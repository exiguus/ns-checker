package testhelper

import (
	"os"
	"path/filepath"
	"testing"
)

// Setup initializes test environment
func Setup() {
	// Set up log directory
	if os.Getenv("LOG_PATH") == "" {
		logPath := filepath.Join(".", "testdata")
		os.Setenv("LOG_PATH", logPath)
		os.MkdirAll(logPath, 0755)
	}
}

// Run provides a standard way to run tests across all packages
func Run(m *testing.M) int {
	Setup()
	return m.Run()
}
