package mesh

import "sync/atomic"

// Coordinator provides simple leader election placeholder.
type Coordinator struct {
	leader atomic.Value
}

// Elect sets a leader id.
func (c *Coordinator) Elect(id string) {
	c.leader.Store(id)
}

// Leader returns current leader id.
func (c *Coordinator) Leader() string {
	if v := c.leader.Load(); v != nil {
		return v.(string)
	}
	return ""
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
