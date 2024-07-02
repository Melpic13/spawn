package sandbox

import "context"

// FirecrackerRuntime wraps microVM execution.
type FirecrackerRuntime struct {
	native *NativeRuntime
	Binary string
}

// NewFirecrackerRuntime returns a firecracker runtime.
func NewFirecrackerRuntime(binary string) *FirecrackerRuntime {
	return &FirecrackerRuntime{native: NewNativeRuntime(), Binary: binary}
}

func (r *FirecrackerRuntime) Create(ctx context.Context, cfg *Config) (Sandbox, error) {
	return r.native.Create(ctx, cfg)
}

func (r *FirecrackerRuntime) List(ctx context.Context) ([]Sandbox, error) {
	return r.native.List(ctx)
}

func (r *FirecrackerRuntime) Supports(feature Feature) bool {
	return r.native.Supports(feature)
}

func (r *FirecrackerRuntime) HealthCheck(ctx context.Context) error {
	return r.native.HealthCheck(ctx)
}
