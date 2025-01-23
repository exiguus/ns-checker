package protocol

import "testing"

func TestDNSType_String(t *testing.T) {
	tests := []struct {
		name string
		t    DNSType
		want string
	}{
		{"A Record", TypeA, "A"},
		{"AAAA Record", TypeAAAA, "AAAA"},
		{"CNAME Record", TypeCNAME, "CNAME"},
		{"MX Record", TypeMX, "MX"},
		{"NS Record", TypeNS, "NS"},
		{"PTR Record", TypePTR, "PTR"},
		{"SOA Record", TypeSOA, "SOA"},
		{"TXT Record", TypeTXT, "TXT"},
		{"Unknown Type", DNSType(999), "TYPE-999"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.t.String(); got != tt.want {
				t.Errorf("DNSType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDNSClass_String(t *testing.T) {
	tests := []struct {
		name string
		c    DNSClass
		want string
	}{
		{"IN Class", ClassIN, "IN"},
		{"CS Class", ClassCS, "CS"},
		{"CH Class", ClassCH, "CH"},
		{"HS Class", ClassHS, "HS"},
		{"Unknown Class", DNSClass(999), "CLASS-999"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.String(); got != tt.want {
				t.Errorf("DNSClass.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDNSFlags_String(t *testing.T) {
	tests := []struct {
		name  string
		flags DNSFlags
		want  string
	}{
		{"Query Flag", FlagQR, "QR"},
		{"Authoritative Flag", FlagAA, "AA"},
		{"Truncation Flag", FlagTC, "TC"},
		{"Recursion Desired", FlagRD, "RD"},
		{"Recursion Available", FlagRA, "RA"},
		{"Multiple Flags", FlagQR | FlagAA, "QR|AA"},
		{"No Flags", DNSFlags(0), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.flags.String(); got != tt.want {
				t.Errorf("DNSFlags.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
