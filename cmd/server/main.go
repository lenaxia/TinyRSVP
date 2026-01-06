package main

import (
	"os"

	"github.com/yourusername/tinyrsvp/internal/config"
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

	logger.Info("Server starting",
		"host", cfg.Server.Host,
		"port", cfg.Server.Port,
		"base_url", cfg.Server.BaseURL,
	)

	logger.Info("TinyRSVP Server initialized successfully")
}
