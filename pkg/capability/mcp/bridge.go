package mcp

import "context"

// Bridge forwards native tool invocations to MCP.
type Bridge struct {
	Client *Client
}

// Invoke forwards a call to MCP.
func (b *Bridge) Invoke(ctx context.Context, tool string, input map[string]interface{}) (map[string]interface{}, error) {
	return b.Client.Call(ctx, tool, input)
}
