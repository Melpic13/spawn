# Agent Specification Reference

This document provides a complete reference for the agent YAML specification.

## Overview

Agent configurations are defined in YAML files that specify the agent's identity, model, capabilities, resources, and behavior.

```yaml
apiVersion: spawn.dev/v1
kind: Agent
metadata:
  # Agent identity
spec:
  # Agent specification
```

## Complete Schema

### Top-Level Structure

```yaml
apiVersion: spawn.dev/v1          # Required: API version
kind: Agent                        # Required: Resource kind
metadata:                          # Required: Agent metadata
  name: string                     # Required: Agent name
  namespace: string                # Optional: Namespace (default: "default")
  labels: map[string]string        # Optional: Labels for selection
  annotations: map[string]string   # Optional: Annotations
spec:                              # Required: Agent specification
  model: ModelConfig               # Required: LLM configuration
  system: string                   # Optional: System prompt
  goal: string                     # Optional: Agent goal
  capabilities: CapabilitiesConfig # Optional: Enabled capabilities
  resources: ResourceConfig        # Optional: Resource limits
  sandbox: SandboxConfig           # Optional: Sandbox settings
  hooks: HooksConfig               # Optional: Lifecycle hooks
  observability: ObservabilityConfig # Optional: Telemetry settings
  scaling: ScalingConfig           # Optional: Scaling rules
  mesh: MeshConfig                 # Optional: Multi-agent mesh
```

---

## Metadata

### `metadata.name`

**Type:** `string`  
**Required:** Yes  
**Pattern:** `^[a-z0-9][a-z0-9-]*[a-z0-9]$`  
**Max Length:** 63 characters

The unique name of the agent within its namespace.

```yaml
metadata:
  name: research-assistant
```

### `metadata.namespace`

**Type:** `string`  
**Required:** No  
**Default:** `"default"`

The namespace for the agent. Namespaces provide isolation and organization.

```yaml
metadata:
  namespace: production
```

### `metadata.labels`

**Type:** `map[string]string`  
**Required:** No

Key-value pairs for organizing and selecting agents.

```yaml
metadata:
  labels:
    team: ai-research
    tier: production
    cost-center: eng-42
```

### `metadata.annotations`

**Type:** `map[string]string`  
**Required:** No

Arbitrary metadata for tooling and documentation.

```yaml
metadata:
  annotations:
    spawn.dev/description: "Research agent for market analysis"
    spawn.dev/owner: "alice@company.com"
    spawn.dev/docs: "https://wiki.company.com/agents/researcher"
```

---

## Model Configuration

### `spec.model`

Configures the LLM provider and model settings.

```yaml
spec:
  model:
    provider: anthropic           # Required: Provider name
    name: claude-sonnet-4-20250514 # Required: Model name
    temperature: 0.7              # Optional: Sampling temperature
    maxTokens: 8192               # Optional: Maximum output tokens
    topP: 0.9                     # Optional: Nucleus sampling
    topK: 40                      # Optional: Top-K sampling
    stopSequences:                # Optional: Stop sequences
      - "\n\nHuman:"
    fallback:                     # Optional: Fallback providers
      - provider: openai
        name: gpt-4o
```

### Model Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `provider` | string | Yes | - | Provider: `anthropic`, `openai`, `custom` |
| `name` | string | Yes | - | Model identifier |
| `temperature` | float | No | 0.7 | Sampling temperature (0.0-2.0) |
| `maxTokens` | int | No | 4096 | Maximum output tokens |
| `topP` | float | No | 1.0 | Nucleus sampling (0.0-1.0) |
| `topK` | int | No | - | Top-K sampling |
| `stopSequences` | []string | No | [] | Stop generation sequences |
| `fallback` | []ModelConfig | No | [] | Fallback models on failure |

### Supported Providers

| Provider | Models | API Key Env |
|----------|--------|-------------|
| `anthropic` | claude-sonnet-4-20250514, claude-sonnet-4-20250514, claude-haiku-4-5-20251001 | `ANTHROPIC_API_KEY` |
| `openai` | gpt-4o, gpt-4o-mini, o1, o1-mini | `OPENAI_API_KEY` |
| `custom` | Any OpenAI-compatible | `CUSTOM_API_KEY` |

