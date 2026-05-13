# Simplified LLM Module

## ADDED Requirements

### Requirement: Direct OpenAI Calls

The `llm.Service` struct SHALL directly call OpenAI SDK without Provider interface abstraction.

#### Scenario: Service.Chat calls OpenAI SDK directly

Given llm.Service with valid config
When Chat(ctx, messages) is called
Then OpenAI SDK client.Chat() is invoked
And response is returned without Provider interface

---

### Requirement: Functions Registry

Functions SHALL be registered via `RegisterFunction(fn FunctionDefinition)` in `functions.go`, replacing auto_register.go.

#### Scenario: Functions are registered at init

Given RegisterFunction is called with FunctionDefinition
When GetHandler(name) is called
Then correct handler function is returned

---

### Requirement: Supported Functions

The LLM module SHALL support these k8s functions:
- `resource_list`: List k8s resources with label/field selectors
- `resource_get`: Get specific k8s resource details
- `get_apiresources`: Discover available API resources
- `use_cluster`: Switch current cluster context

#### Scenario: All four functions are available

Given the function registry
When calling GetHandler for each: resource_list, resource_get, get_apiresources, use_cluster
Then all four return valid handlers

---

### Requirement: Removed Provider Interface

The `Provider` interface in `provider.go` SHALL be removed. Service calls OpenAI directly.

#### Scenario: No Provider interface exists

Given codebase
When searching for "type Provider interface"
Then no result is found
And llm.Service does not reference Provider

---

### Requirement: Config Integration

LLM module SHALL use `cluster.AppConfig.LLM` for configuration, not separate config.go.

#### Scenario: LLMConfig comes from AppConfig

Given cluster.AppConfig with LLM.APIKey, LLM.Model set
When NewService(cfg *LLMConfig) is called with values from AppConfig
Then service is created with correct configuration

---

### Requirement: Role Tool Messages

When sending tool results back to LLM, message role SHALL be "tool" with tool_call_id referencing original call.

#### Scenario: Tool result message has role="tool"

Given a tool call with id="call_123"
When creating message for tool result
Then Role="tool" and ToolCallID="call_123"
And Content contains the result

---

## REMOVED Requirements

### Requirement: ANTHROPIC Support

ANTHROPIC_API_KEY, ANTHROPIC_MODEL environment variables SHALL no longer be supported.

### Requirement: Provider Interface

The Provider interface with Chat/ChatWithFunctions/Name methods SHALL be removed.

### Requirement: Scheduled Task Functions

create_scheduled_task, list_scheduled_tasks, delete_scheduled_task functions SHALL be removed.

---

## Implementation Notes

- `pkg/llm/llm.go` — Service with direct OpenAI calls
- `pkg/llm/openai_sdk.go` — SDK wrapper (kept)
- `pkg/llm/functions.go` — Simplified function registry
- Files removed: provider.go, service.go, config.go, auto_register.go