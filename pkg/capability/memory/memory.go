package memory

import (
	"context"
	"fmt"
	"sync"

	"spawn.dev/pkg/capability"
)

// Capability combines vector, graph and key/value stores.
type Capability struct {
	vector *VectorStore
	graph  *GraphStore
	kv     *KVStore
}

// New returns memory capability backed by embedded stores.
func New(path string) (*Capability, error) {
	kv, err := NewKVStore(path)
	if err != nil {
		return nil, fmt.Errorf("new memory capability: %w", err)
	}
	return &Capability{vector: NewVectorStore(), graph: NewGraphStore(), kv: kv}, nil
}

func (c *Capability) Name() string                                             { return "memory" }
func (c *Capability) Version() string                                          { return "v1" }
func (c *Capability) Description() string                                      { return "Persistent vector/graph/kv memory" }
func (c *Capability) Initialize(context.Context, map[string]interface{}) error { return nil }
func (c *Capability) Shutdown(context.Context) error                           { return c.kv.Close() }
func (c *Capability) HealthCheck(context.Context) error                        { return nil }

func (c *Capability) Schema() *capability.Schema {
	return &capability.Schema{Actions: []capability.Action{{Name: "kv_get"}, {Name: "kv_set"}, {Name: "vector_put"}, {Name: "vector_search"}}}
}

func (c *Capability) Execute(ctx context.Context, req *capability.Request) (*capability.Response, error) {
	if req == nil {
		return &capability.Response{Success: false, Error: &capability.Error{Code: "invalid_request", Message: "nil request"}}, nil
	}
	switch req.Action {
	case "kv_set":
		k, _ := req.Params["key"].(string)
		v, _ := req.Params["value"].(string)
		if err := c.kv.Set(ctx, k, []byte(v)); err != nil {
			return &capability.Response{Success: false, Error: &capability.Error{Code: "kv_set_failed", Message: err.Error()}}, nil
		}
		return &capability.Response{Success: true}, nil
	case "kv_get":
		k, _ := req.Params["key"].(string)
		val, err := c.kv.Get(ctx, k)
		if err != nil {
			return &capability.Response{Success: false, Error: &capability.Error{Code: "kv_get_failed", Message: err.Error()}}, nil
		}
		return &capability.Response{Success: true, Data: string(val)}, nil
	case "vector_put":
		k, _ := req.Params["key"].(string)
		vecAny, _ := req.Params["vector"].([]float32)
		if err := c.vector.Put(ctx, k, vecAny); err != nil {
			return &capability.Response{Success: false, Error: &capability.Error{Code: "vector_put_failed", Message: err.Error()}}, nil
		}
		return &capability.Response{Success: true}, nil
	case "vector_search":
		vecAny, _ := req.Params["vector"].([]float32)
		keys, err := c.vector.Search(ctx, vecAny, 5)
		if err != nil {
			return &capability.Response{Success: false, Error: &capability.Error{Code: "vector_search_failed", Message: err.Error()}}, nil
		}
		return &capability.Response{Success: true, Data: keys}, nil
	default:
		return &capability.Response{Success: false, Error: &capability.Error{Code: "invalid_action", Message: req.Action}}, nil
	}
}

// VectorStore is a tiny in-memory vector store.
type VectorStore struct {
	mu   sync.RWMutex
	data map[string][]float32
}

// NewVectorStore returns an in-memory vector store.
func NewVectorStore() *VectorStore {
	return &VectorStore{data: make(map[string][]float32)}
}
