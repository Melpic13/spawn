# Security Model

This document describes spawn's comprehensive security architecture, threat model, and security controls.

## Security Philosophy

spawn follows the principle of **defense in depth** with **secure defaults**. Every agent runs in an isolated sandbox by default, with capabilities explicitly granted rather than implicitly available.

### Core Principles

1. **Least Privilege**: Agents only get access to explicitly granted capabilities
2. **Isolation by Default**: Every agent runs in its own sandbox
3. **Zero Trust**: All inter-agent communication is authenticated
4. **Auditability**: Every action is logged and traceable
5. **Fail Secure**: On error, default to denying access

---

## Threat Model

### Assets to Protect

| Asset | Description | Sensitivity |
|-------|-------------|-------------|
| Host System | The machine running spawnd | Critical |
| Other Agents | Co-located agents | High |
| User Data | Files, secrets, credentials | High |
| LLM Credentials | API keys for providers | High |
| Network Resources | Internal services | Medium |
| Compute Resources | CPU, memory, disk | Medium |

### Threat Actors

| Actor | Capability | Motivation |
|-------|------------|------------|
| Malicious Prompt | Prompt injection via user input | Data exfiltration, system access |
| Compromised Agent | Agent executing malicious code | Lateral movement, persistence |
| Malicious Tool | External tool with backdoor | Credential theft, code execution |
| Network Attacker | Man-in-the-middle, eavesdropping | Data interception, injection |

### Attack Vectors

```
┌─────────────────────────────────────────────────────────────────┐
│                        ATTACK SURFACE                           │
│                                                                 │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │
│  │   Prompt    │  │    Tool     │  │   Network   │             │
│  │  Injection  │  │  Execution  │  │   Access    │             │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘             │
│         │                │                │                     │
│         ▼                ▼                ▼                     │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                    SPAWN DEFENSES                        │   │
│  │                                                          │   │
│  │  • Input validation    • Sandboxed execution             │   │
│  │  • Output filtering    • Capability restrictions         │   │
│  │  • Tool schema         • Network policies                │   │
│  │    validation          • Rate limiting                   │   │
│  └─────────────────────────────────────────────────────────┘   │
│                              │                                  │
│                              ▼                                  │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │                    PROTECTED ASSETS                      │   │
│  │                                                          │   │
│  │  • Host system         • User data                       │   │
│  │  • Other agents        • Credentials                     │   │
│  │  • Internal services   • Compute resources               │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

---

## Sandbox Architecture

### Isolation Layers

```
┌─────────────────────────────────────────────────────────────────┐
│                         HOST KERNEL                             │
│                                                                 │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │                    gVisor (runsc)                         │  │
│  │                                                           │  │
│  │  ┌────────────────────────────────────────────────────┐  │  │
│  │  │                 SENTRY (User Kernel)               │  │  │
│  │  │                                                    │  │  │
│  │  │  • Implements Linux syscall interface              │  │  │
│  │  │  • Runs in user space                              │  │  │
│  │  │  • Limited host kernel interaction                 │  │  │
│  │  │                                                    │  │  │
│  │  │  ┌────────────────────────────────────────────┐   │  │  │
│  │  │  │              AGENT SANDBOX                  │   │  │  │
│  │  │  │                                             │   │  │  │
│  │  │  │  • Isolated filesystem (overlay)            │   │  │  │
│  │  │  │  • Restricted network namespace             │   │  │  │
│  │  │  │  • Limited capabilities                     │   │  │  │
│  │  │  │  • Resource quotas (cgroups v2)             │   │  │  │
│  │  │  │                                             │   │  │  │
│  │  │  │  ┌─────────────────────────────────────┐   │   │  │  │
│  │  │  │  │           AGENT PROCESS             │   │   │  │  │
│  │  │  │  │                                     │   │   │  │  │
│  │  │  │  │  • Non-root user (uid 1000)         │   │   │  │  │
│  │  │  │  │  • Seccomp filters                  │   │   │  │  │
│  │  │  │  │  • No sensitive mounts              │   │   │  │  │
│  │  │  │  └─────────────────────────────────────┘   │   │  │  │
│  │  │  └────────────────────────────────────────────┘   │  │  │
│  │  └────────────────────────────────────────────────────┘  │  │
│  └──────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
```

### Runtime Options

#### gVisor (Default)

Best balance of security and performance.

```yaml
sandbox:
  runtime: gvisor
