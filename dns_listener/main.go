package dns_listener

import (
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"golang.org/x/time/rate"
)

// Add color constants
const (
	colorReset  = "\033[0m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
)

// Add near the top of the file after imports
var isTestMode = false

// DNSListener represents the structure for a DNS listener
type DNSListener struct {
	Port      string
	LogFile   *os.File
	bufPool   sync.Pool
	cache     *dnsCache
	limiter   *rate.Limiter
	workers   int
	requestCh chan dnsRequest
}

type dnsRequest struct {
	data       []byte
	remoteAddr net.Addr
	protocol   string
	respCh     chan []byte
}

type dnsCache struct {
	sync.RWMutex
	entries map[string]dnsCacheEntry
}

type dnsCacheEntry struct {
	response []byte
	expires  time.Time
}

// calculateOptimalWorkers determines the optimal number of workers based on system resources
func calculateOptimalWorkers() int {
	cpuCount := runtime.NumCPU()

	// Get memory info
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Convert memory to GB
	totalMemoryGB := float64(m.Sys) / (1024 * 1024 * 1024)

	// Base calculation:
	// - Minimum 4 workers
	// - Maximum of (CPU count * 4)
	// - 1 worker per 256MB of system memory
	memoryWorkers := int(totalMemoryGB * 4) // 4 workers per GB of RAM
	cpuWorkers := cpuCount * 3              // CPU count * I/O waiting factor

	// Choose the smaller of the two to avoid overloading
	workers := min(memoryWorkers, cpuWorkers)

	// Ensure minimum and maximum bounds
	workers = max(4, min(workers, cpuCount*4))

	fmt.Printf("System resources: CPU cores: %d, Memory: %.2f GB\n", cpuCount, totalMemoryGB)
	fmt.Printf("Calculated workers: %d (memory workers: %d, CPU workers: %d)\n",
		workers, memoryWorkers, cpuWorkers)

	return workers
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// NewDNSListener initializes a new DNSListener instance
func NewDNSListener(port, logFilePath string) (*DNSListener, error) {
	if p := parsePort(port); p == -1 {
		return nil, fmt.Errorf("invalid port number: %s", port)
	}

	// Use LOG_PATH environment variable if available
	logPath := os.Getenv("LOG_PATH")
	if logPath == "" {
		logPath = "./"
	}

	// Add date prefix and ensure directory exists
	currentDate := time.Now().Format("2006-01-02")
	dateLogFilePath := filepath.Join(logPath, currentDate+"_"+filepath.Base(logFilePath))

	// Ensure log directory exists
	if err := os.MkdirAll(filepath.Dir(dateLogFilePath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	logFile, err := os.OpenFile(dateLogFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	if !hasPermission(port) {
		return nil, fmt.Errorf("insufficient permissions to bind to port %s", port)
	}

	workers := calculateOptimalWorkers()

	d := &DNSListener{
		Port:      port,
		LogFile:   logFile,
		workers:   workers,
		requestCh: make(chan dnsRequest, workers*20), // Buffer size scaled with worker count
		cache: &dnsCache{
			entries: make(map[string]dnsCacheEntry),
		},
		limiter: rate.NewLimiter(rate.Limit(100000), 1000), // 100k requests/second burst 1k
		bufPool: sync.Pool{
			New: func() interface{} {
				return make([]byte, 512)
			},
		},
	}

	return d, nil
}

// printBanner prints the DNS server banner
func printBanner() {
	banner := `
╔═══════════════════════════════════╗
║         DNS Listener Active       ║
╚═══════════════════════════════════╝
`
	fmt.Print(colorGreen, banner, colorReset)
	os.Stdout.Sync()
}

// printStats prints the DNS listener configuration and stats
func (d *DNSListener) printStats() {
	stats := fmt.Sprintf(`
%s=== DNS Listener Configuration ===%s
► Port: %s
► Worker Pool Size: %d workers
► Request Channel Buffer: %d requests
► Rate Limit: %.0f requests/second (burst: %d)
► DNS Message Buffer Size: %d bytes
► Cache TTL: %s
► Cache Cleanup Interval: %s
%s===================================%s
`,
		colorCyan,
		colorReset,
		d.Port,
		d.workers,
		cap(d.requestCh),
		d.limiter.Limit(),
		d.limiter.Burst(),
		512,
		time.Second*600,
		time.Minute,
		colorCyan,
		colorReset,
	)
	fmt.Print(stats)
	os.Stdout.Sync()
	os.Stderr.Sync() // Add stderr flush
}

// Start starts the DNS listener with worker pool
func (d *DNSListener) Start() {
	printBanner()
	d.printStats()

	// Start worker pool
	for i := 0; i < d.workers; i++ {
		go d.worker()
	}

	go d.listenUDP()
	go d.listenTCP()
	go d.cleanCache()
	go d.monitorStats() // Add periodic stats monitoring

	fmt.Printf("%sPress Ctrl+C to stop the listener%s\n", colorYellow, colorReset)
	os.Stdout.Sync()
}

// monitorStats periodically prints runtime statistics
func (d *DNSListener) monitorStats() {
	ticker := time.NewTicker(30 * time.Second)
	for range ticker.C {
		d.cache.RLock()
		cacheSize := len(d.cache.entries)
		d.cache.RUnlock()

		stats := fmt.Sprintf(`
%s=== Runtime Statistics ===%s
► Cache Size: %d entries
► Request Channel Load: %d/%d
%s=========================%s
`,
			colorYellow,
			colorReset,
			cacheSize,
			len(d.requestCh),
			cap(d.requestCh),
			colorYellow,
			colorReset,
		)
		fmt.Print(stats)
		os.Stdout.Sync()
	}
}

// worker processes DNS requests from the queue
func (d *DNSListener) worker() {
	for req := range d.requestCh {
		if !d.limiter.Allow() {
			// Return rate limit exceeded response
			continue
		}

		// Log the request first
		d.logRequest(req.protocol, req.remoteAddr.String(), req.data)

		// Check cache first
		if resp := d.checkCache(req.data); resp != nil {
			req.respCh <- resp
			continue
		}

		response := createDNSResponse(req.data, req.remoteAddr.String())
		d.updateCache(req.data, response)
		req.respCh <- response
	}
}

func (d *DNSListener) checkCache(query []byte) []byte {
	key := hex.EncodeToString(query)
	d.cache.RLock()
	defer d.cache.RUnlock()

	if entry, exists := d.cache.entries[key]; exists && time.Now().Before(entry.expires) {
		return entry.response
	}
	return nil
}

func (d *DNSListener) updateCache(query, response []byte) {
	key := hex.EncodeToString(query)
	d.cache.Lock()
	defer d.cache.Unlock()

	d.cache.entries[key] = dnsCacheEntry{
		response: response,
		expires:  time.Now().Add(600 * time.Second), // Short TTL for testing
	}
}

func (d *DNSListener) cleanCache() {
	ticker := time.NewTicker(1 * time.Minute)
	for range ticker.C {
		d.cache.Lock()
		now := time.Now()
		for key, entry := range d.cache.entries {
			if now.After(entry.expires) {
				delete(d.cache.entries, key)
			}
		}
		d.cache.Unlock()
	}
}

// listenUDP optimized for high performance
func (d *DNSListener) listenUDP() {
	// Use syscall.Socket for UDP to preserve original client IP
	addr := net.UDPAddr{
		Port: parsePort(d.Port),
		IP:   net.IPv4zero,
	}

	conn, err := net.ListenUDP("udp", &addr)
	if err != nil {
		fmt.Println("Failed to start UDP listener:", err)
		return
	}
	defer conn.Close()

	// Set socket options to receive original destination if available
	if rawConn, err := conn.SyscallConn(); err == nil {
		rawConn.Control(func(fd uintptr) {
			// Enable IP_TRANSPARENT if running in container
			if os.Getenv("IN_CONTAINER") == "true" {
				syscall.SetsockoptInt(int(fd), syscall.SOL_IP, syscall.IP_TRANSPARENT, 1)
			}
		})
	}

	fmt.Printf("Listening on UDP port %s with client IP preservation\n", d.Port)

	for {
		buf := d.bufPool.Get().([]byte)
		n, remoteAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			d.bufPool.Put(buf)
			continue
		}

		// Get the real client IP
		realIP := remoteAddr.IP.String()

		respCh := make(chan []byte, 1)
		d.requestCh <- dnsRequest{
			data:       buf[:n],
			remoteAddr: remoteAddr,
			protocol:   "UDP",
			respCh:     respCh,
		}

		go func(buf []byte, remoteAddr *net.UDPAddr) {
			response := <-respCh
			conn.WriteToUDP(response, remoteAddr)
			d.bufPool.Put(buf)
		}(buf, remoteAddr)

		// Log with real client IP
		fmt.Printf("Received request from real client IP: %s\n", realIP)
	}
}

// listenTCP optimized for high performance
func (d *DNSListener) listenTCP() {
	ln, err := net.Listen("tcp", ":"+d.Port)
	if err != nil {
		fmt.Println("Failed to start TCP listener:", err)
		return
	}
	defer ln.Close()
	fmt.Println("Listening on TCP port", d.Port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting TCP connection:", err)
			continue
		}
		go d.handleTCPConnection(conn)
	}
}

// handleTCPConnection optimized
func (d *DNSListener) handleTCPConnection(conn net.Conn) {
	defer conn.Close()

	buf := d.bufPool.Get().([]byte)
	defer d.bufPool.Put(buf)

	for {
		conn.SetReadDeadline(time.Now().Add(2 * time.Second))

		// Read length prefix
		if _, err := conn.Read(buf[:2]); err != nil {
			return
		}
		messageLength := int(buf[0])<<8 | int(buf[1])

		if messageLength > 512 {
			return // Message too large
		}

		// Read message
		if _, err := conn.Read(buf[2 : messageLength+2]); err != nil {
			return
		}

		respCh := make(chan []byte, 1)
		d.requestCh <- dnsRequest{
			data:       buf[2 : messageLength+2],
			remoteAddr: conn.RemoteAddr(),
			protocol:   "TCP",
			respCh:     respCh,
		}

		response := <-respCh
		responseLength := len(response)
		lengthPrefix := []byte{byte(responseLength >> 8), byte(responseLength & 0xFF)}
		conn.Write(append(lengthPrefix, response...))
	}
}

// logRequest logs DNS requests to the file
func (d *DNSListener) logRequest(protocol, remoteAddr string, data []byte) {
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	humanReadable := parseDNSQuery(data)
	hexDump := hex.Dump(data)
	logEntry := fmt.Sprintf("[%s] [%s] Client: %s\n%s\nRaw Query (Hex):\n%s\n",
		timestamp, protocol, remoteAddr, humanReadable, hexDump)

	// Print to console with colors
	fmt.Printf("%s%s%s", colorCyan, logEntry, colorReset)
	os.Stdout.Sync()

	// Write to log file
	d.LogFile.WriteString(logEntry)
	d.LogFile.Sync() // Force flush to file
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

	fmt.Printf("Responding to client %s with DNS response\n", clientIP)
	return response
}

// parsePort parses the port from a string and ensures it is valid
func parsePort(port string) int {
	p, err := net.LookupPort("udp", port)
	if err != nil || p < 1 || p > 65535 {
		return -1 // Return invalid port instead of defaulting
	}
	return p
}

// hasPermission checks if the program has permission to bind to a privileged port
func hasPermission(port string) bool {
	// Skip permission check in test mode
	if isTestMode {
		return true
	}
	// Test if sudo exists
	if _, err := exec.LookPath("sudo"); err != nil {
		return true
	}

	cmd := exec.Command("sudo", "-n", "true")
	if err := cmd.Run(); err != nil {
		fmt.Printf("Permission denied for port %s. Please run with elevated privileges.\n", port)
		return false
	}
	return true
}

// parseDNSQuery converts raw DNS query bytes into a human-readable format
func parseDNSQuery(data []byte) string {
	if len(data) < 12 {
		return "Malformed DNS query"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Transaction ID: %x\n", data[:2]))
	sb.WriteString(fmt.Sprintf("Flags: %x\n", data[2:4]))
	qCount := int(data[4])<<8 | int(data[5])
	sb.WriteString(fmt.Sprintf("Questions: %d\n", qCount))

	offset := 12
	for i := 0; i < qCount; i++ {
		name, newOffset := parseDNSName(data, offset)
		if name == "" {
			return "Error parsing DNS name"
		}
		sb.WriteString(fmt.Sprintf("Question: %s\n", name))

		// Adjust offset to read Type and Class
		offset = newOffset + 1 // Move past the null terminator
		if offset+4 <= len(data) {
			queryType := int(data[offset])<<8 | int(data[offset+1])    // Big-endian
			queryClass := int(data[offset+2])<<8 | int(data[offset+3]) // Big-endian
			sb.WriteString(fmt.Sprintf("Type: %04x (%s)\n", queryType, typeToString(queryType)))
			sb.WriteString(fmt.Sprintf("Class: %04x (%s)\n", queryClass, classToString(queryClass)))
			offset += 4 // Move past Type and Class
		} else {
			sb.WriteString("Type and Class are missing or incomplete\n")
		}
	}

	return sb.String()
}

// Helper function to map query types to human-readable strings
func typeToString(queryType int) string {
	switch queryType {
	case 1:
		return "A"
	case 2:
		return "NS"
	case 3:
		return "MD"
	case 4:
		return "MF"
	case 5:
		return "CNAME"
	case 6:
		return "SOA"
	case 7:
		return "MB"
	case 8:
		return "MG"
	case 9:
		return "MR"
	case 10:
		return "NULL"
	case 11:
		return "WKS"
	case 12:
		return "PTR"
	case 13:
		return "HINFO"
	case 14:
		return "MINFO"
	case 15:
		return "MX"
	case 16:
		return "TXT"
	case 17:
		return "RP"
	case 18:
		return "AFSDB"
	case 19:
		return "X25"
	case 20:
		return "ISDN"
	case 21:
		return "RT"
	case 24:
		return "SIG"
	case 25:
		return "KEY"
	case 28:
		return "AAAA"
	case 33:
		return "SRV"
	case 41:
		return "OPT"
	case 43:
		return "DS"
	case 46:
		return "RRSIG"
	case 47:
		return "NSEC"
	case 48:
		return "DNSKEY"
	case 50:
		return "NSEC3"
	case 51:
		return "NSEC3PARAM"
	case 52:
		return "TLSA"
	case 59:
		return "CDS"
	case 60:
		return "CDNSKEY"
	case 61:
		return "OPENPGPKEY"
	case 62:
		return "CSYNC"
	case 99:
		return "SPF"
	case 249:
		return "TKEY"
	case 250:
		return "TSIG"
	case 251:
		return "IXFR"
	case 252:
		return "AXFR"
	case 253:
		return "MAILB"
	case 254:
		return "MAILA"
	case 255:
		return "ANY"
	case 256:
		return "URI"
	case 257:
		return "CAA"
	case 32768:
		return "TA"
	case 32769:
		return "DLV"
	default:
		return fmt.Sprintf("Unknown (%04x)", queryType)
	}
}

// Helper function to map query classes to human-readable strings
func classToString(queryClass int) string {
	if queryClass == 1 {
		return "IN"
	}
	return fmt.Sprintf("Unknown (%04x)", queryClass)
}

// parseDNSName extracts a DNS name from the query
func parseDNSName(data []byte, offset int) (string, int) {
	var labels []string
	for {
		if offset >= len(data) {
			return "", offset
		}
		length := int(data[offset])
		if length == 0 {
			break
		}
		offset++
		if offset+length > len(data) {
			return "", offset
		}
		labels = append(labels, string(data[offset:offset+length]))
		offset += length
	}
	return strings.Join(labels, "."), offset
}

// Close closes the log file
func (d *DNSListener) Close() {
	d.LogFile.Close()
}

func Run(port string) {
	if port == "" {
		port = "25353"
	}
	listener, err := NewDNSListener(port, "dns_listener.log")
	if err != nil {
		fmt.Println("Error creating DNS listener:", err)
		return
	}
	listener.Start()
	select {}
}
