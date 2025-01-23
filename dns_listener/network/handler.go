package network

import (
	"net"
)

// RequestHandler defines the interface for handling network requests
type RequestHandler interface {
	HandleRequest(data []byte, addr net.Addr, protocol string) ([]byte, error)
}
