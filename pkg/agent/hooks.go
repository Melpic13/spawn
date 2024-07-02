package agent

import "time"

// Hook defines a lifecycle command.
type Hook struct {
	Command []string `yaml:"command" json:"command"`
}

// HealthCheck defines health-check behavior.
type HealthCheck struct {
	Interval time.Duration `yaml:"interval" json:"interval"`
	Timeout  time.Duration `yaml:"timeout" json:"timeout"`
	Command  []string      `yaml:"command" json:"command"`
}

// HooksConfig groups lifecycle hooks.
type HooksConfig struct {
	PreStart    []Hook      `yaml:"preStart" json:"preStart"`
	PostStop    []Hook      `yaml:"postStop" json:"postStop"`
	HealthCheck HealthCheck `yaml:"healthCheck" json:"healthCheck"`
}
