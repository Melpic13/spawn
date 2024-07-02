package agent

// AgentState represents the current state of an agent.
type AgentState string

const (
	StateInitializing AgentState = "initializing"
	StateRunning      AgentState = "running"
	StatePaused       AgentState = "paused"
	StateCompleted    AgentState = "completed"
	StateFailed       AgentState = "failed"
	StateTerminated   AgentState = "terminated"
)
