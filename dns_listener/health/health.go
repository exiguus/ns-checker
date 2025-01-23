package health

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Server struct {
	port    string
	metrics MetricsProvider
}

type MetricsProvider interface {
	GetStats() map[string]interface{}
}

type HealthStatus struct {
	Status    string                 `json:"status"`
	Timestamp string                 `json:"timestamp"`
	Metrics   map[string]interface{} `json:"metrics,omitempty"`
}

func NewServer(port string, metrics MetricsProvider) *Server {
	return &Server{
		port:    port,
		metrics: metrics,
	}
}

func (s *Server) Start() error {
	http.HandleFunc("/health", s.handleHealth)
	http.HandleFunc("/metrics", s.handleMetrics)

	addr := fmt.Sprintf(":%s", s.port)
	return http.ListenAndServe(addr, nil)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	status := HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func (s *Server) handleMetrics(w http.ResponseWriter, r *http.Request) {
	status := HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Metrics:   s.metrics.GetStats(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}
