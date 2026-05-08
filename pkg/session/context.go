package session

import (
	"fmt"
	"k8s-agent/pkg/cluster"
	"k8s-agent/pkg/llm"
	"strings"
)

const (
	// DefaultToolCallRetention is the default number of recent tool calls to retain
	DefaultToolCallRetention = 10
)

// ContextManager manages conversation context with three-level compression strategy
type ContextManager struct {
	config cluster.ContextConfig
}

// NewContextManager creates a new context manager with the given config
func NewContextManager(config cluster.ContextConfig) *ContextManager {
	return &ContextManager{config: config}
}

// isToolCallMessage checks if a message is a tool call related message
func isToolCallMessage(msg *Message) bool {
	// Tool messages contain "[Function Call:" or "[Tool:"
	content := msg.Content
	return strings.Contains(content, "[Function Call:") ||
		strings.Contains(content, "[Tool:") ||
		isToolResultMessage(msg) ||
		strings.Contains(msg.Content, "[Tool:") ||
		strings.Contains(msg.Content, "ToolCallID")
}

// isToolResultMessage checks if a message is a tool execution result
func isToolResultMessage(msg *Message) bool {
	return strings.Contains(msg.Content, "[Tool:") ||
		strings.Contains(msg.Content, "Result:") ||
		strings.Contains(msg.Content, "Error:") ||
		strings.Contains(msg.Content, "Confirmation required")
}

// estimateTokens estimates the token count for a string
func estimateTokens(s string) int {
	if s == "" {
		return 0
	}
	runes := []rune(s)

	cjkCount := 0
	for _, r := range runes {
		if r >= 0x4E00 && r <= 0x9FFF {
			cjkCount++
		}
	}

	nonCjk := len(runes) - cjkCount
	return (nonCjk / 4) + cjkCount
}

// estimateMessagesTokens estimates total tokens for a slice of messages
func estimateMessagesTokens(messages []*Message) int {
	total := 0
	for _, m := range messages {
		total += 4 + estimateTokens(m.Content)
	}
	return total
}

// BuildContextMessages builds the message list for LLM with three-level context compression
func (cm *ContextManager) BuildContextMessages(
	systemPrompt string,
	messages []*Message,
	summaryPrompt string,
) []llm.Message {
	result := []llm.Message{
		{Role: "system", Content: systemPrompt},
	}

	// Level 0: Within limits, return all
	if len(messages) <= cm.config.MaxMessages {
		estimatedTokens := estimateMessagesTokens(messages)
		systemTokens := estimateTokens(systemPrompt)
		if systemTokens+estimatedTokens <= cm.config.MaxTokens {
			for _, m := range messages {
				result = append(result, llm.Message{
					Role:    string(m.Role),
					Content: m.Content,
				})
			}
			return result
		}
	}

	// Level 1: Light compression - drop old tool call results, keep recent 10
	droppedLevel1 := 0
	compressed, dropped := cm.level1Compress(messages)
	droppedLevel1 = dropped

	// Check if Level 1 is sufficient
	if len(compressed) <= cm.config.MaxMessages {
		estimatedTokens := estimateMessagesTokens(compressed)
		systemTokens := estimateTokens(systemPrompt)
		if systemTokens+estimatedTokens <= cm.config.MaxTokens {
			// Add placeholder if any messages were dropped
			if droppedLevel1 > 0 {
				result = append(result, llm.Message{
					Role:    "system",
					Content: fmt.Sprintf("[%d earlier messages including %d tool calls have been condensed]", droppedLevel1, droppedLevel1/2),
				})
			}
			for _, m := range compressed {
				result = append(result, llm.Message{
					Role:    string(m.Role),
					Content: m.Content,
				})
			}
			return result
		}
	}

	// Level 2: Medium compression - keep only recent max-messages or max-tokens
	droppedLevel2 := 0
	compressed, dropped = cm.level2Compress(compressed)
	droppedLevel2 = dropped + droppedLevel1

	// Check if Level 2 is sufficient
	if len(compressed) <= cm.config.MaxMessages {
		estimatedTokens := estimateMessagesTokens(compressed)
		systemTokens := estimateTokens(systemPrompt)
		if systemTokens+estimatedTokens <= cm.config.MaxTokens {
			// Add placeholder if any messages were dropped
			if droppedLevel2 > 0 {
				result = append(result, llm.Message{
					Role:    "system",
					Content: fmt.Sprintf("[%d earlier messages have been condensed due to context window limits]", droppedLevel2),
				})
			}
			for _, m := range compressed {
				result = append(result, llm.Message{
					Role:    string(m.Role),
					Content: m.Content,
				})
			}
			return result
		}
	}

	// Level 3: Deep compression - generate summary if enabled
	if cm.config.SummaryEnabled {
		summary := cm.generateSummary(messages, summaryPrompt)
		result = []llm.Message{
			{Role: "system", Content: systemPrompt},
			{Role: "system", Content: "[Previous conversation summary]: " + summary},
		}

		// Keep most recent messages that fit
		windowSize := cm.config.MaxMessages - 1 // Account for summary
		if windowSize < 1 {
			windowSize = 1
		}
		startIdx := len(messages) - windowSize
		if startIdx < 0 {
			startIdx = 0
		}

		droppedLevel3 := startIdx
		if droppedLevel3 > 0 {
			result = append(result, llm.Message{
				Role:    "system",
				Content: fmt.Sprintf("[%d earlier messages have been summarized into the summary above]", droppedLevel3),
			})
		}

		for i := startIdx; i < len(messages); i++ {
			result = append(result, llm.Message{
				Role:    string(messages[i].Role),
				Content: messages[i].Content,
			})
		}

		// Final token check
		result = cm.trimByTokenLimitWithPlaceholder(result)
	} else {
		// No summarization, just do final token trim
		if droppedLevel2 > 0 {
			result = append(result, llm.Message{
				Role:    "system",
				Content: fmt.Sprintf("[%d earlier messages have been condensed due to context window limits]", droppedLevel2),
			})
		}
		for _, m := range compressed {
			result = append(result, llm.Message{
				Role:    string(m.Role),
				Content: m.Content,
			})
		}
		result = cm.trimByTokenLimitWithPlaceholder(result)
	}

	return result
}

