## Context

当前 `ContextManager.generateSummary()` 使用轻量级关键词统计生成摘要：
- 统计消息中的 K8s 资源类型（pod、service、deployment）
- 提取最近用户查询和工具调用

这种方式无法理解对话语义，摘要质量差。

需要使用 LLM 生成真正理解对话内容的对话式摘要。

## Goals / Non-Goals

**Goals:**
- 使用 LLM 生成对话式摘要，替代轻量级关键词统计
- 保持代码结构合理，无循环依赖问题
- 在 Level 3 压缩时实时生成摘要

**Non-Goals:**
- 不改变现有的 Level 0/1/2 压缩策略
- 不修改 `pkg/llm` 的接口
- 不实现摘要缓存机制

## Decisions

### 1. 推送模式（方案 B）而非拉取模式

**选择**：Agent 外部调用 LLM 摘要

**理由**：
- `pkg/session` 不依赖 `pkg/llm`，避免循环依赖
- ContextManager 保持纯消息压缩逻辑，职责单一
- Agent 作为编排层，协调 LLM 调用和消息压缩

**替代方案 A（拉取模式）**：session 包定义 Summarizer 接口，Agent 注入
- 问题：增加接口层复杂度，或导致循环依赖

### 2. ContextManager.BuildContextMessages() 签名调整

```go
// 旧签名
func (cm *ContextManager) BuildContextMessages(
    systemPrompt string,
    messages []*Message,
    summaryPrompt string,
) []sharedutil.Message

// 新签名
func (cm *ContextManager) BuildContextMessages(
    systemPrompt string,
    messages []*Message,
    summaryPrompt string,
) (llmMessages []sharedutil.Message, needsSummary bool, rawForSummary []*Message)
```

**返回值变更**：
- `needsSummary bool`：是否需要 LLM 摘要
- `rawForSummary []*Message`：待摘要的原始消息（用于 LLM 调用）

**当 needsSummary 为 true 时**：
- llmMessages 仅包含系统提示和摘要请求标记
- Agent 应调用 LLM 生成摘要，然后重新调用 BuildContextMessages()

### 3. Agent.generateLLMSummary() 方法

```go
func (a *Agent) generateLLMSummary(messages []*Message) (string, error) {
    prompt := a.buildSummaryPrompt(messages)
    resp, err := a.llmSvc.Chat(context.Background(), []sharedutil.Message{
        {Role: "user", Content: prompt},
    })
    // ...
}
```

**摘要提示词构建**：
```
请用简洁的自然语言总结以下会话，风格如同助手在回顾对话：

用户与管理集群 'dev' 进行了交互，主要查询了 pod 状态，
查看了 default 和 kube-system 命名空间下的资源。
执行了多个 list/get 操作，未进行危险操作。

[会话内容...]
```

### 4. 摘要后保留最近上下文

Level 3 压缩后的消息结构：
```
1. system prompt
2. [Previous conversation summary]: <LLM生成的对话式摘要>
3. [N条最近消息被省略]
4. 最近 3-5 条消息作为"最近上下文"
```

配置项 `tool-call-retention` 控制保留的最近消息数量。

### 5. 摘要风格：对话式

示例：
```
用户与管理集群 'dev' 进行了交互，主要查询了 pod 状态，
查看了 default 和 kube-system 命名空间下的资源。
执行了多次 list/get 操作，未进行危险操作。
```

特点：
- 自然语言段落，而非结构化列表
- 包含集群名称、操作类型、涉及资源
- 突出危险操作（delete、create 等）

## Risks / Trade-offs

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| LLM 摘要增加延迟 | Level 3 压缩时用户体验下降 | 实时生成是产品需求，可接受延迟 |
| LLM 调用失败 | 摘要无法生成 | fallback 到轻量级摘要或返回错误 |
| token 消耗增加 | 每次 Level 3 压缩消耗额外 token | 摘要长度控制在合理范围（~200 tokens） |

## Migration Plan

1. **Phase 1**: 修改 `ContextManager.BuildContextMessages()` 返回值，新增 `needsSummary` 和 `rawForSummary`
2. **Phase 2**: 在 `Agent` 中新增 `generateLLMSummary()` 方法
3. **Phase 3**: Agent 调用 `BuildContextMessages()` 后检查 `needsSummary`，如需要则调用 LLM 摘要
4. **Phase 4**: 更新测试，确保新逻辑覆盖

**回滚策略**：如有问题，可将 `needsSummary` 默认为 false 快速回滚。

## Open Questions

无。所有关键决策已在 brainstorming 中确定。
