package memory

import (
	"context"
	"sort"
)

// Put inserts a vector by key.
func (s *VectorStore) Put(_ context.Context, key string, vec []float32) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := make([]float32, len(vec))
	copy(cp, vec)
	s.data[key] = cp
	return nil
}

// Search returns the first N keys (placeholder for ANN search).
func (s *VectorStore) Search(_ context.Context, _ []float32, limit int) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	keys := make([]string, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	if limit > 0 && len(keys) > limit {
		keys = keys[:limit]
	}
	return keys, nil
}
