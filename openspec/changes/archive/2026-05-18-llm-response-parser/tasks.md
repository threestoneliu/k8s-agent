## 1. 创建 ResponseParser 接口和实现

- [x] 1.1 创建 `pkg/llm/parser.go`，定义 `TextPart` 结构体
- [x] 1.2 定义 `ResponseParser` 接口
- [x] 1.3 实现 `OpenAIResponseParser` 的 `Parse` 方法（移动现有的 parseThinkTags 逻辑）

## 2. 修改 llm.Service

- [x] 2.1 在 `Service` 结构体中添加 `responseParser` 字段
- [x] 2.2 实现 `ResponseParser() ResponseParser` 方法

## 3. 修改 agent 调用

- [x] 3.1 删除 `pkg/agent/agent.go` 中的 `textPart` 类型定义
- [x] 3.2 删除 `parseThinkTags` 函数
- [x] 3.3 修改调用处为 `a.llmSvc.ResponseParser().Parse(textResp)`

## 4. 验证

- [x] 4.1 运行 `go build ./...` 确保编译通过
- [x] 4.2 运行 `go test ./...` 确保所有测试通过