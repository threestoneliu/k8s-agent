## ADDED Requirements

### Requirement: User can type input and see it in viewport

When user types text and presses Enter, the text SHALL appear in the TUI viewport as a user message.

**Test category**: Unit (no LLM)

#### Scenario: User types "get pods" and presses Enter
- **WHEN** user types "get pods" and presses Enter
- **THEN** the message "get pods" appears in the viewport as a user message

---

### Requirement: User can execute /clusters command

When user types `/clusters` and presses Enter, the TUI SHALL display available clusters.

**Test category**: Unit (no LLM)

#### Scenario: User types /clusters and sees cluster list
- **WHEN** user types "/clusters" and presses Enter
- **THEN** TUI receives Input{Text: "/clusters"}
- **AND** output shows "可用集群:" followed by cluster list

---

### Requirement: User can execute /cluster <name> command

When user types `/cluster <name>` and presses Enter, the TUI SHALL switch cluster context.

**Test category**: Unit (no LLM)

#### Scenario: User types /cluster dev to switch
- **WHEN** user types "/cluster dev" and presses Enter
- **THEN** TUI receives Input{Text: "/cluster dev"}

---

### Requirement: User can execute /exit or /quit to exit

When user types `/exit` or `/quit`, the TUI SHALL quit gracefully.

**Test category**: Unit (no LLM)

#### Scenario: User types /exit
- **WHEN** user types "/exit" and presses Enter
- **THEN** TUI sends tea.Quit signal

#### Scenario: User types /quit
- **WHEN** user types "/quit" and presses Enter
- **THEN** TUI sends tea.Quit signal

---

### Requirement: Agent responses render correctly in viewport

When agent sends Output messages, the TUI SHALL render them with appropriate formatting based on MessageType.

**Test category**: Unit (no LLM) — UI rendering is tested with mock Output

#### Scenario: Agent sends text output
- **WHEN** agent sends Output{Type: "text", Content: "Here are the pods..."}
- **THEN** content appears in viewport with markdown styling

#### Scenario: Agent sends think output
- **WHEN** agent sends Output{Type: "think", Content: "分析中..."}
- **THEN** content appears in viewport prefixed with "💭 "

#### Scenario: Agent sends tool call output
- **WHEN** agent sends Output{Type: "tool_call_start", ToolName: "k8s_get"}
- **THEN** content appears prefixed with "🔧 执行工具: k8s_get(...)"

#### Scenario: Agent sends tool result output (success)
- **WHEN** agent sends Output{Type: "tool_result", ToolSuccess: true}
- **THEN** content appears prefixed with "✅ "

#### Scenario: Agent sends tool result output (failure)
- **WHEN** agent sends Output{Type: "tool_result", ToolSuccess: false, ToolResult: "error message"}
- **THEN** content appears prefixed with "❌ 工具执行失败: error message"

---

### Requirement: Full chat flow with real LLM (integration)

The complete user → Agent → LLM API → UI flow SHALL work correctly with real LLM.

**Test category**: Integration (with real LLM, skipped by default)

#### Scenario: User chat with LLM response
- **WHEN** user types "get pods" and presses Enter
- **AND** real LLM returns a text response
- **THEN** the response appears in viewport with markdown styling

#### Scenario: User chat with LLM tool call
- **WHEN** user types "delete pod nginx" and presses Enter
- **AND** real LLM returns a tool call
- **AND** tool returns success
- **THEN** tool result appears in viewport with "✅ " prefix

#### Scenario: User chat with LLM think tags
- **WHEN** user types "describe pod nginx" and presses Enter
- **AND** real LLM returns response with <think> tags
- **THEN** think content appears with "💭 " prefix
- **AND** regular content appears after with markdown styling
