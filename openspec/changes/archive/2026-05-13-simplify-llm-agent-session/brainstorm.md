# Design Summary

Simplify k8s-agent modules (llm/agent/session/k8s) by removing over-engineering while preserving all functional capabilities.

## Key Problems Identified

1. **LLM module**: Provider interface with only 1 implementation, auto_register.go at 422 lines, duplicate config
2. **Agent module**: 764-line god object handling multiple responsibilities
3. **Session module**: Duplicate `interaction` structs, complex 3-level compression
4. **Engine module**: Overly complex resource normalization, multiple abstraction layers

## Simplification Targets

| Module | Problem | Solution |
|--------|---------|----------|
| pkg/llm/ | Provider interface, auto_register.go | Remove interface, simplify function registry |
| pkg/agent/ | 764-line god object | Split into agent.go, executor.go, session.go |
| pkg/session/ | Duplicate interaction, complex L1/L2/L3 | Keep 3-level compression, unify struct |
| pkg/engine/ | Complex resource handling | Rename to k8s/, simplify internal logic |
| pkg/scheduler/ | Not needed | Delete entire module |

## Agreed Approach

**Core Principle**: Code refactoring, not functional change. All existing capabilities preserved.

**New Module Structure**:
```
pkg/
├── shared/       # Shared types (Message, Function) - NEW
├── config/       # Configuration (AppConfig + Registry)
├── cluster/      # Cluster management (preserve, simplify)
├── k8s/          # K8s operations (simplified from engine)
├── llm/          # LLM calls (simplified)
├── agent/        # Agent (split from god object)
└── session/      # Session with 3-level compression (preserve)
```

**Key Design Decisions**:
1. Extract shared types to prevent circular dependencies
2. Use OpenAI standard roles (user/assistant/system/tool) in Message
3. session.Message embeds shared.Message for consistency
4. 3-level compression preserved but internally simplified
5. K8s functions reduced to: resource_list, resource_get, get_apiresources, use_cluster
6. Remove scheduled task functions (create_scheduled_task, etc.)

## Alternatives Considered

### 方案 A：激进简化
- Remove all interfaces, direct calls everywhere
- Single file per module
- **未采用**: Too aggressive, loses structure clarity

### 方案 B：适度简化（当前方案）
- Extract shared module for type safety
- Split agent.go into logical units
- Simplify internal implementation while preserving interfaces
- **采用**: Balance between simplification and maintainability

### 方案 C：最小改动
- Only remove dead code
- Keep existing structure
- **未采用**: Doesn't address over-engineering issues

## Key Decisions

1. **shared 模块**：防止循环依赖，共享 Message 和 Function 类型
2. **session.Message**：使用 OpenAI 标准 Role，tool_call/tool_result 不需要 MessageType 区分
3. **cluster 保留**：多集群管理能力完整保留，只是代码位置调整
4. **3级压缩不变**：L1 interaction-based, L2 truncation, L3 summary 能力不变，内部实现简化
5. **scheduler 删除**：定时任务暂不需要，整个模块删除

## Open Questions

- [x] config 模块位置：pkg/config/ 或 pkg/cluster/ — 决定使用 pkg/config/
- [x] Message 类型统一：session.Message 和 llm.Message 分离通过 shared/ 共享
- [x] 3级压缩是否保留 — 确认保留，仅简化内部实现