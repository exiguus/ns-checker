package metrics

import (
	"sync"
	"sync/atomic"
	"time"
)

type DNSMetrics struct {
	mu              sync.Mutex
	TotalRequests   uint64
	CacheHits       uint64
	CacheMisses     uint64
	ErrorCount      uint64
	LastRequestTime int64 // Unix timestamp
	ProcessingTimes []time.Duration
	maxSamples      int
}

func New(maxSamples int) *DNSMetrics {
	return &DNSMetrics{
		ProcessingTimes: make([]time.Duration, 0, maxSamples),
		maxSamples:      maxSamples,
	}
}

// RecordRequest increments the total requests counter
func (m *DNSMetrics) RecordRequest() {
	atomic.AddUint64(&m.TotalRequests, 1)
	m.mu.Lock()
	m.LastRequestTime = time.Now().Unix()
	m.mu.Unlock()
}

func (m *DNSMetrics) RecordError() {
	atomic.AddUint64(&m.ErrorCount, 1)
}

func (m *DNSMetrics) RecordCacheHit() {
	atomic.AddUint64(&m.CacheHits, 1)
}

func (m *DNSMetrics) RecordCacheMiss() {
	atomic.AddUint64(&m.CacheMisses, 1)
}

func (m *DNSMetrics) GetStats() map[string]interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()
	return map[string]interface{}{
		"total_requests":   atomic.LoadUint64(&m.TotalRequests),
		"cache_hits":       atomic.LoadUint64(&m.CacheHits),
		"cache_misses":     atomic.LoadUint64(&m.CacheMisses),
		"errors":           atomic.LoadUint64(&m.ErrorCount),
		"last_request":     time.Unix(m.LastRequestTime, 0),
		"processing_times": m.ProcessingTimes,
	}
}
