package main

import (
	"context"
	"encoding/hex"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/config"
	"github.com/lenaxia/tinyrsvp/internal/db"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/events"
	"github.com/lenaxia/tinyrsvp/internal/handlers"
	"github.com/lenaxia/tinyrsvp/internal/invites"
	"github.com/lenaxia/tinyrsvp/internal/jobs"
	"github.com/lenaxia/tinyrsvp/internal/middleware"
	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/internal/rsvp"
	"github.com/lenaxia/tinyrsvp/pkg/token"
)

const appVersion = "0.1.0"

func main() {
	logLevel := config.GetLogLevelFromEnv()
	logger := config.InitLogger(logLevel)

	logger.Info("Starting TinyRSVP Server")

	cfg, err := config.Load()
	if err != nil {
		logger.Error("Failed to load configuration", "error", err)
		os.Exit(1)
	}

	logger.Info("Configuration loaded successfully")
	logger.Debug("Configuration details", "config", cfg.String())

	database, err := db.NewDatabase(db.Config{
		Type:         cfg.Database.Type,
		Path:         cfg.Database.Path,
		MaxOpenConns: cfg.Database.MaxOpenConns,
		MaxIdleConns: cfg.Database.MaxIdleConns,
		MaxLifetime:  cfg.Database.MaxLifetime,
	})
	if err != nil {
		logger.Error("Failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	logger.Info("Database connection established",
		"type", cfg.Database.Type,
		"path", cfg.Database.Path,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := database.Ping(ctx); err != nil {
		logger.Error("Database health check failed", "error", err)
		os.Exit(1)
	}

	logger.Info("Database health check passed")

	logger.Info("Running database migrations")
	migrator, err := db.NewMigrator(database.DB(), "migrations/sqlite")
	if err != nil {
		logger.Error("Failed to create migrator", "error", err)
		os.Exit(1)
	}

	migrationCtx, migrationCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer migrationCancel()

	if err := migrator.Up(migrationCtx); err != nil {
		logger.Error("Failed to run migrations", "error", err)
		os.Exit(1)
	}

	version, dirty, err := migrator.Version(migrationCtx)
	if err != nil {
		logger.Warn("Failed to get migration version", "error", err)
	} else {
		logger.Info("Database migrations completed",
			"version", version,
			"dirty", dirty,
		)
	}

	stats := database.DB().Stats()
	logger.Debug("Database connection pool stats",
		"max_open_connections", stats.MaxOpenConnections,
		"open_connections", stats.OpenConnections,
		"in_use", stats.InUse,
		"idle", stats.Idle,
	)

	userRepo := repositories.NewUserRepository(database)
	sessionRepo := repositories.NewSessionRepository(database)
	eventRepo := repositories.NewEventRepository(database)
	inviteRepo := repositories.NewInviteRepository(database)
	questionRepo := repositories.NewQuestionRepository(database)
	rsvpRepo := repositories.NewRSVPRepository(database)
	answerRepo := repositories.NewAnswerRepository(database)

	sessionMgr := auth.NewSessionManager(sessionRepo, false)
	userService := auth.NewUserService(userRepo)
	authChecker := auth.NewAuthorizationChecker()

	logger.Info("Initialized auth services")

	eventValidator := events.NewValidator(events.NewTimezoneValidator())
	eventService := events.NewService(eventRepo, eventValidator, authChecker)

	logger.Info("Initialized event services")

	questionValidator := events.NewQuestionValidator()
	questionService := events.NewQuestionService(eventRepo, questionRepo, questionValidator, authChecker)

	logger.Info("Initialized question services")

	tokenSecretBytes, err := hex.DecodeString(cfg.Token.Secret)
	if err != nil {
		tokenSecretBytes = []byte(cfg.Token.Secret)
	}
	tokenGenerator := token.NewGenerator(tokenSecretBytes)
	inviteService := invites.NewInviteService(tokenGenerator, inviteRepo)
	individualInviteService := invites.NewIndividualInviteService(tokenGenerator, inviteRepo, eventRepo)

	logger.Info("Initialized invite services")

	var authenticator auth.Authenticator
	if cfg.OIDC.Enabled {
		oidcCfg := &auth.OIDCConfig{
			IssuerURL:    cfg.OIDC.IssuerURL,
			ClientID:     cfg.OIDC.ClientID,
			ClientSecret: cfg.OIDC.ClientSecret,
			RedirectURL:  cfg.OIDC.RedirectURL,
			Scopes:       []string{"openid", "email", "profile"},
		}
		oidcAuth, err := auth.NewOIDCAuthenticator(oidcCfg, userService, sessionMgr)
		if err != nil {
			logger.Error("Failed to create OIDC authenticator", "error", err)
			os.Exit(1)
		}
		authenticator = oidcAuth
		logger.Info("OIDC authentication enabled")
	} else if cfg.ForwardAuth.Enabled {
		fwdAuthCfg := &auth.ForwardAuthConfig{
			UserHeader:  cfg.ForwardAuth.UserHeader,
			EmailHeader: cfg.ForwardAuth.EmailHeader,
			TrustedIPs:  cfg.ForwardAuth.TrustedIPs,
		}
		authenticator = auth.NewForwardAuthenticator(fwdAuthCfg, userService, sessionMgr)
		logger.Info("Forward auth enabled")
	} else {
		logger.Error("No authentication method enabled")
		os.Exit(1)
	}

	loginHandler := auth.NewLoginHandler(authenticator)
	callbackHandler := auth.NewCallbackHandler(authenticator, userService, sessionMgr)
	logoutHandler := auth.NewLogoutHandler(authenticator)

	mux := http.NewServeMux()

	mux.Handle("/login", loginHandler)
	logger.Info("Registered auth endpoint", "path", "/login")

	mux.Handle("/auth/callback", callbackHandler)
	logger.Info("Registered auth endpoint", "path", "/auth/callback")

	mux.Handle("/logout", logoutHandler)
	logger.Info("Registered auth endpoint", "path", "/logout")

	healthHandler := handlers.NewHealthHandler(appVersion)
	mux.Handle("/health", healthHandler)
	logger.Info("Registered health endpoint", "path", "/health")

	readinessHandler := handlers.NewReadinessHandler(appVersion, database, migrator)
	mux.Handle("/ready", readinessHandler)
	logger.Info("Registered readiness endpoint", "path", "/ready")

	userHandler := handlers.NewUserHandler(userService, authChecker)

	requireAuth := middleware.RequireAuth(sessionMgr, userService)
	requireAdmin := middleware.RequireAdmin(authChecker)

	mux.Handle("/api/users", requireAuth(requireAdmin(http.HandlerFunc(userHandler.ListUsers))))
	logger.Info("Registered user management endpoint", "path", "/api/users", "method", "GET", "protection", "admin")

	mux.HandleFunc("/api/users/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/users/")
		if path == "" {
			http.NotFound(w, r)
			return
		}

		parts := strings.Split(path, "/")
		userID := parts[0]

		switch r.Method {
		case http.MethodGet:
			requireAuth(requireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				userHandler.GetUser(w, r, userID)
			}))).ServeHTTP(w, r)
		case http.MethodPatch:
			requireAuth(requireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				userHandler.UpdateUserRole(w, r, userID)
			}))).ServeHTTP(w, r)
		case http.MethodDelete:
			requireAuth(requireAdmin(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				userHandler.DeleteUser(w, r, userID)
			}))).ServeHTTP(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	logger.Info("Registered user management endpoints", "path", "/api/users/{id}", "methods", "GET,PATCH,DELETE", "protection", "admin")

	eventHandlers := handlers.NewEventHandlers(eventService)
	chiRouter := chi.NewRouter()
	chiRouter.Use(func(next http.Handler) http.Handler {
		return requireAuth(next)
	})
	eventHandlers.RegisterRoutes(chiRouter)

	questionHandlers := handlers.NewQuestionHandlers(questionService)
	questionHandlers.RegisterRoutes(chiRouter)

	inviteHandlers := handlers.NewInviteHandlers(individualInviteService, cfg.Server.BaseURL)
	inviteHandlers.RegisterRoutes(chiRouter)

	importInviteHandlers := handlers.NewImportInviteHandlers(inviteService, eventRepo, cfg.Server.BaseURL)
	importInviteHandlers.RegisterRoutes(chiRouter)

	manualInviteHandlers := handlers.NewManualInviteHandlers(inviteService, eventRepo, cfg.Server.BaseURL)
	manualInviteHandlers.RegisterRoutes(chiRouter)

	revokeInviteHandlers := handlers.NewRevokeInviteHandlers(inviteService, eventRepo)
	revokeInviteHandlers.RegisterRoutes(chiRouter)

	regenerateInviteHandlers := handlers.NewRegenerateInviteTokenHandlers(inviteService, eventRepo)
	regenerateInviteHandlers.RegisterRoutes(chiRouter)

	listInviteHandlers := handlers.NewListInviteHandlers(inviteService, eventRepo)
	listInviteHandlers.RegisterRoutes(chiRouter)

	cleanupHandler := handlers.NewCleanupHandler(inviteService)
	mux.Handle("/api/invites/cleanup", requireAuth(requireAdmin(cleanupHandler)))
	logger.Info("Registered invite cleanup endpoint", "path", "/api/invites/cleanup", "method", "POST", "protection", "admin")

	mux.Handle("/api/events", chiRouter)
	mux.Handle("/api/events/", chiRouter)
	logger.Info("Registered event management endpoints", "path", "/api/events", "protection", "authenticated")
	logger.Info("Registered question management endpoints", "path", "/api/events/{id}/questions", "protection", "authenticated")
	logger.Info("Registered invite management endpoints", "path", "/api/events/{eventId}/invites", "protection", "authenticated")
	logger.Info("Registered import invite endpoints", "path", "/api/events/{eventId}/invites/import", "protection", "authenticated")
	logger.Info("Registered manual invite endpoints", "path", "/api/events/{eventId}/invites/manual", "protection", "authenticated")
	logger.Info("Registered revoke invite endpoints", "path", "/api/invites/{inviteId}/revoke", "protection", "authenticated")
	logger.Info("Registered regenerate invite endpoints", "path", "/api/invites/{inviteId}/regenerate", "protection", "authenticated")
	logger.Info("Registered list invite endpoints", "path", "/api/events/{eventId}/invites", "method", "GET", "protection", "authenticated")

	rsvpTemplates, err := template.ParseFiles("templates/web/rsvp_page.html")
	if err != nil {
		logger.Error("Failed to load RSVP templates", "error", err)
		os.Exit(1)
	}
	logger.Info("RSVP templates loaded successfully")

	rsvpService := rsvp.NewService(database, inviteService, inviteRepo, eventRepo, rsvpRepo, answerRepo, questionRepo)
	logger.Info("Initialized RSVP service")

	rsvpHandler := handlers.NewRSVPHandler(inviteService, eventRepo, rsvpRepo, questionRepo)
	rsvpHandler.SetTemplates(rsvpTemplates)
	rsvpHandler.SetRSVPService(rsvpService)
	
	rsvpRouter := chi.NewRouter()
	rsvpRouter.Get("/{token}", rsvpHandler.GetRSVPPage)
	rsvpRouter.Post("/{token}", rsvpHandler.SubmitRSVP)
	rsvpRouter.Put("/{token}", rsvpHandler.UpdateRSVP)
	mux.Handle("/rsvp/", http.StripPrefix("/rsvp", rsvpRouter))
	logger.Info("Registered RSVP page endpoint", "path", "/rsvp/{token}", "method", "GET", "protection", "none")
	logger.Info("Registered RSVP submission endpoint", "path", "/rsvp/{token}", "method", "POST", "protection", "none")
	logger.Info("Registered RSVP update endpoint", "path", "/rsvp/{token}", "method", "PUT", "protection", "none")

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	cleanupCtx, cleanupCancel := context.WithCancel(context.Background())
	defer cleanupCancel()

	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				count, err := sessionMgr.CleanupExpired(context.Background())
				if err != nil {
					logger.Error("Session cleanup failed", "error", err)
				} else {
					logger.Info("Session cleanup completed", "deleted", count)
				}
			case <-cleanupCtx.Done():
				logger.Info("Session cleanup goroutine stopped")
				return
			}
		}
	}()
	logger.Info("Session cleanup background job started")

	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				count, err := inviteService.CleanupExpiredTokens(context.Background())
				if err != nil {
					logger.Error("Invite token cleanup failed", "error", err)
				} else {
					logger.Info("Invite token cleanup completed", "deleted", count)
				}
			case <-cleanupCtx.Done():
				logger.Info("Invite token cleanup goroutine stopped")
				return
			}
		}
	}()
	logger.Info("Invite token cleanup background job started")

	eventArchiver := jobs.NewEventArchiver(eventService, 30)
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				systemUser := &models.User{
					ID:    0,
					Email: "system@tinyrsvp.local",
					Name:  "System",
					Role:  models.RoleAdmin,
				}
				archiveCtx := auth.WithUser(context.Background(), systemUser)

				if err := eventArchiver.Run(archiveCtx); err != nil {
					logger.Error("Event archiving job failed", "error", err)
				}
			case <-cleanupCtx.Done():
				logger.Info("Event archiving goroutine stopped")
				return
			}
		}
	}()
	logger.Info("Event archiving background job started")

	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("Server starting", "address", addr)
		serverErrors <- server.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		logger.Error("Server error", "error", err)
		os.Exit(1)

	case sig := <-shutdown:
		logger.Info("Shutdown signal received", "signal", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			logger.Error("Graceful shutdown failed", "error", err)
			if err := server.Close(); err != nil {
				logger.Error("Force close failed", "error", err)
			}
			os.Exit(1)
		}

		logger.Info("Server stopped gracefully")
	}
}
