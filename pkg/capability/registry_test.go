package capability

import (
	"context"
	"testing"
)

type mockCap struct{ name string }

func (m mockCap) Name() string                                             { return m.name }
func (m mockCap) Version() string                                          { return "v1" }
func (m mockCap) Description() string                                      { return "mock" }
func (m mockCap) Initialize(context.Context, map[string]interface{}) error { return nil }
func (m mockCap) Shutdown(context.Context) error                           { return nil }
func (m mockCap) HealthCheck(context.Context) error                        { return nil }
func (m mockCap) Schema() *Schema                                          { return &Schema{} }
func (m mockCap) Execute(context.Context, *Request) (*Response, error) {
	return &Response{Success: true}, nil
}

func TestRegistryRegisterGetList(t *testing.T) {
	t.Parallel()
	r := NewRegistry()
	if err := r.Register(mockCap{name: "exec"}); err != nil {
		t.Fatalf("register: %v", err)
	}
	if _, err := r.Get("exec"); err != nil {
		t.Fatalf("get: %v", err)
	}
	if got := len(r.List()); got != 1 {
		t.Fatalf("expected 1 capability, got %d", got)
	}
}
