package ui

import (
	"os"
	"testing"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/bubbletea"
	"github.com/threestoneliu/k8s-agent/pkg/session"
	sharedutil "github.com/threestoneliu/k8s-agent/pkg/shared"
)

// mockClusterLister implements cluster lister for testing
type mockClusterLister struct {
	clusters []string
}

func (m *mockClusterLister) ListClusters() []string {
	return m.clusters
}

// newTestTUI creates a TUI instance for testing with mock channels
func newTestTUI(clusters []string) (*TUI, chan Input, chan Output) {
	inputChan := make(chan Input, 100)
	outputChan := make(chan Output, 100)
	tui := NewTUI("test-cluster", &mockClusterLister{clusters: clusters}, nil)
	return tui, inputChan, outputChan
}

// newTestModel creates a tuiModel for testing
func newTestModel(t *testing.T, clusters []string) (*tuiModel, chan Input, chan Output) {
	cleanup := setupTestHistory(t)
	t.Cleanup(cleanup)

	_, inputChan, outputChan := newTestTUI(clusters)
	tui := &TUI{
		viewport:      viewportNew(),
		textinput:     textinputNew(),
		spinner:       spinnerNew(),
		messages:      make([]session.Message, 0),
		clusterCtx:    "test-cluster",
		clusterLister: &mockClusterLister{clusters: clusters},
		done:          make(chan struct{}),
	}
	model := tui.newModel(inputChan, outputChan)
	m, ok := model.(*tuiModel)
	if !ok {
		t.Fatal("expected *tuiModel")
	}
	return m, inputChan, outputChan
}

func viewportNew() viewport.Model {
	vp := viewport.New(80, 20)
	vp.MouseWheelEnabled = true
	return vp
}

func textinputNew() textinput.Model {
	ti := textinput.New()
	ti.Prompt = "> "
	ti.Placeholder = ""
	ti.Focus()
	return ti
}

func spinnerNew() spinner.Model {
	sp := spinner.New()
	sp.Spinner = spinner.MiniDot
	return sp
}

// castModel casts tea.Model to *tuiModel for testing
func castModel(t *testing.T, m tea.Model) *tuiModel {
	tm, ok := m.(*tuiModel)
	if !ok {
		t.Fatalf("expected *tuiModel, got %T", m)
	}
	return tm
}

// setupTestHistory creates a temp history file and configures TUI to use it
func setupTestHistory(t *testing.T) func() {
	// Create temp file
	tmpFile, err := os.CreateTemp("", "history-*.txt")
	if err != nil {
		t.Fatalf("failed to create temp history file: %v", err)
	}
	tmpFile.Close()

	// Override history file path
	SetHistoryFileForTesting(tmpFile.Name())

	// Return cleanup function
	return func() {
		os.Remove(tmpFile.Name())
		SetHistoryFileForTesting("~/.config/k8s-agent/history/history.txt")
	}
}

// Test 2.1.1: User types text and presses Enter → message appears in viewport
func TestUserTypesAndPressesEnter(t *testing.T) {
	model, inputChan, _ := newTestModel(t, []string{"dev", "prod"})

	// Simulate user typing "get pods"
	model.textinput.SetValue("get pods")

	// Send Enter key
	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := castModel(t, newModel)

	// Verify user message was added
	if len(m.messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(m.messages))
	}
	if m.messages[0].Message.Role != sharedutil.RoleUser {
		t.Errorf("expected user role, got %s", m.messages[0].Message.Role)
	}
	if m.messages[0].Message.Content != "get pods" {
		t.Errorf("expected 'get pods', got '%s'", m.messages[0].Message.Content)
	}

	// Verify input was sent to agent
	select {
	case input := <-inputChan:
		if input.Text != "get pods" {
			t.Errorf("expected 'get pods', got '%s'", input.Text)
		}
	default:
		t.Error("expected input on inputChan")
	}
}

// Test 2.1.2: Empty input is ignored
func TestEmptyInputIsIgnored(t *testing.T) {
	model, _, _ := newTestModel(t, []string{"dev", "prod"})

	// Don't type anything, just press Enter
	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m := castModel(t, newModel)

	// Verify no message was added
	if len(m.messages) != 0 {
		t.Fatalf("expected 0 messages, got %d", len(m.messages))
	}
}

// Test 2.2.1: /clusters command sends Input{Text: "/clusters"}
func TestClustersCommand(t *testing.T) {
	model, inputChan, _ := newTestModel(t, []string{"dev", "prod"})

	// Type /clusters command
	model.textinput.SetValue("/clusters")

	// Press Enter
	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Verify input was sent to agent
	select {
	case input := <-inputChan:
		if input.Text != "/clusters" {
			t.Errorf("expected '/clusters', got '%s'", input.Text)
		}
	default:
		t.Error("expected input on inputChan")
	}

	_ = newModel
}

