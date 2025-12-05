package security

import (
	"os"
	"testing"

	"spawn.dev/pkg/agent"
)

func FuzzAgentConfigLoad(f *testing.F) {
	f.Add("apiVersion: spawn.dev/v1\nkind: Agent\nmetadata:\n  name: fuzz\nspec:\n  model:\n    provider: anthropic\n    name: claude\n  sandbox:\n    runtime: gvisor\n")
	f.Fuzz(func(t *testing.T, body string) {
		path := t.TempDir() + "/agent.yaml"
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatalf("write fuzz file: %v", err)
		}
		_, _ = agent.LoadConfig(path)
	})
}