// level1Compress drops old tool call results but keeps recent 10 tool calls
// Returns compressed messages and count of dropped messages
func (cm *ContextManager) level1Compress(messages []*Message) ([]*Message, int) {
	if len(messages) <= cm.config.MaxMessages {
		return messages, 0
	}

	// Identify all tool call related messages
	type toolCallInfo struct {
		index int
		msg   *Message
	}
	var toolCalls []toolCallInfo

	for i, msg := range messages {
		if isToolCallMessage(msg) || isToolResultMessage(msg) {
			toolCalls = append(toolCalls, toolCallInfo{index: i, msg: msg})
		}
	}

	// If no tool calls, return original
	if len(toolCalls) == 0 {
		return messages, 0
	}

	// Keep only the last N tool calls and their results
	keepCount := cm.config.ToolCallRetention
	if keepCount <= 0 {
		keepCount = DefaultToolCallRetention
	}

	// Find the last N tool call indices
	keepStartIdx := 0
	if len(toolCalls) > keepCount {
		keepStartIdx = len(toolCalls) - keepCount
	}
	keepIndices := make(map[int]bool)
	for i := keepStartIdx; i < len(toolCalls); i++ {
		keepIndices[toolCalls[i].index] = true
	}

	// Also keep tool result that follows a kept tool call
	var result []*Message
	droppedCount := 0
	for i, msg := range messages {
		// Keep all non-tool messages
		if !isToolCallMessage(msg) && !isToolResultMessage(msg) {
			result = append(result, msg)
			continue
		}

		// Keep tool messages that are in the keep list
		if keepIndices[i] {
			result = append(result, msg)
			continue
		}

		// For tool results, check if the previous message is kept
		if isToolResultMessage(msg) && i > 0 && keepIndices[i-1] {
			result = append(result, msg)
			continue
		}

		// Otherwise skip (compress) and count
		droppedCount++
	}

	return result, droppedCount
}

// level2Compress keeps only the most recent max-messages or fits within max-tokens
// Returns compressed messages and count of dropped messages
func (cm *ContextManager) level2Compress(messages []*Message) ([]*Message, int) {
	windowSize := cm.config.MaxMessages
	if windowSize <= 0 {
		windowSize = 20
	}

	if len(messages) <= windowSize {
		if cm.config.MaxTokens > 0 {
			totalTokens := estimateMessagesTokens(messages)
			if totalTokens <= cm.config.MaxTokens {
				return messages, 0
			}
		}
		return messages, 0
	}

	// Calculate dropped count
	droppedCount := len(messages) - windowSize

	// Take most recent messages
	startIdx := len(messages) - windowSize
	result := messages[startIdx:]

	// If still over token limit, trim from oldest
	if cm.config.MaxTokens > 0 {
		trimmed, additionalDropped := cm.trimMessagesByTokenLimit(result)
		result = trimmed
		droppedCount += additionalDropped
	}

	return result, droppedCount
}

// trimMessagesByTokenLimit trims messages from oldest to fit token limit
// Returns trimmed messages and count of dropped messages
func (cm *ContextManager) trimMessagesByTokenLimit(messages []*Message) ([]*Message, int) {
	if cm.config.MaxTokens <= 0 || len(messages) == 0 {
		return messages, 0
	}

	// Reserve tokens for system prompt
	remaining := cm.config.MaxTokens - 50
	if remaining <= 0 {
		return messages[len(messages)-1:], len(messages) - 1
	}

	var trimmed []*Message
	droppedCount := 0

	// Work backwards, keep newest messages
	for i := len(messages) - 1; i >= 0 && remaining > 0; i-- {
		msgTokens := estimateTokens(messages[i].Content) + 4
		if msgTokens <= remaining {
			trimmed = append([]*Message{messages[i]}, trimmed...)
			remaining -= msgTokens
		} else {
			droppedCount++
		}
	}

	if len(trimmed) == 0 {
		return messages[len(messages)-1:], droppedCount + len(messages) - 1
	}

	return trimmed, droppedCount
}

