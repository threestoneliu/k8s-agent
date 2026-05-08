package cli

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/bubbletea"
	"k8s-agent/pkg/agent"
	"k8s-agent/pkg/cluster"
	"k8s-agent/pkg/session"
	"k8s-agent/pkg/ui"
)

// precompiled regex for markdown bold
var boldRegex = regexp.MustCompile(`\*\*(.+?)\*\*`)

// model is the Bubble Tea model for the chat TUI
type model struct {
	viewport     viewport.Model
	textinput    textinput.Model
	spinner      spinner.Model
	messages     []session.Message
	sending      bool
	clusterCtx   string
	agentInstance *agent.Agent
	clusterLister interface {
		ListClusters() []string
	}
	err       error
	appConfig *cluster.AppConfig
	height    int
}

// newTUIModel creates a new chat model with shared agent
func newTUIModel(clusterCtx string, agentInstance *agent.Agent, clusterLister interface {
	ListClusters() []string
}, appConfig *cluster.AppConfig) model {
	ti := textinput.New()
	ti.Prompt = "> "
	ti.Placeholder = ""
	ti.Focus()

	sp := spinner.New()
	sp.Spinner = spinner.MiniDot

	vp := viewport.New(80, 20)
	vp.MouseWheelEnabled = true

	return model{
		viewport:      vp,
		textinput:     ti,
		spinner:       sp,
		messages:      make([]session.Message, 0),
		clusterCtx:    clusterCtx,
		agentInstance: agentInstance,
		clusterLister: clusterLister,
		appConfig:     appConfig,
	}
}

// Init implements tea.Model
func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
	)
}

// buildMessageContent builds the message content for viewport
func (m model) buildMessageContent() string {
	var sb strings.Builder
	blue := "\x1b[38;5;75m"
	green := "\x1b[38;5;72m"
	purple := "\x1b[38;5;139m"
	gray := "\x1b[38;5;145m"
	cyan := "\x1b[38;5;45m"
	orange := "\x1b[38;5;208m"
	white := "\x1b[0m"
	reset := "\x1b[0m"

	for _, msg := range m.messages {
		switch msg.Role {
		case session.RoleUser:
			sb.WriteString(blue + "You" + reset + ": " + white + msg.Content + "\n\n")
		case session.RoleAssistant:
			sb.WriteString(green + "Assistant" + reset + ":\n" + white + msg.Content + reset)

			if msg.Think != "" {
				sb.WriteString("\n\n" + orange + "  🤔 Reasoning:" + reset + "\n")
				thinkLines := strings.Split(msg.Think, "\n")
				for _, line := range thinkLines {
					sb.WriteString("  " + gray + "│ " + reset + orange + line + reset + "\n")
				}
				sb.WriteString("  " + gray + "└" + reset)
			}

			if len(msg.ToolCalls) > 0 {
				sb.WriteString("\n\n" + cyan + "  🔧 Tool Calls:" + reset + "\n")
				for _, tc := range msg.ToolCalls {
					sb.WriteString("  " + gray + "┌─ " + reset + cyan + tc.Name + reset + gray + " ─┐" + reset + "\n")
					sb.WriteString("  " + gray + "│ " + reset + white + tc.Arguments + reset + "\n")
					sb.WriteString("  " + gray + "└" + reset + strings.Repeat("─", len(tc.Name)+4) + reset)
				}
			}

			sb.WriteString("\n\n")
		case session.RoleSystem:
			sb.WriteString(purple + "System" + reset + ": " + gray + msg.Content + reset + "\n\n")
		}
	}

	if m.err != nil {
		sb.WriteString("\x1b[38;5;203mError: " + m.err.Error() + reset + "\n")
	}

	return sb.String()
}

