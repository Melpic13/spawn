package llm

import "context"

// Provider represents an LLM provider.
type Provider interface {
	Name() string
	Models() []string
	Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
	ChatStream(ctx context.Context, req *ChatRequest) (<-chan *StreamChunk, error)
	ChatWithTools(ctx context.Context, req *ChatRequest, tools []Tool) (*ChatResponse, error)
	Embed(ctx context.Context, input []string) ([][]float32, error)
	EstimateCost(req *ChatRequest) float64
	HealthCheck(ctx context.Context) error
}

// Message is one chat message.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest represents a chat completion request.
type ChatRequest struct {
	Model         string                 `json:"model"`
	Messages      []Message              `json:"messages"`
	System        string                 `json:"system,omitempty"`
	Temperature   float64                `json:"temperature,omitempty"`
	MaxTokens     int                    `json:"max_tokens,omitempty"`
	StopSequences []string               `json:"stop_sequences,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// ChatResponse represents a chat completion response.
type ChatResponse struct {
	ID         string     `json:"id"`
	Model      string     `json:"model"`
	Content    string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	StopReason StopReason `json:"stop_reason"`
	Usage      *Usage     `json:"usage"`
}

// StreamChunk is one streaming chunk.
type StreamChunk struct {
	Delta      string
	Done       bool
	StopReason StopReason
}

// Tool is a tool available to the LLM.
type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

// ToolCall is an LLM tool invocation request.
type ToolCall struct {
	ID    string                 `json:"id"`
	Name  string                 `json:"name"`
	Input map[string]interface{} `json:"input"`
}

// Usage contains token usage.
type Usage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// StopReason indicates why generation stopped.
type StopReason string

const (
	// StopEndTurn indicates normal response completion.
	StopEndTurn StopReason = "end_turn"
	// StopToolUse indicates a tool call was emitted.
	StopToolUse StopReason = "tool_use"
)

// Router routes requests to providers.
type Router interface {
	Route(ctx context.Context, req *ChatRequest) (Provider, error)
	AddProvider(provider Provider) error
	RemoveProvider(name string) error
	SetStrategy(strategy RoutingStrategy)
}

// RoutingStrategy controls provider selection strategy.
type RoutingStrategy string

const (
	StrategyRoundRobin      RoutingStrategy = "round-robin"
	StrategyCostOptimize    RoutingStrategy = "cost-optimize"
	StrategyLatencyOptimize RoutingStrategy = "latency-optimize"
	StrategyComplexity      RoutingStrategy = "complexity"
	StrategyFallback        RoutingStrategy = "fallback"
)
