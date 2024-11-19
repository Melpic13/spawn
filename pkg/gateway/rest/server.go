package rest

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

// Server hosts REST endpoints.
type Server struct {
	addr string
	e    *echo.Echo
	http *http.Server
	mu   sync.Mutex
}

// New returns a REST server.
func New(addr string) *Server {
	e := echo.New()
	s := &Server{addr: addr, e: e}
	registerRoutes(e)
	return s
}

// Start starts HTTP listener.
func (s *Server) Start(_ context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.http != nil {
		return nil
	}
	if s.addr == "" {
		s.addr = ":8080"
	}
	s.http = &http.Server{
		Addr:              s.addr,
		Handler:           s.e,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		_ = s.http.ListenAndServe()
	}()
	return nil
}

// Stop gracefully stops server.
func (s *Server) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.http == nil {
		return nil
	}
	if err := s.http.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown rest server: %w", err)
	}
	s.http = nil
	return nil
}
