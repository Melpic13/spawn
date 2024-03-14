# SPAWN: Agent Operating System — Master Build Prompt

## Project Identity

**Name:** spawn
**Tagline:** systemd for AI agents
**Repository:** github.com/spawndev/spawn
**Language:** Go 1.22+
**License:** Apache 2.0

---

## Executive Summary

Build a production-grade runtime that enables AI agents to safely execute code, use tools, access files, browse the web, and persist memory inside secure, observable, composable sandboxes. This is the missing operating system layer between LLMs and the real world.

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              SPAWN DAEMON                                   │
│                                                                             │
│  ┌──────────────────────────────────────────────────────────────────────┐   │
│  │                         CONTROL PLANE                                │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  │   │
│  │  │  Scheduler  │  │  Supervisor │  │   Registry  │  │   Gateway   │  │   │
│  │  │             │  │             │  │             │  │   (gRPC +   │  │   │
│  │  │ - Queue     │  │ - Health    │  │ - Agents    │  │    REST)    │  │   │
│  │  │ - Priority  │  │ - Restart   │  │ - Tools     │  │             │  │   │
│  │  │ - Affinity  │  │ - Scaling   │  │ - Schemas   │  │ - Auth      │  │   │
│  │  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘  │   │
│  └──────────────────────────────────────────────────────────────────────┘   │
│                                    │                                        │
│  ┌──────────────────────────────────▼───────────────────────────────────┐   │
│  │                           AGENT MESH                                 │   │
│  │                                                                      │   │
│  │   ┌─────────────┐      ┌─────────────┐      ┌─────────────┐         │   │
│  │   │   Agent A   │◄────►│   Agent B   │◄────►│   Agent C   │         │   │
│  │   │             │      │             │      │             │         │   │
│  │   │ - State     │      │ - State     │      │ - State     │         │   │
│  │   │ - Context   │      │ - Context   │      │ - Context   │         │   │
│  │   │ - Tools     │      │ - Tools     │      │ - Tools     │         │   │
│  │   └──────┬──────┘      └──────┬──────┘      └──────┬──────┘         │   │
│  │          │                    │                    │                │   │
│  │   ┌──────▼────────────────────▼────────────────────▼──────┐         │   │
│  │   │              MESSAGE BUS (NATS Embedded)              │         │   │
│  │   └───────────────────────────────────────────────────────┘         │   │
│  └──────────────────────────────────────────────────────────────────────┘   │
│                                    │                                        │
│  ┌──────────────────────────────────▼───────────────────────────────────┐   │
│  │                        CAPABILITY LAYER                              │   │
│  │                                                                      │   │
│  │  ┌────────────┐ ┌────────────┐ ┌────────────┐ ┌────────────┐        │   │
│  │  │    exec    │ │     fs     │ │    net     │ │   memory   │        │   │
│  │  │            │ │            │ │            │ │            │        │   │
│  │  │ - Sandbox  │ │ - Virtual  │ │ - HTTP     │ │ - Vector   │        │   │
│  │  │ - Timeout  │ │ - Overlay  │ │ - Browser  │ │ - Graph    │        │   │
│  │  │ - Resource │ │ - Snapshot │ │ - WebSocket│ │ - KV       │        │   │
│  │  └────────────┘ └────────────┘ └────────────┘ └────────────┘        │   │
│  │                                                                      │   │
│  │  ┌────────────┐ ┌────────────┐ ┌────────────┐ ┌────────────┐        │   │
│  │  │   tools    │ │  secrets   │ │    mcp     │ │  browser   │        │   │
│  │  │            │ │            │ │            │ │            │        │   │
│  │  │ - Registry │ │ - Vault    │ │ - Client   │ │ - Pool     │        │   │
│  │  │ - Schema   │ │ - Inject   │ │ - Server   │ │ - Stealth  │        │   │
│  │  │ - Invoke   │ │ - Rotate   │ │ - Bridge   │ │ - Capture  │        │   │
│  │  └────────────┘ └────────────┘ └────────────┘ └────────────┘        │   │
│  └──────────────────────────────────────────────────────────────────────┘   │
│                                    │                                        │
│  ┌──────────────────────────────────▼───────────────────────────────────┐   │
│  │                        ISOLATION LAYER                               │   │
│  │                                                                      │   │
│  │  ┌─────────────────────────────────────────────────────────────┐     │   │
│  │  │                    SANDBOX RUNTIME                          │     │   │
│  │  │                                                             │     │   │
│  │  │   Option A: gVisor (runsc) — Default, best security         │     │   │
│  │  │   Option B: Firecracker — MicroVM, strongest isolation      │     │   │
│  │  │   Option C: Docker — Development/compatibility mode         │     │   │
│  │  │   Option D: Native — No isolation (testing only)            │     │   │
│  │  │                                                             │     │   │
│  │  └─────────────────────────────────────────────────────────────┘     │   │
│  └──────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  ┌──────────────────────────────────────────────────────────────────────┐   │
│  │                       OBSERVABILITY LAYER                            │   │
│  │                                                                      │   │
│  │  ┌────────────┐ ┌────────────┐ ┌────────────┐ ┌────────────┐        │   │
│  │  │   Traces   │ │   Metrics  │ │    Logs    │ │   Events   │        │   │
│  │  │  (OTLP)    │ │(Prometheus)│ │  (Structured)│ │  (Stream)  │        │   │
│  │  └────────────┘ └────────────┘ └────────────┘ └────────────┘        │   │
│  └──────────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Directory Structure

