package websocket

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// Server hosts websocket endpoints.
type Server struct {
	addr string
	mux  *http.ServeMux
	http *http.Server
	mu   sync.Mutex
}

// New creates websocket server.
func New(addr string) *Server {
	mux := http.NewServeMux()
	s := &Server{addr: addr, mux: mux}
	mux.HandleFunc("/ws", s.handle)
	return s
}

// Start starts websocket server.
func (s *Server) Start(_ context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.http != nil {
		return nil
	}
	if s.addr == "" {
		s.addr = ":8081"
	}
	s.http = &http.Server{Addr: s.addr, Handler: s.mux}
	go func() {
		_ = s.http.ListenAndServe()
	}()
	return nil
}

// Stop stops websocket server.
func (s *Server) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.http == nil {
		return nil
	}
	if err := s.http.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown websocket server: %w", err)
	}
	s.http = nil
	return nil
}

func (s *Server) handle(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	_ = conn.WriteJSON(map[string]string{"status": "connected"})
}
