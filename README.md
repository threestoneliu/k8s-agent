# K8s-Agent

A conversational Kubernetes agent with human-in-the-loop approvals.

## Features

- Multi-cluster support - manage multiple Kubernetes clusters
- Natural language interface - query and mutate resources using simple commands
- Human-in-the-loop for mutations - destructive operations require explicit confirmation
- Scheduled inspection tasks - automate recurring cluster inspections

## Installation

```bash
go install ./cmd/k8s-agent
```

Or build from source:

```bash
git clone https://github.com/your-org/k8s-agent.git
cd k8s-agent
go build -o k8s-agent ./cmd/k8s-agent
```

## Quick Start

### 1. Add a cluster

```bash
k8s-agent cluster add dev ~/.kube/config
k8s-agent cluster use dev
```

### 2. Start chat mode

```bash
k8s-agent chat
```

### 3. Execute commands

**Query operations (direct execution):**
```
> get pods
> list services
> describe pod nginx
```

**Mutation operations (require confirmation):**
```
> delete pod nginx
Confirmation required: ABC123
Use 'k8s-agent confirm ABC123' to approve
```

### 4. Approve mutations

```bash
k8s-agent confirm ABC123
```

## Commands

### Chat Mode
```
k8s-agent chat
```

### Query Operations
```
k8s-agent get <resource> [name] [-n namespace]
k8s-agent list <resource> [-n namespace]
k8s-agent describe <resource> <name> [-n namespace]
```

### Mutation Operations
```
k8s-agent delete <resource> <name> [-n namespace]
k8s-agent create <resource> <name> [--flags]
k8s-agent scale <resource> <name> --replicas=N
```

### Cluster Management
```
k8s-agent cluster list
k8s-agent cluster add <name> <kubeconfig-path>
k8s-agent cluster use <name>
k8s-agent cluster remove <name>
```

### Task Management
```
k8s-agent task list
k8s-agent task create <name> <cron> <operation>
k8s-agent task delete <id>
k8s-agent task run <id>
k8s-agent task results <id>
```

### Confirmation
```
k8s-agent confirm <key>
```

## Configuration

- **Kubeconfig**: `~/.kube/config` or specify path via `cluster add`
- **Confirmation TTL**: 5 minutes (configurable)
- **Session history**: Stored in memory (current session only)

## Operation Classification

### Query Operations (Direct Execution)
These operations are read-only and execute immediately:
- `get` - Retrieve specific resource
- `list` - List resources
- `describe` - Show resource details
- `watch` - Watch for changes
- `logs` - View pod logs
- `exec` - Execute command in pod

### Mutation Operations (Require Confirmation)
These operations modify resources and require explicit approval:
- `create` - Create new resources
- `update` - Update existing resources
- `patch` - Patch resources
- `delete` - Delete resources
- `scale` - Scale deployments
- `cordon` / `uncordon` - Node maintenance
- `drain` - Drain nodes

### High-Risk Resources
Operations on these resources always require confirmation regardless of verb:
- `nodes`
- `persistentvolumes`
- `namespaces`
- `storageclasses`

## Examples

### Query pods in default namespace
```
k8s-agent get pods
```

### Query pods in specific namespace
```
k8s-agent get pods -n kube-system
```

### Describe a specific pod
```
k8s-agent describe pod nginx
```

### Delete a pod (requires confirmation)
```
k8s-agent delete pod nginx
# Output: Confirmation key: XYZ789
# Then: k8s-agent confirm XYZ789
```

### Scale a deployment
```
k8s-agent scale deployment myapp --replicas=5
```

### Create a scheduled task
```
k8s-agent task create daily-pod-check "0 8 * * *" "get pods"
```

## Architecture

See [docs/architecture.md](docs/architecture.md) for detailed system design.

## License

MIT