package memory

import (
	"context"
	"testing"
)

func TestVectorSearchCosineRanking(t *testing.T) {
	t.Parallel()
	store := NewVectorStore()
	_ = store.Put(context.Background(), "a", []float32{1, 0})
	_ = store.Put(context.Background(), "b", []float32{0, 1})
	_ = store.Put(context.Background(), "c", []float32{1, 1})

	keys, err := store.Search(context.Background(), []float32{1, 0}, 2)
	if err != nil {
		t.Fatalf("search vectors: %v", err)
	}
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
	if keys[0] != "a" {
		t.Fatalf("expected top match to be a, got %s", keys[0])
	}
}
