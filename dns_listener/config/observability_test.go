package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// MockLogger for testing
type MockLogger struct {
	infoEvents  []*ConfigEvent
	errorEvents []*ConfigEvent
}

func (m *MockLogger) Info(event *ConfigEvent) {
	m.infoEvents = append(m.infoEvents, event)
}

func (m *MockLogger) Error(event *ConfigEvent) {
	m.errorEvents = append(m.errorEvents, event)
}

func TestConfigLogging(t *testing.T) {
	mockLogger := &MockLogger{}
	SetLogger(mockLogger)

	// Create a valid configuration for testing
	cfg := &Config{
		Port:                 "8053", // Use non-privileged port
		WorkerCount:          4,
		RateLimit:            1000,
		RateBurst:            100, // Less than RateLimit
		CacheTTL:             5 * time.Minute,
		CacheCleanupInterval: time.Minute,
		HealthPort:           "8088",
		LogPath:              "./test.log",
		LogMaxSize:           10,
		LogMaxBackups:        3,
		LogMaxAge:            30,
		Debug:                false,
	}

	// Ensure log directory exists before validation
	if err := os.MkdirAll(filepath.Dir(cfg.LogPath), 0755); err != nil {
		t.Fatal("Failed to create test log directory:", err)
	}
	defer os.RemoveAll(filepath.Dir(cfg.LogPath)) // Clean up after test

	// Test successful load
	LogConfigLoad(cfg, "test", nil)
	if len(mockLogger.infoEvents) != 1 {
		t.Error("Expected info event for successful load")
	}

	// Test failed load
	testErr := errors.New("test error")
	LogConfigLoad(cfg, "test", testErr)
	if len(mockLogger.errorEvents) != 1 {
		t.Error("Expected error event for failed load")
	}

	// Test validation logging
	mockLogger = &MockLogger{}
	SetLogger(mockLogger)

	ValidateConfig(cfg)

	// Verify event contents
	if len(mockLogger.infoEvents) > 0 {
		event := mockLogger.infoEvents[0]
		if event.EventType != "ConfigValidation" {
			t.Errorf("Expected event type 'ConfigValidation', got %s", event.EventType)
		}
		if event.Source != "validator" {
			t.Errorf("Expected source 'validator', got %s", event.Source)
		}
	}
}
