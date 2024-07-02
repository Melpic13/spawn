package memory

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"go.etcd.io/bbolt"
)

const kvBucket = "spawn_memory"

// KVStore is a bbolt-backed key/value store.
type KVStore struct {
	db *bbolt.DB
}

// NewKVStore creates or opens a kv store at path.
func NewKVStore(path string) (*KVStore, error) {
	if path == "" {
		path = filepath.Join(os.TempDir(), "spawn-memory.db")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("mkdir kv dir: %w", err)
	}
	db, err := bbolt.Open(path, 0o600, nil)
	if err != nil {
		return nil, fmt.Errorf("open kv db: %w", err)
	}
	if err := db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(kvBucket))
		return err
	}); err != nil {
		return nil, fmt.Errorf("init kv bucket: %w", err)
	}
	return &KVStore{db: db}, nil
}

// Set writes a key/value pair.
func (s *KVStore) Set(_ context.Context, key string, value []byte) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		return tx.Bucket([]byte(kvBucket)).Put([]byte(key), value)
	})
}

// Get reads a value by key.
func (s *KVStore) Get(_ context.Context, key string) ([]byte, error) {
	var out []byte
	err := s.db.View(func(tx *bbolt.Tx) error {
		v := tx.Bucket([]byte(kvBucket)).Get([]byte(key))
		if v == nil {
			out = nil
			return nil
		}
		out = append([]byte(nil), v...)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("get kv: %w", err)
	}
	return out, nil
}

// Close closes the db.
func (s *KVStore) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}
