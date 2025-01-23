package cache

import (
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	cfg := DefaultConfig()
	cfg.MaxSize = 1024
	cfg.DefaultTTL = time.Minute

	c := NewLRU(cfg)

	// Add test data
	testData := []byte("test value")
	c.Set("test", testData, time.Minute)

	// Verify value was stored
	if v, ok := c.Get("test"); !ok || string(v) != string(testData) {
		t.Errorf("Get() = %v, %v, want %v, true", string(v), ok, string(testData))
	}

	// Verify stats
	stats := c.Stats()
	if stats.Size != 1 {
		t.Errorf("Stats().Size = %d, want 1", stats.Size)
	}
	if stats.Hits != 1 {
		t.Errorf("Stats().Hits = %d, want 1", stats.Hits)
	}
}
