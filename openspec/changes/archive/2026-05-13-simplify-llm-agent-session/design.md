## Context

**Current State**: k8s-agent has over-engineered module structure:
- `pkg/llm/`: Provider interface with 1 implementation, 422-line auto_register.go, duplicate config
- `pkg/agent/`: 764-line god object handling IPC, LLM calls, tool execution, session management
- `pkg/session/`: Duplicate `interaction` structs, complex 3-level compression implementation
- `pkg/engine/`: Complex resource normalization, multiple abstraction layers (mapper, parser, classifier)
- `pkg/scheduler/`: Unused module consuming ~800 lines

**Why Now**: The code is hard to maintain and understand. New features require navigating complex abstractions. Bugs are hard to trace through multiple layers.

**Stakeholders**: k8s-agent maintainers, developers adding new k8s operations

## Goals / Non-Goals

**Goals:**
- Simplify module structure while preserving all functional capabilities
- Remove dead abstractions (Provider interface with 1 impl)
- Split god objects into focused, single-responsibility units
- Unify duplicate type definitions
- Maintain backward compatibility for existing usage patterns

**Non-Goals:**
- No functional changes to k8s operations (list/get/describe behavior unchanged)
- No changes to 3-level compression capabilities
- No changes to cluster management capabilities
- Not adding new k8s operations or LLM providers
- Not implementing the removed scheduler module

## Decisions

### Decision 1: Extract shared module to prevent circular dependencies

**Choice**: Create `pkg/shared/` with Message and Function types
**Rationale**: Multiple modules (llm, agent, session) need to share these types. Without a shared module, we get either duplication or circular imports.
**Alternatives Considered**:
- Put shared types in session/ — would cause session → llm dependency
- Put in llm/ — would cause agent → llm dependency  
- Duplicate everywhere — violates DRY

### Decision 2: Use OpenAI standard roles in Message

**Choice**: `RoleUser = "user"`, `RoleAssistant = "assistant"`, `RoleSystem = "system"`, `RoleTool = "tool"`
**Rationale**: Aligns with OpenAI API format, simplifies message conversion. tool_call/tool_result distinguished by Role, not MessageType.
**Alternatives Considered**:
- Keep custom role constants with different values — requires conversion layer
- Use enum type instead of string — more type-safe but less compatible with OpenAI SDK

### Decision 3: Keep 3-level compression, simplify internal implementation

**Choice**: Preserve L1 (interaction-based), L2 (message truncation), L3 (summary) but simplify code structure
**Rationale**: Compression is essential for preventing context overflow. The capability is correct, only implementation is complex.
**Alternatives Considered**:
- Remove L2/L3 — would break existing behavior when L1 is insufficient
- Keep as-is — over-engineering issues remain

### Decision 4: Delete scheduler module entirely

**Choice**: Remove `pkg/scheduler/` completely
**Rationale**: Scheduled tasks not currently used. Saves ~800 lines of code.
**Alternatives Considered**:
- Keep but mark as deprecated — still adds to code maintenance burden
- Move to separate repo — overkill for unused code

### Decision 5: Rename engine to k8s, simplify internal logic

**Choice**: `pkg/engine/` → `pkg/k8s/`, keep only executor.go with core operations
**Rationale**: "engine" is too generic. Resource normalization in mapper.go is overly complex for our needs.
**Alternatives Considered**:
- Keep engine/ name — doesn't reflect actual purpose
- Keep mapper/parser/classifier — adds complexity without proportional benefit

## Risks / Trade-offs

[Risk] Tool call format change → Mitigation: Extensive test coverage, backward-compatible Message structure

[Risk] Regression in context compression → Mitigation: Preserve 3-level capability, run existing compression tests

[Risk] Circular dependency after refactor → Mitigation: Dependency audit via `go mod verify`, shared module pattern

[Trade-off] Shared module adds indirection → Benefit: Prevents circular deps, worth the indirection

[Trade-off] Splitting agent.go increases file count → Benefit: Each file has single responsibility, easier to navigate

## Migration Plan

**Phase 1: Create shared module**
- Create `pkg/shared/message.go` and `pkg/shared/function.go`
- Update existing types to use shared types

**Phase 2: Simplify llm module**
- Remove Provider interface and service.go
- Inline function registration from auto_register.go to functions.go
- Remove ANTHROPIC env var support

**Phase 3: Simplify k8s module**
- Rename pkg/engine to pkg/k8s
- Merge executor_query.go into executor.go
- Remove mapper.go, parser.go, classifier.go

**Phase 4: Simplify agent module**
- Split agent.go into agent.go, executor.go, session.go
- Remove global state in llm/executor.go

**Phase 5: Simplify session module**
- Unify interaction structs
- Simplify compression logic internally
- Ensure tests pass

**Rollback**: Git history allows full rollback. Each phase is independently testable.

## Open Questions

- [x] Config module location: pkg/config/ (confirmed)
- [x] Shared types content: Message and Function only (confirmed)
- [x] 3-level compression: Preserve capability, simplify implementation (confirmed)
- [x] Scheduler: Delete entirely (confirmed)
- [x] K8s functions: Keep resource_list, resource_get, get_apiresources, use_cluster (confirmed)