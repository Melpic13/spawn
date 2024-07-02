package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"gopkg.in/yaml.v3"
)

// AgentConfig is the top-level agent configuration.
type AgentConfig struct {
	APIVersion string    `yaml:"apiVersion" json:"apiVersion"`
	Kind       string    `yaml:"kind" json:"kind"`
	Metadata   Metadata  `yaml:"metadata" json:"metadata"`
	Spec       AgentSpec `yaml:"spec" json:"spec"`
}

// Metadata contains identifying labels/annotations.
type Metadata struct {
	Name        string            `yaml:"name" json:"name"`
	Namespace   string            `yaml:"namespace" json:"namespace"`
	Labels      map[string]string `yaml:"labels" json:"labels"`
	Annotations map[string]string `yaml:"annotations" json:"annotations"`
}

// AgentSpec contains runtime behavior settings.
type AgentSpec struct {
	Model         ModelConfig         `yaml:"model" json:"model"`
	System        string              `yaml:"system" json:"system"`
	Goal          string              `yaml:"goal" json:"goal"`
	Capabilities  CapabilitiesConfig  `yaml:"capabilities" json:"capabilities"`
	Resources     ResourceConfig      `yaml:"resources" json:"resources"`
	Sandbox       SandboxConfig       `yaml:"sandbox" json:"sandbox"`
	Hooks         HooksConfig         `yaml:"hooks" json:"hooks"`
	Observability ObservabilityConfig `yaml:"observability" json:"observability"`
	Scaling       ScalingConfig       `yaml:"scaling" json:"scaling"`
	Mesh          MeshConfig          `yaml:"mesh" json:"mesh"`
}

// ModelConfig defines model provider settings.
type ModelConfig struct {
	Provider    string          `yaml:"provider" json:"provider"`
	Name        string          `yaml:"name" json:"name"`
	Temperature float64         `yaml:"temperature" json:"temperature"`
	MaxTokens   int             `yaml:"maxTokens" json:"maxTokens"`
	Fallback    []FallbackModel `yaml:"fallback" json:"fallback"`
}

// FallbackModel defines fallback provider and model.
type FallbackModel struct {
	Provider string `yaml:"provider" json:"provider"`
	Name     string `yaml:"name" json:"name"`
}

// CapabilitiesConfig contains all capability-specific settings.
type CapabilitiesConfig struct {
	Exec    ExecConfig    `yaml:"exec" json:"exec"`
	FS      FSConfig      `yaml:"fs" json:"fs"`
	Net     NetConfig     `yaml:"net" json:"net"`
	Browser BrowserConfig `yaml:"browser" json:"browser"`
	Memory  MemoryConfig  `yaml:"memory" json:"memory"`
	Tools   ToolsConfig   `yaml:"tools" json:"tools"`
	Secrets SecretsConfig `yaml:"secrets" json:"secrets"`
}

// ExecConfig configures execution capability.
type ExecConfig struct {
	Enabled   bool     `yaml:"enabled" json:"enabled"`
	Languages []string `yaml:"languages" json:"languages"`
	Timeout   string   `yaml:"timeout" json:"timeout"`
	Memory    string   `yaml:"memory" json:"memory"`
	CPU       string   `yaml:"cpu" json:"cpu"`
}

// FSConfig configures filesystem capability.
type FSConfig struct {
	Enabled bool      `yaml:"enabled" json:"enabled"`
	Mounts  []FSMount `yaml:"mounts" json:"mounts"`
}

// FSMount defines a mounted path.
type FSMount struct {
	Path   string `yaml:"path" json:"path"`
	Source string `yaml:"source" json:"source"`
	Mode   string `yaml:"mode" json:"mode"`
	Quota  string `yaml:"quota" json:"quota"`
}

// NetConfig configures network policy.
type NetConfig struct {
	Enabled   bool      `yaml:"enabled" json:"enabled"`
	Allowlist []string  `yaml:"allowlist" json:"allowlist"`
	Denylist  []string  `yaml:"denylist" json:"denylist"`
	RateLimit RateLimit `yaml:"rateLimit" json:"rateLimit"`
}

