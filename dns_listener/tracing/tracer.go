package tracing

import (
	"context"
	"sync"
	"time"
)

type Event struct {
	Name      string
	Timestamp time.Time
	Error     error
}

type Trace struct {
	ID        string
	StartTime time.Time
	Events    []Event
	mu        sync.Mutex
}

type Tracer struct {
	traces sync.Map
}

func New() *Tracer {
	return &Tracer{}
}

func (t *Tracer) StartTrace(ctx context.Context) context.Context {
	traceID := generateTraceID()
	trace := &Trace{
		ID:        traceID,
		StartTime: time.Now(),
		Events:    make([]Event, 0),
	}
	t.traces.Store(traceID, trace)
	return context.WithValue(ctx, "trace_id", traceID)
}

func (t *Tracer) AddEvent(ctx context.Context, name string, err error) {
	traceID, ok := ctx.Value("trace_id").(string)
	if !ok {
		return
	}

	if trace, ok := t.traces.Load(traceID); ok {
		tr := trace.(*Trace)
		tr.mu.Lock()
		tr.Events = append(tr.Events, Event{
			Name:      name,
			Timestamp: time.Now(),
			Error:     err,
		})
		tr.mu.Unlock()
	}
}

func generateTraceID() string {
	return time.Now().Format("20060102150405.000000000")
}
