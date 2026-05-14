package agent

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
	"k8s-agent/pkg/ipc"
	"k8s-agent/pkg/llm"
	"k8s-agent/pkg/log"
	"k8s-agent/pkg/session"
	sharedutil "k8s-agent/pkg/shared"
)

// parseThinkTags parses text containing <think>xxx</think> tags
// textPart represents a part of text with its type (text or think)
type textPart struct {
	isThink bool
	content string
}

// parseThinkTags parses text and returns parts in original order
func parseThinkTags(text string) []textPart {
	parts := []textPart{}
	thinkStart := "<think>"
	thinkEnd := "</think>"

	for {
		startIdx := strings.Index(text, thinkStart)
		if startIdx == -1 {
			// No more think tags
			if len(text) > 0 {
				parts = append(parts, textPart{isThink: false, content: text})
			}
			break
		}

		// Add text before think tag
		if startIdx > 0 {
			parts = append(parts, textPart{isThink: false, content: text[:startIdx]})
		}

		// Find end of think tag
		endIdx := strings.Index(text[startIdx:], thinkEnd)
		if endIdx == -1 {
			// Unclosed think tag, treat rest as text
			parts = append(parts, textPart{isThink: false, content: text[startIdx:]})
			break
		}

		endIdx += startIdx + len(thinkEnd)
		thinkContent := text[startIdx+len(thinkStart) : endIdx-len(thinkEnd)]
		parts = append(parts, textPart{isThink: true, content: strings.TrimSpace(thinkContent)})

		// Move to after this think tag
		text = text[endIdx:]
	}

	return parts
}

// clusterLister lists available clusters
type clusterLister interface {
	ListClusters() []string
}

// UI message prefixes for display formatting
const (
	uiThinkPrefix    = "💭 "
	uiToolPrefix     = "🔧 "
	uiSuccessPrefix  = "✅ "
	uiErrorPrefix    = "❌ "
)

// parseThinkTags parses text and returns parts in original order

// Agent handles the conversation flow between LLM and UI
type Agent struct {
	llmSvc        *llm.Service
	fnExec        *llm.Executor
	clusterLister  clusterLister
	store          session.StoreInterface
	ctxManager     *session.ContextManager
	sessionID      string
	clusterName    string
	messages       []*session.Message       // UI display messages (with emojis, formatted content)
	llmMessages    []sharedutil.Message            // LLM interaction messages (proper ToolCalls structure)
	configContent  string                  // raw config file content for /config command
}

// NewAgent creates a new Agent instance
func NewAgent(llmSvc *llm.Service, fnExec *llm.Executor, store session.StoreInterface, sessionID, clusterName string, ctxManager *session.ContextManager) *Agent {
	agent := &Agent{
		llmSvc:      llmSvc,
		fnExec:      fnExec,
		store:      store,
		ctxManager: ctxManager,
		sessionID:  sessionID,
		clusterName: clusterName,
		messages:   make([]*session.Message, 0),
		llmMessages: make([]sharedutil.Message, 0),
	}

	// Try to load existing session from store
	if store != nil {
		if conv, err := store.GetConversation(sessionID); err == nil && conv != nil {
			// Session exists, restore messages
			for _, msg := range conv.Messages {
				agent.messages = append(agent.messages, msg)
			}
			// Reconstruct llmMessages from session
			agent.llmMessages = ReconstructLLMMessages(conv.Messages, clusterName)
		} else {
			// Create new session
			store.CreateConversation(sessionID, clusterName, "")
		}
	}

	return agent
}

