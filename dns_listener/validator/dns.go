package validator

import "github.com/exiguus/ns-checker/dns_listener/protocol"

type DNSMessageValidator struct{}

// NewDNSMessageValidator creates a new DNS message validator
func NewDNSMessageValidator() *DNSMessageValidator {
	return &DNSMessageValidator{}
}

// ValidateQuery validates a DNS query
func (v *DNSMessageValidator) ValidateQuery(data []byte) error {
	return protocol.ValidateDNSMessage(data)
}
