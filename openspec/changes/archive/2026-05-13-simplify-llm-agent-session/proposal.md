## Why

Current module structure has over-engineering issues: Provider interface with single implementation, 422-line auto_register.go, 764-line god object in agent, duplicate interaction structs, complex resource normalization. This makes the code hard to maintain and understand. Simplification preserves all functional capabilities while reducing complexity.

## What Changes

**Module Structure Simplification**
- From: `pkg/llm/` with Provider interface, `pkg/engine/` complex abstraction, `pkg/scheduler/` unused
- To: `pkg/shared/` for shared types, `pkg/k8s/` simplified, scheduler deleted

**LLM Module**
- From: `Provider` interface with 1 implementation, separate `config.go`, 422-line `auto_register.go`
- To: Direct OpenAI calls via `Service`, functions defined in `functions.go`

**Agent Module**
- From: 764-line `agent/agent.go` handling IPC, LLM calls, tool execution, session persistence
- To: Split into `agent.go` (main loop), `executor.go` (tool execution), `session.go` (session management)

**Session Module**
- From: Duplicate `interaction` struct in context.go and interaction.go, complex L1/L2/L3 compression
- To: Unified `Interaction` struct, 3-level compression preserved but internally simplified

**Message Format**
- From: Custom Role constants, separate MessageType for tool_call/tool_result
- To: OpenAI standard roles (user/assistant/system/tool), MessageType only for UI (text/think)

**K8s Module**
- From: `pkg/engine/` with mapper.go, parser.go, classifier.go, executor_query.go
- To: `pkg/k8s/` with simplified executor.go, registry.go, types.go

## Capabilities

### New Capabilities
- `shared-types`: Shared Message and Function types to prevent circular dependencies between modules
- `simplified-llm`: Simplified LLM module without Provider interface, direct OpenAI calls
- `simplified-agent`: Agent split into smaller, focused files with clear responsibilities
- `simplified-k8s`: K8s operations module simplified from complex engine implementation

### Modified Capabilities
- `session-management`: Internal structure simplified, 3-level compression preserved, interaction struct unified

## Impact

**Files Deleted**: `pkg/scheduler/` (entire module), `pkg/llm/provider.go`, `pkg/llm/service.go`, `pkg/llm/auto_register.go` (merged into functions.go), `pkg/llm/config.go` (use pkg/config)

**Files Created**: `pkg/shared/message.go`, `pkg/shared/function.go`, `pkg/k8s/` (from engine simplification)

**Files Modified**: `pkg/llm/llm.go` (new), `pkg/agent/` split from agent.go, `pkg/session/context.go` (simplified), `pkg/session/interaction.go` (unified)

**Functions Removed**: create_scheduled_task, list_scheduled_tasks, delete_scheduled_task

**Functions Preserved**: resource_list, resource_get, get_apiresources, use_cluster