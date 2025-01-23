package cache

import (
	"sync"
	"time"
)

type baseCache struct {
	mu        sync.RWMutex
	entries   map[string]cacheEntry
	hits      int64
	misses    int64
	evictions int64
	config    Config
}

type cacheEntry struct {
	data    []byte
	expires time.Time
}

func (c *baseCache) Stats() Stats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return Stats{
		Size:          len(c.entries),
		BytesInMemory: c.calculateSize(),
		Hits:          c.hits,
		Misses:        c.misses,
		Evictions:     c.evictions,
	}
}

func (c *baseCache) calculateSize() uint64 {
	var total uint64
	for _, entry := range c.entries {
		total += uint64(len(entry.data))
	}
	return total
}
