# LLM Response Parser Implementation Plan

> **For agentic workers:** Use superpowers:subagent-driven-development to implement this plan task-by-task.

**Goal:** Move parseThinkTags from agent package to llm package as a ResponseParser interface.

**Architecture:** Define ResponseParser interface in llm package, implement OpenAI parser, expose via Service.ResponseParser() method.

**Tech Stack:** Go, k8s-agent codebase

---

## Task 1: Create pkg/llm/parser.go

**Files:**
- Create: `pkg/llm/parser.go`

- [ ] **Step 1: Create parser.go with TextPart and ResponseParser**

```go
package llm

// TextPart represents a part of LLM response text
type TextPart struct {
    IsThink  bool
    Content  string
}

// ResponseParser parses LLM output text
type ResponseParser interface {
    Parse(text string) []TextPart
}
```

- [ ] **Step 2: Implement OpenAIResponseParser**

```go
type OpenAIResponseParser struct{}

func (p *OpenAIResponseParser) Parse(text string) []TextPart {
    // Move logic from agent.parseThinkTags
}
```

Run: `go build ./pkg/llm/...`

- [ ] **Step 3: Commit**

```bash
git add pkg/llm/parser.go
git commit -m "feat(llm): add ResponseParser interface and OpenAI implementation"
```

---

## Task 2: Modify llm.Service

**Files:**
- Modify: `pkg/llm/llm.go`

- [ ] **Step 1: Add responseParser field to Service**

```go
type Service struct {
    client          *OpenAISDKProvider
    functions       []sharedutil.Function
    responseParser  ResponseParser
}
```

- [ ] **Step 2: Initialize parser in NewService**

```go
func NewService(cfg *LLMConfig) *Service {
    provider := NewOpenAISDKProvider(cfg)
    return &Service{
        client:          provider,
        functions:       getFunctions(),
        responseParser:  &OpenAIResponseParser{},
    }
}
```

- [ ] **Step 3: Add ResponseParser method**

```go
func (s *Service) ResponseParser() ResponseParser {
    return s.responseParser
}
```

Run: `go build ./pkg/llm/...`

- [ ] **Step 4: Commit**

```bash
git add pkg/llm/llm.go
git commit -m "feat(llm): add ResponseParser() method to Service"
```

---

## Task 3: Modify agent to use interface

**Files:**
- Modify: `pkg/agent/agent.go`

- [ ] **Step 1: Find and remove textPart and parseThinkTags**

Remove from agent.go:
- `textPart` struct definition
- `parseThinkTags` function

- [ ] **Step 2: Update call site**

```go
// OLD:
parts := parseThinkTags(textResp)

// NEW:
parts := a.llmSvc.ResponseParser().Parse(textResp)
```

Note: Need to update the return type handling since TextPart uses `IsThink` instead of `isThink`.

- [ ] **Step 3: Run tests**

```bash
go build ./...
go test ./...
```

- [ ] **Step 4: Commit**

```bash
git add pkg/agent/agent.go
git commit -m "refactor(agent): use llm.ResponseParser interface instead of local parseThinkTags"
```

---

## Task 4: Final verification

- [ ] **Step 1: Run all tests**

```bash
go test ./...
```

- [ ] **Step 2: Run build**

```bash
go build ./...
```

- [ ] **Step 3: Commit**

```bash
git commit -m "chore: final verification for llm-response-parser"
```