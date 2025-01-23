package health

import (
	"runtime"
	"sync/atomic"
	"time"
)

type SystemStats struct {
	CPUUsage       float64
	MemoryUsage    float64
	GCPause        time.Duration
	GoroutineCount int
	ThreadCount    int
	HeapInUse      uint64
	StackInUse     uint64
	LastGC         time.Time
	Uptime         time.Duration
}

type HealthMonitor struct {
	startTime   time.Time
	stats       atomic.Value // holds *SystemStats
	interval    time.Duration
	stopCh      chan struct{}
	lastCPUTime time.Time
	lastCPUStat float64
	lastGC      time.Time
	gcPause     time.Duration
	lastPause   uint32
}

func NewMonitor(interval time.Duration) *HealthMonitor {
	m := &HealthMonitor{
		startTime:   time.Now(),
		interval:    interval,
		stopCh:      make(chan struct{}),
		lastCPUTime: time.Now(),
	}
	m.stats.Store(&SystemStats{})
	go m.collect()
	return m
}

func (m *HealthMonitor) collect() {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-m.stopCh:
			return
		case <-ticker.C:
			var stats runtime.MemStats
			runtime.ReadMemStats(&stats)

			// Track GC stats
			if stats.NumGC > m.lastPause {
				m.gcPause = time.Duration(stats.PauseNs[(stats.NumGC+255)%256])
				m.lastGC = time.Now().Add(-time.Duration(stats.LastGC))
				m.lastPause = stats.NumGC
			}

			// Calculate CPU usage
			now := time.Now()
			duration := now.Sub(m.lastCPUTime).Seconds()

			if duration > 0 {
				// Get number of CPU cores
				numCPU := float64(runtime.NumCPU())

				// Get the number of goroutines as a rough approximation of CPU load
				numGoroutines := float64(runtime.NumGoroutine())

				// Calculate CPU usage as a percentage of available CPU capacity
				cpuUsage := (numGoroutines / numCPU) * float64(runtime.GOMAXPROCS(0))

				// Normalize to a value between 0 and 1
				m.lastCPUStat = cpuUsage / (numCPU * 100)
			}

			m.lastCPUTime = now

			systemStats := &SystemStats{
				CPUUsage:       m.lastCPUStat,
				MemoryUsage:    float64(stats.Alloc) / float64(stats.Sys),
				GCPause:        m.gcPause,
				GoroutineCount: runtime.NumGoroutine(),
				ThreadCount:    runtime.NumCPU(),
				HeapInUse:      stats.HeapInuse,
				StackInUse:     stats.StackInuse,
				LastGC:         m.lastGC,
				Uptime:         time.Since(m.startTime),
			}
			m.stats.Store(systemStats)
		}
	}
}

func (m *HealthMonitor) GetStats() SystemStats {
	return *m.stats.Load().(*SystemStats)
}

func (m *HealthMonitor) Stop() {
	close(m.stopCh)
}
