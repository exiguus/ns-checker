package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

const (
	envDNSPort       = "DNS_PORT"
	envWorkerCount   = "WORKER_COUNT"
	envRateLimit     = "RATE_LIMIT"
	envRateBurst     = "RATE_BURST"
	envCacheTTL      = "CACHE_TTL"
	envCacheCleanup  = "CACHE_CLEANUP"
	envHealthPort    = "HEALTH_CHECK_PORT"
	envLogsDir       = "LOGS_DIR"
	envLogFile       = "LOG_FILE"
	envDebug         = "DEBUG"
	envLogMaxSize    = "LOG_MAX_SIZE"
	envLogMaxBackups = "LOG_MAX_BACKUPS"
	envLogMaxAge     = "LOG_MAX_AGE"
)

// Default values
const (
	DefaultDNSPort         = "25353"
	DefaultHealthPort      = "8088"
	DefaultMaxWorkers      = "4"
	DefaultCacheTTL        = "30m"
	DefaultCleanupInterval = "1m"
	DefaultRateLimit       = "100000"
	DefaultRateBurst       = "1000"
	DefaultLogDir          = "./logs"
	DefaultLogFile         = "dns_listener.log"
	DefaultLogMaxSize      = 10 // MB
	DefaultLogMaxBackups   = 3  // files
	DefaultLogMaxAge       = 30 // days
)

type Config struct {
	Port                 string
	WorkerCount          int
	CacheTTL             time.Duration
	CacheCleanupInterval time.Duration
	LogsDir              string
	LogPath              string
	RateLimit            float64
	RateBurst            int
	HealthPort           string
	Debug                bool
	LogMaxSize           int // Maximum size in megabytes before rotation
	LogMaxBackups        int // Maximum number of old log files to retain
	LogMaxAge            int // Maximum days to retain old log files
}

// Add a flag for testing mode
var isTesting = false

// SetTestMode enables or disables testing mode (disables logging)
func SetTestMode(enabled bool) {
	isTesting = enabled
}

func DefaultConfig() *Config {
	// Get current working directory for default log path
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}

	// Create default log paths
	logDir := filepath.Join(cwd, DefaultLogDir)
	logPath := filepath.Join(logDir, DefaultLogFile)

	cfg := &Config{
		Port:                 "25353",
		WorkerCount:          4,
		RateLimit:            100000,
		RateBurst:            1000,
		CacheTTL:             30 * time.Minute,
		CacheCleanupInterval: time.Minute,
		HealthPort:           "8088",
		LogsDir:              logDir,
		LogPath:              logPath,
		LogMaxSize:           DefaultLogMaxSize,
		LogMaxBackups:        DefaultLogMaxBackups,
		LogMaxAge:            DefaultLogMaxAge,
		Debug:                false, // Add default Debug value
	}

	// Ensure log directory exists
	if err := os.MkdirAll(logDir, 0755); err != nil {
		fmt.Printf("Warning: Failed to create log directory: %v\n", err)
	}

	return cfg
}

func LoadFromEnv() *Config {
	cfg := DefaultConfig()

	cfg.Port = getEnvOrDefault(envDNSPort, cfg.Port)
	cfg.WorkerCount = getEnvAsInt(envWorkerCount, cfg.WorkerCount)
	cfg.RateLimit = getEnvAsFloat(envRateLimit, cfg.RateLimit)
	cfg.RateBurst = getEnvAsInt(envRateBurst, cfg.RateBurst)

	if ttl := os.Getenv(envCacheTTL); ttl != "" {
		if duration, err := time.ParseDuration(ttl); err == nil {
			cfg.CacheTTL = duration
		}
	}

	if cleanup := os.Getenv(envCacheCleanup); cleanup != "" {
		if duration, err := time.ParseDuration(cleanup); err == nil {
			cfg.CacheCleanupInterval = duration
		}
	}

	cfg.HealthPort = getEnvOrDefault(envHealthPort, cfg.HealthPort)

	// Handle log configuration
	if dir := os.Getenv(envLogsDir); dir != "" {
		cfg.LogsDir = dir
		cfg.LogPath = filepath.Join(dir, filepath.Base(cfg.LogPath))
	}

	if file := os.Getenv(envLogFile); file != "" {
		cfg.LogPath = filepath.Join(cfg.LogsDir, file)
	}

	// Log rotation settings
	cfg.LogMaxSize = getEnvAsInt(envLogMaxSize, cfg.LogMaxSize)
	cfg.LogMaxBackups = getEnvAsInt(envLogMaxBackups, cfg.LogMaxBackups)
	cfg.LogMaxAge = getEnvAsInt(envLogMaxAge, cfg.LogMaxAge)

	// Add Debug field loading
	cfg.Debug = getEnvAsBool(envDebug, cfg.Debug)

	// Remove any logging code here
	return cfg
}

