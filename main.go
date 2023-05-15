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

func main() {
	// Load the topic description from a YAML file
	topic := &Topic{}
	yamlFile, err := ioutil.ReadFile("topic.yaml")
	if err != nil {
		log.Fatalf("error reading YAML file: %v", err)
	}
	if err := yaml.Unmarshal(yamlFile, topic); err != nil {
		log.Fatalf("error parsing YAML file: %v", err)
	}

	// Create a Kafka writer
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{"localhost:9092"},
		Topic:   topic.Name,
	})

	// Generate and send messages to Kafka
	for i := 0; i < topic.NumMsgs; i++ {
		// Generate a random message key and value
		key := strconv.Itoa(rand.Intn(100))

		jo := gofakeit.JSONOptions{
			Type:     "object",
			RowCount: 10,
			Fields: []gofakeit.Field{
				{Name: "date", Function: "daterange", Params: gofakeit.MapParams{
					"format":    {"yyyy-MM-dd"},
					"startdate": {"2023-03-13"},
					"enddate":   {"2023-05-16"}}},
				{Name: "Email", Function: "email", Params: gofakeit.MapParams{}},
				{Name: "City", Function: "city", Params: gofakeit.MapParams{}},
				{Name: "Gender", Function: "gender", Params: gofakeit.MapParams{}},
				{Name: "Message", Function: "sentence", Params: gofakeit.MapParams{}},
			},
			Indent: false,
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

		// Send the Kafka message
		err = writer.WriteMessages(context.Background(), kafkaMsg)
		if err != nil {
			log.Fatalf("error sending Kafka message: %v", err)
		}
		log.Println("sent message")

		// Delay before sending the next message
		time.Sleep(time.Duration(topic.MsgDelay) * time.Millisecond)
	}

	// Close the Kafka writer
	err = writer.Close()
	if err != nil {
		log.Fatalf("error closing Kafka writer: %v", err)
	}
}
