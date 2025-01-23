package dns_typo_checker

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// mockDNSFunc is used to replace the real DNS lookup in tests
var mockDNSFunc = func(domain string) bool {
	// Extended mock responses
	validDomains := map[string]bool{
		"example.com": true,
		"example.net": true,
		"example.org": true,
		"exaple.com":  false,
		"exampl.com":  false,
		"xample.com":  false,
		"sz.de":       true,
		"sz.com":      true,
		"s.de":        false,
		"z.de":        false,
		"google.com":  true,
	}
	// Fast response for unknown domains
	if _, exists := validDomains[domain]; !exists {
		return false
	}
	return validDomains[domain]
}

func TestMain(m *testing.M) {
	// Save original function
	originalCheckDNS := CheckDNS
	// Replace with mock for tests
	CheckDNS = mockDNSFunc
	// Run tests
	code := m.Run()
	// Restore original function
	CheckDNS = originalCheckDNS
	os.Exit(code)
}

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
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	tests := []struct {
		name   string
		domain string
		want   bool
	}{
		{
			name:   "Valid domain",
			domain: "example.com",
			want:   true,
		},
		{
			name:   "Invalid domain",
			domain: "thisisaninvaliddomain.com",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckDNS(tt.domain); got != tt.want {
				t.Errorf("CheckDNS() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRun(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	// Create temporary test directory
	tmpDir := t.TempDir()
	os.Setenv("LOG_PATH", tmpDir)

	// Test with limited domain set
	domains := []string{"example.com"}
	tlds := []string{"net", "org"}

	Run(domains, tlds)

	// Verify log files
	files, err := filepath.Glob(filepath.Join(tmpDir, "*dns_typo_checker*.log"))
	if err != nil || len(files) == 0 {
		t.Error("Expected log files were not created")
	}
}
