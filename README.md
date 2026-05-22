# K8s-Agent

A conversational Kubernetes CLI tool powered by LLM with natural language interface.

## Features

- **Natural language interface** - Query and manage Kubernetes clusters using plain English
- **Multi-cluster support** - Manage multiple Kubernetes clusters seamlessly
- **Function calling** - LLM calls Kubernetes APIs directly via function definitions
- **Interactive TUI** - Bubble Tea-based terminal interface with streaming responses
- **Table format output** - kubectl-style compact table format, reduces LLM context usage
- **Change management** - Structured workflow for planning, reviewing, and executing Kubernetes changes with rollback support

## Installation

```bash
go install ./cmd/k8s-agent
```

Or build from source:

```bash
git clone git@github.com:threestoneliu/k8s-agent.git
cd k8s-agent
go build -o k8s-agent ./cmd/k8s-agent
```

## Quick Start

### 1. Configure your clusters

```bash
# Add a cluster (copies kubeconfig entry to ~/.config/k8s-agent/kubeconfigs/)
k8s-agent cluster add dev ~/.kube/config

# Set default cluster
k8s-agent cluster use dev
```

### 2. Start chat mode

```bash
k8s-agent chat
```

### 3. Query in natural language

```
> what pods are running?
> list services in default namespace
> show me the nodes
> get deployment nginx
```

### 4. Manage resources

```
> delete pod nginx
> scale deployment web to 3 replicas
> create configmap app-config from file config.yaml
```

## Commands

### Chat Mode

```bash
k8s-agent chat
```

### Cluster Management

```bash
k8s-agent cluster list
k8s-agent cluster add <name> <kubeconfig-path>
k8s-agent cluster use <name>
k8s-agent cluster remove <name>
```

### Available Functions

The LLM has access to these functions:

| Function | Description |
|----------|-------------|
| `resource_list` | List Kubernetes resources with table format output |
| `resource_get` | Get details of a specific resource |
| `get_apiresources` | List all supported API resource types |
| `use_cluster` | Switch to a different cluster |

## Configuration

Create `~/.config/k8s-agent/config.yaml`:

```yaml
current-cluster: "dev"

llm:
  provider: "openai"
  api-key: "${OPENAI_API_KEY}"  # or set directly
  model: "gpt-4"

context:
  max-messages: 20
  max-tokens: 8000
  summary-enabled: true

session:
  storage-path: "~/.config/k8s-agent/sessions"
  max-cache-size: 100
  max-sessions: 10

logging:
  level: "info"
  format: "text"
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `OPENAI_API_KEY` | OpenAI API key |
| `K8S_AGENT_CONFIG` | Path to config file |

## Architecture

```
User Input (TUI)
    ↓
┌─────────────────────────────────────────┐
│  pkg/agent - Agent loop                  │
│  - Handles /clusters, /cluster, /config │
│  - Manages session context               │
│  - Streams LLM responses                 │
└─────────────────────────────────────────┘
    ↓
┌─────────────────────────────────────────┐
│  pkg/llm - LLM Integration             │
│  - OpenAI SDK integration               │
│  - Function definitions & handlers       │
│  - Auto-registers resource functions    │
└─────────────────────────────────────────┘
    ↓
┌─────────────────────────────────────────┐
│  pkg/k8s - Kubernetes Executor          │
│  - RESTClient with Table Accept header   │
│  - Resource discovery & caching          │
│  - List/Get with selectors              │
└─────────────────────────────────────────┘
    ↓
Kubernetes API Server
```

### Table Format Output

`resource_list` returns kubectl-style compact table format instead of full YAML/JSON:

```
NAME                     READY   STATUS    AGE
nginx-6799fc88d8-abc12    1/1     Running   5d
redis-7d8f9f6e4-xyz34     1/1     Running   10d
```

This reduces LLM context usage by ~90% for large resource lists.

### Change Management Workflow

k8s-agent provides a structured workflow for managing Kubernetes changes:

1. **Parse** - Natural language input is parsed into structured intent
2. **Clarify** - If intent is incomplete, user is asked for clarification
3. **Plan** - System generates an execution plan with steps and risk assessment
4. **Review** - User reviews the plan before it executes
5. **Execute** - Approved changes are executed step-by-step
6. **Rollback** - If something goes wrong, changes can be rolled back

The workflow includes:
- **Risk assessment** - Operations are classified as LOW, MEDIUM, HIGH, or CRITICAL risk
- **Pre-checks** - Validation before execution (resource existence, quota, permissions)
- **Snapshot/rollback** - State capture before mutations for potential rollback
- **Audit logging** - All state transitions and actions are logged

## Skills

k8s-agent supports Skills for standardized workflows. Skills are stored in `~/.config/k8s-agent/skills/`.

### Using Skills

Skills are discovered automatically through the System Prompt. When your request matches a Skill's description, the LLM will read the Skill's SKILL.md and follow its workflow.

**Explicit trigger:**
```
> /skill k8s-inspection
> 用巡检skill
```

**Implicit trigger:**
```
> 巡检一下
> 检查集群健康
```

### Creating Custom Skills

Create a directory at `~/.config/k8s-agent/skills/<skill-name>/SKILL.md`:

```yaml
---
name: my-skill
description: Description of when this skill applies
license: Apache-2.0
---

# My Skill

## Workflows

1. Step 1: call resource_list(resource="...")
2. Step 2: analyze results
3. Step 3: generate output
```

### Available Skills

- `k8s-inspection` - Cluster health inspection workflow (nodes → pods → events → report)

## Development

```bash
# Run tests
go test ./...

# Build
go build -o k8s-agent ./cmd/k8s-agent

# Run with debug logging
K8S_AGENT_LOG_LEVEL=debug ./k8s-agent chat
```

## License

Apache License 2.0 - See [LICENSE](LICENSE) file