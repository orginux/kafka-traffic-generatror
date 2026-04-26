// Package generator provides functionalities to generate and send synthetic Kafka traffic.
package generator

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"math/rand"
	"net"
	"strconv"
	"time"

	"kafka-traffic-generator/internal/config"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/segmentio/kafka-go"
)

// entityPool maps entity name → slice of pre-generated flat field maps.
type entityPool map[string][]map[string]any

// MessageGenerator encapsulates the generation and sending of Kafka messages.
type MessageGenerator struct {
	Config *config.Config
	Logger *slog.Logger
}

// New creates a new instance of MessageGenerator.
func New(cfg *config.Config, logger *slog.Logger) *MessageGenerator {
	return &MessageGenerator{
		Config: cfg,
		Logger: logger.With(slog.String("component", "generator")),
	}
}

// Run generates and sends messages based on the provided configuration.
func (mg *MessageGenerator) Run() error {
	mg.Logger.Info("Generator started")

	pools, err := mg.buildPools()
	if err != nil {
		return err
	}

	duration, err := mg.Config.Topic.ParseDuration()
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", mg.Config.Topic.Duration, err)
	}

	var deadline time.Time
	if duration > 0 {
		deadline = time.Now().Add(duration)
		mg.Logger.Info("Generator will stop after", slog.String("duration", mg.Config.Topic.Duration))
	}

	// Legacy batch mode: when neither rate nor duration is set, use NumMsgs batches.
	if mg.Config.Topic.Rate == nil && duration == 0 {
		return mg.runBatchMode(pools)
	}

	for msgNum := 1; mg.shouldContinue(msgNum, deadline); msgNum++ {
		mg.Logger.Info("Sending message", slog.Int("number", msgNum))

		msg, err := mg.generateMessage(pools)
		if err != nil {
			return err
		}
		if err := mg.sendBatch([]kafka.Message{msg}); err != nil {
			return err
		}

		delayMS := mg.Config.Topic.DelayMS()
		if delayMS > 0 {
			time.Sleep(time.Duration(delayMS) * time.Millisecond)
		}
	}
	return nil
}

// runBatchMode is the original batch-oriented loop (backward compatible).
func (mg *MessageGenerator) runBatchMode(pools entityPool) error {
	for batchNum := 1; batchNum <= mg.Config.Topic.NumBatch || mg.Config.Topic.NumBatch <= 0; batchNum++ {
		mg.Logger.Info("Sending batch", slog.Int("batch number", batchNum))

		batch, err := mg.generateBatch(pools)
		if err != nil {
			return err
		}
		if err := mg.sendBatch(batch); err != nil {
			return err
		}

		mg.handleBatchDelay(batchNum)
	}
	return nil
}

func (mg *MessageGenerator) shouldContinue(n int, deadline time.Time) bool {
	if !deadline.IsZero() && time.Now().After(deadline) {
		return false
	}
	if mg.Config.Topic.NumBatch > 0 && n > mg.Config.Topic.NumBatch {
		return false
	}
	return true
}

// buildPools pre-generates entity pools at startup.
func (mg *MessageGenerator) buildPools() (entityPool, error) {
	pools := make(entityPool)
	for _, e := range mg.Config.Entities {
		objects := make([]map[string]any, e.Count)
		for i := range objects {
			obj, err := generateFieldMap(e.Fields)
			if err != nil {
				return nil, fmt.Errorf("building pool %q: %w", e.Name, err)
			}
			objects[i] = obj
		}
		pools[e.Name] = objects
		mg.Logger.Info("Entity pool built", slog.String("entity", e.Name), slog.Int("count", e.Count))
	}
	return pools, nil
}

// generateMessage produces a single flat JSON message, merging entity fields as needed.
func (mg *MessageGenerator) generateMessage(pools entityPool) (kafka.Message, error) {
	obj := make(map[string]any)
	for _, fc := range mg.Config.Fields {
		switch {
		case fc.Entity != "":
			pool, ok := pools[fc.Entity]
			if !ok || len(pool) == 0 {
				return kafka.Message{}, fmt.Errorf("entity %q not found in pool", fc.Entity)
			}
			entity := pool[rand.Intn(len(pool))]
			for k, v := range entity {
				obj[k] = v
			}

		case fc.Function == "weighted" && len(fc.Values) > 0:
			opts := make([]any, len(fc.Values))
			weights := make([]float32, len(fc.Values))
			for i, v := range fc.Values {
				opts[i] = v.Value
				weights[i] = v.Weight
			}
			result, err := gofakeit.Weighted(opts, weights)
			if err != nil {
				return kafka.Message{}, fmt.Errorf("weighted field %q: %w", fc.Name, err)
			}
			obj[fc.Name] = result

		default:
			val, err := callFaker(fc.Function, fc.Params)
			if err != nil {
				return kafka.Message{}, fmt.Errorf("field %q: %w", fc.Name, err)
			}
			obj[fc.Name] = val
		}
	}

	value, err := json.Marshal(obj)
	if err != nil {
		return kafka.Message{}, err
	}
	return kafka.Message{
		Key:   []byte(strconv.Itoa(rand.Intn(100))),
		Value: value,
	}, nil
}

