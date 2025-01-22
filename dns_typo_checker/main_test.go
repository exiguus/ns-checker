package dns_typo_checker

import (
	"strings"
	"testing"
)

func TestGenerateTypoDomains(t *testing.T) {
	tests := []struct {
		name       string
		domain     string
		commonTLDs []string
		expected   int // number of expected typos
	}{
		{
			name:       "Simple domain",
			domain:     "example.com",
			commonTLDs: []string{"com", "net", "org"},
			expected:   15,
		},
		{
			name:       "Empty domain",
			domain:     "",
			commonTLDs: []string{"com", "net", "org"},
			expected:   0,
		},
		{
			name:       "Domain without TLD",
			domain:     "example",
			commonTLDs: []string{"com", "net", "org"},
			expected:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenerateTypoDomains(tt.domain, tt.commonTLDs)
			if len(result) != tt.expected {
				t.Errorf("GenerateTypoDomains(%s) got %d typos, want %d", tt.domain, len(result), tt.expected)
			}

			if tt.domain != "" && strings.Contains(tt.domain, ".") {
				// Check that none of the generated typos are identical to the original domain
				for _, typo := range result {
					if typo == tt.domain {
						t.Errorf("GenerateTypoDomains(%s) generated the original domain as a typo", tt.domain)
					}
				}
			}
		})
	}
}

func TestCheckDNS(t *testing.T) {
	tests := []struct {
		name     string
		domain   string
		expected bool
	}{
		{
			name:     "Valid domain",
			domain:   "google.com",
			expected: true,
		},
		{
			name:     "Invalid domain",
			domain:   "thisisaninvalid12domain789.com",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckDNS(tt.domain)
			if result != tt.expected {
				t.Errorf("CheckDNS(%s) = %v, want %v", tt.domain, result, tt.expected)
			}
		})
	}
}

func TestRun(t *testing.T) {
	// Test with empty domain list
	Run([]string{}, []string{})

	// Test with valid domains
	domains := []string{"sz.de"}
	commonTLDs := []string{"com", "net", "org"}

	Run(domains, commonTLDs)
	// Note: This is more of an integration test
	// In a real-world scenario, you might want to mock the file operations
	// and DNS checks for more isolated unit testing
}
