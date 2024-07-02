package tools

import (
	"context"
	"fmt"

	"spawn.dev/pkg/capability"
)

// Tool is an invokable tool definition.
type Tool struct {
	Name        string
	Description string
	Schema      map[string]interface{}
	Handler     func(context.Context, map[string]interface{}) (interface{}, error)
}

// Capability manages tool registration and invocation.
type Capability struct {
	registry map[string]Tool
}

// New returns a tool capability.
func New() *Capability {
	return &Capability{registry: make(map[string]Tool)}
}

func (c *Capability) Name() string                                             { return "tools" }
func (c *Capability) Version() string                                          { return "v1" }
func (c *Capability) Description() string                                      { return "Tool registry and invocation" }
func (c *Capability) Initialize(context.Context, map[string]interface{}) error { return nil }
func (c *Capability) Shutdown(context.Context) error                           { return nil }
func (c *Capability) HealthCheck(context.Context) error                        { return nil }

func (c *Capability) Schema() *capability.Schema {
	return &capability.Schema{Actions: []capability.Action{{Name: "register"}, {Name: "invoke"}, {Name: "list"}}}
}

// Register registers a new tool.
func (c *Capability) Register(tool Tool) error {
	if tool.Name == "" {
		return fmt.Errorf("register tool: name required")
	}
	c.registry[tool.Name] = tool
	return nil
}

// List returns registered tool names.
func (c *Capability) List() []string {
	out := make([]string, 0, len(c.registry))
	for name := range c.registry {
		out = append(out, name)
	}
	return out
}

func (c *Capability) Execute(ctx context.Context, req *capability.Request) (*capability.Response, error) {
	if req == nil {
		return &capability.Response{Success: false, Error: &capability.Error{Code: "invalid_request", Message: "nil request"}}, nil
	}
	switch req.Action {
	case "list":
		return &capability.Response{Success: true, Data: c.List()}, nil
	case "invoke":
		name, _ := req.Params["name"].(string)
		tool, ok := c.registry[name]
		if !ok {
			return &capability.Response{Success: false, Error: &capability.Error{Code: "not_found", Message: name}}, nil
		}
		if tool.Handler == nil {
			return &capability.Response{Success: false, Error: &capability.Error{Code: "no_handler", Message: name}}, nil
		}
		params, _ := req.Params["input"].(map[string]interface{})
		result, err := tool.Handler(ctx, params)
		if err != nil {
			return &capability.Response{Success: false, Error: &capability.Error{Code: "invoke_failed", Message: err.Error()}}, nil
		}
		return &capability.Response{Success: true, Data: result}, nil
	default:
		return &capability.Response{Success: false, Error: &capability.Error{Code: "invalid_action", Message: req.Action}}, nil
	}
}
