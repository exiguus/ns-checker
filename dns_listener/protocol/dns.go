package protocol

import (
	"fmt"
	"strings"
)

// ValidationError represents DNS validation errors
type ValidationError struct {
	Field  string
	Reason string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("invalid DNS message: %s - %s", e.Field, e.Reason)
}

// ValidateDNSMessage validates a DNS message
func ValidateDNSMessage(data []byte) error {
	if len(data) < 12 {
		return &ValidationError{Field: "length", Reason: "message too short"}
	}

	questionCount := int(data[4])<<8 | int(data[5])
	if questionCount == 0 {
		return &ValidationError{Field: "questions", Reason: "no questions in query"}
	}
	return nil
}

// CreateDNSResponse creates a DNS response from a query
func CreateDNSResponse(query []byte, clientAddr string) []byte {
	if len(query) < 12 {
		return nil
	}

	response := make([]byte, len(query))
	copy(response, query)

	// Set QR bit to indicate response
	response[2] |= 0x80

	return response
}

// ParseDNSName parses a DNS name from the query bytes starting at the given offset
func ParseDNSName(data []byte, offset int) (string, int) {
	var labels []string
	startOffset := offset

	for {
		if offset >= len(data) {
			return "", startOffset
		}
		length := int(data[offset])
		if length == 0 {
			break
		}
		offset++
		if offset+length > len(data) {
			return "", startOffset
		}
		labels = append(labels, string(data[offset:offset+length]))
		offset += length
	}
	if len(labels) == 0 {
		return "", startOffset
	}
	return strings.Join(labels, "."), offset
}
