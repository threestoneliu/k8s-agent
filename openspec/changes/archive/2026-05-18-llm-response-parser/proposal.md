## Why

`parseThinkTags` 当前定义在 agent 包中，直接依赖 OpenAI 的 `<think>` `</think>` XML 标签格式。这导致 agent 与特定 LLM provider 耦合，扩展到其他 provider（Anthropic、Ollama）时需要修改 agent 代码。将解析逻辑移到 llm 包并通过接口提供，可以解耦并提高可测试性。

## What Changes

**ResponseParser 接口**

- From: `parseThinkTags` 函数定义在 `pkg/agent/agent.go`，解析逻辑硬编码
- To: `ResponseParser` 接口定义在 `pkg/llm/parser.go`，OpenAI 实现可通过 service 获取
- Reason: 解耦 agent 和特定 LLM 输出格式
- Impact: 非破坏性，agent 调用方式微调

**Agent 调用变更**

- From: `parts := parseThinkTags(textResp)` (本地函数)
- To: `parts := a.llmSvc.ResponseParser().Parse(textResp)` (接口调用)
- Reason: 通过接口获取解析器，便于扩展和测试
- Impact: 单一调用点变更

## Capabilities

### New Capabilities

- `llm-response-parser`: LLM 响应解析抽象层，提供 ResponseParser 接口，支持不同 provider 的 think 标签解析

### Modified Capabilities

无

## Impact

**受影响代码：**
- `pkg/llm/parser.go`: 新增文件，定义接口和 OpenAI 实现
- `pkg/llm/llm.go`: 新增 ResponseParser() 方法
- `pkg/agent/agent.go`: 删除 textPart 和 parseThinkTags，修改调用处

**无影响：**
- 其他包无变更
- API 无破坏性变更