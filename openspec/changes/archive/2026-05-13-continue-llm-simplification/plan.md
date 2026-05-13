# Continue LLM/Agent/Session Simplification Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Complete the remaining refactoring tasks: delete scheduler module, rename pkg/engine to pkg/k8s, split agent module, and simplify session compression.

**Architecture:** Refactoring-focused change. Remove unused code (scheduler), rename modules for clarity (engine→k8s), split large files (agent), simplify internal implementation (session compression). No architectural changes to external APIs or IPC protocols.

**Tech Stack:** Go, golang.org/x/oauth2, k8s.io/client-go

---

## Task 1: Delete Scheduler Module

**Files:**
- Delete: `pkg/scheduler/` (entire directory)
- Modify: `pkg/llm/functions.go` (remove scheduled task handlers)

- [ ] **Step 1: Verify scheduler is not imported elsewhere**

Run: `grep -r "k8s-agent/pkg/scheduler" --include="*.go" .`
Expected: Only references in cmd/cli (task command already removed)

- [ ] **Step 2: Delete pkg/scheduler directory**

```bash
rm -rf pkg/scheduler/
```

- [ ] **Step 3: Remove scheduled task handlers from llm/functions.go**

Check for: `create_scheduled_task`, `list_scheduled_tasks`, `delete_scheduled_task` in functions.go
Remove any RegisterFunction calls for these handlers.

- [ ] **Step 4: Verify build**

Run: `go build ./...`
Expected: Success with no scheduler references

- [ ] **Step 5: Run tests**

Run: `go test ./...`
Expected: All tests pass

- [ ] **Step 6: Commit**

```bash
git add -A && git commit -m "refactor: delete scheduler module"
```

---

## Task 2: LLM Module Cleanup

**Files:**
- Modify: `pkg/llm/llm.go`
- Modify: `pkg/llm/functions.go`

- [ ] **Step 1: Check for Provider interface references**

Run: `grep -r "Provider" --include="*.go" pkg/llm/`
Expected: No matches (should already be clean from previous session)

- [ ] **Step 2: Verify go mod dependencies**

Run: `go mod tidy && go build ./pkg/llm/...`
Expected: Success

- [ ] **Step 3: Run LLM tests**

Run: `go test ./pkg/llm/... -v`
Expected: All LLM tests pass

- [ ] **Step 4: Commit**

```bash
git add -A && git commit -m "chore: LLM module already clean from previous session"
```

---

## Task 3: Rename pkg/engine to pkg/k8s

**Files:**
- Rename: `pkg/engine/` → `pkg/k8s/`
- Modify: All files importing `k8s-agent/pkg/engine`

- [ ] **Step 1: Find all files importing pkg/engine**

Run: `grep -rl "k8s-agent/pkg/engine" --include="*.go" .`
Expected: List of files to update

- [ ] **Step 2: Rename directory**

```bash
git mv pkg/engine pkg/k8s
```

- [ ] **Step 3: Update all imports**

Find and replace `"k8s-agent/pkg/engine"` → `"k8s-agent/pkg/k8s"` in all Go files.

Files typically affected:
- `cmd/cli/root.go`
- `cmd/cli/chat.go`
- `pkg/agent/*.go`
- `pkg/llm/*.go`

- [ ] **Step 4: Update CLAUDE.md references**

Check: `grep -r "pkg/engine" CLAUDE.md`
If found, update to `pkg/k8s`

- [ ] **Step 5: Verify build**

Run: `go build ./...`
Expected: Success with no engine references

- [ ] **Step 6: Run tests**

Run: `go test ./...`
Expected: All tests pass

- [ ] **Step 7: Commit**

```bash
git add -A && git commit -m "refactor: rename pkg/engine to pkg/k8s"
```

---

## Task 4: Merge executor_query.go into executor.go

**Files:**
- Modify: `pkg/k8s/executor.go` (merge query functions)
- Delete: `pkg/k8s/executor_query.go`

- [ ] **Step 1: Read executor_query.go content**

Run: `cat pkg/k8s/executor_query.go`
Expected: List resource functions to merge

- [ ] **Step 2: Merge functions into executor.go**

Move all functions from executor_query.go into executor.go.
Delete executor_query.go after merging.

- [ ] **Step 3: Remove parser.go and classifier.go**

```bash
rm pkg/k8s/parser.go pkg/k8s/classifier.go
```

- [ ] **Step 4: Verify build**

Run: `go build ./pkg/k8s/...`
Expected: Success

