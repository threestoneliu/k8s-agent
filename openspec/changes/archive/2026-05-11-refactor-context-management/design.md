# Design: Context Management Refactoring

## Overview

Refactor `pkg/session/context.go` to use interaction-based compression instead of level-based compression. The goal is clearer logic that's easier to understand and maintain.

## Problem

Current implementation:
- ~730 lines in a single file
- Level-based compression (L1/L2/L3) with overlapping responsibilities
- Two parallel implementations for `*Message` and `llm.Message`
- Complex `findCompleteInteractions` with nested conditionals
- Bug: premature compression when interactions were within limits

## Solution

### Core Data Structure

```go
// Interaction represents a user-LLM interaction (internal intermediate representation)
type Interaction struct {
    Query     string    // user query
    ToolNames []string  // e.g., ["k8s_delete", "k8s_get"] - extracted from tool calls
    Summary   string    // final summary (assistant message that is not a tool call)
    Completed bool      // whether tool result and summary have been received

    // Original messages - preserved for uncompressed recent interactions
    OriginalMessages []*Message
}

// Message is the session's internal message format
type Message struct {
    Role    Role
    Content string
}
```

**Why store OriginalMessages?**
- Recent interactions (within ToolCallRetention) are kept fully intact
- Original messages are preserved to maintain exact LLM message format
- Only old interactions beyond retention limit get compressed/rebuilt

### Compression Flow

```
Input: []*Message (from session)

1. Parse: Convert []*Message → []Interaction
   - Scan messages sequentially
   - user message → start new Interaction
   - tool call (contains "[Function Call:" or "[Tool Call:") → extract tool name
   - tool result (contains "[Tool:" or "Result:") → mark Interaction as Completed
   - assistant message that is NOT a tool call → this is the Summary

2. Evaluate: Check if compression is needed
   - Count completed interactions
   - if interaction_count <= ToolCallRetention:
       return original messages (no compression needed)

3. Compress:
   - For each Interaction (oldest to newest):
     - if interaction_index < len(interactions) - ToolCallRetention:
         // Old interaction - rebuild with simplified tool call markers
         [user query, "[Tool: tool_name]", summary]
     - else:
         // Recent interaction - keep original messages intact
         OriginalMessages

4. Output: []*Message (compressed)
   - Add placeholder if any old interactions were compressed:
     "[N msgs + M tool calls condensed]"
```

### Interaction → Message Reconstruction

When an old interaction is compressed, it is rebuilt as:

```
User query:        {Role: "user", Content: interaction.Query}
Tool call marker:  {Role: "assistant", Content: "[Tool: tool_name1]"}
                   {Role: "assistant", Content: "[Tool: tool_name2]"}  // if multiple tools
Final summary:     {Role: "assistant", Content: interaction.Summary}
```

Multiple tool names in one interaction result in multiple `[Tool: name]` markers.

### Configuration

```go
type ContextConfig struct {
    MaxMessages       int  // max messages before compression triggers
    MaxTokens         int  // max tokens before compression triggers
    ToolCallRetention int  // number of RECENT complete interactions to keep fully intact
    SummaryEnabled    bool // enable Level 3 summary generation
}
```

### Key Functions

| Function | Purpose |
|----------|---------|
| `ParseToInteractions([]*Message) []Interaction` | Convert message list to interaction list |
| `CompressInteractions([]Interaction) []*Message` | Apply interaction-based compression |
| `ShouldCompress([]Interaction) bool` | Check if compression is needed |
| `ReconstructInteraction(Interaction) []*Message` | Rebuild compressed interaction as messages |
| `AddPlaceholder([]*Message, int, int) []*Message` | Add condensed message placeholder |

### File Structure

```
pkg/session/
├── context.go        # ContextManager, BuildContextMessages (refactored)
├── interaction.go    # Interaction struct, parsing logic (NEW)
├── compressor.go     # Compression logic (NEW)
├── context_test.go   # Tests (updated)
└── session.go        # Session management (existing)
```

### Placeholder Format

Short and concise:
- `[N msgs + M tool calls condensed]` - when old interactions were compressed
- `[N msgs truncated]` - when message count was truncated at Level 2
- `[conversation summarized]` - when Level 3 summary was generated

### Example

**Original messages:**
```
[0] user: "delete pod nginx"
[1] assistant: "[Function Call: k8s_delete(name='nginx')]"
[2] assistant: "[Tool:result]"
[3] assistant: "已成功删除 nginx pod"

[4] user: "get pods"
[5] assistant: "[Function Call: k8s_get(ns='default')]"
[6] assistant: "[Tool:result]"
[7] assistant: "当前有 5 个 pods"

[8] user: "describe pod web"
[9] assistant: "[Function Call: k8s_describe(name='web')]"
[10] assistant: "[Tool:result]"
[11] assistant: "web pod info..."
```

**ToolCallRetention = 2, 3 interactions total**

Compression result (interaction 1 is compressed, 2 and 3 kept intact):
```
[0] user: "delete pod nginx"
[1] assistant: "[Tool: k8s_delete]"
[2] assistant: "已成功删除 nginx pod"
[3] user: "get pods"
[4] assistant: "[Function Call: k8s_get(ns='default')]"
[5] assistant: "[Tool:result]"
[6] assistant: "当前有 5 个 pods"
[7] user: "describe pod web"
[8] assistant: "[Function Call: k8s_describe(name='web')]"
[9] assistant: "[Tool:result]"
[10] assistant: "web pod info..."
[11] system: "[2 msgs + 1 tool calls condensed]"
```

## Testing Strategy

1. **Unit tests for parsing**:
   - Single complete interaction
   - Multiple complete interactions
   - Incomplete (in-progress) interactions
   - Mixed complete and incomplete

2. **Unit tests for compression**:
   - Interactions within limits (no compression)
   - Old interactions compressed, recent kept intact
   - Correct placeholder content
   - Correct message order after compression

3. **Integration tests**:
   - Full `BuildContextMessages` flow
   - Token limit enforcement
   - Message count limit enforcement

## Implementation Order

1. Create `interaction.go` with Interaction struct and parsing logic
2. Create `compressor.go` with compression logic
3. Refactor `context.go` to use new modules
4. Add/update unit tests
5. Verify all existing tests pass