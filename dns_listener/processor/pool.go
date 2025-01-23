package processor

import (
	"sync"

	"github.com/exiguus/ns-checker/dns_listener/types"
)

// requestPool manages a pool of request buffers
type requestPool struct {
	pool *sync.Pool
}

func newRequestPool() *requestPool {
	return &requestPool{
		pool: &sync.Pool{
			New: func() interface{} {
				return &types.Request{
					Data: make([]byte, 512), // Standard DNS message size
				}
			},
		},
	}
}

func (p *requestPool) get() *types.Request {
	return p.pool.Get().(*types.Request)
}

func (p *requestPool) put(req *types.Request) {
	// Clear sensitive data
	req.Conn = nil
	req.ClientAddr = nil
	req.Protocol = ""
	for i := range req.Data {
		req.Data[i] = 0
	}
	p.pool.Put(req)
}
