package network

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
)

type Server struct {
	udpConn     *net.UDPConn
	tcpListener net.Listener
	handler     RequestHandler
	wg          sync.WaitGroup
	stopChan    chan struct{}
	port        string
	ctx         context.Context
	cancel      context.CancelFunc
}

func NewServer(port string, handler RequestHandler) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		handler:  handler,
		stopChan: make(chan struct{}),
		port:     port,
		ctx:      ctx,
		cancel:   cancel,
	}
}

func (s *Server) Start(ctx context.Context) error {
	errChan := make(chan error, 2)

	s.wg.Add(2) // Add for UDP and TCP servers

	// Start UDP listener
	go func() {
		defer s.wg.Done()
		if err := s.startUDP(); err != nil {
			errChan <- fmt.Errorf("UDP listener failed: %w", err)
		}
	}()

	// Start TCP listener
	go func() {
		defer s.wg.Done()
		if err := s.startTCP(); err != nil {
			errChan <- fmt.Errorf("TCP listener failed: %w", err)
		}
	}()

	// Wait for context cancellation or error
	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return nil
	}
}

func (s *Server) Stop() {
	s.cancel() // Signal all goroutines to stop

	if s.udpConn != nil {
		s.udpConn.Close()
	}
	if s.tcpListener != nil {
		s.tcpListener.Close()
	}

	s.wg.Wait() // Wait for main server goroutines to finish
}

func (s *Server) startUDP() error {
	addr := &net.UDPAddr{
		Port: s.getPort(),
		IP:   net.ParseIP("0.0.0.0"),
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return fmt.Errorf("failed to start UDP listener: %w", err)
	}
	s.udpConn = conn
	fmt.Printf("UDP server listening on %s:%d\n", addr.IP, addr.Port)

	buffer := make([]byte, 4096)
	for {
		select {
		case <-s.ctx.Done():
			return nil
		default:
			n, remoteAddr, err := conn.ReadFromUDP(buffer)
			if err != nil {
				if !strings.Contains(err.Error(), "use of closed network connection") {
					fmt.Printf("UDP read error: %v\n", err)
				}
				return nil
			}

			go s.handleUDPRequest(buffer[:n], remoteAddr)
		}
	}
}

func (s *Server) startTCP() error {
	addr := &net.TCPAddr{
		Port: s.getPort(),
		IP:   net.ParseIP("0.0.0.0"),
	}
	conn, err := net.Listen("tcp", addr.String())
	if err != nil {
		return fmt.Errorf("failed to start TCP listener: %w", err)
	}
	s.tcpListener = conn
	fmt.Printf("TCP server listening on %s:%d\n", addr.IP, addr.Port)

	for {
		select {
		case <-s.ctx.Done():
			return nil
		default:
			conn, err := s.tcpListener.Accept()
			if err != nil {
				if !strings.Contains(err.Error(), "use of closed network connection") {
					fmt.Printf("TCP accept error: %v\n", err)
				}
				return nil
			}

			go func() {
				defer conn.Close()
				s.handleTCPConnection(conn)
			}()
		}
	}
}

func (s *Server) handleUDPRequest(data []byte, addr *net.UDPAddr) {
	// Handle request
	response, err := s.handler.HandleRequest(data, addr, "UDP")
	if err != nil {
		fmt.Printf("Handler error for %s: %v\n", addr.String(), err)
		return
	}

	if response == nil {
		return
	}

	// Send response
	_, err = s.udpConn.WriteToUDP(response, addr)
	if err != nil {
		fmt.Printf("UDP write error to %s: %v\n", addr.String(), err)
	}
}

func (s *Server) handleTCPConnection(conn net.Conn) {
	buffer := make([]byte, 512)
	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			// Read message length
			if _, err := conn.Read(buffer[:2]); err != nil {
				return
			}
			length := int(buffer[0])<<8 | int(buffer[1])

			// Read message
			if length > len(buffer)-2 {
				return
			}
			if _, err := conn.Read(buffer[2 : length+2]); err != nil {
				return
			}

			response, err := s.handler.HandleRequest(buffer[2:length+2], conn.RemoteAddr(), "TCP")
			if err != nil {
				continue
			}

			// Write response length
			respLen := len(response)
			conn.Write([]byte{byte(respLen >> 8), byte(respLen)})
			conn.Write(response)
		}
	}
}

func (s *Server) getPort() int {
	port, err := strconv.Atoi(s.port)
	if err != nil || port < 1 || port > 65535 {
		return 25353 // Default port
	}
	return port
}
