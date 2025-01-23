package protocol

import (
	"fmt"
	"strings"
)

// DNSType represents the type of DNS record
type DNSType uint16

// DNS Record Types
const (
	TypeA     DNSType = 1
	TypeNS    DNSType = 2
	TypeCNAME DNSType = 5
	TypeSOA   DNSType = 6
	TypePTR   DNSType = 12
	TypeMX    DNSType = 15
	TypeTXT   DNSType = 16
	TypeAAAA  DNSType = 28
)

// String returns the string representation of DNSType
func (t DNSType) String() string {
	switch t {
	case TypeA:
		return "A"
	case TypeNS:
		return "NS"
	case TypeCNAME:
		return "CNAME"
	case TypeSOA:
		return "SOA"
	case TypePTR:
		return "PTR"
	case TypeMX:
		return "MX"
	case TypeTXT:
		return "TXT"
	case TypeAAAA:
		return "AAAA"
	default:
		return fmt.Sprintf("TYPE-%d", t)
	}
}

// DNSClass represents the class of DNS record
type DNSClass uint16

// DNS Classes
const (
	ClassIN DNSClass = 1
	ClassCS DNSClass = 2
	ClassCH DNSClass = 3
	ClassHS DNSClass = 4
)

// String returns the string representation of DNSClass
func (c DNSClass) String() string {
	switch c {
	case ClassIN:
		return "IN"
	case ClassCS:
		return "CS"
	case ClassCH:
		return "CH"
	case ClassHS:
		return "HS"
	default:
		return fmt.Sprintf("CLASS-%d", c)
	}
}

// DNSFlags represents DNS header flags
type DNSFlags uint16

// DNS Header Flags
const (
	FlagQR DNSFlags = 1 << 15 // Query/Response
	FlagAA DNSFlags = 1 << 10 // Authoritative Answer
	FlagTC DNSFlags = 1 << 9  // Truncated
	FlagRD DNSFlags = 1 << 8  // Recursion Desired
	FlagRA DNSFlags = 1 << 7  // Recursion Available
)

// String returns the string representation of DNSFlags
func (f DNSFlags) String() string {
	var flags []string
	if f&FlagQR != 0 {
		flags = append(flags, "QR")
	}
	if f&FlagAA != 0 {
		flags = append(flags, "AA")
	}
	if f&FlagTC != 0 {
		flags = append(flags, "TC")
	}
	if f&FlagRD != 0 {
		flags = append(flags, "RD")
	}
	if f&FlagRA != 0 {
		flags = append(flags, "RA")
	}
	if len(flags) == 0 {
		return ""
	}
	return strings.Join(flags, "|")
}
