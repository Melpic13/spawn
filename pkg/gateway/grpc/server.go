package grpc

import (
	"context"
	"fmt"
	"net"
	"sync"

	"google.golang.org/grpc"
)

// Server hosts gRPC services.
type Server struct {
	addr string
	srv  *grpc.Server
	ln   net.Listener
	mu   sync.Mutex
}

// New creates a grpc server wrapper.
func New(addr string) *Server {
	return &Server{addr: addr}
}

// Start starts gRPC server.
func (s *Server) Start(_ context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.srv != nil {
		return nil
	}
	if s.addr == "" {
		s.addr = ":9090"
	}
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("listen grpc: %w", err)
	}
	s.ln = ln
	s.srv = grpc.NewServer()
	go func() {
		_ = s.srv.Serve(ln)
	}()
	return nil
}

// Stop stops gRPC server.
func (s *Server) Stop(_ context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.srv != nil {
		s.srv.GracefulStop()
		s.srv = nil
	}
	if s.ln != nil {
		_ = s.ln.Close()
		s.ln = nil
	}
	return nil
}
