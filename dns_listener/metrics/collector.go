package metrics

import (
	"sync"
	"sync/atomic"
	"time"
)

type Collector struct {
	totalRequests    uint64
	cacheHits        uint64
	cacheMisses      uint64
	errors           uint64
	responseTimes    []time.Duration
	responseTimeLock sync.RWMutex
}

func NewCollector() *Collector {
	return &Collector{
		responseTimes: make([]time.Duration, 0, 1000),
	}
}

func (c *Collector) RecordRequest()           { atomic.AddUint64(&c.totalRequests, 1) }
func (c *Collector) RecordCacheHit()          { atomic.AddUint64(&c.cacheHits, 1) }
func (c *Collector) RecordCacheMiss()         { atomic.AddUint64(&c.cacheMisses, 1) }
func (c *Collector) RecordError()             { atomic.AddUint64(&c.errors, 1) }
func (c *Collector) GetTotalRequests() uint64 { return atomic.LoadUint64(&c.totalRequests) }
func (c *Collector) GetCacheHits() uint64     { return atomic.LoadUint64(&c.cacheHits) }
func (c *Collector) GetCacheMisses() uint64   { return atomic.LoadUint64(&c.cacheMisses) }
func (c *Collector) GetErrors() uint64        { return atomic.LoadUint64(&c.errors) }

func (c *Collector) RecordResponseTime(d time.Duration) {
	c.responseTimeLock.Lock()
	defer c.responseTimeLock.Unlock()

	c.responseTimes = append(c.responseTimes, d)
	if len(c.responseTimes) > 1000 {
		c.responseTimes = c.responseTimes[1:]
	}
}

func (c *Collector) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"total_requests": c.GetTotalRequests(),
		"cache_hits":     c.GetCacheHits(),
		"cache_misses":   c.GetCacheMisses(),
		"errors":         c.GetErrors(),
	}
}

// Add GetRawStats method to Collector
func (c *Collector) GetRawStats() map[string]uint64 {
	return map[string]uint64{
		"total_requests": c.GetTotalRequests(),
		"cache_hits":     c.GetCacheHits(),
		"cache_misses":   c.GetCacheMisses(),
		"errors":         c.GetErrors(),
	}
}
