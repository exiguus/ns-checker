package dns_listener

import (
	"testing"
	"time"
)

func TestCache(t *testing.T) {
	cache := NewCache(10)

	cache.Set("test1", []byte("data1"), 200*time.Millisecond)

	if data, exists := cache.Get("test1"); !exists {
		t.Error("Entry should exist immediately after setting")
	} else if string(data) != "data1" {
		t.Error("Incorrect data retrieved from cache")
	}

	// Wait for entry to expire
	time.Sleep(250 * time.Millisecond)

	if _, exists := cache.Get("test1"); exists {
		t.Error("Expired entry should not be accessible")
	}
}

func TestCacheSize(t *testing.T) {
	cache := NewCache(10)

	// Add entries with different expiration times
	cache.Set("test1", []byte("data1"), 100*time.Millisecond)
	cache.Set("test2", []byte("data2"), 300*time.Millisecond)
	cache.Set("test3", []byte("data3"), 300*time.Millisecond)

	if size := cache.Size(); size != 3 {
		t.Errorf("Initial cache size should be 3, got %d", size)
	}

	// Wait for first entry to expire
	time.Sleep(200 * time.Millisecond)

	size := cache.Size()
	if size != 2 {
		t.Errorf("Cache should have size 2 after expiration, got %d", size)
	}
}

func TestCacheCleanup(t *testing.T) {
	cache := NewCache(10)

	cache.Set("test1", []byte("data1"), 100*time.Millisecond)
	cache.Set("test2", []byte("data2"), 300*time.Millisecond)

	// Wait for first entry to expire
	time.Sleep(200 * time.Millisecond)

	size := cache.Size()
	if size != 1 {
		t.Errorf("Cache should have 1 entry after expiration, got size %d", size)
	}

	if _, exists := cache.Get("test1"); exists {
		t.Error("Should not be able to get expired entry")
	}

	if _, exists := cache.Get("test2"); !exists {
		t.Error("Should be able to get non-expired entry")
	}
}