```
spawn/
├── cmd/
│   ├── spawn/                    # Main CLI binary
│   │   └── main.go
│   ├── spawnd/                   # Daemon binary
│   │   └── main.go
│   └── spawn-sandbox/            # Sandbox helper binary
│       └── main.go
│
├── pkg/
│   ├── agent/                    # Agent lifecycle management
│   │   ├── agent.go              # Core agent struct and methods
│   │   ├── config.go             # Agent configuration parsing
│   │   ├── context.go            # Agent execution context
│   │   ├── hooks.go              # Lifecycle hooks (pre/post)
│   │   ├── state.go              # Agent state machine
│   │   └── supervisor.go         # Agent supervision tree
│   │
│   ├── capability/               # Capability implementations
│   │   ├── capability.go         # Capability interface
│   │   ├── registry.go           # Capability registry
│   │   ├── exec/                 # Code execution capability
│   │   │   ├── exec.go
│   │   │   ├── sandbox.go
│   │   │   ├── languages.go      # Language runtime configs
│   │   │   └── limits.go         # Resource limits
│   │   ├── fs/                   # File system capability
│   │   │   ├── fs.go
│   │   │   ├── virtual.go        # Virtual FS implementation
│   │   │   ├── overlay.go        # Overlay filesystem
│   │   │   └── snapshot.go       # FS snapshots
│   │   ├── net/                  # Network capability
│   │   │   ├── net.go
│   │   │   ├── http.go           # HTTP client
│   │   │   ├── dns.go            # DNS resolution
│   │   │   └── firewall.go       # Network policies
│   │   ├── memory/               # Memory/persistence capability
│   │   │   ├── memory.go
│   │   │   ├── vector.go         # Vector store (embedded)
│   │   │   ├── graph.go          # Graph store
│   │   │   └── kv.go             # Key-value store
│   │   ├── browser/              # Browser automation capability
│   │   │   ├── browser.go
│   │   │   ├── pool.go           # Browser instance pool
│   │   │   ├── stealth.go        # Anti-detection
│   │   │   └── capture.go        # Screenshots, recordings
│   │   ├── tools/                # Tool invocation capability
│   │   │   ├── tools.go
│   │   │   ├── schema.go         # JSON Schema validation
│   │   │   └── invoke.go         # Tool execution
│   │   ├── mcp/                  # MCP protocol support
│   │   │   ├── client.go         # MCP client
│   │   │   ├── server.go         # MCP server (expose agent as tool)
│   │   │   └── bridge.go         # MCP-to-native bridge
│   │   └── secrets/              # Secrets management
│   │       ├── secrets.go
│   │       ├── vault.go          # HashiCorp Vault integration
│   │       └── inject.go         # Secret injection
│   │
│   ├── mesh/                     # Agent mesh / multi-agent
│   │   ├── mesh.go               # Mesh coordinator
│   │   ├── channel.go            # Agent-to-agent channels
│   │   ├── discovery.go          # Agent discovery
│   │   ├── routing.go            # Message routing
│   │   └── consensus.go          # Distributed consensus
│   │
│   ├── scheduler/                # Task scheduling
│   │   ├── scheduler.go          # Core scheduler
│   │   ├── queue.go              # Priority queue
│   │   ├── affinity.go           # Agent affinity rules
│   │   └── backpressure.go       # Load management
│   │
│   ├── sandbox/                  # Sandbox implementations
│   │   ├── sandbox.go            # Sandbox interface
│   │   ├── gvisor.go             # gVisor implementation
│   │   ├── firecracker.go        # Firecracker microVM
│   │   ├── docker.go             # Docker fallback
│   │   ├── native.go             # No isolation (dev only)
│   │   └── config.go             # Sandbox configuration
│   │
│   ├── llm/                      # LLM integrations
│   │   ├── provider.go           # Provider interface
│   │   ├── anthropic.go          # Claude integration
│   │   ├── openai.go             # OpenAI integration
│   │   ├── router.go             # Multi-provider routing
│   │   └── cost.go               # Cost tracking
│   │
│   ├── gateway/                  # API gateway
│   │   ├── gateway.go            # Gateway server
│   │   ├── grpc/                 # gRPC API
│   │   │   ├── server.go
│   │   │   └── handlers.go
│   │   ├── rest/                 # REST API
│   │   │   ├── server.go
│   │   │   ├── handlers.go
│   │   │   └── middleware.go
│   │   ├── websocket/            # WebSocket for streaming
│   │   │   ├── server.go
│   │   │   └── handlers.go
│   │   └── auth/                 # Authentication
│   │       ├── auth.go
│   │       ├── jwt.go
│   │       ├── apikey.go
│   │       └── rbac.go           # Role-based access control
│   │
│   ├── observability/            # Observability
│   │   ├── traces.go             # OpenTelemetry traces
│   │   ├── metrics.go            # Prometheus metrics
│   │   ├── logs.go               # Structured logging
│   │   ├── events.go             # Event streaming
│   │   └── replay.go             # Decision replay system
│   │
│   ├── config/                   # Configuration
│   │   ├── config.go             # Global config
│   │   ├── loader.go             # Config loading
│   │   ├── validate.go           # Config validation
│   │   └── watch.go              # Config hot-reload
│   │
│   └── version/                  # Version info
│       └── version.go
│
├── internal/                     # Internal packages
│   ├── proto/                    # Protobuf definitions
│   │   ├── agent.proto
│   │   ├── capability.proto
│   │   ├── mesh.proto
│   │   └── gateway.proto
│   ├── testutil/                 # Testing utilities
│   │   ├── fixtures.go
│   │   ├── mocks.go
│   │   └── helpers.go
│   └── buildinfo/                # Build information
│       └── buildinfo.go
│
├── api/                          # API specifications
│   ├── openapi/
│   │   └── spawn.yaml            # OpenAPI 3.1 spec
│   └── proto/
│       └── spawn/
│           └── v1/               # gRPC service definitions
│               ├── agent.proto
│               ├── capability.proto
│               └── gateway.proto
│
├── web/                          # Web UI (optional)
│   ├── dashboard/                # React dashboard
│   └── docs/                     # Documentation site
│
├── deploy/                       # Deployment configs
│   ├── docker/
│   │   ├── Dockerfile
│   │   ├── Dockerfile.sandbox
│   │   └── docker-compose.yml
│   ├── kubernetes/
│   │   ├── helm/
│   │   │   └── spawn/
│   │   └── kustomize/
│   └── terraform/
│       ├── aws/
│       ├── gcp/
│       └── azure/
│
├── scripts/                      # Build and utility scripts
│   ├── build.sh
│   ├── install.sh
│   ├── release.sh
│   └── generate.sh               # Code generation
│
├── configs/                      # Example configurations
│   ├── spawn.yaml                # Default daemon config
│   ├── agents/                   # Example agent configs
│   │   ├── researcher.yaml
│   │   ├── coder.yaml
│   │   └── analyst.yaml
│   └── capabilities/             # Capability configs
│
├── docs/                         # Documentation
│   ├── architecture.md
│   ├── quickstart.md
│   ├── configuration.md
│   ├── security.md
│   ├── capabilities/
│   ├── deployment/
│   └── api/
│
├── examples/                     # Example implementations
│   ├── hello-agent/
│   ├── multi-agent-research/
│   ├── code-assistant/
│   └── web-scraper/
│
├── test/                         # Integration tests
│   ├── e2e/
│   ├── benchmark/
│   └── security/
│
├── .github/
│   ├── workflows/
│   │   ├── ci.yml
│   │   ├── release.yml
│   │   └── security.yml
│   ├── ISSUE_TEMPLATE/
│   ├── PULL_REQUEST_TEMPLATE.md
│   └── CODEOWNERS
│
├── .goreleaser.yml               # Release automation
├── Makefile                      # Build commands
├── go.mod
├── go.sum
├── LICENSE
├── README.md
├── CONTRIBUTING.md
├── SECURITY.md
└── CHANGELOG.md
```

