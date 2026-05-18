package ui

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/threestoneliu/k8s-agent/pkg/session"
	sharedutil "github.com/threestoneliu/k8s-agent/pkg/shared"
)

// precompiled regex for markdown bold
var boldRegex = regexp.MustCompile(`\*\*(.+?)\*\*`)

// ansiEscapeRegex matches ANSI escape sequences including mouse events
var ansiEscapeRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]|\x1b[>=]?[^a-zA-Z]*[a-zA-Z]|\[[<][0-9;]+[a-zA-Z]`)

// markdownRenderer renders markdown to ANSI strings
var markdownRenderer *glamour.TermRenderer

// customMarkdownStyle is a custom style that fixes heading rendering issues
const customMarkdownStyle = `{
  "document": {
    "block_prefix": "\n",
    "block_suffix": "\n",
    "margin": 2
  },
  "paragraph": {},
  "heading": {
    "block_suffix": "\n",
    "color": "39",
    "bold": true
  },
  "h1": {
    "prefix": "### "
  },
  "h2": {
    "prefix": "## "
  },
  "h3": {
    "prefix": "# "
  },
  "h4": {
    "prefix": "# "
  },
  "h5": {
    "prefix": "# "
  },
  "h6": {
    "prefix": "# "
  },
  "text": {},
  "strong": {
    "bold": true
  },
  "emph": {
    "italic": true
  },
  "hr": {
    "color": "240",
    "format": "\n--------\n"
  },
  "item": {
    "block_prefix": "• "
  },
  "enumeration": {
    "block_prefix": ". "
  }
}`

// historyFile is the path to the input history file
const historyFile = "~/.config/k8s-agent/history/history.json"

// expandPath expands ~ to the user's home directory
func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err == nil {
			return filepath.Join(home, path[2:])
		}
	}
	return path
}

// loadHistory loads input history from the history file
func loadHistory() ([]string, error) {
	path := expandPath(historyFile)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return []string{}, err
	}
	var history []string
	if err := json.Unmarshal(data, &history); err != nil {
		return []string{}, nil
	}
	return history, nil
}

// saveHistory saves input history to the history file
func saveHistory(history []string) error {
	path := expandPath(historyFile)
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.Marshal(history)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func init() {
	// Initialize markdown renderer with custom style for proper heading rendering
	r, err := glamour.NewTermRenderer(
		glamour.WithStylesFromJSONBytes([]byte(customMarkdownStyle)),
		glamour.WithWordWrap(0),
	)
	if err == nil {
		markdownRenderer = r
	}
}

// tickMsg is used to signal that readOutputChan should run again
type tickMsg struct{}

// TUI implements UI interface using Bubble Tea
type TUI struct {
	viewport      viewport.Model
	textinput     textinput.Model
	spinner       spinner.Model
	messages      []session.Message
	sending       bool
	clusterCtx    string
	clusterLister interface {
		ListClusters() []string
	}
	err    error
	config *AppConfig
	done   chan struct{}
}

// NewTUI creates a new TUI instance
func NewTUI(clusterCtx string, clusterLister interface {
	ListClusters() []string
}, config *AppConfig) *TUI {
	ti := textinput.New()
	ti.Prompt = "> "
	ti.Placeholder = ""
	ti.Focus()

	sp := spinner.New()
	sp.Spinner = spinner.MiniDot

	vp := viewport.New(80, 20)
	vp.MouseWheelEnabled = true

	return &TUI{
		viewport:      vp,
		textinput:     ti,
		spinner:       sp,
		messages:      make([]session.Message, 0),
		clusterCtx:    clusterCtx,
		clusterLister: clusterLister,
		config:        config,
		done:          make(chan struct{}),
	}
}

// Run implements UI.Run
func (t *TUI) Run(inputChan chan<- Input, outputChan <-chan Output) error {
	// Create channel for UI to receive agent outputs
	// This avoids polling - messages are forwarded by a goroutine
	uiOutputChan := make(chan Output, 100)

	// Start goroutine to forward outputChan messages to uiOutputChan
	go func() {
		for output := range outputChan {
			select {
			case uiOutputChan <- output:
			case <-t.done:
				return
			}
		}
		close(uiOutputChan)
	}()

	model := t.newModelWithOutputChan(inputChan, uiOutputChan)
	// WithMouseCellMotion enables click, release, and wheel events for viewport scrolling
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if err := p.Start(); err != nil {
		return fmt.Errorf("failed to start TUI: %w", err)
	}
	return nil
}

// Close implements UI.Close
func (t *TUI) Close() {
	close(t.done)
}

// newModel creates the Bubble Tea model
func (t *TUI) newModel(inputChan chan<- Input, outputChan <-chan Output) tea.Model {
	history, _ := loadHistory()
	m := &tuiModel{
		tui:           t,
		inputChan:     inputChan,
		outputChan:    outputChan,
		viewport:      t.viewport,
		textinput:     t.textinput,
		spinner:       t.spinner,
		messages:      make([]session.Message, 0),
		styles:        make([]outputStyle, 0),
		sending:       false,
		clusterCtx:    t.clusterCtx,
		err:           nil,
		height:        0,
		history:       history,
	}
	return m
}

// newModelWithOutputChan creates a Bubble Tea model with a custom output channel
func (t *TUI) newModelWithOutputChan(inputChan chan<- Input, outputChan <-chan Output) tea.Model {
	return t.newModel(inputChan, outputChan)
}

// outputStyle defines how assistant content should be styled
type outputStyle int

const (
	styleGray outputStyle = iota // gray color for tool execution info
	styleMarkdown                // markdown rendered
	styleNormal                  // normal text
)

// tuiModel is the Bubble Tea model
type tuiModel struct {
	tui           *TUI
	inputChan     chan<- Input
	outputChan    <-chan Output
	viewport      viewport.Model
	textinput     textinput.Model
	spinner       spinner.Model
	messages      []session.Message
	styles        []outputStyle // styling for each assistant message
	sending       bool
	clusterCtx    string
	err           error
	height        int
	history       []string // stores input history
	historyIndex  int      // current browsing position, -1 means not browsing history
	tempInput     string   // saves current input when browsing history
}

func (m *tuiModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m *tuiModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		// Reserve 5 lines for bottom area (separators, input, cluster)
		m.viewport.Height = msg.Height - 5
		m.viewport.Width = msg.Width
		m.viewport.SetContent(m.buildMessageContent())
		cmds = append(cmds, m.readOutputChan())
		return m, tea.Batch(cmds...)

	case spinner.TickMsg:
		sp, cmd := m.spinner.Update(msg)
		m.spinner = sp
		cmds = append(cmds, cmd)
		// Continue the read loop
		cmds = append(cmds, m.readOutputChan())
		return m, tea.Batch(cmds...)

	case tea.KeyMsg:
		vp, vpCmd := m.viewport.Update(msg)
		m.viewport = vp
		if vpCmd != nil {
			cmds = append(cmds, vpCmd)
		}

		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit

		case tea.KeyEnter:
			value := m.textinput.Value()
			if value == "" {
				cmds = append(cmds, m.readOutputChan())
				return m, tea.Batch(cmds...)
			}

			userInput := strings.TrimSpace(ansiEscapeRegex.ReplaceAllString(value, ""))
			// Ignore empty input after cleaning
			if userInput == "" {
				m.textinput.Reset()
				cmds = append(cmds, m.readOutputChan())
				return m, tea.Batch(cmds...)
			}
			m.textinput.Reset()
			m.viewport.GotoBottom()

			// Save to history if not browsing history (historyIndex == -1)
			if m.historyIndex == -1 {
				m.history = append(m.history, userInput)
				// Limit history to 100 items
				if len(m.history) > 100 {
					m.history = m.history[1:]
				}
				go saveHistory(m.history)
			}

			// Handle exit/quit commands in TUI layer
			if userInput == "/exit" || userInput == "/quit" {
				return m, tea.Quit
			}

			// Handle cluster switch command - reset history browsing state
			if strings.HasPrefix(userInput, "/cluster") {
				m.textinput.Reset()
				m.historyIndex = -1
				cmds = append(cmds, m.readOutputChan())
				return m, tea.Batch(cmds...)
			}

			// Handle clear-history command
			if userInput == "/clear-history" {
				m.history = []string{}
				os.Remove(expandPath(historyFile))
				m.textinput.Reset()
				m.historyIndex = -1
				m.tempInput = ""
				cmds = append(cmds, m.readOutputChan())
				return m, tea.Batch(cmds...)
			}

			// Add user message
			m.messages = append(m.messages, session.Message{
				Message: sharedutil.Message{
					Role:    sharedutil.RoleUser,
					Content: userInput,
				},
			})
			m.viewport.SetContent(m.buildMessageContent())

			m.sending = true
			// Send input to agent (including system commands like /clusters)
			m.inputChan <- Input{Text: userInput}
			cmds = append(cmds, m.readOutputChan())
			return m, tea.Batch(cmds...)

		case tea.KeyEscape:
			cmds = append(cmds, m.readOutputChan())
			return m, tea.Batch(cmds...)

		case tea.KeyUp:
			// Navigate history up (older entries)
			if len(m.history) > 0 {
				if m.historyIndex == -1 {
					// First time browsing history - save current input
					m.tempInput = m.textinput.Value()
				}
				// Move to older history entry
				if m.historyIndex < len(m.history)-1 {
					m.historyIndex++
					m.textinput.SetValue(m.history[len(m.history)-1-m.historyIndex])
				}
			}
			cmds = append(cmds, m.readOutputChan())
			return m, tea.Batch(cmds...)

		case tea.KeyDown:
			// Navigate history down (newer entries)
			if m.historyIndex != -1 {
				if m.historyIndex == 0 {
					// At the newest, restore tempInput
					m.textinput.SetValue(m.tempInput)
					m.historyIndex = -1
				} else {
					// Move to newer history entry
					m.historyIndex--
					m.textinput.SetValue(m.history[len(m.history)-1-m.historyIndex])
				}
			}
			cmds = append(cmds, m.readOutputChan())
			return m, tea.Batch(cmds...)

		default:
			// Check if this looks like a mouse sequence (e.g., [<65;24;33M)
			// Mouse sequences typically start with [< followed by numbers and M
			isMouseSequence := false
			if len(msg.Runes) >= 4 && msg.Runes[0] == '[' && msg.Runes[1] == '<' {
				isMouseSequence = true
				for _, r := range msg.Runes[2:] {
					if !((r >= '0' && r <= '9') || r == ';' || r == 'M') {
						isMouseSequence = false
						break
					}
				}
			}
			if isMouseSequence {
				cmds = append(cmds, m.readOutputChan())
				return m, tea.Batch(cmds...)
			}
			ti, cmd := m.textinput.Update(msg)
			m.textinput = ti
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
			cmds = append(cmds, m.readOutputChan())
			return m, tea.Batch(cmds...)
		}

	case tea.MouseMsg:
		// Let viewport handle mouse events
		vp, vpCmd := m.viewport.Update(msg)
		m.viewport = vp
		if vpCmd != nil {
			cmds = append(cmds, vpCmd)
		}
		cmds = append(cmds, m.readOutputChan())
		return m, tea.Batch(cmds...)

	case Output:
		m.handleOutput(msg)
		m.viewport.SetContent(m.buildMessageContent())
		m.viewport.GotoBottom()
		if msg.Type == OutputTypeDone || msg.Type == OutputTypeError {
			m.sending = false
		}
		// Continue the read loop
		cmds = append(cmds, m.readOutputChan())
		return m, tea.Batch(cmds...)

	case tickMsg:
		// Schedule next read
		cmds = append(cmds, m.readOutputChan())
		return m, tea.Batch(cmds...)
	}

	return m, tea.Batch(cmds...)
}

// readOutputChan checks outputChan for messages
// Returns nil if no message is available, ending the read cycle
func (m *tuiModel) readOutputChan() tea.Cmd {
	return func() tea.Msg {
		select {
		case <-m.tui.done:
			return nil
		case output, ok := <-m.outputChan:
			if !ok {
				return nil
			}
			return output
		default:
			// No message available, stop the read cycle
			// The Output case in Update will trigger the next read when needed
			return nil
		}
	}
}

func (m *tuiModel) handleOutput(output Output) {
	// Determine emoji prefix based on MessageType if present, otherwise fall back to OutputType
	emojiPrefix := ""
	switch output.MessageType {
	case "think":
		emojiPrefix = "💭 "
	case "tool_call":
		emojiPrefix = "🔧 "
	case "tool_result":
		if output.ToolSuccess {
			emojiPrefix = "✅ "
		} else {
			emojiPrefix = "❌ "
		}
	default:
		// Fallback to OutputType-based emoji for backward compatibility
		switch output.Type {
		case OutputTypeThink:
			emojiPrefix = "💭 "
		case OutputTypeToolStart:
			emojiPrefix = "🔧 "
		case OutputTypeToolResult:
			if output.ToolSuccess {
				emojiPrefix = "✅ "
			} else {
				emojiPrefix = "❌ "
			}
		}
	}

	switch output.Type {
	case OutputTypeThink:
		// Think content: gray color, always create new message
		content := output.Content
		if emojiPrefix != "" {
			content = emojiPrefix + content
		}
		m.messages = append(m.messages, session.Message{
			Message: sharedutil.Message{
				Role:    sharedutil.RoleAssistant,
				Content: content,
			},
			MessageType: session.MessageTypeThink,
		})
		m.styles = append(m.styles, styleGray)

	case OutputTypeToolStart:
		// Tool call: gray color, always create new message
		content := fmt.Sprintf("执行工具: %s(%s)", output.ToolName, output.ToolArgs)
		if emojiPrefix != "" {
			content = emojiPrefix + content
		}
		m.messages = append(m.messages, session.Message{
			Message: sharedutil.Message{
				Role:    sharedutil.RoleAssistant,
				Content: content,
			},
			MessageType: session.MessageTypeToolCall,
		})
		m.styles = append(m.styles, styleGray)

	case OutputTypeToolResult:
		// Tool result: gray color, always create new message
		content := "工具执行成功"
		if !output.ToolSuccess {
			content = fmt.Sprintf("工具执行失败: %s", output.ToolResult)
		}
		if emojiPrefix != "" {
			content = emojiPrefix + content
		}
		msg := session.Message{
			Message: sharedutil.Message{
				Role:    sharedutil.RoleAssistant,
				Content: content,
			},
			MessageType: session.MessageTypeToolResult,
		}
		m.messages = append(m.messages, msg)
		m.styles = append(m.styles, styleGray)

	case OutputTypeText:
		// Text content: markdown rendered
		content := output.Content
		if emojiPrefix != "" {
			content = emojiPrefix + content
		}
		msg := session.Message{
			Message: sharedutil.Message{
				Role:    sharedutil.RoleAssistant,
				Content: content,
			},
			MessageType: session.MessageTypeText,
		}
		m.messages = append(m.messages, msg)
		m.styles = append(m.styles, styleMarkdown)

		if output.ClusterName != "" {
			m.clusterCtx = output.ClusterName
		}

	case OutputTypeDone:
		m.sending = false

	case OutputTypeError:
		m.err = fmt.Errorf("%s", output.Content)
		m.sending = false
	}
}

func (m *tuiModel) buildMessageContent() string {
	var sb strings.Builder
	blue := "\x1b[38;5;75m"
	green := "\x1b[38;5;72m"
	purple := "\x1b[38;5;139m"
	gray := "\x1b[38;5;145m"
	reset := "\x1b[0m"

	assistantStarted := false
	styleIdx := 0 // Index into styles array for assistant messages

	for _, msg := range m.messages {
		switch msg.Role {
		case session.RoleUser:
			assistantStarted = false
			sb.WriteString(blue + "You" + reset + ": " + msg.Content + "\n\n")

		case session.RoleAssistant:
			// Only output "Assistant:" once at the start of assistant content
			if !assistantStarted {
				sb.WriteString(green + "Assistant" + reset + ":\n")
				assistantStarted = true
			}

			// Determine styling - use styleIdx to access styles array
			style := styleNormal
			if styleIdx < len(m.styles) {
				style = m.styles[styleIdx]
			}
			styleIdx++

			content := msg.Content

			// Apply markdown rendering if needed
			if style == styleMarkdown && markdownRenderer != nil {
				rendered, err := markdownRenderer.Render(content)
				if err == nil {
					content = rendered
				}
			}

			// Apply gray color for tool execution info
			if style == styleGray {
				sb.WriteString(gray)
			}

			sb.WriteString(content)

			if style == styleGray {
				sb.WriteString(reset)
			}
			sb.WriteString("\n")

		case session.RoleSystem:
			assistantStarted = false
			sb.WriteString(purple + "System" + reset + ": " + gray + msg.Content + reset + "\n\n")
		}
	}

	if m.err != nil {
		sb.WriteString("\x1b[38;5;203mError: " + m.err.Error() + reset + "\n")
	}

	return sb.String()
}

func (m *tuiModel) View() string {
	gray := "\x1b[38;5;145m"
	blue := "\x1b[38;5;75m"
	darkGray := "\x1b[38;5;246m"
	reset := "\x1b[0m"

	var sb strings.Builder

	sb.WriteString(m.viewport.View())

	sb.WriteString("\n" + gray + strings.Repeat("═", 80) + reset + "\n")

	if m.sending {
		sb.WriteString(m.spinner.View() + " ")
	}

	sb.WriteString(m.textinput.View())

	sb.WriteString("\n" + darkGray + strings.Repeat("─", 80) + reset + "\n")

	if m.clusterCtx != "" {
		sb.WriteString(darkGray + "● " + reset + blue + m.clusterCtx + reset)
	}

	// Hide cursor to prevent escape sequences from terminal feedback
	sb.WriteString("\x1b[?25l")

	return sb.String()
}

// AppConfig holds TUI configuration
type AppConfig struct {
	CurrentCluster string
}
