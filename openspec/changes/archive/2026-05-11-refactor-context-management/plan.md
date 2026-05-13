# Context Management Refactoring Implementation Plan

> **For agentic workers:** Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Refactor `pkg/session/context.go` to use interaction-based compression with clear Parse → Evaluate → Compress → Output flow, replacing the complex level-based L1/L2/L3 compression.

**Architecture:** New `Interaction` struct serves as intermediate representation. Two new files (`interaction.go`, `compressor.go`) handle parsing and compression respectively. `context.go` refactored to use them.

**Tech Stack:** Go, k8s-agent codebase

---

## Task 1: Create Interaction Struct and Parsing (`interaction.go`)

**Files:**
- Create: `pkg/session/interaction.go`
- Test: `pkg/session/interaction_test.go`

- [ ] **Step 1: Create `pkg/session/interaction.go` with empty struct**

```go
package session

// Interaction represents a user-LLM interaction
type Interaction struct {
    Query            string
    ToolNames        []string
    Summary          string
    Completed        bool
    OriginalMessages []*Message
}

// Message represents session's internal message format
type Message struct {
    Role    Role
    Content string
}
```

- [ ] **Step 2: Create `pkg/session/interaction_test.go` with first failing test**

```go
package session

import "testing"

func TestInteractionStruct(t *testing.T) {
    inter := Interaction{
        Query:     "delete pod nginx",
        ToolNames: []string{"k8s_delete"},
        Summary:   "pod deleted",
        Completed: true,
    }
    if inter.Query != "delete pod nginx" {
        t.Errorf("expected query, got %s", inter.Query)
    }
}
```

- [ ] **Step 3: Run test to verify it passes**

Run: `go test ./pkg/session/... -run TestInteractionStruct -v`
Expected: PASS

- [ ] **Step 4: Add tool name extraction function**

Add to `interaction.go`:

```go
// extractToolName extracts tool name from "[Function Call: k8s_delete]" format
func extractToolName(content string) string {
    // This will be implemented in next step
    return ""
}
```

- [ ] **Step 5: Add test for tool name extraction, run to verify it fails**

```go
func TestExtractToolName(t *testing.T) {
    result := extractToolName("[Function Call: k8s_delete(name='nginx')]")
    if result != "k8s_delete" {
        t.Errorf("expected 'k8s_delete', got '%s'", result)
    }
}
```

Run: `go test ./pkg/session/... -run TestExtractToolName -v`
Expected: FAIL with "expected 'k8s_delete', got ''"

- [ ] **Step 6: Implement extractToolName**

```go
func extractToolName(content string) string {
    const prefix = "[Function Call:"
    idx := strings.Index(content, prefix)
    if idx == -1 {
        idx = strings.Index(content, "[Tool Call:")
        if idx == -1 {
            return ""
        }
        prefix = "[Tool Call:"
    }
    start := idx + len(prefix)
    end := strings.Index(content[start:], "]")
    if end == -1 {
        return ""
    }
    name := strings.TrimSpace(content[start : start+end])
    // Remove parentheses if present: "k8s_delete(name='nginx')" -> "k8s_delete"
    if parenIdx := strings.Index(name, "("); parenIdx != -1 {
        name = name[:parenIdx]
    }
    return name
}
```

- [ ] **Step 7: Run test to verify it passes**

Run: `go test ./pkg/session/... -run TestExtractToolName -v`
Expected: PASS

- [ ] **Step 8: Add helper functions for tool detection**

Add to `interaction.go`:

```go
func isToolCallMessage(msg *Message) bool {
    return strings.Contains(msg.Content, "[Function Call:") ||
        strings.Contains(msg.Content, "[Tool Call:")
}

func isToolResultMessage(msg *Message) bool {
    return strings.Contains(msg.Content, "[Tool:") ||
        strings.Contains(msg.Content, "Result:")
}
```

- [ ] **Step 9: Add ParseToInteractions test, verify it fails**

