package dns_listener

import (
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type FileLogger struct {
	file       *os.File
	mu         sync.Mutex
	debugMode  bool
	debugLevel string
	logPath    string
	flushRate  time.Duration
	lastFlush  time.Time
}

func NewFileLogger(logPath string) (Logger, error) {
	logsDir := filepath.Dir(logPath)
	absLogsDir, err := filepath.Abs(logsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve logs directory path: %w", err)
	}

	if err := os.MkdirAll(absLogsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create logs directory %s: %w", absLogsDir, err)
	}

	// Generate dated log filename
	now := time.Now()
	dateStr := now.Format("2006-01-02")
	baseName := filepath.Base(logPath)
	ext := filepath.Ext(baseName)
	nameWithoutExt := strings.TrimSuffix(baseName, ext)
	fullPath := filepath.Join(absLogsDir, fmt.Sprintf("%s_%s%s", dateStr, nameWithoutExt, ext))

	file, err := os.OpenFile(fullPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file %s: %w", fullPath, err)
	}

	startEntry := fmt.Sprintf("[%s] DNS Listener started\n", now.Format("2006-01-02 15:04:05"))
	if _, err := file.WriteString(startEntry); err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to write initial log entry: %w", err)
	}
	file.Sync()

	logger := &FileLogger{
		file:       file,
		debugMode:  os.Getenv("DEBUG") == "true",
		debugLevel: os.Getenv("DNS_LISTENER_DEBUG_LEVEL"),
		logPath:    fullPath,
		flushRate:  time.Second * 1, // Flush every second
	}

	// Start background flush routine
	go logger.periodicFlush()

	return logger, nil
}

func (l *FileLogger) periodicFlush() {
	ticker := time.NewTicker(l.flushRate)
	for range ticker.C {
		l.mu.Lock()
		if l.file != nil {
			l.file.Sync()
		}
		l.mu.Unlock()
	}
}

func (l *FileLogger) LogRequest(protocol, remoteAddr string, data []byte, err error) {
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	humanReadable := parseDNSQuery(data)

	// Extract IP address without port
	clientIP := remoteAddr
	if idx := strings.LastIndex(remoteAddr, ":"); idx != -1 {
		clientIP = remoteAddr[:idx]
	}

	var sb strings.Builder
	// Basic info with all fields
	sb.WriteString(fmt.Sprintf("[%s] [%s] Client: %s\n", timestamp, protocol, remoteAddr))
	sb.WriteString(fmt.Sprintf("Protocol: %s\n", protocol))
	sb.WriteString(fmt.Sprintf("Client IP: %s\n", clientIP))

	// DNS query details
	sb.WriteString(humanReadable)

	// Raw hex dump in canonical format
	sb.WriteString("Raw Query (Hex):\n")
	sb.WriteString(hex.Dump(data))
	sb.WriteString("\n")

	if err != nil {
		sb.WriteString(fmt.Sprintf("Error: %v\n", err))
	}

	// Write directly to file and console
	l.mu.Lock()
	defer l.mu.Unlock()

	l.file.WriteString(sb.String())
	l.file.Sync()

	// Print to console only if in debug mode or debug level is info/debug
	if l.debugMode || l.debugLevel == "info" || l.debugLevel == "debug" {
		fmt.Printf("%s%s%s", colorCyan, sb.String(), colorReset)
	}
	os.Stdout.Sync()
}

func (l *FileLogger) Write(entry string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Ensure entry ends with newline
	if !strings.HasSuffix(entry, "\n") {
		entry += "\n"
	}

	// Write to file
	if _, err := l.file.WriteString(entry); err != nil {
		fmt.Printf("Error writing to log file: %v\n", err)
		// Try to reopen the file
		if err := l.reopenLogFile(); err != nil {
			fmt.Printf("Failed to reopen log file: %v\n", err)
		}
	}

	// Only print to console if it's not an INFO log in non-debug mode
	if l.debugMode || l.debugLevel == "info" || l.debugLevel == "debug" {
		fmt.Printf("%s%s%s", colorCyan, entry, colorReset)
	}
	os.Stdout.Sync()
}

func (l *FileLogger) reopenLogFile() error {
	if l.file != nil {
		l.file.Close()
	}

	file, err := os.OpenFile(l.logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	l.file = file
	return nil
}

func (l *FileLogger) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file != nil {
		l.file.Close()
	}
}

func (l *FileLogger) Error(msg string, err error) {
	timestamp := time.Now().Format("[2006-01-02 15:04:05.000]")
	l.Write(fmt.Sprintf("%s ERROR: %s: %v\n", timestamp, msg, err))
}
