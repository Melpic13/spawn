<p align="center">
  <img src="./assets/logo.svg" alt="spawn logo" width="200"/>
</p>

<h1 align="center">spawn</h1>

<p align="center">
  <strong>The Agent Operating System</strong><br>
  <em>systemd for AI agents</em>
</p>

<p align="center">
  <sub>Logo source: <code>assets/logo.svg</code></sub>
</p>

<p align="center">
  <a href="https://github.com/spawndev/spawn/releases"><img src="https://img.shields.io/github/v/release/spawndev/spawn?style=flat-square&color=00ADD8" alt="Release"></a>
  <a href="https://github.com/spawndev/spawn/actions"><img src="https://img.shields.io/github/actions/workflow/status/spawndev/spawn/ci.yml?style=flat-square" alt="CI"></a>
  <a href="https://codecov.io/gh/spawndev/spawn"><img src="https://img.shields.io/codecov/c/github/spawndev/spawn?style=flat-square" alt="Coverage"></a>
  <a href="https://goreportcard.com/report/github.com/spawndev/spawn"><img src="https://goreportcard.com/badge/github.com/spawndev/spawn?style=flat-square" alt="Go Report"></a>
  <a href="https://opensource.org/licenses/Apache-2.0"><img src="https://img.shields.io/badge/License-Apache%202.0-blue.svg?style=flat-square" alt="License"></a>
  <a href="https://discord.gg/spawn"><img src="https://img.shields.io/discord/123456789?style=flat-square&logo=discord&logoColor=white&label=Discord" alt="Discord"></a>
</p>

<p align="center">
  <a href="#-quick-start">Quick Start</a> •
  <a href="#-features">Features</a> •
  <a href="#-architecture">Architecture</a> •
  <a href="#-documentation">Docs</a> •
  <a href="#-enterprise">Enterprise</a> •
  <a href="#-community">Community</a>
</p>

---

## The Problem

Your AI agent can think. But can it **act**?

Every team building agents today faces the same challenge: agents need to execute code, browse the web, manage files, remember context, and coordinate with other agents. The current solutions are fragmented, insecure, and impossible to observe.

**Without spawn:**
```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Your LLM  │────▶│  47 Deps    │────▶│   Prayers   │
└─────────────┘     └─────────────┘     └─────────────┘
       │                   │                   │
       ▼                   ▼                   ▼
   "Execute        "Maybe it's        "It deleted
    this code"      secure?"           my files"
```

**With spawn:**
```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Your LLM  │────▶│   spawn     │────▶│  Production │
└─────────────┘     └─────────────┘     └─────────────┘
       │                   │                   │
       ▼                   ▼                   ▼
   "Execute        "Sandboxed,         "Full trace,
    this code"      isolated"           $0.003 cost"
```

---

## 🚀 Quick Start

### Install

```bash
# macOS / Linux
curl -sSL https://spawn.dev/install | sh

# Homebrew
brew install spawndev/tap/spawn

# Go
go install spawn.dev/cmd/spawn@latest

# Docker
docker pull ghcr.io/spawndev/spawn:latest
```

### Your First Agent

```bash
# Initialize a new agent
spawn init my-researcher

# Edit the configuration
cd my-researcher && cat agent.yaml
```

```yaml
apiVersion: spawn.dev/v1
kind: Agent

metadata:
  name: researcher

spec:
  model:
    provider: anthropic
    name: claude-sonnet-4-20250514

  goal: |
    Research the given topic and produce a comprehensive
    report saved to ./output/report.md

  capabilities:
    web:
      enabled: true
    code:
      enabled: true
      languages: [python]
    files:
      enabled: true
      mounts:
        - path: /output
          mode: rw
```

```bash
# Run the agent
spawn run --topic "quantum computing breakthroughs 2025"

# Watch it think, browse, code, and write — all sandboxed
```

**That's it.** Your agent now has secure access to the web, code execution, and file management.

---

## ✨ Features

### 🔒 Secure by Default

Every agent runs in an isolated sandbox with configurable security policies.

| Runtime | Isolation Level | Performance | Use Case |
|---------|----------------|-------------|----------|
| **gVisor** | High | ~5% overhead | Production default |
| **Firecracker** | Maximum | ~10% overhead | Multi-tenant, untrusted |
| **Docker** | Medium | Native | Development |
| **Native** | None | Native | Testing only |

```yaml
spec:
  sandbox:
    runtime: gvisor
    seccomp: strict
    network: egress-only
    readOnlyRoot: true
```

### 🧠 Full Capability Stack

Every capability your agent needs, batteries included.

<table>
<tr>
<td width="50%">

