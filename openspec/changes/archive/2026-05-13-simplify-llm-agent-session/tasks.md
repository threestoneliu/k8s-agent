# Tasks: Simplify LLM/Agent/Session Modules

## 1. Create shared module

- [x] 1.1 Create `pkg/shared/message.go` with Message, ToolCall types and Role constants
- [x] 1.2 Create `pkg/shared/function.go` with Function, FunctionCall types
- [x] 1.3 Verify no circular dependencies (go mod verify)

## 2. Simplify LLM module

- [ ] 2.1 Remove `pkg/llm/provider.go` (Provider interface deleted)
- [ ] 2.2 Remove `pkg/llm/service.go` (merged into llm.go)
- [ ] 2.3 Remove `pkg/llm/config.go` (use pkg/config instead)
- [ ] 2.4 Create `pkg/llm/llm.go` with direct OpenAI calls (replace service.go)
- [ ] 2.5 Simplify `pkg/llm/functions.go` from auto_register.go (remove scheduled task handlers)
- [ ] 2.6 Remove ANTHROPIC environment variable support from LLM config
- [ ] 2.7 Update LLM calls to use shared.Message format

## 3. Simplify K8s module (rename from engine)

- [ ] 3.1 Rename `pkg/engine/` to `pkg/k8s/`
- [ ] 3.2 Merge `executor_query.go` into `executor.go`
- [ ] 3.3 Remove `mapper.go` (simplify resource normalization)
- [ ] 3.4 Remove `parser.go` (no longer needed)
- [ ] 3.5 Remove `classifier.go` (no longer needed)
- [ ] 3.6 Keep only: resource_list, resource_get, get_apiresources, use_cluster handlers

## 4. Delete scheduler module

- [ ] 4.1 Delete `pkg/scheduler/` entire directory
- [ ] 4.2 Remove scheduled task function handlers from llm/functions.go

## 5. Simplify agent module (split god object)

- [ ] 5.1 Create `pkg/agent/agent.go` (main loop, IPC handling, LLM interaction)
- [ ] 5.2 Create `pkg/agent/executor.go` (tool call execution, no global state)
- [ ] 5.3 Create `pkg/agent/session.go` (session management, message creation)
- [ ] 5.4 Remove global state (SetExecutor/SetSchedulerManager) from llm/executor.go
- [ ] 5.5 Use dependency injection for Executor in agent

## 6. Simplify session module

- [ ] 6.1 Unify Interaction struct (remove duplicate in context.go/interaction.go)
- [ ] 6.2 Update session.Message to embed shared.Message
- [ ] 6.3 Simplify 3-level compression internal implementation
- [ ] 6.4 Merge findCompleteInteractions and findLLMCompleteInteractions
- [ ] 6.5 Keep L1/L2/L3 compression triggers unchanged (behavior identical)

## 7. Update imports and dependencies

- [ ] 7.1 Update all imports to use pkg/shared/
- [ ] 7.2 Update llm module to use pkg/config for configuration
- [ ] 7.3 Fix any import cycles introduced by refactor
- [ ] 7.4 Verify go.mod dependencies are clean

## 8. Run tests and verify

- [ ] 8.1 Run existing tests to ensure behavior unchanged
- [ ] 8.2 Add tests for new shared module types
- [ ] 8.3 Verify all 3-level compression scenarios still work
- [ ] 8.4 Verify k8s operations (list/get/describe) work correctly
- [ ] 8.5 Build and ensure no compilation errors