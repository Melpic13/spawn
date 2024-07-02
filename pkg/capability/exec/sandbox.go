package exec

// SandboxMode is the isolation mode for command execution.
type SandboxMode string

const (
	// SandboxNative executes directly on host.
	SandboxNative SandboxMode = "native"
	// SandboxDocker executes via docker wrapper.
	SandboxDocker SandboxMode = "docker"
	// SandboxGVisor executes via runsc wrapper.
	SandboxGVisor SandboxMode = "gvisor"
)
