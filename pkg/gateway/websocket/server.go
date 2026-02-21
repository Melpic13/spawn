package websocket

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Server hosts websocket endpoints.
type Server struct {
	addr     string
	mux      *http.ServeMux
	http     *http.Server
	mu       sync.Mutex
	upgrader websocket.Upgrader
}

// New creates websocket server.
func New(addr string) *Server {
	mux := http.NewServeMux()
	s := &Server{
		addr: addr,
		mux:  mux,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				origin := strings.TrimSpace(r.Header.Get("Origin"))
				if origin == "" {
					return true
				}
				return strings.Contains(origin, r.Host)
			},
		},
	}
	mux.HandleFunc("/ws", s.handle)
	RegisterHandlers(s)
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
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()
	conn.SetReadLimit(1 << 20)
	_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		_ = conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	_ = conn.WriteJSON(map[string]string{"status": "connected"})
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			var payload map[string]interface{}
			if err := conn.ReadJSON(&payload); err != nil {
				return
			}
			payload["receivedAt"] = time.Now().UTC().Format(time.RFC3339)
			if err := conn.WriteJSON(payload); err != nil {
				return
			}
		}
	}()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteMessage(websocket.PingMessage, []byte("ping")); err != nil {
				return
			}
		}
	}
}
