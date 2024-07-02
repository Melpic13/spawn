package exec

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"spawn.dev/pkg/capability"
)

// Capability provides local command execution with timeout.
type Capability struct {
	languages map[string]struct{}
}

// New returns an exec capability with language allowlist.
func New(langs []string) *Capability {
	m := make(map[string]struct{}, len(langs))
	for _, l := range langs {
		m[l] = struct{}{}
	}
	return &Capability{languages: m}
}

func (c *Capability) Name() string                                             { return "exec" }
func (c *Capability) Version() string                                          { return "v1" }
func (c *Capability) Description() string                                      { return "Execute sandboxed commands" }
func (c *Capability) Initialize(context.Context, map[string]interface{}) error { return nil }
func (c *Capability) Shutdown(context.Context) error                           { return nil }
func (c *Capability) HealthCheck(context.Context) error                        { return nil }

func (c *Capability) Schema() *capability.Schema {
	return &capability.Schema{Actions: []capability.Action{{Name: "run", Description: "Run command"}}}
}

func (c *Capability) Execute(ctx context.Context, req *capability.Request) (*capability.Response, error) {
	if req == nil || req.Action != "run" {
		return &capability.Response{Success: false, Error: &capability.Error{Code: "invalid_action", Message: "expected run"}}, nil
	}
	lang, _ := req.Params["language"].(string)
	if lang != "" {
		if _, ok := c.languages[lang]; !ok {
			return &capability.Response{Success: false, Error: &capability.Error{Code: "language_not_allowed", Message: lang}}, nil
		}
	}
	cmdText, _ := req.Params["cmd"].(string)
	if cmdText == "" {
		return &capability.Response{Success: false, Error: &capability.Error{Code: "missing_cmd", Message: "cmd is required"}}, nil
	}

	timeout := req.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(runCtx, "sh", "-lc", cmdText)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return &capability.Response{Success: false, Data: string(out), Error: &capability.Error{Code: "exec_failed", Message: fmt.Sprintf("%v", err)}}, nil
	}
	return &capability.Response{Success: true, Data: string(out)}, nil
}
