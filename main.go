package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/segmentio/kafka-go"
	"gopkg.in/yaml.v2"
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

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "", "config file path")

	flag.Parse()
}

func main() {
	// Load the topic description from a YAML file
	config, err := loadConfig(configFile)
	if err != nil {
		log.Fatalln(err)
	}

	// Generate message parameters
	fields := generateFields(config.Fields)

	// Generate and send batches of messages
	var batchNum int
	for batchNum <= config.Topic.NumBatch {
		batch, err := generateBatch(config.Topic.NumMsgs, fields)
		if err != nil {
			log.Fatalln(err)
		}

		if err := sendBatch(config.Kafka.Host, config.Topic.Name, batch); err != nil {
			log.Fatalln(err)
		}

		// Delay before sending the next batch
		log.Printf("Delaying %d ms before the next batch\n", config.Topic.MsgDelay)
		time.Sleep(time.Duration(config.Topic.MsgDelay) * time.Millisecond)

		if config.Topic.NumBatch > 0 {
			batchNum++
		}
	}
}

// loadConfig loads the configuration from a YAML file.
func loadConfig(filename string) (Config, error) {
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

// generateBatch generates a batch of Kafka messages with random key-value pairs.
func generateFields(fieldConfigs []Field) []gofakeit.Field {
	var fields []gofakeit.Field
	for _, fieldConfig := range fieldConfigs {
		params := gofakeit.NewMapParams()
		for key, value := range fieldConfig.Params {
			params.Add(key, value)
		}

		field := gofakeit.Field{
			Name:     fieldConfig.Name,
			Function: fieldConfig.Function,
			Params:   *params,
		}

		fields = append(fields, field)
	}

	return fields
}

// sendBatch sends a batch of Kafka messages to the specified topic.
func generateBatch(numMsgs int, fields []gofakeit.Field) ([]kafka.Message, error) {
	var batch []kafka.Message
	for i := 0; i < numMsgs; i++ {
		// TODO make it optional
		key := strconv.Itoa(rand.Intn(100))

		// Generate the random fake data
		jo := gofakeit.JSONOptions{
			Type:   "object",
			Fields: fields,
			Indent: false,
		}
		value, err := gofakeit.JSON(&jo)
		if err != nil {
			return nil, fmt.Errorf("Error of generate random data: %v\n", err)
		}

		// Prepare a Kafka message with the random data
		msg := kafka.Message{
			Key:   []byte(key),
			Value: []byte(value),
		}
		batch = append(batch, msg)
	}
	return batch, nil
}

func sendBatch(host, topic string, batch []kafka.Message) error {
	conn := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{host},
		Topic:   topic,
	})

	err := conn.WriteMessages(context.Background(), batch...)
	if err != nil {
		return fmt.Errorf("Failed to write messages: %v\n", err)
	}
	log.Printf("Sent batch of %d messages\n", len(batch))

	if err := conn.Close(); err != nil {
		return fmt.Errorf("Failed to close the Kafka connection: %v\n", err)
	}

	return nil
}