---

## Core Components Specification

### 1. Agent Configuration Schema

```yaml
# agent.yaml - Full specification
apiVersion: spawn.dev/v1
kind: Agent

metadata:
  name: researcher
  namespace: default
  labels:
    team: ai-research
    tier: production
  annotations:
    spawn.dev/description: "Research agent for web analysis"

spec:
  # LLM Configuration
  model:
    provider: anthropic           # anthropic, openai, custom
    name: claude-sonnet-4-20250514
    temperature: 0.7
    maxTokens: 8192
    fallback:
      - provider: openai
        name: gpt-4o

  # System prompt and goal
  system: |
    You are a research assistant with access to web browsing,
    code execution, and file management capabilities.
    
  goal: |
    Research the given topic thoroughly and produce a
    comprehensive report saved to the output directory.

  # Capabilities granted to this agent
  capabilities:
    exec:
      enabled: true
      languages: [python, nodejs, bash]
      timeout: 300s
      memory: 512Mi
      cpu: "1.0"
    
    fs:
      enabled: true
      mounts:
        - path: /workspace
          mode: rw
          quota: 1Gi
        - path: /data
          source: s3://bucket/data
          mode: ro
      
    net:
      enabled: true
      allowlist:
        - "*.wikipedia.org"
        - "*.github.com"
        - "api.anthropic.com"
      denylist:
        - "*.malware.com"
      rateLimit:
        requests: 100
        per: 1m
    
    browser:
      enabled: true
      headless: true
      stealth: true
      timeout: 60s
      viewport:
        width: 1920
        height: 1080
    
    memory:
      enabled: true
      vector:
        dimensions: 1536
        metric: cosine
      graph:
        enabled: true
      ttl: 24h
    
    tools:
      enabled: true
      builtin:
        - calculator
        - datetime
        - json_parser
      mcp:
        - uri: "http://localhost:3000/mcp"
          name: custom-tools
      custom:
        - name: analyze_sentiment
          description: Analyze sentiment of text
          schema:
            type: object
            properties:
              text:
                type: string
            required: [text]
          handler: /plugins/sentiment.wasm
    
    secrets:
      enabled: true
      inject:
        - name: API_KEY
          source: vault://secret/api-key
        - name: DB_PASSWORD
          source: env://DB_PASSWORD

  # Resource limits for the agent
  resources:
    requests:
      memory: 256Mi
      cpu: "0.5"
    limits:
      memory: 1Gi
      cpu: "2.0"
    costLimit:
      daily: 10.00
      monthly: 100.00
      currency: USD

  # Sandbox configuration
  sandbox:
    runtime: gvisor              # gvisor, firecracker, docker, native
    networkPolicy: restricted    # restricted, egress-only, full
    seccompProfile: strict       # strict, moderate, permissive

  # Lifecycle hooks
  hooks:
    preStart:
      - command: ["pip", "install", "-r", "requirements.txt"]
    postStop:
      - command: ["cleanup.sh"]
    healthCheck:
      interval: 30s
      timeout: 5s
      command: ["health.sh"]

  # Observability
  observability:
    traces:
      enabled: true
      sampleRate: 1.0
    metrics:
      enabled: true
    logs:
      level: info
      format: json
    events:
      stream: true

  # Scaling (for multi-instance)
  scaling:
    minReplicas: 1
    maxReplicas: 10
    metrics:
      - type: queue-depth
        target: 5

  # Inter-agent communication
  mesh:
    channels:
      - name: findings
        type: pubsub
        topic: research.findings
      - name: requests
        type: request-reply
        timeout: 30s
```

