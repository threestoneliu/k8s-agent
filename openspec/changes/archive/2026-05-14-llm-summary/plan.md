# LLM Summary Implementation Plan

> **For agentic workers:** Use superpowers:subagent-driven-development to implement this plan task-by-task.

**Goal:** Replace lightweight keyword-based summarization with LLM-generated conversational summaries in Level 3 context compression.

**Architecture:** Agent coordinates LLM summary generation by calling ContextManager which returns `needsSummary=true` when Level 3 is triggered. Agent then calls llmSvc.Chat() to generate the summary, and re-invokes ContextManager with the summary result.

**Tech Stack:** Go, OpenAI SDK, k8s-agent codebase

---

## Task 1: Modify ContextManager.BuildContextMessages() Signature

**Files:**
- Modify: `pkg/session/context.go:68-90` (signature)
- Modify: `pkg/session/context.go:155-207` (Level 3 logic)
- Test: `pkg/session/context_test.go`

- [ ] **Step 1: Write failing test for new return values**

Add test in `pkg/session/context_test.go`:

```go
func TestBuildContextMessages_NeedsSummary(t *testing.T) {
    config := cluster.ContextConfig{
        MaxMessages:      5,
        MaxTokens:       1000,
        SummaryEnabled:  true,
        ToolCallRetention: 3,
    }
    cm := NewContextManager(config)

    // Create messages that will trigger Level 3
    messages := make([]*Message, 20)
    for i := range messages {
        messages[i] = &Message{
            Role:    RoleUser,
            Content: fmt.Sprintf("User query number %d", i),
        }
    }

    llmMessages, needsSummary, rawForSummary := cm.BuildContextMessages("system prompt", messages, "")

    // Verify needsSummary is true when Level 3 is triggered
    if !needsSummary {
        t.Error("Expected needsSummary=true for Level 3 compression")
    }
    if len(rawForSummary) == 0 {
        t.Error("Expected rawForSummary to contain messages for summarization")
    }
}
```

