package fs

import (
	"context"
	"path/filepath"
	"testing"

	"spawn.dev/pkg/capability"
)

func TestFSWriteReadAndTraversalProtection(t *testing.T) {
	t.Parallel()
	base := t.TempDir()
	cap := New(base)

	resp, err := cap.Execute(context.Background(), &capability.Request{
		Action: "write",
		Params: map[string]interface{}{"path": "safe/file.txt", "content": "hello"},
	})
	if err != nil {
		t.Fatalf("write execute: %v", err)
	}
	if !resp.Success {
		t.Fatalf("write failed: %#v", resp.Error)
	}

	resp, err = cap.Execute(context.Background(), &capability.Request{
		Action: "read",
		Params: map[string]interface{}{"path": "safe/file.txt"},
	})
	if err != nil {
		t.Fatalf("read execute: %v", err)
	}
	if got := resp.Data.(string); got != "hello" {
		t.Fatalf("unexpected read contents %q", got)
	}

	resp, err = cap.Execute(context.Background(), &capability.Request{
		Action: "read",
		Params: map[string]interface{}{"path": filepath.Join("..", "escape.txt")},
	})
	if err != nil {
		t.Fatalf("traversal execute: %v", err)
	}
	if resp.Success {
		t.Fatalf("expected traversal to be blocked")
	}
}
