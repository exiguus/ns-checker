package dns_listener

import (
	"fmt"

	"github.com/exiguus/ns-checker/dns_listener/config"
	"github.com/exiguus/ns-checker/dns_listener/types"
)

type Server struct {
	cache     *Cache
	config    *config.Config
	requestCh chan Request
}

func NewServer(cfg *config.Config) (*Server, error) {
	// Create cache with explicit TTL from config
	ttlSeconds := int(cfg.CacheTTL.Seconds())
	fmt.Printf("Creating cache with TTL: %d seconds\n", ttlSeconds)

	server := &Server{
		cache:     NewCache(ttlSeconds),
		config:    cfg,
		requestCh: make(chan Request, cfg.WorkerCount*20),
	}

	return server, nil
}

// Update the status display to show actual TTL
func (s *Server) displayStatus() {
	fmt.Printf(`
=== DNS Listener Configuration ===
► Port: %s
► Worker Pool Size: %d workers
► Request Channel Buffer: %d requests
► Rate Limit: %.0f requests/second (burst: %d)
► DNS Message Buffer Size: %d bytes
► Cache TTL: %v
► Cache Cleanup Interval: %v
===================================
`,
		s.config.Port,
		s.config.WorkerCount,
		cap(s.requestCh),
		s.config.RateLimit,
		s.config.RateBurst,
		types.DefaultBufferSize,
		s.config.CacheTTL,
		s.config.CacheCleanupInterval,
	)
}
