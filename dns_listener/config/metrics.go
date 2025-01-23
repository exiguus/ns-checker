package config

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Metrics collects configuration-related metrics
type Metrics struct {
	ConfigLoads        uint64
	ConfigLoadErrors   uint64
	ValidationErrors   uint64
	LastLoadTime       time.Time
	LastValidationTime time.Time
	ErrorsByField      sync.Map
}

var (
	metrics = &Metrics{
		ErrorsByField: sync.Map{},
	}
)

// RecordConfigLoad records metrics for configuration loading
func RecordConfigLoad(err error) {
	atomic.AddUint64(&metrics.ConfigLoads, 1)
	if err != nil {
		atomic.AddUint64(&metrics.ConfigLoadErrors, 1)
	}
	metrics.LastLoadTime = time.Now()
}

// RecordValidation records metrics for configuration validation
func RecordValidation(err error) {
	if err != nil {
		atomic.AddUint64(&metrics.ValidationErrors, 1)
		if valErr, ok := err.(*ValidationError); ok {
			for _, e := range valErr.Errors {
				if configErr, ok := e.(*ConfigError); ok {
					// Use LoadOrStore to ensure atomic operation
					currentVal, _ := metrics.ErrorsByField.LoadOrStore(configErr.Field, uint64(0))
					metrics.ErrorsByField.Store(configErr.Field, currentVal.(uint64)+1)
				}
			}
		}
	}
	metrics.LastValidationTime = time.Now()
}

// GetMetrics returns the current metrics
func GetMetrics() *Metrics {
	m := &Metrics{
		ConfigLoads:        atomic.LoadUint64(&metrics.ConfigLoads),
		ConfigLoadErrors:   atomic.LoadUint64(&metrics.ConfigLoadErrors),
		ValidationErrors:   atomic.LoadUint64(&metrics.ValidationErrors),
		LastLoadTime:       metrics.LastLoadTime,
		LastValidationTime: metrics.LastValidationTime,
		ErrorsByField:      sync.Map{},
	}

	// Copy the error counts
	metrics.ErrorsByField.Range(func(key, value interface{}) bool {
		m.ErrorsByField.Store(key, value)
		return true
	})

	return m
}

// ResetMetrics resets all metrics to their zero values
func ResetMetrics() {
	atomic.StoreUint64(&metrics.ConfigLoads, 0)
	atomic.StoreUint64(&metrics.ConfigLoadErrors, 0)
	atomic.StoreUint64(&metrics.ValidationErrors, 0)
	metrics.LastLoadTime = time.Time{}
	metrics.LastValidationTime = time.Time{}
	metrics.ErrorsByField = sync.Map{}
}

// ConfigLogger defines the interface for configuration logging
type ConfigLogger interface {
	LogConfigLoad(cfg *Config, source string, err error)
	LogConfigValidation(cfg *Config, err error)
}

// MetricsRecorder defines the interface for recording metrics
type MetricsRecorder interface {
	RecordConfigLoad(err error)
	RecordValidation(err error)
}

// DefaultMetricsRecorder implements MetricsRecorder
type DefaultMetricsRecorder struct{}

func NewMetricsRecorder() MetricsRecorder {
	return &DefaultMetricsRecorder{}
}

// RecordConfigLoad records metrics about configuration loading
func (m *DefaultMetricsRecorder) RecordConfigLoad(err error) {
	atomic.AddUint64(&metrics.ConfigLoads, 1)
	if err != nil {
		atomic.AddUint64(&metrics.ConfigLoadErrors, 1)
	}
	metrics.LastLoadTime = time.Now()
}

func (m *DefaultMetricsRecorder) RecordValidation(err error) {
	if err != nil {
		atomic.AddUint64(&metrics.ValidationErrors, 1)
		if valErr, ok := err.(*ValidationError); ok {
			for _, e := range valErr.Errors {
				if configErr, ok := e.(*ConfigError); ok {
					// Use LoadOrStore to ensure atomic operation
					currentVal, _ := metrics.ErrorsByField.LoadOrStore(configErr.Field, uint64(0))
					metrics.ErrorsByField.Store(configErr.Field, currentVal.(uint64)+1)
				}
			}
		}
	}
	metrics.LastValidationTime = time.Now()
}

// DefaultConfigLogger implements ConfigLogger
type DefaultConfigLogger struct{}

func NewConfigLogger() ConfigLogger {
	return &DefaultConfigLogger{}
}

// LogConfigLoad logs configuration loading events
func (l *DefaultConfigLogger) LogConfigLoad(cfg *Config, source string, err error) {
	level := "INFO"
	if err != nil {
		level = "ERROR"
	}
	fmt.Printf("[%s] ConfigLoad source=%s config=%+v error=%v\n", level, source, cfg, err)
}

// LogConfigValidation logs configuration validation events
func (l *DefaultConfigLogger) LogConfigValidation(cfg *Config, err error) {
	level := "INFO"
	if err != nil {
		level = "ERROR"
	}
	fmt.Printf("[%s] ConfigValidation config=%+v error=%v\n", level, cfg, err)
}

// Add mock logger for testing
type MockConfigLogger struct {
	LoadCalls       int
	ValidateCalls   int
	LastLoadErr     error
	LastValidateErr error
}

func NewMockConfigLogger() *MockConfigLogger {
	return &MockConfigLogger{}
}

func (m *MockConfigLogger) LogConfigLoad(cfg *Config, source string, err error) {
	m.LoadCalls++
	m.LastLoadErr = err
}

func (m *MockConfigLogger) LogConfigValidation(cfg *Config, err error) {
	m.ValidateCalls++
	m.LastValidateErr = err
}
