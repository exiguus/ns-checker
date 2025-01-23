package perf

import (
	"fmt"
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

type Stats struct {
	Goroutines      int
	HeapAlloc       uint64
	HeapObjects     uint64
	GCPauses        uint64
	LastGCTime      time.Duration
	CPUUsage        float64
	ResponseTimes   []time.Duration
	P95             time.Duration
	P99             time.Duration
	AvgResponseTime time.Duration
	MinResponseTime time.Duration
	MaxResponseTime time.Duration
	RequestRate     float64
	LastMinute      struct {
		Count     int
		AvgTime   time.Duration
		ErrorRate float64
	}
}

type Monitor struct {
	stats          atomic.Value // holds *Stats
	samples        []time.Duration
	interval       time.Duration
	lastSampleTime []time.Time
	mu             sync.RWMutex
	lastUpdate     time.Time
	goroutines     uint64
	heapAlloc      uint64
}

func New(sampleInterval time.Duration) *Monitor {
	m := &Monitor{
		interval:       sampleInterval,
		samples:        make([]time.Duration, 0, 1000),
		lastSampleTime: make([]time.Time, 0, 1000),
		lastUpdate:     time.Now(),
	}
	m.stats.Store(&Stats{})

	// Start a goroutine to continuously update runtime stats
	go func() {
		ticker := time.NewTicker(sampleInterval)
		defer ticker.Stop()

		var memStats runtime.MemStats
		for range ticker.C {
			runtime.ReadMemStats(&memStats)
			atomic.StoreUint64(&m.goroutines, uint64(runtime.NumGoroutine()))
			atomic.StoreUint64(&m.heapAlloc, memStats.HeapAlloc)
		}
	}()

	go m.collect()
	return m
}

func (m *Monitor) collect() {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	var lastPause uint32
	var memStats runtime.MemStats

	for range ticker.C {
		runtime.ReadMemStats(&memStats)

		stats := &Stats{
			Goroutines:  runtime.NumGoroutine(),
			HeapAlloc:   memStats.HeapAlloc,
			HeapObjects: memStats.HeapObjects,
			GCPauses:    uint64(memStats.NumGC - lastPause),
			LastGCTime:  time.Duration(memStats.PauseNs[(memStats.NumGC+255)%256]),
		}

		lastPause = memStats.NumGC
		m.stats.Store(stats)
	}
}

func (m *Monitor) RecordResponseTime(d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.samples = append(m.samples, d)
	m.lastSampleTime = append(m.lastSampleTime, time.Now())
	if len(m.samples) > 1000 {
		m.samples = m.samples[1:]
		m.lastSampleTime = m.lastSampleTime[1:]
	}
	m.updatePercentiles()
}

func (m *Monitor) updatePercentiles() {
	if len(m.samples) == 0 {
		return
	}

	sorted := make([]time.Duration, len(m.samples))
	copy(sorted, m.samples)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	var total time.Duration
	for _, d := range sorted {
		total += d
	}

	stats := m.stats.Load().(*Stats)
	stats.ResponseTimes = sorted
	stats.P95 = sorted[len(sorted)*95/100]
	stats.P99 = sorted[len(sorted)*99/100]
	stats.AvgResponseTime = total / time.Duration(len(sorted))
	stats.MinResponseTime = sorted[0]
	stats.MaxResponseTime = sorted[len(sorted)-1]

	// Calculate last minute stats
	now := time.Now()
	lastMinute := 0
	lastMinuteTotal := time.Duration(0)
	for i := len(m.samples) - 1; i >= 0; i-- {
		if now.Sub(m.lastSampleTime[i]) <= time.Minute {
			lastMinute++
			lastMinuteTotal += m.samples[i]
		} else {
			break
		}
	}

	if lastMinute > 0 {
		stats.LastMinute.Count = lastMinute
		stats.LastMinute.AvgTime = lastMinuteTotal / time.Duration(lastMinute)
		stats.RequestRate = float64(lastMinute) / 60.0
	}
}

func (m *Monitor) GetStats() Stats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var stats Stats
	count := len(m.samples)
	if count == 0 {
		return stats
	}

	// Calculate average
	var total time.Duration
	times := make([]time.Duration, count)
	copy(times, m.samples)

	for _, t := range times {
		total += t
	}
	stats.AvgResponseTime = total / time.Duration(count)

	// Calculate percentiles
	sort.Slice(times, func(i, j int) bool {
		return times[i] < times[j]
	})

	p95Index := int(float64(count) * 0.95)
	p99Index := int(float64(count) * 0.99)

	if p95Index < count {
		stats.P95 = times[p95Index]
	}
	if p99Index < count {
		stats.P99 = times[p99Index]
	}

	// Calculate request rate
	now := time.Now()
	duration := now.Sub(m.lastUpdate)
	if duration >= m.interval {
		stats.RequestRate = float64(count) / duration.Seconds()
		m.lastUpdate = now
	}

	stats.Goroutines = int(atomic.LoadUint64(&m.goroutines))
	stats.HeapAlloc = atomic.LoadUint64(&m.heapAlloc)

	return stats
}

// FormatStats returns a formatted string of performance statistics
func (m *Monitor) FormatStats() string {
	stats := m.GetStats()
	return fmt.Sprintf(`Performance Stats:
  • System:
    - Goroutines: %d
    - Heap: %.2f MB
    - GC Pauses: %d
  • Response Times:
    - Average: %v
    - Min/Max: %v/%v
    - P95/P99: %v/%v
  • Last Minute:
    - Requests: %d
    - Rate: %.1f/sec
    - Avg Time: %v`,
		stats.Goroutines,
		float64(stats.HeapAlloc)/1024/1024,
		stats.GCPauses,
		stats.AvgResponseTime,
		stats.MinResponseTime,
		stats.MaxResponseTime,
		stats.P95,
		stats.P99,
		stats.LastMinute.Count,
		stats.RequestRate,
		stats.LastMinute.AvgTime,
	)
}
