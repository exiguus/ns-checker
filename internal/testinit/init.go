package testinit

import (
	"os"
	"sync"
	"testing"
)

var (
	setupOnce   sync.Once
	initialized bool
)

// Initialize testing package before any other initialization
var _ = func() bool {
	testing.Init()
	return true
}()

func init() {
	setupOnce.Do(func() {
		// Just set environment variables, let testing package handle flags
		os.Setenv("TEST_LOG_FILE", "test.log")
		initialized = true
	})
}

// Init ensures initialization is called
func Init() {
	// No-op, just ensures init() is called
}
