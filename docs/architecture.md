# Architecture

## System Overview

k8s-agent is a CLI tool that provides a conversational interface for Kubernetes operations. It combines natural language parsing with operation classification and human-in-the-loop confirmation for safe mutations.

## Components

```
┌─────────────────────────────────────────────────────────────────┐
│                         CLI Layer                               │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐   │
│  │   chat  │ │   get   │ │  list   │ │ delete  │ │  task   │   │
│  └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘ └────┬────┘   │
└───────┼───────────┼───────────┼───────────┼───────────┼────────┘
        │           │           │           │           │
        v           v           v           v           v
┌─────────────────────────────────────────────────────────────────┐
│                      Command Handlers                           │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  RootCommand (session, executor, confirmMgr, scheduler)   │   │
│  └──────────────────────────────────────────────────────────┘   │
└────────────────────────────┬────────────────────────────────────┘
                             │
        ┌────────────────────┼────────────────────┐
        v                    v                    v
┌──────────────┐    ┌────────────────┐    ┌─────────────────┐
│ Session Mgr  │    │ Engine         │    │ Confirmation Mgr│
│              │    │                │    │                 │
│ - conv ID    │    │ - Parse()      │    │ - Create()      │
│ - messages   │    │ - Classify()   │    │ - Approve()     │
│ - context    │    │ - Execute()    │    │ - Get()         │
└──────────────┘    └───────┬────────┘    └─────────────────┘
                           │
           ┌───────────────┼───────────────┐
           v               v               v
    ┌───────────┐   ┌───────────┐   ┌───────────────┐
    │  Parser   │   │ Classifier│   │  Executor     │
    │           │   │           │   │               │
    │ - Parse() │   │ - Query   │   │ - Query Ops   │
    │           │   │ - Mutation│   │ - Mutation    │
    └───────────┘   └───────────┘   └───────┬───────┘
                                            │
                    ┌───────────────────────┼───────────────────────┐
                    v                       v                       v
              ┌───────────┐         ┌───────────┐         ┌───────────┐
              │  Query    │         │ Mutation  │         │  Cluster  │
              │  Handler  │         │  Handler  │         │  Registry │
              └───────────┘         └───────────┘         └───────────┘
```

## Component Descriptions

### CLI Layer (`cmd/cli/`)

The CLI layer provides user-facing commands using Cobra framework:

- `root.go` - Root command and dependency initialization
- `chat.go` - Interactive chat mode
- `get.go`, `list.go`, `describe.go` - Query commands
- `delete.go`, `create.go`, `scale.go` - Mutation commands
- `cluster.go` - Multi-cluster management
- `task.go` - Scheduled task management
- `confirm.go` - Confirmation handling

### Session Manager (`pkg/session/`)

Manages conversation state and history:

- `Manager` - In-memory session storage
- `Conversation` - Contains messages and context
- `Message` - User/Assistant message records

### Engine (`pkg/engine/`)

Core parsing and execution logic:

- `Parser` - Parses natural language input into structured operations
- `Classifier` - Classifies operations as Query or Mutation
- `Executor` - Executes operations against Kubernetes clusters
- `Mapper` - Maps resource aliases to canonical names

### Confirmation Manager (`pkg/confirmation/`)

Handles human-in-the-loop approvals:

- `Manager` - Stores pending confirmations with TTL
- `PendingOperation` - Represents an operation awaiting approval
- Key generation with secure random

### Cluster Registry (`pkg/cluster/`)

Manages Kubernetes cluster configurations:

- `Registry` - Stores cluster kubeconfigs
- `ClusterConfig` - Individual cluster configuration

### Scheduler (`pkg/scheduler/`)

Manages scheduled inspection tasks:

- `Manager` - Task scheduling and execution
- `ScheduledTask` - Task definition with cron expression
- `TaskResult` - Execution results storage

## Data Flow

### Query Operation Flow

```
User: "get pods"
  │
  v
Parser.Parse() ──► ParsedOperation
  │
  v
Classifier.Classify() ──► ClassifiedOperation (OperationTypeQuery)
  │
  v
Executor.ExecuteQuery() ──► ExecutionResult
  │
  v
Output to user
```

### Mutation Operation Flow

```
User: "delete pod nginx"
  │
  v
Parser.Parse() ──► ParsedOperation
  │
  v
Classifier.Classify() ──► ClassifiedOperation (OperationTypeMutation)
  │
  v
ConfirmationManager.CreateConfirmation() ──► confirmKey
  │
  v
"Confirmation required: XYZ" ──► User
  │
  v
User: "k8s-agent confirm XYZ"
  │
  v
ConfirmationManager.ApproveConfirmation()
  │
  v
Executor.ExecuteMutation() ──► ExecutionResult
  │
  v
Output to user
```

## Key Interfaces

### ClassifiedOperation

```go
type ClassifiedOperation struct {
    Type      OperationType  // Query or Mutation
    Verb      string         // get, list, delete, etc.
    Resource  string         // pods, services, etc.
    Name      string         // specific resource name
    Namespace string         // target namespace
    Flags     map[string]string
    RawInput  string
}
```

### ExecutionResult

```go
type ExecutionResult struct {
    Success  bool
    Output   string
    Resource string
    Verb     string
    Error    error
}
```

## Design Decisions

### 1. In-Memory Session Storage
Sessions are stored in-memory for simplicity. This means:
- Sessions don't persist across restarts
- Each chat session starts fresh
- Future: Could add persistence layer

### 2. Confirmation TTL
Confirmations expire after a configurable TTL (default 5 minutes):
- Prevents unclaimed confirmations from accumulating
- Security measure for accidental confirmation keys
- Auto-cleanup via background routine

### 3. Resource Alias Mapping
The mapper handles common aliases:
- `svc` → `services`
- `cm` → `configmaps`
- `ns` → `namespaces`
- etc.

### 4. Operation Classification
Operations are classified as:
- **Query**: Read-only operations that execute immediately
- **Mutation**: Operations that modify resources and require confirmation

## Testing Strategy

- **Unit Tests**: Parser, Classifier, Mapper, Confirmation logic
- **Integration Tests**: Executor with fake Kubernetes client
- **E2E Tests**: Critical flows (query, mutation with confirmation, chat mode)

See [operation_classification.md](operation_classification.md) for detailed classification rules.