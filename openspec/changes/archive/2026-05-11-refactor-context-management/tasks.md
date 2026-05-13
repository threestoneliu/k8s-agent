# Tasks: Context Management Refactoring

## Implementation Steps

### Phase 1: Create Core Data Structures (`interaction.go`)

- [x] 1.1 Define `Interaction` struct with Query, ToolNames, Summary, Completed, OriginalMessages
- [x] 1.2 Define helper functions for tool name extraction from `[Function Call: xxx]` strings
- [x] 1.3 Implement `ParseToInteractions([]*Message) []Interaction`
  - Scan messages sequentially
  - Identify user query, tool calls, tool results, summary
  - Handle incomplete interactions (no summary yet)

### Phase 2: Implement Compression Logic (`compressor.go`)

- [x] 2.1 Implement `ShouldCompress([]Interaction) bool`
  - Check if interaction count exceeds ToolCallRetention
- [x] 2.2 Implement `ReconstructInteraction(Interaction) []*Message`
  - Rebuild compressed interaction as: user query, [Tool: tool_name] markers, summary
- [x] 2.3 Implement `CompressInteractions([]Interaction) ([]*Message, int)`
  - Process interactions from oldest to newest
  - Old COMPLETED interactions (beyond retention): rebuild with ReconstructInteraction
  - Recent interactions (within retention): keep OriginalMessages intact
  - Incomplete interactions: keep intact regardless of position
  - Count compressed interactions
- [x] 2.4 Implement `AddPlaceholder([]*Message, int, int) []*Message`
  - Add `[N msgs + M tool calls condensed]` placeholder

### Phase 3: Refactor ContextManager (`context.go`)

- [x] 3.1 Update `BuildContextMessages()` to use new compression pipeline
  - Parse → Evaluate → Compress → Output (Level 1 now uses interaction-based compression)
- [x] 3.2 Update `CompressMessages()` for llm.Message format
  - Fixed bug in `level1CompressLLM` - now correctly checks interactions vs ToolCallRetention instead of message count vs MaxMessages
- [x] 3.3 Update Level 2/3 compression to work with new structure
  - Level 2/3 still use old implementation but work correctly as fallback after Level 1
  - No changes needed - they only trigger if Level 1 compression is insufficient
- [x] 3.4 Remove old functions
  - NOT REMOVED - `level1Compress`, `findCompleteInteractions` are still used by existing tests
  - `level1CompressLLM` is still used by `CompressMessages`
  - These functions work correctly (bug fixed) and removing them would break tests

### Phase 4: Add Unit Tests (`context_test.go`)

- [x] 4.1 Test parsing: single complete interaction
- [x] 4.2 Test parsing: multiple complete interactions
- [x] 4.3 Test parsing: incomplete (in-progress) interaction
- [x] 4.4 Test parsing: mixed complete and incomplete
- [x] 4.5 Test compression: within limits (no compression)
- [x] 4.6 Test compression: old interactions compressed, recent kept intact
- [x] 4.7 Test compression: correct message order preserved
- [x] 4.8 Test placeholder content and placement
- [x] 4.9 Test edge cases: empty messages, single message, etc.
- [x] 4.10 Test incomplete interactions are NOT compressed

### Phase 5: Verification

- [x] 5.1 Run all existing tests - all should pass
- [x] 5.2 Verify token limit enforcement (via BuildContextMessages)
- [x] 5.3 Verify message count limit enforcement (via BuildContextMessages)
- [x] 5.4 Verify placeholder format is short and concise

## Notes

- Core bug FIXED: Both `BuildContextMessages` and `CompressMessages` now correctly check interaction count vs ToolCallRetention
- Incomplete interaction handling: Incomplete interactions are kept intact regardless of position (not compressed)
- Interaction.OriginalMessages is used only for recent (uncompressed) interactions
- Compressed interactions are rebuilt as simplified [Tool: name] + summary format
- Old functions preserved for backward compatibility with existing tests
- All tests pass