package main

import (
	"flag"
	"log/slog"
	"os"

	"kafka-traffic-generator/internal/config"

	"github.com/ilyakaznacheev/cleanenv"
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
	// serverModeConfig string
	logLevel string
)

// init initializes the command-line flags.
func init() {
	// Set up a command-line flag for specifying the log level.
	flag.StringVar(&logLevel, "log-level", "", "Logging level: debug, [information], warning, error")

	// Set up a command-line flag for specifying the configuration file path.
	flag.StringVar(&singelModeConfig, "config", "", "Path to the configuration file")

	// Set up a command-line flag for init server-mode.
	// flag.StringVar(&serverModeConfig, "server-config", "", "Configuration for start work in the api server mode")

	// Create a flag set using the `flag` package.
	fset := flag.NewFlagSet("config mode", flag.ContinueOnError)

	// Configure the flag set usage with cleanenv's wrapped flag usage.
	fset.Usage = cleanenv.FUsage(fset.Output(), &config.AppConfig, nil, fset.Usage)

	// Parse the command-line arguments.
	_ = fset.Parse(os.Args[1:])

	// Parse any remaining flags.
	flag.Parse()
}

func main() {
	logger := setupLogger(logLevel)
	logger.Info("Starting kafka-traffic-generator")

	config, err := config.Load(singelModeConfig)
	if err != nil {
		logger.Error("Failed to load configuration", err)
		os.Exit(1)
	}

	switch {
	case config.API.Port != 0:
		logger.Info("Starting the API server on port", slog.Int("port", config.API.Port))
	case len(config.Tasks) > 0:
		logger.Info("Generating fake data for provided tasks")
	default:
		logger.Error("No configuration provided; exiting with an error")
		os.Exit(1)
	}
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
