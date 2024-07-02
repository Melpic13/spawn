package mcp

import (
	"context"
	"fmt"
)

// Client is a minimal MCP client abstraction.
type Client struct {
	Endpoint string
}

// Call invokes an MCP tool endpoint.
func (c *Client) Call(_ context.Context, tool string, payload map[string]interface{}) (map[string]interface{}, error) {
	if c.Endpoint == "" {
		return nil, fmt.Errorf("mcp call: endpoint is required")
	}
	return map[string]interface{}{
		"tool":     tool,
		"payload":  payload,
		"endpoint": c.Endpoint,
	}, nil
}
