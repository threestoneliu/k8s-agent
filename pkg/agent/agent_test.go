package agent

import (
	"context"
	"errors"
	"strings"
	"testing"

	"k8s-agent/pkg/session"
	"k8s-agent/pkg/ui"
)

// mockUI implements ui.UI for testing
type mockUI struct {
	messages    []*session.Message
	progresses  []ui.Progress
	doneCalled  bool
	errorCalled bool
	lastError   error
	clusterName string
}

func (m *mockUI) SendMessage(msg *session.Message) {
	m.messages = append(m.messages, msg)
}

func (m *mockUI) SendProgress(progress ui.Progress) {
	m.progresses = append(m.progresses, progress)
}

func (m *mockUI) Done() {
	m.doneCalled = true
}

func (m *mockUI) Error(err error) {
	m.errorCalled = true
	m.lastError = err
}

func (m *mockUI) ClusterName() string {
	return m.clusterName
}

func (m *mockUI) SetClusterName(clusterName string) {
	m.clusterName = clusterName
}

func TestNewAgent_WithStore(t *testing.T) {
	store := &mockStore{
		conversations: make(map[string]*session.Conversation),
	}
	agent := NewAgent(nil, nil, nil, store, "test-session", "test-cluster", nil)

	if agent == nil {
		t.Fatal("NewAgent should not return nil")
	}
	if agent.sessionID != "test-session" {
		t.Errorf("expected sessionID 'test-session', got '%s'", agent.sessionID)
	}
	if agent.clusterName != "test-cluster" {
		t.Errorf("expected clusterName 'test-cluster', got '%s'", agent.clusterName)
	}
	if agent.messages == nil {
		t.Error("messages should be initialized")
	}
}

func TestNewAgent_WithoutStore(t *testing.T) {
	agent := NewAgent(nil, nil, nil, nil, "test-session", "prod-cluster", nil)

	if agent == nil {
		t.Fatal("NewAgent should not return nil")
	}
	if agent.store != nil {
		t.Error("store should be nil when not provided")
	}
	if agent.clusterName != "prod-cluster" {
		t.Errorf("expected clusterName 'prod-cluster', got '%s'", agent.clusterName)
	}
}

func TestAgent_ProcessInput_NilUI(t *testing.T) {
	agent := NewAgent(nil, nil, nil, nil, "test-session", "default", nil)

	err := agent.ProcessInput(context.Background(), "hello", nil)
	if err == nil {
		t.Error("expected error when ui is nil")
	}
	if !strings.Contains(err.Error(), "ui interface is required") {
		t.Errorf("expected 'ui interface is required' error, got: %v", err)
	}
}

// TestAgent_MessagesAccumulation verifies that messages are accumulated
func TestAgent_MessagesAccumulation(t *testing.T) {
	agent := NewAgent(nil, nil, nil, nil, "test-session", "default", nil)

	// Directly add messages to agent (simulating what ProcessInput does)
	agent.messages = append(agent.messages, session.NewMessage(session.RoleUser, "first message", nil))
	agent.messages = append(agent.messages, session.NewMessage(session.RoleAssistant, "response 1", nil))
	agent.messages = append(agent.messages, session.NewMessage(session.RoleUser, "second message", nil))

	if len(agent.messages) != 3 {
		t.Errorf("expected 3 messages, got %d", len(agent.messages))
	}

	// Verify message order
	if agent.messages[0].Content != "first message" {
		t.Errorf("expected 'first message', got '%s'", agent.messages[0].Content)
	}
	if agent.messages[1].Content != "response 1" {
		t.Errorf("expected 'response 1', got '%s'", agent.messages[1].Content)
	}
	if agent.messages[2].Content != "second message" {
		t.Errorf("expected 'second message', got '%s'", agent.messages[2].Content)
	}
}

func TestAgent_GetMessages(t *testing.T) {
	agent := NewAgent(nil, nil, nil, nil, "test-session", "default", nil)
	agent.messages = []*session.Message{
		session.NewMessage(session.RoleUser, "hello", nil),
		session.NewMessage(session.RoleAssistant, "hi there", nil),
	}

	messages := agent.GetMessages()
	if len(messages) != 2 {
		t.Errorf("expected 2 messages, got %d", len(messages))
	}
}

