package health

import "time"

type Status struct {
	Healthy     bool `json:"healthy"`
	SystemStats struct {
		CPUUsage    float64 `json:"cpu_usage"`
		MemoryUsage float64 `json:"memory_usage"`
		Uptime      string  `json:"uptime"`
	} `json:"system_stats"`
	PerformanceStats struct {
		RequestRate    float64       `json:"request_rate"`
		AverageLatency time.Duration `json:"average_latency"`
		P95Latency     time.Duration `json:"p95_latency"`
		ErrorRate      float64       `json:"error_rate"`
		CacheHitRate   float64       `json:"cache_hit_rate"`
		GoroutineCount int           `json:"goroutine_count"`
		LastGCPause    time.Duration `json:"last_gc_pause"`
	} `json:"performance_stats"`
	Thresholds struct {
		MaxCPUUsage    float64       `json:"max_cpu_usage"`
		MaxMemoryUsage float64       `json:"max_memory_usage"`
		MaxLatency     time.Duration `json:"max_latency"`
		MaxErrorRate   float64       `json:"max_error_rate"`
	} `json:"thresholds"`
}

type HealthCheck struct {
	Name     string        `json:"name"`
	Status   bool          `json:"status"`
	Message  string        `json:"message"`
	LastRun  time.Time     `json:"last_run"`
	Duration time.Duration `json:"duration"`
}
