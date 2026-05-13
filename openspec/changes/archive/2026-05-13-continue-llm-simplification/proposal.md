## Why

LLM/Agent/Session 模块在上一轮重构中创建了 shared module，但大部分简化任务未完成。当前代码存在问题：pkg/engine 命名不符合职责、scheduler 模块冗余、agent 模块是 god object、session 压缩内部实现复杂。继续完成重构可提高代码可维护性和可测试性。

## What Changes

**pkg/engine 重命名为 pkg/k8s**
- From: 模块名 `pkg/engine` 包含 parser、classifier、mapper 等不相关的职责
- To: 模块名 `pkg/k8s` 专注 k8s 执行逻辑
- Reason: 清晰表达模块职责，降低认知负担

**删除 scheduler 模块**
- From: pkg/scheduler/ 目录存在，包含 cron、manager、task 等
- To: pkg/scheduler/ 完全移除，CLI task 命令已在上轮删除
- Reason: 功能已被移除，维护负担为零

**Agent 模块拆分**
- From: pkg/agent 是单个大文件（约 600+ 行）
- To: 拆分为 agent.go（主循环）、executor.go（工具执行）、session.go（会话管理）
- Reason: 单一职责原则，提高可测试性

**Session 压缩简化**
- From: findCompleteInteractions 和 findLLMCompleteInteractions 两个函数，L1/L2/L3 逻辑分散
- To: 合并为单一函数，压缩内部实现简化
- Reason: 减少代码重复，提高可维护性

## Capabilities

### New Capabilities

- `llm-cleanup`: 清理 LLM 模块中残留的 provider 引用和未使用的 import
- `k8s-rename`: 将 pkg/engine 重命名为 pkg/k8s，更新所有 imports
- `agent-split`: 将 agent 模块拆分为 agent.go、executor.go、session.go，使用依赖注入
- `session-compression-simplify`: 简化 session 3-level 压缩内部实现，合并重复函数
- `import-update`: 更新所有 imports，确保无循环依赖
- `test-coverage`: 为新拆分的模块添加单元测试

### Modified Capabilities

- 无（当前修改都是实现层面的，不改变外部行为）

## Impact

**受影响代码：**
- pkg/llm/ - 清理残留引用
- pkg/engine/ → pkg/k8s/ - 目录重命名
- pkg/agent/ - 文件拆分
- pkg/session/ - 压缩简化
- 所有导入这些模块的代码需要更新

**受影响测试：**
- pkg/llm/*_test.go
- pkg/agent/*_test.go
- pkg/session/*_test.go
- cmd/cli/root_test.go

**不受影响：**
- CLI 接口（chat、cluster 命令不变）
- IPC 协议（Input/Output 类型不变）
- 外部 API（无变化）
