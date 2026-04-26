package generator

import (
	"log/slog"
	"os"
	"testing"

	"kafka-traffic-generator/internal/config"
)

func newTestGenerator(cfg *config.Config) *MessageGenerator {
	return New(cfg, slog.New(slog.NewTextHandler(os.Stdout, nil)))
}

func TestGenerateFields_Empty(t *testing.T) {
	mg := newTestGenerator(&config.Config{})
	_, err := mg.generateFields()
	if err == nil {
		t.Fatal("expected error for empty fields, got nil")
	}
}

func TestGenerateFields_Valid(t *testing.T) {
	cfg := &config.Config{
		Fields: []config.Field{
			{Name: "Email", Function: "email"},
			{Name: "Name", Function: "firstname"},
		},
	}
	mg := newTestGenerator(cfg)
	fields, err := mg.generateFields()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fields) != 2 {
		t.Errorf("got %d fields, want 2", len(fields))
	}
	if fields[0].Name != "Email" {
		t.Errorf("got field name %q, want %q", fields[0].Name, "Email")
	}
}

func TestGenerateFields_WithParams(t *testing.T) {
	cfg := &config.Config{
		Fields: []config.Field{
			{
				Name:     "Date",
				Function: "daterange",
				Params: map[string]string{
					"startdate": "1993-01-01 00:00:00",
					"enddate":   "1993-12-31 00:00:00",
				},
			},
		},
	}
	mg := newTestGenerator(cfg)
	fields, err := mg.generateFields()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fields) != 1 {
		t.Errorf("got %d fields, want 1", len(fields))
	}
}

func TestGenerateBatch(t *testing.T) {
	cfg := &config.Config{
		Topic: config.Topic{NumMsgs: 5},
		Fields: []config.Field{
			{Name: "Email", Function: "email"},
		},
	}
	mg := newTestGenerator(cfg)
	fields, err := mg.generateFields()
	if err != nil {
		t.Fatalf("generateFields: %v", err)
	}

	batch, err := mg.generateBatch(fields)
	if err != nil {
		t.Fatalf("generateBatch: %v", err)
	}
	if len(batch) != 5 {
		t.Errorf("got %d messages, want 5", len(batch))
	}
	for i, msg := range batch {
		if len(msg.Key) == 0 {
			t.Errorf("message %d has empty key", i)
		}
		if len(msg.Value) == 0 {
			t.Errorf("message %d has empty value", i)
		}
	}
}

func TestCaCertPool_NoTLS(t *testing.T) {
	mg := newTestGenerator(&config.Config{})
	if tlsCfg := mg.caCertPool(); tlsCfg != nil {
		t.Errorf("expected nil TLS config when no paths set, got %v", tlsCfg)
	}
}

func TestHandleBatchDelay_NoDelay(t *testing.T) {
	cfg := &config.Config{
		Topic: config.Topic{BatchDelay: 0, NumBatch: 3},
	}
	mg := newTestGenerator(cfg)
	// Should complete instantly without panicking
	mg.handleBatchDelay(1)
}