// ReconstructLLMMessages reconstructs sharedutil.Messages from session messages
func ReconstructLLMMessages(messages []*session.Message, clusterName string) []sharedutil.Message {
	var llmMessages []sharedutil.Message

	// Add system prompt
	systemPrompt := BuildSystemPrompt(clusterName)
	llmMessages = append(llmMessages, sharedutil.Message{Role: "system", Content: systemPrompt})

	for _, msg := range messages {
		switch msg.Role {
		case session.RoleUser:
			llmMessages = append(llmMessages, sharedutil.Message{
				Role:    "user",
				Content: msg.Content,
			})
		case session.RoleAssistant:
			// Check if this is a tool call message
			if len(msg.ToolCalls) > 0 {
				toolCalls := make([]sharedutil.ToolCall, len(msg.ToolCalls))
				for i, tc := range msg.ToolCalls {
					toolCalls[i] = sharedutil.ToolCall{ID: tc.ID, Name: tc.Name, Arguments: tc.Arguments}
				}
				llmMessages = append(llmMessages, sharedutil.Message{
					Role:    "assistant",
					Content: "",
					ToolCalls: toolCalls,
				})
			} else {
				// Regular assistant message
				// Skip UI-formatted content (with emoji prefix) by checking if it's a display format
				content := msg.Content
				// If content starts with emoji patterns, use the actual content portion
				if strings.HasPrefix(content, "💭 ") || strings.HasPrefix(content, "🔧 ") ||
					strings.HasPrefix(content, "✅ ") || strings.HasPrefix(content, "❌ ") {
					// This is a UI-formatted message, extract the actual content
					// For simplicity, just use the raw content since LLM doesn't need emoji
					// The LLM will regenerate appropriate formatting
				}
				llmMessages = append(llmMessages, sharedutil.Message{
					Role:    "assistant",
					Content: content,
				})
			}
		case session.RoleSystem:
			llmMessages = append(llmMessages, sharedutil.Message{
				Role:    "system",
				Content: msg.Content,
			})
		}
	}

	return llmMessages
}

// SetClusterLister sets the cluster lister for Agent
func (a *Agent) SetClusterLister(lister clusterLister) {
	a.clusterLister = lister
}

// SetConfigContent sets the raw config file content for /config command
func (a *Agent) SetConfigContent(content string) {
	a.configContent = content
}

// Run starts the agent's processing loop
// It reads from inputChan, processes messages, and sends results to outputChan
func (a *Agent) Run(inputChan <-chan ipc.Input, outputChan chan<- ipc.Output) {
	defer close(outputChan)
	log.Info("Agent.Run started")

	for input := range inputChan {
		log.Info("Agent received input", "text", input.Text, "cluster", input.ClusterName)
		a.processInput(input, outputChan)
	}
	log.Info("Agent.Run exiting (inputChan closed)")
}

// processInput handles a single input message
func (a *Agent) processInput(input ipc.Input, outputChan chan<- ipc.Output) {
	clusterName := input.ClusterName
	if clusterName == "" {
		clusterName = a.clusterName
	}

	// Update cluster context if changed
	if input.ClusterName != "" && input.ClusterName != a.clusterName {
		a.SetClusterName(input.ClusterName)
	}

	// Add user message to session
	userMsg := session.NewMessage(session.RoleUser, input.Text, nil)
	userMsg.MessageType = session.MessageTypeUser
	a.messages = append(a.messages, userMsg)
	a.addMessageToSession(userMsg)

	// Handle /clusters command directly without LLM
	if input.Text == "/clusters" {
		a.handleClustersCommand(outputChan)
		return
	}

	// Handle /config command directly without LLM
	if strings.HasPrefix(input.Text, "/config") {
		a.handleConfigCommand(input.Text, outputChan)
		return
	}

	// Handle /cluster <name> command to switch cluster
	if strings.HasPrefix(input.Text, "/cluster ") {
		a.handleClusterCommand(input.Text, outputChan)
		return
	}

	// Build state for processing
	state := NewState(clusterName, a.messages)

	// Process with streaming output
	a.processWithOutput(state, outputChan)
}

// formatSessionText formats text for session storage, matching UI display format
func formatSessionText(role string, content string) string {
	switch role {
	case sharedutil.RoleUser:
		return "You: " + content
	case sharedutil.RoleAssistant:
		return "Assistant: " + content
	case sharedutil.RoleSystem:
		return "System: " + content
	default:
		return content
	}
}