// Update implements tea.Model
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.viewport.Height = msg.Height - 10
		m.viewport.Width = msg.Width
		m.viewport.SetContent(m.buildMessageContent())
		return m, nil

	case spinner.TickMsg:
		sp, cmd := m.spinner.Update(msg)
		m.spinner = sp
		cmds = append(cmds, cmd)

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
				return m, tea.Batch(cmds...)
			}

			userInput := value
			m.textinput.Reset()
			m.viewport.GotoBottom()

			// Handle system commands
			if strings.HasPrefix(userInput, "/") {
				cmd := m.handleSystemCommand(userInput)
				if cmd != nil {
					cmds = append(cmds, cmd)
				}
				return m, tea.Batch(cmds...)
			}

			// Handle cluster switch
			lowerInput := strings.ToLower(userInput)
			if strings.Contains(lowerInput, "切换到") || strings.Contains(lowerInput, "switch to") {
				clusterName := extractClusterName(userInput)
				if clusterName != "" && m.clusterLister != nil {
					clusters := m.clusterLister.ListClusters()
					for _, c := range clusters {
						if c == clusterName {
							m.clusterCtx = clusterName
							if m.agentInstance != nil {
								m.agentInstance.SetClusterName(clusterName)
							}
							m.messages = append(m.messages, session.Message{
								Role:    session.RoleUser,
								Content: userInput,
							})
							m.messages = append(m.messages, session.Message{
								Role:    session.RoleAssistant,
								Content: fmt.Sprintf("已切换到集群 '%s'", clusterName),
							})
							m.viewport.SetContent(m.buildMessageContent())
							return m, nil
						}
					}
					m.messages = append(m.messages, session.Message{
						Role:    session.RoleUser,
						Content: userInput,
					})
					m.messages = append(m.messages, session.Message{
						Role:    session.RoleAssistant,
						Content: fmt.Sprintf("集群 '%s' 不存在。可用集群: %v", clusterName, clusters),
					})
					m.viewport.SetContent(m.buildMessageContent())
					return m, nil
				}
			}

			// Add user message
			m.messages = append(m.messages, session.Message{
				Role:    session.RoleUser,
				Content: userInput,
			})
			m.viewport.SetContent(m.buildMessageContent())

			m.sending = true
			cmds = append(cmds, m.spinner.Tick)
			cmds = append(cmds, m.callAgentCmd())
			return m, tea.Batch(cmds...)

		case tea.KeyEscape:
			return m, nil

		default:
			ti, cmd := m.textinput.Update(msg)
			m.textinput = ti
			if cmd != nil {
				cmds = append(cmds, cmd)
			}
		}

	case tea.MouseMsg:
		vp, vpCmd := m.viewport.Update(msg)
		m.viewport = vp
		if vpCmd != nil {
			cmds = append(cmds, vpCmd)
		}

	case doneMsg:
		m.sending = false
		cmds = append(cmds, m.spinner.Tick)
		m.viewport.SetContent(m.buildMessageContent())
		m.viewport.GotoBottom()

	case errMsg:
		m.err = msg
		m.sending = false
		m.viewport.SetContent(m.buildMessageContent())

	case progressMsg:
		m.handleProgress(msg)
		m.viewport.SetContent(m.buildMessageContent())
		m.viewport.GotoBottom()
	}

	return m, tea.Batch(cmds...)
}

