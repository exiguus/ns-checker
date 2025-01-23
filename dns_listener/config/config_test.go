package config

import (
	"fmt"
	"os"
	"testing"
	"time"
)

// List of all environment variables used in tests
var allEnvVars = []string{
	"DNS_PORT",
	"WORKER_COUNT",
	"RATE_LIMIT",
	"RATE_BURST",
	"CACHE_TTL",
	"CACHE_CLEANUP",
	"HEALTH_CHECK_PORT",
	"LOGS_DIR",
	"LOG_FILE",
	"LOG_MAX_SIZE",
	"LOG_MAX_BACKUPS",
	"LOG_MAX_AGE",
	"DEBUG",
}

func cleanEnvironment() {
	for _, env := range allEnvVars {
		os.Unsetenv(env)
	}
}

func TestLoadFromEnv(t *testing.T) {
	// Clean environment before all tests
	cleanEnvironment()

	tests := []struct {
		name     string
		envVars  map[string]string
		expected *Config
	}{
		{
			name:    "default values when no env vars",
			envVars: map[string]string{},
			expected: &Config{
				Port:                 "25353",            // Match actual default
				WorkerCount:          4,                  // Match actual default
				RateLimit:            100000,             // Match actual default
				RateBurst:            1000,               // Match actual default
				CacheTTL:             30 * time.Minute,   // Match actual default
				CacheCleanupInterval: time.Minute,        // Match actual default
				HealthPort:           "8088",             // Match actual default
				LogPath:              "dns_listener.log", // Match actual default
				LogsDir:              "./logs",
				LogMaxSize:           10,    // DefaultLogMaxSize
				LogMaxBackups:        3,     // DefaultLogMaxBackups
				LogMaxAge:            30,    // DefaultLogMaxAge
				Debug:                false, // Add default Debug value
			},
		},
		{
			name: "custom values from env vars",
			envVars: map[string]string{
				"DNS_PORT":          "45353",
				"WORKER_COUNT":      "8",
				"RATE_LIMIT":        "200",
				"RATE_BURST":        "400",
				"CACHE_TTL":         "10m",
				"CACHE_CLEANUP":     "20m",
				"HEALTH_CHECK_PORT": "9090",
			},
			expected: &Config{
				Port:                 "45353",
				WorkerCount:          8,
				RateLimit:            200,
				RateBurst:            400,
				CacheTTL:             10 * time.Minute,
				CacheCleanupInterval: 20 * time.Minute,
				HealthPort:           "9090",
				LogPath:              "dns_listener.log",
				LogsDir:              "./logs",
				LogMaxSize:           10, // DefaultLogMaxSize
				LogMaxBackups:        3,  // DefaultLogMaxBackups
				LogMaxAge:            30, // DefaultLogMaxAge
			},
		},
		{
			name: "custom log rotation settings",
			envVars: map[string]string{
				"LOG_MAX_SIZE":    "50",
				"LOG_MAX_BACKUPS": "5",
				"LOG_MAX_AGE":     "7",
			},
			expected: &Config{
				Port:                 "25353",            // Match actual default
				WorkerCount:          4,                  // Match actual default
				RateLimit:            100000,             // Match actual default
				RateBurst:            1000,               // Match actual default
				CacheTTL:             30 * time.Minute,   // Match actual default
				CacheCleanupInterval: time.Minute,        // Match actual default
				HealthPort:           "8088",             // Match actual default
				LogPath:              "dns_listener.log", // Match actual default
				LogsDir:              "./logs",
				LogMaxSize:           50,
				LogMaxBackups:        5,
				LogMaxAge:            7,
			},
		},
		{
			name: "debug mode enabled",
			envVars: map[string]string{
				"DEBUG": "true",
			},
			expected: &Config{
				Port:                 "25353",            // Match actual default
				WorkerCount:          4,                  // Match actual default
				RateLimit:            100000,             // Match actual default
				RateBurst:            1000,               // Match actual default
				CacheTTL:             30 * time.Minute,   // Match actual default
				CacheCleanupInterval: time.Minute,        // Match actual default
				HealthPort:           "8088",             // Match actual default
				LogPath:              "dns_listener.log", // Match actual default
				LogsDir:              "./logs",
				LogMaxSize:           10,
				LogMaxBackups:        3,
				LogMaxAge:            30,
				Debug:                true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean environment before each test
			cleanEnvironment()

			// Set test environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			cfg := LoadFromEnv()

			// Compare fields
			if cfg.Port != tt.expected.Port {
				t.Errorf("Port = %s, want %s", cfg.Port, tt.expected.Port)
			}
			if cfg.WorkerCount != tt.expected.WorkerCount {
				t.Errorf("WorkerCount = %d, want %d", cfg.WorkerCount, tt.expected.WorkerCount)
			}
			if cfg.RateLimit != tt.expected.RateLimit {
				t.Errorf("RateLimit = %f, want %f", cfg.RateLimit, tt.expected.RateLimit)
			}
			if cfg.RateBurst != tt.expected.RateBurst {
				t.Errorf("RateBurst = %d, want %d", cfg.RateBurst, tt.expected.RateBurst)
			}
			if cfg.CacheTTL != tt.expected.CacheTTL {
				t.Errorf("CacheTTL = %v, want %v", cfg.CacheTTL, tt.expected.CacheTTL)
			}
			if cfg.CacheCleanupInterval != tt.expected.CacheCleanupInterval {
				t.Errorf("CacheCleanupInterval = %v, want %v", cfg.CacheCleanupInterval, tt.expected.CacheCleanupInterval)
			}
			if cfg.HealthPort != tt.expected.HealthPort {
				t.Errorf("HealthPort = %s, want %s", cfg.HealthPort, tt.expected.HealthPort)
			}
			if cfg.LogMaxSize != tt.expected.LogMaxSize {
				t.Errorf("LogMaxSize = %d, want %d", cfg.LogMaxSize, tt.expected.LogMaxSize)
			}
			if cfg.LogMaxBackups != tt.expected.LogMaxBackups {
				t.Errorf("LogMaxBackups = %d, want %d", cfg.LogMaxBackups, tt.expected.LogMaxBackups)
			}
			if cfg.LogMaxAge != tt.expected.LogMaxAge {
				t.Errorf("LogMaxAge = %d, want %d", cfg.LogMaxAge, tt.expected.LogMaxAge)
			}

			// Add Debug field comparison
			if cfg.Debug != tt.expected.Debug {
				t.Errorf("Debug = %v, want %v", cfg.Debug, tt.expected.Debug)
			}

			// Clean up after test
			cleanEnvironment()
		})
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Port:                 "8053", // Changed from "53" to non-privileged port
				WorkerCount:          4,
				RateLimit:            100,
				RateBurst:            50,
				CacheTTL:             5 * time.Minute,
				CacheCleanupInterval: time.Minute,
				HealthPort:           "9999",
				LogPath:              "./test-log-path.log",
				LogMaxSize:           100,
				LogMaxBackups:        3,
				LogMaxAge:            30,
			},
			wantErr: false,
		},
		{
			name: "invalid port",
			config: &Config{
				Port: "999999",
			},
			wantErr: true,
		},
		{
			name: "invalid worker count",
			config: &Config{
				Port:        "53",
				WorkerCount: 0,
			},
			wantErr: true,
		},
		{
			name: "invalid cache TTL",
			config: &Config{
				Port:        "53",
				WorkerCount: 4,
				CacheTTL:    -1 * time.Minute,
			},
			wantErr: true,
		},
		{
			name: "invalid log max size",
			config: &Config{
				Port:        "53",
				WorkerCount: 4,
				LogMaxSize:  0,
			},
			wantErr: true,
		},
		{
			name: "negative log max size",
			config: &Config{
				Port:        "53",
				WorkerCount: 4,
				LogMaxSize:  -1,
			},
			wantErr: true,
		},
		{
			name: "invalid log max backups",
			config: &Config{
				Port:          "53",
				WorkerCount:   4,
				LogMaxBackups: -1,
				LogMaxSize:    10,
				LogPath:       "./test.log",
				CacheTTL:      time.Minute,
			},
			wantErr: true,
		},
		{
			name: "invalid log max age",
			config: &Config{
				Port:        "53",
				WorkerCount: 4,
				LogMaxAge:   -1,
				LogMaxSize:  10,
				LogPath:     "./test.log",
				CacheTTL:    time.Minute,
			},
			wantErr: true,
		},
		{
			name: "invalid worker count zero",
			config: &Config{
				Port:        "53",
				WorkerCount: 0,
				LogMaxSize:  10,
				LogPath:     "./test.log",
				CacheTTL:    time.Minute,
			},
			wantErr: true,
		},
		{
			name: "invalid health check port",
			config: &Config{
				Port:        "53",
				WorkerCount: 4,
				LogMaxSize:  10,
				LogPath:     "./test.log",
				CacheTTL:    time.Minute,
				HealthPort:  "999999",
			},
			wantErr: true,
		},
		{
			name: "invalid health check port format",
			config: &Config{
				Port:        "53",
				WorkerCount: 4,
				LogMaxSize:  10,
				LogPath:     "./test.log",
				CacheTTL:    time.Minute,
				HealthPort:  "abc",
			},
			wantErr: true,
		},
		{
			name: "invalid rate limit",
			config: &Config{
				Port:        "53",
				WorkerCount: 4,
				RateLimit:   -1,
				LogMaxSize:  10,
				LogPath:     "./test.log",
				CacheTTL:    time.Minute,
			},
			wantErr: true,
		},
		{
			name: "invalid rate burst",
			config: &Config{
				Port:        "53",
				WorkerCount: 4,
				RateLimit:   100,
				RateBurst:   -1,
				LogMaxSize:  10,
				LogPath:     "./test.log",
				CacheTTL:    time.Minute,
			},
			wantErr: true,
		},
		{
			name: "rate burst greater than rate limit",
			config: &Config{
				Port:        "53",
				WorkerCount: 4,
				RateLimit:   100,
				RateBurst:   200,
				LogMaxSize:  10,
				LogPath:     "./test.log",
				CacheTTL:    time.Minute,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// Helper function to test configuration validity
func validateFullConfig(t *testing.T, cfg *Config) {
	err := ValidateConfig(cfg)
	if err != nil {
		t.Errorf("Expected valid config but got error: %v", err)
	}
}

// Helper function to verify rate limits
func checkRateLimits(t *testing.T, cfg *Config, expectedLimit float64, expectedBurst int) {
	if cfg.RateLimit != expectedLimit {
		t.Errorf("RateLimit = %.0f, want %.0f", cfg.RateLimit, expectedLimit)
	}
	if cfg.RateBurst != expectedBurst {
		t.Errorf("RateBurst = %d, want %d", cfg.RateBurst, expectedBurst)
	}
	if float64(cfg.RateBurst) > cfg.RateLimit {
		t.Errorf("RateBurst (%d) is greater than RateLimit (%.0f)", cfg.RateBurst, cfg.RateLimit)
	}
}

// Complete the compareConfigs helper function
func compareConfigs(t *testing.T, got, want *Config) {
	t.Helper()
	if got.Port != want.Port {
		t.Errorf("Port = %s, want %s", got.Port, want.Port)
	}
	if got.WorkerCount != want.WorkerCount {
		t.Errorf("WorkerCount = %d, want %d", got.WorkerCount, want.WorkerCount)
	}
	if got.RateLimit != want.RateLimit {
		t.Errorf("RateLimit = %.0f, want %.0f", got.RateLimit, want.RateLimit)
	}
	if got.RateBurst != want.RateBurst {
		t.Errorf("RateBurst = %d, want %d", got.RateBurst, want.RateBurst)
	}
	if got.CacheTTL != want.CacheTTL {
		t.Errorf("CacheTTL = %v, want %v", got.CacheTTL, want.CacheTTL)
	}
	if got.CacheCleanupInterval != want.CacheCleanupInterval {
		t.Errorf("CacheCleanupInterval = %v, want %v", got.CacheCleanupInterval, want.CacheCleanupInterval)
	}
	if got.HealthPort != want.HealthPort {
		t.Errorf("HealthPort = %s, want %s", got.HealthPort, want.HealthPort)
	}
	if got.LogMaxSize != want.LogMaxSize {
		t.Errorf("LogMaxSize = %d, want %d", got.LogMaxSize, want.LogMaxSize)
	}
	if got.LogMaxBackups != want.LogMaxBackups {
		t.Errorf("LogMaxBackups = %d, want %d", got.LogMaxBackups, want.LogMaxBackups)
	}
	if got.LogMaxAge != want.LogMaxAge {
		t.Errorf("LogMaxAge = %d, want %d", got.LogMaxAge, want.LogMaxAge)
	}
	if got.Debug != want.Debug {
		t.Errorf("Debug = %v, want %v", got.Debug, want.Debug)
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	expected := &Config{
		Port:                 "25353",
		WorkerCount:          4,
		RateLimit:            100000,
		RateBurst:            1000,
		CacheTTL:             30 * time.Minute,
		CacheCleanupInterval: time.Minute,
		HealthPort:           "8088",
		LogPath:              cfg.LogPath,
		LogsDir:              cfg.LogsDir,
		LogMaxSize:           10,
		LogMaxBackups:        3,
		LogMaxAge:            30,
		Debug:                false,
	}

	compareConfigs(t, cfg, expected)
	validateFullConfig(t, cfg)
	checkRateLimits(t, cfg, expected.RateLimit, expected.RateBurst)
}

func TestInvalidEnvironmentVariables(t *testing.T) {
	tests := []struct {
		name     string
		envVars  map[string]string
		checkErr func(*Config) bool
	}{
		{
			name: "invalid worker count format",
			envVars: map[string]string{
				"WORKER_COUNT": "invalid",
			},
			checkErr: func(cfg *Config) bool {
				return cfg.WorkerCount == 4 // Should keep default value
			},
		},
		{
			name: "invalid rate limit format",
			envVars: map[string]string{
				"RATE_LIMIT": "invalid",
			},
			checkErr: func(cfg *Config) bool {
				return cfg.RateLimit == 100000 // Should keep default value
			},
		},
		{
			name: "invalid cache TTL format",
			envVars: map[string]string{
				"CACHE_TTL": "invalid",
			},
			checkErr: func(cfg *Config) bool {
				return cfg.CacheTTL == 30*time.Minute // Should keep default value
			},
		},
		{
			name: "invalid debug format",
			envVars: map[string]string{
				"DEBUG": "not-a-bool",
			},
			checkErr: func(cfg *Config) bool {
				return cfg.Debug == false // Should keep default value
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanEnvironment()
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			cfg := LoadFromEnv()
			if !tt.checkErr(cfg) {
				t.Errorf("LoadFromEnv() did not handle invalid environment variable correctly")
			}

			cleanEnvironment()
		})
	}
}

func BenchmarkLoadFromEnv(b *testing.B) {
	os.Setenv("DNS_PORT", "8053")
	os.Setenv("WORKER_COUNT", "8")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LoadFromEnv()
	}
}

func BenchmarkValidateConfig(b *testing.B) {
	cfg := DefaultConfig()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateConfig(cfg)
	}
}

func ExampleLoadFromEnv() {
	os.Setenv("DNS_PORT", "53")
	os.Setenv("WORKER_COUNT", "8")
	os.Setenv("DEBUG", "true")

	cfg := LoadFromEnv()
	fmt.Printf("DNS Port: %s\n", cfg.Port)
	fmt.Printf("Workers: %d\n", cfg.WorkerCount)
	fmt.Printf("Debug Mode: %v\n", cfg.Debug)

	// Clean up
	os.Unsetenv("DNS_PORT")
	os.Unsetenv("WORKER_COUNT")
	os.Unsetenv("DEBUG")

	// Output:
	// DNS Port: 53
	// Workers: 8
	// Debug Mode: true
}

// MockPortChecker for testing
type MockPortChecker struct{}

func (m *MockPortChecker) IsPortAvailable(port string) error {
	return nil // Always return success for tests
}

func (m *MockPortChecker) IsPortInUse(port string) bool {
	return false // Always return false for tests
}

func ExampleValidateConfig() {
	// Override the default port checker with our mock for the test
	origPortChecker := NewPortChecker
	NewPortChecker = func(timeout time.Duration) PortChecker {
		return &MockPortChecker{}
	}
	defer func() {
		NewPortChecker = origPortChecker // Restore original after test
	}()

	cfg := &Config{
		Port:                 "8053",
		WorkerCount:          4,
		RateLimit:            1000,
		RateBurst:            100,
		LogPath:              "./dns.log",
		CacheTTL:             5 * time.Minute,
		CacheCleanupInterval: time.Minute,
		LogMaxSize:           10,
		LogMaxBackups:        3,
		LogMaxAge:            30,
		HealthPort:           "8080",
	}

	err := ValidateConfig(cfg)
	fmt.Printf("Configuration valid: %v\n", err == nil)

	// Output:
	// Configuration valid: true
}

func BenchmarkLoadFromEnv_WithAllVariables(b *testing.B) {
	envVars := map[string]string{
		"DNS_PORT":          "8053",
		"WORKER_COUNT":      "8",
		"RATE_LIMIT":        "1000",
		"RATE_BURST":        "100",
		"CACHE_TTL":         "5m",
		"CACHE_CLEANUP":     "1m",
		"HEALTH_CHECK_PORT": "8080",
		"LOGS_DIR":          "/tmp/logs",
		"LOG_FILE":          "test.log",
		"LOG_MAX_SIZE":      "20",
		"LOG_MAX_BACKUPS":   "5",
		"LOG_MAX_AGE":       "7",
		"DEBUG":             "true",
	}

	for k, v := range envVars {
		os.Setenv(k, v)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		LoadFromEnv()
	}

	b.StopTimer()
	cleanEnvironment()
}

func BenchmarkValidateConfig_Parallel(b *testing.B) {
	cfg := DefaultConfig()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateConfig(cfg)
	}
}

