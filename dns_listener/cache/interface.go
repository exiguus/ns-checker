package cache

import (
	"fmt"
	"time"
)

type EvictionPolicy int

const (
	LRU EvictionPolicy = iota
	LFU
)

type Cache interface {
	Get(key string) ([]byte, bool)
	Set(key string, value []byte, ttl time.Duration)
	Delete(key string)
	Cleanup()
	Stats() Stats
}

type Stats struct {
	Size          int
	BytesInMemory uint64
	Hits          int64
	Misses        int64
	Evictions     int64
}

type Config struct {
	MaxSize         int64
	DefaultTTL      time.Duration
	CleanupInterval time.Duration
	EvictionPolicy  EvictionPolicy
}

func DefaultConfig() Config {
	return Config{
		MaxSize:         1024 * 1024 * 100, // 100MB
		DefaultTTL:      30 * time.Minute,
		CleanupInterval: time.Minute,
	}
}

func ValidateConfig(cfg *Config) error {
	if cfg.MaxSize <= 0 {
		return fmt.Errorf("max size must be positive")
	}
	if cfg.DefaultTTL <= 0 {
		return fmt.Errorf("default TTL must be positive")
	}
	if cfg.CleanupInterval <= 0 {
		return fmt.Errorf("cleanup interval must be positive")
	}
	return nil
}
