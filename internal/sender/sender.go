package sender

import (
	"context"
	"fmt"
	"kafka-traffic-generator/internal/config"
	"kafka-traffic-generator/internal/generator"
	"log/slog"
	"math/rand"
	"strconv"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/segmentio/kafka-go"
)

type Batch struct {
	Messages []kafka.Message
}

// generateBatch generates a batch of Kafka messages with random key-value pairs.
// TODO use config as parameters
func (b *Batch) generate(numMsgs int, fields config.Fields, logger *slog.Logger) ([]kafka.Message, error) {
	logger = logger.With(
		slog.String("subcomponent", "batch generator"),
	)

	// Generate message parameters
	fields, err := generator.GenerateFields(fields, logger)
	if err != nil {
		return err
	}
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
		*b = append(*b, msg)
		logger.Debug("Message generated", slog.Int("msg num", i+1), slog.Int("ftom", numMsgs))
	}
	logger.Info("Batch generated", slog.Int("messages in batch", len(*batch)))
	return *batch, nil
}

// send sends a batch of Kafka messages to the specified topic.
func (b *Batch) send(host, topic string, logger *slog.Logger) error {
	logger = logger.With(
		slog.String("subcomponent", "batch sender"),
		slog.String("kafka", host),
		slog.String("topic", topic),
		slog.Int("messages", len(batch)),
	)
	conn := kafka.Writer{
		Addr:  kafka.TCP(host),
		Topic: topic,
	}
	logger.Debug("Connected to kafka")
	logger.Debug("Sending batch")
	err := conn.WriteMessages(context.Background(), b...)
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

func ProduceTraffic(config config.Config, logger *slog.Logger) error {

	for _, task := range config.Tasks {
		logger.Info("Task", slog.String("name", task.Name))
		var batchNum int
		for batchNum <= task.Topic.NumBatch {
			logger.Info("Batch statistics", slog.String("task", task.Name), slog.Int("batch number", batchNum))
			var batch Batch
			err := batch.generate(task.Topic.NumMsgs, task.Fields, logger)
			if err != nil {
				return err
			}
			if err := batch.send(config.Kafka.Host, task.Topic.Name, logger); err != nil {
				return err
			}

			// Delay before sending the next batch
			if task.Topic.MsgDelay > 0 {
				logger.Info("Delaying before the next batch:", slog.Int("milliseconds", config.Topic.MsgDelay))
				time.Sleep(time.Duration(task.Topic.MsgDelay) * time.Millisecond)
			}

			// If not unlimited number of batches
			if config.Topic.NumBatch > 0 {
				batchNum++
			}
		}
	}
	return nil
}
