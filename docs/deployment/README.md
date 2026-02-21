# Deployment Guide

This guide covers deploying spawn in various environments, from development to production.

## Deployment Options

| Method | Use Case | Complexity |
|--------|----------|------------|
| Binary | Development, single node | Low |
| Docker | Development, simple production | Low |
| Docker Compose | Multi-service development | Medium |
| Kubernetes | Production, scalable | High |
| Managed (Cloud) | Production, zero-ops | Low |

---

## Binary Installation

### Download

```bash
# Linux (amd64)
curl -sSL https://github.com/spawndev/spawn/releases/latest/download/spawn-linux-amd64.tar.gz | tar xz
sudo mv spawn spawnd /usr/local/bin/

# Linux (arm64)
curl -sSL https://github.com/spawndev/spawn/releases/latest/download/spawn-linux-arm64.tar.gz | tar xz
sudo mv spawn spawnd /usr/local/bin/

# macOS (Apple Silicon)
curl -sSL https://github.com/spawndev/spawn/releases/latest/download/spawn-darwin-arm64.tar.gz | tar xz
sudo mv spawn spawnd /usr/local/bin/

# macOS (Intel)
curl -sSL https://github.com/spawndev/spawn/releases/latest/download/spawn-darwin-amd64.tar.gz | tar xz
sudo mv spawn spawnd /usr/local/bin/
```

### Install Script

```bash
curl -sSL https://spawn.dev/install | sh
```

### Verify Installation

```bash
spawn version
# spawn version 1.0.0 (abc123) built 2025-01-15

spawn doctor
# ✓ spawn binary installed
# ✓ spawnd binary installed
# ✓ gVisor (runsc) available
# ✓ Docker available
# ✓ Configuration valid
```

### Run Daemon

```bash
# Foreground (development)
spawnd --config /etc/spawn/spawn.yaml

# Background (systemd)
sudo systemctl enable spawnd
sudo systemctl start spawnd
```

### Systemd Service

```ini
# /etc/systemd/system/spawnd.service
[Unit]
Description=spawn Agent Daemon
After=network.target

[Service]
Type=simple
User=spawn
Group=spawn
ExecStart=/usr/local/bin/spawnd --config /etc/spawn/spawn.yaml
Restart=always
RestartSec=5
LimitNOFILE=65535
LimitNPROC=65535

# Security
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/spawn /var/log/spawn

[Install]
WantedBy=multi-user.target
```

---

## Docker Deployment

### Quick Start

```bash
# Run spawnd
docker run -d \
  --name spawnd \
  -p 8080:8080 \
  -p 9090:9090 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v spawn-data:/var/lib/spawn \
  -e ANTHROPIC_API_KEY=$ANTHROPIC_API_KEY \
  ghcr.io/spawndev/spawn:latest

# Verify
docker logs spawnd
curl http://localhost:8080/health
```

### Docker Compose

```yaml
# docker-compose.yml
version: '3.8'

services:
  spawnd:
    image: ghcr.io/spawndev/spawn:latest
    container_name: spawnd
    restart: unless-stopped
    ports:
      - "8080:8080"    # REST API
      - "9090:9090"    # gRPC API
      - "9091:9091"    # Metrics
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      - spawn-data:/var/lib/spawn
      - ./spawn.yaml:/etc/spawn/spawn.yaml:ro
    environment:
      - ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY}
      - OPENAI_API_KEY=${OPENAI_API_KEY}
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  # Optional: Postgres for state storage
  postgres:
    image: postgres:16-alpine
    container_name: spawn-postgres
    restart: unless-stopped
    volumes:
      - postgres-data:/var/lib/postgresql/data
    environment:
      - POSTGRES_DB=spawn
      - POSTGRES_USER=spawn
      - POSTGRES_PASSWORD=${POSTGRES_PASSWORD}

  # Optional: NATS for mesh
  nats:
    image: nats:2.10-alpine
    container_name: spawn-nats
    restart: unless-stopped
    ports:
      - "4222:4222"
    command: ["--cluster_name", "spawn", "--js"]

  # Optional: Jaeger for tracing
  jaeger:
    image: jaegertracing/all-in-one:1.53
    container_name: spawn-jaeger
    ports:
      - "16686:16686"   # UI
      - "4317:4317"     # OTLP gRPC

volumes:
  spawn-data:
  postgres-data:
```

### Start Services

```bash
# Create .env file
cat > .env << EOF
ANTHROPIC_API_KEY=sk-ant-...
OPENAI_API_KEY=sk-...
POSTGRES_PASSWORD=$(openssl rand -base64 32)
EOF

# Start
docker-compose up -d

# Check status
docker-compose ps
docker-compose logs -f spawnd
```

---

## Kubernetes Deployment

### Prerequisites

- Kubernetes 1.28+
- kubectl configured
- Helm 3.x (optional)
- gVisor installed on nodes (for gVisor runtime)

