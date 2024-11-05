package main

import (
	"log/slog"
	"os"

	_ "github.com/threefoldtech/tf-kyc-verifier/api/docs"
	"github.com/threefoldtech/tf-kyc-verifier/internal/config"
	"github.com/threefoldtech/tf-kyc-verifier/internal/server"
)

func main() {
	config, err := config.LoadConfigFromEnv()
	if err != nil {
		slog.Error("Failed to load configuration:", "error", err)
		os.Exit(1)
	}
	logLevel := slog.LevelInfo
	if config.Log.Debug {
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel}))
	logger.Debug("Configuration loaded successfully", "config", config.GetPublicConfig())

	server, err := server.New(config, logger)
	if err != nil {
		logger.Error("Failed to create server:", "error", err)
		os.Exit(1)
	}

	logger.Info("Starting server on port", "port", config.Server.Port)
	err = server.Run()
	if err != nil {
		logger.Error("Server exited with error", "error", err)
		os.Exit(1)
	}
}
