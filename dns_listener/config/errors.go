package config

import "fmt"

// ValidationError represents multiple configuration validation errors
type ValidationError struct {
	Errors []error
}

func (v *ValidationError) Error() string {
	return fmt.Sprintf("validation failed with %d errors", len(v.Errors))
}

// ConfigError represents a single configuration error
type ConfigError struct {
	Field   string
	Value   interface{}
	Message string
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("%s: %v - %s", e.Field, e.Value, e.Message)
}

// Error constructors
func NewConfigError(field string, value interface{}, message string) error {
	return &ConfigError{Field: field, Value: value, Message: message}
}

func ErrInvalidPort(port string) error {
	return NewConfigError("Port", port, "invalid port number")
}

func ErrPrivilegedPort(port string) error {
	return NewConfigError("Port", port, "requires root privileges")
}

func ErrPortConflict(port string) error {
	return NewConfigError("Port", port, "port already in use")
}

func ErrInvalidWorkers(count int) error {
	return NewConfigError("WorkerCount", count, "invalid worker count (must be between 1 and 128)")
}

func ErrInvalidRateLimit(limit float64) error {
	return NewConfigError("RateLimit", limit, "invalid rate limit (must be between 1 and 1,000,000)")
}

func ErrInvalidRateBurst(burst int) error {
	return NewConfigError("RateBurst", burst, "invalid rate burst (must be between 1 and 10,000)")
}

func ErrInvalidTTL(ttl string) error {
	return NewConfigError("CacheTTL", ttl, "invalid cache TTL (must be positive duration)")
}

func ErrInvalidCleanup(cleanup string) error {
	return NewConfigError("CacheCleanupInterval", cleanup, "cleanup interval must be less than TTL")
}

func ErrInvalidLogSize(size int) error {
	return NewConfigError("LogMaxSize", size, "invalid log size (must be between 1 and 1024 MB)")
}
