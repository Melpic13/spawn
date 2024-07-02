package browser

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"spawn.dev/pkg/capability"
)

// Capability provides browser automation primitives.
type Capability struct{}

// New returns a browser capability.
func New() *Capability { return &Capability{} }

func (c *Capability) Name() string                                             { return "browser" }
func (c *Capability) Version() string                                          { return "v1" }
func (c *Capability) Description() string                                      { return "Browser automation and capture" }
func (c *Capability) Initialize(context.Context, map[string]interface{}) error { return nil }
func (c *Capability) Shutdown(context.Context) error                           { return nil }
func (c *Capability) HealthCheck(context.Context) error                        { return nil }

func (c *Capability) Schema() *capability.Schema {
	return &capability.Schema{Actions: []capability.Action{{Name: "screenshot"}, {Name: "record"}}}
}

func (c *Capability) Execute(_ context.Context, req *capability.Request) (*capability.Response, error) {
	switch req.Action {
	case "screenshot":
		path, _ := req.Params["path"].(string)
		if path == "" {
			path = filepath.Join(os.TempDir(), "spawn-screenshot.txt")
		}
		if err := os.WriteFile(path, []byte("screenshot placeholder"), 0o644); err != nil {
			return &capability.Response{Success: false, Error: &capability.Error{Code: "write_failed", Message: err.Error()}}, nil
		}
		return &capability.Response{Success: true, Data: path}, nil
	case "record":
		return &capability.Response{Success: true, Data: "recording-started"}, nil
	default:
		return &capability.Response{Success: false, Error: &capability.Error{Code: "invalid_action", Message: fmt.Sprintf("unsupported action: %s", req.Action)}}, nil
	}
}
