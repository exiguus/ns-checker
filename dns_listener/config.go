package dns_listener

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/exiguus/ns-checker/dns_listener/config"
)

func ValidateConfig(cfg *config.Config) error {
	if cfg.Port != "" {
		if port := parsePort(cfg.Port); port == -1 {
			return &ConfigError{Field: "Port", Err: fmt.Errorf("invalid port: %s", cfg.Port)}
		}
	}

	if cfg.CacheTTL <= 0 {
		return &ConfigError{Field: "CacheTTL", Err: fmt.Errorf("must be positive")}
	}

	if cfg.CacheCleanupInterval <= 0 {
		return &ConfigError{Field: "CacheCleanupInterval", Err: fmt.Errorf("must be positive")}
	}

	if cfg.LogPath == "" {
		return &ConfigError{Field: "LogPath", Err: fmt.Errorf("log path cannot be empty")}
	}

	logDir := filepath.Dir(cfg.LogPath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return &ConfigError{Field: "LogPath", Err: fmt.Errorf("failed to create log directory: %v", err)}
	}

	testFile := filepath.Join(logDir, ".test")
	if f, err := os.Create(testFile); err != nil {
		return &ConfigError{Field: "LogPath", Err: fmt.Errorf("directory not writable: %v", err)}
	} else {
		f.Close()
		os.Remove(testFile)
	}

	return nil
}

func SetConfigDefaults(cfg *config.Config) {
	if cfg.Port == "" {
		cfg.Port = DefaultPort
	}
	if cfg.WorkerCount == 0 {
		cfg.WorkerCount = calculateOptimalWorkers()
	}
	if cfg.CacheTTL == 0 {
		cfg.CacheTTL = DefaultCacheTTL
	}
	if cfg.CacheCleanupInterval == 0 {
		cfg.CacheCleanupInterval = DefaultCleanupInterval
	}
	if cfg.RateLimit == 0 {
		cfg.RateLimit = DefaultRateLimit
	}
	if cfg.RateBurst == 0 {
		cfg.RateBurst = DefaultRateBurst
	}

	if cfg.LogPath == "" {
		// Use current working directory + logs
		cwd, err := os.Getwd()
		if err != nil {
			cwd = "."
		}
		cfg.LogPath = filepath.Join(cwd, "logs", "dns_listener.log")
	}

	if !filepath.IsAbs(cfg.LogPath) {
		if absPath, err := filepath.Abs(cfg.LogPath); err == nil {
			cfg.LogPath = absPath
		}
	}
}
