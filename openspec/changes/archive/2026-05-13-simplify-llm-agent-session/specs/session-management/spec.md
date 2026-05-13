# Session Management

## ADDED Requirements

### Requirement: Unified Interaction Struct

Session module SHALL use a single `Interaction` struct (not duplicated in context.go and interaction.go).

```go
type Interaction struct {
    Query            string
    ToolNames        []string
    Summary          string
    Completed        bool
    OriginalMessages []*Message
}
```

#### Scenario: Interaction struct is defined once

Given session/interaction.go defines Interaction
When searching codebase
Then only one definition exists
And no duplicate in context.go

---

### Requirement: OpenAI Standard Roles

Session messages SHALL use OpenAI standard roles: user, assistant, system, tool.

#### Scenario: Message.Role uses shared.RoleUser/RoleAssistant/RoleSystem/RoleTool

Given a session.Message created with Role=shared.RoleUser
When inspecting the embedded Message
Then Role equals "user"

---

### Requirement: MessageType for UI Only

MessageType SHALL only be used for UI rendering (text, think, user). Not for tool call/result distinction.

#### Scenario: MessageType values used for UI display

Given session.Message with MessageType="text" or "think" or "user"
When UI renders the message
Then appropriate formatting is applied
And tool call/result uses Message.Content with embedded Role distinction

---

### Requirement: Embed Shared Message

`session.Message` SHALL embed `shared.Message` fields rather than duplicating them.

#### Scenario: session.Message embeds shared.Message

Given session.Message embeds sharedutil.Message
When accessing Role field
Then it comes from embedded shared.Message
And no duplicate Role field definition exists in session.Message

---

---

## MODIFIED Requirements

### Requirement: 3-Level Compression Preserved

The 3-level context compression SHALL be preserved:
- L1: Interaction-based compression
- L2: Message count truncation
- L3: Summary generation

The internal implementation SHALL be simplified by:
- Merging findCompleteInteractions and findLLMCompleteInteractions into one function
- Removing duplicate logic between L1/L2/L3

The capability (when and how to compress) SHALL remain identical.

#### Scenario: L1 compression triggers on interaction count

Given session with more than ToolCallRetention completed interactions
When building LLM context
Then L1 compression replaces older interactions with summary

#### Scenario: L2 compression triggers on message count

Given session with message count > MaxMessages after L1
When building LLM context
Then L2 truncation removes oldest messages keeping recent

#### Scenario: L3 compression triggers on token limit

Given messages still exceed MaxTokens after L1+L2
When building LLM context
Then L3 summary generation reduces token count

---

### Requirement: Compression Trigger

Compression SHALL trigger when:
- L1: interaction count exceeds ToolCallRetention
- L2: message count exceeds MaxMessages after L1
- L3: token count still exceeds MaxTokens after L2

#### Scenario: Compression triggers in correct order

Given a session with all three compression conditions met
When Compress() is called
Then L1 runs first, then L2, then L3
And each level only runs if previous level didn't fully reduce

---

## Implementation Notes

**Files:**
- `pkg/session/manager.go` — Session creation/management (unchanged)
- `pkg/session/store.go` — Store interface (simplified)
- `pkg/session/context.go` — ContextManager + 3-level compression (internally simplified)
- `pkg/session/interaction.go` — Interaction struct (unified, removed duplicate)
- `pkg/session/compressor.go` — Compression logic (simplified)
- `pkg/session/message.go` — Message with shared.Message embedded
- `pkg/session/filestore.go` — File storage (can be simplified)

**Key Changes:**
1. Remove duplicate Interaction definition — use single definition
2. Simplify compression internal methods — same capability, less code
3. Keep L1/L2/L3 trigger conditions unchanged — behavior identical