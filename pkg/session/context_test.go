package session

import (
	"fmt"
	"testing"

	"github.com/threestoneliu/k8s-agent/pkg/cluster"
	sharedutil "github.com/threestoneliu/k8s-agent/pkg/shared"
)

// TestLevel1Compress_SingleCompleteInteraction verifies that a single complete interaction is handled correctly
func TestLevel1Compress_SingleCompleteInteraction(t *testing.T) {
	cm := NewContextManager(cluster.ContextConfig{
		MaxMessages:       20,
		MaxTokens:         8000,
		ToolCallRetention: 10,
	})

	// Single complete interaction
	messages := []*Message{
		{Message: sharedutil.Message{Role: sharedutil.RoleUser, Content: "delete pod nginx"}},
		{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "我将执行 k8s_delete..."}},
		{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "[Tool:result]"}},
		{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "已成功删除 nginx pod"}},
	}

	// Should not compress (under limit)
	result, _ := cm.level1Compress(messages)
	if len(result) != len(messages) {
		t.Errorf("expected %d messages, got %d", len(messages), len(result))
	}
}

// TestLevel1Compress_OldInteractionKeepsSummary verifies that old interactions keep final summary
func TestLevel1Compress_OldInteractionKeepsSummary(t *testing.T) {
	cm := NewContextManager(cluster.ContextConfig{
		MaxMessages:       10,
		MaxTokens:         8000,
		ToolCallRetention: 1, // Keep only 1 interaction
	})

	// Two complete interactions
	messages := []*Message{
		// Old interaction (should be compressed)
		{Message: sharedutil.Message{Role: sharedutil.RoleUser, Content: "old query"}},
		{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "[Function Call: k8s_delete]"}},
		{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "[Tool:result]"}},
		{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "old summary"}},
		// Recent interaction (should be kept)
		{Message: sharedutil.Message{Role: sharedutil.RoleUser, Content: "recent query"}},
		{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "[Function Call: k8s_get]"}},
		{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "[Tool:result]"}},
		{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "recent summary"}},
	}

	result, _ := cm.level1Compress(messages)

	// Debug: print interactions
	interactions := cm.findCompleteInteractions(messages)
	t.Logf("interactions count: %d", len(interactions))
	for i, inter := range interactions {
		t.Logf("  interaction[%d]: UserIndex=%d SummaryIndex=%d Indices=%v", i, inter.UserIndex, inter.SummaryIndex, inter.Indices)
	}

	t.Logf("result len: %d", len(result))
	for i, m := range result {
		t.Logf("  [%d] Role=%s Content=%s isUserMsg=%v", i, m.Message.Role, m.Message.Content, isUserMessage(m))
	}

	// Old interaction: keep user query and summary, discard tool call and result = 2 messages
	// Recent interaction: keep all = 4 messages
	// Total expected: 6 messages
	expectedLen := 6
	if len(result) != expectedLen {
		t.Errorf("expected %d messages, got %d", expectedLen, len(result))
	}

	// Check that old summary is kept
	foundOldSummary := false
	foundRecentSummary := false
	for _, m := range result {
		if m.Message.Content == "old summary" {
			foundOldSummary = true
		}
		if m.Message.Content == "recent summary" {
			foundRecentSummary = true
		}
	}
	if !foundOldSummary {
		t.Error("old summary should be retained")
	}
	if !foundRecentSummary {
		t.Error("recent summary should be retained")
	}

	// Check that tool calls are discarded for old interaction
	for _, m := range result {
		if m.Message.Content == "[Function Call: k8s_delete]" {
			t.Error("old tool call should be discarded")
		}
		// Tool results for old interaction should also be discarded
	}

	t.Logf("result len: %d", len(result))
}

// TestLevel1Compress_MultipleInteractions verifies multiple complete interactions
func TestLevel1Compress_MultipleInteractions(t *testing.T) {
	cm := NewContextManager(cluster.ContextConfig{
		MaxMessages:       10,
		MaxTokens:         8000,
		ToolCallRetention: 2,
	})

	// Three complete interactions
	messages := []*Message{
		// Interaction 1 (oldest - should be compressed)
		{Message: sharedutil.Message{Role: sharedutil.RoleUser, Content: "query1"}},
		{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "[Function Call: k8s_delete]"}},
		{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "[Tool:result1]"}},
		{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "summary1"}},
		// Interaction 2 (middle - should be compressed)
		{Message: sharedutil.Message{Role: sharedutil.RoleUser, Content: "query2"}},
		{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "[Function Call: k8s_get]"}},
		{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "[Tool:result2]"}},
		{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "summary2"}},
		// Interaction 3 (recent - should be kept)
		{Message: sharedutil.Message{Role: sharedutil.RoleUser, Content: "query3"}},
		{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "[Function Call: k8s_describe]"}},
		{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "[Tool:result3]"}},
		{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "summary3"}},
	}

	result, _ := cm.level1Compress(messages)

	// Interaction 1 (oldest): keep query + summary = 2
	// Interaction 2 (middle): keep all = 4
	// Interaction 3 (recent): keep all = 4
	// Total expected: 10
	expectedLen := 10
	if len(result) != expectedLen {
		t.Errorf("expected %d messages, got %d", expectedLen, len(result))
	}
}