// handleProgress handles progress messages from the agent
func (m *model) handleProgress(msg progressMsg) {
	lastIdx := len(m.messages) - 1

	switch msg.Type {
	case "tool_call_start":
		if lastIdx < 0 || m.messages[lastIdx].Role != session.RoleAssistant {
			m.messages = append(m.messages, session.Message{
				Role:      session.RoleAssistant,
				Content:   "",
				ToolCalls: []session.ToolCall{},
			})
			lastIdx = len(m.messages) - 1
		}
		m.messages[lastIdx].ToolCalls = append(m.messages[lastIdx].ToolCalls, session.ToolCall{
			Name:      msg.ToolName,
			Arguments: msg.ToolArgs,
		})
		m.messages[lastIdx].Content += fmt.Sprintf("\n🔧 执行工具: %s\n", msg.ToolName)
		m.messages[lastIdx].Content += fmt.Sprintf("   参数: %s\n", msg.ToolArgs)

	case "tool_result":
		if lastIdx >= 0 && m.messages[lastIdx].Role == session.RoleAssistant && len(m.messages[lastIdx].ToolCalls) > 0 {
			toolIdx := len(m.messages[lastIdx].ToolCalls) - 1
			m.messages[lastIdx].ToolCalls[toolIdx].Arguments += "\n[Result: " + msg.ToolResult + "]"
			if msg.ToolSuccess {
				m.messages[lastIdx].Content += fmt.Sprintf("✅ 工具完成: %s\n", msg.ToolName)
			} else {
				m.messages[lastIdx].Content += fmt.Sprintf("❌ 工具失败: %s - %s\n", msg.ToolName, msg.ToolResult)
			}
		}

	case "text":
		if lastIdx < 0 || m.messages[lastIdx].Role != session.RoleAssistant {
			m.messages = append(m.messages, session.Message{
				Role:      session.RoleAssistant,
				Content:   "",
				ToolCalls: []session.ToolCall{},
			})
			lastIdx = len(m.messages) - 1
		}
		m.messages[lastIdx].Content += msg.Content

	case "done":
		m.sending = false

	case "error":
		m.err = fmt.Errorf("%s", msg.Content)
		m.sending = false
	}
}

// handleSystemCommand processes slash commands
func (m *model) handleSystemCommand(input string) tea.Cmd {
	switch input {
	case "/clusters":
		m.messages = append(m.messages, session.Message{
			Role:    session.RoleUser,
			Content: input,
		})
		var clusters []string
		if m.clusterLister != nil {
			clusters = m.clusterLister.ListClusters()
		}
		if len(clusters) == 0 {
			clusters = []string{"(no clusters configured)"}
		}
		m.messages = append(m.messages, session.Message{
			Role:    session.RoleAssistant,
			Content: "可用集群:\n" + strings.Join(clusters, "\n"),
		})
		m.viewport.SetContent(m.buildMessageContent())
		return nil

	case "/config":
		m.messages = append(m.messages, session.Message{
			Role:    session.RoleUser,
			Content: input,
		})
		content := formatAppConfig(m.appConfig)
		m.messages = append(m.messages, session.Message{
			Role:    session.RoleAssistant,
			Content: content,
		})
		m.viewport.SetContent(m.buildMessageContent())
		return nil

	case "/exit", "/quit":
		return tea.Quit

	default:
		if strings.HasPrefix(input, "/config ") {
			path := strings.TrimPrefix(input, "/config")
			path = strings.TrimSpace(path)
			path = strings.ToLower(path)

			m.messages = append(m.messages, session.Message{
				Role:    session.RoleUser,
				Content: input,
			})
			content := getConfigValue(m.appConfig, path)
			m.messages = append(m.messages, session.Message{
				Role:    session.RoleAssistant,
				Content: content,
			})
			m.viewport.SetContent(m.buildMessageContent())
			return nil
		}

		if strings.HasPrefix(input, "/cluster ") {
			targetCluster := strings.TrimPrefix(input, "/cluster ")
			targetCluster = strings.TrimSpace(targetCluster)

			m.messages = append(m.messages, session.Message{
				Role:    session.RoleUser,
				Content: input,
			})

			if m.clusterLister == nil {
				m.messages = append(m.messages, session.Message{
					Role:    session.RoleAssistant,
					Content: "集群管理器未初始化",
				})
				m.viewport.SetContent(m.buildMessageContent())
				return nil
			}

			clusters := m.clusterLister.ListClusters()
			found := false
			for _, c := range clusters {
				if c == targetCluster {
					m.clusterCtx = targetCluster
					if m.agentInstance != nil {
						m.agentInstance.SetClusterName(targetCluster)
					}
					m.messages = append(m.messages, session.Message{
						Role:    session.RoleAssistant,
						Content: fmt.Sprintf("已切换到集群 '%s'", targetCluster),
					})
					found = true
					break
				}
			}

			if !found {
				m.messages = append(m.messages, session.Message{
					Role:    session.RoleAssistant,
					Content: fmt.Sprintf("集群 '%s' 不存在。\n可用集群:\n%s", targetCluster, strings.Join(clusters, "\n")),
				})
			}
			m.viewport.SetContent(m.buildMessageContent())
			return nil
		}

		// Unknown command - send to agent
		m.messages = append(m.messages, session.Message{
			Role:    session.RoleUser,
			Content: input,
		})
		m.viewport.SetContent(m.buildMessageContent())
		return m.callAgentCmd()
	}
}