func BenchmarkCompareConfigs(b *testing.B) {
	cfg1 := DefaultConfig()
	cfg2 := DefaultConfig()
	cfg2.Port = "8053"
	cfg2.WorkerCount = 8

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		compareConfigs(&testing.T{}, cfg1, cfg2)
	}
}

func BenchmarkValidateConfig_Scenarios(b *testing.B) {
	scenarios := map[string]*Config{
		"default": DefaultConfig(),
		"minimal": {
			Port:                 "53",
			WorkerCount:          4,
			RateLimit:            100,
			RateBurst:            50,
			CacheTTL:             time.Minute,
			CacheCleanupInterval: time.Minute,
			LogPath:              "./test.log",
			LogMaxSize:           10,
		},
		"maximal": {
			Port:                 "8053",
			WorkerCount:          32,
			RateLimit:            1000000,
			RateBurst:            10000,
			CacheTTL:             time.Hour,
			CacheCleanupInterval: 5 * time.Minute,
			LogPath:              "/var/log/dns.log",
			LogMaxSize:           100,
			LogMaxBackups:        10,
			LogMaxAge:            90,
			HealthPort:           "8080",
			Debug:                true,
		},
	}

	for name, cfg := range scenarios {
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				ValidateConfig(cfg)
			}
		})
	}
}

