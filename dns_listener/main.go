package dns_listener

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/exiguus/ns-checker/dns_listener/config"
	"github.com/exiguus/ns-checker/dns_listener/health"
	"github.com/exiguus/ns-checker/dns_listener/network"
	"github.com/exiguus/ns-checker/dns_listener/protocol/parser"
	"github.com/exiguus/ns-checker/dns_listener/types"
)

const (
	colorReset  = "\033[0m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
)

// Ensure test mode is disabled by default
var isTestMode = false

func init() {
	flag.Parse()
}

func calculateOptimalWorkers() int {
	cpuCount := runtime.NumCPU()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Convert memory to GB
	totalMemoryGB := float64(m.Sys) / (1024 * 1024 * 1024)

	// Base calculation:
	// - Minimum 4 workers
	// - Maximum of (CPU count * 4)
	// - 1 worker per 256MB of system memory
	memoryWorkers := int(totalMemoryGB * 4) // 4 workers per GB of RAM
	cpuWorkers := cpuCount * 3              // CPU count * I/O waiting factor

	// Choose the smaller of the two to avoid overloading
	workers := min(memoryWorkers, cpuWorkers)

	// Ensure minimum and maximum bounds
	workers = max(4, min(workers, cpuCount*4))

	fmt.Printf("System resources: CPU cores: %d, Memory: %.2f GB\n", cpuCount, totalMemoryGB)
	fmt.Printf("Calculated workers: %d (memory workers: %d, CPU workers: %d)\n",
		workers, memoryWorkers, cpuWorkers)

	return workers
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// printBanner prints the DNS server banner
func printBanner() {
	banner := `
╔═══════════════════════════════════╗
║         DNS Listener Active       ║
╚═══════════════════════════════════╝
`
	fmt.Print(colorGreen, banner, colorReset)
	os.Stdout.Sync()
}

// printStats prints the DNS listener configuration and stats
func (d *DNSListener) printStats() {
	stats := fmt.Sprintf(`
%s=== DNS Listener Configuration ===%s
► Port: %s
► Worker Pool Size: %d workers
► Request Channel Buffer: %d requests
► Rate Limit: %.0f requests/second (burst: %d)
► DNS Message Buffer Size: %d bytes
► Cache TTL: %v
► Cache Cleanup Interval: %v
%s===================================%s
`,
		colorCyan,
		colorReset,
		d.config.Port,
		d.config.WorkerCount,
		cap(d.requestCh),
		d.config.RateLimit,
		d.config.RateBurst,
		types.DefaultBufferSize,
		d.config.CacheTTL,
		d.config.CacheCleanupInterval,
		colorCyan,
		colorReset,
	)
	fmt.Print(stats)
	os.Stdout.Sync()
	os.Stderr.Sync() // Add stderr flush
}

func (d *DNSListener) Start() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server without printing message
	server := network.NewServer(d.config.Port, d)

	// Only start cache cleanup if interval is positive
	if d.config.CacheCleanupInterval > 0 {
		go func() {
			ticker := time.NewTicker(d.config.CacheCleanupInterval)
			defer ticker.Stop()
			for range ticker.C {
				d.cache.Cleanup()
			}
		}()
	}
	go d.monitorStats()

	// Handle shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nShutting down gracefully...")
		d.logger.Write("DNS Listener stopped")
		cancel()
		server.Stop()
		d.Close()
	}()

	printBanner()
	d.printStats()

	// Block on server start
	if err := server.Start(ctx); err != nil {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

func parsePort(port string) int {
	p, err := net.LookupPort("udp", port)
	if err != nil || p < 1 || p > 65535 {
		return -1 // Return invalid port instead of defaulting
	}
	return p
}

// hasPermission checks if the program has permission to bind to a privileged port
func hasPermission(port string) bool {
	// Skip permission check in test mode
	if isTestMode {
		return true
	}
	// Test if sudo exists
	if _, err := exec.LookPath("sudo"); err != nil {
		return true
	}

	cmd := exec.Command("sudo", "-n", "true")
	if err := cmd.Run(); err != nil {
		fmt.Printf("Permission denied for port %s. Please run with elevated privileges.\n", port)
		return false
	}
	return true
}

// parseDNSQuery converts raw DNS query bytes into a human-readable format
func parseDNSQuery(data []byte) string {
	p := parser.New(data)
	result, err := p.ParseQuery()
	if err != nil {
		return fmt.Sprintf("Error parsing DNS query: %v", err)
	}
	return result
}

// Close closes the log file
func (d *DNSListener) Close() {
	d.logger.Close()
}

// initializeListener creates and initializes a new DNS listener with validation
func initializeListener(cfg *config.Config) (*DNSListener, error) {
	// Validate ports
	if port := parsePort(cfg.Port); port == -1 {
		return nil, fmt.Errorf("invalid DNS port: %s", cfg.Port)
	}
	if healthPort := parsePort(cfg.HealthPort); healthPort == -1 {
		return nil, fmt.Errorf("invalid health check port: %s", cfg.HealthPort)
	}

	fmt.Printf("Initializing with health check port: %s\n", cfg.HealthPort)

	listener, err := NewDNSListener(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize listener: %w", err)
	}

	// Initialize health check server if enabled
	if cfg.HealthPort != "" {
		healthServer := health.NewServer(cfg.HealthPort, listener.GetMetrics())
		go func() {
			if err := healthServer.Start(); err != nil {
				fmt.Printf("Health check server failed: %v\n", err)
			}
		}()
	}

	return listener, nil
}

// Update request channel initialization to use Request type
func (d *DNSListener) initRequestChannel(size int) {
	d.requestCh = make(chan types.Request, size)
}

func Run(port string) {
	cfg := config.DefaultConfig()
	cfg.Port = port

	if err := run(); err != nil {
		log.Fatalf("Error running DNS listener: %v", err)
	}
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("Fatal error: %v", err)
	}
}

func run() error {
	// Load configuration from environment
	cfg := config.LoadFromEnv()

	// Validate configuration
	if err := config.ValidateConfig(cfg); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Initialize listener
	listener, err := initializeListener(cfg)
	if err != nil {
		return fmt.Errorf("initialization error: %w", err)
	}
	defer listener.Close()

	// Setup signal handling for graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Start listener in a goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- listener.Start()
	}()

	// Wait for either error or shutdown signal
	select {
	case err := <-errChan:
		if err != nil {
			return fmt.Errorf("listener error: %w", err)
		}
	case <-stop:
		fmt.Println("\nReceived shutdown signal")
		cleanup(listener)
	}

	return nil
}

func cleanup(d *DNSListener) {
	if d.cache != nil {
		d.cache.Cleanup()
	}
}
