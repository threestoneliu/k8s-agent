## Design Summary

使用 LLM 对会话内容进行摘要，替代当前基于关键词统计的轻量级摘要生成。

### 核心流程

```
Agent.BuildContextMessages()
  ↓
Level 0: 消息数 ≤ max-messages 且 token 足够 → 直接返回
  ↓
Level 1: 交互压缩（保留最近 N 条工具调用）→ 检查是否足够
  ↓
Level 2: 截断到 max-messages → 检查是否足够
  ↓
Level 3: 返回 [NeedSummary] 标记 + 待摘要消息
  ↓
Agent 检测到 [NeedSummary] 标记
  ↓
Agent 调用 llm.Service.Chat() 生成对话式摘要
  ↓
Agent 重新调用 BuildContextMessages()，传入摘要结果
```

### 摘要风格

对话式摘要示例：
```
用户与管理集群 'dev' 进行了交互，主要查询了 pod 状态，
查看了 default 和 kube-system 命名空间下的资源。
执行了多次 list/get 操作，未进行危险操作。
```

### 设计决策

1. **推送模式（方案 B）**：Agent 外部调用 LLM 摘要，避免 session → llm 循环依赖
2. **实时生成**：Level 3 压缩时实时调用 LLM 生成摘要
3. **摘要 + 最近上下文**：摘要后保留最近 N 条消息作为"最近上下文"
4. **直接使用 Chat()**：使用 `llm.Service.Chat()` 而非 `ChatWithFunctions()`

### 包职责划分

| 包 | 职责 |
|---|---|
| `pkg/session` | ContextManager，消息压缩策略，返回摘要请求 |
| `pkg/llm` | LLM Service，提供 Chat() 接口 |
| `pkg/agent` | 编排层，调用 LLM 生成摘要，协调摘要流程 |

## Alternatives Considered

### 方案 A：ContextManager 内部调用（拉取模式）

- **做法**：在 `session` 包定义 `Summarizer` 接口，Agent 注入实现
- **优点**：封装性好，ContextManager 自主控制摘要时机
- **缺点**：会形成 session → llm 依赖，或需要引入中间接口层
- **为何未采用**：产生循环依赖或增加不必要的接口层

### 方案 C：缓存摘要

- **做法**：摘要结果缓存，仅当新消息积累到一定量时重新生成
- **优点**：减少 LLM 调用次数，降低延迟和成本
- **缺点**：实现复杂度增加，缓存失效策略复杂
- **为何未采用**：当前场景实时性更重要，缓存收益不明显

## Agreed Approach

采用 **方案 B（推送模式）**：
- ContextManager 在 Level 3 返回摘要请求，不自行调用 LLM
- Agent 检测到摘要请求后，调用 llm.Service.Chat() 生成对话式摘要
- 摘要后保留最近 N 条消息作为"最近上下文"

## Key Decisions

1. **ContextManager.BuildContextMessages() 签名调整**
   - 返回 `needsSummary bool` 和 `rawForSummary []*Message`
   - 当 needsSummary 为 true 时，rawForSummary 包含待摘要的消息

2. **Agent 新增 generateLLMSummary() 方法**
   - 使用 llm.Service.Chat() 生成摘要
   - 构建专门的摘要提示词

3. **摘要风格为对话式**
   - 自然语言描述，而非结构化列表
   - 包含操作类型、涉及资源、命名空间等关键信息

## Open Questions

无。所有关键决策已在讨论中确定。
