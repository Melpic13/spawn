package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// AnthropicProvider is a lightweight anthropic provider wrapper.
type AnthropicProvider struct {
	defaultModel string
}

// NewAnthropicProvider returns an Anthropic provider.
func NewAnthropicProvider(defaultModel string) *AnthropicProvider {
	return &AnthropicProvider{defaultModel: defaultModel}
}

func (p *AnthropicProvider) Name() string     { return "anthropic" }
func (p *AnthropicProvider) Models() []string { return []string{p.defaultModel} }

func (p *AnthropicProvider) Chat(_ context.Context, req *ChatRequest) (*ChatResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("anthropic chat: nil request")
	}
	prompt := joinMessages(req.Messages)
	if req.Model == "" {
		req.Model = p.defaultModel
	}
	return &ChatResponse{
		ID:         uuid.NewString(),
		Model:      req.Model,
		Content:    "[anthropic] " + prompt,
		StopReason: StopEndTurn,
		Usage:      &Usage{InputTokens: len(prompt) / 4, OutputTokens: len(prompt) / 8},
	}, nil
}

func (p *AnthropicProvider) ChatStream(ctx context.Context, req *ChatRequest) (<-chan *StreamChunk, error) {
	resp, err := p.Chat(ctx, req)
	if err != nil {
		return nil, err
	}
	ch := make(chan *StreamChunk, 2)
	go func() {
		defer close(ch)
		ch <- &StreamChunk{Delta: resp.Content}
		ch <- &StreamChunk{Done: true, StopReason: resp.StopReason}
	}()
	return ch, nil
}

func (p *AnthropicProvider) ChatWithTools(ctx context.Context, req *ChatRequest, tools []Tool) (*ChatResponse, error) {
	resp, err := p.Chat(ctx, req)
	if err != nil {
		return nil, err
	}
	if len(tools) > 0 {
		resp.ToolCalls = []ToolCall{{ID: uuid.NewString(), Name: tools[0].Name, Input: map[string]interface{}{}}}
		resp.StopReason = StopToolUse
	}
	return resp, nil
}

func (p *AnthropicProvider) Embed(_ context.Context, input []string) ([][]float32, error) {
	out := make([][]float32, 0, len(input))
	for _, item := range input {
		out = append(out, []float32{float32(len(item))})
	}
	return out, nil
}

func (p *AnthropicProvider) EstimateCost(req *ChatRequest) float64 {
	if req == nil {
		return 0
	}
	return float64(len(joinMessages(req.Messages))) * 0.000001
}

func (p *AnthropicProvider) HealthCheck(context.Context) error { return nil }

func joinMessages(messages []Message) string {
	parts := make([]string, 0, len(messages))
	for _, msg := range messages {
		parts = append(parts, msg.Content)
	}
	return strings.TrimSpace(strings.Join(parts, "\n"))
}
