package llm

import (
	"context"
	"testing"
)

func TestRouterRoundRobin(t *testing.T) {
	t.Parallel()
	r := NewRouter(StrategyRoundRobin)
	if err := r.AddProvider(NewAnthropicProvider("claude")); err != nil {
		t.Fatalf("add anthropic: %v", err)
	}
	if err := r.AddProvider(NewOpenAIProvider("gpt")); err != nil {
		t.Fatalf("add openai: %v", err)
	}

	p1, err := r.Route(context.Background(), &ChatRequest{})
	if err != nil {
		t.Fatalf("route1: %v", err)
	}
	p2, err := r.Route(context.Background(), &ChatRequest{})
	if err != nil {
		t.Fatalf("route2: %v", err)
	}
	if p1.Name() == p2.Name() {
		t.Fatalf("expected round robin providers to differ")
	}
}

func TestRouterModelPreference(t *testing.T) {
	t.Parallel()
	r := NewRouter(StrategyRoundRobin)
	_ = r.AddProvider(NewAnthropicProvider("claude-model"))
	_ = r.AddProvider(NewOpenAIProvider("gpt-model"))

	p, err := r.Route(context.Background(), &ChatRequest{Model: "gpt-model"})
	if err != nil {
		t.Fatalf("route model: %v", err)
	}
	if p.Name() != "openai" {
		t.Fatalf("expected openai provider, got %s", p.Name())
	}
}
