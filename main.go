package main

import (
	"fmt"
	"log/slog"
	"os"

	"kafka-traffic-generator/internal/config"
	"kafka-traffic-generator/internal/generator"
	"kafka-traffic-generator/internal/lib/logger/sl"
)

const (
	//  Logging levels:
	// debug, information (default), warning, error.
	logLevelDebug   = "debug"
	logLevelWarning = "warn"
	logLevelError   = "err"
)

func main() {
	// Load the topic description from a YAML file
	config, err := config.Load()
	if err != nil {
		fmt.Println("Failed to load configuration", err)
		os.Exit(1)
	}

	// Setup the logger based on the environment
	logger := setupLogger(config.LogLevel)
	logger.Info("Starting kafka-traffic-generator")

	// Run the Kafka traffic generator with the provided configuration
	generator := generator.New(config, logger)
	if err = generator.Run(); err != nil {
		logger.Error("Failed to run kafka-traffic-generator", sl.Err(err))
	}
	logger.Info("Stopping kafka-traffic-generator")
}

// setupLogger configures and returns a logger based on the environment.
func setupLogger(level string) *slog.Logger {
	var log *slog.Logger

	switch level {
	case logLevelDebug:
		// dev
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
		// prod
	case logLevelWarning:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}),
		)
	case logLevelError:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}),
		)
	default:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