```go
func TestParseToInteractions_SingleComplete(t *testing.T) {
    messages := []*Message{
        {Role: RoleUser, Content: "delete pod nginx"},
        {Role: RoleAssistant, Content: "[Function Call: k8s_delete(name='nginx')]"},
        {Role: RoleAssistant, Content: "[Tool:result]"},
        {Role: RoleAssistant, Content: "已成功删除"},
    }
    interactions := ParseToInteractions(messages)
    if len(interactions) != 1 {
        t.Fatalf("expected 1 interaction, got %d", len(interactions))
    }
    if interactions[0].Query != "delete pod nginx" {
        t.Errorf("expected query 'delete pod nginx', got '%s'", interactions[0].Query)
    }
    if interactions[0].ToolNames[0] != "k8s_delete" {
        t.Errorf("expected tool 'k8s_delete', got '%s'", interactions[0].ToolNames[0])
    }
    if !interactions[0].Completed {
        t.Error("expected interaction to be completed")
    }
}
```

Run: `go test ./pkg/session/... -run TestParseToInteractions_SingleComplete -v`
Expected: FAIL - function not defined

- [ ] **Step 10: Add ParseToInteractions function stub**

```go
func ParseToInteractions(messages []*Message) []Interaction {
    return nil
}
```

Run: `go test ./pkg/session/... -run TestParseToInteractions_SingleComplete -v`
Expected: FAIL - interaction not complete

- [ ] **Step 11: Implement ParseToInteractions**

```go
func ParseToInteractions(messages []*Message) []Interaction {
    var interactions []Interaction
    var current *Interaction

    for i := range messages {
        msg := messages[i]
        if msg.Role == RoleUser {
            if current != nil {
                interactions = append(interactions, *current)
            }
            current = &Interaction{
                Query: msg.Content,
            }
            current.OriginalMessages = append(current.OriginalMessages, msg)
        } else if msg.Role == RoleAssistant {
            if isToolCallMessage(msg) {
                if current == nil {
                    current = &Interaction{}
                }
                if toolName := extractToolName(msg.Content); toolName != "" {
                    current.ToolNames = append(current.ToolNames, toolName)
                }
                current.OriginalMessages = append(current.OriginalMessages, msg)
            } else if current != nil {
                // Final summary (assistant but not tool call)
                current.Summary = msg.Content
                current.OriginalMessages = append(current.OriginalMessages, msg)
            }
        } else if isToolResultMessage(msg) {
            if current == nil {
                current = &Interaction{}
            }
            current.Completed = true
            current.OriginalMessages = append(current.OriginalMessages, msg)
        }
    }

    if current != nil {
        interactions = append(interactions, *current)
    }

    return interactions
}
```

- [ ] **Step 12: Run test to verify it passes**

Run: `go test ./pkg/session/... -run TestParseToInteractions_SingleComplete -v`
Expected: PASS

- [ ] **Step 13: Add test for multiple interactions**

```go
func TestParseToInteractions_Multiple(t *testing.T) {
    messages := []*Message{
        {Role: RoleUser, Content: "query1"},
        {Role: RoleAssistant, Content: "[Function Call: k8s_delete]"},
        {Role: RoleAssistant, Content: "[Tool:result]"},
        {Role: RoleAssistant, Content: "summary1"},
        {Role: RoleUser, Content: "query2"},
        {Role: RoleAssistant, Content: "[Function Call: k8s_get]"},
        {Role: RoleAssistant, Content: "[Tool:result]"},
        {Role: RoleAssistant, Content: "summary2"},
    }
    interactions := ParseToInteractions(messages)
    if len(interactions) != 2 {
        t.Fatalf("expected 2 interactions, got %d", len(interactions))
    }
    if interactions[0].Query != "query1" {
        t.Errorf("expected 'query1', got '%s'", interactions[0].Query)
    }
    if interactions[1].Query != "query2" {
        t.Errorf("expected 'query2', got '%s'", interactions[1].Query)
    }
}
```