// processWithOutput handles LLM function calling with output to channel
func (a *Agent) processWithOutput(state *State, outputChan chan<- ipc.Output) {
	systemPrompt := BuildSystemPrompt(state.ClusterName)

	// For new turn, initialize LLM message list with system prompt
	// Messages from previous turns are already in a.llmMessages
	if len(a.llmMessages) == 0 {
		a.llmMessages = []sharedutil.Message{{Role: "system", Content: systemPrompt}}
	}

	// Add user message from current turn
	a.llmMessages = append(a.llmMessages, sharedutil.Message{
		Role:    "user",
		Content: state.SessionMessages[len(state.SessionMessages)-1].Content,
	})

	// Apply context compression if needed
	if a.ctxManager != nil && len(a.llmMessages) > a.ctxManager.MessageCount() {
		systemPrompt := BuildSystemPrompt(state.ClusterName)
		llmMsgs, err := a.BuildContextMessagesWithSummary(systemPrompt, a.messages)
		if err == nil {
			a.llmMessages = llmMsgs
		}
	}

	log.Info("agent process started", "cluster", state.ClusterName, "messageCount", len(a.llmMessages))

	functions := a.llmSvc.GetFunctions()
	log.Debug("functions available", "count", len(functions), "functions", functions)

	for {
		textResp, fnCall, err := a.llmSvc.ChatWithFunctions(context.Background(), a.llmMessages, functions)
		log.Info("LLM response received", "hasText", textResp != "", "fnCall", fnCall != nil, "err", err)
		if err != nil {
			log.Error("LLM call failed", "error", err)
			outputChan <- ipc.Output{Type: ipc.OutputTypeError, Content: fmt.Sprintf("LLM call failed: %v", err)}
			return
		}

		// Send text response if any
		if textResp != "" {
			// Parse think tags and send parts in original order
			parts := parseThinkTags(textResp)

			for _, part := range parts {
				content := strings.TrimSpace(part.content)
				if len(content) == 0 {
					continue
				}
				if part.isThink {
					outputChan <- ipc.Output{
						Type:    ipc.OutputTypeThink,
						Content: content,
					}
					thinkMsg := session.NewMessage(session.RoleAssistant, content, nil)
					thinkMsg.MessageType = session.MessageTypeThink
					a.messages = append(a.messages, thinkMsg)
					a.addMessageToSession(thinkMsg)
				} else {
					outputChan <- ipc.Output{
						Type:    ipc.OutputTypeText,
						Content: content,
					}
					textMsg := session.NewMessage(session.RoleAssistant, content, nil)
					textMsg.MessageType = session.MessageTypeText
					a.messages = append(a.messages, textMsg)
					a.addMessageToSession(textMsg)
				}
			}
		}

		// If no function call, we're done with this turn
		if fnCall == nil {
			// If we had a text response, add it to LLM messages (single message, no tool calls)
			if textResp != "" {
				a.llmMessages = append(a.llmMessages, sharedutil.Message{
					Role:    "assistant",
					Content: textResp,
				})
			}
			log.Info("Agent sending OutputTypeDone")
			outputChan <- ipc.Output{Type: ipc.OutputTypeDone}
			break
		}

		// Report tool call start
		log.Info("Agent sending OutputTypeToolStart", "name", fnCall.Name)
		outputChan <- ipc.Output{
			Type:        ipc.OutputTypeToolStart,
			ToolName:    fnCall.Name,
			ToolArgs:    fnCall.Arguments,
			MessageType: string(session.MessageTypeToolCall),
		}
		// Record tool start to UI session
		toolCallMsg := &session.Message{
			Message: sharedutil.Message{
				Role:       sharedutil.RoleAssistant,
				Content:    fmt.Sprintf("执行工具: %s(%s)", fnCall.Name, fnCall.Arguments),
				ToolCallID: fnCall.ID,
			},
			MessageType: session.MessageTypeToolCall,
			Timestamp:   time.Now(),
		}
		a.messages = append(a.messages, toolCallMsg)
		a.addMessageToSession(toolCallMsg)

		// Execute function call
		log.Info("Agent executing function call", "name", fnCall.Name, "cluster", state.ClusterName)
		result := a.fnExec.ExecuteFunctionCall(fnCall, state.ClusterName)
		if result.Success {
			if len(result.Result) > 200 {
				log.Info("Agent function call succeeded", "name", fnCall.Name, "result", result.Result[:200]+"...")
			} else {
				log.Info("Agent function call succeeded", "name", fnCall.Name, "result", result.Result)
			}
		} else {
			log.Error("Agent function call failed", "name", fnCall.Name, "cluster", state.ClusterName, "error", result.Error)
		}

		// Handle cluster switch if requested by function
		if result.ClusterSwitch != "" {
			a.SetClusterName(result.ClusterSwitch)
			outputChan <- ipc.Output{
				Type:        ipc.OutputTypeText,
				Content:     fmt.Sprintf("已切换到集群: %s", result.ClusterSwitch),
				ClusterName: result.ClusterSwitch,
			}
		}

		// Report tool result
		log.Info("Agent sending OutputTypeToolResult", "name", fnCall.Name)
		outputChan <- ipc.Output{
			Type:        ipc.OutputTypeToolResult,
			ToolName:    fnCall.Name,
			ToolArgs:    fnCall.Arguments,
			ToolResult:  result.Result,
			ToolSuccess: result.Success,
			MessageType: string(session.MessageTypeToolResult),
		}
		// Record tool result to UI session
		toolResultMsg := &session.Message{
			Message: sharedutil.Message{
				Role:       sharedutil.RoleAssistant,
				Content:    result.Result,
				ToolCallID: fnCall.ID,
			},
			MessageType: session.MessageTypeToolResult,
			Timestamp:   time.Now(),
		}
		a.messages = append(a.messages, toolResultMsg)
		a.addMessageToSession(toolResultMsg)

		// Add function call message to LLM history (proper OpenAI format with ToolCalls)
		a.llmMessages = append(a.llmMessages, sharedutil.Message{
			Role:    "assistant",
			Content: "",
			ToolCalls: []sharedutil.ToolCall{{ID: fnCall.ID, Name: fnCall.Name, Arguments: fnCall.Arguments}},
		})

		// Add function result as tool message to LLM history
		var toolContent string
		if result.Success {
			toolContent = result.Result
		} else {
			toolContent = fmt.Sprintf("Error: %s", result.Error)
		}
		a.llmMessages = append(a.llmMessages, sharedutil.Message{
			Role:       "tool",
			Content:    toolContent,
			ToolCallID: fnCall.ID,
		})
	}
}

