package session

import (
	"errors"
	"fmt"
	"sync"
	"testing"
)

func TestManager_NewManager(t *testing.T) {
	m := NewManager()

	if m == nil {
		t.Fatal("NewManager should not return nil")
	}
	if m.store == nil {
		t.Error("store should be initialized")
	}
}

func TestManager_Errors(t *testing.T) {
	// Ensure error variables are distinct
	errs := []error{
		ErrConversationNotFound,
		ErrConversationAlreadyExists,
		ErrInvalidConversationID,
		ErrMessageEmpty,
	}
	seen := make(map[error]bool)

	for _, err := range errs {
		if seen[err] {
			t.Errorf("duplicate error value detected: %v", err)
		}
		seen[err] = true
	}
}

func TestManager_CreateConversation(t *testing.T) {
	m := NewManager()

	conv, err := m.CreateConversation("session-1", "cluster-a", "default")
	if err != nil {
		t.Fatalf("CreateConversation() error = %v", err)
	}
	if conv == nil {
		t.Fatal("CreateConversation returned nil")
	}
	if conv.ID != "session-1" {
		t.Errorf("ID = %v, want session-1", conv.ID)
	}
	if conv.ClusterName != "cluster-a" {
		t.Errorf("ClusterName = %v, want cluster-a", conv.ClusterName)
	}
	if conv.Namespace != "default" {
		t.Errorf("Namespace = %v, want default", conv.Namespace)
	}
}

func TestManager_CreateConversation_Duplicate(t *testing.T) {
	m := NewManager()

	_, err := m.CreateConversation("dup", "cluster", "ns")
	if err != nil {
		t.Fatalf("First CreateConversation() failed: %v", err)
	}

	_, err = m.CreateConversation("dup", "cluster", "ns")
	if !errors.Is(err, ErrConversationAlreadyExists) {
		t.Errorf("Duplicate error = %v, want ErrConversationAlreadyExists", err)
	}
}

func TestManager_GetConversation(t *testing.T) {
	m := NewManager()

	// Non-existent
	_, err := m.GetConversation("non-existent")
	if !errors.Is(err, ErrConversationNotFound) {
		t.Errorf("GetConversation() = %v, want ErrConversationNotFound", err)
	}

	// Existing
	created, _ := m.CreateConversation("get-test", "cluster", "ns")
	retrieved, err := m.GetConversation("get-test")
	if err != nil {
		t.Fatalf("GetConversation() error = %v", err)
	}
	if retrieved.ID != created.ID {
		t.Errorf("Retrieved ID = %v, want %v", retrieved.ID, created.ID)
	}
}

func TestManager_AddMessage(t *testing.T) {
	m := NewManager()
	m.CreateConversation("msg-test", "cluster", "ns")

	err := m.AddMessage("msg-test", RoleUser, "Hello, how are you?", nil)
	if err != nil {
		t.Fatalf("AddMessage() error = %v", err)
	}

	conv, _ := m.GetConversation("msg-test")
	if len(conv.Messages) != 1 {
		t.Errorf("Messages count = %d, want 1", len(conv.Messages))
	}
	if conv.Messages[0].Content != "Hello, how are you?" {
		t.Errorf("Message content = %v, want 'Hello, how are you?'", conv.Messages[0].Content)
	}
}

func TestManager_AddMessage_EmptyContent(t *testing.T) {
	m := NewManager()
	m.CreateConversation("empty-msg-test", "", "")

	err := m.AddMessage("empty-msg-test", RoleUser, "", nil)
	if !errors.Is(err, ErrMessageEmpty) {
		t.Errorf("AddMessage() with empty content = %v, want ErrMessageEmpty", err)
	}
}

func TestManager_AddMessage_ConversationNotFound(t *testing.T) {
	m := NewManager()

	err := m.AddMessage("non-existent", RoleUser, "Hello", nil)
	if !errors.Is(err, ErrConversationNotFound) {
		t.Errorf("AddMessage() = %v, want ErrConversationNotFound", err)
	}
}

func TestManager_SetClusterContext(t *testing.T) {
	m := NewManager()
	m.CreateConversation("ctx-test", "", "")

	err := m.SetClusterContext("ctx-test", "new-cluster")
	if err != nil {
		t.Fatalf("SetClusterContext() error = %v", err)
	}

	conv, _ := m.GetConversation("ctx-test")
	if conv.ClusterName != "new-cluster" {
		t.Errorf("ClusterName = %v, want new-cluster", conv.ClusterName)
	}
}

func TestManager_SetClusterContext_ConversationNotFound(t *testing.T) {
	m := NewManager()

	err := m.SetClusterContext("non-existent", "cluster")
	if !errors.Is(err, ErrConversationNotFound) {
		t.Errorf("SetClusterContext() = %v, want ErrConversationNotFound", err)
	}
}

func TestManager_SetNamespaceContext(t *testing.T) {
	m := NewManager()
	m.CreateConversation("ns-test", "", "")

	err := m.SetNamespaceContext("ns-test", "kube-system")
	if err != nil {
		t.Fatalf("SetNamespaceContext() error = %v", err)
	}

	conv, _ := m.GetConversation("ns-test")
	if conv.Namespace != "kube-system" {
		t.Errorf("Namespace = %v, want kube-system", conv.Namespace)
	}
}

