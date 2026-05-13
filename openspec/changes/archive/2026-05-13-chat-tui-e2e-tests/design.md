## Context

k8s-agent 的 TUI 基于 Bubble Tea 框架实现。Bubble Tea 提供了一套测试模式，允许模拟用户输入（按键）和验证 UI 输出。

当前状态：
- `pkg/ui/tui.go` - TUI 实现，无测试
- `pkg/agent/agent.go` - Agent 逻辑，有部分单元测试
- `cmd/cli/chat.go` - CLI 命令入口，有基础测试

## Goals / Non-Goals

**Goals:**
- 为 TUI 交互流程编写测试
- 验证用户输入 → UI 的完整流程（不调用 LLM）
- 验证 UI 输出渲染（根据 MessageType 添加 emoji）
- 提供可选的 LLM 集成测试（默认跳过）

**Non-Goals:**
- 不在常规测试中调用真实 LLM
- 不测试 Bubble Tea 内部渲染细节
- 不测试真实的 Kubernetes 集群交互

## Decisions

### Decision 1: 测试分为两部分

**选择**: 单元测试（无 LLM）+ 集成测试（调用 LLM，默认跳过）

**理由**:
- 单元测试快速可靠，CI/CD 默认执行
- 集成测试需要真实 LLM API，成本高且不稳定
- 使用 Go 的 `//go:build integration` 标签区分

### Decision 2: 使用 Bubble Tea 测试模式

**选择**: 使用 `tea.Model` 的测试方法

**理由**:
- Bubble Tea 提供 `tea.Batch()` 和 `tea.Exec()` 测试助手
- 可以模拟按键输入并获取完整视图状态
- 与 Bubble Tea 官方推荐方式一致

### Decision 3: 测试文件结构

**选择**: 单元测试在 `pkg/ui/tui_test.go`，集成测试在 `pkg/ui/tui_integration_test.go`

**理由**:
- 单元测试访问私有字段和方法，在同一包内
- 集成测试使用 `//go:build integration` 标签，默认不执行
- 符合 Go 测试惯例

### Decision 4: 集成测试使用真实 LLM

**选择**: 集成测试调用真实 LLM API（单元测试不调用 LLM）

**理由**:
- 集成测试验证完整流程：用户输入 → Agent → LLM → UI
- 真实 LLM 才能验证 Function Calling 和工具调用的完整闭环
- 通过环境变量 `OPENAI_API_KEY` 或 `ANTHROPIC_API_KEY` 配置
- 集成测试默认跳过，只在手动执行时调用真实 LLM

## Risks / Trade-offs

**[Risk]** TUI 代码直接创建 goroutine，难以模拟并发
→ **Mitigation**: 通过 channel 控制输入输出，测试时使用 buffer channel

**[Risk]** Bubble Tea 测试需要初始化 tea.Model
→ **Mitigation**: 提取 model 创建逻辑到可注入的工厂函数

**[Risk]** UI 状态分散在多个字段
→ **Mitigation**: 通过 `tea.Model.Update()` 返回的 model 验证状态

## Open Questions

1. TUI.Run() 需要 TTY，是否需要提取可测试的接口？
2. 消息渲染结果的验证方式：检查 model.messages 还是检查 viewport content？