func TestAgent_GetClusterName(t *testing.T) {
	agent := NewAgent(nil, nil, nil, nil, "test-session", "my-cluster", nil)
	if agent.GetClusterName() != "my-cluster" {
		t.Errorf("expected 'my-cluster', got '%s'", agent.GetClusterName())
	}
}

func TestAgent_SetClusterName(t *testing.T) {
	agent := NewAgent(nil, nil, nil, nil, "test-session", "original", nil)
	agent.SetClusterName("new-cluster")

	if agent.clusterName != "new-cluster" {
		t.Errorf("expected 'new-cluster', got '%s'", agent.clusterName)
	}
}

func TestNewState(t *testing.T) {
	messages := []*session.Message{
		session.NewMessage(session.RoleUser, "hello", nil),
		session.NewMessage(session.RoleAssistant, "hi", nil),
	}
	state := NewState("prod", messages)

	if state.ClusterName != "prod" {
		t.Errorf("expected ClusterName 'prod', got '%s'", state.ClusterName)
	}
	if len(state.SessionMessages) != 2 {
		t.Errorf("expected 2 messages, got %d", len(state.SessionMessages))
	}
}

func TestBuildSystemPrompt_EmptyCluster(t *testing.T) {
	prompt := BuildSystemPrompt("")
	if !strings.Contains(prompt, "default") {
		t.Error("expected 'default' cluster in prompt")
	}
	if !strings.Contains(prompt, "Kubernetes") {
		t.Error("expected 'Kubernetes' in prompt")
	}
}

func TestBuildSystemPrompt_ProdCluster(t *testing.T) {
	prompt := BuildSystemPrompt("prod-cluster")
	if !strings.Contains(prompt, "prod-cluster") {
		t.Error("expected 'prod-cluster' in prompt")
	}
}

func TestBuildSystemPrompt_ChineseContent(t *testing.T) {
	prompt := BuildSystemPrompt("test")
	if !strings.Contains(prompt, "中文") {
		t.Error("expected Chinese content in prompt")
	}
}

// mockStore implements session.StoreInterface for testing
type mockStore struct {
	conversations map[string]*session.Conversation
}

func (m *mockStore) GetConversation(id string) (*session.Conversation, error) {
	if conv, ok := m.conversations[id]; ok {
		return conv, nil
	}
	return nil, nil
}

func (m *mockStore) CreateConversation(id, clusterName, initialMessage string) (*session.Conversation, error) {
	conv := &session.Conversation{
		ID:           id,
		ClusterName:  clusterName,
		Messages:     []*session.Message{},
	}
	m.conversations[id] = conv
	return conv, nil
}

func (m *mockStore) UpdateConversation(id string, update func(*session.Conversation) error) (*session.Conversation, error) {
	conv, ok := m.conversations[id]
	if !ok {
		conv, _ = m.CreateConversation(id, "", "")
	}
	return conv, update(conv)
}

func (m *mockStore) ListConversations() []*session.Conversation {
	result := make([]*session.Conversation, 0, len(m.conversations))
	for _, conv := range m.conversations {
		result = append(result, conv)
	}
	return result
}

func (m *mockStore) DeleteConversation(id string) error {
	delete(m.conversations, id)
	return nil
}

func TestAgent_SetClusterName_WithStore(t *testing.T) {
	store := &mockStore{conversations: make(map[string]*session.Conversation)}
	agent := NewAgent(nil, nil, nil, store, "test-session", "original", nil)

	agent.SetClusterName("new-cluster")

	if agent.clusterName != "new-cluster" {
		t.Errorf("expected 'new-cluster', got '%s'", agent.clusterName)
	}

	// Verify store was updated
	conv, _ := store.GetConversation("test-session")
	if conv.ClusterName != "new-cluster" {
		t.Errorf("expected store clusterName 'new-cluster', got '%s'", conv.ClusterName)
	}
}

