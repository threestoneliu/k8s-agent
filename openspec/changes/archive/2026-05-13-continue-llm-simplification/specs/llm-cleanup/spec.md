# LLM Cleanup

## ADDED Requirements

### Requirement: Remove Unused Provider References

The `pkg/llm` module SHALL NOT contain any references to `Provider` interface or deleted provider files.

```go
// Provider interface SHALL NOT exist in codebase
type Provider interface { ... }
```

#### Scenario: Provider interface is removed

Given the codebase
When searching for "type Provider interface"
Then no result is found

---

### Requirement: No Import of Deleted Files

The `pkg/llm` module SHALL NOT import any deleted files: `provider.go`, `service.go`, `config.go`.

#### Scenario: No import of deleted files

Given `pkg/llm` module
When checking imports
Then imports do not reference provider.go, service.go, or config.go

---

### Requirement: Service Uses OpenAISDKProvider Directly

The `llm.Service` SHALL use `OpenAISDKProvider` for LLM calls without indirection.

#### Scenario: Service calls OpenAI SDK directly

Given `llm.Service` with valid `LLMConfig`
When `Chat()` is called
Then `OpenAISDKProvider.Chat()` is invoked directly
