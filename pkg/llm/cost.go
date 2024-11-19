package llm

import "sync"

// CostTracker tracks aggregate LLM costs.
type CostTracker struct {
	mu    sync.Mutex
	spent float64
}

// Add adds a cost amount.
func (c *CostTracker) Add(amount float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.spent += amount
}

// Spent returns aggregate cost.
func (c *CostTracker) Spent() float64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.spent
}
