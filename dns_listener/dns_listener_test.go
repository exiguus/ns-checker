package dns_listener_test

import (
	"context"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/exiguus/ns-checker/dns_listener"
	"github.com/exiguus/ns-checker/dns_listener/config"
	"github.com/exiguus/ns-checker/internal/testflags"
)

func init() {
	testflags.Setup()
	config.SetTestMode(true)
}

type testConfig struct {
	tempDir string
	logFile string
	port    string
}

func setupTest(t *testing.T) (*testConfig, func()) {
	t.Helper()

	tempDir := t.TempDir()

	logFile := testflags.GetLogFile()
	if logFile == "" {
		logFile = filepath.Join(tempDir, "test.log")
	}

	tc := &testConfig{
		tempDir: tempDir,
		logFile: logFile,
		port:    "45353", // Use high port for tests
	}

	if err := os.MkdirAll(tempDir, 0755); err != nil {
		t.Fatal(err)
	}

	return tc, func() {} // Simplified cleanup function
}

func createTestConfig(tc *testConfig) *config.Config {
	return &config.Config{
		Port:                 tc.port,
		LogPath:              tc.logFile,
		WorkerCount:          4,
		CacheTTL:             time.Minute,
		RateLimit:            1000,
		RateBurst:            100,
		CacheCleanupInterval: time.Minute,
	}
}

func TestNewDNSListener(t *testing.T) {
	tc, cleanup := setupTest(t)
	defer cleanup()

	cfg := createTestConfig(tc)
	listener, err := dns_listener.NewDNSListener(cfg)
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	if listener == nil {
		t.Fatal("Expected listener to be non-nil")
	}
	defer listener.Close()
}

func TestRateLimiting(t *testing.T) {
	tc, cleanup := setupTest(t)
	defer cleanup()

	cfg := createTestConfig(tc)
	cfg.RateLimit = 1 // Only allow 1 request per second
	cfg.RateBurst = 1 // No burst

	listener, cancel := setupTestListener(t, cfg)
	defer cancel()
	defer listener.Close()

	addr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 12345}

	// Create a minimal valid DNS query
	query := []byte{
		0x00, 0x01, // ID
		0x01, 0x00, // Standard query
		0x00, 0x01, // One question
		0x00, 0x00, // No answers
		0x00, 0x00, // No authority
		0x00, 0x00, // No additional
		0x07, 'e', 'x', 'a', 'm', 'p', 'l', 'e',
		0x03, 'c', 'o', 'm',
		0x00,       // Root label
		0x00, 0x01, // Type A
		0x00, 0x01, // Class IN
	}

	resp1, err := listener.HandleRequest(query, addr, "UDP")
	if err != nil {
		t.Errorf("First request should succeed, got error: %v", err)
	}
	if resp1 == nil {
		t.Error("Expected response from first request, got nil")
	}

	resp2, err := listener.HandleRequest(query, addr, "UDP")
	if err == nil {
		t.Error("Second request should be rate limited")
	}
	if resp2 != nil {
		t.Error("Rate limited request should not return response")
	}

	// Wait for rate limit to reset
	time.Sleep(time.Second)

	resp3, err := listener.HandleRequest(query, addr, "UDP")
	if err != nil {
		t.Errorf("Third request should succeed after waiting, got error: %v", err)
	}
	if resp3 == nil {
		t.Error("Expected response from third request, got nil")
	}
}

func setupTestListener(t *testing.T, cfg *config.Config) (*dns_listener.DNSListener, context.CancelFunc) {
	t.Helper()
	_, cancel := context.WithCancel(context.Background())

	listener, err := dns_listener.NewDNSListener(cfg)
	if err != nil {
		cancel()
		t.Fatalf("Failed to create listener: %v", err)
	}

	return listener, cancel
}

