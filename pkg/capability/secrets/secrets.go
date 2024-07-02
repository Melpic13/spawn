package secrets

import (
	"context"
	"os"
	"strings"
)

// Resolver resolves secret references.
type Resolver interface {
	Resolve(ctx context.Context, ref string) (string, error)
}

// EnvResolver resolves env:// refs.
type EnvResolver struct{}

// Resolve resolves env references.
func (EnvResolver) Resolve(_ context.Context, ref string) (string, error) {
	name := strings.TrimPrefix(ref, "env://")
	return os.Getenv(name), nil
}
