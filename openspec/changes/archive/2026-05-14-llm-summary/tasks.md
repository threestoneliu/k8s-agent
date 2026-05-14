## 1. ContextManager.BuildContextMessages() 签名修改

- [ ] 1.1 修改 `pkg/session/context.go` 中 `BuildContextMessages()` 返回值，新增 `needsSummary bool` 和 `rawForSummary []*Message`
- [ ] 1.2 更新 Level 3 逻辑：当需要深度压缩时，返回 `needsSummary=true` 和待摘要消息，而非调用内部 generateSummary
- [ ] 1.3 保留旧的轻量级 generateSummary() 方法作为 fallback（LLM 调用失败时使用）

## 2. Agent.generateLLMSummary() 方法

- [ ] 2.1 在 `pkg/agent/agent.go` 中新增 `generateLLMSummary(messages []*Message) (string, error)` 方法
- [ ] 2.2 新增 `buildSummaryPrompt(messages []*Message) string` 辅助方法，构建 LLM 摘要提示词
- [ ] 2.3 调用 `a.llmSvc.Chat()` 生成对话式摘要
- [ ] 2.4 实现 fallback 逻辑：LLM 调用失败时使用旧的轻量级摘要

## 3. Agent.BuildContextMessages() 协调逻辑

- [ ] 3.1 修改 Agent 调用 ContextManager.BuildContextMessages() 的方式，接收新的三返回值
- [ ] 3.2 当 `needsSummary=true` 时，调用 generateLLMSummary() 获取摘要
- [ ] 3.3 重新调用 ContextManager.BuildContextMessages()，传入摘要结果
- [ ] 3.4 确保最终发送给 LLM 的消息包含：系统提示 + LLM摘要 + 最近上下文

## 4. 测试更新

- [ ] 4.1 更新 `pkg/session/context_test.go` 中的 BuildContextMessages 测试，验证新返回值
- [ ] 4.2 更新 `pkg/agent/agent_test.go` 中的相关测试，验证 LLM 摘要流程
- [ ] 4.3 添加单元测试覆盖 generateLLMSummary() 及其 fallback 逻辑

## 5. 验证

- [ ] 5.1 运行 `go test ./pkg/session/...` 确保通过
- [ ] 5.2 运行 `go test ./pkg/agent/...` 确保通过
- [ ] 5.3 运行 `go build ./...` 确保编译通过