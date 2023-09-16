// Package generator provides functionalities to generate and send synthetic Kafka traffic.
package generator

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"strconv"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl/plain"

	"kafka-traffic-generator/internal/config"
)

// Run generates and sends batches of Kafka messages based on the provided configuration.
func Run(config config.Config, logger *slog.Logger) error {
	logger = logger.With(
		slog.String("component", "generator"),
	)
	logger.Info("Generator started")

	// Generate message parameters
	fields, err := generateFields(config.Fields, logger)
	if err != nil {
		return err
	}

	// Generate and send batches of messages
	var batchNum int
	for batchNum <= config.Topic.NumBatch {
		logger.Info("Batch statistics", slog.Int("batch number", batchNum))
		batch, err := generateBatch(config.Topic.NumMsgs, fields, logger)
		if err != nil {
			return err
		}

		if err := sendBatch(config.Kafka.Host, config.Topic.Name, batch, logger); err != nil {
			logger.Debug("Failed to send the batch")
			return err
		}

		// Delay before sending the next batch
		if config.Topic.MsgDelay > 0 {
			logger.Info("Delaying before the next batch:", slog.Int("milliseconds", config.Topic.MsgDelay))
			time.Sleep(time.Duration(config.Topic.MsgDelay) * time.Millisecond)
		}

		// If NumBatch <= 0 (default) -- Unlimited number of batches
		if config.Topic.NumBatch > 0 {
			batchesLeft := config.Topic.NumBatch - batchNum
			logger.Info("Batch statistics", slog.Int("left:", batchesLeft))
			batchNum++
		}
	}
	return nil
}

// generateFields generates a slice of gofakeit.Field based on the provided field configurations.
func generateFields(fieldConfigs []config.Field, logger *slog.Logger) ([]gofakeit.Field, error) {
	if len(fieldConfigs) == 0 {
		return nil, fmt.Errorf("Fields are not defined in the config file")
	}

	var fields []gofakeit.Field
	for _, fieldConfig := range fieldConfigs {
		logger.Debug("Preparing fake data", slog.String("field", fieldConfig.Name), slog.String("function", fieldConfig.Function))
		params := gofakeit.NewMapParams()
		for key, value := range fieldConfig.Params {
			logger.Debug(
				"Preparing fake data",
				slog.String("field", fieldConfig.Name),
				slog.String("function", fieldConfig.Function),
				slog.String("param", key),
			)
			params.Add(key, value)
		}

		field := gofakeit.Field{
			Name:     fieldConfig.Name,
			Function: fieldConfig.Function,
			Params:   *params,
		}

		fields = append(fields, field)
	}
	logger.Debug("Fields generated")
	return fields, nil
}

// generateBatch generates a batch of Kafka messages with random key-value pairs.
func generateBatch(numMsgs int, fields []gofakeit.Field, logger *slog.Logger) ([]kafka.Message, error) {
	logger = logger.With(
		slog.String("subcomponent", "batch generator"),
	)
	var batch []kafka.Message

	for i := 0; i < numMsgs; i++ {
		// TODO make it optional
		key := strconv.Itoa(rand.Intn(100))
		logger.Debug("Generate batch", slog.String("Kafka message key", key))

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
		logger.Debug("Message generated", slog.Int("msg num", i+1), slog.Int("ftom", numMsgs))
	}
	logger.Info("Batch generated", slog.Int("messages in batch", len(batch)))
	return batch, nil
}

// sendBatch sends a batch of Kafka messages to the specified topic.
func sendBatch(host, topic string, batch []kafka.Message, logger *slog.Logger) error {
	logger = logger.With(
		slog.String("subcomponent", "batch sender"),
		slog.String("kafka", host),
		slog.String("topic", topic),
		slog.Int("messages", len(batch)),
	)

	/////
	// Transports are responsible for managing connection pools and other resources,
	// it's generally best to create a few of these and share them across your
	// application.

	sharedTransport := &kafka.Transport{
		SASL: plain.Mechanism{
			Username: "adminplain",
			Password: "admin-secret",
		},
	}

	/////
	conn := kafka.Writer{
		Addr:      kafka.TCP(host),
		Topic:     topic,
		Transport: sharedTransport,
	}
	logger.Debug("Connected to kafka")
	logger.Debug("Sending batch")
	err := conn.WriteMessages(context.Background(), batch...)
	if err != nil {
		return fmt.Errorf("Failed to write messages: %v\n", err)
	}
	logger.Info("Sent batch")

	if err := conn.Close(); err != nil {
		return fmt.Errorf("Failed to close the Kafka connection: %v\n", err)
	}
	logger.Debug("Connection with kafka closed")

	return nil
}
