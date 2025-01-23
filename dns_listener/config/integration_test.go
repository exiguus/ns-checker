package config

import (
	"os"
	"testing"
	"time"
)

func TestConfigIntegration(t *testing.T) {
	// Clean environment before tests
	cleanEnvironment()
	defer cleanEnvironment()

	// Enable test mode to disable logging
	SetTestMode(true)

	tests := []struct {
		name    string
		setup   func()
		verify  func(*Config) error
		cleanup func()
	}{
		{
			name: "full configuration flow",
			setup: func() {
				os.Setenv("DNS_PORT", "8053")
				os.Setenv("WORKER_COUNT", "8")
				os.Setenv("DEBUG", "true")
				os.MkdirAll("./testdata/logs", 0755)
			},
			verify: func(cfg *Config) error {
				if err := ValidateConfig(cfg); err != nil {
					return err
				}

				// Verify configuration values directly
				if cfg.Port != "8053" || cfg.WorkerCount != 8 || !cfg.Debug {
					t.Error("Configuration values don't match expected values")
				}

				return nil
			},
			cleanup: func() {
				os.RemoveAll("./testdata")
			},
		},
		{
			name: "error handling",
			setup: func() {
				os.Setenv("DNS_PORT", "invalid")
			},
			verify: func(cfg *Config) error {
				err := ValidateConfig(cfg)
				if err == nil {
					t.Error("Expected validation error for invalid port")
				}
				return nil
			},
			cleanup: func() {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			if tt.setup != nil {
				tt.setup()
			}

			// Load and verify configuration
			cfg := LoadFromEnv()
			if err := tt.verify(cfg); err != nil {
				t.Errorf("Verification failed: %v", err)
			}

			// Cleanup
			if tt.cleanup != nil {
				tt.cleanup()
			}
		})
	}
}

func TestPortCheckerIntegration(t *testing.T) {
	// Set up test server
	portChecker := NewPortChecker(time.Second)

	tests := []struct {
		name    string
		setup   func() (cleanup func())
		port    string
		wantErr bool
	}{
		{
			name: "available port",
			port: "48053",
			setup: func() func() {
				return func() {}
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := tt.setup()
			defer cleanup()

			err := portChecker.IsPortAvailable(tt.port)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsPortAvailable() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
