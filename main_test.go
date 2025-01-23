package main

import (
	"os"
	"testing"

	_ "github.com/exiguus/ns-checker/internal/testinit"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func TestRunCommand(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tests := []struct {
		name     string
		args     []string
		wantExit int
	}{
		{
			name:     "help command",
			args:     []string{"ns-checker", "help"},
			wantExit: 0,
		},
		{
			name:     "invalid command",
			args:     []string{"ns-checker", "invalid"},
			wantExit: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := runCommand(tt.args); got != tt.wantExit {
				t.Errorf("runCommand() = %v, want %v", got, tt.wantExit)
			}
		})
	}
}
