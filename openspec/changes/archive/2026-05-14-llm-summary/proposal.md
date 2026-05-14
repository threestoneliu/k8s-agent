## Why

当前 `ContextManager.generateSummary()` 使用轻量级关键词统计生成摘要，仅能提取消息中出现的 K8s 资源类型和工具调用，无法理解对话语义。这导致摘要质量差，在长对话恢复时上下文理解不准确。使用 LLM 生成对话式摘要可以真正理解会话内容，生成自然语言形式的上下文回顾。

## What Changes

**会话摘要生成逻辑**

- From: ContextManager 内部使用关键词统计生成轻量级摘要
- To: Agent 在 Level 3 压缩时调用 LLM Service.Chat() 生成对话式摘要
- Reason: 提升摘要质量，更好地恢复长对话上下文
- Impact: 非破坏性变更，向后兼容

**ContextManager.BuildContextMessages() 签名调整**

- From: 返回 `[]sharedutil.Message`
- To: 返回 `([]sharedutil.Message, bool, []*Message)` — 新增 needsSummary 和 rawForSummary
- Reason: 支持 Agent 判断是否需要调用 LLM 摘要
- Impact: 调用方需适配新签名

## Capabilities

### New Capabilities

- `llm-summary`: LLM驱动的会话摘要生成，在 Level 3 压缩时触发，生成自然语言形式的对话式摘要

## Impact

**受影响的代码：**
- `pkg/session/context.go`: BuildContextMessages() 签名变更
- `pkg/agent/agent.go`: 新增 generateLLMSummary() 方法

**无影响：**
- `pkg/llm`: 无接口变更，仅被 Agent 调用
- `pkg/session`: 其他压缩逻辑保持不变
