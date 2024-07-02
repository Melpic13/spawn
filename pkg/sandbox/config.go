package sandbox

import "time"

// DefaultConfig returns secure baseline settings.
func DefaultConfig() *Config {
	return &Config{
		Runtime:      RuntimeGVisor,
		Memory:       256 * 1024 * 1024,
		CPU:          0.5,
		Network:      NetworkRestricted,
		Seccomp:      SeccompStrict,
		ReadOnlyRoot: true,
		StartTimeout: 30 * time.Second,
		ExecTimeout:  2 * time.Minute,
	}
}