Run: `go test ./pkg/session/... -run TestParseToInteractions_Multiple -v`
Expected: PASS

- [ ] **Step 14: Add test for incomplete interaction**

```go
func TestParseToInteractions_Incomplete(t *testing.T) {
    messages := []*Message{
        {Role: RoleUser, Content: "query1"},
        {Role: RoleAssistant, Content: "[Function Call: k8s_delete]"},
        // No tool result or summary yet
    }
    interactions := ParseToInteractions(messages)
    if len(interactions) != 1 {
        t.Fatalf("expected 1 interaction, got %d", len(interactions))
    }
    if interactions[0].Completed {
        t.Error("incomplete interaction should not be marked completed")
    }
    if interactions[0].Summary != "" {
        t.Errorf("expected empty summary, got '%s'", interactions[0].Summary)
    }
}
```

Run: `go test ./pkg/session/... -run TestParseToInteractions_Incomplete -v`
Expected: PASS

- [ ] **Step 15: Commit**

```bash
git add pkg/session/interaction.go pkg/session/interaction_test.go
git commit -m "feat(session): add Interaction struct and ParseToInteractions
- Add Interaction struct with Query, ToolNames, Summary, Completed, OriginalMessages
- Add extractToolName helper to parse [Function Call: xxx] format
- Add ParseToInteractions to convert []*Message to []Interaction
- Add isToolCallMessage and isToolResultMessage helpers
- Add tests for single, multiple, and incomplete interactions"
```

---

## Task 2: Create Compression Logic (`compressor.go`)

**Files:**
- Create: `pkg/session/compressor.go`
- Test: `pkg/session/compressor_test.go`

- [ ] **Step 1: Create `pkg/session/compressor.go` with stub functions**

```go
package session

// ShouldCompress checks if interactions need compression
func ShouldCompress(interactions []Interaction, retentionLimit int) bool {
    return len(interactions) > retentionLimit
}
```

- [ ] **Step 2: Create `pkg/session/compressor_test.go` with failing test**

```go
package session

import "testing"

func TestShouldCompress(t *testing.T) {
    interactions := []Interaction{
        {Query: "query1"},
        {Query: "query2"},
        {Query: "query3"},
    }
    // 3 interactions, retention 2 -> should compress
    if !ShouldCompress(interactions, 2) {
        t.Error("expected ShouldCompress to return true")
    }
    // 2 interactions, retention 2 -> should NOT compress
    if ShouldCompress(interactions[:2], 2) {
        t.Error("expected ShouldCompress to return false")
    }
}
```

- [ ] **Step 3: Run test to verify it passes**

Run: `go test ./pkg/session/... -run TestShouldCompress -v`
Expected: PASS

- [ ] **Step 4: Add ReconstructInteraction test, verify it fails**

```go
func TestReconstructInteraction(t *testing.T) {
    inter := Interaction{
        Query:     "delete pod nginx",
        ToolNames: []string{"k8s_delete"},
        Summary:   "pod deleted successfully",
        Completed: true,
    }
    result := ReconstructInteraction(inter)
    // Should produce: [user query, [Tool: k8s_delete], summary]
    if len(result) != 3 {
        t.Fatalf("expected 3 messages, got %d", len(result))
    }
    if result[0].Role != RoleUser {
        t.Errorf("expected user role, got %v", result[0].Role)
    }
    if result[0].Content != "delete pod nginx" {
        t.Errorf("expected query content, got '%s'", result[0].Content)
    }
    if result[1].Content != "[Tool: k8s_delete]" {
        t.Errorf("expected '[Tool: k8s_delete]', got '%s'", result[1].Content)
    }
    if result[2].Content != "pod deleted successfully" {
        t.Errorf("expected summary, got '%s'", result[2].Content)
    }
}
```

Run: `go test ./pkg/session/... -run TestReconstructInteraction -v`
Expected: FAIL - function not defined

- [ ] **Step 5: Implement ReconstructInteraction**

