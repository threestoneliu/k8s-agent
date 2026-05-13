# Shared Types

## ADDED Requirements

### Requirement: Shared Message Structure

The `shared.Message` struct SHALL contain Role, Content, ToolCalls, and ToolCallID fields aligned with OpenAI message format.

```go
type Message struct {
    Role       string    `json:"role"`        // user/assistant/system/tool
    Content    string    `json:"content"`
    ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
    ToolCallID string    `json:"tool_call_id,omitempty"`
}
```

#### Scenario: Message marshals to OpenAI-compatible JSON

Given a Message with Role="user", Content="get pods", and no tool calls
When marshaled to JSON
Then the output contains "role":"user" and "content":"get pods"
And matches OpenAI API message format

#### Scenario: Message with tool calls serializes correctly

Given a Message with Role="assistant", tool_calls=[{ID:"call_123", Name:"k8s_get", Arguments:"{}"}]
When marshaled to JSON
Then tool_calls array is present with correct structure

---

### Requirement: Role Constants

Role constants SHALL use OpenAI standard values: user, assistant, system, tool.

#### Scenario: Role constants match OpenAI spec

Given shared.RoleUser, shared.RoleAssistant, shared.RoleSystem, shared.RoleTool
When compared to OpenAI API
Then values are "user", "assistant", "system", "tool" respectively

---

### Requirement: ToolCall Structure

ToolCall struct SHALL contain ID, Name, and Arguments fields.

```go
type ToolCall struct {
    ID        string `json:"id"`
    Name      string `json:"name"`
    Arguments string `json:"arguments"`
}
```

#### Scenario: ToolCall serializes to JSON with all fields

Given a ToolCall with ID="call_abc", Name="resource_list", Arguments='{"namespace":"default"}'
When marshaled to JSON
Then all three fields appear in output
And Arguments is a JSON string (escaped)

---

### Requirement: Shared Function Types

The `Function` and `FunctionCall` types SHALL be in `shared` module for common access.

```go
type Function struct {
    Name        string
    Description string
    Parameters  map[string]interface{}
}

type FunctionCall struct {
    ID        string
    Name      string
    Arguments string
}
```

#### Scenario: Function struct contains required fields

Given a Function with Name="resource_get", Description="Get k8s resource", Parameters={"type":"object"}
Then all three fields are accessible and match input values

#### Scenario: FunctionCall struct used by LLM executor

Given a FunctionCall returned from LLM with Name="k8s_get", Arguments='{"name":"nginx"}'
When passed to ExecuteFunctionCall
Then the handler receives correctly parsed arguments

---

### Requirement: No Circular Dependencies

No module using shared types SHALL import back into shared module, maintaining acyclic dependency graph.

#### Scenario: pkg/llm imports pkg/shared

Given pkg/llm imports pkg/shared
When checking dependency graph
Then pkg/shared does not import pkg/llm

#### Scenario: pkg/session imports pkg/shared

Given pkg/session imports pkg/shared
When checking dependency graph
Then pkg/shared does not import pkg/session

#### Scenario: go mod verify passes

Given all module dependencies
When running `go mod verify`
Then no cyclic dependencies detected

---

## Implementation Notes

- `pkg/shared/message.go` — Message, ToolCall, Role constants
- `pkg/shared/function.go` — Function, FunctionCall types
- Both `pkg/llm/` and `pkg/session/` import `pkg/shared/`