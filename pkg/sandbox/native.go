package sandbox

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/google/uuid"
)

// NativeRuntime executes commands directly on host.
type NativeRuntime struct {
	sandboxes map[string]*nativeSandbox
}

// NewNativeRuntime creates native runtime.
func NewNativeRuntime() *NativeRuntime {
	return &NativeRuntime{sandboxes: map[string]*nativeSandbox{}}
}

func (r *NativeRuntime) Create(_ context.Context, config *Config) (Sandbox, error) {
	s := &nativeSandbox{
		id:      uuid.NewString(),
		config:  config,
		state:   StateCreated,
		started: time.Now(),
	}
	r.sandboxes[s.id] = s
	return s, nil
}

func (r *NativeRuntime) List(_ context.Context) ([]Sandbox, error) {
	out := make([]Sandbox, 0, len(r.sandboxes))
	for _, sb := range r.sandboxes {
		out = append(out, sb)
	}
	return out, nil
}

func (r *NativeRuntime) Supports(feature Feature) bool {
	switch feature {
	case FeatureNetworking:
		return true
	default:
		return false
	}
}

func (r *NativeRuntime) HealthCheck(context.Context) error { return nil }

type nativeSandbox struct {
	id      string
	config  *Config
	state   SandboxState
	started time.Time
}

func (s *nativeSandbox) ID() string { return s.id }

func (s *nativeSandbox) Start(context.Context) error {
	s.state = StateRunning
	s.started = time.Now()
	return nil
}

func (s *nativeSandbox) Stop(context.Context) error {
	s.state = StateStopped
	return nil
}

func (s *nativeSandbox) Pause(context.Context) error {
	s.state = StatePaused
	return nil
}

func (s *nativeSandbox) Resume(context.Context) error {
	s.state = StateRunning
	return nil
}

func (s *nativeSandbox) Destroy(context.Context) error {
	s.state = StateStopped
	return nil
}

func (s *nativeSandbox) Exec(ctx context.Context, cmd *Command) (*ExecResult, error) {
	if cmd == nil || cmd.Path == "" {
		return nil, fmt.Errorf("exec sandbox command: path is required")
	}
	start := time.Now()
	timeout := cmd.Timeout
	if timeout <= 0 && s.config != nil {
		timeout = s.config.ExecTimeout
	}
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ec := exec.CommandContext(runCtx, cmd.Path, cmd.Args...)
	env := os.Environ()
	for k, v := range cmd.Env {
		env = append(env, k+"="+v)
	}
	ec.Env = env
	out, err := ec.CombinedOutput()
	res := &ExecResult{Stdout: string(out), Duration: time.Since(start)}
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			res.ExitCode = exitErr.ExitCode()
			res.Stderr = err.Error()
			return res, nil
		}
		return nil, fmt.Errorf("exec command: %w", err)
	}
	res.ExitCode = 0
	return res, nil
}

func (s *nativeSandbox) CopyIn(context.Context, string, string) error  { return nil }
func (s *nativeSandbox) CopyOut(context.Context, string, string) error { return nil }
func (s *nativeSandbox) NetworkConfig() *NetworkConfig {
	return &NetworkConfig{Policy: s.config.Network}
}
func (s *nativeSandbox) State() SandboxState { return s.state }
func (s *nativeSandbox) Metrics() *SandboxMetrics {
	return &SandboxMetrics{Uptime: time.Since(s.started)}
}
func (s *nativeSandbox) Stdout() io.ReadCloser { return io.NopCloser(nilReader{}) }
func (s *nativeSandbox) Stderr() io.ReadCloser { return io.NopCloser(nilReader{}) }
func (s *nativeSandbox) Stdin() io.WriteCloser { return nopWriteCloser{Writer: io.Discard} }

type nilReader struct{}

func (nilReader) Read(_ []byte) (int, error) { return 0, io.EOF }

type nopWriteCloser struct{ io.Writer }

func (nopWriteCloser) Close() error { return nil }