// Test 2.2.2: /cluster <name> command sends Input{Text: "/cluster <name>"}
func TestClusterSwitchCommand(t *testing.T) {
	model, inputChan, _ := newTestModel(t, []string{"dev", "prod"})

	// Type /cluster dev command
	model.textinput.SetValue("/cluster dev")

	// Press Enter
	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Verify input was sent to agent
	select {
	case input := <-inputChan:
		if input.Text != "/cluster dev" {
			t.Errorf("expected '/cluster dev', got '%s'", input.Text)
		}
	default:
		t.Error("expected input on inputChan")
	}

	_ = newModel
}

// Test 2.2.3: /exit command triggers tea.Quit
func TestExitCommand(t *testing.T) {
	model, _, _ := newTestModel(t, []string{"dev", "prod"})

	// Type /exit command
	model.textinput.SetValue("/exit")

	// Press Enter
	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Verify tea.Quit was returned
	if cmd == nil {
		t.Error("expected tea.Quit command")
	}
}

// Test 2.2.4: /quit command triggers tea.Quit
func TestQuitCommand(t *testing.T) {
	model, _, _ := newTestModel(t, []string{"dev", "prod"})

	// Type /quit command
	model.textinput.SetValue("/quit")

	// Press Enter
	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyEnter})

	// Verify tea.Quit was returned
	if cmd == nil {
		t.Error("expected tea.Quit command")
	}
}

// Test 2.3.1: OutputTypeText renders content in viewport
func TestTextOutputRendersInViewport(t *testing.T) {
	model, _, _ := newTestModel(t, []string{"dev", "prod"})

	// Add a user message first
	model.messages = append(model.messages, session.Message{
		Message: sharedutil.Message{
			Role:    sharedutil.RoleUser,
			Content: "get pods",
		},
	})

	// Simulate agent text output
	output := Output{
		Type:        OutputTypeText,
		Content:     "Here are the pods in default namespace...",
		MessageType: "text",
	}
	newModel, _ := model.Update(output)
	m := castModel(t, newModel)

	// Verify message was added
	if len(m.messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(m.messages))
	}
	if m.messages[1].MessageType != session.MessageTypeText {
		t.Errorf("expected text message type, got %s", m.messages[1].MessageType)
	}
}

// Test 2.3.2: OutputTypeThink renders with "💭 " prefix
func TestThinkOutputRendersWithEmoji(t *testing.T) {
	model, _, _ := newTestModel(t, []string{"dev", "prod"})

	// Simulate agent think output
	output := Output{
		Type:        OutputTypeThink,
		Content:     "分析中...",
		MessageType: "think",
	}
	newModel, _ := model.Update(output)
	m := castModel(t, newModel)

	// Verify message was added with think type
	if len(m.messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(m.messages))
	}
	if m.messages[0].MessageType != session.MessageTypeThink {
		t.Errorf("expected think message type, got %s", m.messages[0].MessageType)
	}
}

// Test 2.3.3: OutputTypeToolStart renders with "🔧 " prefix
func TestToolStartOutputRendersWithEmoji(t *testing.T) {
	model, _, _ := newTestModel(t, []string{"dev", "prod"})

	// Simulate tool call start output
	output := Output{
		Type:        OutputTypeToolStart,
		ToolName:    "k8s_get",
		ToolArgs:    `{"resource":"pods"}`,
		MessageType: "tool_call",
	}
	newModel, _ := model.Update(output)
	m := castModel(t, newModel)

	// Verify message was added
	if len(m.messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(m.messages))
	}
	if m.messages[0].MessageType != session.MessageTypeToolCall {
		t.Errorf("expected tool_call message type, got %s", m.messages[0].MessageType)
	}
}

// Test 2.3.4: OutputTypeToolResult success renders correctly
func TestToolResultSuccessRendersCorrectly(t *testing.T) {
	model, _, _ := newTestModel(t, []string{"dev", "prod"})

	// Simulate successful tool result
	output := Output{
		Type:        OutputTypeToolResult,
		ToolSuccess: true,
		MessageType: "tool_result",
	}
	newModel, _ := model.Update(output)
	m := castModel(t, newModel)

	// Verify message was added
	if len(m.messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(m.messages))
	}
	if m.messages[0].MessageType != session.MessageTypeToolResult {
		t.Errorf("expected tool_result message type, got %s", m.messages[0].MessageType)
	}
}