### 2. CLI Commands

```bash
# Core commands
spawn init [name]              # Initialize new agent project
spawn run [config]             # Run agent(s) from config
spawn start                    # Start spawn daemon
spawn stop                     # Stop spawn daemon
spawn status                   # Show daemon and agent status

# Agent management
spawn agent list               # List all agents
spawn agent get <name>         # Get agent details
spawn agent logs <name>        # Stream agent logs
spawn agent exec <name> <cmd>  # Execute command in agent
spawn agent kill <name>        # Terminate agent
spawn agent restart <name>     # Restart agent

# Capability management
spawn capability list          # List available capabilities
spawn capability install <n>   # Install capability plugin
spawn capability config <n>    # Configure capability

# Tool management
spawn tool list                # List registered tools
spawn tool register <schema>   # Register new tool
spawn tool invoke <name>       # Manually invoke tool

# Mesh commands
spawn mesh status              # Show mesh topology
spawn mesh channels            # List communication channels
spawn mesh send <channel>      # Send message to channel

# Observability
spawn logs                     # Stream all logs
spawn metrics                  # Show metrics
spawn trace <id>               # Get trace details
spawn replay <id>              # Replay agent decision

# Development
spawn dev                      # Start development mode
spawn validate <config>        # Validate configuration
spawn lint                     # Lint agent configs
spawn test                     # Run agent tests

# System
spawn version                  # Show version
spawn doctor                   # Diagnose installation
spawn upgrade                  # Upgrade spawn
spawn config                   # Manage global config
```

### 3. Core Interfaces

```go
// pkg/agent/agent.go
package agent

import (
    "context"
    "time"
    
    "spawn.dev/pkg/capability"
    "spawn.dev/pkg/llm"
)

// AgentState represents the current state of an agent
type AgentState string

const (
    StateInitializing AgentState = "initializing"
    StateRunning      AgentState = "running"
    StatePaused       AgentState = "paused"
    StateCompleted    AgentState = "completed"
    StateFailed       AgentState = "failed"
    StateTerminated   AgentState = "terminated"
)

// Agent represents a running AI agent instance
type Agent struct {
    ID          string
    Name        string
    Namespace   string
    Config      *AgentConfig
    State       AgentState
    StartedAt   time.Time
    
    // Runtime components
    LLM         llm.Provider
    Capabilities map[string]capability.Capability
    Context     *ExecutionContext
    
    // Communication
    Inbox       chan Message
    Outbox      chan Message
    
    // Metrics
    TokensUsed  int64
    CostUSD     float64
    TasksRun    int64
}

// AgentConfig represents the parsed agent configuration
type AgentConfig struct {
    APIVersion  string            `yaml:"apiVersion"`
    Kind        string            `yaml:"kind"`
    Metadata    Metadata          `yaml:"metadata"`
    Spec        AgentSpec         `yaml:"spec"`
}

type Metadata struct {
    Name        string            `yaml:"name"`
    Namespace   string            `yaml:"namespace"`
    Labels      map[string]string `yaml:"labels"`
    Annotations map[string]string `yaml:"annotations"`
}

type AgentSpec struct {
    Model         ModelConfig                    `yaml:"model"`
    System        string                         `yaml:"system"`
    Goal          string                         `yaml:"goal"`
    Capabilities  CapabilitiesConfig             `yaml:"capabilities"`
    Resources     ResourceConfig                 `yaml:"resources"`
    Sandbox       SandboxConfig                  `yaml:"sandbox"`
    Hooks         HooksConfig                    `yaml:"hooks"`
    Observability ObservabilityConfig            `yaml:"observability"`
    Scaling       ScalingConfig                  `yaml:"scaling"`
    Mesh          MeshConfig                     `yaml:"mesh"`
}

// Manager handles agent lifecycle
type Manager interface {
    // Lifecycle
    Create(ctx context.Context, config *AgentConfig) (*Agent, error)
    Start(ctx context.Context, id string) error
    Stop(ctx context.Context, id string) error
    Restart(ctx context.Context, id string) error
    Delete(ctx context.Context, id string) error
    
    // Query
    Get(ctx context.Context, id string) (*Agent, error)
    List(ctx context.Context, opts ListOptions) ([]*Agent, error)
    
    // Interaction
    SendMessage(ctx context.Context, id string, msg Message) error
    Execute(ctx context.Context, id string, task Task) (*TaskResult, error)
    
    // Observation
    Logs(ctx context.Context, id string, opts LogOptions) (<-chan LogEntry, error)
    Metrics(ctx context.Context, id string) (*AgentMetrics, error)
    
    // Events
    Watch(ctx context.Context, opts WatchOptions) (<-chan Event, error)
}

// ExecutionContext holds the runtime context for an agent
type ExecutionContext struct {
    // Working directory
    WorkDir     string
    
    // Environment
    Env         map[string]string
    
    // Secrets (injected)
    Secrets     map[string]string
    
    // Conversation history
    Messages    []llm.Message
    
    // Memory stores
    VectorStore capability.VectorStore
    GraphStore  capability.GraphStore
    KVStore     capability.KVStore
    
    // Tool results cache
    ToolCache   map[string]interface{}
    
    // Parent context for cancellation
    ctx         context.Context
    cancel      context.CancelFunc
}
```

