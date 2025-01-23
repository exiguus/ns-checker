package errors

import (
	"errors"
	"testing"
)

func TestDNSErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantType ErrorType
	}{
		{
			name:     "validation error",
			err:      NewValidationError("test", "invalid request", nil),
			wantType: ValidationError,
		},
		{
			name:     "network error",
			err:      NewNetworkError("test", "connection failed", errors.New("network down")),
			wantType: NetworkError,
		},
		{
			name:     "internal error",
			err:      NewInternalError("test", "processing failed", nil),
			wantType: InternalError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var dnsErr *DNSError
			if !errors.As(tt.err, &dnsErr) {
				t.Errorf("error %v is not a DNSError", tt.err)
			}
			if dnsErr.Type != tt.wantType {
				t.Errorf("error type = %v, want %v", dnsErr.Type, tt.wantType)
			}
		})
	}
}