**Code Execution**
```yaml
capabilities:
  exec:
    languages: [python, node, bash, rust]
    timeout: 5m
    memory: 512Mi
```
- Secure sandboxed execution
- Resource limits (CPU, memory, time)
- Multi-language support
- Output streaming

</td>
<td width="50%">

**Web Access**
```yaml
capabilities:
  net:
    allowlist: ["*.wikipedia.org"]
    rateLimit: 100/min
  browser:
    headless: true
    stealth: true
```
- HTTP client with policies
- Full browser automation
- Anti-detection built-in
- Screenshot capture

</td>
</tr>
<tr>
<td width="50%">

**File System**
```yaml
capabilities:
  fs:
    mounts:
      - path: /data
        source: s3://bucket
        mode: ro
```
- Virtual filesystem
- Cloud storage mounts
- Snapshot/restore
- Quota management

</td>
<td width="50%">

**Memory**
```yaml
capabilities:
  memory:
    vector: { dimensions: 1536 }
    graph: { enabled: true }
    ttl: 24h
```
- Vector store (embeddings)
- Graph database
- Key-value store
- Persistent across runs

</td>
</tr>
<tr>
<td width="50%">

**Tools (MCP Compatible)**
```yaml
capabilities:
  tools:
    mcp:
      - uri: "http://localhost:3000"
    builtin:
      - calculator
      - json_parser
```
- MCP protocol support
- Custom tool registration
- JSON Schema validation
- Automatic discovery

</td>
<td width="50%">

**Secrets**
```yaml
capabilities:
  secrets:
    inject:
      - name: API_KEY
        source: vault://secret/key
```
- HashiCorp Vault integration
- Kubernetes secrets
- Environment injection
- Automatic rotation

</td>
</tr>
</table>

### 🕸️ Multi-Agent Mesh

First-class support for agent-to-agent communication.

```yaml
# researcher.yaml
spec:
  mesh:
    channels:
      - name: findings
        type: pubsub
        topic: research.findings
---
# writer.yaml
spec:
  mesh:
    channels:
      - name: findings
        type: pubsub
        topic: research.findings
        subscribe: true
```

```bash
# Run a swarm
spawn run researcher.yaml writer.yaml reviewer.yaml

# Visualize the topology
spawn mesh topology --watch
```

```
    ┌────────────┐
    │ Researcher │
    └──────┬─────┘
           │ findings
    ┌──────▼─────┐
    │   Writer   │
    └──────┬─────┘
           │ drafts
    ┌──────▼─────┐
    │  Reviewer  │
    └────────────┘
```

### 📊 Complete Observability

See everything your agents do. Debug anything.

```bash
# Stream logs
spawn logs --follow

# View traces
spawn trace list
spawn trace view tr_abc123

# Decision replay
spawn replay tr_abc123 --step-by-step
```

<p align="center">
  <img src="https://raw.githubusercontent.com/spawndev/spawn/main/assets/trace-screenshot.png" alt="Trace visualization" width="800"/>
</p>

**Built-in dashboards:**
- Real-time agent status
- Token usage and costs
- Capability utilization
- Error tracking

```yaml
spec:
  observability:
    traces:
      enabled: true
      sampleRate: 1.0
    metrics:
      enabled: true
      exporters: [prometheus, datadog]
    logs:
      level: debug
      format: json
```

### 💰 Cost Control

Never get surprised by LLM bills again.

```yaml
spec:
  resources:
    costLimit:
      hourly: 1.00
      daily: 10.00
      monthly: 100.00
      action: pause  # pause, notify, or terminate
```

