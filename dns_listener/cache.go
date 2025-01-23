package dns_listener

import (
	"sync"
	"time"
)

type cacheEntry struct {
	data    []byte
	expires time.Time
}

type Cache struct {
	mu      sync.RWMutex
	entries map[string]cacheEntry
	size    int
}

func NewCache(size int) *Cache {
	return &Cache{
		entries: make(map[string]cacheEntry, size),
		size:    size,
	}
}

func (c *Cache) Set(key string, value []byte, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = cacheEntry{
		data:    value,
		expires: time.Now().Add(ttl),
	}
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(entry.expires) {
		delete(c.entries, key)
		return nil, false
	}

	return entry.data, true
}

func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	count := 0
	now := time.Now()
	for _, entry := range c.entries {
		if !now.After(entry.expires) {
			count++
		}
	}
	return count
}

func (c *Cache) Cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.entries {
		if now.After(entry.expires) {
			delete(c.entries, key)
		}
	}
}
