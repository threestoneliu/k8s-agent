package ui

import (
	"testing"
	"time"

	"k8s-agent/pkg/session"
)

func TestProgress_Fields(t *testing.T) {
	now := time.Now()
	progress := Progress{
		Type:        "tool_call_start",
		Content:     "test content",
		ToolName:    "get_pods",
		ToolArgs:    `{"namespace":"default"}`,
		ToolResult:  "pods list",
		ToolSuccess: true,
		Timestamp:   now,
	}

	if progress.Type != "tool_call_start" {
		t.Errorf("expected Type 'tool_call_start', got '%s'", progress.Type)
	}
	if progress.ToolName != "get_pods" {
		t.Errorf("expected ToolName 'get_pods', got '%s'", progress.ToolName)
	}
	if progress.ToolSuccess != true {
		t.Errorf("expected ToolSuccess true, got %v", progress.ToolSuccess)
	}
}

func TestMessage_Fields(t *testing.T) {
	now := time.Now()
	msg := Message{
		Role:      session.RoleAssistant,
		Content:   "Hello, how can I help you?",
		ToolCalls: []session.ToolCall{},
		Timestamp: now,
	}

	if msg.Role != session.RoleAssistant {
		t.Errorf("expected Role 'assistant', got '%s'", msg.Role)
	}
	if msg.Content != "Hello, how can I help you?" {
		t.Errorf("expected Content 'Hello, how can I help you?', got '%s'", msg.Content)
	}
	if len(msg.ToolCalls) != 0 {
		t.Errorf("expected 0 ToolCalls, got %d", len(msg.ToolCalls))
	}
}

func TestMessage_WithToolCalls(t *testing.T) {
	msg := Message{
		Role:    session.RoleAssistant,
		Content: "Found 3 pods",
		ToolCalls: []session.ToolCall{
			{Name: "get_pods", Arguments: `{"namespace":"default"}`},
		},
	}

	if len(msg.ToolCalls) != 1 {
		t.Errorf("expected 1 ToolCall, got %d", len(msg.ToolCalls))
	}
	if msg.ToolCalls[0].Name != "get_pods" {
		t.Errorf("expected ToolCall Name 'get_pods', got '%s'", msg.ToolCalls[0].Name)
	}
}