// RateLimit defines request limits.
type RateLimit struct {
	Requests int    `yaml:"requests" json:"requests"`
	Per      string `yaml:"per" json:"per"`
}

// BrowserConfig configures browser capability.
type BrowserConfig struct {
	Enabled  bool     `yaml:"enabled" json:"enabled"`
	Headless bool     `yaml:"headless" json:"headless"`
	Stealth  bool     `yaml:"stealth" json:"stealth"`
	Timeout  string   `yaml:"timeout" json:"timeout"`
	Viewport Viewport `yaml:"viewport" json:"viewport"`
}

// Viewport defines browser viewport size.
type Viewport struct {
	Width  int `yaml:"width" json:"width"`
	Height int `yaml:"height" json:"height"`
}

// MemoryConfig configures memory capability.
type MemoryConfig struct {
	Enabled bool         `yaml:"enabled" json:"enabled"`
	Vector  VectorConfig `yaml:"vector" json:"vector"`
	Graph   GraphConfig  `yaml:"graph" json:"graph"`
	TTL     string       `yaml:"ttl" json:"ttl"`
}

// VectorConfig defines vector settings.
type VectorConfig struct {
	Dimensions int    `yaml:"dimensions" json:"dimensions"`
	Metric     string `yaml:"metric" json:"metric"`
}

// GraphConfig defines graph settings.
type GraphConfig struct {
	Enabled bool `yaml:"enabled" json:"enabled"`
}

// ToolsConfig configures tool capability.
type ToolsConfig struct {
	Enabled bool         `yaml:"enabled" json:"enabled"`
	Builtin []string     `yaml:"builtin" json:"builtin"`
	MCP     []MCPTool    `yaml:"mcp" json:"mcp"`
	Custom  []CustomTool `yaml:"custom" json:"custom"`
}

// MCPTool defines an MCP tool endpoint.
type MCPTool struct {
	URI  string `yaml:"uri" json:"uri"`
	Name string `yaml:"name" json:"name"`
}

// CustomTool defines a custom tool schema+handler.
type CustomTool struct {
	Name        string                 `yaml:"name" json:"name"`
	Description string                 `yaml:"description" json:"description"`
	Schema      map[string]interface{} `yaml:"schema" json:"schema"`
	Handler     string                 `yaml:"handler" json:"handler"`
}

// SecretsConfig configures secret injection.
type SecretsConfig struct {
	Enabled bool            `yaml:"enabled" json:"enabled"`
	Inject  []SecretBinding `yaml:"inject" json:"inject"`
}

// SecretBinding binds env var names to secret refs.
type SecretBinding struct {
	Name   string `yaml:"name" json:"name"`
	Source string `yaml:"source" json:"source"`
}

// ResourceConfig configures resource requests/limits/cost controls.
type ResourceConfig struct {
	Requests  ResourceValues `yaml:"requests" json:"requests"`
	Limits    ResourceValues `yaml:"limits" json:"limits"`
	CostLimit CostLimit      `yaml:"costLimit" json:"costLimit"`
}

// ResourceValues stores cpu/memory values.
type ResourceValues struct {
	Memory string `yaml:"memory" json:"memory"`
	CPU    string `yaml:"cpu" json:"cpu"`
}

// CostLimit stores budget constraints.
type CostLimit struct {
	Daily    float64 `yaml:"daily" json:"daily"`
	Monthly  float64 `yaml:"monthly" json:"monthly"`
	Currency string  `yaml:"currency" json:"currency"`
}

// SandboxConfig configures agent sandboxing.
type SandboxConfig struct {
	Runtime        string `yaml:"runtime" json:"runtime"`
	NetworkPolicy  string `yaml:"networkPolicy" json:"networkPolicy"`
	SeccompProfile string `yaml:"seccompProfile" json:"seccompProfile"`
}

// ObservabilityConfig configures traces/metrics/logs/events.
type ObservabilityConfig struct {
	Traces struct {
		Enabled    bool    `yaml:"enabled" json:"enabled"`
		SampleRate float64 `yaml:"sampleRate" json:"sampleRate"`
	} `yaml:"traces" json:"traces"`
	Metrics struct {
		Enabled bool `yaml:"enabled" json:"enabled"`
	} `yaml:"metrics" json:"metrics"`
	Logs struct {
		Level  string `yaml:"level" json:"level"`
		Format string `yaml:"format" json:"format"`
	} `yaml:"logs" json:"logs"`
	Events struct {
		Stream bool `yaml:"stream" json:"stream"`
	} `yaml:"events" json:"events"`
}