```go
// pkg/capability/capability.go
package capability

import (
    "context"
)

// Capability represents a capability that can be granted to an agent
type Capability interface {
    // Metadata
    Name() string
    Version() string
    Description() string
    
    // Lifecycle
    Initialize(ctx context.Context, config map[string]interface{}) error
    Shutdown(ctx context.Context) error
    
    // Health
    HealthCheck(ctx context.Context) error
    
    // Schema
    Schema() *Schema
    
    // Execution
    Execute(ctx context.Context, request *Request) (*Response, error)
}

// Schema defines the capability's interface
type Schema struct {
    Actions     []Action           `json:"actions"`
    Events      []EventType        `json:"events"`
    Config      map[string]Field   `json:"config"`
}

type Action struct {
    Name        string             `json:"name"`
    Description string             `json:"description"`
    Input       map[string]Field   `json:"input"`
    Output      map[string]Field   `json:"output"`
}

type Field struct {
    Type        string             `json:"type"`
    Description string             `json:"description"`
    Required    bool               `json:"required"`
    Default     interface{}        `json:"default,omitempty"`
}

// Request represents a capability execution request
type Request struct {
    Action      string                 `json:"action"`
    Params      map[string]interface{} `json:"params"`
    Context     *ExecutionContext      `json:"context"`
    Timeout     time.Duration          `json:"timeout"`
}

// Response represents a capability execution response
type Response struct {
    Success     bool                   `json:"success"`
    Data        interface{}            `json:"data,omitempty"`
    Error       *Error                 `json:"error,omitempty"`
    Metrics     *ExecutionMetrics      `json:"metrics,omitempty"`
}

// Registry manages capability registration and discovery
type Registry interface {
    Register(cap Capability) error
    Unregister(name string) error
    Get(name string) (Capability, error)
    List() []Capability
    
    // Discovery
    Discover(ctx context.Context) ([]Capability, error)
}
```

```go
// pkg/sandbox/sandbox.go
package sandbox

import (
    "context"
    "io"
)

// Runtime represents a sandbox runtime implementation
type Runtime interface {
    // Lifecycle
    Create(ctx context.Context, config *Config) (Sandbox, error)
    List(ctx context.Context) ([]Sandbox, error)
    
    // Capabilities
    Supports(feature Feature) bool
    
    // Health
    HealthCheck(ctx context.Context) error
}

// Sandbox represents an isolated execution environment
type Sandbox interface {
    // Identity
    ID() string
    
    // Lifecycle
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Pause(ctx context.Context) error
    Resume(ctx context.Context) error
    Destroy(ctx context.Context) error
    
    // Execution
    Exec(ctx context.Context, cmd *Command) (*ExecResult, error)
    
    // Filesystem
    CopyIn(ctx context.Context, src string, dst string) error
    CopyOut(ctx context.Context, src string, dst string) error
    
    // Networking
    NetworkConfig() *NetworkConfig
    
    // State
    State() SandboxState
    Metrics() *SandboxMetrics
    
    // Streams
    Stdout() io.ReadCloser
    Stderr() io.ReadCloser
    Stdin() io.WriteCloser
}

// Config represents sandbox configuration
type Config struct {
    Runtime     RuntimeType       `yaml:"runtime"`
    Image       string            `yaml:"image"`
    
    // Resource limits
    Memory      int64             `yaml:"memory"`      // bytes
    CPU         float64           `yaml:"cpu"`         // cores
    Disk        int64             `yaml:"disk"`        // bytes
    Pids        int               `yaml:"pids"`        // max processes
    
    // Network
    Network     NetworkPolicy     `yaml:"network"`
    
    // Security
    Seccomp     SeccompProfile    `yaml:"seccomp"`
    Capabilities []string         `yaml:"capabilities"`
    ReadOnlyRoot bool             `yaml:"readOnlyRoot"`
    
    // Mounts
    Mounts      []Mount           `yaml:"mounts"`
    
    // Environment
    Env         map[string]string `yaml:"env"`
    
    // Timeouts
    StartTimeout time.Duration    `yaml:"startTimeout"`
    ExecTimeout  time.Duration    `yaml:"execTimeout"`
}

type RuntimeType string

const (
    RuntimeGVisor      RuntimeType = "gvisor"
    RuntimeFirecracker RuntimeType = "firecracker"
    RuntimeDocker      RuntimeType = "docker"
    RuntimeNative      RuntimeType = "native"
)

type NetworkPolicy string

const (
    NetworkNone       NetworkPolicy = "none"
    NetworkRestricted NetworkPolicy = "restricted"
    NetworkEgressOnly NetworkPolicy = "egress-only"
    NetworkFull       NetworkPolicy = "full"
)
```

