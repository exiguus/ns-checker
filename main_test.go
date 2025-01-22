package main

import (
	"testing"
)

func TestRunCommand(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantExit int
		skip     bool
	}{
		{
			name:     "no arguments",
			args:     []string{"ns-checker"},
			wantExit: 1,
		},
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
		{
			name:     "listen with custom port",
			args:     []string{"ns-checker", "listen", "48053"},
			wantExit: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skip {
				t.Skip("Skipping test requiring elevated privileges")
			}
			if got := runCommand(tt.args); got != tt.wantExit {
				t.Errorf("runCommand() = %v, want %v", got, tt.wantExit)
			}
		})
	}
}
