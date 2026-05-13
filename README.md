# K8s-Agent

A conversational Kubernetes CLI tool powered by LLM with natural language interface.

## Features

- **Natural language interface** - Query and manage Kubernetes clusters using plain English
- **Multi-cluster support** - Manage multiple Kubernetes clusters seamlessly
- **Function calling** - LLM calls Kubernetes APIs directly via function definitions
- **Interactive TUI** - Bubble Tea-based terminal interface with streaming responses

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
| `resource_list` | List Kubernetes resources with optional label/field selectors |
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
│  - Dynamic client for all resources    │
│  - Resource discovery & caching          │
│  - List/Get with selectors              │
└─────────────────────────────────────────┘
    ↓
Kubernetes API Server
```

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

MIT
