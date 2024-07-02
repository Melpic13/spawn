package agent

import (
	"context"
	"time"

	"spawn.dev/pkg/capability"
	"spawn.dev/pkg/llm"
)

// Message is an inter-agent message.
type Message struct {
	ID        string
	Topic     string
	Payload   interface{}
	Timestamp time.Time
}

// Task defines one unit of work.
type Task struct {
	ID      string
	Prompt  string
	Timeout time.Duration
}

// TaskResult is the result of task execution.
type TaskResult struct {
	TaskID   string
	Output   string
	Error    string
	Duration time.Duration
}

// LogEntry is a streamable structured log line.
type LogEntry struct {
	Time    time.Time
	Level   string
	Message string
}

// AgentMetrics exposes high-level agent counters.
type AgentMetrics struct {
	TokensUsed int64
	CostUSD    float64
	TasksRun   int64
}

// ListOptions filters list queries.
type ListOptions struct {
	Namespace string
}

// LogOptions defines log stream options.
type LogOptions struct {
	Follow bool
}

// WatchOptions defines event watch options.
type WatchOptions struct {
	Namespace string
}

// Event is an emitted lifecycle event.
type Event struct {
	Type      string
	AgentID   string
	Timestamp time.Time
}

// Agent represents a running AI agent instance.
type Agent struct {
	ID           string
	Name         string
	Namespace    string
	Config       *AgentConfig
	State        AgentState
	StartedAt    time.Time
	LLM          llm.Provider
	Capabilities map[string]capability.Capability
	Context      *ExecutionContext
	Inbox        chan Message
	Outbox       chan Message
	TokensUsed   int64
	CostUSD      float64
	TasksRun     int64
}

// Manager handles agent lifecycle.
type Manager interface {
	Create(ctx context.Context, config *AgentConfig) (*Agent, error)
	Start(ctx context.Context, id string) error
	Stop(ctx context.Context, id string) error
	Restart(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*Agent, error)
	List(ctx context.Context, opts ListOptions) ([]*Agent, error)
	SendMessage(ctx context.Context, id string, msg Message) error
	Execute(ctx context.Context, id string, task Task) (*TaskResult, error)
	Logs(ctx context.Context, id string, opts LogOptions) (<-chan LogEntry, error)
	Metrics(ctx context.Context, id string) (*AgentMetrics, error)
	Watch(ctx context.Context, opts WatchOptions) (<-chan Event, error)
}
