package errors

import "fmt"

// ErrorType represents the type of DNS error
type ErrorType int

const (
	ParseError ErrorType = iota
	ValidationError
	NetworkError
	CacheError
	InternalError
	ConfigError
)

// String representations for error types
func (e ErrorType) String() string {
	switch e {
	case ParseError:
		return "ParseError"
	case ValidationError:
		return "ValidationError"
	case NetworkError:
		return "NetworkError"
	case CacheError:
		return "CacheError"
	case InternalError:
		return "InternalError"
	case ConfigError:
		return "ConfigError"
	default:
		return "UnknownError"
	}
}

// DNSError represents a DNS operation error with context
type DNSError struct {
	Type    ErrorType
	Op      string
	Message string
	Err     error
}

func (e *DNSError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s - %v", e.Op, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Op, e.Message)
}

// Error constructors
func NewParseError(op string, msg string, err error) error {
	return &DNSError{Type: ParseError, Op: op, Message: msg, Err: err}
}

func NewValidationError(op string, msg string, err error) error {
	return &DNSError{Type: ValidationError, Op: op, Message: msg, Err: err}
}

func NewNetworkError(op string, msg string, err error) error {
	return &DNSError{Type: NetworkError, Op: op, Message: msg, Err: err}
}

func NewCacheError(op string, msg string, err error) error {
	return &DNSError{Type: CacheError, Op: op, Message: msg, Err: err}
}

func NewInternalError(op string, msg string, err error) error {
	return &DNSError{Type: InternalError, Op: op, Message: msg, Err: err}
}

func NewConfigError(op string, msg string, err error) error {
	return &DNSError{Type: ConfigError, Op: op, Message: msg, Err: err}
}

// Is checks if target error is of the given type
func Is(err error, errType ErrorType) bool {
	if dnsErr, ok := err.(*DNSError); ok {
		return dnsErr.Type == errType
	}
	return false
}