// Helper functions
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	strValue := os.Getenv(key)
	if strValue == "" {
		return defaultValue
	}
	if value, err := strconv.Atoi(strValue); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsFloat(key string, defaultValue float64) float64 {
	strValue := os.Getenv(key)
	if strValue == "" {
		return defaultValue
	}
	if value, err := strconv.ParseFloat(strValue, 64); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	strValue := os.Getenv(key)
	if strValue == "" {
		return defaultValue
	}
	value, err := strconv.ParseBool(strValue)
	if err != nil {
		return defaultValue
	}
	return value
}

// Add new validation functions
func checkPortConflict(port string) error {
	p, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("invalid port number: %s", port)
	}
	if p < 1024 && os.Getuid() != 0 {
		return fmt.Errorf("port %d requires root privileges", p)
	}
	return nil
}

func checkFilePermissions(path string, requireWrite bool) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) && requireWrite {
			// Try to create directory
			return os.MkdirAll(path, 0755)
		}
		return err
	}

	if requireWrite {
		if info.IsDir() {
			// Check if directory is writable
			testFile := filepath.Join(path, ".permissions_test")
			f, err := os.Create(testFile)
			if err != nil {
				return fmt.Errorf("directory not writable: %v", err)
			}
			f.Close()
			os.Remove(testFile)
		} else {
			// Check if file's parent directory is writable
			dir := filepath.Dir(path)
			if err := checkFilePermissions(dir, true); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateRateLimits(rateLimit float64, rateBurst int) error {
	if rateLimit <= 0 {
		return fmt.Errorf("rate limit must be positive")
	}
	if rateBurst <= 0 {
		return fmt.Errorf("rate burst must be positive")
	}
	if float64(rateBurst) > rateLimit {
		return fmt.Errorf("rate burst (%d) cannot be greater than rate limit (%.0f)", rateBurst, rateLimit)
	}
	// Add sanity checks
	if rateLimit > 1000000 {
		return fmt.Errorf("rate limit exceeds maximum allowed value (1,000,000)")
	}
	if rateBurst > 10000 {
		return fmt.Errorf("rate burst exceeds maximum allowed value (10,000)")
	}
	return nil
}

// Update ValidateConfig function to use local error types
func ValidateConfig(config *Config) error {
	var errors []error

	// Port availability checks
	portChecker := NewPortChecker(5 * time.Second)

	// Validate DNS port
	if config.Port != "" {
		if err := portChecker.IsPortAvailable(config.Port); err != nil {
			errors = append(errors, NewConfigError("Port", config.Port, err.Error()))
		}
	}

	// Validate health check port
	if config.HealthPort != "" {
		if err := portChecker.IsPortAvailable(config.HealthPort); err != nil {
			errors = append(errors, NewConfigError("HealthPort", config.HealthPort, err.Error()))
		}

		// Check for port conflict between DNS and health check ports
		if config.Port == config.HealthPort {
			errors = append(errors, NewConfigError("HealthPort", config.HealthPort,
				"health check port cannot be the same as DNS port"))
		}
	}

	// Worker count validation
	if config.WorkerCount < 1 || config.WorkerCount > 128 {
		errors = append(errors, NewConfigError("WorkerCount",
			config.WorkerCount,
			fmt.Sprintf("must be between 1 and 128, got %d", config.WorkerCount)))
	}

	// Rate limit validation
	if config.RateLimit <= 0 || config.RateLimit > 1000000 {
		errors = append(errors, ErrInvalidRateLimit(config.RateLimit))
	}
	if config.RateBurst <= 0 || config.RateBurst > 10000 {
		errors = append(errors, ErrInvalidRateBurst(config.RateBurst))
	}
	if float64(config.RateBurst) > config.RateLimit {
		errors = append(errors, NewConfigError("RateBurst", config.RateBurst,
			fmt.Sprintf("cannot be greater than rate limit (%.0f)", config.RateLimit)))
	}

	// Cache settings validation
	if config.CacheTTL <= 0 {
		errors = append(errors, ErrInvalidTTL(config.CacheTTL.String()))
	}
	if config.CacheCleanupInterval > config.CacheTTL {
		errors = append(errors, ErrInvalidCleanup(config.CacheCleanupInterval.String()))
	}

	// Log settings validation
	if config.LogMaxSize < 1 || config.LogMaxSize > 1024 {
		errors = append(errors, ErrInvalidLogSize(config.LogMaxSize))
	}

	// Remove logging and just return the error if any
	if len(errors) > 0 {
		return &ValidationError{Errors: errors}
	}

	return nil
}
