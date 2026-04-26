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
