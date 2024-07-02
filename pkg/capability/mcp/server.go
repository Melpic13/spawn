package mcp

import "context"

// Handler handles inbound MCP calls.
type Handler func(context.Context, map[string]interface{}) (map[string]interface{}, error)

// Server stores MCP handlers.
type Server struct {
	handlers map[string]Handler
}

// NewServer returns a new MCP server.
func NewServer() *Server {
	return &Server{handlers: make(map[string]Handler)}
}

// Register registers a handler.
func (s *Server) Register(name string, h Handler) {
	s.handlers[name] = h
}

// Handle executes a registered handler.
func (s *Server) Handle(ctx context.Context, name string, input map[string]interface{}) (map[string]interface{}, error) {
	if h, ok := s.handlers[name]; ok {
		return h(ctx, input)
	}
	return map[string]interface{}{"error": "not found"}, nil
}
