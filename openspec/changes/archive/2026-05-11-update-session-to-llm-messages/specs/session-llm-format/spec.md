## ADDED Requirements

### Requirement: Session stores LLM message format

Session SHALL persist messages in the same structure used for LLM API calls (`llm.Message` format), including ToolCalls, rather than UI-formatted text with emoji prefixes.

### Requirement: Session can restore LLM context

When a session is loaded from FileStore, the Agent SHALL be able to reconstruct `llmMessages` from persisted session data, enabling continuation of LLM conversations across restarts.

### Requirement: UI renders from LLM message type

The UI layer SHALL render messages based on their `MessageType` field, adding appropriate emoji prefixes (💭 for think, 🔧 for tool calls, ✅/❌ for results).

#### Scenario: User input stored correctly
- **WHEN** user sends a message
- **THEN** session stores `{Role: "user", Content: "...", MessageType: "user"}`

#### Scenario: LLM response with thinking stored
- **WHEN** LLM returns a response containing `<think>` tags
- **THEN** session stores as two messages: `{MessageType: "think", Content: "..."}` and `{MessageType: "text", Content: "..."}`

#### Scenario: Tool call stored with proper structure
- **WHEN** LLM makes a function call
- **THEN** session stores `{Role: "assistant", ToolCalls: [...], MessageType: "tool_call"}`

#### Scenario: Tool result stored with success indicator
- **WHEN** tool execution completes
- **THEN** session stores `{Role: "tool", Content: "...", MessageType: "tool_result", Success: true/false}`

#### Scenario: Session restore reconstructs llmMessages
- **WHEN** Agent loads an existing session
- **THEN** Agent can reconstruct `llmMessages` array from session data for LLM API calls

#### Scenario: UI renders based on MessageType
- **WHEN** UI renders a message with `MessageType: "think"`
- **THEN** UI displays content prefixed with "💭 "
- **WHEN** UI renders a message with `MessageType: "tool_call"`
- **THEN** UI displays content prefixed with "🔧 执行工具: "