```go
// pkg/mesh/mesh.go
package mesh

import (
    "context"
    "time"
)

// Mesh coordinates multi-agent communication
type Mesh interface {
    // Agent registration
    Register(ctx context.Context, agent *AgentInfo) error
    Deregister(ctx context.Context, agentID string) error
    
    // Discovery
    Discover(ctx context.Context, query *DiscoveryQuery) ([]*AgentInfo, error)
    
    // Messaging
    Send(ctx context.Context, msg *Message) error
    Request(ctx context.Context, msg *Message, timeout time.Duration) (*Message, error)
    Subscribe(ctx context.Context, topic string, handler MessageHandler) (Subscription, error)
    
    // Channels
    CreateChannel(ctx context.Context, config *ChannelConfig) (Channel, error)
    GetChannel(ctx context.Context, name string) (Channel, error)
    
    // Topology
    Topology(ctx context.Context) (*TopologyGraph, error)
}

// Channel represents a communication channel between agents
type Channel interface {
    Name() string
    Type() ChannelType
    
    Send(ctx context.Context, msg *Message) error
    Receive(ctx context.Context) (*Message, error)
    
    Subscribe(handler MessageHandler) (Subscription, error)
    
    Close() error
}

type ChannelType string

const (
    ChannelPubSub      ChannelType = "pubsub"
    ChannelRequestReply ChannelType = "request-reply"
    ChannelStream      ChannelType = "stream"
    ChannelBroadcast   ChannelType = "broadcast"
)

// Message represents an inter-agent message
type Message struct {
    ID          string                 `json:"id"`
    From        string                 `json:"from"`
    To          string                 `json:"to,omitempty"`
    Topic       string                 `json:"topic,omitempty"`
    Type        MessageType            `json:"type"`
    Payload     interface{}            `json:"payload"`
    Metadata    map[string]string      `json:"metadata,omitempty"`
    Timestamp   time.Time              `json:"timestamp"`
    ReplyTo     string                 `json:"replyTo,omitempty"`
    CorrelationID string               `json:"correlationId,omitempty"`
}

// TopologyGraph represents the mesh topology
type TopologyGraph struct {
    Agents      []*AgentNode           `json:"agents"`
    Channels    []*ChannelEdge         `json:"channels"`
    Connections []*Connection          `json:"connections"`
}
```

```go
// pkg/llm/provider.go
package llm

import (
    "context"
)

// Provider represents an LLM provider
type Provider interface {
    // Identity
    Name() string
    Models() []string
    
    // Chat
    Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
    ChatStream(ctx context.Context, req *ChatRequest) (<-chan *StreamChunk, error)
    
    // Tools
    ChatWithTools(ctx context.Context, req *ChatRequest, tools []Tool) (*ChatResponse, error)
    
    // Embeddings
    Embed(ctx context.Context, input []string) ([][]float32, error)
    
    // Cost
    EstimateCost(req *ChatRequest) float64
    
    // Health
    HealthCheck(ctx context.Context) error
}

// ChatRequest represents a chat completion request
type ChatRequest struct {
    Model       string                 `json:"model"`
    Messages    []Message              `json:"messages"`
    System      string                 `json:"system,omitempty"`
    Temperature float64                `json:"temperature,omitempty"`
    MaxTokens   int                    `json:"max_tokens,omitempty"`
    StopSequences []string             `json:"stop_sequences,omitempty"`
    Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ChatResponse represents a chat completion response
type ChatResponse struct {
    ID          string                 `json:"id"`
    Model       string                 `json:"model"`
    Content     string                 `json:"content"`
    ToolCalls   []ToolCall             `json:"tool_calls,omitempty"`
    StopReason  StopReason             `json:"stop_reason"`
    Usage       *Usage                 `json:"usage"`
}

// Tool represents a tool available to the LLM
type Tool struct {
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    InputSchema map[string]interface{} `json:"input_schema"`
}

// ToolCall represents a tool invocation by the LLM
type ToolCall struct {
    ID          string                 `json:"id"`
    Name        string                 `json:"name"`
    Input       map[string]interface{} `json:"input"`
}

// Router routes requests to appropriate providers
type Router interface {
    Route(ctx context.Context, req *ChatRequest) (Provider, error)
    AddProvider(provider Provider) error
    RemoveProvider(name string) error
    SetStrategy(strategy RoutingStrategy)
}

type RoutingStrategy string

const (
    StrategyRoundRobin   RoutingStrategy = "round-robin"
    StrategyCostOptimize RoutingStrategy = "cost-optimize"
    StrategyLatencyOptimize RoutingStrategy = "latency-optimize"
    StrategyComplexity   RoutingStrategy = "complexity"
    StrategyFallback     RoutingStrategy = "fallback"
)
```

