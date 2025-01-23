package processor

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"

	dnserr "github.com/exiguus/ns-checker/dns_listener/errors"
	"github.com/exiguus/ns-checker/dns_listener/metrics"
	"github.com/exiguus/ns-checker/dns_listener/tracing"
	"github.com/exiguus/ns-checker/dns_listener/types"
)

const (
	DefaultTimeout = 5 * time.Second
	maxRetries     = 3
)

type Processor struct {
	workers    int
	timeout    time.Duration
	handler    RequestHandler
	metrics    *metrics.Collector
	requestCh  chan types.Request
	pool       *sync.Pool
	ctx        context.Context
	cancelFunc context.CancelFunc
	reqPool    *requestPool
	tracer     *tracing.Tracer
}

type RequestHandler interface {
	HandleRequest(data []byte, addr net.Addr, protocol string) ([]byte, error)
}

type ProcessorConfig struct {
	Workers    int
	Timeout    time.Duration
	BufferSize int
}

func New(cfg ProcessorConfig, handler RequestHandler, metrics *metrics.Collector) *Processor {
	ctx, cancel := context.WithCancel(context.Background())

	return &Processor{
		workers:    cfg.Workers,
		timeout:    cfg.Timeout,
		handler:    handler,
		metrics:    metrics,
		requestCh:  make(chan types.Request, cfg.BufferSize),
		pool:       &sync.Pool{New: func() interface{} { return make([]byte, 512) }},
		ctx:        ctx,
		cancelFunc: cancel,
		reqPool:    newRequestPool(),
		tracer:     tracing.New(),
	}
}

func (p *Processor) Start() {
	for i := 0; i < p.workers; i++ {
		go p.worker()
	}
}

func (p *Processor) Stop() {
	p.cancelFunc()
}

func (p *Processor) Process(req types.Request) {
	select {
	case p.requestCh <- req:
		// Request accepted
	case <-p.ctx.Done():
		// Processor is shutting down
		p.metrics.RecordError()
	default:
		// Channel full, handle overflow
		p.metrics.RecordError()
	}
}

func (p *Processor) worker() {
	for {
		select {
		case <-p.ctx.Done():
			return
		case req := <-p.requestCh:
			p.handleRequest(req)
		}
	}
}

func (p *Processor) handleRequest(req types.Request) {
	// Create timeout context for request
	ctx, cancel := context.WithTimeout(p.ctx, p.timeout)
	defer cancel()

	// Start trace
	ctx = p.tracer.StartTrace(ctx)
	p.tracer.AddEvent(ctx, "request_received", nil)

	// Get request from pool
	pooledReq := p.reqPool.get()
	defer p.reqPool.put(pooledReq)

	// Copy request data
	pooledReq.Conn = req.Conn
	pooledReq.ClientAddr = req.ClientAddr
	pooledReq.Protocol = req.Protocol
	copy(pooledReq.Data, req.Data)

	var response []byte
	var err error

	// Handle request with retries
	for attempt := 1; attempt <= maxRetries; attempt++ {
		select {
		case <-ctx.Done():
			p.tracer.AddEvent(ctx, "request_timeout", ctx.Err())
			p.metrics.RecordError()
			return
		default:
			p.tracer.AddEvent(ctx, fmt.Sprintf("attempt_%d_start", attempt), nil)
			response, err = p.handler.HandleRequest(pooledReq.Data, pooledReq.ClientAddr, pooledReq.Protocol)
			if err == nil {
				p.tracer.AddEvent(ctx, fmt.Sprintf("attempt_%d_success", attempt), nil)
				break
			}
			p.tracer.AddEvent(ctx, fmt.Sprintf("attempt_%d_failed", attempt), err)

			if attempt == maxRetries {
				p.metrics.RecordError()
				return
			}

			// Simple exponential backoff
			time.Sleep(time.Duration(attempt*100) * time.Millisecond)
		}
	}

	if err != nil {
		p.metrics.RecordError()
		return
	}

	// Send response with context
	select {
	case <-ctx.Done():
		p.metrics.RecordError()
		return
	default:
		if err := p.sendResponse(req.Conn, response); err != nil {
			p.metrics.RecordError()
		}
	}
}

func (p *Processor) sendResponse(conn net.Conn, response []byte) error {
	_, err := conn.Write(response)
	if err != nil {
		return dnserr.NewNetworkError("sendResponse", "failed to write response", err)
	}
	return nil
}
