package agent

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{
			name: "valid config",
			body: `apiVersion: spawn.dev/v1
kind: Agent
metadata:
  name: tester
spec:
  model:
    provider: anthropic
    name: claude-sonnet-4-20250514
  sandbox:
    runtime: gvisor
`,
			wantErr: false,
		},
		{
			name: "missing metadata name",
			body: `apiVersion: spawn.dev/v1
kind: Agent
metadata: {}
spec:
  model:
    provider: anthropic
    name: claude-sonnet-4-20250514
  sandbox:
    runtime: gvisor
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			path := t.TempDir() + "/agent.yaml"
			if err := os.WriteFile(path, []byte(tt.body), 0o644); err != nil {
				t.Fatalf("write temp config: %v", err)
			}
			_, err := LoadConfig(path)
			if tt.wantErr && err == nil {
				t.Fatalf("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestMergeConfig(t *testing.T) {
	t.Parallel()
	base := &AgentConfig{
		APIVersion: "spawn.dev/v1",
		Kind:       "Agent",
		Metadata:   Metadata{Name: "base", Labels: map[string]string{"team": "a"}},
		Spec: AgentSpec{
			Model:   ModelConfig{Provider: "anthropic", Name: "claude"},
			Sandbox: SandboxConfig{Runtime: "gvisor"},
		},
	}
	override := &AgentConfig{
		Metadata: Metadata{Name: "override", Labels: map[string]string{"tier": "prod"}},
		Spec:     AgentSpec{Goal: "new goal"},
	}

	merged := MergeConfig(base, override)
	if merged.Metadata.Name != "override" {
		t.Fatalf("expected merged name override, got %q", merged.Metadata.Name)
	}
	if merged.Metadata.Labels["team"] != "a" || merged.Metadata.Labels["tier"] != "prod" {
		t.Fatalf("expected merged labels, got %#v", merged.Metadata.Labels)
	}
	if merged.Spec.Goal != "new goal" {
		t.Fatalf("expected merged goal")
	}
}
