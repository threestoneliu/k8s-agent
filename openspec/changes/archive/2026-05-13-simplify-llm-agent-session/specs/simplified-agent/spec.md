# Simplified Agent Module

## ADDED Requirements

### Requirement: Agent Split into Focused Files

The agent module SHALL be split into three files:
- `agent.go`: Main agent loop, IPC handling, LLM interaction
- `executor.go`: Tool call execution, result formatting
- `session.go`: Session management, message creation

#### Scenario: Three focused files exist

Given `pkg/agent/` directory
When listing files
Then agent.go, executor.go, session.go all exist
And each contains only related functionality

---

### Requirement: No Global State

The `llm.Executor` SHALL use dependency injection, not global `SetExecutor()`/`SetSchedulerManager()`.

#### Scenario: Agent receives dependencies via constructor

Given Agent struct with llmSvc, executor fields
When creating agent via NewAgent(llmSvc, executor, ...)
Then no global state is set or read

---

### Requirement: Message Conversion

Agent SHALL convert `shared.Message` to OpenAI SDK format when calling LLM.

#### Scenario: session.Message converts to OpenAI format

Given session.Message with Role=shared.RoleUser, Content="get pods"
When ReconstructLLMMessages() is called
Then returned messages have Role="user" matching OpenAI SDK format

---

### Requirement: Session Message Structure

Agent SHALL create `session.Message` with appropriate Role and MessageType for UI rendering.

#### Scenario: Agent sets MessageType for UI rendering

Given agent processing user input "get pods"
When creating session.Message
Then MessageType is set appropriately (user/text/think/tool_call)
And UI uses MessageType to apply correct formatting

---

## Implementation Notes

**File Responsibilities:**

`pkg/agent/agent.go`:
- Agent struct with dependencies: llmSvc, k8sExecutor, sessionMgr
- Main loop: receive input → build messages → call LLM → handle tool calls → return output
- IPC handling (Input/Output channels)

`pkg/agent/executor.go`:
- ExecuteFunctionCall(call *shared.FunctionCall, clusterName string) *FunctionResult
- Uses k8s executor for actual k8s operations
- No global state

`pkg/agent/session.go`:
- Session creation and management
- Message creation helpers
- Cluster context management

**Dependencies:**
- imports pkg/shared/, pkg/llm/, pkg/k8s/, pkg/session/