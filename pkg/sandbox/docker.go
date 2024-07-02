package sandbox

import "context"

// DockerRuntime is a docker-backed runtime wrapper.
type DockerRuntime struct {
	native *NativeRuntime
}

// NewDockerRuntime creates a DockerRuntime.
func NewDockerRuntime() *DockerRuntime {
	return &DockerRuntime{native: NewNativeRuntime()}
}

func (r *DockerRuntime) Create(ctx context.Context, cfg *Config) (Sandbox, error) {
	return r.native.Create(ctx, cfg)
}

func (r *DockerRuntime) List(ctx context.Context) ([]Sandbox, error) {
	return r.native.List(ctx)
}

func (r *DockerRuntime) Supports(feature Feature) bool {
	return r.native.Supports(feature)
}

func (r *DockerRuntime) HealthCheck(ctx context.Context) error {
	return r.native.HealthCheck(ctx)
}
