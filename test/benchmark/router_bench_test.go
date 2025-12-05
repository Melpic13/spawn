package benchmark

import (
	"context"
	"testing"

	"spawn.dev/pkg/llm"
)

func BenchmarkRouterRoute(b *testing.B) {
	r := llm.NewRouter(llm.StrategyRoundRobin)
	_ = r.AddProvider(llm.NewAnthropicProvider("claude-sonnet-4-20250514"))
	_ = r.AddProvider(llm.NewOpenAIProvider("gpt-4o"))
	req := &llm.ChatRequest{Messages: []llm.Message{{Role: "user", Content: "hi"}}}
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = r.Route(ctx, req)
	}
}
