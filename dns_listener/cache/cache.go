package cache

import (
	"sync"
	"sync/atomic"
	"time"
)

type basicCacheItem struct {
	value      []byte
	expiration time.Time
	size       int64
	hits       int64
}

type BasicCache struct {
	mu              sync.RWMutex
	items           map[string]*basicCacheItem
	maxSize         int64
	currentSize     int64
	defaultTTL      time.Duration
	cleanupInterval time.Duration
	stats           Stats
	evictions       uint64
}

func New(cfg Config) Cache {
	c := &BasicCache{
		items:           make(map[string]*basicCacheItem),
		maxSize:         cfg.MaxSize,
		defaultTTL:      cfg.DefaultTTL,
		cleanupInterval: cfg.CleanupInterval,
	}

	if cfg.CleanupInterval > 0 {
		go c.startCleanup()
	}

	return c
}

func (c *BasicCache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists || time.Now().After(item.expiration) {
		c.stats.Misses++
		return nil, false
	}

	c.stats.Hits++
	item.hits++
	return item.value, true
}

func (c *BasicCache) Set(key string, value []byte, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ttl <= 0 {
		ttl = c.defaultTTL
	}

	size := int64(len(value))
	if oldItem, exists := c.items[key]; exists {
		c.currentSize -= oldItem.size
	}
	c.currentSize += size

	// If we're at capacity, evict the oldest entry
	if int64(len(c.items)) >= c.maxSize {
		var oldestKey string
		var oldestTime time.Time
		for k, v := range c.items {
			if oldestTime.IsZero() || v.expiration.Before(oldestTime) {
				oldestTime = v.expiration
				oldestKey = k
			}
		}
		if oldestKey != "" {
			delete(c.items, oldestKey)
			atomic.AddUint64(&c.evictions, 1)
			atomic.AddInt64(&c.stats.Evictions, 1)
		}
	}

	c.items[key] = &basicCacheItem{
		value:      value,
		expiration: time.Now().Add(ttl),
		size:       size,
		hits:       0,
	}

	c.cleanup()
}

func (c *BasicCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if item, exists := c.items[key]; exists {
		c.currentSize -= item.size
		delete(c.items, key)
	}
}

func (c *BasicCache) Stats() Stats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return Stats{
		Size:          len(c.items),
		BytesInMemory: uint64(c.currentSize),
		Hits:          c.stats.Hits,
		Misses:        c.stats.Misses,
		Evictions:     c.stats.Evictions,
	}
}

func (c *BasicCache) Cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cleanup()
}

func (c *BasicCache) cleanup() {
	now := time.Now()
	for key, item := range c.items {
		if now.After(item.expiration) {
			c.currentSize -= item.size
			delete(c.items, key)
			c.stats.Evictions++
			atomic.AddUint64(&c.evictions, 1)
			atomic.AddInt64(&c.stats.Evictions, 1)
		}
	}

	for c.currentSize > c.maxSize {
		c.evictOldest()
	}
}

func (c *BasicCache) startCleanup() {
	ticker := time.NewTicker(c.cleanupInterval)
	for range ticker.C {
		c.Cleanup()
	}
}

func (c *BasicCache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, item := range c.items {
		if oldestTime.IsZero() || item.expiration.Before(oldestTime) {
			oldestKey = key
			oldestTime = item.expiration
		}
	}

	if oldestKey != "" {
		c.Delete(oldestKey)
	}
}