// callAgentCmd creates a command to call the agent
func (m model) callAgentCmd() tea.Cmd {
	var userInput string
	if len(m.messages) > 0 {
		for i := len(m.messages) - 1; i >= 0; i-- {
			if m.messages[i].Role == session.RoleUser {
				userInput = m.messages[i].Content
				break
			}
		}
	}

	if m.agentInstance == nil {
		return func() tea.Msg { return errMsg(fmt.Errorf("agent not configured")) }
	}

	progressChan := make(chan progressMsg, 500)

	go func() {
		uiAdapter := &tuiAdapter{
			messages:     &m.messages,
			progressChan: progressChan,
			model:       &m,
		}

		err := m.agentInstance.ProcessInput(context.Background(), userInput, uiAdapter)
		if err != nil {
			progressChan <- progressMsg{Type: "error", Content: err.Error()}
		}
		close(progressChan)
	}()

	return readProgressChanCmd(progressChan)
}

// readProgressChanCmd reads from channel and dispatches messages to TUI
func readProgressChanCmd(ch <-chan progressMsg) tea.Cmd {
	return func() tea.Msg {
		select {
		case msg, ok := <-ch:
			if !ok {
				return doneMsg{}
			}
			return msg
		default:
			return tea.Tick(15*time.Millisecond, func(time.Time) tea.Msg {
				return readProgressChanCmd(ch)
			})
		}
	}
}

// tuiAdapter adapts the TUI model to the ui.UI interface
type tuiAdapter struct {
	messages     *[]session.Message
	progressChan chan<- progressMsg
	model        *model
}

func (a *tuiAdapter) SendMessage(msg *session.Message) {
	*a.messages = append(*a.messages, *msg)
}

func (a *tuiAdapter) SendProgress(progress ui.Progress) {
	a.progressChan <- progressMsg{
		Type:        progress.Type,
		Content:     progress.Content,
		ToolName:    progress.ToolName,
		ToolArgs:    progress.ToolArgs,
		ToolResult:  progress.ToolResult,
		ToolSuccess: progress.ToolSuccess,
	}
}

func (a *tuiAdapter) Done() {
	a.progressChan <- progressMsg{Type: "done"}
}

func (a *tuiAdapter) Error(err error) {
	a.progressChan <- progressMsg{Type: "error", Content: err.Error()}
}

func (a *tuiAdapter) ClusterName() string {
	return a.model.clusterCtx
}

func (a *tuiAdapter) SetClusterName(clusterName string) {
	a.model.clusterCtx = clusterName
}

// progressMsg is used internally by the TUI
type progressMsg struct {
	Type        string
	Content     string
	ToolName    string
	ToolArgs    string
	ToolResult  string
	ToolSuccess bool
}

// doneMsg signals completion
type doneMsg struct{}

// errMsg signals an error
type errMsg error