```go
func ReconstructInteraction(inter Interaction) []*Message {
    var messages []*Message
    messages = append(messages, &Message{
        Role:    RoleUser,
        Content: inter.Query,
    })
    for _, toolName := range inter.ToolNames {
        messages = append(messages, &Message{
            Role:    RoleAssistant,
            Content: "[Tool: " + toolName + "]",
        })
    }
    if inter.Summary != "" {
        messages = append(messages, &Message{
            Role:    RoleAssistant,
            Content: inter.Summary,
        })
    }
    return messages
}
```

- [ ] **Step 6: Run test to verify it passes**

Run: `go test ./pkg/session/... -run TestReconstructInteraction -v`
Expected: PASS

- [ ] **Step 7: Add CompressInteractions test, verify it fails**

```go
func TestCompressInteractions_Basic(t *testing.T) {
    interactions := []Interaction{
        {Query: "query1", ToolNames: []string{"tool1"}, Summary: "sum1", Completed: true},
        {Query: "query2", ToolNames: []string{"tool2"}, Summary: "sum2", Completed: true},
        {Query: "query3", ToolNames: []string{"tool3"}, Summary: "sum3", Completed: true},
    }
    // retention=2, 3 interactions -> only first should be compressed
    result, count := CompressInteractions(interactions, 2)
    if count != 1 {
        t.Errorf("expected 1 compressed, got %d", count)
    }
    // interaction 1 compressed -> 3 messages (query + tool + summary)
    // interaction 2 and 3 kept intact -> depends on OriginalMessages
}
```

Run: `go test ./pkg/session/... -run TestCompressInteractions_Basic -v`
Expected: FAIL - function not defined

- [ ] **Step 8: Implement CompressInteractions**

```go
func CompressInteractions(interactions []Interaction, retentionLimit int) ([]*Message, int) {
    if len(interactions) <= retentionLimit {
        // No compression needed, return original messages
        var result []*Message
        for _, inter := range interactions {
            result = append(result, inter.OriginalMessages...)
        }
        return result, 0
    }

    var result []*Message
    compressedCount := 0

    for i := range interactions {
        isRecent := i >= len(interactions)-retentionLimit
        if isRecent {
            // Keep original messages intact
            result = append(result, interactions[i].OriginalMessages...)
        } else {
            // Compress: rebuild with simplified tool markers
            result = append(result, ReconstructInteraction(interactions[i])...)
            compressedCount++
        }
    }

    return result, compressedCount
}
```

- [ ] **Step 9: Run test to verify it passes**

Run: `go test ./pkg/session/... -run TestCompressInteractions_Basic -v`
Expected: PASS

- [ ] **Step 10: Add test for no compression when within limits**

```go
func TestCompressInteractions_WithinLimits(t *testing.T) {
    interactions := []Interaction{
        {Query: "query1", ToolNames: []string{"tool1"}, Summary: "sum1", Completed: true},
        {Query: "query2", ToolNames: []string{"tool2"}, Summary: "sum2", Completed: true},
    }
    // retention=2, 2 interactions -> no compression
    result, count := CompressInteractions(interactions, 2)
    if count != 0 {
        t.Errorf("expected 0 compressed, got %d", count)
    }
    if len(result) != len(interactions[0].OriginalMessages)*2 {
        t.Errorf("expected original messages preserved, got %d", len(result))
    }
}
```

Run: `go test ./pkg/session/... -run TestCompressInteractions_WithinLimits -v`
Expected: PASS

- [ ] **Step 11: Add AddPlaceholder test, verify it fails**

```go
func TestAddPlaceholder(t *testing.T) {
    messages := []*Message{
        {Role: RoleUser, Content: "query"},
        {Role: RoleAssistant, Content: "summary"},
    }
    result := AddPlaceholder(messages, 3, 2)
    // Should add system message at end
    if len(result) != 3 {
        t.Errorf("expected 3 messages, got %d", len(result))
    }
    if result[2].Role != RoleSystem {
        t.Errorf("expected system role, got %v", result[2].Role)
    }
    if result[2].Content != "[5 msgs + 2 tool calls condensed]" {
        t.Errorf("unexpected placeholder: '%s'", result[2].Content)
    }
}
```