Run: `go test ./pkg/session/... -run TestBuildContextMessages_NeedsSummary -v`
Expected: FAIL (method doesn't return these values yet)

- [ ] **Step 2: Update BuildContextMessages signature to return three values**

Modify `pkg/session/context.go:68-73`:

```go
func (cm *ContextManager) BuildContextMessages(
    systemPrompt string,
    messages []*Message,
    summaryPrompt string,
) (llmMessages []sharedutil.Message, needsSummary bool, rawForSummary []*Message) {
    llmMessages = []sharedutil.Message{
        {Role: "system", Content: systemPrompt},
    }
    // ... rest of Level 0/1/2 logic unchanged, but return early with (result, false, nil)
    // ...
}
```

Update Level 0 (line 78-90) to return early:
```go
if len(messages) <= cm.config.MaxMessages {
    estimatedTokens := estimateMessagesTokens(messages)
    systemTokens := estimateTokens(systemPrompt)
    if systemTokens+estimatedTokens <= cm.config.MaxTokens {
        for _, m := range messages {
            llmMessages = append(llmMessages, sharedutil.Message{
                Role:    string(m.Role),
                Content: m.Content,
            })
        }
        return llmMessages, false, nil
    }
}
```

- [ ] **Step 3: Update Level 1 return values**

After Level 1 compression (around line 124), add:
```go
// After existing Level 1 check, before Level 2:
return llmMessages, false, nil
```

- [ ] **Step 4: Modify Level 3 to return needsSummary=true**

In `pkg/session/context.go:155-207`, change Level 3 logic to:

```go
// Level 3: Deep compression - return needsSummary signal
if cm.config.SummaryEnabled {
    // Return needsSummary=true and all messages for LLM summarization
    return llmMessages, true, messages
}
```

Also update the else branch to return proper values.

- [ ] **Step 5: Run test to verify new signature works**

Run: `go test ./pkg/session/... -run TestBuildContextMessages_NeedsSummary -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add pkg/session/context.go pkg/session/context_test.go
git commit -m "feat(session): modify BuildContextMessages to return needsSummary and rawForSummary"
```

---

## Task 2: Add generateLLMSummary() to Agent

**Files:**
- Modify: `pkg/agent/agent.go` (add methods around line 300)
- Test: `pkg/agent/agent_test.go`

- [ ] **Step 1: Write failing test for generateLLMSummary**

Add test in `pkg/agent/agent_test.go`:

```go
func TestAgent_GenerateLLMSummary(t *testing.T) {
    // Skip if no API key
    if os.Getenv("OPENAI_API_KEY") == "" {
        t.Skip("OPENAI_API_KEY not set")
    }

    // Create agent with mock LLM service
    cfg := &llm.LLMConfig{APIKey: os.Getenv("OPENAI_API_KEY"), Model: "gpt-4"}
    llmSvc := llm.NewService(cfg)
    agent := &Agent{llmSvc: llmSvc}

    messages := []*session.Message{
        session.NewMessage(session.RoleUser, "list pods in default", nil),
        session.NewMessage(session.RoleAssistant, "执行工具: resource_list", nil),
    }

    summary, err := agent.generateLLMSummary(messages)
    if err != nil {
        t.Errorf("generateLLMSummary failed: %v", err)
    }
    if len(summary) == 0 {
        t.Error("Expected non-empty summary")
    }
    // Verify summary mentions cluster or operations
    if !strings.Contains(strings.ToLower(summary), "cluster") &&
       !strings.Contains(strings.ToLower(summary), "default") {
        t.Logf("Summary: %s", summary) // Just log, not fail
    }
}
```

Run: `go test ./pkg/agent/... -run TestAgent_GenerateLLMSummary -v`
Expected: FAIL (method doesn't exist)

- [ ] **Step 2: Add generateLLMSummary method**

Add to `pkg/agent/agent.go` (around line 300):

```go
// generateLLMSummary generates a conversational summary using LLM
func (a *Agent) generateLLMSummary(messages []*session.Message) (string, error) {
    if a.llmSvc == nil {
        return "", fmt.Errorf("LLM service not available")
    }

    prompt := a.buildSummaryPrompt(messages)
    resp, err := a.llmSvc.Chat(context.Background(), []sharedutil.Message{
        {Role: "user", Content: prompt},
    })
    if err != nil {
        return "", fmt.Errorf("LLM summary failed: %w", err)
    }
    return resp, nil
}

// buildSummaryPrompt constructs the prompt for LLM summarization
func (a *Agent) buildSummaryPrompt(messages []*session.Message) string {
    var sb strings.Builder
    sb.WriteString("请用简洁的自然语言总结以下会话，风格如同助手在回顾对话。")
    sb.WriteString("重点描述：涉及的集群、操作类型、资源类型、命名空间、是否有危险操作。\n\n")

    for _, m := range messages {
        role := "用户"
        if m.Role == session.RoleAssistant {
            role = "助手"
        }
        sb.WriteString(fmt.Sprintf("%s: %s\n", role, m.Content))
    }

    sb.WriteString("\n请生成一段简洁的对话式摘要（中文，100字以内）。")
    return sb.String()
}
```

- [ ] **Step 3: Add fallback to lightweight summary**

Update `generateLLMSummary` to handle error gracefully:

```go
func (a *Agent) generateLLMSummary(messages []*session.Message) (string, error) {
    if a.llmSvc == nil {
        return a.fallbackSummary(messages), nil
    }

    prompt := a.buildSummaryPrompt(messages)
    resp, err := a.llmSvc.Chat(context.Background(), []sharedutil.Message{
        {Role: "user", Content: prompt},
    })
    if err != nil {
        log.Warn("LLM summary failed, using fallback", "error", err)
        return a.fallbackSummary(messages), nil
    }
    return resp, nil
}

// fallbackSummary uses the old lightweight keyword-based summary
func (a *Agent) fallbackSummary(messages []*session.Message) string {
    // Use ContextManager's generateSummary for fallback
    ctxMgr := a.ctxManager
    if ctxMgr == nil {
        return "会话摘要不可用"
    }
    return ctxMgr.GenerateFallbackSummary(messages)
}
```

- [ ] **Step 4: Run test to verify**

Run: `go test ./pkg/agent/... -run TestAgent_GenerateLLMSummary -v`
Expected: PASS (or SKIP if no API key)

- [ ] **Step 5: Commit**

```bash
git add pkg/agent/agent.go pkg/agent/agent_test.go
git commit -m "feat(agent): add generateLLMSummary method with fallback"
```

---

## Task 3: Update Agent to Use New BuildContextMessages API

**Files:**
- Modify: `pkg/agent/agent.go` (processWithOutput around line 275)
- Modify: `pkg/session/context.go` (add GenerateFallbackSummary public method)

- [ ] **Step 1: Write failing test for full integration**

Add test in `pkg/agent/agent_test.go`:

```go
func TestAgent_BuildContextMessages_WithLLMSummary(t *testing.T) {
    if os.Getenv("OPENAI_API_KEY") == "" {
        t.Skip("OPENAI_API_KEY not set")
    }

    cfg := &llm.LLMConfig{APIKey: os.Getenv("OPENAI_API_KEY"), Model: "gpt-4"}
    llmSvc := llm.NewService(cfg)

    ctxCfg := cluster.ContextConfig{
        MaxMessages:       5,
        MaxTokens:        2000,
        SummaryEnabled:   true,
        ToolCallRetention: 3,
    }
    ctxMgr := session.NewContextManager(ctxCfg)

    agent := &Agent{
        llmSvc:   llmSvc,
        ctxManager: ctxMgr,
    }

    // Create enough messages to trigger Level 3
    messages := make([]*session.Message, 20)
    for i := range messages {
        messages[i] = session.NewMessage(session.RoleUser, fmt.Sprintf("query %d", i), nil)
    }

    result, err := agent.BuildContextMessagesWithSummary("system prompt", messages)
    if err != nil {
        t.Errorf("BuildContextMessagesWithSummary failed: %v", err)
    }
    // Verify result contains summary message
    hasSummary := false
    for _, msg := range result {
        if strings.Contains(msg.Content, "[Previous conversation summary]") {
            hasSummary = true
            break
        }
    }
    if !hasSummary {
        t.Error("Expected summary message in result")
    }
}
```

Run: `go test ./pkg/agent/... -run TestAgent_BuildContextMessages_WithLLMSummary -v`
Expected: FAIL (method doesn't exist)

- [ ] **Step 2: Add BuildContextMessagesWithSummary to Agent**

Add to `pkg/agent/agent.go`:

```go
// BuildContextMessagesWithSummary builds context messages with LLM summary if needed
func (a *Agent) BuildContextMessagesWithSummary(systemPrompt string, messages []*session.Message) ([]sharedutil.Message, error) {
    if a.ctxManager == nil {
        return nil, fmt.Errorf("context manager not available")
    }

    ctxMgr := a.ctxManager
    llmMessages, needsSummary, rawForSummary := ctxMgr.BuildContextMessages(systemPrompt, messages, "")

    if !needsSummary {
        return llmMessages, nil
    }

    // Generate LLM summary
    summary, err := a.generateLLMSummary(rawForSummary)
    if err != nil {
        return nil, err
    }

    // Rebuild with summary - create new messages with summary included
    result := []sharedutil.Message{
        {Role: "system", Content: systemPrompt},
        {Role: "system", Content: "[Previous conversation summary]: " + summary},
    }

    // Keep recent messages as "recent context"
    windowSize := ctxMgr.GetConfig().ToolCallRetention
    if windowSize <= 0 {
        windowSize = 5
    }
    startIdx := len(messages) - windowSize
    if startIdx < 0 {
        startIdx = 0
    }

    result = append(result, sharedutil.Message{
        Role:    "system",
        Content: fmt.Sprintf("[%d earlier messages have been summarized]", startIdx),
    })

    for i := startIdx; i < len(messages); i++ {
        result = append(result, sharedutil.Message{
            Role:    string(messages[i].Role),
            Content: messages[i].Content,
        })
    }

    return result, nil
}
```

- [ ] **Step 3: Expose GenerateFallbackSummary in ContextManager**

Add to `pkg/session/context.go`:

```go
// GenerateFallbackSummary provides the old lightweight summary as fallback
func (cm *ContextManager) GenerateFallbackSummary(messages []*Message) string {
    return cm.generateSummary(messages, "")
}
```

Also add getter for config:

```go
// GetConfig returns the context manager configuration
func (cm *ContextManager) GetConfig() cluster.ContextConfig {
    return cm.config
}
```

- [ ] **Step 4: Update processWithOutput to use new method**

In `pkg/agent/agent.go:275-278`, replace:

```go
// OLD:
if a.ctxManager != nil && len(a.llmMessages) > a.ctxManager.MessageCount() {
    a.llmMessages = a.ctxManager.CompressMessages(a.llmMessages)
}

// NEW:
if a.ctxManager != nil && len(a.llmMessages) > 5 { // arbitrary threshold before Level 3
    llmMsgs, err := a.BuildContextMessagesWithSummary(systemPrompt, a.messages)
    if err == nil {
        a.llmMessages = llmMsgs
    }
}
```

- [ ] **Step 5: Run integration test**

Run: `go test ./pkg/agent/... -run TestAgent_BuildContextMessages_WithLLMSummary -v`
Expected: PASS (or SKIP if no API key)

- [ ] **Step 6: Commit**

```bash
git add pkg/agent/agent.go pkg/session/context.go pkg/agent/agent_test.go
git commit -m "feat(agent): integrate LLM summary into context building"
```

---

## Task 4: Update All Tests

**Files:**
- Test: `pkg/session/context_test.go`
- Test: `pkg/agent/agent_test.go`

- [ ] **Step 1: Update session/context_test.go for new signature**

Run existing tests:
```bash
go test ./pkg/session/... -v 2>&1 | head -50
```

Fix any compilation errors or test failures by updating test assertions to match new return values.

- [ ] **Step 2: Update agent tests**

Run:
```bash
go test ./pkg/agent/... -v 2>&1 | head -50
```

Fix any failures related to signature changes.

- [ ] **Step 3: Run all tests**

```bash
go test ./pkg/session/... ./pkg/agent/... -v
```

Expected: All pass (or some skip due to no API key)

- [ ] **Step 4: Commit**

```bash
git add pkg/session/context_test.go pkg/agent/agent_test.go
git commit -m "test: update tests for LLM summary integration"
```

---

## Task 5: Final Verification

- [ ] **Step 1: Run build**

```bash
go build ./...
```

Expected: Success

- [ ] **Step 2: Run all tests**

```bash
go test ./...
```

Expected: All pass (some may skip if no API key)

- [ ] **Step 3: Commit**

```bash
git commit -m "chore: final verification for LLM summary feature"
```