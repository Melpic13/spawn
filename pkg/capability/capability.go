package capability

import (
	"context"
	"time"
)

// Capability represents a capability that can be granted to an agent.
type Capability interface {
	Name() string
	Version() string
	Description() string
	Initialize(ctx context.Context, config map[string]interface{}) error
	Shutdown(ctx context.Context) error
	HealthCheck(ctx context.Context) error
	Schema() *Schema
	Execute(ctx context.Context, request *Request) (*Response, error)
}

// EventType is a named emitted event.
type EventType string

// Schema defines the capability interface.
type Schema struct {
	Actions []Action         `json:"actions"`
	Events  []EventType      `json:"events"`
	Config  map[string]Field `json:"config"`
}

// Action defines an executable operation.
type Action struct {
	Name        string           `json:"name"`
	Description string           `json:"description"`
	Input       map[string]Field `json:"input"`
	Output      map[string]Field `json:"output"`
}

// Field defines a typed field.
type Field struct {
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Required    bool        `json:"required"`
	Default     interface{} `json:"default,omitempty"`
}

// ExecutionContext is lightweight execution context passed into capability invocations.
type ExecutionContext struct {
	AgentID  string            `json:"agentId"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Request is a capability execution request.
type Request struct {
	Action  string                 `json:"action"`
	Params  map[string]interface{} `json:"params"`
	Context *ExecutionContext      `json:"context"`
	Timeout time.Duration          `json:"timeout"`
}

// Response is a capability execution response.
type Response struct {
	Success bool              `json:"success"`
	Data    interface{}       `json:"data,omitempty"`
	Error   *Error            `json:"error,omitempty"`
	Metrics *ExecutionMetrics `json:"metrics,omitempty"`
}

// Error represents a structured capability error.
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ExecutionMetrics captures invocation timings.
type ExecutionMetrics struct {
	Duration time.Duration `json:"duration"`
}

// Registry manages capability registration and discovery.
type Registry interface {
	Register(cap Capability) error
	Unregister(name string) error
	Get(name string) (Capability, error)
	List() []Capability
	Discover(ctx context.Context) ([]Capability, error)
}

// VectorStore represents vector memory storage.
type VectorStore interface {
	Put(ctx context.Context, key string, vec []float32) error
	Search(ctx context.Context, vec []float32, limit int) ([]string, error)
}

// GraphStore represents graph memory storage.
type GraphStore interface {
	UpsertNode(ctx context.Context, id string, payload map[string]interface{}) error
}

// KVStore represents key/value memory storage.
type KVStore interface {
	Set(ctx context.Context, key string, value []byte) error
	Get(ctx context.Context, key string) ([]byte, error)
}
