package metrics

import "time"

// MetricsCollector interface defines the methods that a metrics collector must implement
type MetricsCollector interface {
	RecordRequest()
	RecordCacheHit()
	RecordCacheMiss()
	RecordError()
	RecordResponseTime(time.Duration)
	GetTotalRequests() uint64
	GetCacheHits() uint64
	GetCacheMisses() uint64
	GetErrors() uint64
	GetStats() map[string]interface{}
	GetRawStats() map[string]uint64
}