- [ ] **Step 5: Run tests**

Run: `go test ./pkg/k8s/...`
Expected: All tests pass

- [ ] **Step 6: Commit**

```bash
git add -A && git commit -m "refactor: merge executor_query into executor, remove parser/classifier"
```

---

## Task 5: Agent Module Split

**Files:**
- Create: `pkg/agent/executor.go`
- Create: `pkg/agent/session.go`
- Modify: `pkg/agent/agent.go`

- [ ] **Step 1: Read current agent.go structure**

Run: `head -100 pkg/agent/agent.go`
Expected: Understand main loop and dependencies

- [ ] **Step 2: Create executor.go with ExecuteFunctionCall**

Extract from agent.go:
- ExecuteFunctionCall function
- Any tool result formatting

```go
// pkg/agent/executor.go
package agent

import (
    sharedutil "k8s-agent/pkg/shared"
)

// ExecuteFunctionCall executes a function call and returns the result
func ExecuteFunctionCall(call *sharedutil.FunctionCall, clusterName string) *sharedutil.FunctionResult {
    // Move this logic from agent.go
}
```

- [ ] **Step 3: Create session.go with session helpers**

Extract from agent.go:
- NewMessageWithType function
- Session message creation helpers
- Cluster context management

```go
// pkg/agent/session.go
package agent

import (
    session "k8s-agent/pkg/session"
    sharedutil "k8s-agent/pkg/shared"
)

// NewMessageWithType creates a session message with explicit MessageType
func NewMessageWithType(role, content string, msgType string, metadata map[string]string) *session.Message {
    // Move this logic from agent.go
}
```

- [ ] **Step 4: Refactor agent.go to use new files**

Remove duplicated code, keep only:
- Agent struct
- Run method (main loop)
- processInput, processWithOutput methods

- [ ] **Step 5: Remove global state setters**

Search for: `SetExecutor`, `SetSchedulerManager`
Remove these if found (use dependency injection via constructor instead)

- [ ] **Step 6: Update agent tests**

Check `pkg/agent/*_test.go` and update imports.

- [ ] **Step 7: Verify build**

Run: `go build ./pkg/agent/...`
Expected: Success

- [ ] **Step 8: Run tests**

Run: `go test ./pkg/agent/...`
Expected: All agent tests pass

- [ ] **Step 9: Commit**

```bash
git add -A && git commit -m "refactor: split agent into agent.go, executor.go, session.go"
```

---

## Task 6: Session Compression Simplify

**Files:**
- Modify: `pkg/session/context.go` (merge compression functions)
- Modify: `pkg/session/interaction.go` (unify Interaction struct)

- [ ] **Step 1: Read current compression functions**

Run: `grep -n "findCompleteInteractions" pkg/session/*.go`
Expected: Two functions to merge

- [ ] **Step 2: Merge findCompleteInteractions functions**

Combine `findCompleteInteractions` and `findLLMCompleteInteractions` into a single function.
Keep L1/L2/L3 behavior identical.

- [ ] **Step 3: Verify Interaction struct is unified**

Run: `grep -n "type Interaction struct" pkg/session/*.go`
Expected: Only one definition in interaction.go

- [ ] **Step 4: Run compression tests**

Run: `go test ./pkg/session/... -v -run Compress`
Expected: All compression tests pass

- [ ] **Step 5: Run all session tests**

Run: `go test ./pkg/session/...`
Expected: All session tests pass

- [ ] **Step 6: Commit**

```bash
git add -A && git commit -m "refactor: simplify session compression, merge duplicate functions"
```

---

## Task 7: Final Build and Test Verification

**Files:**
- All packages

- [ ] **Step 1: Run go mod tidy**

Run: `go mod tidy`
Expected: Clean dependency graph

- [ ] **Step 2: Run full build**

Run: `go build ./...`
Expected: Success

- [ ] **Step 3: Run all tests**

Run: `go test ./...`
Expected: All tests pass

- [ ] **Step 4: Verify no circular deps**

Run: `go mod verify`
Expected: All OK

- [ ] **Step 5: Build binary**

Run: `go build -o k8s-agent ./cmd/k8s-agent`
Expected: Binary created

- [ ] **Step 6: Final commit**

```bash
git add -A && git commit -m "chore: final verification - all tests pass"
```

---

## Execution Options

**1. Subagent-Driven (recommended)** - Dispatch fresh subagent per task, review between tasks

**2. Inline Execution** - Execute tasks in this session using executing-plans

Which approach?
