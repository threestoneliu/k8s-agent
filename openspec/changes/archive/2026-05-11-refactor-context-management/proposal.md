## Why

Current `pkg/session/context.go` (~730 lines) uses level-based compression (L1/L2/L3) with overlapping responsibilities. A bug in level1 compression causes premature message dropping even when interactions are within limits. The two parallel implementations for `*Message` and `llm.Message` formats are hard to maintain. Refactoring to interaction-based compression will make the logic clearer and easier to understand.

## What Changes

**Context Compression Logic**

- From: Level-based compression (L1/L2/L3) with complex `findCompleteInteractions` nested conditionals
- To: Interaction-based compression with clear Parse → Evaluate → Compress → Output flow
- Reason: Simpler logic, easier to understand and maintain
- Impact: Non-breaking internal refactor

**Interaction Data Structure**

- From: Messages compressed ad-hoc based on message count and tool call retention
- To: `Interaction { Query, ToolNames, Summary, Completed, OriginalMessages }` as intermediate representation
- Reason: Enables clear separation between parsing, compression decision, and reconstruction
- Impact: Internal data structure change

**Compression Behavior**

- From: Messages could be dropped even when interaction count was within ToolCallRetention limit
- To: Only interactions beyond ToolCallRetention are compressed; recent ones stay fully intact
- Reason: Fix premature compression bug
- Impact: Fixes incorrect behavior

**Compressed Interaction Format**

- From: Old tool calls and results completely discarded
- To: Compressed interactions rebuilt as `[Tool: tool_name]` + summary format
- Reason: Preserve tool execution context for LLM
- Impact: Non-breaking, improved context preservation

## Capabilities

### New Capabilities

- `interaction-compression`: Refactored context compression using interaction as the unit of compression. Replaces the current level-based L1/L2/L3 compression with a clearer Parse → Evaluate → Compress → Output flow.

### Modified Capabilities

- (none - this is internal refactoring of existing context management behavior)

## Impact

**Affected Code:**
- `pkg/session/context.go` - Refactored to use new interaction-based compression
- `pkg/session/interaction.go` - NEW: Interaction struct and parsing logic
- `pkg/session/compressor.go` - NEW: Compression logic
- `pkg/session/context_test.go` - Updated tests

**Dependencies:**
- No external API changes
- No configuration changes
- Session management continues to work unchanged

**Testing:**
- Existing tests updated to match new compression behavior
- New unit tests for Interaction parsing and compression