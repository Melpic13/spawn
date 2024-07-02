package sandbox

import (
	"context"
	"io"
	"time"
)

// Runtime represents a sandbox runtime implementation.
type Runtime interface {
	Create(ctx context.Context, config *Config) (Sandbox, error)
	List(ctx context.Context) ([]Sandbox, error)
	Supports(feature Feature) bool
	HealthCheck(ctx context.Context) error
}

// Sandbox represents an isolated execution environment.
type Sandbox interface {
	ID() string
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Pause(ctx context.Context) error
	Resume(ctx context.Context) error
	Destroy(ctx context.Context) error
	Exec(ctx context.Context, cmd *Command) (*ExecResult, error)
	CopyIn(ctx context.Context, src string, dst string) error
	CopyOut(ctx context.Context, src string, dst string) error
	NetworkConfig() *NetworkConfig
	State() SandboxState
	Metrics() *SandboxMetrics
	Stdout() io.ReadCloser
	Stderr() io.ReadCloser
	Stdin() io.WriteCloser
}

// Config represents sandbox configuration.
type Config struct {
	Runtime      RuntimeType       `yaml:"runtime"`
	Image        string            `yaml:"image"`
	Memory       int64             `yaml:"memory"`
	CPU          float64           `yaml:"cpu"`
	Disk         int64             `yaml:"disk"`
	Pids         int               `yaml:"pids"`
	Network      NetworkPolicy     `yaml:"network"`
	Seccomp      SeccompProfile    `yaml:"seccomp"`
	Capabilities []string          `yaml:"capabilities"`
	ReadOnlyRoot bool              `yaml:"readOnlyRoot"`
	Mounts       []Mount           `yaml:"mounts"`
	Env          map[string]string `yaml:"env"`
	StartTimeout time.Duration     `yaml:"startTimeout"`
	ExecTimeout  time.Duration     `yaml:"execTimeout"`
}

// Command is a command executed inside a sandbox.
type Command struct {
	Path    string
	Args    []string
	Env     map[string]string
	Timeout time.Duration
}

// ExecResult captures command output and metadata.
type ExecResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
	Duration time.Duration
}

// Mount is a filesystem mount definition.
type Mount struct {
	Source string `yaml:"source"`
	Target string `yaml:"target"`
	Mode   string `yaml:"mode"`
}

// NetworkConfig contains sandbox network settings.
type NetworkConfig struct {
	Policy NetworkPolicy
}

// SandboxMetrics captures current resource usage.
type SandboxMetrics struct {
	CPUPercent   float64
	MemoryBytes  int64
	Uptime       time.Duration
	RestartCount int64
}

// Feature describes runtime features.
type Feature string

const (
	FeaturePause      Feature = "pause"
	FeatureSnapshots  Feature = "snapshots"
	FeatureNetworking Feature = "networking"
)

// SandboxState describes current sandbox lifecycle state.
type SandboxState string

const (
	StateCreated SandboxState = "created"
	StateRunning SandboxState = "running"
	StatePaused  SandboxState = "paused"
	StateStopped SandboxState = "stopped"
)

// RuntimeType identifies a runtime backend.
type RuntimeType string

const (
	RuntimeGVisor      RuntimeType = "gvisor"
	RuntimeFirecracker RuntimeType = "firecracker"
	RuntimeDocker      RuntimeType = "docker"
	RuntimeNative      RuntimeType = "native"
)

// NetworkPolicy controls network access.
type NetworkPolicy string

const (
	NetworkNone       NetworkPolicy = "none"
	NetworkRestricted NetworkPolicy = "restricted"
	NetworkEgressOnly NetworkPolicy = "egress-only"
	NetworkFull       NetworkPolicy = "full"
)

// SeccompProfile defines syscall filter mode.
type SeccompProfile string

const (
	SeccompStrict     SeccompProfile = "strict"
	SeccompModerate   SeccompProfile = "moderate"
	SeccompPermissive SeccompProfile = "permissive"
)