Run: `go test ./pkg/session/... -run TestAddPlaceholder -v`
Expected: FAIL - function not defined

- [ ] **Step 12: Implement AddPlaceholder**

```go
func AddPlaceholder(messages []*Message, msgDropped, toolCallDropped int) []*Message {
    placeholder := &Message{
        Role:    RoleSystem,
        Content: fmt.Sprintf("[%d msgs + %d tool calls condensed]", msgDropped, toolCallDropped),
    }
    return append(messages, placeholder)
}
```

- [ ] **Step 13: Run test to verify it passes**

Run: `go test ./pkg/session/... -run TestAddPlaceholder -v`
Expected: PASS

- [ ] **Step 14: Commit**

```bash
git add pkg/session/compressor.go pkg/session/compressor_test.go
git commit -m "feat(session): add compression logic in compressor.go
- Add ShouldCompress to check if interaction count exceeds limit
- Add ReconstructInteraction to rebuild compressed interaction as messages
- Add CompressInteractions to apply interaction-based compression
- Add AddPlaceholder to add condensed message placeholder
- Add tests for compression behavior"
```

---

## Task 3: Refactor ContextManager (`context.go`)

**Files:**
- Modify: `pkg/session/context.go:68-200`
- Test: `pkg/session/context_test.go` (update existing)

- [ ] **Step 1: Read current context.go to understand what needs changing**

Review the `BuildContextMessages` function (lines 68-200) to see how it currently calls level1Compress/level2Compress.

- [ ] **Step 2: Update BuildContextMessages to use new compression pipeline**

Replace the level-based compression calls with:

```go
// Build context with interaction-based compression
interactions := ParseToInteractions(messages)
if ShouldCompress(interactions, cm.config.ToolCallRetention) {
    compressed, compressedCount := CompressInteractions(interactions, cm.config.ToolCallRetention)
    // Add placeholder if any interactions were compressed
    if compressedCount > 0 {
        result = append(result, llm.Message{
            Role:    "system",
            Content: fmt.Sprintf("[%d msgs + %d tool calls condensed]", compressedCount*3, compressedCount),
        })
    }
    // ... continue with token check
} else {
    // No compression needed
}
```

- [ ] **Step 3: Run tests to verify existing tests still pass**

Run: `go test ./pkg/session/... -v 2>&1 | head -50`
Expected: Most tests pass, some may need updates due to refactored code

- [ ] **Step 4: Update or remove old functions**

Remove (or deprecate):
- `findCompleteInteractions`
- `level1Compress`
- `level1CompressLLM`

These will cause compile errors if referenced - update the callers.

- [ ] **Step 5: Verify build passes**

Run: `go build ./...`
Expected: SUCCESS

- [ ] **Step 6: Run all tests**

Run: `go test ./... -v 2>&1 | tail -30`
Expected: All tests pass

- [ ] **Step 7: Commit**

```bash
git add pkg/session/context.go
git commit -m "refactor(session): update BuildContextMessages to use interaction-based compression
- Replace level-based compression with ParseToInteractions + CompressInteractions
- Add placeholder when old interactions are compressed
- Remove deprecated findCompleteInteractions, level1Compress, level1CompressLLM"
```

---

## Task 4: Update Unit Tests

**Files:**
- Modify: `pkg/session/context_test.go`

- [ ] **Step 1: Review existing tests that may need updates**

Look at tests that call `level1Compress` - these need to be rewritten to test the new flow.

- [ ] **Step 2: Add test for compression with OriginalMessages preserved**

