package dns_listener

import (
	"net"
	"testing"
	"time"

	"github.com/exiguus/ns-checker/dns_listener/config"
	"github.com/exiguus/ns-checker/dns_listener/types"
)

func TestCalculateOptimalWorkers(t *testing.T) {
	workers := calculateOptimalWorkers()
	if workers < 4 {
		t.Errorf("Expected minimum 4 workers, got %d", workers)
	}
	cpuCount := 4 * 4 // max workers should be CPU count * 4
	if workers > cpuCount {
		t.Errorf("Expected maximum %d workers, got %d", cpuCount, workers)
	}
}

func TestParsePort(t *testing.T) {
	tests := []struct {
		name     string
		port     string
		expected int
	}{
		{"Valid port", "53", 53},
		{"Invalid port", "999999", -1},
		{"Non-numeric port", "abc", -1},
		{"Empty port", "", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parsePort(tt.port)
			if result != tt.expected {
				t.Errorf("parsePort(%s) = %d; want %d", tt.port, result, tt.expected)
			}
		})
	}
}

func TestParseDNSQuery(t *testing.T) {
	// Sample DNS query bytes for example.com
	queryBytes := []byte{
		0x00, 0x01, // Transaction ID
		0x01, 0x00, // Flags
		0x00, 0x01, // Questions
		0x00, 0x00, // Answer RRs
		0x00, 0x00, // Authority RRs
		0x00, 0x00, // Additional RRs
		0x07, 'e', 'x', 'a', 'm', 'p', 'l', 'e',
		0x03, 'c', 'o', 'm',
		0x00,       // Root label
		0x00, 0x01, // Type A
		0x00, 0x01, // Class IN
	}

	result := parseDNSQuery(queryBytes)
	if result == "" {
		t.Error("Expected non-empty query parse result")
	}
}

func TestInitializeListener(t *testing.T) {
	cfg := &config.Config{
		Port:                 "53",
		HealthPort:           "8080",
		WorkerCount:          4,
		RateLimit:            100,
		RateBurst:            200,
		CacheTTL:             5 * time.Minute,
		CacheCleanupInterval: 10 * time.Minute,
	}

	isTestMode = true // Enable test mode
	defer func() { isTestMode = false }()

	listener, err := initializeListener(cfg)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if listener == nil {
		t.Error("Expected non-nil listener")
	}
	defer listener.Close()
}

func TestDNSListenerIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	cfg := config.DefaultConfig()
	cfg.Port = "15353" // Use non-privileged port for testing

	listener, err := NewDNSListener(cfg)
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()

	go func() {
		if err := listener.Start(); err != nil {
			t.Errorf("Listener start error: %v", err)
		}
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	// Test sending a DNS query
	conn, err := net.Dial("udp", "127.0.0.1:"+cfg.Port)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Simple DNS query for example.com
	query := []byte{
		0x00, 0x01, // Transaction ID
		0x01, 0x00, // Flags
		0x00, 0x01, // Questions
		0x00, 0x00, // Answer RRs
		0x00, 0x00, // Authority RRs
		0x00, 0x00, // Additional RRs
		0x07, 'e', 'x', 'a', 'm', 'p', 'l', 'e',
		0x03, 'c', 'o', 'm',
		0x00,       // Root label
		0x00, 0x01, // Type A
		0x00, 0x01, // Class IN
	}

	_, err = conn.Write(query)
	if err != nil {
		t.Fatalf("Failed to send query: %v", err)
	}

	// Read response
	response := make([]byte, types.DefaultBufferSize)
	conn.SetReadDeadline(time.Now().Add(time.Second))
	_, err = conn.Read(response)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}
}