// formatAppConfig formats the app config for display
func formatAppConfig(cfg *cluster.AppConfig) string {
	if cfg == nil {
		return "配置信息不可用"
	}

	var sb strings.Builder
	sb.WriteString("当前配置:\n\n")

	sb.WriteString("LLM 配置:\n")
	sb.WriteString(fmt.Sprintf("  Provider:    %s\n", cfg.LLM.Provider))
	sb.WriteString(fmt.Sprintf("  Model:      %s\n", cfg.LLM.Model))
	sb.WriteString(fmt.Sprintf("  BaseURL:    %s\n", cfg.LLM.BaseURL))
	sb.WriteString(fmt.Sprintf("  Timeout:    %.0f秒\n", cfg.LLM.Timeout))
	sb.WriteString(fmt.Sprintf("  Temperature: %.1f\n", cfg.LLM.Temperature))
	sb.WriteString(fmt.Sprintf("  MaxTokens:  %d\n", cfg.LLM.MaxTokens))
	if cfg.LLM.APIKey != "" {
		sb.WriteString("  APIKey:     ****\n")
	} else {
		sb.WriteString("  APIKey:     (未设置)\n")
	}

	sb.WriteString("\nContext 配置:\n")
	sb.WriteString(fmt.Sprintf("  MaxMessages:    %d\n", cfg.Context.MaxMessages))
	sb.WriteString(fmt.Sprintf("  MaxTokens:      %d\n", cfg.Context.MaxTokens))
	sb.WriteString(fmt.Sprintf("  SummaryEnabled: %t\n", cfg.Context.SummaryEnabled))

	sb.WriteString("\n日志配置:\n")
	sb.WriteString(fmt.Sprintf("  Level:  %s\n", cfg.Log.Level))
	sb.WriteString(fmt.Sprintf("  Format: %s\n", cfg.Log.Format))

	sb.WriteString(fmt.Sprintf("\n当前集群: %s\n", cfg.CurrentCluster))

	return sb.String()
}

// getConfigValue returns the value of a specific config path (case-insensitive)
func getConfigValue(cfg *cluster.AppConfig, path string) string {
	if cfg == nil {
		return "配置信息不可用"
	}

	if path == "" {
		return formatAppConfig(cfg)
	}

	switch path {
	case "llm", "llm.provider":
		return cfg.LLM.Provider
	case "llm.model":
		return cfg.LLM.Model
	case "llm.baseurl", "llm.base_url":
		return cfg.LLM.BaseURL
	case "llm.apikey", "llm.api_key":
		if cfg.LLM.APIKey != "" {
			return "****"
		}
		return "(未设置)"
	case "llm.timeout":
		return fmt.Sprintf("%.0f秒", cfg.LLM.Timeout)
	case "llm.temperature":
		return fmt.Sprintf("%.1f", cfg.LLM.Temperature)
	case "llm.maxtokens", "llm.max_tokens":
		return fmt.Sprintf("%d", cfg.LLM.MaxTokens)
	case "context", "context.maxmessages", "context.max_messages":
		return fmt.Sprintf("%d", cfg.Context.MaxMessages)
	case "context.maxtokens", "context.max_tokens":
		return fmt.Sprintf("%d", cfg.Context.MaxTokens)
	case "context.summaryenabled", "context.summary_enabled":
		return fmt.Sprintf("%t", cfg.Context.SummaryEnabled)
	case "log", "log.level":
		return cfg.Log.Level
	case "log.format":
		return cfg.Log.Format
	case "cluster", "currentcluster", "current_cluster":
		return cfg.CurrentCluster
	default:
		return fmt.Sprintf("未知的配置路径: %s", path)
	}
}

// extractClusterName extracts cluster name from user input
func extractClusterName(input string) string {
	if strings.Contains(input, "切换到") {
		parts := strings.Split(input, "切换到")
		if len(parts) > 1 {
			return strings.TrimSpace(parts[1])
		}
	}
	if strings.Contains(input, "switch to") {
		parts := strings.Split(input, "switch to")
		if len(parts) > 1 {
			return strings.TrimSpace(parts[1])
		}
	}
	return ""
}

// View implements tea.Model
func (m model) View() string {
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

	return sb.String()
}
