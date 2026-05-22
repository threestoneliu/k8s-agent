# State Machine Specification

## ADDED Requirements

### Requirement: State Machine States

The state machine SHALL have exactly 7 states: PARSING, CLARIFYING, PLANNING, REVIEWING, EXECUTING, COMPLETED, FAILED

#### Scenario: States Exist

Given a newly created ChangeSession
When the session is initialized
Then the session state SHALL be PARSING

#### Scenario: All 7 States Defined

Given the state machine implementation
When all states are enumerated
Then there SHALL be exactly 7 states

### Requirement: State Transitions

The state machine SHALL transition based on user signals (Confirm, Abort, Modify, Proceed).

#### Scenario: Confirm Advances State

Given a session in PARSING state with valid intent
When SignalConfirm is received
Then the session SHALL transition to PLANNING

### Requirement: Signal Handling

The state machine SHALL handle signals as follows:
- SignalConfirm SHALL advance the current step
- SignalAbort SHALL transition to FAILED
- SignalModify SHALL return to PLANNING
- SignalProceed SHALL allow dangerous operations to proceed

#### Scenario: Abort Transitions to FAILED

Given a session in any non-terminal state
When SignalAbort is received
Then the session SHALL transition to FAILED

#### Scenario: Terminal State Rejects All Signals

Given a session in COMPLETED or FAILED state
When any signal is received
Then the state SHALL remain unchanged

---

## Overview

ChangeSession implements a deterministic state machine for managing the lifecycle of a change request.

## States

| State | Description |
|-------|-------------|
| PARSING | Parsing user intent into ParsedIntent |
| CLARIFYING | Requesting clarification from user |
| PLANNING | Generating ChangePlan |
| REVIEWING | User reviewing plan and impact |
| EXECUTING | Executing change with checks |
| COMPLETED | Change completed successfully |
| FAILED | Change failed or aborted |

## State Transitions

```
StartSession()
    │
    ▼
[PARSING] ──(intent incomplete)──→ [CLARIFYING]
    │                                  │
    │ (intent valid)                   │ (clarified)
    ▼                                  ▼
[PLANNING] ←────────────────────── [PARSING]
    │
    │ (plan generated)
    ▼
[REVIEWING] ──(modified)──→ [PLANNING]
    │
    │ (approved)
    ▼
[EXECUTING] ──(failed)──→ [FAILED]
    │
    │ (success)
    ▼
[COMPLETED]
```

## Signals

| Signal | Description |
|--------|-------------|
| SignalConfirm | User confirms current step |
| SignalAbort | User aborts execution |
| SignalModify | User modifies parameters |
| SignalProceed | User proceeds with dangerous operation |

## Transition Rules

### PARSING

- **Entry**: Validate ParsedIntent completeness
- **Exit (valid)**: Transition to PLANNING
- **Exit (incomplete)**: Transition to CLARIFYING with ClarifyQuestion

### CLARIFYING

- **Entry**: Generate ClarifyQuestion for missing fields
- **Exit**: Return to PARSING with updated intent
- **Loop**: Stay in CLARIFYING until intent complete

### PLANNING

- **Entry**: Generate ChangePlan with ResourceDiff, RiskLevel
- **Exit**: Transition to REVIEWING
- **Output**: ChangePlan struct

### REVIEWING

- **Entry**: Present plan to user with risk assessment
- **Exit (approved)**: Transition to EXECUTING
- **Exit (modified)**: Return to PLANNING with modifications

### EXECUTING

- **Entry**: Run pre-checks
- **Loop**: Execute ChangeSteps sequentially
- **Exit (success)**: Transition to COMPLETED
- **Exit (failed)**: Transition to FAILED
- **Exit (abort)**: Transition to FAILED

### COMPLETED / FAILED

- **Terminal states**: No further transitions
- **Rollback available**: From FAILED state

## Session ID

- UUID format for session identification
- Sessions stored in memory map (future: persistent storage)

## Concurrency

- Sessions are independent
- Single session is single-threaded
- Session operations are mutex-protected