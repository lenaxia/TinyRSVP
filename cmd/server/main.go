package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/yourusername/tinyrsvp/internal/auth"
	"github.com/yourusername/tinyrsvp/internal/config"
	"github.com/yourusername/tinyrsvp/internal/db"
	"github.com/yourusername/tinyrsvp/internal/db/repositories"
	"github.com/yourusername/tinyrsvp/internal/handlers"
	"github.com/yourusername/tinyrsvp/internal/middleware"
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

	sessionMgr := auth.NewSessionManager(sessionRepo, false)
	userService := auth.NewUserService(userRepo)
	authChecker := auth.NewAuthorizationChecker()

	logger.Info("Initialized auth services")

	mux := http.NewServeMux()

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

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

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
