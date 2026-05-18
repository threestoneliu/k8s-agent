## Design Summary

将 `parseThinkTags` 从 agent 包移到 llm 包，定义 `ResponseParser` 接口，让不同 LLM provider 实现自己的 think 标签解析逻辑。

### 核心设计

```go
// llm 包
type TextPart struct {
    IsThink  bool
    Content  string
}

type ResponseParser interface {
    Parse(text string) []TextPart
}

// agent 包调用
parts := a.llmSvc.ResponseParser().Parse(textResp)
```

### 文件变更

| 文件 | 变更 |
|---|---|
| `pkg/llm/parser.go` | 新增：ResponseParser 接口和 OpenAI 实现 |
| `pkg/llm/llm.go` | 新增 ResponseParser() 方法 |
| `pkg/agent/agent.go` | 删除 textPart 和 parseThinkTags，修改调用处 |

### 依赖关系

```
agent → llm (调用 ResponseParser 接口)
```

无循环依赖。

## Agreed Approach

采用上述设计方案：
- 接口定义在 llm 包
- OpenAIResponseParser 作为默认实现
- Agent 直接调用 `llmSvc.ResponseParser().Parse()`

## Key Decisions

1. **TextPart 类型放在 llm 包** - 作为 ResponseParser 的返回值
2. **Service 添加 ResponseParser() 方法** - 返回默认的 OpenAI parser
3. **Agent 直接使用接口** - 便于后续扩展其他 provider

## Open Questions

无。