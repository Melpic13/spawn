package net

import (
	"context"
	"net/http"
	"strings"
	"time"

	"spawn.dev/pkg/capability"
)

// Capability provides HTTP and DNS operations with policy checks.
type Capability struct {
	allow []string
	deny  []string
	http  *http.Client
}

// New returns a network capability.
func New(allowlist, denylist []string) *Capability {
	return &Capability{
		allow: allowlist,
		deny:  denylist,
		http:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Capability) Name() string                                             { return "net" }
func (c *Capability) Version() string                                          { return "v1" }
func (c *Capability) Description() string                                      { return "HTTP and DNS with policy controls" }
func (c *Capability) Initialize(context.Context, map[string]interface{}) error { return nil }
func (c *Capability) Shutdown(context.Context) error                           { return nil }
func (c *Capability) HealthCheck(context.Context) error                        { return nil }

func (c *Capability) Schema() *capability.Schema {
	return &capability.Schema{Actions: []capability.Action{{Name: "get"}, {Name: "resolve"}}}
}

func (c *Capability) Execute(_ context.Context, req *capability.Request) (*capability.Response, error) {
	if req == nil {
		return &capability.Response{Success: false, Error: &capability.Error{Code: "invalid_request", Message: "nil request"}}, nil
	}
	switch req.Action {
	case "get":
		url, _ := req.Params["url"].(string)
		if !c.allowed(url) {
			return &capability.Response{Success: false, Error: &capability.Error{Code: "blocked", Message: "url blocked by policy"}}, nil
		}
		resp, err := c.http.Get(url)
		if err != nil {
			return &capability.Response{Success: false, Error: &capability.Error{Code: "http_failed", Message: err.Error()}}, nil
		}
		defer resp.Body.Close()
		return &capability.Response{Success: true, Data: map[string]interface{}{"status": resp.StatusCode}}, nil
	case "resolve":
		host, _ := req.Params["host"].(string)
		ips, err := Lookup(host)
		if err != nil {
			return &capability.Response{Success: false, Error: &capability.Error{Code: "dns_failed", Message: err.Error()}}, nil
		}
		return &capability.Response{Success: true, Data: ips}, nil
	default:
		return &capability.Response{Success: false, Error: &capability.Error{Code: "invalid_action", Message: req.Action}}, nil
	}
}

func (c *Capability) allowed(url string) bool {
	for _, denied := range c.deny {
		if strings.Contains(url, strings.TrimPrefix(denied, "*")) {
			return false
		}
	}
	if len(c.allow) == 0 {
		return true
	}
	for _, allowed := range c.allow {
		if strings.Contains(url, strings.TrimPrefix(allowed, "*")) {
			return true
		}
	}
	return false
}
