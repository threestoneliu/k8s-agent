## ADDED Requirements

### Requirement: LLM Summary Generation Trigger

When `ContextManager.BuildContextMessages()` is called and Level 3 compression is triggered (i.e., after Level 0/1/2 compression still exceeds max-messages or max-tokens limits), the system SHALL return `needsSummary = true` and include the raw messages requiring summarization in `rawForSummary`.

### Requirement: Agent LLM Summary Invocation

When `Agent.BuildContextMessages()` receives `needsSummary = true` from `ContextManager`, it SHALL call `llmSvc.Chat()` to generate a conversational summary of the messages in `rawForSummary`.

### Requirement: Summary Prompt Construction

The summary prompt SHALL include:
- Instruction to generate a conversational summary in natural language
- Style guidance: describe cluster, operations performed, resources involved, and any dangerous operations
- The raw messages content

### Requirement: Context Rebuild with Summary

After LLM generates the summary, `Agent.BuildContextMessages()` SHALL call `ContextManager.BuildContextMessages()` again with the summary passed as a parameter, such that the final context includes:
1. System prompt
2. LLM-generated conversational summary
3. Recent N messages (controlled by tool-call-retention config)

---

## MODIFIED Requirements

None.

---

## REMOVED Requirements

None.

---

## RENAMED Requirements

None.