---

## System Prompt and Goal

### `spec.system`

**Type:** `string`  
**Required:** No

The system prompt that defines the agent's persona and instructions.

```yaml
spec:
  system: |
    You are a senior software engineer with expertise in distributed systems.
    
    Guidelines:
    - Always write production-quality code
    - Include comprehensive error handling
    - Follow language-specific best practices
    - Document your reasoning
```

### `spec.goal`

**Type:** `string`  
**Required:** No

The specific goal or task for this agent instance.

```yaml
spec:
  goal: |
    Research the competitive landscape for AI code assistants.
    Produce a detailed report saved to /output/report.md.
```

---

## Capabilities Configuration

### `spec.capabilities`

Capabilities grant agents access to system resources in a controlled manner.

```yaml
spec:
  capabilities:
    exec:
      enabled: true
      # Exec-specific config
    fs:
      enabled: true
      # FS-specific config
    net:
      enabled: true
      # Net-specific config
    browser:
      enabled: true
      # Browser-specific config
    memory:
      enabled: true
      # Memory-specific config
    tools:
      enabled: true
      # Tools-specific config
    secrets:
      enabled: true
      # Secrets-specific config
```

### Exec Capability

Sandboxed code execution.

```yaml
capabilities:
  exec:
    enabled: true
    languages:                    # Enabled languages
      - python
      - nodejs
      - bash
      - rust
    timeout: 300s                 # Max execution time
    memory: 512Mi                 # Memory limit
    cpu: "1.0"                    # CPU limit (cores)
    workdir: /workspace           # Working directory
    env:                          # Environment variables
      PYTHONPATH: /workspace/lib
    packages:                     # Pre-installed packages
      python:
        - numpy
        - pandas
      nodejs:
        - axios
```

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `enabled` | bool | No | false | Enable capability |
| `languages` | []string | No | [python, bash] | Allowed languages |
| `timeout` | duration | No | 5m | Maximum execution time |
| `memory` | quantity | No | 256Mi | Memory limit |
| `cpu` | string | No | "0.5" | CPU cores limit |
| `workdir` | string | No | /workspace | Working directory |
| `env` | map | No | {} | Environment variables |
| `packages` | map | No | {} | Pre-installed packages |

### Filesystem Capability

Virtual filesystem access.

```yaml
capabilities:
  fs:
    enabled: true
    mounts:
      - path: /workspace          # Mount path in sandbox
        mode: rw                  # Access mode: ro, rw
        quota: 1Gi                # Storage quota
      - path: /data
        source: s3://bucket/data  # External source
        mode: ro
      - path: /models
        source: gs://bucket/models
        mode: ro
        cache: true               # Cache locally
    snapshot:
      enabled: true               # Enable snapshots
      interval: 5m                # Auto-snapshot interval
      retain: 10                  # Snapshots to retain
```

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `enabled` | bool | No | false | Enable capability |
| `mounts` | []Mount | No | [] | Filesystem mounts |
| `snapshot.enabled` | bool | No | false | Enable snapshots |
| `snapshot.interval` | duration | No | - | Auto-snapshot interval |
| `snapshot.retain` | int | No | 5 | Snapshots to retain |

