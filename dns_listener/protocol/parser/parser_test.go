package parser

import (
	"testing"
)

func TestParseDNSHeader(t *testing.T) {
	tests := []struct {
		name    string
		query   []byte
		wantID  uint16
		wantErr bool
	}{
		{
			name: "valid header",
			query: []byte{
				0x12, 0x34, // ID
				0x01, 0x00, // Standard query
				0x00, 0x01, // Questions
				0x00, 0x00, // Answers
				0x00, 0x00, // Authority
				0x00, 0x00, // Additional
			},
			wantID:  0x1234,
			wantErr: false,
		},
		{
			name:    "too short",
			query:   []byte{0x12, 0x34},
			wantID:  0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header, err := ParseDNSHeader(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDNSHeader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && header.ID != tt.wantID {
				t.Errorf("ParseDNSHeader() ID = %v, want %v", header.ID, tt.wantID)
			}
		})
	}
}

func TestParseDNSQuestion(t *testing.T) {
	tests := []struct {
		name       string
		query      []byte
		wantDomain string
		wantErr    bool
	}{
		{
			name: "valid question",
			query: []byte{
				// Header
				0x00, 0x01, 0x01, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				// Question
				0x07, 'e', 'x', 'a', 'm', 'p', 'l', 'e',
				0x03, 'c', 'o', 'm',
				0x00,       // Root label
				0x00, 0x01, // Type A
				0x00, 0x01, // Class IN
			},
			wantDomain: "example.com",
			wantErr:    false,
		},
		{
			name:       "invalid query",
			query:      []byte{0x00},
			wantDomain: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			domain, err := ParseDNSQuestion(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDNSQuestion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && domain != tt.wantDomain {
				t.Errorf("ParseDNSQuestion() domain = %v, want %v", domain, tt.wantDomain)
			}
		})
	}
}

func TestParseDNSHeader_AdditionalCases(t *testing.T) {
	tests := []struct {
		name    string
		query   []byte
		want    *DNSHeader
		wantErr bool
	}{
		{
			name: "all fields populated",
			query: []byte{
				0x12, 0x34, // ID
				0x01, 0x00, // Flags
				0x00, 0x01, // Questions
				0x00, 0x02, // Answers
				0x00, 0x03, // Authority
				0x00, 0x04, // Additional
			},
			want: &DNSHeader{
				ID:      0x1234,
				Flags:   0x0100,
				QDCount: 1,
				ANCount: 2,
				NSCount: 3,
				ARCount: 4,
			},
			wantErr: false,
		},
		{
			name:    "nil input",
			query:   nil,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDNSHeader(tt.query)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDNSHeader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !headerEqual(got, tt.want) {
				t.Errorf("ParseDNSHeader() = %v, want %v", got, tt.want)
			}
		})
	}
}

func headerEqual(a, b *DNSHeader) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.ID == b.ID &&
		a.Flags == b.Flags &&
		a.QDCount == b.QDCount &&
		a.ANCount == b.ANCount &&
		a.NSCount == b.NSCount &&
		a.ARCount == b.ARCount
}
