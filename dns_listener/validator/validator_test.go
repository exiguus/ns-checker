package validator

import "testing"

func TestValidator(t *testing.T) {
	v := New()

	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name: "valid query",
			data: []byte{
				0x00, 0x01, // ID
				0x01, 0x00, // Flags
				0x00, 0x01, // QDCOUNT
				0x00, 0x00, // ANCOUNT
				0x00, 0x00, // NSCOUNT
				0x00, 0x00, // ARCOUNT
				0x03, 'w', 'w', 'w', // QNAME
				0x07, 'e', 'x', 'a', 'm', 'p', 'l', 'e',
				0x03, 'c', 'o', 'm',
				0x00,       // QNAME terminator
				0x00, 0x01, // QTYPE
				0x00, 0x01, // QCLASS
			},
			wantErr: false,
		},
		{
			name:    "empty query",
			data:    []byte{},
			wantErr: true,
		},
		{
			name:    "invalid length",
			data:    []byte{0, 1},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateQuery(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateQuery() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
