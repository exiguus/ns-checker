package cache

import (
	"hash/fnv"
	"sync"
	"sync/atomic"
	"time"
)

type ShardedCache struct {
	shards    []*cacheShard
	numShards int
	mask      uint32
	config    Config
	stats     struct {
		hits      uint64
		misses    uint64
		evictions uint64
		bytes     int64
	}
}

type cacheShard struct {
	sync.RWMutex
	items map[string]*cacheItem
}

type cacheItem struct {
	value      []byte
	expiration time.Time
	size       int64
	hits       uint64
}

func NewSharded(config Config, shards int) Cache {
	if shards <= 0 {
		shards = 32
	}
	shards = 1 << uint(32-numberOfLeadingZeros32(uint32(shards-1)))

	sc := &ShardedCache{
		shards:    make([]*cacheShard, shards),
		numShards: shards,
		mask:      uint32(shards - 1),
		config:    config,
	}

	for i := 0; i < shards; i++ {
		sc.shards[i] = &cacheShard{
			items: make(map[string]*cacheItem),
		}
	}

	return sc
}

func numberOfLeadingZeros32(x uint32) uint32 {
	if x == 0 {
		return 32
	}
	n := uint32(0)
	if x <= 0x0000FFFF {
		n += 16
		x <<= 16
	}
	if x <= 0x00FFFFFF {
		n += 8
		x <<= 8
	}
	if x <= 0x0FFFFFFF {
		n += 4
		x <<= 4
	}
	if x <= 0x3FFFFFFF {
		n += 2
		x <<= 2
	}
	if x <= 0x7FFFFFFF {
		n++
	}
	return n
}

func (sc *ShardedCache) getShard(key string) *cacheShard {
	h := fnv.New32a()
	h.Write([]byte(key))
	return sc.shards[h.Sum32()&sc.mask]
}

func (sc *ShardedCache) Get(key string) ([]byte, bool) {
	shard := sc.getShard(key)
	shard.RLock()
	item, exists := shard.items[key]
	shard.RUnlock()

	if !exists {
		atomic.AddUint64(&sc.stats.misses, 1)
		return nil, false
	}

	if time.Now().After(item.expiration) {
		shard.Lock()
		delete(shard.items, key)
		atomic.AddUint64(&sc.stats.evictions, 1)
		atomic.AddInt64(&sc.stats.bytes, -item.size)
		shard.Unlock()
		return nil, false
	}

	atomic.AddUint64(&item.hits, 1)
	atomic.AddUint64(&sc.stats.hits, 1)
	return item.value, true
}

func (sc *ShardedCache) Set(key string, value []byte, ttl time.Duration) {
	if ttl == 0 {
		ttl = sc.config.DefaultTTL
	}

	shard := sc.getShard(key)
	shard.Lock()
	defer shard.Unlock()

	// Check size before adding
	valueSize := int64(len(value))
	if atomic.LoadInt64(&sc.stats.bytes)+valueSize > int64(sc.config.MaxSize) {
		sc.evict()
	}

	// Update or add item
	if existing, exists := shard.items[key]; exists {
		atomic.AddInt64(&sc.stats.bytes, -existing.size)
	}

	shard.items[key] = &cacheItem{
		value:      value,
		expiration: time.Now().Add(ttl),
		size:       valueSize,
	}
	atomic.AddInt64(&sc.stats.bytes, valueSize)
}

func (sc *ShardedCache) Delete(key string) {
	shard := sc.getShard(key)
	shard.Lock()
	if item, exists := shard.items[key]; exists {
		atomic.AddInt64(&sc.stats.bytes, -item.size)
		delete(shard.items, key)
	}
	shard.Unlock()
}

func (sc *ShardedCache) Size() int {
	var size int
	for _, shard := range sc.shards {
		shard.RLock()
		size += len(shard.items)
		shard.RUnlock()
	}
	return size
}

func (sc *ShardedCache) Cleanup() {
	now := time.Now()
	for _, shard := range sc.shards {
		shard.Lock()
		for key, item := range shard.items {
			if now.After(item.expiration) {
				atomic.AddInt64(&sc.stats.bytes, -item.size)
				delete(shard.items, key)
				atomic.AddUint64(&sc.stats.evictions, 1)
			}
		}
		shard.Unlock()
	}
}

func (sc *ShardedCache) startCleanup() {
	ticker := time.NewTicker(sc.config.CleanupInterval)
	for range ticker.C {
		sc.Cleanup()
	}
}

func (sc *ShardedCache) Stats() Stats {
	var stats Stats
	for _, shard := range sc.shards {
		shard.RLock()
		stats.Size += len(shard.items)
		for _, item := range shard.items {
			stats.BytesInMemory += uint64(item.size)
		}
		shard.RUnlock()
	}
	stats.Hits = int64(atomic.LoadUint64(&sc.stats.hits))
	stats.Misses = int64(atomic.LoadUint64(&sc.stats.misses))
	stats.Evictions = int64(atomic.LoadUint64(&sc.stats.evictions))
	return stats
}

func (sc *ShardedCache) evict() {
	switch sc.config.EvictionPolicy {
	case LRU:
		sc.evictLRU()
	case LFU:
		sc.evictLFU()
	default:
		sc.evictFIFO() // Default to FIFO if policy not specified
	}
}

func (sc *ShardedCache) evictLRU() {
	var maxShard *cacheShard
	maxItems := 0

	for _, shard := range sc.shards {
		shard.RLock()
		if len(shard.items) > maxItems {
			maxItems = len(shard.items)
			maxShard = shard
		}
		shard.RUnlock()
	}

	if maxShard != nil {
		maxShard.Lock()
		for key, item := range maxShard.items {
			atomic.AddInt64(&sc.stats.bytes, -item.size)
			delete(maxShard.items, key)
			atomic.AddUint64(&sc.stats.evictions, 1)
			break // Just remove one item
		}
		maxShard.Unlock()
	}
}

func (sc *ShardedCache) evictLFU() {
	// Similar to LRU but based on hit count
	// Implementation omitted for brevity
}

func (sc *ShardedCache) evictFIFO() {
	// Similar to LRU but simpler removal
	// Implementation omitted for brevity
}
