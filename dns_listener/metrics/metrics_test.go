package metrics

import (
	"testing"
	"time"
)

func TestMetrics(t *testing.T) {
	m := New(100)

	// Record some metrics
	m.RecordRequest()
	m.RecordRequest()
	m.RecordCacheHit()
	m.RecordCacheMiss()
	m.RecordError()

	// Get stats
	stats := m.GetStats()

	// Check counts using uint64 type assertions
	if requests := stats["total_requests"].(uint64); requests != 2 {
		t.Errorf("Expected 2 requests, got %d", requests)
	}
	if hits := stats["cache_hits"].(uint64); hits != 1 {
		t.Errorf("Expected 1 cache hit, got %d", hits)
	}
	if misses := stats["cache_misses"].(uint64); misses != 1 {
		t.Errorf("Expected 1 cache miss, got %d", misses)
	}
	if errors := stats["errors"].(uint64); errors != 1 {
		t.Errorf("Expected 1 error, got %d", errors)
	}

	// Check last request time
	if _, ok := stats["last_request"].(time.Time); !ok {
		t.Error("Last request time should be a time.Time value")
	}
}
