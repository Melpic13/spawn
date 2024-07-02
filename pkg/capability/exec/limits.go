package exec

import "time"

// Limits define execution resource limits.
type Limits struct {
	MemoryMB int
	CPUCores float64
	Timeout  time.Duration
}