// addMessageToSession adds a structured message to the session and persists it
func (a *Agent) addMessageToSession(msg *session.Message) {
	if a.store == nil {
		return
	}

	a.store.UpdateConversation(a.sessionID, func(conv *session.Conversation) error {
		conv.AddMessage(msg)
		return nil
	})
}

// GetMessages returns the conversation history
func (a *Agent) GetMessages() []*session.Message {
	return a.messages
}

// GetClusterName returns the current cluster name
func (a *Agent) GetClusterName() string {
	return a.clusterName
}

// SetClusterName sets the current cluster name
func (a *Agent) SetClusterName(clusterName string) {
	a.clusterName = clusterName
	if a.store != nil {
		a.store.UpdateConversation(a.sessionID, func(conv *session.Conversation) error {
			conv.SetClusterContext(clusterName)
			return nil
		})
	}
}

// handleClustersCommand handles the /clusters command
func (a *Agent) handleClustersCommand(outputChan chan<- ipc.Output) {
	var clusters []string
	if a.clusterLister != nil {
		clusters = a.clusterLister.ListClusters()
	}
	if len(clusters) == 0 {
		clusters = []string{"(no clusters configured)"}
	}

	// Use markdown list format for proper rendering
	content := "可用集群:\n"
	for _, c := range clusters {
		content += "- " + c + "\n"
	}

	outputChan <- ipc.Output{
		Type:    ipc.OutputTypeText,
		Content: content,
	}
	clustersMsg := session.NewMessage(session.RoleAssistant, content, nil)
	clustersMsg.MessageType = session.MessageTypeText
	a.messages = append(a.messages, clustersMsg)
	a.addMessageToSession(clustersMsg)

	outputChan <- ipc.Output{Type: ipc.OutputTypeDone}
}

// handleConfigCommand handles the /config command
// Supports: /config (full config), /config <section> (e.g., /config llm), /config <section>.<field> (e.g., /config llm.apikey)
func (a *Agent) handleConfigCommand(cmd string, outputChan chan<- ipc.Output) {
	content := ""

	if a.configContent == "" {
		content = "(no config file content available)"
	} else {
		// Parse the path from command (e.g., "/config llm.apikey" -> "llm.apikey")
		path := strings.TrimPrefix(cmd, "/config")
		path = strings.TrimSpace(path)

		if path == "" {
			// No path specified, show full config (parse and re-marshal to strip comments)
			content = "```\n" + a.getConfigByPath("") + "\n```"
		} else {
			// Parse YAML and extract the requested path
			content = "```\n" + a.getConfigByPath(path) + "\n```"
		}
	}

	outputChan <- ipc.Output{
		Type:    ipc.OutputTypeText,
		Content: content,
	}
	configMsg := session.NewMessage(session.RoleAssistant, content, nil)
	configMsg.MessageType = session.MessageTypeText
	a.messages = append(a.messages, configMsg)
	a.addMessageToSession(configMsg)

	outputChan <- ipc.Output{Type: ipc.OutputTypeDone}
}

