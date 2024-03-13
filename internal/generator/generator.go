// Package generator provides functionalities to generate and send synthetic Kafka traffic.
package generator

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"log/slog"
	"math/rand"
	"net"
	"strconv"
	"time"

	"kafka-traffic-generator/internal/config"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/compress"
)

// MessageGenerator encapsulates the generation and sending of Kafka messages.
type MessageGenerator struct {
	Config *config.Config
	Logger *slog.Logger
}

// NewMessageGenerator creates a new instance of MessageGenerator.
func New(cfg *config.Config, logger *slog.Logger) *MessageGenerator {
	return &MessageGenerator{
		Config: cfg,
		Logger: logger.With(slog.String("component", "generator")),
	}
}

// Run generates and sends batches of Kafka messages based on the provided configuration.
func (mg *MessageGenerator) Run() error {
	mg.Logger.Info("Generator started")
	fields, err := mg.generateFields()
	if err != nil {
		return err
	}

	for batchNum := 1; batchNum <= mg.Config.Topic.NumBatch || mg.Config.Topic.NumBatch <= 0; batchNum++ {
		mg.Logger.Info("Sending batch", slog.Int("batch number", batchNum))

		if err := mg.processBatch(fields); err != nil {
			return err
		}

		// Delay between batches
		mg.handleBatchDelay(batchNum)
	}
	return nil
}

// generateFields generates a slice of gofakeit.Field based on the provided field configurations.
func (mg *MessageGenerator) generateFields() ([]gofakeit.Field, error) {
	if len(mg.Config.Fields) == 0 {
		return nil, fmt.Errorf("no fields defined in the configuration")
	}

	var fields []gofakeit.Field
	for _, fc := range mg.Config.Fields {
		params := gofakeit.NewMapParams()
		for key, value := range fc.Params {
			params.Add(key, value)
		}
		fields = append(fields, gofakeit.Field{
			Name:     fc.Name,
			Function: fc.Function,
			Params:   *params,
		})
	}
	mg.Logger.Debug("Fields generated")
	return fields, nil
}

// processBatch generates and sends a single batch of messages.
func (mg *MessageGenerator) processBatch(fields []gofakeit.Field) error {
	batch, err := mg.generateBatch(fields)
	if err != nil {
		return err
	}

	return mg.sendBatch(batch)
}

// generateBatch generates a batch of Kafka messages with random key-value pairs.
func (mg *MessageGenerator) generateBatch(fields []gofakeit.Field) ([]kafka.Message, error) {
	var batch []kafka.Message
	for i := 0; i < mg.Config.Topic.NumMsgs; i++ {
		key := strconv.Itoa(rand.Intn(100))
		value, err := gofakeit.JSON(&gofakeit.JSONOptions{
			Type:   "object",
			Fields: fields,
			Indent: false,
		})
		if err != nil {
			return nil, fmt.Errorf("error generating message: %v", err)
		}

		batch = append(batch, kafka.Message{
			Key:   []byte(key),
			Value: []byte(value),
		})
	}
	mg.Logger.Info("Batch generated", slog.Int("messages in batch", len(batch)))
	return batch, nil
}

// sendBatch sends a batch of Kafka messages to the specified topic.
func (mg *MessageGenerator) sendBatch(batch []kafka.Message) error {
	transport := &kafka.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).DialContext,
		TLS: mg.caCertPool(),
	}

	conn := kafka.Writer{
		Addr:         kafka.TCP(mg.Config.Kafka.Host),
		Topic:        mg.Config.Topic.Name,
		Transport:    transport,
		Compression:  compress.Gzip,
		Async:        true,
		RequiredAcks: kafka.RequireNone,
	}

	if err := conn.WriteMessages(context.Background(), batch...); err != nil {
		return fmt.Errorf("failed to write messages: %v", err)
	}
	if err := conn.Close(); err != nil {
		return fmt.Errorf("failed to close Kafka connection: %v", err)
	}
	mg.Logger.Info("Batch sent successfully")
	return nil
}

// caCertPool initializes the CA certificate pool for TLS connections.
func (mg *MessageGenerator) caCertPool() *tls.Config {
	// Skip TLS configuration if not enabled.
	if mg.Config.Kafka.TLS.CaPath == "" || mg.Config.Kafka.TLS.CertPath == "" || mg.Config.Kafka.TLS.KeyPath == "" {
		mg.Logger.Debug("TLS not enabled")
		return nil
	}

	mg.Logger.Debug("TLS enabled")

	// Read CA, certificate, and key files.
	caPEM, certPEM, keyPEM, err := mg.Config.Kafka.TLS.Read()
	if err != nil {
		fmt.Println(err)
		return nil
	}

	// Define TLS configuration
	certificate, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		log.Fatal("Failed to load client certificate", err)
		return nil
	}

	// Create CA certificate pool
	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM(caPEM); !ok {
		log.Fatal("Failed to append CA certificate to pool")
	}

	return &tls.Config{
		Certificates:       []tls.Certificate{certificate},
		RootCAs:            caCertPool,
		InsecureSkipVerify: true,
	}
}

// handleBatchDelay introduces a delay between message batches if configured.
func (mg *MessageGenerator) handleBatchDelay(batchNum int) {
	if mg.Config.Topic.BatchDelay > 0 {
		time.Sleep(time.Duration(mg.Config.Topic.BatchDelay) * time.Millisecond)
		mg.Logger.Info("Delaying before the next batch", slog.Int("milliseconds", mg.Config.Topic.BatchDelay))
	}

	if mg.Config.Topic.NumBatch > 0 {
		batchesLeft := mg.Config.Topic.NumBatch - batchNum
		mg.Logger.Info("Batch progress", slog.Int("batches left", batchesLeft))
	}
}
