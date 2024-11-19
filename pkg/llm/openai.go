package llm

import "context"

// OpenAIProvider is an OpenAI provider wrapper.
type OpenAIProvider struct {
	defaultModel string
}

// NewOpenAIProvider returns an OpenAI provider.
func NewOpenAIProvider(defaultModel string) *OpenAIProvider {
	return &OpenAIProvider{defaultModel: defaultModel}
}

func (p *OpenAIProvider) Name() string     { return "openai" }
func (p *OpenAIProvider) Models() []string { return []string{p.defaultModel} }
func (p *OpenAIProvider) Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
	if req == nil || req.Model == "" {
		if req == nil {
			req = &ChatRequest{}
		}
		req.Model = p.defaultModel
	}
	resp, err := NewAnthropicProvider(p.defaultModel).Chat(ctx, req)
	if err != nil {
		return nil, err
	}
	resp.Content = "[openai] " + resp.Content
	return resp, nil
}
func (p *OpenAIProvider) ChatStream(ctx context.Context, req *ChatRequest) (<-chan *StreamChunk, error) {
	return NewAnthropicProvider(p.defaultModel).ChatStream(ctx, req)
}
func (p *OpenAIProvider) ChatWithTools(ctx context.Context, req *ChatRequest, tools []Tool) (*ChatResponse, error) {
	return NewAnthropicProvider(p.defaultModel).ChatWithTools(ctx, req, tools)
}
func (p *OpenAIProvider) Embed(ctx context.Context, input []string) ([][]float32, error) {
	return NewAnthropicProvider(p.defaultModel).Embed(ctx, input)
}
func (p *OpenAIProvider) EstimateCost(req *ChatRequest) float64 {
	return NewAnthropicProvider(p.defaultModel).EstimateCost(req)
}
func (p *OpenAIProvider) HealthCheck(context.Context) error { return nil }