#### Mount Configuration

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `path` | string | Yes | - | Mount path in sandbox |
| `source` | string | No | - | External source (s3://, gs://, local path) |
| `mode` | string | No | ro | Access mode: `ro`, `rw` |
| `quota` | quantity | No | 1Gi | Storage quota |
| `cache` | bool | No | false | Cache remote files locally |

### Network Capability

Network access control.

```yaml
capabilities:
  net:
    enabled: true
    allowlist:                    # Allowed domains
      - "*.wikipedia.org"
      - "api.github.com"
      - "*.anthropic.com"
    denylist:                     # Blocked domains
      - "*.malware.com"
      - "internal.company.com"
    rateLimit:
      requests: 100               # Requests per window
      per: 1m                     # Time window
    proxy:
      http: http://proxy:8080     # HTTP proxy
      https: http://proxy:8080    # HTTPS proxy
    dns:
      servers:                    # Custom DNS servers
        - 8.8.8.8
        - 8.8.4.4
```

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `enabled` | bool | No | false | Enable capability |
| `allowlist` | []string | No | ["*"] | Allowed domain patterns |
| `denylist` | []string | No | [] | Blocked domain patterns |
| `rateLimit.requests` | int | No | 1000 | Rate limit requests |
| `rateLimit.per` | duration | No | 1m | Rate limit window |
| `proxy.http` | string | No | - | HTTP proxy URL |
| `proxy.https` | string | No | - | HTTPS proxy URL |
| `dns.servers` | []string | No | system | DNS servers |

### Browser Capability

Headless browser automation.

```yaml
capabilities:
  browser:
    enabled: true
    headless: true                # Headless mode
    stealth: true                 # Anti-detection
    timeout: 60s                  # Page load timeout
    viewport:
      width: 1920
      height: 1080
    userAgent: "Mozilla/5.0..."   # Custom user agent
    pool:
      size: 5                     # Browser pool size
      idleTimeout: 5m             # Idle browser timeout
    capture:
      screenshots: true           # Enable screenshots
      video: false                # Enable video recording
      har: true                   # Capture HAR files
    proxy:
      url: http://proxy:8080
      username: user
      password: ${PROXY_PASSWORD}
```

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `enabled` | bool | No | false | Enable capability |
| `headless` | bool | No | true | Headless mode |
| `stealth` | bool | No | true | Anti-detection mode |
| `timeout` | duration | No | 30s | Page load timeout |
| `viewport.width` | int | No | 1920 | Viewport width |
| `viewport.height` | int | No | 1080 | Viewport height |
| `userAgent` | string | No | Chrome default | User agent string |
| `pool.size` | int | No | 3 | Browser pool size |
| `pool.idleTimeout` | duration | No | 5m | Idle timeout |
| `capture.screenshots` | bool | No | true | Enable screenshots |
| `capture.video` | bool | No | false | Enable video |
| `capture.har` | bool | No | false | Capture HAR |

### Memory Capability

Persistent memory stores.

```yaml
capabilities:
  memory:
    enabled: true
    vector:
      enabled: true
      dimensions: 1536            # Embedding dimensions
      metric: cosine              # Distance metric
      maxItems: 100000            # Maximum items
      indexType: hnsw             # Index type
    graph:
      enabled: true
      maxNodes: 100000
      maxEdges: 1000000
    kv:
      enabled: true
      maxKeys: 100000
      maxValueSize: 1Mi
    ttl: 24h                      # Default TTL for all stores
    persistence:
      enabled: true               # Persist to disk
      path: /data/memory
```

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `enabled` | bool | No | false | Enable capability |
| `vector.enabled` | bool | No | true | Enable vector store |
| `vector.dimensions` | int | No | 1536 | Embedding dimensions |
| `vector.metric` | string | No | cosine | Distance: cosine, euclidean, dot |
| `vector.maxItems` | int | No | 100000 | Maximum items |
| `graph.enabled` | bool | No | true | Enable graph store |
| `kv.enabled` | bool | No | true | Enable KV store |
| `ttl` | duration | No | - | Default TTL |
| `persistence.enabled` | bool | No | true | Persist to disk |

### Tools Capability

External tool integration.

```yaml
capabilities:
  tools:
    enabled: true
    builtin:                      # Built-in tools
      - calculator
      - datetime
      - json_parser
      - regex
      - uuid
    mcp:                          # MCP servers
      - uri: "http://localhost:3000/mcp"
        name: custom-tools
        auth:
          type: bearer
          token: ${MCP_TOKEN}
    custom:                       # Custom tool definitions
      - name: analyze_sentiment
        description: "Analyze sentiment of text"
        schema:
          type: object
          properties:
            text:
              type: string
              description: "Text to analyze"
          required: [text]
        handler: /plugins/sentiment.wasm
    timeout: 30s                  # Tool execution timeout
    retries: 3                    # Retry failed invocations
```

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `enabled` | bool | No | false | Enable capability |
| `builtin` | []string | No | [] | Built-in tool names |
| `mcp` | []MCPConfig | No | [] | MCP server connections |
| `custom` | []ToolConfig | No | [] | Custom tool definitions |
| `timeout` | duration | No | 30s | Tool execution timeout |
| `retries` | int | No | 3 | Retry count |

#### Built-in Tools

| Tool | Description |
|------|-------------|
| `calculator` | Mathematical calculations |
| `datetime` | Date/time operations |
| `json_parser` | JSON parsing and manipulation |
| `regex` | Regular expression matching |
| `uuid` | UUID generation |
| `hash` | Hashing functions (SHA-256, MD5, etc.) |
| `base64` | Base64 encoding/decoding |
| `url_parser` | URL parsing and manipulation |

### Secrets Capability

Secret management and injection.

```yaml
capabilities:
  secrets:
    enabled: true
    inject:
      - name: API_KEY             # Environment variable name
        source: vault://secret/data/api-key#key
      - name: DB_PASSWORD
        source: env://DB_PASSWORD
      - name: SERVICE_ACCOUNT
        source: file:///etc/secrets/sa.json
        encoding: base64
    vault:
      address: https://vault.example.com
      auth:
        method: kubernetes        # kubernetes, token, approle
        role: spawn-agent
    refresh:
      enabled: true               # Auto-refresh secrets
      interval: 1h                # Refresh interval
```

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `enabled` | bool | No | false | Enable capability |
| `inject` | []SecretRef | No | [] | Secrets to inject |
| `vault.address` | string | No | - | Vault server address |
| `vault.auth.method` | string | No | token | Auth method |
| `refresh.enabled` | bool | No | false | Auto-refresh |
| `refresh.interval` | duration | No | 1h | Refresh interval |

#### Secret Sources

| Source | Format | Description |
|--------|--------|-------------|
| `vault://` | `vault://path/to/secret#key` | HashiCorp Vault |
| `env://` | `env://VAR_NAME` | Environment variable |
| `file://` | `file:///path/to/file` | Local file |
| `k8s://` | `k8s://namespace/secret#key` | Kubernetes secret |

---

## Resource Configuration

### `spec.resources`

Resource requests, limits, and cost controls.

```yaml
spec:
  resources:
    requests:
      memory: 256Mi               # Requested memory
      cpu: "0.5"                  # Requested CPU cores
    limits:
      memory: 1Gi                 # Maximum memory
      cpu: "2.0"                  # Maximum CPU cores
      disk: 10Gi                  # Maximum disk space
      pids: 100                   # Maximum processes
    costLimit:
      hourly: 1.00                # Hourly cost limit (USD)
      daily: 10.00                # Daily cost limit
      monthly: 100.00             # Monthly cost limit
      action: pause               # Action: notify, pause, terminate
    timeout:
      session: 1h                 # Maximum session duration
      idle: 10m                   # Idle timeout
```

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `requests.memory` | quantity | No | 256Mi | Requested memory |
| `requests.cpu` | string | No | "0.5" | Requested CPU |
| `limits.memory` | quantity | No | 1Gi | Memory limit |
| `limits.cpu` | string | No | "2.0" | CPU limit |
| `limits.disk` | quantity | No | 10Gi | Disk limit |
| `limits.pids` | int | No | 100 | Process limit |
| `costLimit.hourly` | float | No | - | Hourly cost limit |
| `costLimit.daily` | float | No | - | Daily cost limit |
| `costLimit.monthly` | float | No | - | Monthly cost limit |
| `costLimit.action` | string | No | notify | Limit action |
| `timeout.session` | duration | No | 1h | Session timeout |
| `timeout.idle` | duration | No | 10m | Idle timeout |

---

## Sandbox Configuration

### `spec.sandbox`

Sandbox runtime and security settings.

```yaml
spec:
  sandbox:
    runtime: gvisor               # Runtime: gvisor, firecracker, docker, native
    networkPolicy: restricted     # Network: none, restricted, egress-only, full
    seccompProfile: strict        # Seccomp: strict, moderate, permissive
    capabilities: []              # Linux capabilities to add
    readOnlyRoot: true            # Read-only root filesystem
    privileged: false             # Privileged mode (dangerous!)
    user:
      uid: 1000                   # User ID
      gid: 1000                   # Group ID
```

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `runtime` | string | No | gvisor | Sandbox runtime |
| `networkPolicy` | string | No | restricted | Network policy |
| `seccompProfile` | string | No | strict | Seccomp profile |
| `capabilities` | []string | No | [] | Added capabilities |
| `readOnlyRoot` | bool | No | true | Read-only root FS |
| `privileged` | bool | No | false | Privileged mode |
| `user.uid` | int | No | 1000 | User ID |
| `user.gid` | int | No | 1000 | Group ID |

### Runtime Comparison

| Runtime | Isolation | Overhead | Startup | Use Case |
|---------|-----------|----------|---------|----------|
| `gvisor` | High | ~5% | 150ms | Production default |
| `firecracker` | Maximum | ~10% | 300ms | Multi-tenant |
| `docker` | Medium | ~1% | 50ms | Development |
| `native` | None | 0% | 0ms | Testing only |

### Seccomp Profiles

| Profile | Blocked Syscalls | Use Case |
|---------|-----------------|----------|
| `strict` | 200+ dangerous syscalls | Production |
| `moderate` | Known dangerous syscalls | Compatibility |
| `permissive` | Minimal blocking | Debugging |

---

## Lifecycle Hooks

### `spec.hooks`

Lifecycle hooks for setup, cleanup, and health checking.

```yaml
spec:
  hooks:
    preStart:
      - command: ["pip", "install", "-r", "requirements.txt"]
        timeout: 5m
      - command: ["./setup.sh"]
        env:
          SETUP_MODE: production
    postStart:
      - command: ["./warmup.sh"]
    preStop:
      - command: ["./save-state.sh"]
    postStop:
      - command: ["./cleanup.sh"]
    healthCheck:
      command: ["./health.sh"]
      interval: 30s
      timeout: 5s
      retries: 3
      startPeriod: 10s
```

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `preStart` | []Hook | No | [] | Before agent starts |
| `postStart` | []Hook | No | [] | After agent starts |
| `preStop` | []Hook | No | [] | Before agent stops |
| `postStop` | []Hook | No | [] | After agent stops |
| `healthCheck` | HealthCheck | No | - | Health check config |

---

## Observability Configuration

### `spec.observability`

Telemetry and observability settings.

```yaml
spec:
  observability:
    traces:
      enabled: true
      sampleRate: 1.0             # 1.0 = 100% sampling
      exporter: otlp              # Exporter: otlp, jaeger, zipkin
      endpoint: localhost:4317    # Collector endpoint
    metrics:
      enabled: true
      exporter: prometheus        # Exporter: prometheus, otlp
      port: 9091                  # Metrics port
      path: /metrics              # Metrics path
    logs:
      level: info                 # Level: debug, info, warn, error
      format: json                # Format: json, text
      output: stdout              # Output: stdout, file
      file:
        path: /var/log/agent.log
        maxSize: 100Mi
        maxAge: 7d
        compress: true
    events:
      stream: true                # Stream events
      webhooks:
        - url: https://hooks.example.com/spawn
          events: [started, completed, failed]
          secret: ${WEBHOOK_SECRET}
```

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `traces.enabled` | bool | No | true | Enable tracing |
| `traces.sampleRate` | float | No | 1.0 | Sample rate (0.0-1.0) |
| `traces.exporter` | string | No | otlp | Trace exporter |
| `metrics.enabled` | bool | No | true | Enable metrics |
| `metrics.exporter` | string | No | prometheus | Metrics exporter |
| `logs.level` | string | No | info | Log level |
| `logs.format` | string | No | json | Log format |
| `events.stream` | bool | No | true | Stream events |

---

## Scaling Configuration

### `spec.scaling`

Auto-scaling configuration for agent replicas.

```yaml
spec:
  scaling:
    minReplicas: 1                # Minimum replicas
    maxReplicas: 10               # Maximum replicas
    metrics:
      - type: queue-depth         # Metric type
        target: 5                 # Target value
      - type: cpu
        target: 80                # Target percentage
      - type: memory
        target: 80
    behavior:
      scaleUp:
        stabilizationWindow: 60s
        policies:
          - type: Pods
            value: 4
            periodSeconds: 60
      scaleDown:
        stabilizationWindow: 300s
        policies:
          - type: Percent
            value: 10
            periodSeconds: 60
```

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `minReplicas` | int | No | 1 | Minimum replicas |
| `maxReplicas` | int | No | 10 | Maximum replicas |
| `metrics` | []ScalingMetric | No | [] | Scaling metrics |

---

## Mesh Configuration

### `spec.mesh`

Multi-agent mesh and communication settings.

```yaml
spec:
  mesh:
    enabled: true
    channels:
      - name: findings            # Channel name
        type: pubsub              # Type: pubsub, request-reply, stream
        topic: research.findings  # Topic name
        publish: true             # Can publish
        subscribe: false          # Can subscribe
      - name: requests
        type: request-reply
        timeout: 30s
        subscribe: true
    discovery:
      enabled: true
      labels:                     # Discover agents with labels
        team: research
    consensus:
      enabled: false              # Enable distributed consensus
      algorithm: raft             # Algorithm: raft
```

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `enabled` | bool | No | true | Enable mesh |
| `channels` | []Channel | No | [] | Communication channels |
| `discovery.enabled` | bool | No | true | Enable discovery |
| `discovery.labels` | map | No | {} | Label selector |
| `consensus.enabled` | bool | No | false | Enable consensus |

---

## Complete Example

```yaml
apiVersion: spawn.dev/v1
kind: Agent

metadata:
  name: senior-researcher
  namespace: production
  labels:
    team: ai-research
    tier: production
    cost-center: research-42
  annotations:
    spawn.dev/description: "Production research agent for market analysis"
    spawn.dev/owner: "research-team@company.com"

spec:
  model:
    provider: anthropic
    name: claude-sonnet-4-20250514
    temperature: 0.7
    maxTokens: 8192
    fallback:
      - provider: openai
        name: gpt-4o

  system: |
    You are a senior research analyst with expertise in market research,
    competitive analysis, and technology trends.
    
    Guidelines:
    - Verify information from multiple sources
    - Cite sources for all claims
    - Present balanced perspectives
    - Structure findings clearly

  goal: |
    Research the topic provided and produce a comprehensive report
    saved to /output/report.md with proper citations.

  capabilities:
    exec:
      enabled: true
      languages: [python, bash]
      timeout: 5m
      memory: 512Mi
      packages:
        python: [requests, beautifulsoup4, pandas]
    
    fs:
      enabled: true
      mounts:
        - path: /output
          mode: rw
          quota: 1Gi
        - path: /reference
          source: s3://company-data/research
          mode: ro
    
    net:
      enabled: true
      allowlist:
        - "*.wikipedia.org"
        - "*.arxiv.org"
        - "*.github.com"
        - "news.ycombinator.com"
      rateLimit:
        requests: 100
        per: 1m
    
    browser:
      enabled: true
      headless: true
      stealth: true
      capture:
        screenshots: true
    
    memory:
      enabled: true
      vector:
        dimensions: 1536
      ttl: 7d
    
    tools:
      enabled: true
      builtin: [calculator, datetime, json_parser]
    
    secrets:
      enabled: true
      inject:
        - name: SERP_API_KEY
          source: vault://secret/research/serp-api

  resources:
    requests:
      memory: 512Mi
      cpu: "1.0"
    limits:
      memory: 2Gi
      cpu: "4.0"
    costLimit:
      daily: 50.00
      action: pause
    timeout:
      session: 2h

  sandbox:
    runtime: gvisor
    networkPolicy: restricted
    seccompProfile: strict

  hooks:
    preStart:
      - command: ["pip", "install", "-r", "/workspace/requirements.txt"]
    healthCheck:
      command: ["python", "-c", "print('healthy')"]
      interval: 30s

  observability:
    traces:
      enabled: true
      sampleRate: 1.0
    metrics:
      enabled: true
    logs:
      level: info
      format: json

  mesh:
    channels:
      - name: findings
        type: pubsub
        topic: research.findings
        publish: true
```

## Validation

Validate configurations before deployment:

```bash
# Validate single file
spawn validate agent.yaml

# Validate all files in directory
spawn validate ./agents/

# Validate with strict mode
spawn validate --strict agent.yaml
```

## Next Steps

- [Daemon Configuration](daemon-config.md)
- [Capabilities Deep Dive](../capabilities/overview.md)
- [Configuration Examples](examples.md)
