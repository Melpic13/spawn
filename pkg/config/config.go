package config

import "time"

// DaemonConfig represents top-level daemon configuration.
type DaemonConfig struct {
	APIVersion    string              `mapstructure:"apiVersion" yaml:"apiVersion"`
	Kind          string              `mapstructure:"kind" yaml:"kind"`
	Server        ServerConfig        `mapstructure:"server" yaml:"server"`
	Auth          AuthConfig          `mapstructure:"auth" yaml:"auth"`
	Storage       StorageConfig       `mapstructure:"storage" yaml:"storage"`
	Sandbox       SandboxConfig       `mapstructure:"sandbox" yaml:"sandbox"`
	LLM           LLMConfig           `mapstructure:"llm" yaml:"llm"`
	Mesh          MeshConfig          `mapstructure:"mesh" yaml:"mesh"`
	Observability ObservabilityConfig `mapstructure:"observability" yaml:"observability"`
	Security      SecurityConfig      `mapstructure:"security" yaml:"security"`
	Plugins       PluginsConfig       `mapstructure:"plugins" yaml:"plugins"`
}

type ServerConfig struct {
	Host  string      `mapstructure:"host" yaml:"host"`
	Ports ServerPorts `mapstructure:"ports" yaml:"ports"`
	TLS   TLSConfig   `mapstructure:"tls" yaml:"tls"`
}

type ServerPorts struct {
	GRPC    int `mapstructure:"grpc" yaml:"grpc"`
	REST    int `mapstructure:"rest" yaml:"rest"`
	Metrics int `mapstructure:"metrics" yaml:"metrics"`
}

type TLSConfig struct {
	Enabled bool   `mapstructure:"enabled" yaml:"enabled"`
	Cert    string `mapstructure:"cert" yaml:"cert"`
	Key     string `mapstructure:"key" yaml:"key"`
}

type AuthConfig struct {
	Enabled   bool           `mapstructure:"enabled" yaml:"enabled"`
	Providers []AuthProvider `mapstructure:"providers" yaml:"providers"`
	RBAC      RBACConfig     `mapstructure:"rbac" yaml:"rbac"`
}

type AuthProvider struct {
	Type     string `mapstructure:"type" yaml:"type"`
	Issuer   string `mapstructure:"issuer" yaml:"issuer"`
	Audience string `mapstructure:"audience" yaml:"audience"`
	Header   string `mapstructure:"header" yaml:"header"`
}

type RBACConfig struct {
	Enabled     bool   `mapstructure:"enabled" yaml:"enabled"`
	DefaultRole string `mapstructure:"defaultRole" yaml:"defaultRole"`
}

type StorageConfig struct {
	State  DriverConfig `mapstructure:"state" yaml:"state"`
	Vector DriverConfig `mapstructure:"vector" yaml:"vector"`
	Files  DriverConfig `mapstructure:"files" yaml:"files"`
}

type DriverConfig struct {
	Driver string `mapstructure:"driver" yaml:"driver"`
	DSN    string `mapstructure:"dsn" yaml:"dsn"`
	Path   string `mapstructure:"path" yaml:"path"`
}

type SandboxConfig struct {
	DefaultRuntime string         `mapstructure:"defaultRuntime" yaml:"defaultRuntime"`
	GVisor         RuntimeConfig  `mapstructure:"gvisor" yaml:"gvisor"`
	Firecracker    RuntimeConfig  `mapstructure:"firecracker" yaml:"firecracker"`
	Docker         RuntimeConfig  `mapstructure:"docker" yaml:"docker"`
	Defaults       SandboxDefault `mapstructure:"defaults" yaml:"defaults"`
}

type RuntimeConfig struct {
	Binary     string `mapstructure:"binary" yaml:"binary"`
	Platform   string `mapstructure:"platform" yaml:"platform"`
	KernelPath string `mapstructure:"kernelPath" yaml:"kernelPath"`
	Socket     string `mapstructure:"socket" yaml:"socket"`
}

type SandboxDefault struct {
	Memory  string        `mapstructure:"memory" yaml:"memory"`
	CPU     string        `mapstructure:"cpu" yaml:"cpu"`
	Timeout time.Duration `mapstructure:"timeout" yaml:"timeout"`
}

type LLMConfig struct {
	Providers LLMProviders `mapstructure:"providers" yaml:"providers"`
	Routing   Routing      `mapstructure:"routing" yaml:"routing"`
	Costs     CostConfig   `mapstructure:"costs" yaml:"costs"`
}

type LLMProviders struct {
	Anthropic ProviderConfig `mapstructure:"anthropic" yaml:"anthropic"`
	OpenAI    ProviderConfig `mapstructure:"openai" yaml:"openai"`
}

type ProviderConfig struct {
	APIKey       string `mapstructure:"apiKey" yaml:"apiKey"`
	DefaultModel string `mapstructure:"defaultModel" yaml:"defaultModel"`
}

type Routing struct {
	Strategy      string   `mapstructure:"strategy" yaml:"strategy"`
	FallbackChain []string `mapstructure:"fallbackChain" yaml:"fallbackChain"`
}

type CostConfig struct {
	TrackEnabled bool        `mapstructure:"trackEnabled" yaml:"trackEnabled"`
	Alerts       []CostAlert `mapstructure:"alerts" yaml:"alerts"`
}

type CostAlert struct {
	Threshold float64 `mapstructure:"threshold" yaml:"threshold"`
	Action    string  `mapstructure:"action" yaml:"action"`
}

type MeshConfig struct {
	Enabled bool   `mapstructure:"enabled" yaml:"enabled"`
	Backend string `mapstructure:"backend" yaml:"backend"`
	NATS    struct {
		URL string `mapstructure:"url" yaml:"url"`
	} `mapstructure:"nats" yaml:"nats"`
}

type ObservabilityConfig struct {
	Traces struct {
		Enabled  bool   `mapstructure:"enabled" yaml:"enabled"`
		Exporter string `mapstructure:"exporter" yaml:"exporter"`
		Endpoint string `mapstructure:"endpoint" yaml:"endpoint"`
	} `mapstructure:"traces" yaml:"traces"`
	Metrics struct {
		Enabled  bool   `mapstructure:"enabled" yaml:"enabled"`
		Exporter string `mapstructure:"exporter" yaml:"exporter"`
		Path     string `mapstructure:"path" yaml:"path"`
	} `mapstructure:"metrics" yaml:"metrics"`
	Logs struct {
		Level  string `mapstructure:"level" yaml:"level"`
		Format string `mapstructure:"format" yaml:"format"`
		Output string `mapstructure:"output" yaml:"output"`
	} `mapstructure:"logs" yaml:"logs"`
}

type SecurityConfig struct {
	Secrets struct {
		Provider string `mapstructure:"provider" yaml:"provider"`
		Vault    struct {
			Address    string `mapstructure:"address" yaml:"address"`
			AuthMethod string `mapstructure:"authMethod" yaml:"authMethod"`
		} `mapstructure:"vault" yaml:"vault"`
	} `mapstructure:"secrets" yaml:"secrets"`
	Audit struct {
		Enabled bool   `mapstructure:"enabled" yaml:"enabled"`
		Path    string `mapstructure:"path" yaml:"path"`
	} `mapstructure:"audit" yaml:"audit"`
}

type PluginsConfig struct {
	Directory string `mapstructure:"directory" yaml:"directory"`
	Autoload  bool   `mapstructure:"autoload" yaml:"autoload"`
}