func TestAgent_AddMessageToSession(t *testing.T) {
	store := &mockStore{conversations: make(map[string]*session.Conversation)}
	agent := NewAgent(nil, nil, nil, store, "test-session", "default", nil)

	agent.addMessageToSession(session.RoleUser, "test message", nil)

	conv, _ := store.GetConversation("test-session")
	if len(conv.Messages) != 1 {
		t.Errorf("expected 1 message in store, got %d", len(conv.Messages))
	}
	if conv.Messages[0].Content != "test message" {
		t.Errorf("expected 'test message', got '%s'", conv.Messages[0].Content)
	}
}

func TestAgent_AddMessageToSession_NoStore(t *testing.T) {
	agent := NewAgent(nil, nil, nil, nil, "test-session", "default", nil)

	// Should not panic when store is nil
	agent.addMessageToSession(session.RoleUser, "test", nil)
}

func TestUI_Progress_Fields(t *testing.T) {
	progress := ui.Progress{
		Type:        "tool_result",
		Content:     "result content",
		ToolName:    "get_pods",
		ToolArgs:    `{"ns":"default"}`,
		ToolResult:  "pods list",
		ToolSuccess: true,
	}

	if progress.Type != "tool_result" {
		t.Errorf("expected Type 'tool_result', got '%s'", progress.Type)
	}
	if progress.ToolName != "get_pods" {
		t.Errorf("expected ToolName 'get_pods', got '%s'", progress.ToolName)
	}
	if !progress.ToolSuccess {
		t.Error("expected ToolSuccess true")
	}
}

func TestUI_Message_Fields(t *testing.T) {
	msg := ui.Message{
		Role:      session.RoleAssistant,
		Content:   "hello",
		ToolCalls: []session.ToolCall{},
	}

	if msg.Role != session.RoleAssistant {
		t.Errorf("expected RoleAssistant, got %s", msg.Role)
	}
	if msg.Content != "hello" {
		t.Errorf("expected 'hello', got '%s'", msg.Content)
	}
}

func TestMockUI_SendMessage(t *testing.T) {
	mock := &mockUI{}
	msg := session.NewMessage(session.RoleAssistant, "test response", nil)

	mock.SendMessage(msg)

	if len(mock.messages) != 1 {
		t.Errorf("expected 1 message, got %d", len(mock.messages))
	}
	if mock.messages[0].Content != "test response" {
		t.Errorf("expected 'test response', got '%s'", mock.messages[0].Content)
	}
}

func TestMockUI_SendProgress(t *testing.T) {
	mock := &mockUI{}
	progress := ui.Progress{
		Type:      "text",
		Content:   "streaming text",
		ToolName:  "test_tool",
		ToolSuccess: true,
	}

	mock.SendProgress(progress)

	if len(mock.progresses) != 1 {
		t.Errorf("expected 1 progress, got %d", len(mock.progresses))
	}
	if mock.progresses[0].Content != "streaming text" {
		t.Errorf("expected 'streaming text', got '%s'", mock.progresses[0].Content)
	}
}

func TestMockUI_Done(t *testing.T) {
	mock := &mockUI{}
	mock.Done()

	if !mock.doneCalled {
		t.Error("expected doneCalled to be true")
	}
}

func TestMockUI_Error(t *testing.T) {
	mock := &mockUI{}
	err := errors.New("test error")

	mock.Error(err)

	if !mock.errorCalled {
		t.Error("expected errorCalled to be true")
	}
	if mock.lastError.Error() != "test error" {
		t.Errorf("expected 'test error', got '%s'", mock.lastError.Error())
	}
}

func TestMockUI_ClusterName(t *testing.T) {
	mock := &mockUI{clusterName: "prod"}

	if mock.ClusterName() != "prod" {
		t.Errorf("expected 'prod', got '%s'", mock.ClusterName())
	}
}

func TestMockUI_SetClusterName(t *testing.T) {
	mock := &mockUI{}
	mock.SetClusterName("dev")

	if mock.clusterName != "dev" {
		t.Errorf("expected 'dev', got '%s'", mock.clusterName)
	}
}