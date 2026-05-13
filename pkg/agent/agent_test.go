package agent

import (
	"strings"
	"testing"
	"time"

	"k8s-agent/pkg/session"
	sharedutil "k8s-agent/pkg/shared"
	"k8s-agent/pkg/ui"
)

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

func (m *mockStore) CreateConversation(id, clusterName, namespace string) (*session.Conversation, error) {
	conv := &session.Conversation{
		ID:          id,
		ClusterName: clusterName,
		Namespace:   namespace,
		Messages:    []*session.Message{},
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

func TestNewAgent_WithStore(t *testing.T) {
	store := &mockStore{
		conversations: make(map[string]*session.Conversation),
	}
	agent := NewAgent(nil, nil, store, "test-session", "test-cluster", nil)

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
	agent := NewAgent(nil, nil, nil, "test-session", "prod-cluster", nil)

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

func TestAgent_Run_CloseInput(t *testing.T) {
	agent := NewAgent(nil, nil, nil, "test-session", "default", nil)

	inputChan := make(chan ui.Input, 10)
	outputChan := make(chan ui.Output, 10)

	// Close input channel immediately
	close(inputChan)

	// Start agent - should exit when input channel is closed
	done := make(chan struct{})
	go func() {
		agent.Run(inputChan, outputChan)
		close(done)
	}()

	select {
	case <-done:
		// Expected - agent exited
	case <-time.After(100 * time.Millisecond):
		t.Error("agent should have exited when input channel closed")
	}
}

func TestAgent_MessagesAccumulation(t *testing.T) {
	agent := NewAgent(nil, nil, nil, "test-session", "default", nil)

	// Directly add messages to agent (simulating what processInput does)
	agent.messages = append(agent.messages, session.NewMessage(sharedutil.RoleUser, "first message", nil))
	agent.messages = append(agent.messages, session.NewMessage(sharedutil.RoleAssistant, "response 1", nil))
	agent.messages = append(agent.messages, session.NewMessage(sharedutil.RoleUser, "second message", nil))

	if len(agent.messages) != 3 {
		t.Errorf("expected 3 messages, got %d", len(agent.messages))
	}

	// Verify message order
	if agent.messages[0].Message.Content != "first message" {
		t.Errorf("expected 'first message', got '%s'", agent.messages[0].Message.Content)
	}
	if agent.messages[1].Message.Content != "response 1" {
		t.Errorf("expected 'response 1', got '%s'", agent.messages[1].Message.Content)
	}
	if agent.messages[2].Message.Content != "second message" {
		t.Errorf("expected 'second message', got '%s'", agent.messages[2].Message.Content)
	}
}

func TestAgent_GetMessages(t *testing.T) {
	agent := NewAgent(nil, nil, nil, "test-session", "default", nil)
	agent.messages = []*session.Message{
		session.NewMessage(sharedutil.RoleUser, "hello", nil),
		session.NewMessage(sharedutil.RoleAssistant, "hi there", nil),
	}

	messages := agent.GetMessages()
	if len(messages) != 2 {
		t.Errorf("expected 2 messages, got %d", len(messages))
	}
}

func TestAgent_GetClusterName(t *testing.T) {
	agent := NewAgent(nil, nil, nil, "test-session", "my-cluster", nil)
	if agent.GetClusterName() != "my-cluster" {
		t.Errorf("expected 'my-cluster', got '%s'", agent.GetClusterName())
	}
}

func TestAgent_SetClusterName(t *testing.T) {
	agent := NewAgent(nil, nil, nil, "test-session", "original", nil)
	agent.SetClusterName("new-cluster")

	if agent.clusterName != "new-cluster" {
		t.Errorf("expected 'new-cluster', got '%s'", agent.clusterName)
	}
}

func TestNewState(t *testing.T) {
	messages := []*session.Message{
		session.NewMessage(sharedutil.RoleUser, "hello", nil),
		session.NewMessage(sharedutil.RoleAssistant, "hi", nil),
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

func TestAgent_SetClusterName_WithStore(t *testing.T) {
	store := &mockStore{conversations: make(map[string]*session.Conversation)}
	agent := NewAgent(nil, nil, store, "test-session", "original", nil)

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
	agent := NewAgent(nil, nil, store, "test-session", "default", nil)

	testMsg := session.NewMessage(sharedutil.RoleUser, "test message", nil)
	agent.addMessageToSession(testMsg)

	conv, _ := store.GetConversation("test-session")
	if len(conv.Messages) != 1 {
		t.Errorf("expected 1 message in store, got %d", len(conv.Messages))
	}
	if conv.Messages[0].Message.Content != "test message" {
		t.Errorf("expected 'test message', got '%s'", conv.Messages[0].Message.Content)
	}
}

func TestAgent_AddMessageToSession_NoStore(t *testing.T) {
	agent := NewAgent(nil, nil, nil, "test-session", "default", nil)

	// Should not panic when store is nil
	testMsg := session.NewMessage(sharedutil.RoleUser, "test", nil)
	agent.addMessageToSession(testMsg)
}

func TestOutputType_Constants(t *testing.T) {
	if ui.OutputTypeText != "text" {
		t.Errorf("expected OutputTypeText 'text', got '%s'", ui.OutputTypeText)
	}
	if ui.OutputTypeToolStart != "tool_call_start" {
		t.Errorf("expected OutputTypeToolStart 'tool_call_start', got '%s'", ui.OutputTypeToolStart)
	}
	if ui.OutputTypeToolResult != "tool_result" {
		t.Errorf("expected OutputTypeToolResult 'tool_result', got '%s'", ui.OutputTypeToolResult)
	}
	if ui.OutputTypeDone != "done" {
		t.Errorf("expected OutputTypeDone 'done', got '%s'", ui.OutputTypeDone)
	}
	if ui.OutputTypeError != "error" {
		t.Errorf("expected OutputTypeError 'error', got '%s'", ui.OutputTypeError)
	}
}

func TestInput_Fields(t *testing.T) {
	input := ui.Input{
		Text:        "hello",
		ClusterName: "test-cluster",
	}

	if input.Text != "hello" {
		t.Errorf("expected Text 'hello', got '%s'", input.Text)
	}
	if input.ClusterName != "test-cluster" {
		t.Errorf("expected ClusterName 'test-cluster', got '%s'", input.ClusterName)
	}
}

func TestOutput_Fields(t *testing.T) {
	output := ui.Output{
		Type:        ui.OutputTypeToolResult,
		Content:     "result content",
		ToolName:    "get_pods",
		ToolArgs:    `{"ns":"default"}`,
		ToolResult:  "pods list",
		ToolSuccess: true,
		ClusterName: "prod",
	}

	if output.Type != ui.OutputTypeToolResult {
		t.Errorf("expected Type 'tool_result', got '%s'", output.Type)
	}
	if output.ToolName != "get_pods" {
		t.Errorf("expected ToolName 'get_pods', got '%s'", output.ToolName)
	}
	if !output.ToolSuccess {
		t.Error("expected ToolSuccess true")
	}
	if output.ClusterName != "prod" {
		t.Errorf("expected ClusterName 'prod', got '%s'", output.ClusterName)
	}
}

func TestReconstructLLMMessages_UserMessage(t *testing.T) {
	messages := []*session.Message{
		{
			Message: sharedutil.Message{
				Role:    sharedutil.RoleUser,
				Content: "get pods",
			},
			MessageType: session.MessageTypeUser,
			Timestamp:   time.Now(),
		},
	}

	llmMessages := ReconstructLLMMessages(messages, "test-cluster")

	// Should have system prompt + user message
	if len(llmMessages) != 2 {
		t.Errorf("expected 2 messages, got %d", len(llmMessages))
	}
	if llmMessages[0].Role != "system" {
		t.Errorf("expected first message role 'system', got '%s'", llmMessages[0].Role)
	}
	if llmMessages[1].Role != "user" {
		t.Errorf("expected second message role 'user', got '%s'", llmMessages[1].Role)
	}
	if llmMessages[1].Content != "get pods" {
		t.Errorf("expected content 'get pods', got '%s'", llmMessages[1].Content)
	}
}

func TestReconstructLLMMessages_ToolCallMessage(t *testing.T) {
	messages := []*session.Message{
		{
			Message: sharedutil.Message{
				Role:    sharedutil.RoleUser,
				Content: "delete pod nginx",
			},
			MessageType: session.MessageTypeUser,
			Timestamp:   time.Now(),
		},
		{
			Message: sharedutil.Message{
				Role:       sharedutil.RoleAssistant,
				Content:    "执行工具: k8s_delete(...)",
				ToolCallID: "call_123",
				ToolCalls: []sharedutil.ToolCall{
					{ID: "call_123", Name: "k8s_delete", Arguments: `{"name":"nginx"}`},
				},
			},
			MessageType: session.MessageTypeToolCall,
			Timestamp:   time.Now(),
		},
	}

	llmMessages := ReconstructLLMMessages(messages, "test-cluster")

	// Should have system + user + assistant tool call
	if len(llmMessages) != 3 {
		t.Errorf("expected 3 messages, got %d", len(llmMessages))
	}
	// Third message should be assistant with tool calls
	if llmMessages[2].Role != "assistant" {
		t.Errorf("expected third message role 'assistant', got '%s'", llmMessages[2].Role)
	}
	if len(llmMessages[2].ToolCalls) != 1 {
		t.Errorf("expected 1 tool call, got %d", len(llmMessages[2].ToolCalls))
	}
	if llmMessages[2].ToolCalls[0].Name != "k8s_delete" {
		t.Errorf("expected tool call name 'k8s_delete', got '%s'", llmMessages[2].ToolCalls[0].Name)
	}
}

func TestReconstructLLMMessages_ThinkAndTextMessages(t *testing.T) {
	messages := []*session.Message{
		{
			Message: sharedutil.Message{
				Role:    sharedutil.RoleUser,
				Content: "describe pod nginx",
			},
			MessageType: session.MessageTypeUser,
			Timestamp:   time.Now(),
		},
		{
			Message: sharedutil.Message{
				Role:    sharedutil.RoleAssistant,
				Content: "分析中...",
			},
			MessageType: session.MessageTypeThink,
			Timestamp:   time.Now(),
		},
		{
			Message: sharedutil.Message{
				Role:    sharedutil.RoleAssistant,
				Content: "好的，我来描述 pod nginx 的信息",
			},
			MessageType: session.MessageTypeText,
			Timestamp:   time.Now(),
		},
	}

	llmMessages := ReconstructLLMMessages(messages, "test-cluster")

	// Should have system + user + 2 assistant messages
	if len(llmMessages) != 4 {
		t.Errorf("expected 4 messages, got %d", len(llmMessages))
	}
}

func TestReconstructLLMMessages_EmptySession(t *testing.T) {
	messages := []*session.Message{}

	llmMessages := ReconstructLLMMessages(messages, "prod-cluster")

	// Should only have system prompt
	if len(llmMessages) != 1 {
		t.Errorf("expected 1 message (system only), got %d", len(llmMessages))
	}
	if llmMessages[0].Role != "system" {
		t.Errorf("expected first message role 'system', got '%s'", llmMessages[0].Role)
	}
}

func TestNewAgent_RestoresSessionAndLLMContext(t *testing.T) {
	store := &mockStore{
		conversations: make(map[string]*session.Conversation),
	}

	// Create a session with existing messages
	existingMessages := []*session.Message{
		{
			Message: sharedutil.Message{
				Role:    sharedutil.RoleUser,
				Content: "get pods",
			},
			MessageType: session.MessageTypeUser,
			Timestamp:   time.Now(),
		},
		{
			Message: sharedutil.Message{
				Role:    sharedutil.RoleAssistant,
				Content: "OK, 获取 pods 信息",
			},
			MessageType: session.MessageTypeText,
			Timestamp:   time.Now(),
		},
	}
	conv, _ := store.CreateConversation("session-123", "test-cluster", "")
	conv.Messages = existingMessages

	// Create new agent with the store
	agent := NewAgent(nil, nil, store, "session-123", "test-cluster", nil)

	// Verify messages were restored
	if len(agent.messages) != 2 {
		t.Errorf("expected 2 messages restored, got %d", len(agent.messages))
	}

	// Verify llmMessages were reconstructed (system + user + assistant)
	if len(agent.llmMessages) != 3 {
		t.Errorf("expected 3 llmMessages (system + user + assistant), got %d", len(agent.llmMessages))
	}
}

func TestMessageType_Inference(t *testing.T) {
	// Test that session.NewMessage does NOT set MessageType by default
	// (MessageType is set by the calling code in processWithOutput)
	msg := session.NewMessage(sharedutil.RoleUser, "test", nil)

	// NewMessage should not set MessageType (it defaults to empty)
	if msg.MessageType != "" {
		t.Errorf("expected empty MessageType from NewMessage, got '%s'", msg.MessageType)
	}
	if msg.Message.Role != sharedutil.RoleUser {
		t.Errorf("expected RoleUser, got '%s'", msg.Message.Role)
	}
	if msg.Message.Content != "test" {
		t.Errorf("expected content 'test', got '%s'", msg.Message.Content)
	}
}

func TestMessageType_Constants(t *testing.T) {
	if session.MessageTypeUser != "user" {
		t.Errorf("expected MessageTypeUser 'user', got '%s'", session.MessageTypeUser)
	}
	if session.MessageTypeText != "text" {
		t.Errorf("expected MessageTypeText 'text', got '%s'", session.MessageTypeText)
	}
	if session.MessageTypeThink != "think" {
		t.Errorf("expected MessageTypeThink 'think', got '%s'", session.MessageTypeThink)
	}
	if session.MessageTypeToolCall != "tool_call" {
		t.Errorf("expected MessageTypeToolCall 'tool_call', got '%s'", session.MessageTypeToolCall)
	}
	if session.MessageTypeToolResult != "tool_result" {
		t.Errorf("expected MessageTypeToolResult 'tool_result', got '%s'", session.MessageTypeToolResult)
	}
}

func TestToolCall_IDField(t *testing.T) {
	tc := sharedutil.ToolCall{
		ID:        "call_abc123",
		Name:      "k8s_get",
		Arguments: `{"resource":"pods"}`,
	}

	if tc.ID != "call_abc123" {
		t.Errorf("expected ID 'call_abc123', got '%s'", tc.ID)
	}
	if tc.Name != "k8s_get" {
		t.Errorf("expected Name 'k8s_get', got '%s'", tc.Name)
	}
	if tc.Arguments != `{"resource":"pods"}` {
		t.Errorf("expected Arguments '{}', got '%s'", tc.Arguments)
	}
}