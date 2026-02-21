package fs

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"spawn.dev/pkg/capability"
)

// Capability provides a basic virtual filesystem rooted at baseDir.
type Capability struct {
	baseDir string
	baseAbs string
}

// New returns a filesystem capability.
func New(baseDir string) *Capability {
	if baseDir == "" {
		baseDir = "."
	}
	baseAbs, err := filepath.Abs(baseDir)
	if err != nil {
		baseAbs = baseDir
	}
	return &Capability{baseDir: baseDir, baseAbs: filepath.Clean(baseAbs)}
}

func (c *Capability) Name() string                                             { return "fs" }
func (c *Capability) Version() string                                          { return "v1" }
func (c *Capability) Description() string                                      { return "Virtual filesystem operations" }
func (c *Capability) Initialize(context.Context, map[string]interface{}) error { return nil }
func (c *Capability) Shutdown(context.Context) error                           { return nil }
func (c *Capability) HealthCheck(context.Context) error                        { return nil }

func (c *Capability) Schema() *capability.Schema {
	return &capability.Schema{Actions: []capability.Action{{Name: "read"}, {Name: "write"}, {Name: "copy"}}}
}

func (c *Capability) Execute(_ context.Context, req *capability.Request) (*capability.Response, error) {
	if req == nil {
		return &capability.Response{Success: false, Error: &capability.Error{Code: "invalid_request", Message: "nil request"}}, nil
	}
	switch req.Action {
	case "read":
		path, _ := req.Params["path"].(string)
		resolved, err := c.resolve(path)
		if err != nil {
			return &capability.Response{Success: false, Error: &capability.Error{Code: "access_denied", Message: err.Error()}}, nil
		}
		b, err := os.ReadFile(resolved)
		if err != nil {
			return &capability.Response{Success: false, Error: &capability.Error{Code: "read_failed", Message: err.Error()}}, nil
		}
		return &capability.Response{Success: true, Data: string(b)}, nil
	case "write":
		path, _ := req.Params["path"].(string)
		resolved, err := c.resolve(path)
		if err != nil {
			return &capability.Response{Success: false, Error: &capability.Error{Code: "access_denied", Message: err.Error()}}, nil
		}
		content, _ := req.Params["content"].(string)
		if err := os.MkdirAll(filepath.Dir(resolved), 0o755); err != nil {
			return &capability.Response{Success: false, Error: &capability.Error{Code: "mkdir_failed", Message: err.Error()}}, nil
		}
		if err := os.WriteFile(resolved, []byte(content), 0o644); err != nil {
			return &capability.Response{Success: false, Error: &capability.Error{Code: "write_failed", Message: err.Error()}}, nil
		}
		return &capability.Response{Success: true}, nil
	case "copy":
		src, _ := req.Params["src"].(string)
		dst, _ := req.Params["dst"].(string)
		if src == "" || dst == "" {
			return &capability.Response{Success: false, Error: &capability.Error{Code: "invalid_params", Message: "src and dst required"}}, nil
		}
		resolvedSrc, err := c.resolve(src)
		if err != nil {
			return &capability.Response{Success: false, Error: &capability.Error{Code: "access_denied", Message: err.Error()}}, nil
		}
		resolvedDst, err := c.resolve(dst)
		if err != nil {
			return &capability.Response{Success: false, Error: &capability.Error{Code: "access_denied", Message: err.Error()}}, nil
		}
		if err := c.copy(resolvedSrc, resolvedDst); err != nil {
			return &capability.Response{Success: false, Error: &capability.Error{Code: "copy_failed", Message: err.Error()}}, nil
		}
		return &capability.Response{Success: true}, nil
	default:
		return &capability.Response{Success: false, Error: &capability.Error{Code: "invalid_action", Message: req.Action}}, nil
	}
}

func (c *Capability) resolve(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", fmt.Errorf("path is required")
	}
	cleanRel := filepath.Clean(path)
	resolved := filepath.Join(c.baseAbs, cleanRel)
	absResolved, err := filepath.Abs(resolved)
	if err != nil {
		return "", fmt.Errorf("resolve path: %w", err)
	}
	rel, err := filepath.Rel(c.baseAbs, absResolved)
	if err != nil {
		return "", fmt.Errorf("resolve path: %w", err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path escapes root")
	}
	return absResolved, nil
}

func (c *Capability) copy(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open src: %w", err)
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return fmt.Errorf("mkdir dst: %w", err)
	}
	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create dst: %w", err)
	}
	defer out.Close()
	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("copy bytes: %w", err)
	}
	return nil
}
