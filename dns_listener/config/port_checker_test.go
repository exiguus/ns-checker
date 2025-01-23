package config

import (
	"net"
	"testing"
	"time"
)

func TestPortChecker(t *testing.T) {
	// Create a test listener to simulate a port in use
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	// Get the actual port that was assigned
	_, portStr, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}

	pc := NewPortChecker(time.Second)

	tests := []struct {
		name    string
		port    string
		wantErr bool
	}{
		{"valid unused port", "8081", false},
		{"invalid port format", "abc", true},
		{"port out of range", "99999", true},
		{"port in use", portStr, true},
		{"privileged port", "80", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := pc.IsPortAvailable(tt.port)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsPortAvailable() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMockPortChecker(t *testing.T) {
	mock := &MockPortChecker{}

	if err := mock.IsPortAvailable("any"); err != nil {
		t.Error("MockPortChecker.IsPortAvailable() should always return nil")
	}

	if mock.IsPortInUse("any") {
		t.Error("MockPortChecker.IsPortInUse() should always return false")
	}
}

func BenchmarkPortChecker(b *testing.B) {
	pc := NewPortChecker(time.Second)
	b.Run("IsPortAvailable", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			pc.IsPortAvailable("8081")
		}
	})

	b.Run("IsPortInUse", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			pc.IsPortInUse("8081")
		}
	})
}