```go
func TestCompressInteractions_OriginalMessagesPreserved(t *testing.T) {
    interactions := []Interaction{
        {
            Query:            "query1",
            ToolNames:        []string{"k8s_delete"},
            Summary:          "sum1",
            Completed:        true,
            OriginalMessages: []*Message{
                {Role: RoleUser, Content: "query1"},
                {Role: RoleAssistant, Content: "[Function Call: k8s_delete]"},
                {Role: RoleAssistant, Content: "[Tool:result]"},
                {Role: RoleAssistant, Content: "sum1"},
            },
        },
    }
    result, count := CompressInteractions(interactions, 1)
    // Within limits - should preserve original messages
    if count != 0 {
        t.Errorf("expected 0 compressed, got %d", count)
    }
    if len(result) != 4 {
        t.Errorf("expected 4 original messages, got %d", len(result))
    }
}
```

- [ ] **Step 3: Run test to verify it passes**

Run: `go test ./pkg/session/... -run TestCompressInteractions_OriginalMessagesPreserved -v`
Expected: PASS

- [ ] **Step 4: Add test for message order after compression**

```go
func TestCompression_MessageOrder(t *testing.T) {
    // Create interactions where interaction 1 is old (to be compressed)
    // and interaction 2 is recent (to be kept intact)
    interactions := []Interaction{
        {
            Query:     "old query",
            ToolNames: []string{"old_tool"},
            Summary:   "old summary",
            Completed: true,
        },
        {
            Query: "recent query",
            OriginalMessages: []*Message{
                {Role: RoleUser, Content: "recent query"},
                {Role: RoleAssistant, Content: "recent response"},
            },
        },
    }
    result, _ := CompressInteractions(interactions, 1)
    // First interaction compressed: [user query, [Tool: old_tool], summary] = 3
    // Second interaction kept: [user query, response] = 2
    if len(result) != 5 {
        t.Errorf("expected 5 messages, got %d", len(result))
    }
    // Verify order: old query first, then recent
    if result[0].Content != "old query" {
        t.Errorf("expected first message to be 'old query', got '%s'", result[0].Content)
    }
    if result[3].Content != "recent query" {
        t.Errorf("expected 4th message to be 'recent query', got '%s'", result[3].Content)
    }
}
```

- [ ] **Step 5: Run test to verify it passes**

Run: `go test ./pkg/session/... -run TestCompression_MessageOrder -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add pkg/session/context_test.go
git commit -m "test(session): add tests for interaction-based compression
- Add test for OriginalMessages preservation
- Add test for message order after compression
- Update existing tests to use new compression flow"
```

---

## Task 5: Final Verification

- [ ] **Step 1: Run all tests**

Run: `go test ./... -v 2>&1 | tail -50`
Expected: All tests pass

- [ ] **Step 2: Verify build**

Run: `go build -o k8s-agent ./cmd/k8s-agent`
Expected: SUCCESS

- [ ] **Step 3: Run race detection tests**

Run: `go test -race ./pkg/session/...`
Expected: No race conditions detected

- [ ] **Step 4: Final commit**

```bash
git add -A
git commit -m "chore: finalize context management refactoring
- All tests passing
- Build succeeds
- No race conditions"
```

---

## Self-Review Checklist

**Spec coverage:**
- [x] Interaction Parsing - Task 1
- [x] Interaction Completion Detection - Task 1
- [x] Compression Trigger - Task 2 (ShouldCompress)
- [x] Recent Interaction Preservation - Task 2 (CompressInteractions keeps OriginalMessages)
- [x] Old Interaction Compression - Task 2 (ReconstructInteraction)
- [x] Placeholder Addition - Task 2 (AddPlaceholder)
- [x] Incomplete Interaction Handling - Task 1 (parse leaves Completed=false)

**Placeholder scan:** No TBD/TODO found.

**Type consistency:** All function names and signatures match across tasks.

---

## Execution Options

**Plan complete and saved to `openspec/changes/refactor-context-management/plan.md`. Two execution options:**

**1. Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration

**2. Inline Execution** - Execute tasks in this session using `/opsx:apply`

**Which approach?**