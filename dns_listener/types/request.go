package types

import "net"

// Request represents a DNS request
type Request struct {
	Conn       net.Conn
	Protocol   string
	ClientAddr net.Addr
	Data       []byte
}