// handleClusterCommand handles the /cluster <name> command to switch cluster
func (a *Agent) handleClusterCommand(cmd string, outputChan chan<- ipc.Output) {
	// Extract cluster name from command
	clusterName := strings.TrimPrefix(cmd, "/cluster ")
	clusterName = strings.TrimSpace(clusterName)

	if clusterName == "" {
		content := "用法: /cluster <集群名称>\n当前集群: " + a.clusterName
		outputChan <- ipc.Output{
			Type:    ipc.OutputTypeText,
			Content: content,
		}
		usageMsg := session.NewMessage(session.RoleAssistant, content, nil)
		usageMsg.MessageType = session.MessageTypeText
		a.messages = append(a.messages, usageMsg)
		a.addMessageToSession(usageMsg)
		outputChan <- ipc.Output{Type: ipc.OutputTypeDone}
		return
	}

	// Validate cluster exists
	if a.clusterLister != nil {
		found := false
		for _, c := range a.clusterLister.ListClusters() {
			if c == clusterName {
				found = true
				break
			}
		}
		if !found {
			content := fmt.Sprintf("集群 '%s' 不存在", clusterName)
			outputChan <- ipc.Output{
				Type:    ipc.OutputTypeText,
				Content: content,
			}
			notFoundMsg := session.NewMessage(session.RoleAssistant, content, nil)
			notFoundMsg.MessageType = session.MessageTypeText
			a.messages = append(a.messages, notFoundMsg)
			a.addMessageToSession(notFoundMsg)
			outputChan <- ipc.Output{Type: ipc.OutputTypeDone}
			return
		}
	}

	// Switch cluster
	a.SetClusterName(clusterName)
	content := fmt.Sprintf("已切换到集群: %s", clusterName)
	outputChan <- ipc.Output{
		Type:        ipc.OutputTypeText,
		Content:     content,
		ClusterName: clusterName,
	}
	switchedMsg := session.NewMessage(session.RoleAssistant, content, nil)
	switchedMsg.MessageType = session.MessageTypeText
	a.messages = append(a.messages, switchedMsg)
	a.addMessageToSession(switchedMsg)

	outputChan <- ipc.Output{Type: ipc.OutputTypeDone}
}

// getConfigByPath extracts a specific path from YAML config
func (a *Agent) getConfigByPath(path string) string {
	// Parse the YAML content
	var result map[string]interface{}
	if err := yaml.Unmarshal([]byte(a.configContent), &result); err != nil {
		return fmt.Sprintf("failed to parse config: %v", err)
	}

	// If path is empty, return full config as key-value pairs
	if path == "" {
		return formatConfigFlat(result, "")
	}

	// Navigate through the path (e.g., "llm.apikey" -> result["llm"]["apikey"])
	parts := strings.Split(path, ".")
	current := interface{}(result)

	for i, part := range parts {
		if current == nil {
			return fmt.Sprintf("path not found: %s (at '%s')", path, strings.Join(parts[:i], "."))
		}

		if m, ok := current.(map[string]interface{}); ok {
			current = m[part]
		} else {
			return fmt.Sprintf("path not found: %s (at '%s' - not a map)", path, strings.Join(parts[:i], "."))
		}
	}

	if current == nil {
		return fmt.Sprintf("path not found: %s", path)
	}

	// Format the result
	switch v := current.(type) {
	case string:
		if v == "" {
			return fmt.Sprintf("%s: (empty)", path)
		}
		return fmt.Sprintf("%s: %s", path, v)
	case int, float64, bool:
		return fmt.Sprintf("%s: %v", path, v)
	case map[string]interface{}:
		// If it's a map, show it as flat key-value pairs
		return formatConfigFlat(v, path)
	default:
		return fmt.Sprintf("%s: %v", path, v)
	}
}

// formatYAMLOutput formats compact YAML into properly indented multi-line output
func formatYAMLOutput(compact string) string {
	var buf bytes.Buffer
	lines := strings.Split(compact, "\n")
	for _, line := range lines {
		trimmed := strings.TrimRight(line, " ")
		if trimmed != "" {
			buf.WriteString(trimmed + "\n")
		}
	}
	return buf.String()
}

