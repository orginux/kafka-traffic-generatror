package main

import (
	"context"
	"io/ioutil"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/segmentio/kafka-go"
	"gopkg.in/yaml.v2"
)

// Topic defines the structure of a Kafka topic
type Topic struct {
	Name     string `yaml:"name"`
	NumMsgs  int    `yaml:"num_msgs"`
	MsgDelay int    `yaml:"msg_delay"`
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

func main() {
	// Load the topic description from a YAML file
	yamlFile, err := ioutil.ReadFile("topic.yaml")
	if err != nil {
		log.Fatalf("error reading YAML file: %v", err)
	}

	config := Config{}
	if err := yaml.Unmarshal(yamlFile, &config); err != nil {
		log.Fatalf("error parsing YAML file: %v", err)
	}

	// Create a Kafka writer
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{config.Kafka.Host},
		Topic:   config.Topic.Name,
	})

	// Genereate message parampetrs
	var fields []gofakeit.Field
	for _, fild := range config.Fields {
		params := gofakeit.NewMapParams()
		if len(fild.Params) > 0 {
			for key, value := range fild.Params {
				log.Println(key, value)
				params.Add(key, value)
			}
		}
		fields = append(fields, gofakeit.Field{
			Name:     fild.Name,
			Function: fild.Function,
			Params:   *params,
		})
	}

	// Generate and send messages to Kafka
	for i := 0; i < config.Topic.NumMsgs; i++ {
		// Generate a random message key and value
		key := strconv.Itoa(rand.Intn(100))

		jo := gofakeit.JSONOptions{
			Type:     "object",
			RowCount: 10,
			Fields:   fields,
			Indent:   false,
		}

		value, err := gofakeit.JSON(&jo)
		if err != nil {
			log.Fatalln(err)
		}

		// Create a Kafka message with the formatted key and value
		kafkaMsg := kafka.Message{
			Key:   []byte(key),
			Value: []byte(value),
		}

		conn.SetWriteDeadline(time.Now().Add(1 * time.Second))

		// Send the Kafka message
		err = writer.WriteMessages(context.Background(), kafkaMsg)
		if err != nil {
			log.Fatalf("error sending Kafka message: %v", err)
		}
		log.Println("sent message")

		// Delay before sending the next message
		time.Sleep(time.Duration(config.Topic.MsgDelay) * time.Millisecond)
	}

	// Close the Kafka writer
	if err := writer.Close(); err != nil {
		log.Fatal("failed to close writer:", err)
	}
}
