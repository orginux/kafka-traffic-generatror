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

var (
	configPath string
	config     Config
)

// init initializes the command-line flags.
func init() {
	flag.StringVar(&configPath, "config", "", "config file path")

	// create flag set using `flag` package
	fset := flag.NewFlagSet("Environment variables", flag.ContinueOnError)

	// get config usage with wrapped flag usage
	fset.Usage = cleanenv.FUsage(fset.Output(), &config, nil, fset.Usage)

	_ = fset.Parse(os.Args[1:])
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

	if err := cleanenv.ReadConfig(configPath, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
