## Context

当前实现中，Agent 维护两套独立的消息：

1. **`a.messages` (`[]*session.Message`)** — UI 展示用，带 emoji 前缀（💭 🔧 ✅ ❌），持久化到 FileStore
2. **`a.llmMessages` (`[]llm.Message`)** — LLM API 调用用，干净的 OpenAI ToolCalls 结构，不持久化

```
用户输入 "delete pod nginx"
    │
    ├─► a.messages = "You: delete pod nginx" + "💭 分析中..." + "🔧 执行工具..."
    └─► a.llmMessages = {Role: "user", Content: "delete pod nginx"} + ...
```

Session 恢复后，LLM 无法从 session 中重建对话上下文，只能看到格式化后的文本。

## Goals / Non-Goals

**Goals:**
- Session 持久化 LLM 对话原始格式，跨次对话可恢复完整 LLM 上下文
- 消除双份消息维护的复杂性
- UI 展示层从 LLM 消息反向渲染

**Non-Goals:**
- 不改变 LLM API 调用格式（仍使用 OpenAI ToolCalls 结构）
- 不修改 Agent 的流式处理逻辑
- 不改变 FileStore 的 LRU 缓存机制

## Decisions

### Decision 1: Message 结构统一

**选择**: 让 `session.Message` 兼容 `llm.Message` 的 ToolCalls 结构

**理由**:
- `session.Message` 已有 `ToolCalls []ToolCall` 字段
- `llm.Message` 的 ToolCalls 结构是 `[]struct{ID, Name, Arguments}` 
- 可以扩展 `session.Message` 的 ToolCall 类型以匹配

### Decision 2: UI 渲染反向转换

**选择**: 展示时从 `session.Message` 反向渲染 UI 格式（添加 emoji 前缀）

**理由**:
- 存储时保持干净的结构化数据
- UI 渲染逻辑集中在一处（可在 Agent 或 UI 层）
- emoji 等展示细节是 UI concern，不应污染数据层

### Decision 3: Session 恢复时重建 LLM 上下文

**选择**: 从 FileStore 加载 session 时，直接构建 `llmMessages` 供 LLM 使用

**理由**:
- `session.Message` 结构与 `llm.Message` 兼容，可直接映射
- 避免在运行时做复杂的格式转换

## Risks / Trade-offs

**[Risk]** UI 渲染逻辑需要从消息内容推断类型（是思考？工具调用？结果？）
→ **Mitigation**: 在 `session.Message` 中增加 `MessageType` 字段（text/think/tool_call/tool_result），UI 渲染据此添加 emoji

**[Risk]** 历史 session 中的 emoji 内容会被视为普通文本
→ **Mitigation**: 迁移期间不做处理，现有 session 继续以旧格式工作；新 session 使用新格式

**[Risk]** ToolCalls 结构需要保持与 `llm.Message` 兼容
→ **Mitigation**: 在设计阶段明确定义 ToolCall 字段映射，确保 API 兼容

## Migration Plan

1. 扩展 `session.Message` 增加 `MessageType` 字段
2. 修改 `agent.go` 中消息同步逻辑，从 `llmMessages` 同步到 `messages` 时附带类型信息
3. UI 渲染层根据 `MessageType` 添加 emoji 前缀
4. Session 恢复时从 `session.Message` 直接构建 `llmMessages`

## Open Questions

1. Session 恢复时是否需要压缩（token 限制）？如果需要，压缩逻辑放在哪层？
2. 旧的 UI 格式 session 如何兼容？是否需要一次性的格式迁移？