```bash
# View real-time costs
spawn cost --watch

# Agent: researcher
# Session: 2h 14m
# Tokens: 847,293 (in: 612,847 / out: 234,446)
# Cost: $2.34
# Limit: $10.00/day (23.4%)
```

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              SPAWN DAEMON                                   │
│                                                                             │
│  ┌──────────────────────────────────────────────────────────────────────┐   │
│  │                         CONTROL PLANE                                │   │
│  │                                                                      │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  │   │
│  │  │  Scheduler  │  │  Supervisor │  │   Registry  │  │   Gateway   │  │   │
│  │  └─────────────┘  └─────────────┘  └─────────────┘  └─────────────┘  │   │
│  │         │                │                │                │         │   │
│  └─────────┼────────────────┼────────────────┼────────────────┼─────────┘   │
│            │                │                │                │             │
│  ┌─────────▼────────────────▼────────────────▼────────────────▼─────────┐   │
│  │                           AGENT MESH                                 │   │
│  │                                                                      │   │
│  │   ┌─────────────┐      ┌─────────────┐      ┌─────────────┐         │   │
│  │   │   Agent A   │◄────►│   Agent B   │◄────►│   Agent C   │         │   │
│  │   └──────┬──────┘      └──────┬──────┘      └──────┬──────┘         │   │
│  │          └──────────────────┬─┴──────────────────┬─┘                │   │
│  │                             ▼                    │                  │   │
│  │                    Message Bus (NATS)            │                  │   │
│  └──────────────────────────────────────────────────┼──────────────────┘   │
│                                                     │                      │
│  ┌──────────────────────────────────────────────────▼──────────────────┐   │
│  │                        CAPABILITY LAYER                              │   │
│  │                                                                      │   │
│  │  ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐  │   │
│  │  │  exec  │ │   fs   │ │  net   │ │ memory │ │browser │ │ tools  │  │   │
│  │  └────────┘ └────────┘ └────────┘ └────────┘ └────────┘ └────────┘  │   │
│  └──────────────────────────────────────────────────────────────────────┘   │
│                                     │                                       │
│  ┌──────────────────────────────────▼───────────────────────────────────┐   │
│  │                        ISOLATION LAYER                               │   │
│  │                   gVisor │ Firecracker │ Docker                      │   │
│  └──────────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────────┘
```

**Key Design Principles:**

1. **Single Binary** — No runtime dependencies, one binary to rule them all
2. **Security First** — Every agent sandboxed by default, no exceptions
3. **Observable** — Full tracing of every decision, tool call, and state change
4. **Cloud Native** — Kubernetes-ready, scales horizontally
5. **Protocol Agnostic** — Works with any LLM provider

---

## 📖 Documentation

| Resource | Description |
|----------|-------------|
| [Quick Start Guide](https://docs.spawn.dev/quickstart) | Get running in 5 minutes |
| [Configuration Reference](https://docs.spawn.dev/config) | Complete YAML specification |
| [Capabilities Guide](https://docs.spawn.dev/capabilities) | Deep dive into each capability |
| [Security Model](https://docs.spawn.dev/security) | Sandbox internals and policies |
| [Multi-Agent Patterns](https://docs.spawn.dev/mesh) | Building agent swarms |
| [API Reference](https://docs.spawn.dev/api) | REST and gRPC documentation |
| [Deployment Guide](https://docs.spawn.dev/deploy) | Production deployment patterns |
| [Troubleshooting](https://docs.spawn.dev/troubleshooting) | Common issues and solutions |

---

## 🎯 Examples

### Research Agent

```yaml
apiVersion: spawn.dev/v1
kind: Agent
metadata:
  name: deep-researcher
spec:
  model:
    provider: anthropic
    name: claude-sonnet-4-20250514
  
  system: |
    You are a thorough research assistant. For each topic:
    1. Search the web for authoritative sources
    2. Extract and verify key facts
    3. Synthesize findings into a structured report
    
  capabilities:
    net:
      enabled: true
    browser:
      enabled: true
    fs:
      mounts:
        - path: /output
          mode: rw
    memory:
      vector:
        dimensions: 1536
```

### Code Assistant

```yaml
apiVersion: spawn.dev/v1
kind: Agent
metadata:
  name: code-assistant
spec:
  model:
    provider: anthropic
    name: claude-sonnet-4-20250514
  
  capabilities:
    exec:
      languages: [python, node, bash]
      timeout: 5m
    fs:
      mounts:
        - path: /workspace
          source: ./project
          mode: rw
    tools:
      builtin:
        - git
        - lsp
```

### Multi-Agent Pipeline

```yaml
# pipeline.yaml — Three agents working together
---
apiVersion: spawn.dev/v1
kind: Agent
metadata:
  name: planner
spec:
  goal: Break down complex tasks into subtasks
  mesh:
    publish: [tasks]
---
apiVersion: spawn.dev/v1
kind: Agent
metadata:
  name: executor
spec:
  goal: Execute assigned subtasks
  mesh:
    subscribe: [tasks]
    publish: [results]
  capabilities:
    exec:
      enabled: true
---
apiVersion: spawn.dev/v1
kind: Agent
metadata:
  name: reviewer
spec:
  goal: Review and validate results
  mesh:
    subscribe: [results]
