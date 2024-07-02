package sandbox

import (
	"context"
	"testing"
)

func TestNativeSandboxExec(t *testing.T) {
	t.Parallel()
	r := NewNativeRuntime()
	sb, err := r.Create(context.Background(), DefaultConfig())
	if err != nil {
		t.Fatalf("create sandbox: %v", err)
	}
	if err := sb.Start(context.Background()); err != nil {
		t.Fatalf("start sandbox: %v", err)
	}
	res, err := sb.Exec(context.Background(), &Command{Path: "sh", Args: []string{"-lc", "echo hello"}})
	if err != nil {
		t.Fatalf("exec sandbox: %v", err)
	}
	if res.ExitCode != 0 {
		t.Fatalf("expected exit code 0, got %d", res.ExitCode)
	}
}
