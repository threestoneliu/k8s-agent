## 1. Extend session.Message for LLM compatibility

- [x] 1.1 Add `MessageType` field to `session.Message` (text/think/tool_call/tool_result/user)
- [x] 1.2 Add `ToolCallID` field to `session.Message` for tool response correlation
- [x] 1.3 Ensure `ToolCall` struct in session matches `llm.Message` ToolCalls structure (added ID field)

## 2. Update Agent message synchronization

- [x] 2.1 Modify `processWithOutput()` to sync from `llmMessages` to `messages` with MessageType
- [x] 2.2 Add MessageType inference when creating session messages from LLM responses
- [x] 2.3 Handle think tag parsing with MessageType: think vs text

## 3. Add UI rendering based on MessageType

- [x] 3.1 Update UI output logic to render emoji based on MessageType
- [x] 3.2 Test UI displays correct emoji for think/tool_call/tool_result messages

## 4. Implement session restore with LLM context reconstruction

- [x] 4.1 Modify Agent initialization to load session and reconstruct `llmMessages`
- [x] 4.2 Add mapping from `session.Message` to `llm.Message` format
- [x] 4.3 Test session restore across restart maintains LLM context

## 5. Update FileStore persistence format

- [x] 5.1 Update FileMessage to include new MessageType and ToolCallID fields
- [x] 5.2 Ensure backward compatibility with existing session files (graceful handling of missing fields)

## 6. Verify and test

- [x] 6.1 Write unit tests for message type inference
- [x] 6.2 Integration test: create session, restart, verify LLM context continuation (existing tests pass, session restore implemented)
- [x] 6.3 Run existing tests to ensure no regressions