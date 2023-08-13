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

	"kafka-traffic-generator/internal/config"
)

// Run generates and sends batches of Kafka messages based onthe provided configuration.
func Run(config config.Config, logger *slog.Logger) error {
	logger = logger.With(
		slog.String("component", "generator"),
	)
	logger.Info("Sarting generator")

	// Generate message parameters
	fields, err := generateFields(config.Fields, logger)
	if err != nil {
		return err
	}

	// Generate and send batches of messages
	var batchNum int
	for batchNum <= config.Topic.NumBatch {
		logger.Info("Batches statistics", slog.Int("batch num", batchNum))
		batch, err := generateBatch(config.Topic.NumMsgs, fields, logger)
		if err != nil {
			return err
		}

		if err := sendBatch(config.Kafka.Host, config.Topic.Name, batch, logger); err != nil {
			return err
		}

		// Delay before sending the next batch
		logger.Info("Delaying before the next batch:", slog.Int("ms", config.Topic.MsgDelay))
		time.Sleep(time.Duration(config.Topic.MsgDelay) * time.Millisecond)

		// If NumBatch =< 0 (default) -- Unlimited number of batches
		if config.Topic.NumBatch > 0 {
			batchesLeft := config.Topic.NumBatch - batchNum
			logger.Info("Batches statistics", slog.Int("left:", batchesLeft))
			batchNum++
		}
	}
	return nil
}

// generateFields generates a slice of gofakeit.Field based on the provided field configurations.
func generateFields(fieldConfigs []config.Field, logger *slog.Logger) ([]gofakeit.Field, error) {
	if len(fieldConfigs) == 0 {
		logger.Debug("No fields to generate")
		return nil, fmt.Errorf("Fields are not defined in config file")
	}

	var fields []gofakeit.Field
	for _, fieldConfig := range fieldConfigs {
		logger.Debug("Preparing fake data", slog.String("field", fieldConfig.Name), slog.String("function", fieldConfig.Function))
		params := gofakeit.NewMapParams()
		for key, value := range fieldConfig.Params {
			logger.Debug(
				"log",
				slog.String("Field", fieldConfig.Name),
				slog.String("Function", fieldConfig.Function),
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
		logger.Debug("Generate batch", slog.String("kafka msg key", key))

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
	logger.Debug("Batch generated", slog.Int("msg in batch", len(batch)))
	return batch, nil
}

// sendBatch sends a batch of Kafka messages to the specified topic.
func sendBatch(host, topic string, batch []kafka.Message, logger *slog.Logger) error {
	conn := kafka.Writer{
		Addr:  kafka.TCP(host),
		Topic: topic,
	}

	err := conn.WriteMessages(context.Background(), batch...)
	if err != nil {
		return fmt.Errorf("Failed to write messages: %v\n", err)
	}
	logger.Info("Sent batch", slog.Int("messages", len(batch)))

	if err := conn.Close(); err != nil {
		return fmt.Errorf("Failed to close the Kafka connection: %v\n", err)
	}

	return nil
}
