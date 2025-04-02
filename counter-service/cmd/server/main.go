package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/yourusername/counter-service/internal/api"
	"github.com/yourusername/counter-service/internal/config"
	"github.com/yourusername/counter-service/internal/counter"
	"github.com/yourusername/counter-service/internal/metrics"
	"github.com/yourusername/counter-service/pkg/logging"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Setup logging
	logger := logging.NewLogger(cfg.LogLevel, cfg.Environment)
	logger.Info().
		Str("version", config.Version).
		Str("environment", cfg.Environment).
		Msg("Counter service starting")

	// Initialize metrics
	metrics := metrics.NewMetrics()

	// Initialize counter service
	counterService, err := counter.NewService(cfg, logger, metrics)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize counter service")
	}

	// Initialize API server
	server := api.NewServer(cfg, logger, counterService, metrics)

	// Handle graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		if err := server.Start(); err != nil {
			logger.Fatal().Err(err).Msg("Server failed")
		}
	}()

	logger.Info().Str("port", cfg.Port).Msg("Server started successfully")

	// Wait for interrupt signal
	<-stop
	logger.Info().Msg("Shutdown signal received")

	// Shutdown server
	if err := server.Shutdown(); err != nil {
		logger.Error().Err(err).Msg("Error during server shutdown")
	}

	logger.Info().Msg("Server shutdown complete")
}
