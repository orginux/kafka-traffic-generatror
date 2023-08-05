package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Topic defines the structure of a Kafka topic
type Topic struct {
	Name     string `yaml:"name"`
	NumMsgs  int    `yaml:"batch_msgs"`
	NumBatch int    `yaml:"batch_count"`
	MsgDelay int    `yaml:"batch_delay_ms"`
}

type Field struct {
	Name     string            `yaml:"name"`
	Function string            `yaml:"function"`
	Params   map[string]string `yaml:"params"`
}

type Kafka struct {
	Host string `yaml:"host"`
}

type Config struct {
	Kafka  Kafka   `yaml:"kafka"`
	Topic  Topic   `yaml:"topic"`
	Fields []Field `yaml:"fields"`
}

// Load loads the configuration from a YAML file.
func Load(filename string) (Config, error) {
	yamlFile, err := os.ReadFile(filename)
	if err != nil {
		return Config{}, fmt.Errorf("Error reading YAML file: %v\n", err)
	}

	var config Config
	if err := yaml.Unmarshal(yamlFile, &config); err != nil {
		return Config{}, fmt.Errorf("Error parsing YAML file: %v\n", err)
	}

	return config, nil
}