// ScalingConfig configures autoscaling behavior.
type ScalingConfig struct {
	MinReplicas int             `yaml:"minReplicas" json:"minReplicas"`
	MaxReplicas int             `yaml:"maxReplicas" json:"maxReplicas"`
	Metrics     []ScalingMetric `yaml:"metrics" json:"metrics"`
}

// ScalingMetric defines one scaling target metric.
type ScalingMetric struct {
	Type   string  `yaml:"type" json:"type"`
	Target float64 `yaml:"target" json:"target"`
}

// MeshConfig configures inter-agent channels.
type MeshConfig struct {
	Channels []MeshChannel `yaml:"channels" json:"channels"`
}

// MeshChannel defines one mesh channel.
type MeshChannel struct {
	Name    string `yaml:"name" json:"name"`
	Type    string `yaml:"type" json:"type"`
	Topic   string `yaml:"topic" json:"topic"`
	Timeout string `yaml:"timeout" json:"timeout"`
}

// LoadConfig reads and validates an agent config file.
func LoadConfig(path string) (*AgentConfig, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("load agent config: %w", err)
	}
	var cfg AgentConfig
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, fmt.Errorf("decode agent config: %w", err)
	}
	if err := ValidateConfig(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// ValidateConfig validates required fields.
func ValidateConfig(cfg *AgentConfig) error {
	if cfg == nil {
		return fmt.Errorf("validate agent config: nil config")
	}
	if cfg.APIVersion == "" {
		return fmt.Errorf("validate agent config: apiVersion is required")
	}
	if cfg.Kind != "Agent" {
		return fmt.Errorf("validate agent config: kind must be Agent")
	}
	if cfg.Metadata.Name == "" {
		return fmt.Errorf("validate agent config: metadata.name is required")
	}
	if cfg.Spec.Model.Provider == "" || cfg.Spec.Model.Name == "" {
		return fmt.Errorf("validate agent config: spec.model provider and name are required")
	}
	if cfg.Spec.Sandbox.Runtime == "" {
		return fmt.Errorf("validate agent config: spec.sandbox.runtime is required")
	}
	return nil
}

// GenerateJSONSchema generates a minimal JSON schema for editor support.
func GenerateJSONSchema() ([]byte, error) {
	schema := map[string]interface{}{
		"$schema":  "https://json-schema.org/draft/2020-12/schema",
		"title":    "spawn Agent",
		"type":     "object",
		"required": []string{"apiVersion", "kind", "metadata", "spec"},
		"properties": map[string]interface{}{
			"apiVersion": map[string]interface{}{"type": "string", "const": "spawn.dev/v1"},
			"kind":       map[string]interface{}{"type": "string", "const": "Agent"},
			"metadata": map[string]interface{}{
				"type":     "object",
				"required": []string{"name"},
			},
			"spec": map[string]interface{}{"type": "object"},
		},
	}
	b, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("generate schema: %w", err)
	}
	return b, nil
}

