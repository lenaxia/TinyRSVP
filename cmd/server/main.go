package main

import (
	"context"
	"os"
	"time"

	"github.com/yourusername/tinyrsvp/internal/config"
	"github.com/yourusername/tinyrsvp/internal/db"
)

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

	stats := database.DB().Stats()
	logger.Debug("Database connection pool stats",
		"max_open_connections", stats.MaxOpenConnections,
		"open_connections", stats.OpenConnections,
		"in_use", stats.InUse,
		"idle", stats.Idle,
	)

	logger.Info("Server starting",
		"host", cfg.Server.Host,
		"port", cfg.Server.Port,
		"base_url", cfg.Server.BaseURL,
	)

	logger.Info("TinyRSVP Server initialized successfully")
}
