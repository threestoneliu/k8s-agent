# Architecture

## System Overview

k8s-agent is a CLI tool that provides a conversational interface for Kubernetes operations. It combines natural language parsing with operation classification and human-in-the-loop confirmation for safe mutations. The Core change management system adds a structured workflow for planning, reviewing, and executing Kubernetes changes.

## Core Change Management System

The Core change management system implements a state machine for managing Kubernetes change operations. It ensures changes are properly planned, reviewed, and can be rolled back if something goes wrong.

### State Machine

The change workflow follows a structured state machine:

```
                    ┌─────────────┐
                    │  PARSING    │
                    └──────┬──────┘
                           │
           ┌───────────────┼───────────────┐
           │               │               │
           v               v               v
    ┌──────────┐    ┌───────────┐    ┌──────────┐
    │ CLARIFY  │    │ PLANNING  │    │  FAILED  │
    └────┬─────┘    └─────┬─────┘    └──────────┘
         │                │
         │                v
         │         ┌───────────┐
         │         │ REVIEWING│
         │         └─────┬─────┘
         │               │
         │               v
         │         ┌───────────┐
         └────────►│ EXECUTING │
                   └─────┬─────┘
                         │
              ┌──────────┴──────────┐
              │                     │
              v                     v
        ┌──────────┐          ┌──────────┐
        │COMPLETED │          │  FAILED  │
        └──────────┘          └──────────┘
```

**States:**
- `PARSING` - Parsing natural language input into structured intent
- `CLARIFYING` - Requesting clarification from the user for incomplete intent
- `PLANNING` - Generating execution plan with steps and risk assessment
- `REVIEWING` - Presenting plan to user for review and confirmation
- `EXECUTING` - Executing approved changes step by step
- `COMPLETED` - Change successfully completed
- `FAILED` - Change failed or was aborted

**Signals:**
- `SignalConfirm` - Approve current state and proceed
- `SignalModify` - Request modification (return to previous state)
- `SignalAbort` - Abort the operation
- `SignalProceed` - Force progression to next state

## Components

### pkg/core/

The core package contains the change management state machine and related logic:

- `session.go` - ChangeSession struct and session lifecycle
- `state.go` - State machine transition logic
- `validator.go` - Intent validation and risk assessment
- `planner.go` - ChangePlan generation
- `executor.go` - Plan execution with pre-checks
- `diff.go` - Resource diff calculation
- `rollback.go` - Snapshot and rollback functionality
- `audit.go` - Audit logging for all state transitions

### pkg/agent/

The agent package handles interaction between UI and core:

- `translator.go` - Converts ChangePlan and ResourceDiff to human-readable text

### pkg/agent/intent/

- `parser.go` - Parses user input into ParsedIntent structures

### pkg/ipc/

- `ipc.go` - Input/Output structures for UI-Agent communication

## Data Structures

### ChangeSession

Represents a single change management session:

```go
type ChangeSession struct {
    ID    string  // Unique session identifier
    State State   // Current position in state machine
}
```

### ParsedIntent

Represents parsed user intent:

```go
type ParsedIntent struct {
    Action    Action           // CREATE, UPDATE, DELETE, INSPECT
    Target    ResourceTarget   // Target Kubernetes resource
    Params    map[string]interface{}
    RiskLevel RiskLevel        // LOW, MEDIUM, HIGH, CRITICAL
    Reason    string           // Justification for high-risk operations
}
```

### ChangePlan

Generated from ParsedIntent:

```go
type ChangePlan struct {
    ID           string
    Summary      string
    Steps        []ChangeStep   // Ordered execution steps
    PreCheck     []string       // Pre-execution validation checks
    RollbackPlan []ChangeStep   // Steps to undo the change
    RiskLevel    RiskLevel
    Impact       string
    Duration     time.Duration
}
```

### ResourceDiff

Shows differences between current and desired state:

```go
type ResourceDiff struct {
    HasChanges    bool
    ChangedFields []string
    OldValues     map[string]interface{}
    NewValues     map[string]interface{}
}
```

### Snapshot

Point-in-time capture of a resource for rollback:

```go
type Snapshot struct {
    ID         string
    SessionID  string
    ResourceID ResourceID
    Object     *unstructured.Unstructured
    CreatedAt  time.Time
}
```

## Risk Assessment

Risk levels are determined by:

1. **Action type**: DELETE > UPDATE > CREATE > INSPECT
2. **Resource kind**: Critical resources (Node, PV, Namespace, etc.) increase risk
3. **Scope**: Cluster-scoped resources have higher risk than namespaced

| Action    | Default Risk |
|-----------|--------------|
| INSPECT   | LOW          |
| CREATE    | LOW          |
| UPDATE    | MEDIUM       |
| DELETE    | HIGH         |

## Pre-checks

Before execution, the following pre-checks run:

- `resource_exists` - Verifies resource existence for non-CREATE operations
- `sufficient_quota` - Checks namespace resource quota
- `no_conflicting_name` - Ensures no naming conflicts
- `backup_snapshot` - Verifies snapshots exist for rollback

## Audit Log

All state transitions and significant actions are logged:

```go
type AuditEntry struct {
    SessionID string
    Timestamp time.Time
    Action    string    // e.g., "state_transition", "execute_start"
    Actor     string    // e.g., "session", "executor"
    Details   map[string]interface{}
}
```

## Snapshot and Rollback

The rollback system:

1. Creates snapshots before mutations (UPDATE, DELETE)
2. Stores snapshots in memory indexed by session and resource
3. Enables rollback to previous state if something goes wrong
4. Cleans up snapshots when session closes

## Integration

The Core change management system integrates with:

- **UI Layer** - Receives user signals and displays state changes
- **LLM Integration** - Parses natural language into ParsedIntent
- **Kubernetes API** - Executes planned changes (placeholder in current implementation)

## Future Enhancements

- K8s API integration for actual resource operations
- Persistent storage for sessions and snapshots
- LLM-powered intent parsing and clarification
- Detailed step-by-step execution with progress tracking