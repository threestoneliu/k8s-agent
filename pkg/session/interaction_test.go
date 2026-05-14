package session

import (
	"testing"

	sharedutil "github.com/threestoneliu/k8s-agent/pkg/shared"
)

func TestInteractionStruct(t *testing.T) {
	inter := Interaction{
		Query:     "delete pod nginx",
		ToolNames: []string{"k8s_delete"},
		Summary:   "pod deleted",
		Completed: true,
	}
	if inter.Query != "delete pod nginx" {
		t.Errorf("expected query, got %s", inter.Query)
	}
}

func TestExtractToolName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"[Function Call: k8s_delete(name='nginx')]", "k8s_delete"},
		{"[Function Call: k8s_get(ns='default')]", "k8s_get"},
		{"[Tool Call: k8s_describe]", "k8s_describe"},
		{"some random text", ""},
		{"[Function Call: k8s_delete]", "k8s_delete"},
	}

	for _, tt := range tests {
		result := extractToolName(tt.input)
		if result != tt.expected {
			t.Errorf("extractToolName(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestParseToInteractions_SingleComplete(t *testing.T) {
	messages := []*Message{
		{Message: sharedutil.Message{Role: sharedutil.RoleUser, Content: "delete pod nginx"}},
		{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "[Function Call: k8s_delete(name='nginx')]"}},
		{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "[Tool:result]"}},
		{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "已成功删除"}},
	}
	interactions := ParseToInteractions(messages)
	if len(interactions) != 1 {
		t.Fatalf("expected 1 interaction, got %d", len(interactions))
	}
	if interactions[0].Query != "delete pod nginx" {
		t.Errorf("expected query 'delete pod nginx', got '%s'", interactions[0].Query)
	}
	if interactions[0].ToolNames[0] != "k8s_delete" {
		t.Errorf("expected tool 'k8s_delete', got '%s'", interactions[0].ToolNames[0])
	}
	if !interactions[0].Completed {
		t.Error("expected interaction to be completed")
	}
}

func TestParseToInteractions_Multiple(t *testing.T) {
	messages := []*Message{
		{Message: sharedutil.Message{Role: sharedutil.RoleUser, Content: "query1"}},
		{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "[Function Call: k8s_delete]"}},
		{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "[Tool:result]"}},
		{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "summary1"}},
		{Message: sharedutil.Message{Role: sharedutil.RoleUser, Content: "query2"}},
		{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "[Function Call: k8s_get]"}},
		{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "[Tool:result]"}},
		{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "summary2"}},
	}
	interactions := ParseToInteractions(messages)
	if len(interactions) != 2 {
		t.Fatalf("expected 2 interactions, got %d", len(interactions))
	}
	if interactions[0].Query != "query1" {
		t.Errorf("expected 'query1', got '%s'", interactions[0].Query)
	}
	if interactions[1].Query != "query2" {
		t.Errorf("expected 'query2', got '%s'", interactions[1].Query)
	}
}

func TestParseToInteractions_Incomplete(t *testing.T) {
	messages := []*Message{
		{Message: sharedutil.Message{Role: sharedutil.RoleUser, Content: "query1"}},
		{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "[Function Call: k8s_delete]"}},
		// No tool result or summary yet
	}
	interactions := ParseToInteractions(messages)
	if len(interactions) != 1 {
		t.Fatalf("expected 1 interaction, got %d", len(interactions))
	}
	if interactions[0].Completed {
		t.Error("incomplete interaction should not be marked completed")
	}
	if interactions[0].Summary != "" {
		t.Errorf("expected empty summary, got '%s'", interactions[0].Summary)
	}
}

func TestParseToInteractions_Mixed(t *testing.T) {
	messages := []*Message{
		// Complete interaction 1
		{Message: sharedutil.Message{Role: sharedutil.RoleUser, Content: "query1"}},
		{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "[Function Call: k8s_delete]"}},
		{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "[Tool:result]"}},
		{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "summary1"}},
		// Incomplete interaction 2
		{Message: sharedutil.Message{Role: sharedutil.RoleUser, Content: "query2"}},
		{Message: sharedutil.Message{Role: sharedutil.RoleAssistant, Content: "[Function Call: k8s_get]"}},
	}
	interactions := ParseToInteractions(messages)
	if len(interactions) != 2 {
		t.Fatalf("expected 2 interactions, got %d", len(interactions))
	}
	if !interactions[0].Completed {
		t.Error("first interaction should be completed")
	}
	if interactions[1].Completed {
		t.Error("second interaction should be incomplete")
	}
}