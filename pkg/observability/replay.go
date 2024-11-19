package observability

import "sync"

// ReplayStore stores ordered decision steps by trace id.
type ReplayStore struct {
	mu    sync.RWMutex
	steps map[string][]string
}

// NewReplayStore creates replay store.
func NewReplayStore() *ReplayStore {
	return &ReplayStore{steps: map[string][]string{}}
}

// AddStep appends a replay step.
func (r *ReplayStore) AddStep(traceID, step string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.steps[traceID] = append(r.steps[traceID], step)
}

// Steps returns replay steps.
func (r *ReplayStore) Steps(traceID string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	steps := r.steps[traceID]
	out := make([]string, len(steps))
	copy(out, steps)
	return out
}
