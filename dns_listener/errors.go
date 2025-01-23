package dns_listener

import "fmt"

type DNSError struct {
	Op  string
	Err error
}

func (e *DNSError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("dns operation %s failed: %v", e.Op, e.Err)
	}
	return fmt.Sprintf("dns operation %s failed", e.Op)
}

type ConfigError struct {
	Field string
	Err   error
}

func (e *ConfigError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("invalid configuration for %s: %v", e.Field, e.Err)
	}
	return fmt.Sprintf("invalid configuration for %s", e.Field)
}
