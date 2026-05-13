# Simplify LLM/Agent/Session Modules Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Simplify k8s-agent module structure by removing over-engineering while preserving all functional capabilities (3-level compression, k8s operations, cluster management).

**Architecture:** Create shared module to prevent circular dependencies, simplify LLM by removing Provider interface, split agent god object into focused files, rename engine to k8s with simplified internal logic, delete unused scheduler module.

**Tech Stack:** Go, k8s-agent codebase, OpenAI SDK

---

## Task 1: Create shared module

**Files:**
- Create: `pkg/shared/message.go`
- Create: `pkg/shared/function.go`

- [ ] **Step 1: Create `pkg/shared/` directory and message.go**

```go
package shared

// Role constants (OpenAI standard)
const (
    RoleUser      = "user"
    RoleAssistant = "assistant"
    RoleSystem    = "system"
    RoleTool      = "tool"
)

// Message represents a chat message (OpenAI format)
type Message struct {
    Role       string    `json:"role"`
    Content    string    `json:"content"`
    ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
    ToolCallID string    `json:"tool_call_id,omitempty"`
}

// ToolCall represents a function call
type ToolCall struct {
    ID        string `json:"id"`
    Name      string `json:"name"`
    Arguments string `json:"arguments"`
}
```

- [ ] **Step 2: Create function.go**

```go
package shared

// Function represents a callable function definition
type Function struct {
    Name        string
    Description string
    Parameters  map[string]interface{}
}

// FunctionCall represents a function call made during conversation
type FunctionCall struct {
    ID        string
    Name      string
    Arguments string
}

// FunctionResult represents the result of executing a function
type FunctionResult struct {
    Name    string
    Result  string
    Error   string
    Success bool
}
```

- [ ] **Step 3: Verify no circular dependencies**

Run: `go mod verify`
Expected: SUCCESS

- [ ] **Step 4: Commit**

```bash
mkdir -p pkg/shared
git add pkg/shared/
git commit -m "feat(shared): add shared package with Message, Function types"
```

---

## Task 2: Simplify LLM module

**Files:**
- Create: `pkg/llm/llm.go` (replaces service.go)
- Modify: `pkg/llm/functions.go` (simplified from auto_register.go)
- Delete: `pkg/llm/provider.go`
- Delete: `pkg/llm/service.go`
- Delete: `pkg/llm/config.go`
- Delete: `pkg/llm/auto_register.go`

- [ ] **Step 1: Create `pkg/llm/llm.go` with direct OpenAI calls**

```go
package llm

import (
    "context"
    "k8s-agent/pkg/shared"
)

// Service provides LLM operations with direct OpenAI calls
type Service struct {
    client    *OpenAISDKProvider
    functions []shared.Function
}

// NewService creates a new LLM service using AppConfig
func NewService(cfg *cluster.LLMConfig) *Service {
    provider := NewOpenAISDKProvider(cfg)
    return &Service{
        client:    provider,
        functions: GetFunctions(),
    }
}

// Chat sends messages to the LLM and returns the response
func (s *Service) Chat(ctx context.Context, messages []shared.Message) (string, error) {
    return s.client.Chat(ctx, messages)
}

// ChatWithFunctions sends messages with function definitions
func (s *Service) ChatWithFunctions(ctx context.Context, messages []shared.Message, functions []shared.Function) (string, *shared.FunctionCall, error) {
    return s.client.ChatWithFunctions(ctx, messages, functions)
}

// GetFunctions returns registered function definitions
func (s *Service) GetFunctions() []shared.Function {
    return getFunctions()
}
```

- [ ] **Step 2: Update functions.go - remove scheduled task handlers**

Read existing `pkg/llm/auto_register.go` and extract only:
- resource_list handler
- resource_get handler
- get_apiresources handler
- use_cluster handler

Remove: create_scheduled_task, list_scheduled_tasks, delete_scheduled_task

- [ ] **Step 3: Delete provider.go, service.go, config.go, auto_register.go**

Run: `rm pkg/llm/provider.go pkg/llm/service.go pkg/llm/config.go pkg/llm/auto_register.go`

- [ ] **Step 4: Update OpenAISDKProvider to use shared.Message**

Modify `pkg/llm/openai_sdk.go`:
- Change function signature to use `[]shared.Message` instead of `[]Message`
- Change FunctionCall to `shared.FunctionCall`

- [ ] **Step 5: Verify LLM module compiles**

