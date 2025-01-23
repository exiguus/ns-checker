package monitoring

import (
	"runtime"
	"sync"
	"time"
)

type SystemStats struct {
	CPUUsage    float64
	MemoryUsage float64
	Goroutines  int
	HeapAlloc   uint64
	StackInUse  uint64
	LastGC      time.Time
}

type Monitor struct {
	mu           sync.RWMutex
	stats        SystemStats
	updateTicker *time.Ticker
	stopChan     chan struct{}
}

func NewMonitor(interval time.Duration) *Monitor {
	m := &Monitor{
		updateTicker: time.NewTicker(interval),
		stopChan:     make(chan struct{}),
	}
	go m.run()
	return m
}

func (m *Monitor) run() {
	for {
		select {
		case <-m.updateTicker.C:
			m.updateStats()
		case <-m.stopChan:
			m.updateTicker.Stop()
			return
		}
	}
}

func (m *Monitor) updateStats() {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)

	m.mu.Lock()
	defer m.mu.Unlock()

	m.stats = SystemStats{
		CPUUsage:    getCPUUsage(),
		MemoryUsage: float64(stats.Alloc) / float64(stats.Sys),
		Goroutines:  runtime.NumGoroutine(),
		HeapAlloc:   stats.HeapAlloc,
		StackInUse:  stats.StackInuse,
		LastGC:      time.Unix(0, int64(stats.LastGC)),
	}
}

func (m *Monitor) GetStats() SystemStats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.stats
}

func (m *Monitor) Stop() {
	close(m.stopChan)
}

// getCPUUsage returns a value between 0 and 1 representing CPU usage
func getCPUUsage() float64 {
	// Implementation would depend on the OS
	// This is a placeholder that should be implemented
	// using actual CPU measurements
	return 0.0
}
