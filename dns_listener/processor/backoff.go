package processor

import (
	"math"
	"math/rand"
	"time"
)

const (
	minBackoff = 100 * time.Millisecond
	maxBackoff = 2 * time.Second
	factor     = 2.0
	jitter     = 0.1
)

// backoff calculates the next backoff duration using exponential backoff with jitter
func backoff(attempt int) time.Duration {
	if attempt <= 0 {
		return minBackoff
	}

	// Calculate base backoff
	backoff := float64(minBackoff) * math.Pow(factor, float64(attempt-1))

	// Add jitter
	backoff = backoff * (1 + jitter*rand.Float64())

	// Ensure we don't exceed max backoff
	if backoff > float64(maxBackoff) {
		backoff = float64(maxBackoff)
	}

	return time.Duration(backoff)
}
