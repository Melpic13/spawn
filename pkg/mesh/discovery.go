package mesh

// AgentInfo defines a discoverable agent.
type AgentInfo struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Labels    map[string]string `json:"labels,omitempty"`
	Healthy   bool              `json:"healthy"`
}

// DiscoveryQuery defines discovery filters.
type DiscoveryQuery struct {
	Namespace string
	Labels    map[string]string
	Healthy   *bool
}
