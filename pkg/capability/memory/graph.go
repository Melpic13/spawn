package memory

import (
	"context"
	"sync"
)

// GraphStore is an embedded graph store.
type GraphStore struct {
	mu    sync.RWMutex
	nodes map[string]map[string]interface{}
}

// NewGraphStore creates a graph store.
func NewGraphStore() *GraphStore {
	return &GraphStore{nodes: make(map[string]map[string]interface{})}
}

// UpsertNode inserts or updates a graph node.
func (g *GraphStore) UpsertNode(_ context.Context, id string, payload map[string]interface{}) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.nodes[id] = payload
	return nil
}
