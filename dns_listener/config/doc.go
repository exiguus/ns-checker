/*
Package config provides configuration management for the DNS listener service.

The package handles configuration through environment variables and provides validation
for all settings. It supports configuration for:

  - DNS and health check ports
  - Worker pool size
  - Rate limiting
  - Cache settings
  - Log rotation
  - Debug mode

Example usage:

	cfg := config.LoadFromEnv()
	if err := config.ValidateConfig(cfg); err != nil {
	    log.Fatalf("Invalid configuration: %v", err)
	}

Environment Variables:

	DNS_PORT          - DNS server port (default: 25353)
	WORKER_COUNT      - Number of workers (default: 4)
	RATE_LIMIT        - Rate limit per second (default: 100000)
	RATE_BURST        - Rate limit burst (default: 1000)
	CACHE_TTL         - Cache time-to-live (default: 30m)
	CACHE_CLEANUP     - Cache cleanup interval (default: 1m)
	HEALTH_CHECK_PORT - Health check port (default: 8088)
	LOGS_DIR         - Log directory (default: ./logs)
	LOG_FILE         - Log file name (default: dns_listener.log)
	LOG_MAX_SIZE     - Maximum log file size in MB (default: 10)
	LOG_MAX_BACKups  - Maximum number of old log files (default: 3)
	LOG_MAX_AGE      - Maximum age of old log files in days (default: 30)
	DEBUG            - Enable debug mode (default: false)
*/
package config