### 4. Daemon Configuration

```yaml
# spawn.yaml - Daemon configuration
apiVersion: spawn.dev/v1
kind: DaemonConfig

server:
  host: 0.0.0.0
  ports:
    grpc: 9090
    rest: 8080
    metrics: 9091
  tls:
    enabled: true
    cert: /etc/spawn/tls/server.crt
    key: /etc/spawn/tls/server.key

auth:
  enabled: true
  providers:
    - type: jwt
      issuer: https://auth.example.com
      audience: spawn
    - type: apikey
      header: X-API-Key
  rbac:
    enabled: true
    defaultRole: viewer

storage:
  # Agent state storage
  state:
    driver: sqlite          # sqlite, postgres, mysql
    dsn: /var/lib/spawn/state.db
  
  # Vector store for memory capability
  vector:
    driver: embedded        # embedded, pinecone, weaviate
    path: /var/lib/spawn/vectors
  
  # File storage
  files:
    driver: local           # local, s3, gcs
    path: /var/lib/spawn/files

sandbox:
  defaultRuntime: gvisor
  gvisor:
    binary: /usr/local/bin/runsc
    platform: systrap
  firecracker:
    binary: /usr/local/bin/firecracker
    kernelPath: /var/lib/spawn/vmlinux
  docker:
    socket: /var/run/docker.sock
  
  # Resource defaults
  defaults:
    memory: 256Mi
    cpu: "0.5"
    timeout: 5m

llm:
  providers:
    anthropic:
      apiKey: ${ANTHROPIC_API_KEY}
      defaultModel: claude-sonnet-4-20250514
    openai:
      apiKey: ${OPENAI_API_KEY}
      defaultModel: gpt-4o
  
  routing:
    strategy: complexity
    fallbackChain:
      - anthropic
      - openai
  
  costs:
    trackEnabled: true
    alerts:
      - threshold: 100.00
        action: notify
      - threshold: 500.00
        action: pause

mesh:
  enabled: true
  backend: embedded-nats    # embedded-nats, nats, redis
  nats:
    url: nats://localhost:4222

observability:
  traces:
    enabled: true
    exporter: otlp
    endpoint: localhost:4317
  
  metrics:
    enabled: true
    exporter: prometheus
    path: /metrics
  
  logs:
    level: info
    format: json
    output: stdout

security:
  secrets:
    provider: vault         # vault, env, file
    vault:
      address: https://vault.example.com
      authMethod: kubernetes
  
  audit:
    enabled: true
    path: /var/log/spawn/audit.log

plugins:
  directory: /var/lib/spawn/plugins
  autoload: true
```

---

## Implementation Requirements

### Phase 1: Core Foundation (Week 1-2)

1. **Project Setup**
   - Initialize Go module with proper structure
   - Set up Makefile with build, test, lint targets
   - Configure GitHub Actions CI/CD
   - Set up pre-commit hooks (golangci-lint, gofmt)

2. **CLI Framework**
   - Implement using `cobra` + `viper`
   - All core commands stubbed
   - Config loading and validation
   - Colored output with `lipgloss`

3. **Agent Config Parser**
   - YAML parsing with validation
   - JSON Schema generation for IDE support
   - Config inheritance/merging

4. **Basic Sandbox (Docker)**
   - Docker-based sandbox implementation
   - Container lifecycle management
   - Basic exec functionality

### Phase 2: Capabilities (Week 3-4)

5. **Exec Capability**
   - Python, Node.js, Bash support
   - Resource limits (memory, CPU, time)
   - Output capture and streaming

6. **Filesystem Capability**
   - Virtual filesystem with overlayfs
   - Mount management
   - Snapshot/restore

7. **Network Capability**
   - HTTP client with policies
   - DNS resolution
   - Allow/deny lists

8. **Memory Capability**
   - Embedded vector store (use `hnswlib`)
   - Key-value store (bbolt)
   - Graph store (embedded dgraph or custom)

### Phase 3: LLM Integration (Week 5-6)

9. **Provider Implementations**
   - Anthropic Claude integration
   - OpenAI integration
   - Streaming support

10. **Tool System**
    - Tool registry
    - Schema validation
    - Execution pipeline

11. **Agent Loop**
    - ReAct-style agent loop
    - Conversation management
    - Tool use handling

### Phase 4: Advanced Sandbox (Week 7-8)

12. **gVisor Integration**
    - runsc wrapper
    - OCI runtime compliance
    - Security policies

13. **Firecracker Integration** (Optional)
    - MicroVM management
    - Kernel/rootfs setup
    - Network virtualization

### Phase 5: Mesh & Multi-Agent (Week 9-10)

14. **Message Bus**
    - Embedded NATS
    - Channel abstractions
    - Request/reply patterns

15. **Agent Discovery**
    - Service registry
    - Health checking
    - Load balancing

16. **Coordination**
    - Distributed consensus (Raft)
    - Leader election
    - State synchronization

### Phase 6: Observability & API (Week 11-12)

17. **Tracing**
    - OpenTelemetry integration
    - Span propagation
    - Decision replay

18. **Metrics**
    - Prometheus metrics
    - Cost tracking
    - Performance metrics

