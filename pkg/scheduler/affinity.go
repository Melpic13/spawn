package scheduler

// AffinityRule influences task placement decisions.
type AffinityRule struct {
	AgentID string
	Tag     string
	Weight  int
}