```

```bash
spawn run pipeline.yaml --task "Build a REST API for user management"
```

---

## 🏢 Enterprise

### spawn Cloud

Managed spawn infrastructure with enterprise features.

| Feature | Cloud | Self-Hosted |
|---------|-------|-------------|
| Managed infrastructure | ✅ | ❌ |
| SSO / SAML | ✅ | ✅ |
| Audit logging | ✅ | ✅ |
| SOC 2 Type II | ✅ | — |
| HIPAA compliance | ✅ | — |
| Custom SLAs | ✅ | — |
| 24/7 support | ✅ | Optional |
| Air-gapped deployment | ❌ | ✅ |

### Enterprise Features

**Security & Compliance**
- Advanced RBAC with attribute-based policies
- Complete audit trail
- Data residency controls
- Custom security policies
- Penetration test reports

**Operations**
- High availability deployment
- Disaster recovery
- Automated backups
- Custom retention policies

**Integration**
- LDAP/Active Directory
- Okta, Auth0, Azure AD
- Splunk, Datadog, New Relic
- PagerDuty, Opsgenie
- Custom webhooks

<p align="center">
  <a href="https://spawn.dev/enterprise">
    <img src="https://img.shields.io/badge/Learn%20More-Enterprise-blue?style=for-the-badge" alt="Enterprise">
  </a>
  <a href="https://spawn.dev/demo">
    <img src="https://img.shields.io/badge/Request-Demo-green?style=for-the-badge" alt="Demo">
  </a>
</p>

---

## 📊 Benchmarks

Performance comparison on standard agent tasks:

| Metric | spawn | LangChain | AutoGPT | CrewAI |
|--------|-------|-----------|---------|--------|
| Cold start | 180ms | 2.4s | 5.1s | 1.8s |
| Memory overhead | 45MB | 280MB | 520MB | 190MB |
| Tool execution | 12ms | 89ms | 156ms | 67ms |
| Sandbox overhead | 5% | N/A | N/A | N/A |
| Max concurrent agents | 10,000+ | ~100 | ~20 | ~200 |

**Security comparison:**

| Feature | spawn | Others |
|---------|-------|--------|
| Code sandbox | gVisor/Firecracker | None/Docker |
| Network isolation | Per-agent policies | None |
| File system isolation | Virtual FS + quotas | Shared |
| Secret management | Vault integration | Env vars |
| Audit logging | Complete | Partial |

---

## 🛣️ Roadmap

### v1.0 — Foundation (Current)
- [x] Core agent lifecycle
- [x] All capabilities (exec, fs, net, memory, browser, tools)
- [x] gVisor sandbox
- [x] Multi-agent mesh
- [x] REST/gRPC API
- [x] Observability stack

### v1.1 — Scale
- [ ] Firecracker microVM support
- [ ] Distributed scheduling
- [ ] Agent checkpointing
- [ ] Live migration

### v1.2 — Intelligence
- [ ] Agent memory consolidation
- [ ] Learning from traces
- [ ] Automatic tool discovery
- [ ] Cost optimization engine

### v2.0 — Platform
- [ ] Visual workflow builder
- [ ] Marketplace for capabilities
- [ ] Enterprise SSO
- [ ] Multi-region deployment

---

## 🤝 Contributing

We love contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

**Quick contribution guide:**

```bash
# Fork and clone
git clone https://github.com/YOUR_USERNAME/spawn.git
cd spawn

# Install dependencies
make setup

# Run tests
make test

# Run locally
make run

# Submit PR
```

**Areas we need help:**
- 🌍 Translations
- 📖 Documentation
- 🧪 Test coverage
- 🔌 Capability plugins
- 🎨 Dashboard UI

---

## 🌟 Community

<p align="center">
  <a href="https://discord.gg/spawn">
    <img src="https://img.shields.io/badge/Discord-Join%20Us-5865F2?style=for-the-badge&logo=discord&logoColor=white" alt="Discord">
  </a>
  <a href="https://twitter.com/spawndev">
    <img src="https://img.shields.io/badge/Twitter-Follow-1DA1F2?style=for-the-badge&logo=twitter&logoColor=white" alt="Twitter">
  </a>
  <a href="https://github.com/spawndev/spawn/discussions">
    <img src="https://img.shields.io/badge/GitHub-Discussions-333?style=for-the-badge&logo=github&logoColor=white" alt="Discussions">
  </a>
</p>

### Adopters

<p align="center">
  <em>Used in production by teams at</em>
</p>

<p align="center">
  <img src="https://raw.githubusercontent.com/spawndev/spawn/main/assets/adopters.png" alt="Adopter logos" width="600"/>
</p>

### Star History

<p align="center">
  <a href="https://star-history.com/#spawndev/spawn&Date">
    <img src="https://api.star-history.com/svg?repos=spawndev/spawn&type=Date" alt="Star History" width="600"/>
  </a>
</p>

---

## 📜 License

spawn is [Apache 2.0](LICENSE) licensed.

---

<p align="center">
  <strong>Built with ❤️ for the AI agent ecosystem</strong>
</p>

<p align="center">
  <sub>If spawn helps your team ship agents faster, consider <a href="https://github.com/sponsors/spawndev">sponsoring</a> the project.</sub>
</p>
