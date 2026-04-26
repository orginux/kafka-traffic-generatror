package generator

import (
	"encoding/json"
	"log/slog"
	"os"
	"testing"

	"kafka-traffic-generator/internal/config"
)

func newTestGenerator(cfg *config.Config) *MessageGenerator {
	return New(cfg, slog.New(slog.NewTextHandler(os.Stdout, nil)))
}

func emptyPools() entityPool { return make(entityPool) }

// --- callFaker ---

func TestCallFaker_KnownFunction(t *testing.T) {
	val, err := callFaker("email", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val == nil {
		t.Error("expected non-nil value")
	}
}

func TestCallFaker_UnknownFunction(t *testing.T) {
	_, err := callFaker("notafunction", nil)
	if err == nil {
		t.Fatal("expected error for unknown function")
	}
}

func TestCallFaker_WithParams(t *testing.T) {
	val, err := callFaker("price", map[string]string{
		"min": "10",
		"max": "100",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val == nil {
		t.Error("expected non-nil value")
	}
}

// --- generateFieldMap ---

func TestGenerateFieldMap_Valid(t *testing.T) {
	fields := []config.Field{
		{Name: "Email", Function: "email"},
		{Name: "Name", Function: "firstname"},
	}
	obj, err := generateFieldMap(fields)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(obj) != 2 {
		t.Errorf("got %d fields, want 2", len(obj))
	}
	if _, ok := obj["Email"]; !ok {
		t.Error("missing Email field")
	}
}

func TestGenerateFieldMap_Weighted(t *testing.T) {
	fields := []config.Field{
		{
			Name:     "status",
			Function: "weighted",
			Values: []config.WeightedValue{
				{Value: "active", Weight: 80},
				{Value: "inactive", Weight: 20},
			},
		},
	}
	obj, err := generateFieldMap(fields)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	v, ok := obj["status"]
	if !ok {
		t.Fatal("missing status field")
	}
	s, ok := v.(string)
	if !ok {
		t.Fatalf("expected string, got %T", v)
	}
	if s != "active" && s != "inactive" {
		t.Errorf("unexpected value %q", s)
	}
}

// --- buildPools ---

func TestBuildPools_Empty(t *testing.T) {
	mg := newTestGenerator(&config.Config{})
	pools, err := mg.buildPools()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pools) != 0 {
		t.Errorf("expected empty pools, got %d", len(pools))
	}
}

func TestBuildPools_CreatesEntries(t *testing.T) {
	cfg := &config.Config{
		Entities: []config.Entity{
			{
				Name:  "user",
				Count: 5,
				Fields: []config.Field{
					{Name: "user_id", Function: "uuid"},
					{Name: "user_email", Function: "email"},
				},
			},
		},
	}
	mg := newTestGenerator(cfg)
	pools, err := mg.buildPools()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	pool, ok := pools["user"]
	if !ok {
		t.Fatal("user pool not found")
	}
	if len(pool) != 5 {
		t.Errorf("got %d entries, want 5", len(pool))
	}
	// Verify all entries have both fields
	for i, entry := range pool {
		if _, ok := entry["user_id"]; !ok {
			t.Errorf("entry %d missing user_id", i)
		}
		if _, ok := entry["user_email"]; !ok {
			t.Errorf("entry %d missing user_email", i)
		}
	}
}

// --- generateMessage ---

func TestGenerateMessage_BasicFields(t *testing.T) {
	cfg := &config.Config{
		Fields: []config.Field{
			{Name: "email", Function: "email"},
			{Name: "name", Function: "firstname"},
		},
	}
	mg := newTestGenerator(cfg)
	msg, err := mg.generateMessage(emptyPools())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msg.Key) == 0 {
		t.Error("empty key")
	}
	var obj map[string]any
	if err := json.Unmarshal(msg.Value, &obj); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if _, ok := obj["email"]; !ok {
		t.Error("missing email field")
	}
}

func TestGenerateMessage_EntityMerge(t *testing.T) {
	pools := entityPool{
		"user": []map[string]any{
			{"user_id": "fixed-id", "user_name": "Test User"},
		},
	}
	cfg := &config.Config{
		Fields: []config.Field{
			{Entity: "user"},
			{Name: "tx_id", Function: "uuid"},
		},
	}
	mg := newTestGenerator(cfg)
	msg, err := mg.generateMessage(pools)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var obj map[string]any
	if err := json.Unmarshal(msg.Value, &obj); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if obj["user_id"] != "fixed-id" {
		t.Errorf("user_id not merged, got %v", obj["user_id"])
	}
	if _, ok := obj["tx_id"]; !ok {
		t.Error("missing tx_id field")
	}
}

func TestGenerateMessage_WeightedField(t *testing.T) {
	cfg := &config.Config{
		Fields: []config.Field{
			{
				Name:     "status",
				Function: "weighted",
				Values: []config.WeightedValue{
					{Value: "ok", Weight: 100},
				},
			},
		},
	}
	mg := newTestGenerator(cfg)
	msg, err := mg.generateMessage(emptyPools())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var obj map[string]any
	if err := json.Unmarshal(msg.Value, &obj); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if obj["status"] != "ok" {
		t.Errorf("expected status=ok, got %v", obj["status"])
	}
}

func TestGenerateMessage_UnknownEntity(t *testing.T) {
	cfg := &config.Config{
		Fields: []config.Field{
			{Entity: "missing"},
		},
	}
	mg := newTestGenerator(cfg)
	_, err := mg.generateMessage(emptyPools())
	if err == nil {
		t.Fatal("expected error for missing entity")
	}
}

// --- generateBatch ---

func TestGenerateBatch(t *testing.T) {
	cfg := &config.Config{
		Topic: config.Topic{NumMsgs: 5},
		Fields: []config.Field{
			{Name: "Email", Function: "email"},
		},
	}
	mg := newTestGenerator(cfg)
	batch, err := mg.generateBatch(emptyPools())
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

// --- TLS and delay ---

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
	mg.handleBatchDelay(1)
}
