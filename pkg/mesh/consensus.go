package mesh

import (
	"sync"
	"time"
)

// Coordinator provides simple lease-based leader election semantics.
type Coordinator struct {
	mu      sync.RWMutex
	leader  string
	expires time.Time
}

// Elect sets a leader with a lease duration.
func (c *Coordinator) Elect(id string, lease time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.leader = id
	if lease <= 0 {
		lease = 30 * time.Second
	}
	c.expires = time.Now().UTC().Add(lease)
}

// Leader returns current leader id.
func (c *Coordinator) Leader() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.leader == "" {
		return ""
	}
	if time.Now().UTC().After(c.expires) {
		return ""
	}
	return c.leader
}

// Renew extends current leader lease if caller is the current leader.
func (c *Coordinator) Renew(id string, lease time.Duration) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.leader != id || c.leader == "" {
		return false
	}
	if lease <= 0 {
		lease = 30 * time.Second
	}
	c.expires = time.Now().UTC().Add(lease)
	return true
}

// Clear clears current leadership.
func (c *Coordinator) Clear(id string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if id != "" && c.leader != id {
		return false
	}
	c.leader = ""
	c.expires = time.Time{}
	return true
}

// AgentNode is a graph node.
type AgentNode struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// ChannelEdge is a graph channel edge.
type ChannelEdge struct {
	Name string      `json:"name"`
	Type ChannelType `json:"type"`
}

// Connection links two agents.
type Connection struct {
	From string `json:"from"`
	To   string `json:"to"`
}
