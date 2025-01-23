package perf

import (
	"testing"
	"time"
)

func TestPerformanceMonitor(t *testing.T) {
	mon := New(100 * time.Millisecond) // Shorter interval for testing

	// Simulate requests
	for i := 0; i < 10; i++ {
		mon.RecordResponseTime(100 * time.Millisecond)
		time.Sleep(10 * time.Millisecond) // Space out requests
	}

	// Wait for metrics to be calculated
	time.Sleep(200 * time.Millisecond)

	stats := mon.GetStats()
	if stats.RequestRate <= 0 {
		t.Error("expected non-zero request rate")
	}

	if stats.AvgResponseTime != 100*time.Millisecond {
		t.Errorf("average response time = %v, want %v", stats.AvgResponseTime, 100*time.Millisecond)
	}
}
