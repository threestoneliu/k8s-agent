package session

import (
	"testing"

	sharedutil "k8s-agent/pkg/shared"
)

func TestShouldCompress(t *testing.T) {
	interactions := []Interaction{
		{Query: "query1"},
		{Query: "query2"},
		{Query: "query3"},
	}
	// 3 interactions, retention 2 -> should compress
	if !ShouldCompress(interactions, 2) {
		t.Error("expected ShouldCompress to return true")
	}
	// 2 interactions, retention 2 -> should NOT compress
	if ShouldCompress(interactions[:2], 2) {
		t.Error("expected ShouldCompress to return false")
	}
}

func TestReconstructInteraction(t *testing.T) {
	inter := Interaction{
		Query:     "delete pod nginx",
		ToolNames: []string{"k8s_delete"},
		Summary:   "pod deleted successfully",
		Completed: true,
	}
	result := ReconstructInteraction(inter)
	// Should produce: [user query, [Tool: k8s_delete], summary]
	if len(result) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(result))
	}
	if result[0].Message.Role != sharedutil.RoleUser {
		t.Errorf("expected user role, got %v", result[0].Message.Role)
	}
	if result[0].Message.Content != "delete pod nginx" {
		t.Errorf("expected query content, got '%s'", result[0].Message.Content)
	}
	if result[1].Message.Content != "[Tool: k8s_delete]" {
		t.Errorf("expected '[Tool: k8s_delete]', got '%s'", result[1].Message.Content)
	}
	if result[2].Message.Content != "pod deleted successfully" {
		t.Errorf("expected summary, got '%s'", result[2].Message.Content)
	}
}

func TestReconstructInteraction_MultipleTools(t *testing.T) {
	inter := Interaction{
		Query:     "get pods and describe web",
		ToolNames: []string{"k8s_get", "k8s_describe"},
		Summary:   "operations complete",
		Completed: true,
	}
	result := ReconstructInteraction(inter)
	if len(result) != 4 {
		t.Fatalf("expected 4 messages, got %d", len(result))
	}
	if result[1].Message.Content != "[Tool: k8s_get]" {
		t.Errorf("expected '[Tool: k8s_get]', got '%s'", result[1].Message.Content)
	}
	if result[2].Message.Content != "[Tool: k8s_describe]" {
		t.Errorf("expected '[Tool: k8s_describe]', got '%s'", result[2].Message.Content)
	}
}

func TestCompressInteractions_WithinLimits(t *testing.T) {
	interactions := []Interaction{
		{
			Query:            "query1",
			ToolNames:        []string{"tool1"},
			Summary:          "sum1",
			Completed:        true,
			OriginalMessages: []*Message{
				{Message: sharedutil.Message{Role: sharedutil.RoleUser, Content: "query1"}},
				{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "[Function Call: tool1]"}},
				{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "[Tool:result]"}},
				{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "sum1"}},
			},
		},
		{
			Query:            "query2",
			ToolNames:        []string{"tool2"},
			Summary:          "sum2",
			Completed:        true,
			OriginalMessages: []*Message{
				{Message: sharedutil.Message{Role: sharedutil.RoleUser, Content: "query2"}},
				{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "[Function Call: tool2]"}},
				{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "[Tool:result]"}},
				{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "sum2"}},
			},
		},
	}
	// retention=2, 2 interactions -> no compression
	result, count := CompressInteractions(interactions, 2)
	if count != 0 {
		t.Errorf("expected 0 compressed, got %d", count)
	}
	if len(result) != 8 {
		t.Errorf("expected 8 original messages, got %d", len(result))
	}
}

func TestCompressInteractions_OldCompressed(t *testing.T) {
	interactions := []Interaction{
		{
			Query:            "old query",
			ToolNames:        []string{"old_tool"},
			Summary:          "old summary",
			Completed:        true,
			OriginalMessages: []*Message{
				{Message: sharedutil.Message{Role: sharedutil.RoleUser, Content: "old query"}},
				{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "[Function Call: old_tool]"}},
				{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "[Tool:result]"}},
				{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "old summary"}},
			},
		},
		{
			Query:            "recent query",
			ToolNames:        []string{"recent_tool"},
			Summary:          "recent summary",
			Completed:        true,
			OriginalMessages: []*Message{
				{Message: sharedutil.Message{Role: sharedutil.RoleUser, Content: "recent query"}},
				{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "[Function Call: recent_tool]"}},
				{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "[Tool:result]"}},
				{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "recent summary"}},
			},
		},
	}
	// retention=1, 2 interactions -> first should be compressed
	result, count := CompressInteractions(interactions, 1)
	if count != 1 {
		t.Errorf("expected 1 compressed, got %d", count)
	}
	// Compressed: 3 msgs (query + [Tool:] + summary)
	// Recent: 4 msgs (original)
	// Total: 7
	if len(result) != 7 {
		t.Errorf("expected 7 messages, got %d", len(result))
	}
	// Verify first message is old query (compressed)
	if result[0].Message.Content != "old query" {
		t.Errorf("expected first message to be 'old query', got '%s'", result[0].Message.Content)
	}
}

