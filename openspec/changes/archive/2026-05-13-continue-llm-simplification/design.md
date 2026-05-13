## Context

继续完成 simplify-llm-agent-session 中未完成的重构任务。当前代码状态：
- build 通过 ✅
- 测试通过 ✅
- shared module 已创建 ✅
- session.Message 已嵌入 shared.Message ✅

剩余任务：
- 简化 LLM 模块 (2.1-2.7)
- 重命名 pkg/engine 为 pkg/k8s (3.1-3.6)
- 删除 scheduler 模块 (4.1-4.2)
- 拆分 agent 模块 (5.1-5.5)
- 简化 session 压缩 (6.1-6.5)
- 更新 imports (7.1-7.4)
- 测试覆盖 (8.1-8.5)

## Goals / Non-Goals

**Goals:**
- 完成 LLM 模块简化（移除残留 provider 引用）
- 重命名 pkg/engine → pkg/k8s
- 删除 pkg/scheduler/ 模块
- 拆分 agent 模块为 agent.go、executor.go、session.go
- 使用依赖注入替代全局状态
- 简化 session 3-level 压缩内部实现
- 所有测试通过，功能验证通过

**Non-Goals:**
- 不引入新的外部依赖
- 不改变外部 API（CLI 接口不变）
- 不改变 3-level 压缩的用户可见行为
- 不重新设计 IPC 协议

## Decisions

### Decision 1: pkg/engine → pkg/k8s 重命名

**选择**：直接重命名目录并更新所有 imports

**理由**：
- 清晰表达模块职责
- 与 plan.md 一致
- Go module 路径变更需要协调更新

**替代方案**：
- 保持 pkg/engine 名称 → 不符合简化目标

### Decision 2: Agent 拆分为三个文件

**选择**：拆分为 agent.go（主循环）、executor.go（工具调用执行）、session.go（会话管理）

**理由**：
- 每个文件职责清晰
- 与 plan.md 一致
- 便于单独测试

**替代方案**：
- 拆分为更多文件 → 增加复杂度
- 不拆分 → god object 问题持续

### Decision 3: 依赖注入替代全局状态

**选择**：通过 NewAgent 构造函数注入依赖

**理由**：
- 消除全局状态
- 便于单元测试
- 显式依赖更清晰

### Decision 4: Scheduler 模块整体删除

**选择**：删除 pkg/scheduler/ 目录，移除 CLI task 子命令

**理由**：
- 已在 CLI 中移除 task 命令
- 避免未来维护负担
- 用户不需要 scheduled task 功能

### Decision 5: Session 压缩保持行为不变

**选择**：L1/L2/L3 压缩触发条件和行为保持不变

**理由**：
- 用户依赖现有压缩行为
- 内部实现可以简化
- 避免回归问题

## Risks / Trade-offs

| Risk | Mitigation |
|------|------------|
| pkg/k8s 重命名导致 imports 遗漏 | 逐个包检查 `go build ./...` |
| Agent 拆分破坏 IPC 通信 | 运行 TUI 集成测试验证 |
| Session 压缩行为变化 | 对比压缩前后消息数量 |
| 删除 scheduler 影响其他模块 | 检查所有 imports |

## Migration Plan

1. **Phase 1**: 删除 scheduler 模块 (4.1-4.2)
   - 删除 pkg/scheduler/
   - 运行测试确认无破坏

2. **Phase 2**: LLM 模块清理 (2.1-2.7)
   - 移除 provider.go 等引用（如果还有）
   - 更新 imports

3. **Phase 3**: pkg/k8s 重命名 (3.1-3.6)
   - `git mv pkg/engine pkg/k8s`
   - 更新所有 imports
   - 运行 `go build ./...`

4. **Phase 4**: Agent 拆分 (5.1-5.5)
   - 创建新文件，移动代码
   - 添加依赖注入
   - 更新 imports

5. **Phase 5**: Session 压缩简化 (6.1-6.5)
   - 简化内部方法
   - 运行压缩相关测试

6. **Phase 6**: 功能验证
   - 运行 `go test ./...`
   - 运行 k8s-agent chat 实际测试

**Rollback**: `git revert` 最后一个 commit 恢复所有变更

## Open Questions

无重大开放问题。设计已足够清晰可以执行。