func TestValidateConfig_AdvancedChecks(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "privileged port without root",
			config: &Config{
				Port:                 "80",
				WorkerCount:          4,
				RateLimit:            1000,
				RateBurst:            100,
				CacheTTL:             time.Minute,
				CacheCleanupInterval: time.Minute,
				LogPath:              "./test.log",
				LogMaxSize:           10,
				LogMaxBackups:        3,
				LogMaxAge:            30,
			},
			wantErr: true,
		},
		{
			name: "same ports for DNS and health check",
			config: &Config{
				Port:                 "8053",
				HealthPort:           "8053",
				WorkerCount:          4,
				RateLimit:            1000,
				RateBurst:            100,
				CacheTTL:             time.Minute,
				CacheCleanupInterval: time.Minute,
				LogPath:              "./test.log",
				LogMaxSize:           10,
				LogMaxBackups:        3,
				LogMaxAge:            30,
			},
			wantErr: true,
		},
		{
			name: "excessive worker count",
			config: &Config{
				Port:                 "8053",
				WorkerCount:          256,
				RateLimit:            1000,
				RateBurst:            100,
				CacheTTL:             time.Minute,
				CacheCleanupInterval: time.Minute,
				LogPath:              "./test.log",
				LogMaxSize:           10,
				LogMaxBackups:        3,
				LogMaxAge:            30,
			},
			wantErr: true,
		},
		{
			name: "excessive rate limit",
			config: &Config{
				Port:                 "8053",
				WorkerCount:          4,
				RateLimit:            2000000,
				RateBurst:            1000,
				CacheTTL:             time.Minute,
				CacheCleanupInterval: time.Minute,
				LogPath:              "./test.log",
				LogMaxSize:           10,
				LogMaxBackups:        3,
				LogMaxAge:            30,
			},
			wantErr: true,
		},
		{
			name: "cleanup interval greater than TTL",
			config: &Config{
				Port:                 "8053",
				WorkerCount:          4,
				RateLimit:            1000,
				RateBurst:            100,
				CacheTTL:             time.Minute,
				CacheCleanupInterval: 2 * time.Minute,
				LogPath:              "./test.log",
				LogMaxSize:           10,
				LogMaxBackups:        3,
				LogMaxAge:            30,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConfig(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func BenchmarkConfig_Parallel(b *testing.B) {
	scenarios := map[string]struct {
		setup    func() *Config
		validate bool
	}{
		"load-env": {
			setup: func() *Config {
				os.Setenv("DNS_PORT", "8053")
				os.Setenv("WORKER_COUNT", "8")
				return nil
			},
			validate: false,
		},
		"validate-minimal": {
			setup: func() *Config {
				return &Config{
					Port:                 "8053",
					WorkerCount:          4,
					RateLimit:            100,
					RateBurst:            50,
					CacheTTL:             time.Minute,
					CacheCleanupInterval: time.Minute,
					LogPath:              "./test.log",
					LogMaxSize:           10,
				}
			},
			validate: true,
		},
		"validate-full": {
			setup: func() *Config {
				return &Config{
					Port:                 "8053",
					WorkerCount:          32,
					RateLimit:            1000,
					RateBurst:            100,
					CacheTTL:             time.Hour,
					CacheCleanupInterval: 5 * time.Minute,
					LogPath:              "./test.log",
					LogMaxSize:           100,
					LogMaxBackups:        10,
					LogMaxAge:            90,
					HealthPort:           "8080",
					Debug:                true,
				}
			},
			validate: true,
		},
	}

	for name, sc := range scenarios {
		b.Run(name, func(b *testing.B) {
			if sc.setup != nil {
				cfg := sc.setup()
				b.ResetTimer()
				b.RunParallel(func(pb *testing.PB) {
					for pb.Next() {
						if sc.validate {
							ValidateConfig(cfg)
						} else {
							LoadFromEnv()
						}
					}
				})
			}
		})
	}
}

func BenchmarkConfigInitialization(b *testing.B) {
	tests := []struct {
		name string
		init func() *Config
	}{
		{
			name: "default",
			init: DefaultConfig,
		},
		{
			name: "custom",
			init: func() *Config {
				return &Config{
					Port:        "8053",
					WorkerCount: 4,
					RateLimit:   1000,
					RateBurst:   100,
				}
			},
		},
	}

	for _, tt := range tests {
		b.Run(tt.name, func(b *testing.B) {
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					_ = tt.init()
				}
			})
		})
	}
}

// Add cleanup for benchmarks
func init() {
	// Ensure test directories exist
	os.MkdirAll("./testdata", 0755)
}

func BenchmarkCleanup(b *testing.B) {
	if err := os.RemoveAll("./testdata"); err != nil {
		b.Fatal("Failed to clean up test data:", err)
	}
}
