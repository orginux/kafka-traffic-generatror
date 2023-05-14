package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/segmentio/kafka-go"
	"gopkg.in/yaml.v2"
)

// Message defines the structure of a Kafka message format
type Message struct {
	Key   string `yaml:"key"`
	Value Foo    `yaml:"value"`
}

type Foo struct {
	Str           string
	Int           int
	Pointer       *int
	Name          string         `fake:"{firstname}"`  // Any available function all lowercase
	Sentence      string         `fake:"{sentence:3}"` // Can call with parameters
	RandStr       string         `fake:"{randomstring:[hello,world]}"`
	Number        string         `fake:"{number:1,10}"`       // Comma separated for multiple values
	Regex         string         `fake:"{regex:[abcdef]{5}}"` // Generate string from regex
	Map           map[string]int `fakesize:"2"`
	Array         []string       `fakesize:"2"`
	ArrayRange    []string       `fakesize:"2,6"`
	Skip          *string        `fake:"skip"` // Set to "skip" to not generate data for
	Created       time.Time      // Can take in a fake tag as well as a format tag
	CreatedFormat time.Time      `fake:"{year}-{month}-{day}" format:"2006-01-02"`
}

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
				{Name: "date", Function: "daterange", Params: gofakeit.MapParams{"format": {"2006-01-02"}}},
				{Name: "message", Function: "sentence", Params: gofakeit.MapParams{}},
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
		fmt.Println("sent message")

		// Delay before sending the next message
		time.Sleep(time.Duration(topic.MsgDelay) * time.Millisecond)
	}

	// Close the Kafka writer
	err = writer.Close()
	if err != nil {
		log.Fatalf("error closing Kafka writer: %v", err)
	}
}