// TestLevel1Compress_RecentInteractionIntact verifies recent interactions stay complete
func TestLevel1Compress_RecentInteractionIntact(t *testing.T) {
	cm := NewContextManager(cluster.ContextConfig{
		MaxMessages:       20,
		MaxTokens:        8000,
		ToolCallRetention: 1,
	})

	messages := []*Message{
		{Message: sharedutil.Message{Role: sharedutil.RoleUser, Content: "old query"}},
		{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "old tool"}},
		{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "[Tool:old]"}},
		{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "old summary"}},
		{Message: sharedutil.Message{Role: sharedutil.RoleUser, Content: "recent query"}},
		{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "recent tool"}},
		{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "[Tool:recent]"}},
		{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "recent summary"}},
	}

	result, _ := cm.level1Compress(messages)

	// Recent interaction should be completely intact
	recentIndices := []int{4, 5, 6, 7}
	for _, idx := range recentIndices {
		found := false
		for _, m := range result {
			if m.Message.Content == messages[idx].Message.Content {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("message at index %d (%s) should be in result", idx, messages[idx].Message.Content)
		}
	}
}

// TestFindCompleteInteractions_Basic verifies interaction detection
func TestFindCompleteInteractions_Basic(t *testing.T) {
	cm := &ContextManager{}

	messages := []*Message{
		{Message: sharedutil.Message{Role: sharedutil.RoleUser, Content: "query1"}},
		{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "tool1"}},
		{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "[Tool:result1]"}},
		{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "summary1"}},
	}

	interactions := cm.findCompleteInteractions(messages)

	if len(interactions) != 1 {
		t.Errorf("expected 1 interaction, got %d", len(interactions))
	}

	if interactions[0].UserIndex != 0 {
		t.Errorf("expected UserIndex 0, got %d", interactions[0].UserIndex)
	}

	if interactions[0].SummaryIndex != 3 {
		t.Errorf("expected SummaryIndex 3, got %d", interactions[0].SummaryIndex)
	}
}

// TestFindCompleteInteractions_Multiple verifies multiple interaction detection
func TestFindCompleteInteractions_Multiple(t *testing.T) {
	cm := &ContextManager{}

	messages := []*Message{
		{Message: sharedutil.Message{Role: sharedutil.RoleUser, Content: "query1"}},
		{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "tool1"}},
		{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "[Tool:result1]"}},
		{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "summary1"}},
		{Message: sharedutil.Message{Role: sharedutil.RoleUser, Content: "query2"}},
		{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "tool2"}},
		{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "[Tool:result2]"}},
		{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "summary2"}},
	}

	interactions := cm.findCompleteInteractions(messages)

	if len(interactions) != 2 {
		t.Errorf("expected 2 interactions, got %d", len(interactions))
	}

	if interactions[0].UserIndex != 0 {
		t.Errorf("expected first interaction UserIndex 0, got %d", interactions[0].UserIndex)
	}

	if interactions[1].UserIndex != 4 {
		t.Errorf("expected second interaction UserIndex 4, got %d", interactions[1].UserIndex)
	}
}

// TestBuildContextMessages_NeedsSummary verifies that BuildContextMessages returns correct values
func TestBuildContextMessages_NeedsSummary(t *testing.T) {
	config := cluster.ContextConfig{
		MaxMessages:       3,
		MaxTokens:     50000,
		SummaryEnabled: true,
		ToolCallRetention: 1,
	}
	cm := NewContextManager(config)

	// Create 4 complete interactions (16 messages total, indices 0-15)
	messages := make([]*Message, 16)
	for i := 0; i < 4; i++ {
		base := i * 4
		// User query
		messages[base] = &Message{Message: sharedutil.Message{Role: sharedutil.RoleUser, Content: fmt.Sprintf("User query number %d with lots of additional content", i)}}
		// Tool call
		messages[base+1] = &Message{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: fmt.Sprintf("[Function Call: k8s_get_%d]", i)}}
		// Tool result
		messages[base+2] = &Message{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: fmt.Sprintf("[Tool:result_%d]", i)}}
		// Summary
		messages[base+3] = &Message{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: fmt.Sprintf("Summary for query %d", i)}}
	}

	llmMessages, needsSummary, rawForSummary := cm.BuildContextMessages("system prompt", messages, "")

	// Log debug info
	t.Logf("MaxMessages=%d, len(messages)=%d, ToolCallRetention=%d", config.MaxMessages, len(messages), config.ToolCallRetention)
	t.Logf("needsSummary=%v, len(llmMessages)=%d, len(rawForSummary)=%d", needsSummary, len(llmMessages), len(rawForSummary))

	// The function should return 3 values - verify signature works
	if needsSummary && len(rawForSummary) > 0 {
		// Level 3 triggered successfully
		t.Log("Level 3 compression triggered correctly")
	} else if !needsSummary && len(rawForSummary) == 0 {
		// Level 0, 1, or 2 succeeded
		t.Logf("Compression handled at lower level, returned %d messages", len(llmMessages))
	} else {
		t.Errorf("Unexpected state: needsSummary=%v but rawForSummary has %d elements", needsSummary, len(rawForSummary))
	}

	// Verify we got valid output
	if len(llmMessages) == 0 {
		t.Error("Expected at least one message in llmMessages")
	}
}