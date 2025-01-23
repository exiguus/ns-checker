package ratelimit

import (
	"sync"
	"sync/atomic"
	"time"
)

// RateLimiter implements a token bucket rate limiter
type RateLimiter struct {
	mu           sync.RWMutex
	limits       map[string]*bucket
	rate         float64
	burst        int
	cleanupEvery time.Duration
	stats        struct {
		allowed    uint64
		limited    uint64
		activeKeys int32
	}
}

type bucket struct {
	tokens    float64
	lastCheck time.Time
}

// Stats represents rate limiter statistics
type Stats struct {
	Allowed    uint64
	Limited    uint64
	ActiveKeys int32
	BurstUsage float64
}

// New creates a new rate limiter
func New(rate float64, burst int) *RateLimiter {
	rl := &RateLimiter{
		limits:       make(map[string]*bucket),
		rate:         rate,
		burst:        burst,
		cleanupEvery: 5 * time.Minute,
	}
	go rl.cleanup()
	return rl
}

// Allow checks if a request should be allowed
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	b, exists := rl.limits[key]
	if !exists {
		b = &bucket{
			tokens:    float64(rl.burst),
			lastCheck: now,
		}
		rl.limits[key] = b
		atomic.AddInt32(&rl.stats.activeKeys, 1)
	}

	elapsed := now.Sub(b.lastCheck).Seconds()
	b.tokens += elapsed * rl.rate
	if b.tokens > float64(rl.burst) {
		b.tokens = float64(rl.burst)
	}
	b.lastCheck = now

	if b.tokens >= 1 {
		b.tokens--
		atomic.AddUint64(&rl.stats.allowed, 1)
		return true
	}

	atomic.AddUint64(&rl.stats.limited, 1)
	return false
}

// cleanup removes inactive buckets periodically
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.cleanupEvery)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, bucket := range rl.limits {
			if now.Sub(bucket.lastCheck) > rl.cleanupEvery {
				delete(rl.limits, key)
				atomic.AddInt32(&rl.stats.activeKeys, -1)
			}
		}
		rl.mu.Unlock()
	}
}

// GetStats returns current rate limiter statistics
func (rl *RateLimiter) GetStats() Stats {
	rl.mu.RLock()
	defer rl.mu.RUnlock()

	stats := Stats{
		Allowed:    atomic.LoadUint64(&rl.stats.allowed),
		Limited:    atomic.LoadUint64(&rl.stats.limited),
		ActiveKeys: atomic.LoadInt32(&rl.stats.activeKeys),
	}

	var totalTokens float64
	for _, b := range rl.limits {
		totalTokens += b.tokens
	}
	if len(rl.limits) > 0 {
		stats.BurstUsage = 1 - (totalTokens / (float64(len(rl.limits)) * float64(rl.burst)))
	}

	return stats
}