### Namespace

```yaml
# namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: spawn
  labels:
    app.kubernetes.io/name: spawn
```

### ConfigMap

```yaml
# configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: spawn-config
  namespace: spawn
data:
  spawn.yaml: |
    server:
      host: 0.0.0.0
      ports:
        grpc: 9090
        rest: 8080
        metrics: 9091
    
    storage:
      state:
        driver: postgres
        dsn: postgres://spawn:$(POSTGRES_PASSWORD)@postgres:5432/spawn?sslmode=require
      
    sandbox:
      defaultRuntime: gvisor
    
    mesh:
      enabled: true
      backend: nats
      nats:
        url: nats://nats:4222
    
    observability:
      traces:
        enabled: true
        exporter: otlp
        endpoint: jaeger-collector:4317
```

### Secrets

```yaml
# secrets.yaml
apiVersion: v1
kind: Secret
metadata:
  name: spawn-secrets
  namespace: spawn
type: Opaque
stringData:
  ANTHROPIC_API_KEY: "sk-ant-..."
  OPENAI_API_KEY: "sk-..."
  POSTGRES_PASSWORD: "..."
```

### Deployment

```yaml
# deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: spawnd
  namespace: spawn
  labels:
    app: spawnd
spec:
  replicas: 3
  selector:
    matchLabels:
      app: spawnd
  template:
    metadata:
      labels:
        app: spawnd
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "9091"
    spec:
      serviceAccountName: spawnd
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
        fsGroup: 1000
      
      containers:
        - name: spawnd
          image: ghcr.io/spawndev/spawn:1.0.0
          imagePullPolicy: IfNotPresent
          
          ports:
            - name: grpc
              containerPort: 9090
            - name: rest
              containerPort: 8080
            - name: metrics
              containerPort: 9091
          
          envFrom:
            - secretRef:
                name: spawn-secrets
          
          volumeMounts:
            - name: config
              mountPath: /etc/spawn
              readOnly: true
            - name: data
              mountPath: /var/lib/spawn
          
          resources:
            requests:
              memory: "512Mi"
              cpu: "500m"
            limits:
              memory: "2Gi"
              cpu: "2000m"
          
          livenessProbe:
            httpGet:
              path: /health
              port: rest
            initialDelaySeconds: 10
            periodSeconds: 10
          
          readinessProbe:
            httpGet:
              path: /health
              port: rest
            initialDelaySeconds: 5
            periodSeconds: 5
      
      volumes:
        - name: config
          configMap:
            name: spawn-config
        - name: data
          emptyDir: {}
      
      affinity:
        podAntiAffinity:
          preferredDuringSchedulingIgnoredDuringExecution:
            - weight: 100
              podAffinityTerm:
                labelSelector:
                  matchLabels:
                    app: spawnd
                topologyKey: kubernetes.io/hostname
```

### Service

```yaml
# service.yaml
apiVersion: v1
kind: Service
metadata:
  name: spawnd
  namespace: spawn
spec:
  selector:
    app: spawnd
  ports:
    - name: grpc
      port: 9090
      targetPort: grpc
    - name: rest
      port: 8080
      targetPort: rest
    - name: metrics
      port: 9091
      targetPort: metrics
---
apiVersion: v1
kind: Service
metadata:
  name: spawnd-headless
  namespace: spawn
spec:
  clusterIP: None
  selector:
    app: spawnd
  ports:
    - name: grpc
      port: 9090
```

### Ingress

```yaml
# ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: spawnd
  namespace: spawn
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/backend-protocol: "GRPC"
spec:
  tls:
    - hosts:
        - spawn.example.com
      secretName: spawn-tls
  rules:
    - host: spawn.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: spawnd
                port:
                  number: 8080
```

### RBAC

```yaml
# rbac.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: spawnd
  namespace: spawn
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: spawnd
rules:
  - apiGroups: [""]
    resources: ["pods", "pods/exec", "pods/log"]
    verbs: ["get", "list", "watch", "create", "delete"]
  - apiGroups: [""]
    resources: ["secrets"]
    verbs: ["get"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: spawnd
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: spawnd
subjects:
  - kind: ServiceAccount
    name: spawnd
    namespace: spawn
```

### HPA

```yaml
# hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: spawnd
  namespace: spawn
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: spawnd
  minReplicas: 3
  maxReplicas: 20
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
    - type: Resource
      resource:
        name: memory
        target:
          type: Utilization
          averageUtilization: 80
```

### Apply All

```bash
kubectl apply -f namespace.yaml
kubectl apply -f secrets.yaml
kubectl apply -f configmap.yaml
kubectl apply -f rbac.yaml
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml
kubectl apply -f ingress.yaml
kubectl apply -f hpa.yaml

# Verify
kubectl get pods -n spawn
kubectl logs -n spawn -l app=spawnd -f
```