19. **Gateway API**
    - gRPC service implementation
    - REST API with OpenAPI
    - WebSocket streaming

### Phase 7: Polish & Security (Week 13-14)

20. **Security Hardening**
    - Seccomp profiles
    - Capability dropping
    - Audit logging

21. **Browser Capability**
    - Chromium pool management
    - Stealth mode
    - Screenshot/recording

22. **MCP Integration**
    - MCP client
    - MCP server (expose agent)
    - Tool bridging

23. **Documentation**
    - Comprehensive docs
    - API reference
    - Tutorials

24. **Testing**
    - Unit tests (>80% coverage)
    - Integration tests
    - Security tests
    - Benchmarks

---

## Code Quality Requirements

### Style
- Follow Uber Go Style Guide
- All exported types/functions documented
- Consistent error handling with `fmt.Errorf("operation: %w", err)`
- Context propagation throughout

### Testing
- Table-driven tests
- Mocks using `gomock` or `testify/mock`
- Integration tests with testcontainers
- Fuzzing for parsers

### Performance
- Benchmark critical paths
- Profile memory allocations
- Connection pooling
- Efficient serialization (protobuf for internal)

### Security
- No hardcoded secrets
- Input validation everywhere
- Secure defaults
- Regular dependency updates

---

## Dependencies (go.mod)

```go
module spawn.dev

go 1.22

require (
    // CLI
    github.com/spf13/cobra v1.8.0
    github.com/spf13/viper v1.18.0
    github.com/charmbracelet/lipgloss v0.9.1
    github.com/charmbracelet/bubbletea v0.25.0
    
    // Server
    google.golang.org/grpc v1.61.0
    google.golang.org/protobuf v1.32.0
    github.com/grpc-ecosystem/grpc-gateway/v2 v2.19.0
    github.com/labstack/echo/v4 v4.11.4
    github.com/gorilla/websocket v1.5.1
    
    // Storage
    github.com/dgraph-io/badger/v4 v4.2.0
    go.etcd.io/bbolt v1.3.8
    github.com/lib/pq v1.10.9
    
    // Containers
    github.com/docker/docker v25.0.0
    github.com/containerd/containerd v1.7.12
    github.com/opencontainers/runtime-spec v1.2.0
    
    // Messaging
    github.com/nats-io/nats.go v1.32.0
    github.com/nats-io/nats-server/v2 v2.10.9
    
    // LLM
    github.com/anthropics/anthropic-sdk-go v0.1.0
    github.com/sashabaranov/go-openai v1.19.0
    
    // Observability
    go.opentelemetry.io/otel v1.23.0
    go.opentelemetry.io/otel/trace v1.23.0
    go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.23.0
    github.com/prometheus/client_golang v1.18.0
    go.uber.org/zap v1.26.0
    
    // Security
    github.com/golang-jwt/jwt/v5 v5.2.0
    golang.org/x/crypto v0.18.0
    
    // Utilities
    github.com/google/uuid v1.6.0
    github.com/hashicorp/go-multierror v1.1.1
    golang.org/x/sync v0.6.0
    gopkg.in/yaml.v3 v3.0.1
    github.com/xeipuuv/gojsonschema v1.2.0
)
```

---

## Build Commands

```makefile
.PHONY: all build test lint clean

VERSION := $(shell git describe --tags --always --dirty)
LDFLAGS := -ldflags "-X spawn.dev/internal/buildinfo.Version=$(VERSION)"

all: lint test build

build:
	go build $(LDFLAGS) -o bin/spawn ./cmd/spawn
	go build $(LDFLAGS) -o bin/spawnd ./cmd/spawnd
	go build $(LDFLAGS) -o bin/spawn-sandbox ./cmd/spawn-sandbox

test:
	go test -race -cover ./...

test-integration:
	go test -race -tags=integration ./test/...

lint:
	golangci-lint run

generate:
	go generate ./...
	buf generate

clean:
	rm -rf bin/

install:
	go install $(LDFLAGS) ./cmd/spawn

docker:
	docker build -t spawn:$(VERSION) .

release:
	goreleaser release --clean
```

---

## Success Criteria

The project is complete when:

1. ✅ `spawn run agent.yaml` successfully runs an agent with all capabilities
2. ✅ Multiple agents can communicate via mesh
3. ✅ gVisor sandbox provides secure isolation
4. ✅ Full observability (traces, metrics, logs)
5. ✅ REST and gRPC APIs functional
6. ✅ 80%+ test coverage
7. ✅ All examples in `/examples` work
8. ✅ Documentation complete
9. ✅ CI/CD pipeline green
10. ✅ Security audit passed

---

## Additional Context

This project aims to become the standard runtime for AI agents, similar to how Docker became the standard for containers. Key differentiators:

1. **Security-first**: Unlike other agent frameworks, we prioritize sandboxing
2. **Multi-agent native**: Built for swarms, not just single agents
3. **Observability**: Full tracing of agent decisions for debugging/replay
4. **Cloud-native**: Kubernetes-ready from day one
5. **Protocol-agnostic**: Works with any LLM provider

The target users are:
- AI/ML engineers building agent systems
- Platform teams deploying agents in production
- Researchers experimenting with multi-agent coordination
- Enterprises needing secure, auditable AI systems