// trimByTokenLimitWithPlaceholder trims messages to fit within token limit and adds a placeholder
func (cm *ContextManager) trimByTokenLimitWithPlaceholder(messages []llm.Message) []llm.Message {
	if cm.config.MaxTokens <= 0 || len(messages) == 0 {
		return messages
	}

	systemPromptTokens := estimateTokens(messages[0].Content)
	if systemPromptTokens > cm.config.MaxTokens {
		messages[0].Content = truncateToTokens(messages[0].Content, cm.config.MaxTokens)
		return messages
	}

	remaining := cm.config.MaxTokens - systemPromptTokens - 50
	if remaining <= 0 {
		return messages[:1]
	}

	var trimmed []llm.Message
	trimmed = append(trimmed, messages[0])
	droppedCount := 0

	for i := len(messages) - 1; i >= 1 && remaining > 0; i-- {
		msgTokens := estimateTokens(messages[i].Content) + 4
		if msgTokens <= remaining {
			trimmed = append([]llm.Message{messages[i]}, trimmed...)
			remaining -= msgTokens
		} else {
			droppedCount++
		}
	}

	// Add placeholder if any messages were dropped
	if droppedCount > 0 {
		placeholder := llm.Message{
			Role:    "system",
			Content: fmt.Sprintf("[%d earlier messages have been removed due to context window limits]", droppedCount),
		}
		// Insert after system prompt(s)
		if len(trimmed) > 1 {
			trimmed = append([]llm.Message{trimmed[0], placeholder}, trimmed[1:]...)
		} else {
			trimmed = append(trimmed, placeholder)
		}
	}

	if len(trimmed) == 1 {
		return messages[:1]
	}

	return trimmed
}

// truncateToTokens truncates string to fit within token limit
func truncateToTokens(s string, maxTokens int) string {
	runes := []rune(s)
	charLimit := maxTokens * 4
	if len(runes) <= charLimit {
		return s
	}
	return string(runes[:charLimit]) + "..."
}

// generateSummary creates a summary of the conversation using LLM
func (cm *ContextManager) generateSummary(messages []*Message, summaryPrompt string) string {
	if len(messages) == 0 {
		return "Empty conversation"
	}

	var sb strings.Builder
	sb.WriteString("Conversation summary:\n")

	// Analyze topics
	topics := make(map[string]int)
	for _, m := range messages {
		content := strings.ToLower(m.Content)
		if strings.Contains(content, "pod") {
			topics["pods"]++
		}
		if strings.Contains(content, "service") || strings.Contains(content, "svc") {
			topics["services"]++
		}
		if strings.Contains(content, "deployment") || strings.Contains(content, "deploy") {
			topics["deployments"]++
		}
		if strings.Contains(content, "delete") {
			topics["delete operations"]++
		}
		if strings.Contains(content, "create") {
			topics["create operations"]++
		}
		if strings.Contains(content, "scale") {
			topics["scale operations"]++
		}
		if strings.Contains(content, "get") || strings.Contains(content, "list") || strings.Contains(content, "describe") {
			topics["query operations"]++
		}
		if strings.Contains(content, "configmap") || strings.Contains(content, "secret") {
			topics["config/secrets"]++
		}
		if strings.Contains(content, "node") {
			topics["node operations"]++
		}
		if strings.Contains(content, "namespace") || strings.Contains(content, "ns") {
			topics["namespace operations"]++
		}
	}

	if len(topics) > 0 {
		sb.WriteString("Topics: ")
		for topic := range topics {
			sb.WriteString(topic)
			sb.WriteString(", ")
		}
		sb.WriteString("\n")
	}

	// Recent user requests
	sb.WriteString("Recent user requests:\n")
	userCount := 0
	for i := len(messages) - 1; i >= 0 && userCount < 5; i-- {
		if messages[i].Role == RoleUser {
			content := messages[i].Content
			if len(content) > 150 {
				content = content[:150] + "..."
			}
			sb.WriteString("- ")
			sb.WriteString(content)
			sb.WriteString("\n")
			userCount++
		}
	}

	// Operations performed
	sb.WriteString("Operations performed:\n")
	opCount := 0
	for i := len(messages) - 1; i >= 0 && opCount < 10; i-- {
		if isToolCallMessage(messages[i]) {
			content := messages[i].Content
			// Extract just the function name
			if idx := strings.Index(content, "[Function Call:"); idx >= 0 {
				start := idx + len("[Function Call:")
				end := strings.Index(content[start:], "]")
				if end > 0 {
					fnName := strings.TrimSpace(content[start : start+end])
					sb.WriteString("- ")
					sb.WriteString(fnName)
					sb.WriteString("\n")
					opCount++
				}
			}
		}
	}

	return sb.String()
}

// MessageCount returns the configured max messages
func (cm *ContextManager) MessageCount() int {
	return cm.config.MaxMessages
}

// IsSummaryEnabled returns whether summarization is enabled
func (cm *ContextManager) IsSummaryEnabled() bool {
	return cm.config.SummaryEnabled
}