package tools

import (
	"context"
	"fmt"

	"spawn.dev/pkg/capability"
)

// Invoke invokes a tool by name.
func (c *Capability) Invoke(ctx context.Context, name string, input map[string]interface{}) (interface{}, error) {
	resp, err := c.Execute(ctx, &capability.Request{
		Action: "invoke",
		Params: map[string]interface{}{"name": name, "input": input},
	})
	if err != nil {
		return nil, err
	}
	if !resp.Success {
		if resp.Error != nil {
			return nil, fmt.Errorf("invoke tool: %s", resp.Error.Message)
		}
		return nil, fmt.Errorf("invoke tool: invocation failed")
	}
	return resp.Data, nil
}
