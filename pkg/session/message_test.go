package session

import (
	"testing"
	"time"
)

func TestMessage_Roles(t *testing.T) {
	tests := []struct {
		name     string
		role     Role
		expected string
	}{
		{"user role", RoleUser, "user"},
		{"assistant role", RoleAssistant, "assistant"},
		{"system role", RoleSystem, "system"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.role) != tt.expected {
				t.Errorf("Role = %v, want %v", tt.role, tt.expected)
			}
		})
	}
}

func TestMessage_NewMessage(t *testing.T) {
	tests := []struct {
		name      string
		role      Role
		content   string
		metadata  map[string]string
		wantPanic bool
	}{
		{
			name:    "user message",
			role:    RoleUser,
			content: "Hello, how are you?",
		},
		{
			name:    "assistant message",
			role:    RoleAssistant,
			content: "I'm doing well, thank you!",
		},
		{
			name:    "system message",
			role:    RoleSystem,
			content: "You are a helpful assistant.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := NewMessage(tt.role, tt.content, tt.metadata)

			if msg.Role != tt.role {
				t.Errorf("Role = %v, want %v", msg.Role, tt.role)
			}
			if msg.Content != tt.content {
				t.Errorf("Content = %v, want %v", msg.Content, tt.content)
			}
			if tt.metadata != nil {
				for k, v := range tt.metadata {
					if msg.Metadata[k] != v {
						t.Errorf("Metadata[%s] = %v, want %v", k, msg.Metadata[k], v)
					}
				}
			}
			if msg.Timestamp.IsZero() {
				t.Error("Timestamp should not be zero")
			}
		})
	}
}

func TestMessage_NewMessage_EmptyContentPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("NewMessage with empty content should panic")
		}
	}()

	NewMessage(RoleUser, "", nil)
}

func TestMessage_TimestampIsSet(t *testing.T) {
	before := time.Now()
	msg := NewMessage(RoleUser, "test", nil)
	after := time.Now()

	if msg.Timestamp.Before(before) || msg.Timestamp.After(after) {
		t.Errorf("Timestamp %v not between %v and %v", msg.Timestamp, before, after)
	}
}

func TestConversation_NewConversation(t *testing.T) {
	conv := NewConversation("test-session", "test-cluster", "default")

	if conv.ID != "test-session" {
		t.Errorf("ID = %v, want test-session", conv.ID)
	}
	if conv.ClusterName != "test-cluster" {
		t.Errorf("ClusterName = %v, want test-cluster", conv.ClusterName)
	}
	if conv.Namespace != "default" {
		t.Errorf("Namespace = %v, want default", conv.Namespace)
	}
	if conv.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
	if conv.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should not be zero")
	}
	if len(conv.Messages) != 0 {
		t.Errorf("Messages should be empty, got %d", len(conv.Messages))
	}
}

func TestConversation_AddMessage(t *testing.T) {
	conv := NewConversation("test-session", "", "")

	userMsg := NewMessage(RoleUser, "Hello", nil)
	assistantMsg := NewMessage(RoleAssistant, "Hi there!", nil)

	conv.AddMessage(userMsg)
	conv.AddMessage(assistantMsg)

	if len(conv.Messages) != 2 {
		t.Errorf("Messages count = %d, want 2", len(conv.Messages))
	}

	if conv.Messages[0].Role != RoleUser {
		t.Errorf("First message role = %v, want RoleUser", conv.Messages[0].Role)
	}
	if conv.Messages[1].Role != RoleAssistant {
		t.Errorf("Second message role = %v, want RoleAssistant", conv.Messages[1].Role)
	}
}

func TestConversation_UpdatedAtChanges(t *testing.T) {
	conv := NewConversation("test-session", "", "")
	originalUpdatedAt := conv.UpdatedAt

	// Small sleep to ensure time difference
	time.Sleep(time.Millisecond)

	userMsg := NewMessage(RoleUser, "Hello", nil)
	conv.AddMessage(userMsg)

	if !conv.UpdatedAt.After(originalUpdatedAt) {
		t.Error("UpdatedAt should be updated after AddMessage")
	}
}

func TestConversation_GetLastMessage(t *testing.T) {
	conv := NewConversation("test-session", "", "")

	// No messages yet
	if conv.GetLastMessage() != nil {
		t.Error("GetLastMessage should return nil when no messages")
	}

	userMsg := NewMessage(RoleUser, "Hello", nil)
	conv.AddMessage(userMsg)

	lastMsg := conv.GetLastMessage()
	if lastMsg == nil {
		t.Error("GetLastMessage should return last message")
	}
	if lastMsg.Content != "Hello" {
		t.Errorf("Last message content = %v, want Hello", lastMsg.Content)
	}
}

func TestConversation_GetMessagesByRole(t *testing.T) {
	conv := NewConversation("test-session", "", "")

	conv.AddMessage(NewMessage(RoleUser, "Hello", nil))
	conv.AddMessage(NewMessage(RoleAssistant, "Hi", nil))
	conv.AddMessage(NewMessage(RoleUser, "How are you?", nil))
	conv.AddMessage(NewMessage(RoleSystem, "Be helpful", nil))

	userMessages := conv.GetMessagesByRole(RoleUser)
	if len(userMessages) != 2 {
		t.Errorf("User messages count = %d, want 2", len(userMessages))
	}

	assistantMessages := conv.GetMessagesByRole(RoleAssistant)
	if len(assistantMessages) != 1 {
		t.Errorf("Assistant messages count = %d, want 1", len(assistantMessages))
	}

	systemMessages := conv.GetMessagesByRole(RoleSystem)
	if len(systemMessages) != 1 {
		t.Errorf("System messages count = %d, want 1", len(systemMessages))
	}
}

func TestConversation_GetClusterContext(t *testing.T) {
	conv := NewConversation("test-session", "", "")
	if conv.GetClusterContext() != "" {
		t.Error("Empty cluster context expected")
	}

	conv.SetClusterContext("prod-cluster")
	if conv.GetClusterContext() != "prod-cluster" {
		t.Errorf("Cluster context = %v, want prod-cluster", conv.GetClusterContext())
	}
}

func TestConversation_SetNamespaceContext(t *testing.T) {
	conv := NewConversation("test-session", "", "")

	conv.SetNamespaceContext("kube-system")
	if conv.GetNamespace() != "kube-system" {
		t.Errorf("Namespace = %v, want kube-system", conv.GetNamespace())
	}
}

func TestConversation_MessageCount(t *testing.T) {
	conv := NewConversation("test-session", "", "")

	if conv.MessageCount() != 0 {
		t.Errorf("Initial count = %d, want 0", conv.MessageCount())
	}

	conv.AddMessage(NewMessage(RoleUser, "Hello", nil))
	conv.AddMessage(NewMessage(RoleAssistant, "Hi", nil))

	if conv.MessageCount() != 2 {
		t.Errorf("Count after adding = %d, want 2", conv.MessageCount())
	}
}
