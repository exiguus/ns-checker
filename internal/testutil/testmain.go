package testutil

import (
	"os"
	"path/filepath"
	"sync"
	"testing"

	_ "github.com/exiguus/ns-checker/internal/testinit"
)

var (
	setupOnce sync.Once
	logPath   = "testdata"
)

// SetupTests sets up the test environment
func SetupTests() {
	setupOnce.Do(func() {
		// Set default log path if not set
		if os.Getenv("LOG_PATH") == "" {
			logPath = filepath.Join(".", logPath)
			os.Setenv("LOG_PATH", logPath)
		}

		if err := os.MkdirAll(os.Getenv("LOG_PATH"), 0755); err != nil {
			panic("Failed to create log directory: " + err.Error())
		}
	})
}

// Run executes the test suite
func Run(m *testing.M) int {
	SetupTests()
	return m.Run()
}
