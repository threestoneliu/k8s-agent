# Implementation Tasks

## Phase 1: Core State Machine

### Task 1: Create pkg/core/ directory structure

- [x] Create directory `pkg/core/`
- [x] Create `state.go` with State type and constants
- [x] Create `session.go` with ChangeSession struct
- [x] Create `api.go` with ChangeManager interface

### Task 2: Implement State type and transitions

- [x] Define State type with 7 states: PARSING, CLARIFYING, PLANNING, REVIEWING, EXECUTING, COMPLETED, FAILED
- [x] Define SessionSignal type with 4 signals: Confirm, Abort, Modify, Proceed
- [x] Implement `Transition(session, signal) (State, error)` function
- [x] Add transition validation rules
- [x] Write unit tests for state transitions

### Task 3: Implement ChangeSession

- [x] Define ChangeSession struct with ID, State, Intent, Plan fields
- [x] Implement `NewChangeSession(sessionID string, intent ParsedIntent) *ChangeSession`
- [x] Add session mutex for concurrency safety
- [x] Write unit tests for ChangeSession

### Task 4: Implement ChangeManager API

- [x] Implement `StartSession(sessionID string, intent ParsedIntent) (State, error)`
- [x] Implement `Advance(sessionID string, signal SessionSignal) (State, error)`
- [x] Implement `GetSession(sessionID string) *ChangeSession`
- [x] Implement `Abort(sessionID string) error`
- [x] Implement `ListSessions() []*ChangeSession`
- [x] Write integration tests

## Phase 2: Intent Validation

### Task 5: Implement ParsedIntent validator

- [x] Create `validator.go`
- [x] Define `ValidateIntent(intent *ParsedIntent) *ClarifyQuestion`
- [x] Implement required field checks
- [x] Implement risk level defaults
- [x] Write unit tests

### Task 6: Implement ClarifyQuestion generation

- [x] Define ClarifyQuestion struct
- [x] Implement question generation rules
- [x] Handle blocking vs optional questions
- [x] Write unit tests

## Phase 3: ChangePlan Generation

### Task 7: Implement ChangePlan generator

- [x] Create `planner.go`
- [x] Define ChangePlan and ChangeStep structs
- [x] Implement `GeneratePlan(intent ParsedIntent) *ChangePlan`
- [x] Implement ResourceDiff calculation
- [x] Implement RiskLevel assessment
- [x] Write unit tests

### Task 8: Implement ResourceDiff

- [ ] Create `diff.go`
- [ ] Define ResourceDiff and FieldChange structs
- [ ] Implement `CalculateDiff(before, after *unstructured.Unstructured) *ResourceDiff`
- [ ] Write unit tests

## Phase 4: Change Execution

### Task 9: Implement ChangeExecutor

- [ ] Create `executor.go`
- [ ] Define PreCheck struct and list
- [ ] Implement `Execute(sessionID string) error`
- [ ] Implement pre-check hooks
- [ ] Implement step-by-step execution
- [ ] Write integration tests

### Task 10: Integrate executor with state machine

- [ ] Modify EXECUTING state to call executor
- [ ] Handle execution errors → FAILED transition
- [ ] Handle execution success → COMPLETED transition
- [ ] Write integration tests

## Phase 5: Snapshot & Rollback

### Task 11: Implement Snapshot

- [x] Create `rollback.go`
- [x] Define Snapshot struct
- [x] Implement `CreateSnapshot(sessionID string, resource ResourceID) (*Snapshot, error)`
- [x] Implement in-memory storage
- [x] Write unit tests

### Task 12: Implement Rollback

- [x] Implement `Rollback(sessionID string) (*RollbackResult, error)`
- [x] Implement latest snapshot lookup
- [x] Implement resource restoration
- [x] Write unit tests

## Phase 6: Audit Logging

### Task 13: Implement AuditLog

- [x] Create `audit.go`
- [x] Define AuditEntry struct
- [x] Implement `Log(sessionID, action, actor string, details map[string]interface{})`
- [x] Implement in-memory storage
- [x] Implement `GetAuditLog(sessionID string) []AuditEntry`
- [x] Write unit tests

### Task 14: Integrate audit into state machine

- [ ] Log all state transitions
- [ ] Log all executor operations
- [ ] Log rollback operations
- [ ] Write integration tests

## Phase 7: Agent Integration

### Task 15: Extend IPC protocol

- [ ] Add SessionID to Output struct
- [ ] Add State to Output struct
- [ ] Add Plan/Diff/ClarifyQuestion to Output struct
- [ ] Add RequiresConfirm field

### Task 16: Implement Agent translator

- [ ] Create `pkg/agent/translator.go`
- [ ] Implement `PlanToNaturalLanguage(plan *ChangePlan) string`
- [ ] Implement `DiffToNaturalLanguage(diff *ResourceDiff) string`
- [ ] Write unit tests

### Task 17: Implement Agent intent parser

- [ ] Create `pkg/agent/intent/parser.go`
- [ ] Implement `ParseToIntent(userInput string) (*ParsedIntent, error)`
- [ ] Integrate with LLM
- [ ] Write unit tests

## Phase 8: Testing & Documentation

### Task 18: Integration tests

- [ ] Write end-to-end test for complete flow
- [ ] Test abort/resume/rollback flows
- [ ] Test multi-session concurrency

### Task 19: Documentation

- [ ] Add godoc comments to all public APIs
- [ ] Create ARCHITECTURE.md for overall design
- [ ] Update README with new features