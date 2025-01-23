package parser

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strings"

	"github.com/exiguus/ns-checker/dns_listener/protocol"
)

// DNSHeader represents the header section of a DNS message
type DNSHeader struct {
	ID      uint16
	Flags   uint16
	QDCount uint16
	ANCount uint16
	NSCount uint16
	ARCount uint16
}

// ParseDNSHeader parses the header section of a DNS message
func ParseDNSHeader(query []byte) (*DNSHeader, error) {
	if len(query) < 12 {
		return nil, errors.New("DNS message too short")
	}

	header := &DNSHeader{
		ID:      binary.BigEndian.Uint16(query[0:2]),
		Flags:   binary.BigEndian.Uint16(query[2:4]),
		QDCount: binary.BigEndian.Uint16(query[4:6]),
		ANCount: binary.BigEndian.Uint16(query[6:8]),
		NSCount: binary.BigEndian.Uint16(query[8:10]),
		ARCount: binary.BigEndian.Uint16(query[10:12]),
	}

	return header, nil
}

// ParseDNSQuestion parses the question section of a DNS message and returns the domain name
func ParseDNSQuestion(query []byte) (string, error) {
	if len(query) < 12 {
		return "", errors.New("DNS message too short")
	}

	// Skip header
	pos := 12
	var labels []string

	// Parse domain name labels
	for pos < len(query) {
		labelLen := int(query[pos])
		if labelLen == 0 {
			break
		}
		if pos+1+labelLen > len(query) {
			return "", errors.New("invalid domain name length")
		}
		labels = append(labels, string(query[pos+1:pos+1+labelLen]))
		pos += 1 + labelLen
	}

	if len(labels) == 0 {
		return "", errors.New("no domain name found")
	}

	return strings.Join(labels, "."), nil
}

// Parser handles DNS message parsing
type Parser struct {
	data []byte
}

func New(data []byte) *Parser {
	return &Parser{data: data}
}

// ParseQuery parses a DNS query and returns a human-readable string
func (p *Parser) ParseQuery() (string, error) {
	if len(p.data) < 12 {
		return "", fmt.Errorf("malformed DNS query: message too short")
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Transaction ID: %x\n", p.data[:2]))
	sb.WriteString(fmt.Sprintf("Flags: %x\n", p.data[2:4]))

	qCount := int(p.data[4])<<8 | int(p.data[5])
	sb.WriteString(fmt.Sprintf("Questions: %d\n", qCount))

	offset := 12
	for i := 0; i < qCount; i++ {
		name, newOffset := p.parseName(offset)
		if name == "" {
			return "", fmt.Errorf("error parsing DNS name")
		}
		sb.WriteString(fmt.Sprintf("Question: %s\n", name))

		// Parse Type and Class
		offset = newOffset + 1
		if offset+4 <= len(p.data) {
			queryType := protocol.DNSType(int(p.data[offset])<<8 | int(p.data[offset+1]))
			queryClass := protocol.DNSClass(int(p.data[offset+2])<<8 | int(p.data[offset+3]))
			sb.WriteString(fmt.Sprintf("Type: %s\n", queryType))
			sb.WriteString(fmt.Sprintf("Class: %s\n", queryClass))
			offset += 4
		} else {
			return "", fmt.Errorf("malformed DNS query: incomplete type and class")
		}
	}

	return sb.String(), nil
}

// parseName extracts a DNS name from the query bytes
func (p *Parser) parseName(offset int) (string, int) {
	var labels []string

	for {
		if offset >= len(p.data) {
			return "", offset
		}
		length := int(p.data[offset])
		if length == 0 {
			break
		}
		offset++
		if offset+length > len(p.data) {
			return "", offset
		}
		labels = append(labels, string(p.data[offset:offset+length]))
		offset += length
	}

	if len(labels) == 0 {
		return "", offset
	}
	return strings.Join(labels, "."), offset
}
