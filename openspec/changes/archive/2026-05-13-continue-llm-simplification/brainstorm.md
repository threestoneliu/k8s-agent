## Design Summary

继续完成 simplify-llm-agent-session 中未完成的重构任务，包括：
- 简化 LLM 模块（移除残留的 provider 引用）
- 重命名 pkg/engine 为 pkg/k8s
- 删除 scheduler 模块
- 拆分 agent 模块为多个文件
- 简化 session 压缩内部实现

## Alternatives Considered

### 方案 A：按原始 plan 顺序执行
- **做法**：严格按照 simplify-llm-agent-session/plan.md 的任务顺序执行
- **優點**：与上一个 change 保持连续性
- **缺點**：任务间可能有依赖导致效率低下

### 方案 B：按依赖关系重排序
- **做法**：删除 scheduler → 更新 imports → 拆分 agent → 简化 session
- **優點**：减少任务间依赖
- **缺點**：与原始 plan 偏离

### 方案 C：渐进式重构
- **做法**：每次只重构一个小模块，确保测试通过后再继续
- **優點**：降低风险
- **缺點**：耗时长

## Agreed Approach

**方案 A：按原始 plan 顺序执行**

选择理由：
- 与上一个 change (simplify-llm-agent-session) 保持连续性
- 用户已明确选择方案 A
- plan.md 已经过充分设计，不需要大改

## Key Decisions

1. **pkg/engine → pkg/k8s 重命名** — 更新所有 imports
2. **删除 pkg/scheduler/** — 移除 scheduled task 功能
3. **Agent 拆分为 3 文件** — agent.go, executor.go, session.go
4. **Session 压缩简化** — 保持 L1/L2/L3 行为，内部实现简化
5. **依赖注入** — 移除全局状态，使用构造函数注入

## Open Questions

无重大开放问题。所有设计决策已在 plan.md 中明确。
