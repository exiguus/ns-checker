package logger

import (
	"bytes"
	"errors"
	"log"
	"strings"
	"testing"
)

func TestConfigLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := NewConfigLogger(true)
	logger.logger = log.New(&buf, "", 0) // Use buffer for testing

	tests := []struct {
		name     string
		action   func()
		contains []string
	}{
		{
			name: "log successful config load",
			action: func() {
				logger.LogConfigLoad("environment", nil)
			},
			contains: []string{"[INFO]", "ConfigLoad", "source=environment"},
		},
		{
			name: "log config load error",
			action: func() {
				logger.LogConfigLoad("file", errors.New("test error"))
			},
			contains: []string{"[ERROR]", "ConfigLoad", "source=file", "test error"},
		},
		{
			name: "log successful validation",
			action: func() {
				logger.LogConfigValidation(nil)
			},
			contains: []string{"[INFO]", "ConfigValidation"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.action()
			output := buf.String()

			for _, want := range tt.contains {
				if !strings.Contains(output, want) {
					t.Errorf("log output missing %q, got: %s", want, output)
				}
			}
		})
	}
}
