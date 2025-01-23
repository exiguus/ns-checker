package config

import (
	"errors"
	"fmt"
	"testing"
)

func TestMetrics(t *testing.T) {
	// Reset metrics before testing
	ResetMetrics()

	tests := []struct {
		name          string
		action        func()
		checkMetrics  func(m *Metrics) bool
		expectedError bool
	}{
		{
			name: "config load success",
			action: func() {
				RecordConfigLoad(nil)
			},
			checkMetrics: func(m *Metrics) bool {
				return m.ConfigLoads == 1 && m.ConfigLoadErrors == 0
			},
		},
		{
			name: "config load error",
			action: func() {
				RecordConfigLoad(errors.New("test error"))
			},
			checkMetrics: func(m *Metrics) bool {
				return m.ConfigLoads == 1 && m.ConfigLoadErrors == 1
			},
		},
		{
			name: "validation error",
			action: func() {
				err := &ValidationError{
					Errors: []error{
						NewConfigError("Port", "invalid", "test error"),
					},
				}
				RecordValidation(err)
			},
			checkMetrics: func(m *Metrics) bool {
				value, exists := m.ErrorsByField.Load("Port")
				return m.ValidationErrors == 1 && exists && value.(uint64) == 1
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ResetMetrics()
			tt.action()
			m := GetMetrics()
			if !tt.checkMetrics(m) {
				t.Errorf("Metrics check failed for %s", tt.name)
			}
		})
	}
}

func BenchmarkMetrics(b *testing.B) {
	b.Run("RecordConfigLoad", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			RecordConfigLoad(nil)
		}
	})

	b.Run("RecordValidation", func(b *testing.B) {
		err := &ValidationError{
			Errors: []error{
				NewConfigError("Port", "invalid", "test error"),
			},
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			RecordValidation(err)
		}
	})
}

func TestConfigLogger(t *testing.T) {
	mockLogger := NewMockConfigLogger()
	cfg := &Config{Port: "8053"}
	testErr := fmt.Errorf("test error")

	tests := []struct {
		name     string
		action   func()
		checkLog func() bool
	}{
		{
			name: "log successful config load",
			action: func() {
				mockLogger.LogConfigLoad(cfg, "test", nil)
			},
			checkLog: func() bool {
				return mockLogger.LoadCalls == 1 && mockLogger.LastLoadErr == nil
			},
		},
		{
			name: "log config load error",
			action: func() {
				mockLogger.LogConfigLoad(cfg, "test", testErr)
			},
			checkLog: func() bool {
				return mockLogger.LoadCalls == 2 && mockLogger.LastLoadErr == testErr
			},
		},
		{
			name: "log successful validation",
			action: func() {
				mockLogger.LogConfigValidation(cfg, nil)
			},
			checkLog: func() bool {
				return mockLogger.ValidateCalls == 1 && mockLogger.LastValidateErr == nil
			},
		},
		{
			name: "log validation error",
			action: func() {
				mockLogger.LogConfigValidation(cfg, testErr)
			},
			checkLog: func() bool {
				return mockLogger.ValidateCalls == 2 && mockLogger.LastValidateErr == testErr
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.action()
			if !tt.checkLog() {
				t.Errorf("Logger check failed for %s", tt.name)
			}
		})
	}
}
