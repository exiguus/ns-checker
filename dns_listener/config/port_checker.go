package config

import (
	"fmt"
	"net"
	"strconv"
	"time"
)

// PortChecker provides methods to validate port configuration
type PortChecker interface {
	IsPortAvailable(port string) error
	IsPortInUse(port string) bool
}

// DefaultPortChecker implements PortChecker interface
type DefaultPortChecker struct {
	timeout time.Duration
}

// Make NewPortChecker variable for testing
var NewPortChecker = func(timeout time.Duration) PortChecker {
	return &DefaultPortChecker{
		timeout: timeout,
	}
}

// IsPortAvailable checks if a port is valid and available
func (pc *DefaultPortChecker) IsPortAvailable(port string) error {
	portNum, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("invalid port number format: %s", port)
	}

	if portNum < 1 || portNum > 65535 {
		return fmt.Errorf("port number %d out of valid range (1-65535)", portNum)
	}

	if pc.IsPortInUse(port) {
		return fmt.Errorf("port %s is already in use", port)
	}

	return nil
}

// IsPortInUse checks if a port is currently in use
func (pc *DefaultPortChecker) IsPortInUse(port string) bool {
	addr := fmt.Sprintf(":%s", port)
	conn, err := net.DialTimeout("tcp", addr, pc.timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}
