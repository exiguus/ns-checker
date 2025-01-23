package dns_listener

import (
	"context"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/exiguus/ns-checker/dns_listener/cache"
	"github.com/exiguus/ns-checker/dns_listener/config"
	dnserr "github.com/exiguus/ns-checker/dns_listener/errors"
	"github.com/exiguus/ns-checker/dns_listener/health"
	"github.com/exiguus/ns-checker/dns_listener/metrics"
	"github.com/exiguus/ns-checker/dns_listener/network"
	"github.com/exiguus/ns-checker/dns_listener/perf"
	"github.com/exiguus/ns-checker/dns_listener/processor"
	"github.com/exiguus/ns-checker/dns_listener/protocol"
	"github.com/exiguus/ns-checker/dns_listener/ratelimit"
	"github.com/exiguus/ns-checker/dns_listener/tracing"
	"github.com/exiguus/ns-checker/dns_listener/types"
	"github.com/exiguus/ns-checker/dns_listener/validator"
)

type DNSListener struct {
	port        string
	metrics     *metrics.Collector
	config      *config.Config
	cache       cache.Cache
	logger      Logger
	rateLimiter *ratelimit.RateLimiter
	validator   validator.MessageValidator
	bufPool     sync.Pool
	stopChan    chan struct{}
	wg          sync.WaitGroup
	processor   *processor.Processor
	requestCh   chan types.Request
	tracer      *tracing.Tracer
	perfMon     *perf.Monitor
	healthMon   *health.HealthMonitor
}