// MergeConfig merges child over parent.
func MergeConfig(parent, child *AgentConfig) *AgentConfig {
	if parent == nil {
		return child
	}
	if child == nil {
		return parent
	}
	merged := *parent

	if child.APIVersion != "" {
		merged.APIVersion = child.APIVersion
	}
	if child.Kind != "" {
		merged.Kind = child.Kind
	}

	merged.Metadata = parent.Metadata
	if child.Metadata.Name != "" {
		merged.Metadata.Name = child.Metadata.Name
	}
	if child.Metadata.Namespace != "" {
		merged.Metadata.Namespace = child.Metadata.Namespace
	}
	merged.Metadata.Labels = mergeStringMap(parent.Metadata.Labels, child.Metadata.Labels)
	merged.Metadata.Annotations = mergeStringMap(parent.Metadata.Annotations, child.Metadata.Annotations)

	merged.Spec = parent.Spec
	if child.Spec.Model.Provider != "" {
		merged.Spec.Model = child.Spec.Model
	}
	if child.Spec.System != "" {
		merged.Spec.System = child.Spec.System
	}
	if child.Spec.Goal != "" {
		merged.Spec.Goal = child.Spec.Goal
	}
	if child.Spec.Sandbox.Runtime != "" {
		merged.Spec.Sandbox = child.Spec.Sandbox
	}
	if len(child.Spec.Mesh.Channels) > 0 {
		merged.Spec.Mesh = child.Spec.Mesh
	}
	if hasCapabilitiesConfig(child.Spec.Capabilities) {
		merged.Spec.Capabilities = child.Spec.Capabilities
	}
	if hasResourceConfig(child.Spec.Resources) {
		merged.Spec.Resources = child.Spec.Resources
	}
	if hasObservabilityConfig(child.Spec.Observability) {
		merged.Spec.Observability = child.Spec.Observability
	}
	if hasScalingConfig(child.Spec.Scaling) {
		merged.Spec.Scaling = child.Spec.Scaling
	}
	if hasHooksConfig(child.Spec.Hooks) {
		merged.Spec.Hooks = child.Spec.Hooks
	}
	return &merged
}

func mergeStringMap(a, b map[string]string) map[string]string {
	if len(a) == 0 && len(b) == 0 {
		return nil
	}
	out := make(map[string]string, len(a)+len(b))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		out[k] = v
	}
	return out
}

func hasCapabilitiesConfig(c CapabilitiesConfig) bool {
	return c.Exec.Enabled ||
		c.FS.Enabled ||
		c.Net.Enabled ||
		c.Browser.Enabled ||
		c.Memory.Enabled ||
		c.Tools.Enabled ||
		c.Secrets.Enabled ||
		len(c.Exec.Languages) > 0 ||
		len(c.FS.Mounts) > 0 ||
		len(c.Net.Allowlist) > 0 ||
		len(c.Net.Denylist) > 0 ||
		len(c.Tools.Builtin) > 0 ||
		len(c.Tools.MCP) > 0 ||
		len(c.Tools.Custom) > 0 ||
		len(c.Secrets.Inject) > 0
}

func hasResourceConfig(r ResourceConfig) bool {
	return r.Requests.Memory != "" ||
		r.Requests.CPU != "" ||
		r.Limits.Memory != "" ||
		r.Limits.CPU != "" ||
		r.CostLimit.Daily > 0 ||
		r.CostLimit.Monthly > 0 ||
		r.CostLimit.Currency != ""
}

func hasObservabilityConfig(o ObservabilityConfig) bool {
	return o.Traces.Enabled ||
		o.Traces.SampleRate > 0 ||
		o.Metrics.Enabled ||
		o.Logs.Level != "" ||
		o.Logs.Format != "" ||
		o.Events.Stream
}

func hasScalingConfig(s ScalingConfig) bool {
	return s.MinReplicas > 0 || s.MaxReplicas > 0 || len(s.Metrics) > 0
}

func hasHooksConfig(h HooksConfig) bool {
	return len(h.PreStart) > 0 || len(h.PostStop) > 0 || len(h.HealthCheck.Command) > 0 || h.HealthCheck.Interval > 0 || h.HealthCheck.Timeout > 0
}

// CapabilityNames returns a sorted list of enabled capabilities.
func (cfg *AgentConfig) CapabilityNames() []string {
	if cfg == nil {
		return nil
	}
	out := []string{}
	if cfg.Spec.Capabilities.Exec.Enabled {
		out = append(out, "exec")
	}
	if cfg.Spec.Capabilities.FS.Enabled {
		out = append(out, "fs")
	}
	if cfg.Spec.Capabilities.Net.Enabled {
		out = append(out, "net")
	}
	if cfg.Spec.Capabilities.Browser.Enabled {
		out = append(out, "browser")
	}
	if cfg.Spec.Capabilities.Memory.Enabled {
		out = append(out, "memory")
	}
	if cfg.Spec.Capabilities.Tools.Enabled {
		out = append(out, "tools")
	}
	if cfg.Spec.Capabilities.Secrets.Enabled {
		out = append(out, "secrets")
	}
	sort.Strings(out)
	return out
}
