package dns_listener

import (
	"time"
)

type Logger interface {
	Write(string)
	Error(msg string, err error)
	LogRequest(protocol, client string, data []byte, err error)
	Close()
}

type Request struct {
	Data       []byte
	ClientAddr string
	Protocol   string
}

const (
	DefaultPort            = "25353"
	DefaultCacheTTL        = 5 * time.Minute
	DefaultCleanupInterval = time.Minute
	DefaultRateLimit       = 1000
	DefaultRateBurst       = 100
)
