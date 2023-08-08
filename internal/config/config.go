// Package config provides functionality for loading and handling configuration settings.
package config

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

// Config defines the overall configuration structure.
type Config struct {
	Kafka  Kafka   `yaml:"kafka"`
	Topic  Topic   `yaml:"topic"`
	Fields []Field `yaml:"fields"`
}

// Kafka defines the Kafka-related configuration settings.
type Kafka struct {
	Host string `yaml:"host" env:"KTG_KAFKA" env-description:"Kafka host address" env-default:"localhost:9092"`
}

// Topic defines the structure of a Kafka topic
type Topic struct {
	Name     string `yaml:"name" env:"KTG_TOPIC" env-description:"Kafka topic name" env-required:"true"`
	NumMsgs  int    `yaml:"batch_msgs" env:"KTG_MSGNUM" env-description:"Number of messages per batch" env-default:"100" env-upd:"true"`
	NumBatch int    `yaml:"batch_count" env:"KTG_BATCHNUM" env-description:"Number of batches" env-default:"0" env-upd:"true"`
	MsgDelay int    `yaml:"batch_delay_ms" env:"KTG_DELAY" env-description:"Delay between batches in milliseconds" env-default:"1000" env-upd:"true"`
}

// Field defines the structure of a field configuration for generating fake data.
type Field struct {
	Name     string            `yaml:"name"`
	Function string            `yaml:"function"`
	Params   map[string]string `yaml:"params"`
}

const (
	envConfigPath = "KTG_CONFIG"
)

var (
	configPath string
	config     Config
)

// init initializes the command-line flags.
func init() {
	// Set up a command-line flag for specifying the configuration file path.
	flag.StringVar(&configPath, "config", "", "Path to the configuration file")

	// Create a flag set using the `flag` package.
	fset := flag.NewFlagSet("Environment variables", flag.ContinueOnError)

	// Configure the flag set usage with cleanenv's wrapped flag usage.
	fset.Usage = cleanenv.FUsage(fset.Output(), &config, nil, fset.Usage)

	// Parse the command-line arguments.
	_ = fset.Parse(os.Args[1:])

	// Parse any remaining flags.
	flag.Parse()
}

// Load loads the configuration from a YAML file.
func Load() (*Config, error) {
	// Check if a configuration path is provided; otherwise, fetch from environment variable.
	if configPath == "" {
		log.Println("Configuration not provided via flag, checking environment variables")

		configPath = os.Getenv(envConfigPath)
		if configPath == "" {
			return nil, fmt.Errorf("%s environment variable not set\n", envConfigPath)
		}
	}

	log.Println("Detected configuration file at:", configPath)

	// Check if the configuration file exists.
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, err
	}

	// Read and parse the configuration.
	if err := cleanenv.ReadConfig(configPath, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
