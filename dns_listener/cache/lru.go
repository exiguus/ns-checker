package cache

import (
	"container/list"
	"sync"
	"sync/atomic"
	"time"
)

// LRUCache implements a thread-safe LRU cache
type LRUCache struct {
	mu        sync.RWMutex
	items     map[string]*entry
	evictList *list.List
	config    Config
	stats     struct {
		hits      uint64
		misses    uint64
		evictions uint64
		bytes     int64
		size      int64
	}
}

type entry struct {
	key     string
	value   []byte
	size    int64
	expires time.Time
	element *list.Element
}

func NewLRU(config Config) Cache {
	return &LRUCache{
		items:     make(map[string]*entry),
		evictList: list.New(),
		config:    config,
	}
}

func (c *LRUCache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	entry, exists := c.items[key]
	c.mu.RUnlock()

	if !exists {
		atomic.AddUint64(&c.stats.misses, 1)
		return nil, false
	}

	if time.Now().After(entry.expires) {
		c.Delete(key)
		atomic.AddUint64(&c.stats.misses, 1)
		return nil, false
	}

	c.mu.Lock()
	c.evictList.MoveToFront(entry.element)
	c.mu.Unlock()

	atomic.AddUint64(&c.stats.hits, 1)
	return entry.value, true
}

func (c *LRUCache) Set(key string, value []byte, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ttl == 0 {
		ttl = c.config.DefaultTTL
	}

	if existing, exists := c.items[key]; exists {
		c.removeElement(existing.element)
	}

	valueSize := int64(len(value))
	for atomic.LoadInt64(&c.stats.bytes)+valueSize > int64(c.config.MaxSize) && c.evictList.Len() > 0 {
		c.removeOldest()
	}

	ent := &entry{
		key:     key,
		value:   value,
		size:    valueSize,
		expires: time.Now().Add(ttl),
	}
	ent.element = c.evictList.PushFront(ent)
	c.items[key] = ent
	atomic.AddInt64(&c.stats.bytes, valueSize)

	// Update size tracking
	atomic.StoreInt64(&c.stats.size, int64(len(c.items)))
}

func (c *LRUCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if ent, exists := c.items[key]; exists {
		c.removeElement(ent.element)
	}
}

func (c *LRUCache) removeElement(e *list.Element) {
	c.evictList.Remove(e)
	ent := e.Value.(*entry)
	delete(c.items, ent.key)
	atomic.AddInt64(&c.stats.bytes, -ent.size)
	atomic.AddUint64(&c.stats.evictions, 1)
}

func (c *LRUCache) removeOldest() {
	if ele := c.evictList.Back(); ele != nil {
		c.removeElement(ele)
	}
}

func (c *LRUCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

func (c *LRUCache) Cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for _, ent := range c.items {
		if now.After(ent.expires) {
			c.removeElement(ent.element)
		}
	}
}

func (c *LRUCache) startCleanup() {
	ticker := time.NewTicker(c.config.CleanupInterval)
	for range ticker.C {
		c.Cleanup()
	}
}

func (c *LRUCache) Stats() Stats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return Stats{
		Size:          int(atomic.LoadInt64(&c.stats.size)),
		BytesInMemory: uint64(atomic.LoadInt64(&c.stats.bytes)),
		Hits:          int64(atomic.LoadUint64(&c.stats.hits)),
		Misses:        int64(atomic.LoadUint64(&c.stats.misses)),
		Evictions:     int64(atomic.LoadUint64(&c.stats.evictions)),
	}
}
