package logger

import (
	"fmt"
	"log"
	"time"
)

// LogLevel represents logging severity
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
)

// String returns the string representation of LogLevel
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// ConfigLogger handles configuration-related logging
type ConfigLogger struct {
	logger *log.Logger
	debug  bool
}

// NewConfigLogger creates a new ConfigLogger instance
func NewConfigLogger(debug bool) *ConfigLogger {
	return &ConfigLogger{
		logger: log.Default(),
		debug:  debug,
	}
}

// LogConfigLoad logs configuration loading events
func (l *ConfigLogger) LogConfigLoad(source string, err error) {
	level := INFO
	if err != nil {
		level = ERROR
	}
	l.log(level, "ConfigLoad", map[string]interface{}{
		"source":    source,
		"timestamp": time.Now(),
		"error":     err,
	})
}

// LogConfigValidation logs configuration validation events
func (l *ConfigLogger) LogConfigValidation(err error) {
	level := INFO
	if err != nil {
		level = ERROR
	}
	l.log(level, "ConfigValidation", map[string]interface{}{
		"timestamp": time.Now(),
		"error":     err,
	})
}

func (l *ConfigLogger) log(level LogLevel, event string, fields map[string]interface{}) {
	if level == DEBUG && !l.debug {
		return
	}

	msg := fmt.Sprintf("[%s] %s", level.String(), event)
	for k, v := range fields {
		if v != nil {
			msg += fmt.Sprintf(" %s=%v", k, v)
		}
	}
	l.logger.Println(msg)
}
