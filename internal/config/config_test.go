package config

import (
	"testing"

	"github.com/segmentio/kafka-go"
)

func TestParseAcks(t *testing.T) {
	tests := []struct {
		acks    string
		want    kafka.RequiredAcks
		wantErr bool
	}{
		{"all", kafka.RequireAll, false},
		{"one", kafka.RequireOne, false},
		{"none", kafka.RequireNone, false},
		{"invalid", kafka.RequireNone, true},
		{"", kafka.RequireNone, true},
	}

	for _, tt := range tests {
		k := &Kafka{Acks: tt.acks}
		got, err := k.ParseAcks()
		if (err != nil) != tt.wantErr {
			t.Errorf("ParseAcks(%q) error = %v, wantErr %v", tt.acks, err, tt.wantErr)
		}
		if !tt.wantErr && got != tt.want {
			t.Errorf("ParseAcks(%q) = %v, want %v", tt.acks, got, tt.want)
		}
	}
}

func TestDelayMS_FallbackToBatchDelay(t *testing.T) {
	topic := &Topic{BatchDelay: 500}
	if got := topic.DelayMS(); got != 500 {
		t.Errorf("got %d, want 500", got)
	}
}

func TestDelayMS_WithRate(t *testing.T) {
	topic := &Topic{Rate: &Rate{Min: 10, Max: 15}}
	got := topic.DelayMS()
	// 10/min → 6000ms, 15/min → 4000ms
	if got < 4000 || got > 6000 {
		t.Errorf("DelayMS() = %d, want between 4000 and 6000", got)
	}
}

func TestDelayMS_EqualMinMax(t *testing.T) {
	topic := &Topic{Rate: &Rate{Min: 12, Max: 12}}
	got := topic.DelayMS()
	if got != 5000 {
		t.Errorf("DelayMS() = %d, want 5000", got)
	}
}

func TestParseDuration_Empty(t *testing.T) {
	topic := &Topic{}
	d, err := topic.ParseDuration()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d != 0 {
		t.Errorf("got %v, want 0", d)
	}
}

func TestParseDuration_Valid(t *testing.T) {
	topic := &Topic{Duration: "10m"}
	d, err := topic.ParseDuration()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.Minutes() != 10 {
		t.Errorf("got %v, want 10m", d)
	}
}

func TestParseDuration_Invalid(t *testing.T) {
	topic := &Topic{Duration: "badvalue"}
	_, err := topic.ParseDuration()
	if err == nil {
		t.Fatal("expected error for invalid duration")
	}
}

func TestParseCompression(t *testing.T) {
	tests := []struct {
		compression string
		want        kafka.Compression
		wantErr     bool
	}{
		{"", kafka.Compression(0), false},
		{"gzip", kafka.Gzip, false},
		{"snappy", kafka.Snappy, false},
		{"lz4", kafka.Lz4, false},
		{"zstd", kafka.Zstd, false},
		{"invalid", kafka.Compression(0), true},
	}

	for _, tt := range tests {
		k := &Kafka{Compression: tt.compression}
		got, err := k.ParseCompression()
		if (err != nil) != tt.wantErr {
			t.Errorf("ParseCompression(%q) error = %v, wantErr %v", tt.compression, err, tt.wantErr)
		}
		if !tt.wantErr && got != tt.want {
			t.Errorf("ParseCompression(%q) = %v, want %v", tt.compression, got, tt.want)
		}
	}
}
