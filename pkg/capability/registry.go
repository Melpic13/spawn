package capability

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
)

// InMemoryRegistry is a concurrency-safe registry implementation.
type InMemoryRegistry struct {
	mu    sync.RWMutex
	items map[string]Capability
}

// NewRegistry returns a new in-memory capability registry.
func NewRegistry() *InMemoryRegistry {
	return &InMemoryRegistry{items: make(map[string]Capability)}
}

// Register registers a capability by normalized name.
func (r *InMemoryRegistry) Register(cap Capability) error {
	if cap == nil {
		return fmt.Errorf("register capability: nil capability")
	}
	name := strings.ToLower(cap.Name())
	if name == "" {
		return fmt.Errorf("register capability: empty name")
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.items[name]; exists {
		return fmt.Errorf("register capability: %s already exists", name)
	}
	r.items[name] = cap
	return nil
}

// Unregister removes a capability by name.
func (r *InMemoryRegistry) Unregister(name string) error {
	name = strings.ToLower(name)
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.items[name]; !exists {
		return fmt.Errorf("unregister capability: %s not found", name)
	}
	delete(r.items, name)
	return nil
}

// Get returns a capability by name.
func (r *InMemoryRegistry) Get(name string) (Capability, error) {
	name = strings.ToLower(name)
	r.mu.RLock()
	defer r.mu.RUnlock()
	cap, ok := r.items[name]
	if !ok {
		return nil, fmt.Errorf("get capability: %s not found", name)
	}
	return cap, nil
}

// List returns all registered capabilities sorted by name.
func (r *InMemoryRegistry) List() []Capability {
	r.mu.RLock()
	defer r.mu.RUnlock()
	keys := make([]string, 0, len(r.items))
	for k := range r.items {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out := make([]Capability, 0, len(keys))
	for _, k := range keys {
		out = append(out, r.items[k])
	}
	return out
}

// Discover returns all locally registered capabilities.
func (r *InMemoryRegistry) Discover(_ context.Context) ([]Capability, error) {
	return r.List(), nil
}
