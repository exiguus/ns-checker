package metrics

import "time"

// MetricType represents different types of metrics
type MetricType int

const (
	CounterMetric MetricType = iota
	GaugeMetric
	HistogramMetric
)

// MetricValue represents a metric value with timestamp
type MetricValue struct {
	Value     float64
	Timestamp time.Time
	Labels    map[string]string
}

// MetricDefinition defines a metric
type MetricDefinition struct {
	Name        string
	Type        MetricType
	Description string
	Labels      []string
}
