// Package config provides functionality for loading and handling configuration settings.
package config

import (
	"flag"
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
	Host string `yaml:"host" env:"KTG_KAFKA" env-default:"localhost:9092"`
}

// Topic defines the structure of a Kafka topic
type Topic struct {
	Name     string `yaml:"name" env:"KTG_TOPIC"`
	NumMsgs  int    `yaml:"batch_msgs" env:"KTG_MSGNUM" env-upd`
	NumBatch int    `yaml:"batch_count" env:"KTG_BATCHNUM" env-default:"0" env-upd`
	MsgDelay int    `yaml:"batch_delay_ms" env:"KTG_DELAY" env-default:"1000" env-upd`
}

// Field defines the structure of a field configuration for generating fake data.
type Field struct {
	Name     string            `yaml:"name"`
	Function string            `yaml:"function"`
	Params   map[string]string `yaml:"params"`
}

var (
	configPath string
)

// init initializes the command-line flags.
func init() {
	flag.StringVar(&configPath, "config", "", "config file path")
	flag.Parse()
}

// Load loads the configuration from a YAML file.
func Load() (*Config, error) {
	log.Println(configPath)
	if configPath == "" {
		configPath = os.Getenv("KTG_CONFIG")
	}
	log.Println(configPath)

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, err
	}

	var config Config
	if err := cleanenv.ReadConfig(configPath, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
