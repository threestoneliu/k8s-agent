# Session Compression Simplify

## ADDED Requirements

### Requirement: Single Compression Function

The session module SHALL use a single function for finding complete interactions, merging `findCompleteInteractions` and `findLLMCompleteInteractions`.

#### Scenario: No duplicate find functions

Given `pkg/session/` module
When searching for findCompleteInteractions functions
Then only one function exists with merged logic

---

### Requirement: L1/L2/L3 Triggers Preserved

The 3-level compression triggers SHALL remain identical:
- L1: Interaction count exceeds ToolCallRetention
- L2: Message count exceeds MaxMessages after L1
- L3: Token count exceeds MaxTokens after L2

#### Scenario: L1 compression triggers correctly

Given session with completed interactions > ToolCallRetention
When calling Compress()
Then L1 compression replaces older interactions with summaries

#### Scenario: L2 compression triggers correctly

Given session with messages > MaxMessages after L1
When calling Compress()
Then L2 truncation removes oldest messages keeping recent

#### Scenario: L3 compression triggers correctly

Given session with tokens > MaxTokens after L1+L2
When calling Compress()
Then L3 summary generation reduces token count

---

### Requirement: Unified Interaction Struct

Session SHALL use a single `Interaction` struct, not duplicated in context.go and interaction.go.

#### Scenario: Interaction defined once

Given session module
When searching for Interaction struct definition
Then only one definition exists