```

**Security Properties:**
- Syscall interception and emulation
- User-space kernel implementation
- ~280 syscalls blocked by default
- OCI-compatible

**Blocked Capabilities:**
- Direct hardware access
- Kernel module loading
- Raw sockets (by default)
- ptrace

#### Firecracker (Maximum Isolation)

For multi-tenant or untrusted workloads.

```yaml
sandbox:
  runtime: firecracker
```

**Security Properties:**
- Full VM isolation
- Dedicated kernel per agent
- Hardware-enforced boundaries
- Minimal device model

**Use Cases:**
- Multi-tenant deployments
- Processing untrusted input
- Compliance requirements (FedRAMP, etc.)

#### Docker (Development)

For development and testing only.

```yaml
sandbox:
  runtime: docker
```

**Security Properties:**
- Container namespaces
- Seccomp profiles
- AppArmor/SELinux
- cgroups resource limits

**Warning:** Docker provides weaker isolation than gVisor or Firecracker.

---

## Capability Security

### Capability Model

Each capability is:
1. **Disabled by default** — Must be explicitly enabled
2. **Scoped** — Limited to specific operations
3. **Audited** — All invocations logged
4. **Rate-limited** — Prevents abuse

```yaml
capabilities:
  exec:
    enabled: true           # Explicitly enable
    languages: [python]     # Scope to specific languages
    timeout: 5m             # Limit duration
    memory: 512Mi           # Limit resources
```

### Exec Capability Security

Code execution is the highest-risk capability.

**Controls:**

| Control | Description |
|---------|-------------|
| Language allowlist | Only specified languages can run |
| Timeout | Maximum execution time |
| Resource limits | CPU, memory, PIDs, disk |
| No network | Network disabled inside exec by default |
| Read-only code | Code directory is read-only |
| No setuid | setuid binaries disabled |
| Seccomp | Strict syscall filtering |

**Blocked Operations:**

```
- fork bombs (PID limit)
- disk fills (quota)
- network access (disabled)
- privilege escalation (no setuid, capabilities dropped)
- container escape (gVisor syscall filtering)
- host filesystem access (isolated namespace)
```

### Network Capability Security

**Controls:**

| Control | Description |
|---------|-------------|
| Domain allowlist | Only approved domains accessible |
| Domain denylist | Blocked domains (internal, malicious) |
| Rate limiting | Requests per time window |
| TLS enforcement | Can require TLS for all connections |
| DNS filtering | Custom DNS resolution |
| Egress proxy | Route through proxy for inspection |

**Default Policy (restricted):**

```yaml
net:
  allowlist:
    - "*.wikipedia.org"
    - "*.github.com"
  denylist:
    - "*.internal.company.com"
    - "10.*"
    - "172.16.*"
    - "192.168.*"
    - "metadata.google.internal"      # Cloud metadata
    - "169.254.169.254"               # AWS metadata