// generateBatch produces NumMsgs messages for the legacy batch mode.
func (mg *MessageGenerator) generateBatch(pools entityPool) ([]kafka.Message, error) {
	batch := make([]kafka.Message, 0, mg.Config.Topic.NumMsgs)
	for i := 0; i < mg.Config.Topic.NumMsgs; i++ {
		msg, err := mg.generateMessage(pools)
		if err != nil {
			return nil, err
		}
		batch = append(batch, msg)
	}
	mg.Logger.Info("Batch generated", slog.Int("messages in batch", len(batch)))
	return batch, nil
}

// generateFieldMap builds a flat map from a slice of Field configs (used for entity pools).
func generateFieldMap(fields []config.Field) (map[string]any, error) {
	obj := make(map[string]any)
	for _, fc := range fields {
		if fc.Function == "weighted" && len(fc.Values) > 0 {
			opts := make([]any, len(fc.Values))
			weights := make([]float32, len(fc.Values))
			for i, v := range fc.Values {
				opts[i] = v.Value
				weights[i] = v.Weight
			}
			result, err := gofakeit.Weighted(opts, weights)
			if err != nil {
				return nil, fmt.Errorf("weighted field %q: %w", fc.Name, err)
			}
			obj[fc.Name] = result
		} else {
			val, err := callFaker(fc.Function, fc.Params)
			if err != nil {
				return nil, fmt.Errorf("field %q: %w", fc.Name, err)
			}
			obj[fc.Name] = val
		}
	}
	return obj, nil
}

// callFaker invokes a gofakeit function by name with the given params.
func callFaker(function string, params map[string]string) (any, error) {
	info := gofakeit.GetFuncLookup(function)
	if info == nil {
		return nil, fmt.Errorf("unknown gofakeit function %q", function)
	}
	mp := gofakeit.NewMapParams()
	for k, v := range params {
		mp.Add(k, v)
	}
	return info.Generate(gofakeit.GlobalFaker, mp, info)
}

// sendBatch sends a batch of Kafka messages to the configured topic.
func (mg *MessageGenerator) sendBatch(batch []kafka.Message) error {
	transport := &kafka.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).DialContext,
		TLS: mg.caCertPool(),
	}

	acks, err := mg.Config.Kafka.ParseAcks()
	if err != nil {
		return fmt.Errorf("failed to set up Kafka acks %s: %v", mg.Config.Kafka.Acks, err)
	}
	mg.Logger.Info("Acks configuration", slog.String("acks", acks.String()))

	compressionType, err := mg.Config.Kafka.ParseCompression()
	if err != nil {
		return fmt.Errorf("failed to set up Kafka compression")
	}
	mg.Logger.Info("Compression configuration", slog.String("compression", compressionType.String()))

	conn := kafka.Writer{
		Addr:         kafka.TCP(mg.Config.Kafka.Host),
		Topic:        mg.Config.Topic.Name,
		Transport:    transport,
		Compression:  compressionType,
		Async:        true,
		RequiredAcks: acks,
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
	if mg.Config.Kafka.TLS.CaPath == "" || mg.Config.Kafka.TLS.CertPath == "" || mg.Config.Kafka.TLS.KeyPath == "" {
		mg.Logger.Debug("TLS not enabled")
		return nil
	}
	mg.Logger.Debug("TLS enabled")

	caPEM, certPEM, keyPEM, err := mg.Config.Kafka.TLS.Read()
	if err != nil {
		fmt.Println(err)
		return nil
	}

	certificate, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		log.Fatal("Failed to load client certificate", err)
		return nil
	}

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

// handleBatchDelay introduces a delay between batches in legacy batch mode.
func (mg *MessageGenerator) handleBatchDelay(batchNum int) {
	delay := mg.Config.Topic.BatchDelay
	if delay > 0 {
		time.Sleep(time.Duration(delay) * time.Millisecond)
		mg.Logger.Info("Delaying before the next batch", slog.Int("milliseconds", delay))
	}
	if mg.Config.Topic.NumBatch > 0 {
		mg.Logger.Info("Batch progress", slog.Int("batches left", mg.Config.Topic.NumBatch-batchNum))
	}
}