// formatConfigFlat formats config as flat key-value pairs for better readability
func formatConfigFlat(data map[string]interface{}, prefix string) string {
	var result string
	if prefix != "" {
		result = prefix + ":\n"
	} else {
		result = "配置内容:\n"
	}
	for key, value := range data {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}
		switch v := value.(type) {
		case string:
			result += fmt.Sprintf("  %s: %s\n", key, v)
		case int, float64, bool:
			result += fmt.Sprintf("  %s: %v\n", key, v)
		case map[string]interface{}:
			result += formatConfigFlat(v, fullKey)
		}
	}
	return result
}

// generateLLMSummary generates a conversational summary using LLM
func (a *Agent) generateLLMSummary(messages []*session.Message) (string, error) {
	if a.llmSvc == nil {
		return a.fallbackSummary(messages), nil
	}

	prompt := a.buildSummaryPrompt(messages)
	resp, err := a.llmSvc.Chat(context.Background(), []sharedutil.Message{
		{Role: "user", Content: prompt},
	})
	if err != nil {
		log.Warn("LLM summary failed, using fallback", "error", err)
		return a.fallbackSummary(messages), nil
	}
	return resp, nil
}

// buildSummaryPrompt constructs the prompt for LLM summarization
func (a *Agent) buildSummaryPrompt(messages []*session.Message) string {
	var sb strings.Builder
	sb.WriteString("请用简洁的自然语言总结以下会话，风格如同助手在回顾对话。")
	sb.WriteString("重点描述：涉及的集群、操作类型、资源类型、命名空间、是否有危险操作。\n\n")

	for _, m := range messages {
		role := "用户"
		if m.Role == session.RoleAssistant {
			role = "助手"
		}
		sb.WriteString(fmt.Sprintf("%s: %s\n", role, m.Content))
	}

	sb.WriteString("\n请生成一段简洁的对话式摘要（中文，100字以内）。")
	return sb.String()
}

// BuildContextMessagesWithSummary builds context messages with LLM summary if needed
func (a *Agent) BuildContextMessagesWithSummary(systemPrompt string, messages []*session.Message) ([]sharedutil.Message, error) {
	if a.ctxManager == nil {
		return nil, fmt.Errorf("context manager not available")
	}

	ctxMgr := a.ctxManager
	llmMessages, needsSummary, rawForSummary := ctxMgr.BuildContextMessages(systemPrompt, messages, "")

	if !needsSummary {
		return llmMessages, nil
	}

	// Generate LLM summary
	summary, err := a.generateLLMSummary(rawForSummary)
	if err != nil {
		return nil, err
	}

	// Rebuild with summary - create new messages with summary included
	result := []sharedutil.Message{
		{Role: "system", Content: systemPrompt},
		{Role: "system", Content: "[Previous conversation summary]: " + summary},
	}

	// Keep recent messages as "recent context"
	windowSize := ctxMgr.GetConfig().ToolCallRetention
	if windowSize <= 0 {
		windowSize = 5
	}
	startIdx := len(messages) - windowSize
	if startIdx < 0 {
		startIdx = 0
	}

	result = append(result, sharedutil.Message{
		Role:    "system",
		Content: fmt.Sprintf("[%d earlier messages have been summarized]", startIdx),
	})

	for i := startIdx; i < len(messages); i++ {
		result = append(result, sharedutil.Message{
			Role:    string(messages[i].Role),
			Content: messages[i].Content,
		})
	}

	return result, nil
}

// fallbackSummary uses the old lightweight keyword-based summary
func (a *Agent) fallbackSummary(messages []*session.Message) string {
	ctxMgr := a.ctxManager
	if ctxMgr == nil {
		return "会话摘要不可用"
	}
	return ctxMgr.GenerateFallbackSummary(messages)
}

// yamlMarshalIndent marshals v to YAML with proper indentation
func yamlMarshalIndent(v interface{}) ([]byte, error) {
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(v); err != nil {
		return nil, err
	}
	encoder.Close()
	return buf.Bytes(), nil
}

// State tracks the conversation state
type State struct {
	ClusterName     string
	SessionMessages []*session.Message
}

// NewState creates a new conversation state
func NewState(clusterName string, messages []*session.Message) *State {
	return &State{
		ClusterName:     clusterName,
		SessionMessages: messages,
	}
}
