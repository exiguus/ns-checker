package testflags

import (
	"flag"

	// Ensure testing is initialized first
	_ "github.com/exiguus/ns-checker/internal/testinit"
)

var (
	// LogFile is the path to the test log file
	LogFile = flag.String("testlogfile", "", "test log file path")
)

// Setup initializes test flags
func Setup() {
	if !flag.Parsed() {
		flag.Parse()
	}
}

// GetLogFile returns the current log file path
func GetLogFile() string {
	if LogFile == nil {
		return ""
	}
	return *LogFile
}
