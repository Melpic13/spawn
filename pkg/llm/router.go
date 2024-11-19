package llm

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
)

// ProviderRouter routes across multiple providers.
type ProviderRouter struct {
	mu        sync.RWMutex
	providers map[string]Provider
	strategy  RoutingStrategy
	index     int
}

// NewRouter returns a provider router.
func NewRouter(strategy RoutingStrategy) *ProviderRouter {
	if strategy == "" {
		strategy = StrategyRoundRobin
	}
	return &ProviderRouter{providers: map[string]Provider{}, strategy: strategy}
}

// AddProvider adds a provider.
func (r *ProviderRouter) AddProvider(provider Provider) error {
	if provider == nil {
		return fmt.Errorf("add provider: nil provider")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[strings.ToLower(provider.Name())] = provider
	return nil
}

// RemoveProvider removes by name.
func (r *ProviderRouter) RemoveProvider(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.providers, strings.ToLower(name))
	return nil
}

// SetStrategy updates routing strategy.
func (r *ProviderRouter) SetStrategy(strategy RoutingStrategy) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.strategy = strategy
}

// Route selects a provider for the request.
func (r *ProviderRouter) Route(_ context.Context, req *ChatRequest) (Provider, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.providers) == 0 {
		return nil, fmt.Errorf("route provider: no providers configured")
	}

	if req != nil && req.Model != "" {
		for _, p := range r.providers {
			for _, m := range p.Models() {
				if m == req.Model {
					return p, nil
				}
			}
		}
	}

	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	sort.Strings(names)

	switch r.strategy {
	case StrategyCostOptimize:
		var best Provider
		bestCost := 1e18
		for _, name := range names {
			p := r.providers[name]
			cost := p.EstimateCost(req)
			if cost < bestCost {
				bestCost = cost
				best = p
			}
		}
		return best, nil
	case StrategyRoundRobin, StrategyFallback, StrategyComplexity, StrategyLatencyOptimize:
		idx := r.index % len(names)
		r.index++
		return r.providers[names[idx]], nil
	default:
		idx := r.index % len(names)
		r.index++
		return r.providers[names[idx]], nil
	}
}
