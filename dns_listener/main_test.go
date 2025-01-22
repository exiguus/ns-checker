package dns_listener

import (
	"net"
	"os"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	// Set test mode before running tests
	isTestMode = true

	// Run tests
	code := m.Run()

	// Exit
	os.Exit(code)
}

func TestNewDNSListener(t *testing.T) {
	// Create temporary log file
	tmpLog := "test_dns.log"
	defer os.Remove(tmpLog)

	tests := []struct {
		name    string
		port    string
		wantErr bool
	}{
		{
			name:    "Valid port",
			port:    "45353", // Using higher port number
			wantErr: false,
		},
		{
			name:    "Invalid port",
			port:    "999999",
			wantErr: true,
		},
		{
			name:    "Invalid port string",
			port:    "invalid",
			wantErr: true,
		},
		{
			name:    "Empty port",
			port:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			listener, err := NewDNSListener(tt.port, tmpLog)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDNSListener() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && listener == nil {
				t.Error("NewDNSListener() returned nil but wanted valid listener")
			}
			if listener != nil {
				listener.Close()
			}
		})
	}
}

func TestDNSCache(t *testing.T) {
	cache := &dnsCache{
		entries: make(map[string]dnsCacheEntry),
	}

	testResponse := []byte("test.response")

	// Test cache update
	cache.Lock()
	cache.entries["test"] = dnsCacheEntry{
		response: testResponse,
		expires:  time.Now().Add(time.Second),
	}
	cache.Unlock()

	// Test cache hit
	cache.RLock()
	entry, exists := cache.entries["test"]
	cache.RUnlock()

	if !exists {
		t.Error("Cache entry not found")
	}
	if string(entry.response) != string(testResponse) {
		t.Error("Cache entry mismatch")
	}
}

func TestCreateDNSResponse(t *testing.T) {
	testQuery := []byte{
		0x00, 0x01, // Transaction ID
		0x01, 0x00, // Flags
		0x00, 0x01, // Questions
		0x00, 0x00, // Answer RRs
		0x00, 0x00, // Authority RRs
		0x00, 0x00, // Additional RRs
	}

	response := createDNSResponse(testQuery, "127.0.0.1")

	if len(response) == 0 {
		t.Error("Expected non-empty DNS response")
	}

	if response[2] != 0x81 || response[3] != 0x80 {
		t.Error("Invalid response flags")
	}
}

func TestParseDNSQuery(t *testing.T) {
	testQuery := []byte{
		0x00, 0x01, // Transaction ID
		0x01, 0x00, // Flags
		0x00, 0x01, // Questions
		0x00, 0x00, // Answer RRs
		0x00, 0x00, // Authority RRs
		0x00, 0x00, // Additional RRs
		0x03, 'w', 'w', 'w', // Label
		0x07, 'e', 'x', 'a', 'm', 'p', 'l', 'e',
		0x03, 'c', 'o', 'm',
		0x00,       // Root label
		0x00, 0x01, // Type A
		0x00, 0x01, // Class IN
	}

	result := parseDNSQuery(testQuery)

	if result == "" {
		t.Error("Expected non-empty parsed query")
	}
	if result == "Malformed DNS query" {
		t.Error("Query incorrectly marked as malformed")
	}
}

func TestUDPListener(t *testing.T) {
	tmpLog := "test_dns.log"
	defer os.Remove(tmpLog)

	testPort := "45354" // Using different high port
	listener, err := NewDNSListener(testPort, tmpLog)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	go listener.Start()
	time.Sleep(time.Second) // Wait for server to start

	// Create UDP client
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:"+testPort)
	if err != nil {
		t.Fatal(err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// Send test query
	testQuery := []byte{
		0x00, 0x01, // Transaction ID
		0x01, 0x00, // Flags
		0x00, 0x01, // Questions
		0x00, 0x00, // Answer RRs
		0x00, 0x00, // Authority RRs
		0x00, 0x00, // Additional RRs
	}

	_, err = conn.Write(testQuery)
	if err != nil {
		t.Fatal(err)
	}

	// Read response
	buffer := make([]byte, 512)
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, err = conn.Read(buffer)
	if err != nil {
		t.Fatal(err)
	}
}

func TestTCPListener(t *testing.T) {
	tmpLog := "test_dns.log"
	defer os.Remove(tmpLog)

	testPort := "45355" // Different port from UDP test
	listener, err := NewDNSListener(testPort, tmpLog)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	go listener.Start()
	time.Sleep(time.Second) // Wait for server to start

	// Create TCP client
	conn, err := net.Dial("tcp", "127.0.0.1:"+testPort)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	// Send test query with length prefix
	testQuery := []byte{
		0x00, 0x0c, // Length prefix (12 bytes)
		0x00, 0x01, // Transaction ID
		0x01, 0x00, // Flags
		0x00, 0x01, // Questions
		0x00, 0x00, // Answer RRs
		0x00, 0x00, // Authority RRs
		0x00, 0x00, // Additional RRs
	}

	_, err = conn.Write(testQuery)
	if err != nil {
		t.Fatal(err)
	}

	// Read response
	buffer := make([]byte, 512)
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, err = conn.Read(buffer)
	if err != nil {
		t.Fatal(err)
	}
}

func TestParsePort(t *testing.T) {
	tests := []struct {
		name     string
		port     string
		expected int
	}{
		{"Valid port", "8053", 8053},
		{"Invalid port high", "999999", -1},
		{"Invalid port string", "invalid", -1},
		{"Empty port", "", -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parsePort(tt.port)
			if result != tt.expected {
				t.Errorf("parsePort(%s) = %d, want %d", tt.port, result, tt.expected)
			}
		})
	}
}