func TestCompressInteractions_IncompleteNotCompressed(t *testing.T) {
	interactions := []Interaction{
		{
			Query:            "incomplete query",
			ToolNames:        []string{"incomplete_tool"},
			Summary:          "",
			Completed:        false, // Incomplete - no tool result yet
			OriginalMessages: []*Message{
				{Message: sharedutil.Message{Role: sharedutil.RoleUser, Content: "incomplete query"}},
				{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "[Function Call: incomplete_tool]"}},
			},
		},
		{
			Query:            "recent query",
			ToolNames:        []string{"recent_tool"},
			Summary:          "recent summary",
			Completed:        true,
			OriginalMessages: []*Message{
				{Message: sharedutil.Message{Role: sharedutil.RoleUser, Content: "recent query"}},
				{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "[Function Call: recent_tool]"}},
				{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "[Tool:result]"}},
				{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "recent summary"}},
			},
		},
	}
	// retention=1, 2 interactions -> first (incomplete) should NOT be compressed
	result, count := CompressInteractions(interactions, 1)
	if count != 0 {
		t.Errorf("expected 0 compressed (incomplete interaction), got %d", count)
	}
	// Both should keep original messages
	if len(result) != 6 {
		t.Errorf("expected 6 messages (both interactions intact), got %d", len(result))
	}
	// Verify first message is incomplete query (not compressed)
	if result[0].Message.Content != "incomplete query" {
		t.Errorf("expected first message to be 'incomplete query', got '%s'", result[0].Message.Content)
	}
	// Verify original format is preserved (not reconstructed)
	if result[1].Message.Content != "[Function Call: incomplete_tool]" {
		t.Errorf("expected '[Function Call: incomplete_tool]', got '%s'", result[1].Message.Content)
	}
}

func TestAddPlaceholder(t *testing.T) {
	messages := []*Message{
		{Message: sharedutil.Message{Role: sharedutil.RoleUser, Content: "query"}},
		{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "summary"}},
	}
	result := AddPlaceholder(messages, 3, 2)
	// Should add system message at end
	if len(result) != 3 {
		t.Errorf("expected 3 messages, got %d", len(result))
	}
	if result[2].Message.Role != sharedutil.RoleSystem {
		t.Errorf("expected system role, got %v", result[2].Message.Role)
	}
	if result[2].Message.Content != "[3 msgs + 2 tool calls condensed]" {
		t.Errorf("unexpected placeholder: '%s'", result[2].Message.Content)
	}
}

func TestCompressInteractions_MessageOrder(t *testing.T) {
	// Create interactions where interaction 1 is old (to be compressed)
	// and interaction 2 is recent (to be kept intact)
	interactions := []Interaction{
		{
			Query:     "old query",
			ToolNames: []string{"old_tool"},
			Summary:   "old summary",
			Completed: true,
		},
		{
			Query: "recent query",
			OriginalMessages: []*Message{
				{Message: sharedutil.Message{Role: sharedutil.RoleUser, Content: "recent query"}},
				{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "recent response"}},
			},
		},
	}
	result, _ := CompressInteractions(interactions, 1)
	// First interaction compressed: [user query, [Tool: old_tool], summary] = 3
	// Second interaction kept: [user query, response] = 2
	if len(result) != 5 {
		t.Errorf("expected 5 messages, got %d", len(result))
	}
	// Verify order: old query first, then recent
	if result[0].Message.Content != "old query" {
		t.Errorf("expected first message to be 'old query', got '%s'", result[0].Message.Content)
	}
	if result[3].Message.Content != "recent query" {
		t.Errorf("expected 4th message to be 'recent query', got '%s'", result[3].Message.Content)
	}
}