func TestManager_SetNamespaceContext_ConversationNotFound(t *testing.T) {
	m := NewManager()

	err := m.SetNamespaceContext("non-existent", "ns")
	if !errors.Is(err, ErrConversationNotFound) {
		t.Errorf("SetNamespaceContext() = %v, want ErrConversationNotFound", err)
	}
}

func TestManager_ListConversations(t *testing.T) {
	m := NewManager()

	// Empty list
	convs := m.ListConversations()
	if len(convs) != 0 {
		t.Errorf("Empty manager, ListConversations() = %d, want 0", len(convs))
	}

	// Add some
	m.CreateConversation("conv-1", "cluster-a", "ns-1")
	m.CreateConversation("conv-2", "cluster-b", "ns-2")

	convs = m.ListConversations()
	if len(convs) != 2 {
		t.Errorf("ListConversations() = %d, want 2", len(convs))
	}
}

func TestManager_DeleteConversation(t *testing.T) {
	m := NewManager()
	m.CreateConversation("del-test", "", "")

	// Delete non-existent
	err := m.DeleteConversation("non-existent")
	if !errors.Is(err, ErrConversationNotFound) {
		t.Errorf("DeleteConversation() = %v, want ErrConversationNotFound", err)
	}

	// Delete existing
	err = m.DeleteConversation("del-test")
	if err != nil {
		t.Fatalf("DeleteConversation() error = %v", err)
	}

	// Verify deleted
	_, err = m.GetConversation("del-test")
	if !errors.Is(err, ErrConversationNotFound) {
		t.Errorf("After delete, GetConversation() = %v, want ErrConversationNotFound", err)
	}
}

func TestManager_ConversationWorkflow(t *testing.T) {
	m := NewManager()

	// Create conversation
	_, err := m.CreateConversation("workflow", "prod-cluster", "default")
	if err != nil {
		t.Fatalf("CreateConversation() failed: %v", err)
	}

	// Add messages
	m.AddMessage("workflow", RoleUser, "List pods", nil)
	m.AddMessage("workflow", RoleAssistant, "Here are the pods...", nil)
	m.AddMessage("workflow", RoleUser, "Show me deployments", nil)

	// Change namespace
	m.SetNamespaceContext("workflow", "kube-system")

	// Get and verify
	updated, _ := m.GetConversation("workflow")
	if updated.MessageCount() != 3 {
		t.Errorf("MessageCount = %d, want 3", updated.MessageCount())
	}
	if updated.Namespace != "kube-system" {
		t.Errorf("Namespace = %v, want kube-system", updated.Namespace)
	}
	if updated.ClusterName != "prod-cluster" {
		t.Errorf("ClusterName = %v, want prod-cluster", updated.ClusterName)
	}

	// Delete
	m.DeleteConversation("workflow")
	if len(m.ListConversations()) != 0 {
		t.Error("After delete, ListConversations() should be empty")
	}
}

func TestManager_AddMessage_MultipleRoles(t *testing.T) {
	m := NewManager()
	m.CreateConversation("roles-test", "", "")

	m.AddMessage("roles-test", RoleUser, "Hello", nil)
	m.AddMessage("roles-test", RoleAssistant, "Hi there!", nil)
	m.AddMessage("roles-test", RoleSystem, "You are helpful.", nil)

	conv, _ := m.GetConversation("roles-test")

	if len(conv.Messages) != 3 {
		t.Errorf("Messages count = %d, want 3", len(conv.Messages))
	}

	if conv.Messages[0].Role != RoleUser {
		t.Errorf("First message role = %v, want RoleUser", conv.Messages[0].Role)
	}
	if conv.Messages[1].Role != RoleAssistant {
		t.Errorf("Second message role = %v, want RoleAssistant", conv.Messages[1].Role)
	}
	if conv.Messages[2].Role != RoleSystem {
		t.Errorf("Third message role = %v, want RoleSystem", conv.Messages[2].Role)
	}
}

func TestManager_ConcurrentOperations(t *testing.T) {
	m := NewManager()
	var wg sync.WaitGroup

	// Concurrent creates
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			m.CreateConversation(fmt.Sprintf("conv-%d", id), "", "")
		}(i)
	}

	// Concurrent gets
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			m.GetConversation(fmt.Sprintf("conv-%d", id))
		}(i)
	}

	wg.Wait()

	count := len(m.ListConversations())
	if count != 50 {
		t.Errorf("After concurrent ops, count = %d, want 50", count)
	}
}

func TestManager_AddMessage_WithMetadata(t *testing.T) {
	m := NewManager()
	m.CreateConversation("metadata-test", "", "")

	metadata := map[string]string{
		"source":    "cli",
		"command":   "get pods",
		"timestamp": "1234567890",
	}

	err := m.AddMessage("metadata-test", RoleUser, "Get pods", metadata)
	if err != nil {
		t.Fatalf("AddMessage() error = %v", err)
	}

	conv, _ := m.GetConversation("metadata-test")
	msg := conv.GetLastMessage()
	if msg.Metadata["source"] != "cli" {
		t.Errorf("Metadata[source] = %v, want cli", msg.Metadata["source"])
	}
	if msg.Metadata["command"] != "get pods" {
		t.Errorf("Metadata[command] = %v, want get pods", msg.Metadata["command"])
	}
}
