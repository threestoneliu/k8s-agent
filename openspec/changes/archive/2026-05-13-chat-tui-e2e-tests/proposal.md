## Why

k8s-agent 的 TUI 实现目前没有自动化测试覆盖。chat 命令是用户主要交互方式，但 UI 逻辑（消息显示、命令处理、状态更新）缺乏测试验证。现有测试只覆盖了 Agent 和底层组件，没有覆盖完整的 UI 交互流程。

## What Changes

新增 TUI 测试，分两部分：

**单元测试（默认执行，无 LLM）：**
- TUI 消息输入/输出处理
- 命令解析（`/clusters`, `/config`, `/cluster`）
- UI 渲染（根据 MessageType 添加 emoji）
- viewport 更新和边界处理

**集成测试（默认跳过，真实 LLM）：**
- 用户输入 → Agent → LLM API → UI 的完整流程
- 验证 Function Calling 和工具调用的完整闭环
- 通过环境变量配置 API key
- 使用 `//go:build integration` 标签
- 执行方式：`OPENAI_API_KEY=xxx go test ./pkg/ui/... -tags=integration -v`

## Capabilities

### New Capabilities
- `chat-tui-e2e`: TUI 测试，覆盖 UI 交互流程

## Impact

- 新增 `pkg/ui/tui_test.go` - 单元测试（无 LLM）
- 新增 `pkg/ui/tui_integration_test.go` - 集成测试（mock LLM）
- 使用 Bubble Tea 的测试模式验证 UI 行为
- 可能需要重构部分 TUI 代码以支持测试
