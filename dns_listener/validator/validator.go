package validator

import (
	"errors"
	"sync/atomic"
)

// Ensure DNSValidator implements MessageValidator
var _ MessageValidator = (*DNSValidator)(nil)

var (
	ErrMessageTooShort      = errors.New("DNS message too short")
	ErrInvalidHeaderSize    = errors.New("invalid DNS header size")
	ErrInvalidQuestionCount = errors.New("invalid question count")
	ErrMalformedQuestion    = errors.New("malformed question section")
	ErrUnsupportedOpcode    = errors.New("unsupported opcode")
)

// DNSValidator implements MessageValidator interface
type DNSValidator struct {
	stats ValidationStats // Use ValidationStats from interface.go
}

func New() *DNSValidator {
	return &DNSValidator{}
}

func (v *DNSValidator) ValidateQuery(data []byte) error {
	atomic.AddUint64(&v.stats.TotalValidated, 1)

	if err := v.validateBasics(data); err != nil {
		atomic.AddUint64(&v.stats.InvalidQueries, 1)
		return err
	}

	// Validate opcode
	opcode := (data[2] >> 3) & 0x0F
	if opcode != 0 {
		atomic.AddUint64(&v.stats.InvalidQueries, 1)
		return ErrUnsupportedOpcode
	}

	// Validate question section
	if err := v.validateQuestions(data); err != nil {
		atomic.AddUint64(&v.stats.InvalidQueries, 1)
		return err
	}

	return nil
}

func (v *DNSValidator) ValidateResponse(data []byte) error {
	if err := v.validateBasics(data); err != nil {
		atomic.AddUint64(&v.stats.InvalidResponses, 1)
		return err
	}

	// Check QR bit is set
	if (data[2] & 0x80) == 0 {
		atomic.AddUint64(&v.stats.InvalidResponses, 1)
		return errors.New("response bit not set")
	}

	return nil
}

func (v *DNSValidator) validateBasics(data []byte) error {
	if len(data) < 12 {
		return ErrMessageTooShort
	}

	questionCount := int(data[4])<<8 | int(data[5])
	if questionCount == 0 {
		return ErrInvalidQuestionCount
	}

	return nil
}

func (v *DNSValidator) validateQuestions(data []byte) error {
	offset := 12
	questionCount := int(data[4])<<8 | int(data[5])

	for i := 0; i < questionCount; i++ {
		// Parse name
		for offset < len(data) {
			length := int(data[offset])
			if length == 0 {
				offset++
				break
			}
			offset += length + 1
			if offset >= len(data) {
				return ErrMalformedQuestion
			}
		}

		// Check type and class fields
		if offset+4 > len(data) {
			return ErrMalformedQuestion
		}
		offset += 4
	}

	return nil
}

func (v *DNSValidator) GetStats() ValidationStats {
	return ValidationStats{
		TotalValidated:   atomic.LoadUint64(&v.stats.TotalValidated),
		InvalidQueries:   atomic.LoadUint64(&v.stats.InvalidQueries),
		InvalidResponses: atomic.LoadUint64(&v.stats.InvalidResponses),
	}
}
