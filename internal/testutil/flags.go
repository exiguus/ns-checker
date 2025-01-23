package testutil

import (
	"flag"
	"time"
)

var (
	TestLogFile = flag.String("test.testlogfile", "", "Test log file path")
	TestTimeout = flag.Duration("test.timeout", 30*time.Second, "Test timeout duration")
)

func init() {
	// Common test flags
	flag.Bool("test.v", false, "verbose")

	if !flag.Parsed() {
		flag.Parse()
	}
}
