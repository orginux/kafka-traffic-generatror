package main

import (
	"flag"
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
	logLevelWarning = "warning"
	logLevelError   = "err"
)

var (
	singelModeConfig string
	serverModeConfig string
)

// init initializes the command-line flags.
func init() {
	// Set up a command-line flag for specifying the configuration file path.
	flag.StringVar(&singelModeConfig, "config", "", "Path to the configuration file")

	// Set up a command-line flag for init server-mode.
	flag.StringVar(&serverModeConfig, "server-config", "", "Configuration for start work in the api server mode")

	// Create a flag set using the `flag` package.
	// fset := flag.NewFlagSet("config mode", flag.ContinueOnError)

	// Configure the flag set usage with cleanenv's wrapped flag usage.
	// fset.Usage = cleanenv.FUsage(fset.Output(), &config.AppConfig, nil, fset.Usage)

	// Parse the command-line arguments.
	// _ = fset.Parse(os.Args[1:])

	// Parse any remaining flags.
	flag.Parse()
}

func main() {

	switch {
	case singelModeConfig != "":
		// Load the topic description from a YAML file
		config, err := config.Load(singelModeConfig)
		if err != nil {
			fmt.Println("Failed to load configuration", err)
			os.Exit(1)
		}
	case serverModeConfig != "":
		fmt.Println("server mode")
	}

	// Setup the logger based on the environment
	logger := setupLogger(config.LogLevel)
	logger.Info("Starting kafka-traffic-generator")

	// Run the Kafka traffic generator with the provided configuration
	if err = generator.Run(*config, logger); err != nil {
		logger.Error("Failed to run generator", sl.Err(err))
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
