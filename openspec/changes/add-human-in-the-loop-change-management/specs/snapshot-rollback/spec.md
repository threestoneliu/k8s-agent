# Snapshot & Rollback Specification

## ADDED Requirements

### Requirement: Snapshot Before Change

The system SHALL create a snapshot before executing any change step.

#### Scenario: Snapshot Created Before Execution

Given a session in EXECUTING state with a ChangeStep for a resource
When the step is about to execute
Then a snapshot SHALL be created before the change is applied

### Requirement: Snapshot Contains Full Object State

The snapshot SHALL contain the complete resource object (*unstructured.Unstructured) for restoration.

#### Scenario: Snapshot Stores Complete Object

Given a Deployment resource with replicas=3, image=nginx:1.19
When a snapshot is created
Then the snapshot SHALL contain the full object including all fields

### Requirement: Rollback Restores Object State

Rollback SHALL restore the resource to its state at snapshot creation time.

#### Scenario: Rollback Restores Previous State

Given a snapshot with replicas=3 for a Deployment
When Rollback is executed
Then the Deployment SHALL be restored to replicas=3

### Requirement: Snapshot Isolation Per Session

Each session's snapshots SHALL be isolated from other sessions.

#### Scenario: Sessions Have Isolated Snapshots

Given session A with snapshots and session B with no snapshots
When session B queries for session A's snapshots
Then no snapshots SHALL be returned

### Requirement: Rollback Returns Result

Rollback SHALL return a RollbackResult containing lists of successfully rolled back and failed resources.

#### Scenario: Rollback Returns Result With Lists

Given a valid rollback scenario
When Rollback is executed
Then a RollbackResult SHALL be returned with RolledBack and Failed lists

---

## Overview

Automatic snapshot creation before each change and user-triggered rollback capability.

## Snapshot

### Creation Trigger

- Before each ChangeStep execution in EXECUTING state
- Automatic, no user action required

### Snapshot Storage

```go
type Snapshot struct {
    ID         string
    SessionID  string
    StepSeq    int
    ResourceID ResourceID
    Object     *unstructured.Unstructured
    CreatedAt  time.Time
}
```

### Storage Backend

- In-memory map: `map[ResourceID]*Snapshot`
- Key: `{namespace}/{kind}/{name}`
- Future: persistent storage (etcd/file)

## Rollback

### Trigger

- User calls `ChangeManager.Rollback(sessionID)`
- Change failed (automatic option presented to user)

### Rollback Process

1. Find latest snapshot for target resource
2. Restore object using Kubernetes client
3. Verify restore success
4. Log audit entry

### Rollback Steps

```go
type RollbackStep struct {
    Action   string
    Resource *unstructured.Unstructured
}

func (e *Executor) Rollback(sessionID string, target ResourceID) error {
    snapshot := findLatestSnapshot(sessionID, target)
    if snapshot == nil {
        return fmt.Errorf("no snapshot found for %v", target)
    }

    // Restore using dynamic client
    err := dynClient.Resource(gvr).Namespace(target.Namespace).Update(ctx, snapshot.Object)
    if err != nil {
        return fmt.Errorf("rollback failed: %v", err)
    }

    // Audit log
    audit.Log(sessionID, "ROLLBACK", target, "restored from snapshot "+snapshot.ID)
    return nil
}
```

### RollbackResult

```go
type RollbackResult struct {
    SessionID  string
    RolledBack []ResourceID
    Failed     []ResourceID
    At         time.Time
}
```

## Pre-Check Snapshot

### Check Name

`backup_snapshot`

### Behavior

- **Critical**: true — if snapshot creation fails, block execution
- **Before**: Create snapshot of target resource
- **After**: Verification not required (rollback is manual)

## Audit Integration

All snapshot and rollback operations logged:

```go
type AuditEntry struct {
    SessionID  string
    Timestamp time.Time
    Action    string  // SNAPSHOT / ROLLBACK / ROLLBACK_SUCCESS / ROLLBACK_FAILED
    Actor     string  // system
    Details   map[string]interface{}{
        "resource_id": ResourceID,
        "snapshot_id": string,
    }
}
```