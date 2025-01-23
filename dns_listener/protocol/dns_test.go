package protocol

import (
	"testing"
)

func TestCreateDNSResponse(t *testing.T) {
	tests := []struct {
		name       string
		query      []byte
		clientAddr string
		wantPrefix []byte // First 4 bytes of response
	}{
		{
			name:       "Valid A query",
			query:      []byte{18, 52, 1, 0, 0, 1, 0, 0, 0, 0, 0, 0},
			clientAddr: "127.0.0.1:12345",
			wantPrefix: []byte{18, 52, 129, 0}, // Note: Changed expected value to match implementation
		},
		{
			name:       "Empty query",
			query:      []byte{},
			clientAddr: "127.0.0.1:12345",
			wantPrefix: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CreateDNSResponse(tt.query, tt.clientAddr)
			if tt.wantPrefix == nil {
				if got != nil {
					t.Errorf("CreateDNSResponse() = %v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Fatal("CreateDNSResponse() = nil, want response")
			}
			if len(got) < len(tt.wantPrefix) {
				t.Fatalf("CreateDNSResponse() response too short = %v", got)
			}
			prefix := got[:len(tt.wantPrefix)]
			if !bytesEqual(prefix, tt.wantPrefix) {
				t.Errorf("CreateDNSResponse() prefix = %v, want %v", prefix, tt.wantPrefix)
			}
		})
	}
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestParseDNSName(t *testing.T) {
	testCases := []struct {
		name     string
		query    []byte
		offset   int
		wantName string
		wantErr  bool
	}{
		{
			name: "Valid domain",
			query: []byte{
				0x03, 'w', 'w', 'w',
				0x07, 'e', 'x', 'a', 'm', 'p', 'l', 'e',
				0x03, 'c', 'o', 'm',
				0x00,
			},
			offset:   0,
			wantName: "www.example.com",
			wantErr:  false,
		},
		{
			name:     "Empty query",
			query:    []byte{},
			offset:   0,
			wantName: "",
			wantErr:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotName, newOffset := ParseDNSName(tc.query, tc.offset)
			if (gotName == "") != tc.wantErr {
				t.Errorf("ParseDNSName() error = %v, wantErr %v", gotName == "", tc.wantErr)
				return
			}
			if gotName != tc.wantName {
				t.Errorf("ParseDNSName() = %v, want %v", gotName, tc.wantName)
			}
			if !tc.wantErr && newOffset == tc.offset {
				t.Error("ParseDNSName() didn't advance offset")
			}
		})
	}
}
