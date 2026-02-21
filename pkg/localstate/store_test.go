package localstate

import (
	"testing"
	"time"
)

func TestStoreUpdateAndLoad(t *testing.T) {
	t.Parallel()
	store := OpenAt(t.TempDir() + "/state.json")
	if err := store.Update(func(st *State) error {
		st.Daemon.Running = true
		st.Logs = append(st.Logs, LogEntry{Time: time.Now().UTC(), Level: "info", Message: "started"})
		st.Agents["alpha"] = AgentRecord{Name: "alpha", State: "running", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
		return nil
	}); err != nil {
		t.Fatalf("update state: %v", err)
	}

	st, err := store.Load()
	if err != nil {
		t.Fatalf("load state: %v", err)
	}
	if !st.Daemon.Running {
		t.Fatalf("expected daemon running")
	}
	if len(st.Logs) != 1 {
		t.Fatalf("expected 1 log, got %d", len(st.Logs))
	}
	if _, ok := st.Agents["alpha"]; !ok {
		t.Fatalf("expected alpha agent")
	}
}

func TestStoreDedupesCapabilities(t *testing.T) {
	t.Parallel()
	store := OpenAt(t.TempDir() + "/state.json")
	if err := store.Update(func(st *State) error {
		st.InstalledCapabilities = []string{"net", "exec", "net"}
		return nil
	}); err != nil {
		t.Fatalf("update state: %v", err)
	}
	st, err := store.Load()
	if err != nil {
		t.Fatalf("load state: %v", err)
	}
	if len(st.InstalledCapabilities) != 2 {
		t.Fatalf("expected deduped capability list, got %#v", st.InstalledCapabilities)
	}
}
