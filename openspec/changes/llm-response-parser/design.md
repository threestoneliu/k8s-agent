## Context

当前 `parseThinkTags` 函数定义在 `pkg/agent/agent.go` 中，解析 OpenAI 的 `<think>` `</think>` XML 标签。这个函数与 LLM 输出格式紧密相关，但放在 agent 包中不合理：

1. **耦合问题**：agent 包依赖具体的 LLM 输出格式
2. **扩展性差**：如果支持其他 LLM provider（Anthropic、Ollama），它们的 think 格式不同
3. **测试困难**：难以 mock 不同的解析逻辑

## Goals / Non-Goals

**Goals:**
- 将 ResponseParser 接口定义在 llm 包
- 让不同 LLM provider 可以提供自己的解析实现
- Agent 通过接口调用，保持对展示逻辑的关注

**Non-Goals:**
- 不实现其他 provider 的解析器（暂只支持 OpenAI）
- 不改变 agent 的展示逻辑
- 不修改现有的压缩/摘要逻辑

## Decisions

### 1. 接口定义在 llm 包

```go
// pkg/llm/parser.go
type TextPart struct {
    IsThink  bool
    Content  string
}

type ResponseParser interface {
    Parse(text string) []TextPart
}
```

**为何**：解析逻辑与 LLM 输出格式相关，放在 llm 包更合理。

### 2. Service 提供默认 parser

```go
// pkg/llm/llm.go
func (s *Service) ResponseParser() ResponseParser {
    return &OpenAIResponseParser{}
}
```

**为何**：Agent 通过 service 获取 parser，保持简单的调用链。

### 3. Agent 直接使用接口

```go
// pkg/agent/agent.go
parts := a.llmSvc.ResponseParser().Parse(textResp)
```

**为何**：接口调用，便于测试时注入 mock parser。

### 4. 保留 OpenAIResponseParser 实现

```go
// pkg/llm/parser.go
type OpenAIResponseParser struct{}

func (p *OpenAIResponseParser) Parse(text string) []TextPart {
    // 移动现有的 parseThinkTags 逻辑
}
```

**替代方案**：直接在 Service 中返回 `func(text string) []TextPart`

**为何选择当前方案**：
- 接口更扩展，未来可以添加更多方法
- 便于实现多个 provider 的 parser

## Risks / Trade-offs

| 风险 | 影响 | 缓解 |
|------|------|------|
| 引入接口增加复杂度 | 小 | 只增加一个接口和一个实现 |
| 需要修改 agent 调用 | 小 | 只需改一处调用 |

## Migration Plan

1. 创建 `pkg/llm/parser.go` - 定义接口和 OpenAI 实现
2. 修改 `pkg/llm/llm.go` - 添加 ResponseParser() 方法
3. 修改 `pkg/agent/agent.go` - 删除本地实现，使用接口
4. 运行测试确保一切正常

**回滚策略**：如果出问题，恢复 agent.go 中的 `textPart` 和 `parseThinkTags` 定义。

## Open Questions

无。