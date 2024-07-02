package sandbox

import "context"

// GVisorRuntime wraps runsc compatible execution.
type GVisorRuntime struct {
	native *NativeRuntime
	Binary string
}

// NewGVisorRuntime returns a gVisor runtime.
func NewGVisorRuntime(binary string) *GVisorRuntime {
	return &GVisorRuntime{native: NewNativeRuntime(), Binary: binary}
}

func (r *GVisorRuntime) Create(ctx context.Context, cfg *Config) (Sandbox, error) {
	return r.native.Create(ctx, cfg)
}

func (r *GVisorRuntime) List(ctx context.Context) ([]Sandbox, error) {
	return r.native.List(ctx)
}

func (r *GVisorRuntime) Supports(feature Feature) bool {
	return r.native.Supports(feature)
}

func (r *GVisorRuntime) HealthCheck(ctx context.Context) error {
	return r.native.HealthCheck(ctx)
}
