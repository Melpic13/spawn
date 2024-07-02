package config

import "fmt"

// Validate validates daemon configuration.
func Validate(cfg *DaemonConfig) error {
	if cfg == nil {
		return fmt.Errorf("validate config: nil config")
	}
	if cfg.APIVersion == "" {
		return fmt.Errorf("validate config: apiVersion is required")
	}
	if cfg.Kind == "" {
		return fmt.Errorf("validate config: kind is required")
	}
	if cfg.Server.Ports.GRPC <= 0 {
		return fmt.Errorf("validate config: server.ports.grpc must be > 0")
	}
	if cfg.Server.Ports.REST <= 0 {
		return fmt.Errorf("validate config: server.ports.rest must be > 0")
	}
	if cfg.Sandbox.DefaultRuntime == "" {
		return fmt.Errorf("validate config: sandbox.defaultRuntime is required")
	}
	return nil
}
