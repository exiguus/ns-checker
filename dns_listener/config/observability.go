package config

import (
	"time"
)

// ConfigEvent represents a configuration-related event
type ConfigEvent struct {
	EventType string
	Source    string
	Timestamp time.Time
	Config    *Config
	Error     error
}

// Logger defines the interface for logging configuration events
type Logger interface {
	Info(event *ConfigEvent)
	Error(event *ConfigEvent)
}

var globalLogger Logger = &defaultLogger{}

// SetLogger allows setting a custom logger implementation
func SetLogger(l Logger) {
	globalLogger = l
}

// defaultLogger implements Logger interface
type defaultLogger struct{}

func (l *defaultLogger) Info(event *ConfigEvent) {
	// Default implementation uses the existing log functionality
	if event.EventType == "ConfigLoad" {
		NewConfigLogger().LogConfigLoad(event.Config, event.Source, event.Error)
	} else {
		NewConfigLogger().LogConfigValidation(event.Config, event.Error)
	}
}

func (l *defaultLogger) Error(event *ConfigEvent) {
	// Default implementation uses the existing log functionality
	if event.EventType == "ConfigLoad" {
		NewConfigLogger().LogConfigLoad(event.Config, event.Source, event.Error)
	} else {
		NewConfigLogger().LogConfigValidation(event.Config, event.Error)
	}
}

func LogConfigLoad(cfg *Config, source string, err error) {
	event := &ConfigEvent{
		EventType: "ConfigLoad",
		Source:    source,
		Timestamp: time.Now(),
		Config:    cfg,
		Error:     err,
	}
	if err != nil {
		globalLogger.Error(event)
	} else {
		globalLogger.Info(event)
	}
}

func LogConfigValidation(cfg *Config, err error) {
	event := &ConfigEvent{
		EventType: "ConfigValidation",
		Source:    "validator",
		Timestamp: time.Now(),
		Config:    cfg,
		Error:     err,
	}
	if err != nil {
		globalLogger.Error(event)
	} else {
		globalLogger.Info(event)
	}
}
