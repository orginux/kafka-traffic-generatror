// Package generator provides functionalities to generate and send synthetic Kafka traffic.
package generator

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/segmentio/kafka-go"

	"kafka-traffic-generator/internal/config"
)

// Run generates and sends batches of Kafka messages based onthe provided configuration.
func Run(config config.Config) error {
	// Generate message parameters
	fields := generateFields(config.Fields)

	// Generate and send batches of messages
	var batchNum int
	for batchNum <= config.Topic.NumBatch {
		batch, err := generateBatch(config.Topic.NumMsgs, fields)
		if err != nil {
			return err
		}

		if err := sendBatch(config.Kafka.Host, config.Topic.Name, batch); err != nil {
			return err
		}

		// Delay before sending the next batch
		log.Printf("Delaying %d ms before the next batch\n", config.Topic.MsgDelay)
		time.Sleep(time.Duration(config.Topic.MsgDelay) * time.Millisecond)

		// If NumBatch =< 0 (default) -- Unlimited number of batches
		if config.Topic.NumBatch > 0 {
			batchNum++
		}
	}
	return nil
}

// generateFields generates a slice of gofakeit.Field based on the provided field configurations.
func generateFields(fieldConfigs []config.Field) []gofakeit.Field {
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

// generateBatch generates a batch of Kafka messages with random key-value pairs.
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
			return nil, fmt.Errorf("Error generating random data: %v\n", err)
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

// sendBatch sends a batch of Kafka messages to the specified topic.
func sendBatch(host, topic string, batch []kafka.Message) error {
	conn := kafka.Writer{
		Addr:  kafka.TCP(host),
		Topic: topic,
	}

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