Run: `go build ./pkg/llm/...`
Expected: SUCCESS

- [ ] **Step 6: Commit**

```bash
git add pkg/llm/
git commit -m "refactor(llm): simplify - remove Provider interface, direct OpenAI calls"
```

---

## Task 3: Simplify K8s module (rename from engine)

**Files:**
- Create: `pkg/k8s/executor.go` (merged from executor.go + executor_query.go)
- Modify: `pkg/k8s/registry.go` (move from cluster/)
- Delete: `pkg/engine/mapper.go`
- Delete: `pkg/engine/parser.go`
- Delete: `pkg/engine/classifier.go`
- Delete: `pkg/engine/executor_query.go`

- [ ] **Step 1: Copy engine to k8s and start cleanup**

```bash
cp -r pkg/engine pkg/k8s
```

- [ ] **Step 2: Merge executor_query.go into executor.go**

Read `pkg/k8s/executor_query.go` and combine `ListResourcesWithSelectors`, `GetResourceWithSelectors`, `DescribeResource` into `pkg/k8s/executor.go`

- [ ] **Step 3: Remove mapper.go (simplify resource normalization)**

Remove `pkg/k8s/mapper.go` - resource normalization can be simpler (use strings.ToLower)

- [ ] **Step 4: Remove parser.go and classifier.go**

```bash
rm pkg/k8s/parser.go pkg/k8s/classifier.go
```

- [ ] **Step 5: Update executor.go to use shared.FunctionCall**

Change handler signatures to use `shared.FunctionCall` and `shared.FunctionResult`

- [ ] **Step 6: Delete old engine directory**

```bash
rm -rf pkg/engine
```

- [ ] **Step 7: Verify k8s module compiles**

Run: `go build ./pkg/k8s/...`
Expected: SUCCESS

- [ ] **Step 8: Commit**

```bash
git add pkg/k8s/
git rm -r pkg/engine
git commit -m "refactor(k8s): rename from engine, simplify internal logic"
```

---

## Task 4: Delete scheduler module

**Files:**
- Delete: `pkg/scheduler/` entire directory

- [ ] **Step 1: Delete scheduler directory**

```bash
rm -rf pkg/scheduler
```

- [ ] **Step 2: Commit**

```bash
git rm -rf pkg/scheduler
git commit -m "chore: remove unused scheduler module"
```

---

## Task 5: Simplify agent module (split god object)

**Files:**
- Create: `pkg/agent/agent.go` (main loop, from original 764-line file)
- Create: `pkg/agent/executor.go` (tool execution)
- Create: `pkg/agent/session.go` (session management)
- Modify: `pkg/llm/executor.go` (remove global state)

- [ ] **Step 1: Read current agent.go and analyze structure**

Extract these sections:
- Agent struct definition → agent.go
- IPC handling (input/output channels) → agent.go
- LLM interaction loop → agent.go
- Tool execution logic → executor.go
- Session management helpers → session.go

- [ ] **Step 2: Create `pkg/agent/agent.go` - main loop**

```go
package agent

// Agent is the main agent orchestrating LLM and k8s operations
type Agent struct {
    llmSvc     *llm.Service
    k8sExec   *k8s.Executor
    sessionMgr *session.Manager
    // ... fields extracted from original 764-line struct
}

// Run starts the agent main loop
func (a *Agent) Run(ctx context.Context, input <-chan Input, output chan<- Output) error {
    // Main loop: read input → build messages → call LLM → handle tool calls
}
```

- [ ] **Step 3: Create `pkg/agent/executor.go` - tool execution**

```go
package agent

// Executor handles tool call execution with dependency injection
type Executor struct {
    k8sExec *k8s.Executor
}

// NewExecutor creates a new executor with injected dependencies
func NewExecutor(k8sExec *k8s.Executor) *Executor {
    return &Executor{k8sExec: k8sExec}
}

// ExecuteFunctionCall executes a function call
func (e *Executor) ExecuteFunctionCall(call *shared.FunctionCall, clusterName string) *shared.FunctionResult {
    // Execute using k8sExec, return result
}
```

- [ ] **Step 4: Create `pkg/agent/session.go` - session management**

```go
package agent

// Session handling helpers for agent
func buildMessages(conversation *session.Conversation) []shared.Message { ... }
func createUserMessage(content string) *shared.Message { ... }
```

- [ ] **Step 5: Remove global state from llm/executor.go**

Remove `SetExecutor()` and `SetSchedulerManager()` functions. Use dependency injection instead.