func NewDNSListener(cfg *config.Config) (*DNSListener, error) {
	parsedPort, err := network.ParsePort(cfg.Port)
	if err != nil {
		return nil, err
	}

	// Update port in config with parsed value
	cfg.Port = fmt.Sprintf("%d", parsedPort)

	logger, err := NewFileLogger(cfg.LogPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	// Ensure config has valid TTL
	if cfg.CacheTTL == 0 {
		cfg.CacheTTL = 30 * time.Minute
	}
	if cfg.CacheCleanupInterval == 0 {
		cfg.CacheCleanupInterval = time.Minute
	}

	ttlSeconds := int(cfg.CacheTTL.Seconds())
	if ttlSeconds <= 0 {
		ttlSeconds = 1800 // Default to 30 minutes if invalid
	}

	// Initialize cache with proper configuration
	cacheConfig := cache.Config{
		MaxSize:         1024 * 1024 * 100,
		DefaultTTL:      cfg.CacheTTL,
		CleanupInterval: cfg.CacheCleanupInterval,
	}

	// Use New instead of NewBasicCache to match the interface
	cacheImpl := cache.New(cacheConfig)

	listener := &DNSListener{
		port:        cfg.Port,
		metrics:     metrics.NewCollector(),
		config:      cfg,
		cache:       cacheImpl,
		logger:      logger,
		rateLimiter: ratelimit.New(cfg.RateLimit, cfg.RateBurst),
		validator:   validator.New(),
		bufPool:     sync.Pool{New: func() interface{} { return make([]byte, types.DefaultBufferSize) }},
		stopChan:    make(chan struct{}),
		requestCh:   make(chan types.Request, cfg.WorkerCount*20),
		tracer:      tracing.New(),
		perfMon:     perf.New(time.Second),
		healthMon:   health.NewMonitor(time.Second),
	}

	// Initialize processor after listener is created
	procConfig := processor.ProcessorConfig{
		Workers:    cfg.WorkerCount,
		Timeout:    30 * time.Second,
		BufferSize: cfg.WorkerCount * 20,
	}
	listener.processor = processor.New(procConfig, listener, metrics.NewCollector())

	return listener, nil
}

func (d *DNSListener) GetPort() string {
	return d.port
}

func (d *DNSListener) GetMetrics() metrics.MetricsCollector {
	return d.metrics
}

func (d *DNSListener) HandleRequest(data []byte, addr net.Addr, protocolType string) ([]byte, error) {
	start := time.Now()
	defer func() {
		d.perfMon.RecordResponseTime(time.Since(start))
	}()

	if !d.rateLimiter.Allow(addr.String()) {
		d.metrics.RecordError()
		return nil, dnserr.NewValidationError("HandleRequest", "rate limit exceeded", nil)
	}

	ctx := d.tracer.StartTrace(context.Background())
	d.tracer.AddEvent(ctx, "request_start", nil)

	d.logger.LogRequest(protocolType, addr.String(), data, nil)

	d.metrics.RecordRequest()

	if cachedResponse := d.checkCache(data); cachedResponse != nil {
		d.metrics.RecordCacheHit()
		d.logger.Write(fmt.Sprintf("Cache hit for %s\n", addr.String()))
		d.tracer.AddEvent(ctx, "cache_hit", nil)
		d.tracer.AddEvent(ctx, "request_complete", nil)

		// Create fresh response instead of using cached one
		response := protocol.CreateDNSResponse(data, addr.String())
		if response != nil {
			return response, nil
		}
	}
	d.metrics.RecordCacheMiss()

	if err := d.validator.ValidateQuery(data); err != nil {
		d.metrics.RecordError()
		d.logger.Write(fmt.Sprintf("Validation error for %s: %v\n", addr.String(), err))
		d.tracer.AddEvent(ctx, "validation_error", err)
		d.tracer.AddEvent(ctx, "request_complete", nil)
		return nil, dnserr.NewValidationError("HandleRequest", "invalid query", err)
	}

	response := protocol.CreateDNSResponse(data, addr.String())
	if response == nil {
		err := dnserr.NewInternalError("HandleRequest", "failed to create response", nil)
		d.metrics.RecordError()
		d.logger.Write(fmt.Sprintf("Response creation error for %s: %v\n", addr.String(), err))
		d.tracer.AddEvent(ctx, "response_creation_error", err)
		d.tracer.AddEvent(ctx, "request_complete", nil)
		return nil, err
	}

	if err := d.validator.ValidateResponse(response); err != nil {
		d.metrics.RecordError()
		d.tracer.AddEvent(ctx, "response_validation_error", err)
		d.tracer.AddEvent(ctx, "request_complete", nil)
		return nil, dnserr.NewValidationError("HandleRequest", "invalid response", err)
	}

	d.logger.Write(fmt.Sprintf("Created response for %s (%d bytes)\n", addr.String(), len(response)))

	d.updateCache(data, response)
	d.tracer.AddEvent(ctx, "request_complete", nil)
	return response, nil
}

func (d *DNSListener) handleRequest(conn net.Conn, protocol string, clientAddr net.Addr) {
	req := types.Request{
		Conn:       conn,
		Protocol:   protocol,
		ClientAddr: clientAddr,
	}
	d.processor.Process(req)
}

func (d *DNSListener) sendResponse(conn net.Conn, response []byte) error {
	_, err := conn.Write(response)
	return err
}

func (d *DNSListener) checkCache(query []byte) []byte {
	key := cacheKeyFromQuery(query)

	if response, ok := d.cache.Get(key); ok {
		return response
	}
	return nil
}

func (d *DNSListener) updateCache(query, response []byte) {
	key := cacheKeyFromQuery(query)
	d.cache.Set(key, response, d.config.CacheTTL)
}

func cacheKeyFromQuery(query []byte) string {
	if len(query) < 12 {
		return hex.EncodeToString(query)
	}

	pos := 12
	questionCount := int(query[4])<<8 | int(query[5])

	// Skip questions to find end of question section
	for i := 0; i < questionCount && pos < len(query); i++ {
		// Skip name
		for pos < len(query) {
			length := int(query[pos])
			if length == 0 {
				pos++
				break
			}
			pos += length + 1
		}
		pos += 4 // Skip QTYPE and QCLASS
	}

	// Use only question section for cache key
	return hex.EncodeToString(query[12:pos])
}

func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%.2fµs", float64(d.Microseconds()))
	}
	if d < time.Second {
		return fmt.Sprintf("%.2fms", float64(d.Milliseconds()))
	}
	return d.Round(time.Millisecond).String()
}

func humanizeBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func formatResponseTime(d time.Duration) string {
	if d == 0 {
		return "0"
	}
	if d < time.Microsecond {
		return fmt.Sprintf("%.2fns", float64(d.Nanoseconds()))
	}
	if d < time.Millisecond {
		return fmt.Sprintf("%.2fµs", float64(d.Microseconds()))
	}
	if d < time.Second {
		return fmt.Sprintf("%.2fms", float64(d.Milliseconds()))
	}
	return fmt.Sprintf("%.2fs", d.Seconds())
}

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func formatGCTime(t time.Time) string {
	if t.IsZero() {
		return "never"
	}
	d := time.Since(t)
	if d < time.Minute {
		return fmt.Sprintf("%.2fs ago", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm%ds ago", d/time.Minute, (d%time.Minute)/time.Second)
	}
	return fmt.Sprintf("%dh%dm ago", d/time.Hour, (d%time.Hour)/time.Minute)
}

// Add this new method
func (d *DNSListener) getChannelStats() struct {
	current     int
	capacity    int
	utilization int
} {
	current := len(d.requestCh)
	capacity := cap(d.requestCh)
	utilization := 0
	if capacity > 0 {
		// Calculate utilization as a percentage with floating-point precision
		// before converting to integer to avoid integer division truncation
		utilization = int((float64(current) / float64(capacity)) * 100.0)
	}

	return struct {
		current     int
		capacity    int
		utilization int
	}{
		current:     current,
		capacity:    capacity,
		utilization: utilization,
	}
}

func (d *DNSListener) monitorStats() {
	ticker := time.NewTicker(30 * time.Second)
	startTime := time.Now()
	for range ticker.C {
		cacheStats := d.cache.Stats()
		rawStats := d.metrics.GetRawStats()
		rlStats := d.rateLimiter.GetStats()
		valStats := d.validator.GetStats()
		perfStats := d.perfMon.GetStats()
		healthStats := d.healthMon.GetStats()

		// Convert RateBurst to int32 for calculation
		rateBurst := int32(d.config.RateBurst)
		activeClientsPercent := float64(rlStats.ActiveKeys) / float64(rateBurst) * 100

		// Replace the Channel Load stats calculation with:
		channelStats := d.getChannelStats()

		stats := fmt.Sprintf(`
%s=== Runtime Statistics ===%s
► System Health:
  • CPU Usage: %.1f%%
  • Memory Usage: %.1f%%
  • Uptime: %s
  • Last GC: %s
  • GC Pause: %s
► Cache:
  • Size: %d entries (%s)
  • Hit Ratio: %.1f%% (%d/%d)
  • Evictions: %d
► Processing:
  • Channel Load: %d/%d (%d%% utilized)
  • Total Requests: %d (%.1f/sec avg)
  • Goroutines: %d
  • Heap Usage: %s
► Performance:
  • Request Rate: %.1f/sec current
  • Response Times:
    - Avg: %s
    - P95: %s
    - P99: %s
► Rate Limiting:
  • Limited Requests: %d
  • Active Clients: %d (%d%% of limit)
  • Burst Usage: %.1f%%
► Validation:
  • Success Rate: %.1f%% (%d/%d total)
  • Invalid Queries: %d
  • Invalid Responses: %d
%s=========================%s
`,
			colorYellow,
			colorReset,
			healthStats.CPUUsage*100,
			healthStats.MemoryUsage*100,
			formatDuration(time.Since(startTime)),
			formatGCTime(healthStats.LastGC),
			formatResponseTime(healthStats.GCPause),
			cacheStats.Size,
			humanizeBytes(cacheStats.BytesInMemory),
			float64(cacheStats.Hits)/(float64(cacheStats.Hits+cacheStats.Misses))*100,
			cacheStats.Hits,
			cacheStats.Hits+cacheStats.Misses,
			cacheStats.Evictions,
			channelStats.current, channelStats.capacity, channelStats.utilization,
			rawStats["total_requests"],
			float64(rawStats["total_requests"])/time.Since(startTime).Seconds(),
			perfStats.Goroutines,
			humanizeBytes(perfStats.HeapAlloc),
			perfStats.RequestRate,
			formatResponseTime(perfStats.AvgResponseTime),
			formatResponseTime(perfStats.P95),
			formatResponseTime(perfStats.P99),
			rlStats.Limited,
			rlStats.ActiveKeys,
			int(activeClientsPercent), // Convert to int for display
			rlStats.BurstUsage*100,
			float64(valStats.TotalValidated-valStats.InvalidQueries-valStats.InvalidResponses)/float64(valStats.TotalValidated)*100,
			valStats.TotalValidated-valStats.InvalidQueries-valStats.InvalidResponses,
			valStats.TotalValidated,
			valStats.InvalidQueries,
			valStats.InvalidResponses,
			colorYellow,
			colorReset,
		)

		fmt.Print(stats)
		os.Stdout.Sync()
	}
}

// Cache returns the cache instance for testing
func (d *DNSListener) Cache() cache.Cache {
	return d.cache
}
