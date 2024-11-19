package testutil

// AgentConfigYAML is a reusable test fixture.
const AgentConfigYAML = `apiVersion: spawn.dev/v1
kind: Agent
metadata:
  name: test-agent
spec:
  model:
    provider: anthropic
    name: claude-sonnet-4-20250514
  sandbox:
    runtime: gvisor
  capabilities:
    exec:
      enabled: true
`