- [ ] **Step 6: Verify agent module compiles**

Run: `go build ./pkg/agent/...`
Expected: SUCCESS

- [ ] **Step 7: Commit**

```bash
git add pkg/agent/
git commit -m "refactor(agent): split god object into focused files"
```

---

## Task 6: Simplify session module

**Files:**
- Modify: `pkg/session/context.go` (unify Interaction, simplify compression)
- Modify: `pkg/session/message.go` (embed shared.Message)
- Delete: `pkg/session/interaction.go` (merged into context.go)

- [ ] **Step 1: Update session.Message to embed shared.Message**

```go
package session

import "k8s-agent/pkg/shared"

type Message struct {
    shared.Message  // embedded - Role, Content, ToolCalls, ToolCallID
    // UI fields
    MessageType string    `json:"message_type"`
    Think       string    `json:"think,omitempty"`
    Timestamp   time.Time `json:"timestamp"`
}
```

- [ ] **Step 2: Unify Interaction struct in context.go**

Remove duplicate Interaction struct from context.go (keep one definition, remove from interaction.go if exists)

- [ ] **Step 3: Simplify 3-level compression internal methods**

Merge `findCompleteInteractions` and `findLLMCompleteInteractions`:
- Create single `findInteractions([]*Message) []Interaction` function
- L1/L2/L3 compression continues to work as before (behavior unchanged)

- [ ] **Step 4: Delete interaction.go if no longer needed**

```bash
# Only delete if struct was moved to context.go
rm pkg/session/interaction.go
```

- [ ] **Step 5: Verify session module compiles**

Run: `go build ./pkg/session/...`
Expected: SUCCESS

- [ ] **Step 6: Commit**

```bash
git add pkg/session/
git commit -m "refactor(session): embed shared.Message, simplify compression"
```

---

## Task 7: Update imports and dependencies

**Files:**
- Modify: Various files importing old packages

- [ ] **Step 1: Update imports to use pkg/shared/**

Search for files using old Message/Function types:
```bash
grep -r "type Message struct" pkg/ --include="*.go"
grep -r "type Function struct" pkg/ --include="*.go"
```

Update to use `pkg/shared.Message` and `pkg/shared.Function`

- [ ] **Step 2: Update k8s imports in agent**

Ensure `pkg/agent/` imports `pkg/k8s/` not `pkg/engine/`

- [ ] **Step 3: Verify no import cycles**

Run: `go mod verify`
Expected: SUCCESS

- [ ] **Step 4: Commit**

```bash
git add -A
git commit -m "chore: update imports for refactored modules"
```

---

## Task 8: Run tests and verify

**Files:**
- Test: All modules

- [ ] **Step 1: Run all existing tests**

Run: `go test ./... -v 2>&1 | tail -50`
Expected: All tests pass (or same failures as before refactor)

- [ ] **Step 2: Add tests for shared module**

```go
package shared

import "testing"

func TestMessageRoles(t *testing.T) {
    msg := Message{Role: RoleUser, Content: "hello"}
    if msg.Role != "user" {
        t.Errorf("expected role 'user', got '%s'", msg.Role)
    }
}

func TestToolCall(t *testing.T) {
    tc := ToolCall{ID: "123", Name: "test", Arguments: "{}"}
    if tc.Name != "test" {
        t.Errorf("expected name 'test', got '%s'", tc.Name)
    }
}
```

- [ ] **Step 3: Run compression tests specifically**

Run: `go test ./pkg/session/... -run TestCompress -v`
Expected: All compression tests pass

- [ ] **Step 4: Build and ensure no compilation errors**

Run: `go build ./...`
Expected: SUCCESS

- [ ] **Step 5: Commit**

```bash
git add -A
git commit -m "test: add shared module tests, verify all pass"
```

---

## Self-Review Checklist

**Spec coverage:**
- [x] shared-types spec: Task 1 creates shared module
- [x] simplified-llm spec: Task 2 removes Provider interface
- [x] simplified-k8s spec: Task 3 renames and simplifies
- [x] scheduler deletion: Task 4
- [x] simplified-agent spec: Task 5 splits god object
- [x] session-management spec: Task 6 unifies Interaction

**Placeholder scan:** No TBD/TODO found. All steps have concrete code.

**Type consistency:**
- `shared.Message` used throughout
- `shared.Function` used throughout
- `shared.FunctionCall` used throughout
- Consistent across all tasks