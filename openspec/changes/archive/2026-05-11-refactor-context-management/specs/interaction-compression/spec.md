## ADDED Requirements

### Requirement: Interaction Parsing

The ContextManager SHALL parse session messages into Interaction objects by scanning messages sequentially and identifying: user query (user role), tool calls (assistant messages containing "[Function Call:" or "[Tool Call:"), tool results (messages containing "[Tool:" or "Result:"), and final summary (assistant message that is not a tool call).

#### Scenario: Parse complete interaction
- **GIVEN** messages with user query, tool call, tool result, and final summary
- **WHEN** ParseToInteractions is called
- **THEN** a single Interaction is returned with Query, ToolNames, Summary, and Completed=true

#### Scenario: Parse incomplete interaction
- **GIVEN** messages with user query and tool call but no tool result
- **WHEN** ParseToInteractions is called
- **THEN** an Interaction is returned with Completed=false

### Requirement: Interaction Completion Detection

An Interaction SHALL be marked as Completed when a tool result message is encountered after the tool call. The final assistant message that is not a tool call SHALL be set as the Summary.

#### Scenario: Mark interaction complete on tool result
- **GIVEN** an Interaction with user query and tool call
- **WHEN** a message containing "[Tool:" is encountered
- **THEN** Interaction.Completed is set to true

#### Scenario: Set summary on final assistant message
- **GIVEN** an Interaction with user query, tool call, and tool result
- **WHEN** an assistant message that is NOT a tool call is encountered
- **THEN** that message content is set as Interaction.Summary

### Requirement: Compression Trigger

The ContextManager SHALL trigger compression when the number of completed interactions exceeds the ToolCallRetention configuration. Interactions within the retention limit SHALL NOT be compressed.

#### Scenario: No compression when within limits
- **GIVEN** 3 completed interactions and ToolCallRetention=5
- **WHEN** ShouldCompress is called
- **THEN** returns false (no compression needed)

#### Scenario: Compression when exceeding limits
- **GIVEN** 8 completed interactions and ToolCallRetention=5
- **WHEN** ShouldCompress is called
- **THEN** returns true (compression needed)

### Requirement: Recent Interaction Preservation

Recent interactions (within ToolCallRetention count from the newest) SHALL be kept fully intact with their OriginalMessages preserved. The original message format SHALL be maintained exactly.

#### Scenario: Preserve recent interactions with original messages
- **GIVEN** 5 interactions with ToolCallRetention=2
- **WHEN** CompressInteractions is called
- **THEN** the 2 most recent interactions keep their OriginalMessages intact

### Requirement: Old Interaction Compression

Old interactions (beyond ToolCallRetention from the newest) SHALL be compressed to: user query, tool call markers `[Tool: tool_name]` for each tool executed, and final summary. The original detailed tool call parameters and results SHALL be discarded.

#### Scenario: Compress old interaction to simplified format
- **GIVEN** an old completed interaction with detailed tool call and result
- **WHEN** ReconstructInteraction is called
- **THEN** returns [user query, "[Tool: tool_name]", summary]

#### Scenario: Multiple tool calls in one interaction
- **GIVEN** an interaction with tool names ["k8s_get", "k8s_describe"]
- **WHEN** ReconstructInteraction is called
- **THEN** returns [user query, "[Tool: k8s_get]", "[Tool: k8s_describe]", summary]

### Requirement: Placeholder Addition

When old interactions are compressed, the ContextManager SHALL add a system placeholder message: `[N msgs + M tool calls condensed]` where N is the number of messages dropped and M is the number of tool calls dropped.

#### Scenario: Add placeholder after compression
- **GIVEN** 3 old interactions were compressed
- **WHEN** AddPlaceholder is called with msgDropped=9, toolCallDropped=3
- **THEN** a system message "[9 msgs + 3 tool calls condensed]" is appended

### Requirement: Incomplete Interaction Handling

Incomplete interactions (without tool result or summary yet) SHALL be kept as-is without compression, even if beyond ToolCallRetention limit.

#### Scenario: Incomplete interaction not compressed
- **GIVEN** 3 interactions where the first is incomplete (no tool result), and ToolCallRetention=1
- **WHEN** CompressInteractions is called
- **THEN** the incomplete interaction is kept intact with OriginalMessages (not compressed)

#### Scenario: Incomplete interaction at recent position
- **GIVEN** interactions where the most recent is incomplete
- **WHEN** CompressInteractions is called
- **THEN** the incomplete interaction is preserved with full OriginalMessages

---

## MODIFIED Requirements

### Requirement: Level-based Compression Removal

The previous level-based compression (L1/L2/L3) logic SHALL be removed. The functions `findCompleteInteractions`, `level1Compress`, `level1CompressLLM`, and `level2Compress` SHALL be deprecated and replaced with the new interaction-based compression flow.

#### Scenario: Transition from Level-based to Interaction-based
- **WHEN** BuildContextMessages is called with messages that exceed limits
- **THEN** The system SHALL use Parse → Evaluate → Compress → Output flow instead of level-based compression

---

## REMOVED Requirements

### Requirement: Premature Compression Bug

The bug where messages could be dropped even when interaction count was within ToolCallRetention limit SHALL be fixed. The compression decision MUST be based on the interaction count relative to ToolCallRetention, not on raw message count.

**Reason**: This was a bug in the original implementation where `level1Compress` would trigger compression even when `len(interactions) <= keepCount`.

**Migration**: No migration needed - this was incorrect behavior that should not have been relied upon.

### Requirement: Raw Message-based Level1 Threshold

The previous threshold check using `len(messages) <= cm.config.MaxMessages` for triggering level1 compression SHALL be removed. Compression trigger is now based solely on interaction count relative to ToolCallRetention.

**Reason**: Message count is not a reliable indicator for interaction-based compression.

**Migration**: Use the new interaction count check instead.