// Package config provides functionality for loading and handling configuration settings.
package config

import (
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

// Kafka defines the Kafka-related configuration settings.
type Kafka struct {
	Host string `yaml:"host" env:"KTG_KAFKA" env-description:"Kafka host address" env-default:"localhost:9092"`
}

// Task defines the structure of each task
type Task struct {
	Name   string
	Topic  Topic   `yaml:"topic"`
	Fields []Field `yaml:"fields"`
}

// Topic defines the structure of a Kafka topic
type Topic struct {
	Name     string `yaml:"name" env:"KTG_TOPIC" env-description:"Kafka topic name" env-required:"true"`
	NumMsgs  int    `yaml:"batch_msgs" env:"KTG_MSGNUM" env-description:"Number of messages per batch" env-default:"100" env-upd:"true"`
	NumBatch int    `yaml:"batch_count" env:"KTG_BATCHNUM" env-description:"Number of batches (0 - unlimited)" env-default:"0" env-upd:"true"`
	MsgDelay int    `yaml:"batch_delay_ms" env:"KTG_DELAY" env-description:"Delay between batches in milliseconds" env-default:"1000" env-upd:"true"`
}

// Field defines the structure of a field configuration for generating fake data.
type Field struct {
	Name     string            `yaml:"name"`
	Function string            `yaml:"function"`
	Params   map[string]string `yaml:"params"`
}

// API defines the structure of a fields configuration for the api server.
type API struct {
	Port int `yaml:"port"`
}

// Config defines the overall configuration structure.
type Config struct {
	Kafka Kafka  `yaml:"kafka"`
	API   API    `yaml:"api"`
	Tasks []Task `yaml:"tasks"`
}

const (
	envConfigPath = "KTG_CONFIG"
)

// Load loads the configuration from a YAML file.
func Load(configPath string) (*Config, error) {
	var appConfig Config

	// Check if a configuration path is provided; otherwise, fetch from environment variable.
	if configPath == "" {
		fmt.Println("Configuration not provided via flag, checking environment variables")

		configPath = os.Getenv(envConfigPath)
		if configPath == "" {
			return nil, fmt.Errorf("%s environment variable not set\n", envConfigPath)
		}
	}

	fmt.Println("Detected configuration file at:", configPath)

	// Check if the configuration file exists.
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, err
	}

	// Read and parse the configuration.
	if err := cleanenv.ReadConfig(configPath, &appConfig); err != nil {
		return nil, err
	}

	return &appConfig, nil
}
