package dns_listener

import (
	"fmt"
)

// DNSMessage represents a DNS message structure
type DNSMessage struct {
	TransactionID uint16
	Flags         uint16
	Questions     uint16
	Answers       uint16
	Authority     uint16
	Additional    uint16
	Payload       []byte
}

func parseDNSMessage(data []byte) (*DNSMessage, error) {
	if len(data) < 12 {
		return nil, &DNSError{Op: "parse", Err: fmt.Errorf("message too short")}
	}

	msg := &DNSMessage{
		TransactionID: uint16(data[0])<<8 | uint16(data[1]),
		Flags:         uint16(data[2])<<8 | uint16(data[3]),
		Questions:     uint16(data[4])<<8 | uint16(data[5]),
		Answers:       uint16(data[6])<<8 | uint16(data[7]),
		Authority:     uint16(data[8])<<8 | uint16(data[9]),
		Additional:    uint16(data[10])<<8 | uint16(data[11]),
		Payload:       data[12:],
	}

	return msg, nil
}

// createDNSResponse creates a simple DNS response
func createDNSResponse(request []byte, clientIP string) []byte {
	if len(request) < 12 {
		return []byte{}
	}

	response := make([]byte, len(request))
	copy(response, request)
	response[2] = 0x81 // Set QR (response), Opcode (0), AA, TC, RD
	response[3] = 0x80 // RA

	response[6] = 0x00 // Answer RRs high byte
	response[7] = 0x01 // Answer RRs low byte

	response = append(response, 0xC0, 0x0C)             // Name pointer
	response = append(response, 0x00, 0x01)             // Type: A
	response = append(response, 0x00, 0x01)             // Class: IN
	response = append(response, 0x00, 0x00, 0x01, 0x2C) // TTL: 300
	response = append(response, 0x00, 0x04)             // Data length: 4 bytes
	response = append(response, 0x7F, 0x00, 0x00, 0x01) // Address: 127.0.0.1

	return response
}
