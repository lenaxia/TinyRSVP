package main

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lenaxia/tinyrsvp/internal/assets"
	"github.com/lenaxia/tinyrsvp/internal/auth"
	"github.com/lenaxia/tinyrsvp/internal/config"
	"github.com/lenaxia/tinyrsvp/internal/db"
	"github.com/lenaxia/tinyrsvp/internal/db/repositories"
	"github.com/lenaxia/tinyrsvp/internal/email"
	"github.com/lenaxia/tinyrsvp/internal/events"
	"github.com/lenaxia/tinyrsvp/internal/handlers"
	"github.com/lenaxia/tinyrsvp/internal/invites"
	"github.com/lenaxia/tinyrsvp/internal/jobs"
	"github.com/lenaxia/tinyrsvp/internal/middleware"
	"github.com/lenaxia/tinyrsvp/internal/models"
	"github.com/lenaxia/tinyrsvp/internal/rsvp"
	"github.com/lenaxia/tinyrsvp/internal/storage"
	"github.com/lenaxia/tinyrsvp/internal/templates"
	"github.com/lenaxia/tinyrsvp/pkg/ics"
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

	logger.Info("Ensuring system user exists")
	bootstrapCtx, bootstrapCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer bootstrapCancel()

	systemUser, err := userRepo.GetByEmail(bootstrapCtx, "system@tinyrsvp.local")
	if err != nil {
		var notFoundErr *models.NotFoundError
		if errors.As(err, &notFoundErr) {
			systemUser = &models.User{
				Email: "system@tinyrsvp.local",
				Name:  "System",
				Role:  models.RoleAdmin,
			}
			if err := userRepo.Create(bootstrapCtx, systemUser); err != nil {
				logger.Error("Failed to create system user", "error", err)
				os.Exit(1)
			}
			logger.Info("System user created", "id", systemUser.ID)
		} else {
			logger.Error("Failed to check for system user", "error", err)
			os.Exit(1)
		}
	} else {
		logger.Info("System user already exists", "id", systemUser.ID)
	}

	logger.Info("Seeding default templates")
	templateRepo := repositories.NewTemplateRepository(database)
	seeder := templates.NewSeeder(templateRepo, systemUser.ID)

	seedCtx, seedCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer seedCancel()

	if err := seeder.SeedDefaults(seedCtx); err != nil {
		logger.Error("Failed to seed default templates", "error", err)
		os.Exit(1)
	}
	logger.Info("Default templates seeded successfully")

	templateEngine := templates.NewEngine()
	templateValidator := templates.NewValidator(templateEngine)
	templateService := templates.NewService(templateRepo, templateValidator)
	logger.Info("Template service initialized")

	sessionRepo := repositories.NewSessionRepository(database)
	eventRepo := repositories.NewEventRepository(database)
	inviteRepo := repositories.NewInviteRepository(database)
	questionRepo := repositories.NewQuestionRepository(database)
	rsvpRepo := repositories.NewRSVPRepository(database)
	answerRepo := repositories.NewAnswerRepository(database)
	emailQueueRepo := repositories.NewEmailQueueRepository(database)

	templateRenderer, err := email.NewTemplateRenderer(&email.TemplateConfig{
		TemplateDir:  "templates/email",
		CacheEnabled: true,
	})
	if err != nil {
		logger.Error("Failed to initialize template renderer", "error", err)
		os.Exit(1)
	}
	logger.Info("Email template renderer initialized")

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

	templateHandlers := handlers.NewTemplateHandlers(templateService)
	templateHandlers.RegisterRoutes(chiRouter)
	logger.Info("Registered template management endpoints", "path", "/api/templates", "protection", "authenticated")

	storageType := os.Getenv("STORAGE_TYPE")
	if storageType == "" {
		storageType = "local"
	}

	storagePath := os.Getenv("STORAGE_PATH")
	if storagePath == "" {
		storagePath = "/data/uploads"
	}

	storageBaseURL := os.Getenv("STORAGE_BASE_URL")
	if storageBaseURL == "" {
		storageBaseURL = cfg.Server.BaseURL
	}

	if storageType == "local" {
		if err := os.MkdirAll(storagePath, 0755); err != nil {
			logger.Error("Failed to create storage directory", "error", err, "path", storagePath)
			os.Exit(1)
		}
		logger.Info("Storage directory ready", "path", storagePath)
	}

	storageProvider, err := storage.NewProvider(&storage.Config{
		Type:     storageType,
		BasePath: storagePath,
		BaseURL:  storageBaseURL,
	})
	if err != nil {
		logger.Error("Failed to create storage provider", "error", err)
		os.Exit(1)
	}
	logger.Info("Storage provider initialized", "type", storageType, "path", storagePath, "baseURL", storageBaseURL)

	imageService := assets.NewImageService(storageProvider)
	imageHandlers := handlers.NewImageHandlers(imageService, eventService, authChecker)
	imageHandlers.RegisterRoutes(chiRouter)
	logger.Info("Registered image management endpoints", "path", "/api/events/{event_id}/images", "protection", "authenticated")

	assetHandler := handlers.NewAssetHandler(storageProvider)
	mux.HandleFunc("/assets/", assetHandler.ServeAsset)
	logger.Info("Registered asset serving endpoint", "path", "/assets/*", "method", "GET,HEAD", "protection", "none")

	staticFS := http.FileServer(http.Dir("static"))
	mux.Handle("/static/", http.StripPrefix("/static/", staticFS))
	logger.Info("Registered static file serving endpoint", "path", "/static/*", "method", "GET,HEAD", "protection", "none")

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

	icsGenerator := ics.NewGenerator()
	emailService := email.NewConfirmationService(templateRenderer, emailQueueRepo, icsGenerator)
	logger.Info("Initialized email confirmation service")

	rsvpService := rsvp.NewServiceWithEmail(database, inviteService, inviteRepo, eventRepo, rsvpRepo, answerRepo, questionRepo, emailService)
	logger.Info("Initialized RSVP service with email support")

	rsvpHandler := handlers.NewRSVPHandler(inviteService, eventRepo, rsvpRepo, questionRepo)
	rsvpHandler.SetTemplates(rsvpTemplates)
	rsvpHandler.SetRSVPService(rsvpService)
	rsvpHandler.SetAnswerRepository(answerRepo)
	
	rsvpRouter := chi.NewRouter()
	rsvpRouter.Get("/{token}", rsvpHandler.GetRSVPPage)
	rsvpRouter.Get("/{token}/confirmation", rsvpHandler.GetConfirmationPage)
	rsvpRouter.Post("/{token}", rsvpHandler.SubmitRSVP)
	rsvpRouter.Put("/{token}", rsvpHandler.UpdateRSVP)
	mux.Handle("/rsvp/", http.StripPrefix("/rsvp", rsvpRouter))
	logger.Info("Registered RSVP page endpoint", "path", "/rsvp/{token}", "method", "GET", "protection", "none")
	logger.Info("Registered RSVP confirmation endpoint", "path", "/rsvp/{token}/confirmation", "method", "GET", "protection", "none")
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

	emailConfig, err := email.LoadConfig()
	if err != nil {
		logger.Error("Failed to load email configuration", "error", err)
		os.Exit(1)
	}
	logger.Info("Email configuration loaded", "config", emailConfig.Sanitized())

	smtpSender, err := email.NewSMTPSender(emailConfig)
	if err != nil {
		logger.Error("Failed to create SMTP sender", "error", err)
		os.Exit(1)
	}

	if emailConfig.TestOnStartup {
		testConnCtx, testConnCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer testConnCancel()
		if err := smtpSender.TestConnection(testConnCtx); err != nil {
			logger.Warn("SMTP connection test failed", "error", err)
		} else {
			logger.Info("SMTP connection test passed")
		}
	}

	emailHealthChecker := email.NewHealthChecker(emailQueueRepo, smtpSender)
	mux.Handle("/api/email/health", requireAuth(requireAdmin(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			status, err := emailHealthChecker.GetStatus(r.Context())
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			statusCode := http.StatusOK
			if !status.Healthy {
				statusCode = http.StatusServiceUnavailable
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(statusCode)
			json.NewEncoder(w).Encode(status)
		}),
	)))
	logger.Info("Registered email health endpoint", "path", "/api/email/health", "method", "GET", "protection", "admin")

	rateLimiter := email.NewRateLimiter(emailConfig.RateLimit)
	emailMetrics := email.NewNoOpMetrics()
	emailLogger := email.NewLogger(logger)
	
	emailProcessor := email.NewQueueProcessor(
		emailQueueRepo,
		smtpSender,
		rateLimiter,
		emailConfig.QueueBatchSize,
		emailConfig.QueuePollInterval,
		emailMetrics,
		emailLogger,
	)

	processorCtx, processorCancel := context.WithCancel(context.Background())
	defer processorCancel()

	processorErrors := make(chan error, 1)
	go func() {
		if err := emailProcessor.Start(processorCtx); err != nil && err != context.Canceled {
			logger.Error("Email processor error", "error", err)
			processorErrors <- err
		}
	}()
	logger.Info("Email queue processor started",
		"batch_size", cfg.Email.ProcessorBatchSize,
		"poll_interval", cfg.Email.ProcessorPollInterval,
	)

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

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		logger.Info("Stopping email processor")
		processorCancel()
		processorStopCtx, processorStopCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer processorStopCancel()
		if err := emailProcessor.Stop(processorStopCtx); err != nil {
			logger.Error("Email processor shutdown failed", "error", err)
		} else {
			logger.Info("Email processor stopped gracefully")
		}

		logger.Info("Closing SMTP sender")
		if err := smtpSender.Close(); err != nil {
			logger.Error("SMTP sender close failed", "error", err)
		} else {
			logger.Info("SMTP sender closed gracefully")
		}

		logger.Info("Stopping background jobs")
		cleanupCancel()

		logger.Info("Stopping HTTP server")
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("Graceful shutdown failed", "error", err)
			if err := server.Close(); err != nil {
				logger.Error("Force close failed", "error", err)
			}
			os.Exit(1)
		}

		logger.Info("Server stopped gracefully")
	}
}
