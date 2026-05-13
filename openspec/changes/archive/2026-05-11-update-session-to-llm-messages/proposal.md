## Why

当前 session 持久化的是 UI 展示格式（带 emoji 前缀），而不是 LLM 对话的原始结构。这导致两个问题：

1. **跨次对话丢失 LLM 上下文** — session 恢复后，LLM 无法获得历史对话上下文，只能看到 UI 格式化后的文本
2. **数据冗余** — 同时维护两套消息格式（session.Message 带 emoji vs llm.Message 干净结构），维护成本高

如果用户关闭 TUI 再重新打开同一个 session，当前对话历史无法被 LLM 继续利用。

## What Changes

- Session 持久化改为存储 `llm.Message` 结构，而非 UI 格式化的 `session.Message`
- UI 展示时从 LLM 消息格式反向渲染（添加 emoji 前缀等）
- 消除双份消息维护的复杂性

## Capabilities

### New Capabilities
- `session-llm-format`: Session 持久化 LLM 消息格式，支持跨次对话的上下文恢复

### Modified Capabilities
- 无（当前 session 并未用于 LLM 上下文恢复）

## Impact

- `pkg/agent/agent.go` — 消息同步逻辑需要重构
- `pkg/session/message.go` — Message 结构可能需要调整
- `pkg/session/filestore.go` — 持久化格式变更
- UI 层需要从 LLM 格式反向渲染展示