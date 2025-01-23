package validator

// MessageValidator defines the interface for DNS message validation
type MessageValidator interface {
	ValidateQuery(data []byte) error
	ValidateResponse(data []byte) error
	GetStats() ValidationStats
}

// ValidationStats represents validation statistics
type ValidationStats struct {
	TotalValidated   uint64
	InvalidQueries   uint64
	InvalidResponses uint64
}
