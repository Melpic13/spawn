package net

import (
	"context"
	"io"
	"net"
	"net/http"
	"net/url"
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

func (c *Capability) Execute(ctx context.Context, req *capability.Request) (*capability.Response, error) {
	if req == nil {
		return &capability.Response{Success: false, Error: &capability.Error{Code: "invalid_request", Message: "nil request"}}, nil
	}
	switch req.Action {
	case "get":
		rawURL, _ := req.Params["url"].(string)
		targetURL, err := url.Parse(rawURL)
		if err != nil || targetURL.Scheme == "" || targetURL.Host == "" {
			return &capability.Response{Success: false, Error: &capability.Error{Code: "invalid_url", Message: "valid absolute url is required"}}, nil
		}
		if targetURL.Scheme != "http" && targetURL.Scheme != "https" {
			return &capability.Response{Success: false, Error: &capability.Error{Code: "invalid_url", Message: "only http/https are allowed"}}, nil
		}
		if !c.allowed(targetURL.Hostname()) {
			return &capability.Response{Success: false, Error: &capability.Error{Code: "blocked", Message: "url blocked by policy"}}, nil
		}
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, targetURL.String(), nil)
		if err != nil {
			return &capability.Response{Success: false, Error: &capability.Error{Code: "http_failed", Message: err.Error()}}, nil
		}
		request.Header.Set("User-Agent", "spawn-net-capability/1.0")
		resp, err := c.http.Do(request)
		if err != nil {
			return &capability.Response{Success: false, Error: &capability.Error{Code: "http_failed", Message: err.Error()}}, nil
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(io.LimitReader(resp.Body, 8*1024))
		if err != nil {
			return &capability.Response{Success: false, Error: &capability.Error{Code: "http_failed", Message: err.Error()}}, nil
		}
		return &capability.Response{Success: true, Data: map[string]interface{}{
			"status":     resp.StatusCode,
			"host":       targetURL.Hostname(),
			"headers":    resp.Header,
			"body":       string(body),
			"body_bytes": len(body),
		}}, nil
	case "resolve":
		host, _ := req.Params["host"].(string)
		if !c.allowed(host) {
			return &capability.Response{Success: false, Error: &capability.Error{Code: "blocked", Message: "host blocked by policy"}}, nil
		}
		ips, err := Lookup(host)
		if err != nil {
			return &capability.Response{Success: false, Error: &capability.Error{Code: "dns_failed", Message: err.Error()}}, nil
		}
		return &capability.Response{Success: true, Data: ips}, nil
	default:
		return &capability.Response{Success: false, Error: &capability.Error{Code: "invalid_action", Message: req.Action}}, nil
	}
}

func (c *Capability) allowed(host string) bool {
	normalizedHost := normalizeHost(host)
	if normalizedHost == "" {
		return false
	}
	for _, denied := range c.deny {
		if matchesDomain(normalizedHost, denied) {
			return false
		}
	}
	if len(c.allow) == 0 {
		return true
	}
	for _, allowed := range c.allow {
		if matchesDomain(normalizedHost, allowed) {
			return true
		}
	}
	return false
}

func normalizeHost(host string) string {
	host = strings.TrimSpace(host)
	if host == "" {
		return ""
	}
	if parsedHost, _, err := net.SplitHostPort(host); err == nil {
		host = parsedHost
	}
	return strings.TrimSuffix(strings.ToLower(host), ".")
}

func matchesDomain(host, pattern string) bool {
	pattern = normalizeHost(pattern)
	if pattern == "" {
		return false
	}
	if pattern == "*" {
		return true
	}
	if strings.HasPrefix(pattern, "*.") {
		suffix := strings.TrimPrefix(pattern, "*.")
		return host == suffix || strings.HasSuffix(host, "."+suffix)
	}
	return host == pattern
}
