package network

import (
	"context"
	"net"
	"testing"
	"time"
)

type mockHandler struct{}

func (m *mockHandler) HandleRequest(data []byte, addr net.Addr, proto string) ([]byte, error) {
	return data, nil
}

func TestServer(t *testing.T) {
	handler := &mockHandler{}
	server := NewServer("0", handler)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Start(ctx)
	}()

	// Wait for server to start
	time.Sleep(100 * time.Millisecond)

	server.Stop()

	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("server error: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Error("server didn't stop in time")
	}
}
