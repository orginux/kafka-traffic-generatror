package main

import (
	"log"
	"log/slog"
	"os"

	"kafka-traffic-generator/internal/config"
	"kafka-traffic-generator/internal/generator"
	"kafka-traffic-generator/internal/lib/logger/sl"
)

const (
	// envLocal = "local"
	envDev  = "dev"
	envProd = "prod"
)

func main() {

	// Load the topic description from a YAML file
	config, err := config.Load()
	if err != nil {
		log.Fatalln("Failed to load configuration", err)
	}

	// Setup the logger based on the environment
	logger := setupLogger(config.Env)
	logger.Info("Starting kafka-traffic-generator")

	// Run the Kafka traffic generator with the provided configuration
	if err = generator.Run(*config); err != nil {
		logger.Error("Failed to run generator", sl.Err(err))
	}
	logger.Info("Stopping kafka-traffic-generator")
}

// setupLogger configures and returns a logger based on the environment.
func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	// Local environment
	default:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	}

	return log
}