func TestDNSListenerCache(t *testing.T) {
	// Setup with proper rate limits
	cfg := &config.Config{
		Port:                 "25353",
		LogPath:              "/tmp/dns.log",
		CacheTTL:             time.Minute,
		CacheCleanupInterval: time.Second * 30,
		RateLimit:            100, // Allow 100 requests per second
		RateBurst:            10,  // Allow burst of 10
		WorkerCount:          4,   // Add worker count
	}

	listener, err := dns_listener.NewDNSListener(cfg)
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()

	// Create test data - use proper DNS query format
	testQuery := []byte{
		0x00, 0x01, // ID
		0x01, 0x00, // Standard query
		0x00, 0x01, // One question
		0x00, 0x00, // No answers
		0x00, 0x00, // No authority
		0x00, 0x00, // No additional
		0x07, 'e', 'x', 'a', 'm', 'p', 'l', 'e',
		0x03, 'c', 'o', 'm',
		0x00,       // Root label
		0x00, 0x01, // Type A
		0x00, 0x01, // Class IN
	}

	testAddr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 25353}

	// First request - should miss cache
	resp1, err := listener.HandleRequest(testQuery, testAddr, "UDP")
	if err != nil {
		t.Fatalf("First request failed: %v", err)
	}
	if resp1 == nil {
		t.Fatal("Expected response from first request, got nil")
	}

	// Second request with same query - should hit cache
	resp2, err := listener.HandleRequest(testQuery, testAddr, "UDP")
	if err != nil {
		t.Fatalf("Second request failed: %v", err)
	}
	if resp2 == nil {
		t.Fatal("Expected response from second request, got nil")
	}
	if !reflect.DeepEqual(resp1, resp2) {
		t.Error("Cache hit should return same response")
	}

	// Get cache stats
	stats := listener.Cache().Stats()
	if stats.Hits != int64(1) {
		t.Errorf("Expected 1 cache hit, got %d", stats.Hits)
	}
	if stats.Misses != int64(1) {
		t.Errorf("Expected 1 cache miss, got %d", stats.Misses)
	}
}

func TestCacheExpiration(t *testing.T) {
	cfg := &config.Config{
		Port:                 "53",
		LogPath:              "/tmp/dns.log",
		CacheTTL:             time.Millisecond * 100,
		CacheCleanupInterval: time.Millisecond * 50,
		RateLimit:            100, // Allow 100 requests per second
		RateBurst:            10,  // Allow burst of 10
		WorkerCount:          4,   // Add worker count
	}

	listener, err := dns_listener.NewDNSListener(cfg)
	if err != nil {
		t.Fatalf("Failed to create listener: %v", err)
	}
	defer listener.Close()

	// Create a proper DNS query format
	testQuery := []byte{
		0x00, 0x01, // ID
		0x01, 0x00, // Standard query
		0x00, 0x01, // One question
		0x00, 0x00, // No answers
		0x00, 0x00, // No authority
		0x00, 0x00, // No additional
		0x04, 't', 'e', 's', 't',
		0x07, 'e', 'x', 'a', 'm', 'p', 'l', 'e',
		0x03, 'c', 'o', 'm',
		0x00,       // Root label
		0x00, 0x01, // Type A
		0x00, 0x01, // Class IN
	}

	testAddr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 12345}

	// First request
	resp1, err := listener.HandleRequest(testQuery, testAddr, "UDP")
	if err != nil {
		t.Fatalf("First request failed: %v", err)
	}
	if resp1 == nil {
		t.Fatal("Expected response from first request, got nil")
	}

	// Wait for cache to expire
	time.Sleep(time.Millisecond * 150)

	// Request after expiration - should miss cache
	resp2, err := listener.HandleRequest(testQuery, testAddr, "UDP")
	if err != nil {
		t.Fatalf("Second request failed: %v", err)
	}
	if resp2 == nil {
		t.Fatal("Expected response from second request, got nil")
	}

	// Get cache stats
	stats := listener.Cache().Stats()
	if stats.Hits != int64(0) {
		t.Errorf("Expected 0 cache hits after expiration, got %d", stats.Hits)
	}
	if stats.Misses != int64(2) {
		t.Errorf("Expected 2 cache misses, got %d", stats.Misses)
	}
}
