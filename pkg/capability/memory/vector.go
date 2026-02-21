package memory

import (
	"context"
	"math"
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

// Search returns top keys ranked by cosine similarity.
func (s *VectorStore) Search(_ context.Context, query []float32, limit int) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if limit <= 0 {
		limit = 5
	}
	type scored struct {
		key   string
		score float64
	}
	scores := make([]scored, 0, len(s.data))
	for key, vec := range s.data {
		score := cosine(query, vec)
		scores = append(scores, scored{key: key, score: score})
	}
	sort.Slice(scores, func(i, j int) bool {
		if scores[i].score == scores[j].score {
			return scores[i].key < scores[j].key
		}
		return scores[i].score > scores[j].score
	})
	if len(scores) > limit {
		scores = scores[:limit]
	}
	out := make([]string, 0, len(scores))
	for _, item := range scores {
		out = append(out, item.key)
	}
	return out, nil
}

func cosine(a, b []float32) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	max := len(a)
	if len(b) < max {
		max = len(b)
	}
	var dot float64
	var magA float64
	var magB float64
	for i := 0; i < max; i++ {
		av := float64(a[i])
		bv := float64(b[i])
		dot += av * bv
		magA += av * av
		magB += bv * bv
	}
	if magA == 0 || magB == 0 {
		return 0
	}
	return dot / (math.Sqrt(magA) * math.Sqrt(magB))
}