```

### Filesystem Capability Security

**Controls:**

| Control | Description |
|---------|-------------|
| Mount isolation | Virtual filesystem per agent |
| Quotas | Storage limits per mount |
| Access modes | Read-only or read-write |
| Path restrictions | Allowed paths only |
| No device files | /dev restricted |
| No special mounts | /proc, /sys restricted |

**Blocked Paths:**

```
/etc/passwd, /etc/shadow          # System credentials
/root, /home/*                    # User directories
/proc, /sys                       # Kernel interfaces
/dev (except /dev/null, /dev/urandom)
```

### Secrets Capability Security

**Controls:**

| Control | Description |
|---------|-------------|
| Vault integration | Secrets fetched at runtime |
| No disk storage | Secrets never written to disk |
| Memory encryption | Secrets encrypted in memory |
| Audit logging | All secret access logged |
| Rotation | Automatic secret rotation |
| Scope | Secrets scoped to specific agents |

---

## Authentication & Authorization

### Authentication Methods

#### API Keys

```
X-API-Key: spawn_sk_live_abc123...
```

**Properties:**
- Prefixed for environment identification (`sk_live_`, `sk_test_`)
- 256-bit entropy
- Hashicorp-style checksums
- Revocable

#### JWT Tokens

```
Authorization: Bearer eyJhbGciOiJSUzI1NiIs...
```

**Properties:**
- RS256 or ES256 signing
- Short-lived (1 hour default)
- Refresh token rotation
- Claims-based authorization

#### mTLS (Enterprise)

```yaml
server:
  tls:
    enabled: true
    clientAuth: require
    clientCA: /etc/spawn/ca.crt
```

### Role-Based Access Control (RBAC)

```yaml
# roles.yaml
apiVersion: spawn.dev/v1
kind: Role
metadata:
  name: agent-operator
  namespace: production
rules:
  - resources: [agents]
    verbs: [get, list, create, start, stop]
  - resources: [agents/logs]
    verbs: [get]
  - resources: [tasks]
    verbs: [get, list, create]
---
apiVersion: spawn.dev/v1
kind: RoleBinding
metadata:
  name: alice-operator
  namespace: production
subjects:
  - kind: User
    name: alice@company.com
roleRef:
  kind: Role
  name: agent-operator
```

### Built-in Roles

| Role | Permissions |
|------|-------------|
| `admin` | Full access to all resources |
| `operator` | Manage agents, view logs, no config changes |
| `developer` | Create/run agents in dev namespace |
| `viewer` | Read-only access to all resources |
| `auditor` | Read access to logs, traces, audit events |

---

## Network Security

### TLS Configuration

All external communication uses TLS 1.3:

```yaml
server:
  tls:
    enabled: true
    minVersion: "1.3"
    cert: /etc/spawn/tls/server.crt
    key: /etc/spawn/tls/server.key
    cipherSuites:
      - TLS_AES_256_GCM_SHA384
      - TLS_CHACHA20_POLY1305_SHA256
```

### Service-to-Service mTLS

Internal communication uses mutual TLS:

```yaml
mesh:
  tls:
    enabled: true
    cert: /etc/spawn/tls/mesh.crt
    key: /etc/spawn/tls/mesh.key
    ca: /etc/spawn/tls/ca.crt
```

### Network Policies

Agent network isolation:

```
┌─────────────────────────────────────────────────────────────┐
│                      SPAWN NETWORK                          │
│                                                             │
│  ┌─────────────┐     ┌─────────────┐     ┌─────────────┐   │
│  │   Agent A   │     │   Agent B   │     │   Agent C   │   │
│  │             │     │             │     │             │   │
│  │ net: egress │     │ net: none   │     │ net: full   │   │
│  └──────┬──────┘     └─────────────┘     └──────┬──────┘   │
│         │                                       │           │
│         │           ┌─────────────┐             │           │
│         └──────────►│   Egress    │◄────────────┘           │
│                     │   Proxy     │                         │
│                     └──────┬──────┘                         │
│                            │                                │
└────────────────────────────┼────────────────────────────────┘
                             │
                             ▼
                      ┌─────────────┐
                      │  Internet   │
                      └─────────────┘
```

### Policy Types

| Policy | Description |
|--------|-------------|
| `none` | No network access |
| `restricted` | Allowlist-only egress, no ingress |
| `egress-only` | All egress allowed, no ingress |
| `full` | Full network access (dangerous) |

---

## Audit & Compliance

### Audit Logging

All security-relevant events are logged:

```json
{
  "timestamp": "2025-01-15T10:30:00Z",
  "eventType": "agent.capability.invoke",
  "principal": {
    "type": "agent",
    "id": "ag_abc123",
    "name": "researcher"
  },
  "resource": {
    "type": "capability",
    "name": "exec"
  },
  "action": "execute",
  "outcome": "success",
  "details": {
    "language": "python",
    "codeHash": "sha256:abc123...",
    "duration": "2.5s"
  },
  "sourceIP": "10.0.0.5",
  "requestId": "req_xyz789"
}
```

### Audit Event Types

| Event Type | Description |
|------------|-------------|
| `auth.login` | Authentication attempt |
| `auth.logout` | Session termination |
| `agent.create` | Agent created |
| `agent.delete` | Agent deleted |
| `agent.start` | Agent started |
| `agent.stop` | Agent stopped |
| `agent.capability.invoke` | Capability used |
| `agent.tool.invoke` | Tool invoked |
| `agent.secret.access` | Secret accessed |
| `mesh.message.send` | Inter-agent message |
| `rbac.permission.denied` | Authorization failure |

### Compliance Frameworks

| Framework | Support |
|-----------|---------|
| SOC 2 Type II | Enterprise |
| HIPAA | Enterprise |
| GDPR | Built-in |
| FedRAMP | Enterprise |
| PCI DSS | Enterprise |

### Data Residency

Configure data residency requirements:

```yaml
storage:
  dataResidency:
    enabled: true
    regions:
      - us-east-1
      - eu-west-1
    encryption:
      atRest: true
      inTransit: true
      keyManagement: aws-kms
```

---

## Secrets Management

### Secret Lifecycle

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Secret    │────►│    Vault    │────►│   spawn     │
│   Created   │     │   Stored    │     │   Fetches   │
└─────────────┘     └─────────────┘     └──────┬──────┘
                                               │
                                               ▼
                                        ┌─────────────┐
                                        │   Agent     │
                                        │  (Memory)   │
                                        └──────┬──────┘
                                               │
                                               ▼
                                        ┌─────────────┐
                                        │   Audit     │
                                        │    Log      │
                                        └─────────────┘
```

### Secret Storage

Secrets are **never**:
- Written to disk in the agent sandbox
- Logged in plaintext
- Passed as command-line arguments
- Stored in environment dumps

Secrets are:
- Fetched at runtime from Vault
- Injected as environment variables
- Encrypted in memory (where supported)
- Rotated automatically

### Vault Integration

```yaml
capabilities:
  secrets:
    vault:
      address: https://vault.example.com
      auth:
        method: kubernetes
        role: spawn-agent
        mountPath: auth/kubernetes
      tls:
        ca: /etc/spawn/vault-ca.crt
```

---

## Incident Response

### Security Events

spawn emits security events for:

| Event | Severity | Response |
|-------|----------|----------|
| Sandbox escape attempt | Critical | Terminate agent, alert |
| Capability abuse | High | Rate limit, audit |
| Authentication failure | Medium | Log, possible lockout |
| Policy violation | Medium | Block, log |
| Resource exhaustion | Low | Throttle, warn |

### Automated Response

```yaml
security:
  automation:
    rules:
      - name: sandbox-escape-detection
        condition: event.type == "sandbox.escape.attempt"
        actions:
          - terminate_agent
          - notify_security_team
          - create_incident
      
      - name: auth-brute-force
        condition: count(event.type == "auth.failure" AND event.principal == principal) > 5 IN 5m
        actions:
          - block_principal
          - notify_security_team
```

### Forensics

Decision replay for post-incident analysis:

```bash
# Replay agent decisions
spawn replay tr_abc123 --step-by-step

# Export trace for analysis
spawn trace export tr_abc123 --format json > trace.json

# Analyze capability usage
spawn audit capabilities --agent ag_abc123 --since 24h
```

---

## Security Checklist

### Production Deployment

- [ ] TLS enabled for all endpoints
- [ ] mTLS for service-to-service communication
- [ ] API keys rotated regularly
- [ ] Audit logging enabled and forwarded
- [ ] Network policies configured
- [ ] Secrets in Vault (not environment variables)
- [ ] gVisor or Firecracker runtime (not Docker)
- [ ] Resource limits configured
- [ ] Cost limits configured
- [ ] RBAC policies defined
- [ ] Security events monitored
- [ ] Incident response plan documented

### Agent Configuration

- [ ] Minimum required capabilities only
- [ ] Network allowlist defined
- [ ] Secrets injected (not hardcoded)
- [ ] Resource limits set
- [ ] Timeout configured
- [ ] Sandbox runtime specified
- [ ] Seccomp profile set to strict

---

## Vulnerability Disclosure

Report security vulnerabilities to: security@spawn.dev

**PGP Key:** Available at https://spawn.dev/.well-known/security.txt

**Bug Bounty:** https://hackerone.com/spawn (Enterprise customers)
