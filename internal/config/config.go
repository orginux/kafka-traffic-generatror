// Package config provides functionality for loading and handling configuration settings.
package config

import (
	"flag"
	"fmt"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/segmentio/kafka-go"
)

// Config defines the overall configuration structure.
type Config struct {
	Kafka    Kafka   `yaml:"kafka"`
	Topic    Topic   `yaml:"topic"`
	Fields   []Field `yaml:"fields"`
	LogLevel string  `yaml:"loglevel" env:"KTG_LOGLEVEL" env-description:"Logging level [debug, info, warn, error]"`
}

// Kafka defines the Kafka-related configuration settings.
type Kafka struct {
	Host string `yaml:"host" env:"KTG_KAFKA" env-description:"Kafka host address" env-default:"localhost:9092"`
	Acks string `yaml:"acks" env:"KTG_ACKS" env-description:"Kafka acks setting [all, one, none]" env-default:"none"`
	TLS  KafkaCerts
}

// Kafka acks settings, available options
const (
	kafkaAcksAll  = "all"
	kafkaAcksOne  = "one"
	kafkaAcksNone = "none"
)

// ParseAcks parses the acks configuration setting
func (k *Kafka) ParseAcks() (kafka.RequiredAcks, error) {
	switch k.Acks {
	case kafkaAcksAll:
		return kafka.RequireAll, nil
	case kafkaAcksOne:
		return kafka.RequireOne, nil
	case kafkaAcksNone:
		return kafka.RequireNone, nil
	default:
		return kafka.RequireNone, fmt.Errorf("invalid acks configuration")
	}
}

// Kafka TLS
type KafkaCerts struct {
	CaPath   string `yaml:"ca_path" env:"KTG_CA_PATH" env-description:"Path to the CA certificate"`
	CertPath string `yaml:"cert_path" env:"KTG_CERT_PATH" env-description:"Path to the client certificate"`
	KeyPath  string `yaml:"key_path" env:"KTG_KEY_PATH" env-description:"Path to the client key"`
}

// Read certicates from files
func (kc *KafkaCerts) Read() (caCert, clientCert, clientKey []byte, err error) {
	caCert, err = os.ReadFile(kc.CaPath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to read CA certificate: %w", err)
	}
	clientCert, err = os.ReadFile(kc.CertPath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to read client certificate: %w", err)
	}
	clientKey, err = os.ReadFile(kc.KeyPath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to read client key: %w", err)
	}
	return caCert, clientCert, clientKey, nil
}

// Topic defines the structure of a Kafka topic
type Topic struct {
	Name       string `yaml:"name" env:"KTG_TOPIC" env-description:"Kafka topic name" env-required:"true"`
	NumMsgs    int    `yaml:"batch_msgs" env:"KTG_MSGNUM" env-description:"Number of messages per batch" env-default:"100" env-upd:"true"`
	NumBatch   int    `yaml:"batch_count" env:"KTG_BATCHNUM" env-description:"Number of batches (0 - unlimited)" env-default:"0" env-upd:"true"`
	BatchDelay int    `yaml:"batch_delay_ms" env:"KTG_BATCHDELAY" env-description:"Delay between batches in milliseconds" env-default:"0" env-upd:"true"`
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
	if err := cleanenv.ReadConfig(configPath, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
