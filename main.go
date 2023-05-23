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

func main() {
	// Load the topic description from a YAML file
	yamlFile, err := ioutil.ReadFile("topic.yaml")
	if err != nil {
		log.Fatalf("error reading YAML file: %v\n", err)
	}

	config := Config{}
	if err := yaml.Unmarshal(yamlFile, &config); err != nil {
		log.Fatalf("error parsing YAML file: %v\n", err)
	}

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

	for i := 0; i < config.Topic.NumBatch; i++ {
		// Generate a batch of messages
		var batch []kafka.Message
		for j := 0; j < config.Topic.NumMsgs; j++ {
			// Generate a random message key and value
			key := strconv.Itoa(rand.Intn(100))

			jo := gofakeit.JSONOptions{
				Type:   "object", // array or object
				Fields: fields,   // internal_exampleFields
				Indent: false,    // indent
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
			batch = append(batch, kafkaMsg)

		}

		// Create a Kafka connection
		conn := kafka.NewWriter(kafka.WriterConfig{
			Brokers: []string{config.Kafka.Host},
			Topic:   config.Topic.Name,
		})

		// Send the Kafka message
		err = conn.WriteMessages(context.Background(), batch...)
		if err != nil {
			log.Fatalf("failed to write messages: %v\n", err)
		}
		log.Printf("sent batch of %d messages\n", config.Topic.NumMsgs)

		// Close the Kafka connection
		if err := conn.Close(); err != nil {
			log.Fatalf("failed to close the Kafka connection: %v\n", err)
		}

		// Delay before sending the next message
		log.Printf("delay %d ms before the next batch\n", config.Topic.MsgDelay)
		time.Sleep(time.Duration(config.Topic.MsgDelay) * time.Millisecond)
	}
}
