package testbase

import (
	"testing"
)

// Setup performs any necessary test setup
func Setup() {
	// We'll let the testing package handle all flags
}

// Run handles common test setup and execution
func Run(m *testing.M) int {
	Setup()
	return m.Run()
}