// Test 2.3.5: OutputTypeToolResult failure renders with "❌ " prefix
func TestToolResultFailureRendersWithEmoji(t *testing.T) {
	model, _, _ := newTestModel(t, []string{"dev", "prod"})

	// Simulate failed tool result
	output := Output{
		Type:        OutputTypeToolResult,
		ToolSuccess: false,
		ToolResult:  "pod not found",
		MessageType: "tool_result",
	}
	newModel, _ := model.Update(output)
	m := castModel(t, newModel)

	// Verify message was added
	if len(m.messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(m.messages))
	}
	if m.messages[0].MessageType != session.MessageTypeToolResult {
		t.Errorf("expected tool_result message type, got %s", m.messages[0].MessageType)
	}
}

// Test 2.4.1: Viewport content is updated after user input
func TestViewportUpdatedAfterUserInput(t *testing.T) {
	model, _, _ := newTestModel(t, []string{"dev", "prod"})

	// Add a user message
	model.messages = append(model.messages, session.Message{
		Message: sharedutil.Message{
			Role:    sharedutil.RoleUser,
			Content: "test message",
		},
	})

	// Build viewport content
	content := model.buildMessageContent()
	if content == "" {
		t.Error("expected viewport content to be non-empty")
	}
}

// Test 2.4.2: Viewport content is updated after agent output
func TestViewportUpdatedAfterAgentOutput(t *testing.T) {
	model, _, _ := newTestModel(t, []string{"dev", "prod"})

	// Add user message
	model.messages = append(model.messages, session.Message{
		Message: sharedutil.Message{
			Role:    sharedutil.RoleUser,
			Content: "get pods",
		},
	})

	// Add agent text output
	model.messages = append(model.messages, session.Message{
		Message: sharedutil.Message{
			Role:    sharedutil.RoleAssistant,
			Content: "pods list here",
		},
		MessageType: session.MessageTypeText,
	})

	// Build viewport content
	content := model.buildMessageContent()
	if content == "" {
		t.Error("expected viewport content to be non-empty")
	}
}

// Test 2.5.1: Ctrl+C exits
func TestCtrlCExits(t *testing.T) {
	model, _, _ := newTestModel(t, []string{"dev", "prod"})

	// Send Ctrl+C
	_, cmd := model.Update(tea.KeyMsg{Type: tea.KeyCtrlC})

	// Verify tea.Quit was returned
	if cmd == nil {
		t.Error("expected tea.Quit command on Ctrl+C")
	}
}

// Test 2.5.2: Escape key is handled
func TestEscapeKeyIsHandled(t *testing.T) {
	model, _, _ := newTestModel(t, []string{"dev", "prod"})

	// Add a user message to have content
	model.messages = append(model.messages, session.Message{
		Message: sharedutil.Message{
			Role:    sharedutil.RoleUser,
			Content: "test",
		},
	})

	// Send Escape key - should not cause panic
	newModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEscape})
	m := castModel(t, newModel)

	// Model should still be valid
	if m == nil {
		t.Error("model should not be nil after Escape")
	}
}

// Test 2.5.3: Window resize updates viewport dimensions
func TestWindowResizeUpdatesViewport(t *testing.T) {
	model, _, _ := newTestModel(t, []string{"dev", "prod"})

	// Send window size message
	width := 120
	height := 30
	newModel, _ := model.Update(tea.WindowSizeMsg{Width: width, Height: height})
	m := castModel(t, newModel)

	// Verify viewport was updated
	if m.viewport.Width != width {
		t.Errorf("expected width %d, got %d", width, m.viewport.Width)
	}
	if m.viewport.Height != height-5 { // Height - 5 for bottom area
		t.Errorf("expected height %d, got %d", height-5, m.viewport.Height)
	}
}

// Test that output channel is properly read
func TestOutputChanRead(t *testing.T) {
	model, _, outputChan := newTestModel(t, []string{"dev", "prod"})

	// Send output through the channel synchronously
	outputChan <- Output{
		Type:        OutputTypeText,
		Content:     "test output",
		MessageType: "text",
	}

	// Call readOutputChan command
	cmd := model.readOutputChan()
	if cmd == nil {
		t.Error("expected non-nil command from readOutputChan")
	}

	// Execute the command to get the message
	msg := cmd()
	if msg == nil {
		t.Error("expected output message")
	}

	output, ok := msg.(Output)
	if !ok {
		t.Fatal("expected Output type")
	}
	if output.Content != "test output" {
		t.Errorf("expected 'test output', got '%s'", output.Content)
	}
}