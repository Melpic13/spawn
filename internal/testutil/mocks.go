package testutil

import (
	"context"

	"spawn.dev/pkg/llm"
)

// MockProvider is a simple llm provider for tests.
type MockProvider struct{}

func (MockProvider) Name() string     { return "mock" }
func (MockProvider) Models() []string { return []string{"mock"} }
func (MockProvider) Chat(context.Context, *llm.ChatRequest) (*llm.ChatResponse, error) {
	return &llm.ChatResponse{Content: "ok", Usage: &llm.Usage{}}, nil
}
func (MockProvider) ChatStream(context.Context, *llm.ChatRequest) (<-chan *llm.StreamChunk, error) {
	ch := make(chan *llm.StreamChunk, 1)
	ch <- &llm.StreamChunk{Done: true}
	close(ch)
	return ch, nil
}
func (MockProvider) ChatWithTools(context.Context, *llm.ChatRequest, []llm.Tool) (*llm.ChatResponse, error) {
	return &llm.ChatResponse{Content: "ok", Usage: &llm.Usage{}}, nil
}
func (MockProvider) Embed(context.Context, []string) ([][]float32, error) {
	return [][]float32{{1}}, nil
}
func (MockProvider) EstimateCost(*llm.ChatRequest) float64 { return 0 }
func (MockProvider) HealthCheck(context.Context) error     { return nil }