### Helm Chart

```bash
# Add repo
helm repo add spawn https://charts.spawn.dev
helm repo update

# Install
helm install spawn spawn/spawn \
  --namespace spawn \
  --create-namespace \
  --set anthropicApiKey=$ANTHROPIC_API_KEY \
  --set postgres.enabled=true \
  --set nats.enabled=true \
  --values values.yaml

# values.yaml
replicaCount: 3

image:
  repository: ghcr.io/spawndev/spawn
  tag: "1.0.0"

resources:
  requests:
    memory: 512Mi
    cpu: 500m
  limits:
    memory: 2Gi
    cpu: 2000m

sandbox:
  runtime: gvisor

postgres:
  enabled: true
  persistence:
    size: 100Gi

nats:
  enabled: true
  jetstream:
    enabled: true

ingress:
  enabled: true
  hostname: spawn.example.com
  tls: true
```

---

## Production Checklist

### Security

- [ ] TLS enabled for all endpoints
- [ ] API keys in Kubernetes Secrets or Vault
- [ ] Network policies configured
- [ ] Pod security policies/standards enforced
- [ ] RBAC policies defined
- [ ] Audit logging enabled

### High Availability

- [ ] Minimum 3 replicas
- [ ] Pod anti-affinity configured
- [ ] PodDisruptionBudget defined
- [ ] Multi-AZ deployment
- [ ] Database HA (Postgres with replication)
- [ ] NATS cluster (3+ nodes)

### Observability

- [ ] Prometheus metrics enabled
- [ ] Tracing configured (Jaeger/Tempo)
- [ ] Logs forwarded (Loki/CloudWatch)
- [ ] Dashboards created
- [ ] Alerts configured

### Performance

- [ ] Resource limits tuned
- [ ] HPA configured
- [ ] Connection pooling enabled
- [ ] Cache configured

### Backup & Recovery

- [ ] Database backups scheduled
- [ ] Backup verification tested
- [ ] Recovery procedure documented
- [ ] RTO/RPO defined

### Operations

- [ ] Runbooks documented
- [ ] On-call rotation defined
- [ ] Incident response plan
- [ ] Change management process

---

## Scaling

### Horizontal Scaling

spawn scales horizontally by adding more replicas:

```yaml
spec:
  replicas: 10
```

### Agent Scaling

Individual agents can scale:

```yaml
spec:
  scaling:
    minReplicas: 1
    maxReplicas: 50
    metrics:
      - type: queue-depth
        target: 10
```

### Database Scaling

For high throughput:

```yaml
storage:
  state:
    driver: postgres
    dsn: postgres://...
    pool:
      maxConnections: 100
      maxIdleConnections: 20
```

### NATS Scaling

```yaml
mesh:
  nats:
    cluster:
      enabled: true
      replicas: 3
```

---

## Monitoring

### Prometheus Metrics

Key metrics to monitor:

| Metric | Description |
|--------|-------------|
| `spawn_agents_total` | Total agents |
| `spawn_agents_running` | Running agents |
| `spawn_agent_tokens_total` | Tokens used |
| `spawn_agent_cost_usd_total` | Cost in USD |
| `spawn_capability_invocations_total` | Capability calls |
| `spawn_tool_invocations_total` | Tool calls |
| `spawn_mesh_messages_total` | Mesh messages |
| `spawn_api_requests_total` | API requests |
| `spawn_api_latency_seconds` | API latency |

### Grafana Dashboard

Import dashboard: `https://grafana.com/dashboards/12345`

### Alerting

```yaml
# prometheus-rules.yaml
groups:
  - name: spawn
    rules:
      - alert: SpawnDaemonDown
        expr: up{job="spawnd"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "spawn daemon is down"
      
      - alert: SpawnHighErrorRate
        expr: rate(spawn_api_requests_total{status="error"}[5m]) > 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High error rate detected"
      
      - alert: SpawnHighCost
        expr: sum(spawn_agent_cost_usd_total) > 1000
        for: 1h
        labels:
          severity: warning
        annotations:
          summary: "High LLM cost detected"
```

---

## Troubleshooting

### Common Issues

#### Daemon Won't Start

```bash
# Check logs
journalctl -u spawnd -f

# Verify config
spawn validate /etc/spawn/spawn.yaml

# Check permissions
ls -la /var/lib/spawn
```

#### Agents Failing to Start

```bash
# Check agent logs
spawn agent logs <agent-id>

# Check sandbox runtime
spawn doctor

# Verify gVisor
runsc --version
```

#### Network Issues

```bash
# Check connectivity
spawn mesh status

# Verify NATS
nats-server --version
nats pub test "hello"
```

### Debug Mode

```bash
# Enable debug logging
spawnd --log-level debug

# Trace specific agent
spawn run --trace agent.yaml
```
