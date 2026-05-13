# Brainstorm Summary

## Problem Statement

The current `pkg/session/context.go` implementation has complex level1 compression logic that is hard to understand and maintain. Issues:

- ~730 lines in a single file with too many responsibilities
- Two parallel implementations for `*Message` and `llm.Message` formats
- `findCompleteInteractions` has deeply nested if-else branches
- `ToolCallRetention` parameter name is semantically misaligned with actual behavior
- Bug: premature compression when interactions were within limits

## Key Decisions

1. **仅 Agent 内部使用** - For building LLM messages to avoid context window overflow
2. **需要占位符但内容简短** - Placeholder like `[N msgs + M tool calls condensed]` not verbose
3. **触发机制：消息数量 OR token 数量** - Either exceeds limit triggers compression
4. **重新设计分层** - Not by level (L1/L2/L3) but by interaction completeness
5. **未完成的交互不压缩** - Wait until interaction truly completes (receives summary)
6. **按消息内容有效性压缩** - Only keep messages with meaningful content
7. **保留 tool call 标记** - Rebuild compressed interactions as `[Tool: tool_name]`
8. **优先完整性** - First ensure interactions are complete, then truncate if needed
9. **保留原始消息用于未压缩交互** - Recent interactions (within retention) stay fully intact

## Design

### 1. Core Data Structure

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
```

**Why store OriginalMessages?**
- Recent interactions (within ToolCallRetention) are kept fully intact
- Original messages are preserved to maintain exact LLM message format
- Only old interactions beyond retention limit get compressed/rebuilt

### 2. Compression Flow

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

### 3. Interaction → Message Reconstruction

When an old interaction is compressed, it is rebuilt as:

```
User query:        {Role: "user", Content: interaction.Query}
Tool call marker:  {Role: "assistant", Content: "[Tool: tool_name1]"}
                   {Role: "assistant", Content: "[Tool: tool_name2]"}  // if multiple tools
Final summary:     {Role: "assistant", Content: interaction.Summary}
```

### 4. Example

**Original messages (3 interactions):**
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

### 5. Placeholder Format

Short and concise:
- `[N msgs + M tool calls condensed]` - when old interactions were compressed
- `[N msgs truncated]` - when message count was truncated at Level 2
- `[conversation summarized]` - when Level 3 summary was generated

## Alternatives Considered

### 方案 B：Token-aware two-stage compression

- **Approach**: First compress by interaction, then truncate by token
- **Pros**: More precise token control
- **Cons**: Two-stage logic may interfere with each other
- **Why not chosen**: More complex without clear benefit given current usage

### 方案 C：Discard tool call markers entirely

- **Approach**: Compressed interactions only keep query + summary
- **Pros**: Simpler structure
- **Cons**: LLM loses awareness that tools were executed
- **Why not chosen**: Tool execution awareness is important context for LLM

## Agreed Approach

**方案 A：交互完整性优先压缩 with original message preservation**

- Core idea: Group by interaction, each group is either kept complete or compressed
- Recent interactions (within ToolCallRetention) stay fully intact with original messages
- Old interactions are rebuilt as simplified `[Tool: name]` + summary format
- Placeholder indicates what was compressed
- Logic is clear and matches the "interaction completeness" semantic