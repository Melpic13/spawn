package gateway

import (
	"context"
	"fmt"
	"sync"

	grpcgw "spawn.dev/pkg/gateway/grpc"
	restgw "spawn.dev/pkg/gateway/rest"
	wsgw "spawn.dev/pkg/gateway/websocket"
)

// Config configures the gateway listeners.
type Config struct {
	GRPCAddr string
	RESTAddr string
	WSAddr   string
}

// Gateway aggregates gRPC, REST and WebSocket servers.
type Gateway struct {
	cfg   Config
	grpc  *grpcgw.Server
	rest  *restgw.Server
	ws    *wsgw.Server
	mu    sync.Mutex
	alive bool
}

// New creates a gateway.
func New(cfg Config) *Gateway {
	return &Gateway{
		cfg:  cfg,
		grpc: grpcgw.New(cfg.GRPCAddr),
		rest: restgw.New(cfg.RESTAddr),
		ws:   wsgw.New(cfg.WSAddr),
	}
}

// Start starts all gateway servers.
func (g *Gateway) Start(ctx context.Context) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.alive {
		return nil
	}
	if err := g.grpc.Start(ctx); err != nil {
		return fmt.Errorf("start grpc: %w", err)
	}
	if err := g.rest.Start(ctx); err != nil {
		return fmt.Errorf("start rest: %w", err)
	}
	if err := g.ws.Start(ctx); err != nil {
		return fmt.Errorf("start websocket: %w", err)
	}
	g.alive = true
	return nil
}

// Stop stops all gateway servers.
func (g *Gateway) Stop(ctx context.Context) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if !g.alive {
		return nil
	}
	if err := g.grpc.Stop(ctx); err != nil {
		return err
	}
	if err := g.rest.Stop(ctx); err != nil {
		return err
	}
	if err := g.ws.Stop(ctx); err != nil {
		return err
	}
	g.alive = false
	return nil
}
