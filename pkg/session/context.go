package session

import (
	"fmt"
	"k8s-agent/pkg/cluster"
	sharedutil "k8s-agent/pkg/shared"
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

// isToolCallMessage checks if a message is a tool call (not a result)
func isToolCallMessage(msg *Message) bool {
	content := msg.Content
	return strings.Contains(content, "[Function Call:") ||
		strings.Contains(content, "[Tool Call:") ||
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
) []sharedutil.Message {
	result := []sharedutil.Message{
		{Role: "system", Content: systemPrompt},
	}

	// Level 0: Within limits, return all
	if len(messages) <= cm.config.MaxMessages {
		estimatedTokens := estimateMessagesTokens(messages)
		systemTokens := estimateTokens(systemPrompt)
		if systemTokens+estimatedTokens <= cm.config.MaxTokens {
			for _, m := range messages {
				result = append(result, sharedutil.Message{
					Role:    string(m.Role),
					Content: m.Content,
				})
			}
			return result
		}
	}

	// Level 1: Interaction-based compression
	// Parse messages to interactions and compress old ones
	interactions := ParseToInteractions(messages)
	retentionLimit := cm.config.ToolCallRetention
	if retentionLimit <= 0 {
		retentionLimit = DefaultToolCallRetention
	}
	droppedLevel1 := 0
	var compressed []*Message
	if ShouldCompress(interactions, retentionLimit) {
		compressed, droppedLevel1 = CompressInteractions(interactions, retentionLimit)
	}

	// Check if Level 1 is sufficient
	if len(compressed) <= cm.config.MaxMessages {
		estimatedTokens := estimateMessagesTokens(compressed)
		systemTokens := estimateTokens(systemPrompt)
		if systemTokens+estimatedTokens <= cm.config.MaxTokens {
			// Add placeholder if any interactions were compressed
			if droppedLevel1 > 0 {
				result = append(result, sharedutil.Message{
					Role:    "system",
					Content: fmt.Sprintf("[%d msgs + %d tool calls condensed]", droppedLevel1*3, droppedLevel1),
				})
			}
			for _, m := range compressed {
				result = append(result, sharedutil.Message{
					Role:    string(m.Role),
					Content: m.Content,
				})
			}
			return result
		}
	}

	// Level 2: Medium compression - keep only recent max-messages or max-tokens
	droppedLevel2 := 0
	compressed, dropped := cm.level2Compress(compressed)
	droppedLevel2 = dropped + droppedLevel1

	// Check if Level 2 is sufficient
	if len(compressed) <= cm.config.MaxMessages {
		estimatedTokens := estimateMessagesTokens(compressed)
		systemTokens := estimateTokens(systemPrompt)
		if systemTokens+estimatedTokens <= cm.config.MaxTokens {
			// Add placeholder if any messages were dropped
			if droppedLevel2 > 0 {
				result = append(result, sharedutil.Message{
					Role:    "system",
					Content: fmt.Sprintf("[%d earlier messages have been condensed due to context window limits]", droppedLevel2),
				})
			}
			for _, m := range compressed {
				result = append(result, sharedutil.Message{
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
		result = []sharedutil.Message{
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
			result = append(result, sharedutil.Message{
				Role:    "system",
				Content: fmt.Sprintf("[%d earlier messages have been summarized into the summary above]", droppedLevel3),
			})
		}

		for i := startIdx; i < len(messages); i++ {
			result = append(result, sharedutil.Message{
				Role:    string(messages[i].Role),
				Content: messages[i].Content,
			})
		}

		// Final token check
		result = cm.trimByTokenLimitWithPlaceholder(result)
	} else {
		// No summarization, just do final token trim
		if droppedLevel2 > 0 {
			result = append(result, sharedutil.Message{
				Role:    "system",
				Content: fmt.Sprintf("[%d earlier messages have been condensed due to context window limits]", droppedLevel2),
			})
		}
		for _, m := range compressed {
			result = append(result, sharedutil.Message{
				Role:    string(m.Role),
				Content: m.Content,
			})
		}
		result = cm.trimByTokenLimitWithPlaceholder(result)
	}

	return result
}

// level1Compress drops old tool call results but keeps recent N complete interactions.
// A complete interaction is: user query + tool call + tool result + final summary (assistant message that is not a tool call).
// For old interactions being dropped, we keep the user query and final summary, discard tool call and tool result.
// Returns compressed messages and count of dropped messages.
func (cm *ContextManager) level1Compress(messages []*Message) ([]*Message, int) {
	if len(messages) == 0 {
		return messages, 0
	}

	// Find all complete interactions (from back to front)
	// A complete interaction ends with an assistant message that is NOT a tool call
	interactions := cm.findCompleteInteractions(messages)
	if len(interactions) == 0 {
		return messages, 0
	}

	// Keep only the last N complete interactions
	keepCount := cm.config.ToolCallRetention
	if keepCount <= 0 {
		keepCount = DefaultToolCallRetention
	}

	// If number of interactions is within retention limit, no need to compress
	if len(interactions) <= keepCount {
		return messages, 0
	}

	// Build result by processing interactions
	var result []*Message
	droppedCount := 0

	// Process from oldest to newest
	for i := 0; i < len(interactions); i++ {
		interaction := interactions[i]
		isRecent := i >= len(interactions)-keepCount

		if isRecent {
			// Keep entire interaction intact
			for _, idx := range interaction.Indices {
				result = append(result, messages[idx])
			}
		} else {
			// For old interactions, keep only user query and final summary
			// Discard tool call and tool result
			for _, idx := range interaction.Indices {
				msg := messages[idx]
				if isUserMessage(msg) || isFinalSummaryMessage(msg, &interaction, idx) {
					result = append(result, msg)
				} else {
					droppedCount++
				}
			}
		}
	}

	// If result is still too long, let level2 handle it
	return result, droppedCount
}

// interaction represents a complete user-LLM interaction
type interaction struct {
	Indices     []int // indices into the messages array
	UserIndex   int  // index of user query
	SummaryIndex int // index of final summary (assistant message that is not a tool call)
}

// isUserMessage checks if a message is from user
func isUserMessage(msg *Message) bool {
	return msg.Role == RoleUser
}

// isFinalSummaryMessage checks if this message is the final summary of an interaction
// by checking both that the message is assistant role AND the index matches SummaryIndex
func isFinalSummaryMessage(msg *Message, interaction *interaction, idx int) bool {
	return msg.Role == RoleAssistant && interaction.SummaryIndex >= 0 && idx == interaction.SummaryIndex
}

// findCompleteInteractions identifies all complete interactions in the message list
// A complete interaction is: user query + tool call(s) + tool result(s) + final summary
// Returns interactions from oldest to newest
func (cm *ContextManager) findCompleteInteractions(messages []*Message) []interaction {
	var interactions []interaction
	var currentInteraction *interaction

	for i := 0; i < len(messages); i++ {
		msg := messages[i]

		if msg.Role == RoleUser {
			// Start new interaction
			if currentInteraction != nil && currentInteraction.UserIndex >= 0 {
				interactions = append(interactions, *currentInteraction)
			}
			currentInteraction = &interaction{
				UserIndex:   i,
				SummaryIndex: -1,
			}
			currentInteraction.Indices = append(currentInteraction.Indices, i)
		} else if msg.Role == RoleAssistant {
			if isToolCallMessage(msg) {
				// Tool call - add to current interaction
				if currentInteraction == nil {
					// Edge case: tool call without user message before it, start new interaction
					currentInteraction = &interaction{
						UserIndex:   -1,
						SummaryIndex: -1,
					}
				}
				currentInteraction.Indices = append(currentInteraction.Indices, i)
			} else {
				// Final summary (assistant but not tool call)
				if currentInteraction == nil {
					// Edge case: summary without prior interaction, treat as its own interaction
					currentInteraction = &interaction{
						UserIndex:    -1,
						SummaryIndex: i,
					}
				} else {
					currentInteraction.SummaryIndex = i
				}
				currentInteraction.Indices = append(currentInteraction.Indices, i)
			}
		} else if isToolResultMessage(msg) {
			// Tool result - add to current interaction
			if currentInteraction == nil {
				currentInteraction = &interaction{
					UserIndex:   -1,
					SummaryIndex: -1,
				}
			}
			currentInteraction.Indices = append(currentInteraction.Indices, i)
		}
	}

	// Don't forget the last interaction
	if currentInteraction != nil && currentInteraction.UserIndex >= 0 {
		interactions = append(interactions, *currentInteraction)
	}

	return interactions
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
func (cm *ContextManager) trimByTokenLimitWithPlaceholder(messages []sharedutil.Message) []sharedutil.Message {
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

	var trimmed []sharedutil.Message
	trimmed = append(trimmed, messages[0])
	droppedCount := 0

	for i := len(messages) - 1; i >= 1 && remaining > 0; i-- {
		msgTokens := estimateTokens(messages[i].Content) + 4
		if msgTokens <= remaining {
			trimmed = append([]sharedutil.Message{messages[i]}, trimmed...)
			remaining -= msgTokens
		} else {
			droppedCount++
		}
	}

	// Add placeholder if any messages were dropped
	if droppedCount > 0 {
		placeholder := sharedutil.Message{
			Role:    "system",
			Content: fmt.Sprintf("[%d earlier messages have been removed due to context window limits]", droppedCount),
		}
		// Insert after system prompt(s)
		if len(trimmed) > 1 {
			trimmed = append([]sharedutil.Message{trimmed[0], placeholder}, trimmed[1:]...)
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

// CompressMessages compresses LLM messages to fit within context limits
func (cm *ContextManager) CompressMessages(messages []sharedutil.Message) []sharedutil.Message {
	if len(messages) <= cm.config.MaxMessages {
		return messages
	}

	// Level 1: Light compression
	compressed, _ := cm.level1CompressLLM(messages)
	if len(compressed) <= cm.config.MaxMessages {
		return compressed
	}

	// Level 2: Medium compression - keep only recent messages
	windowSize := cm.config.MaxMessages
	if windowSize <= 0 {
		windowSize = 20
	}
	if len(compressed) > windowSize {
		compressed = compressed[len(compressed)-windowSize:]
	}

	return compressed
}

// llmInteraction represents a complete user-LLM interaction in LLM message format
type llmInteraction struct {
	Indices      []int // indices into the messages array
	UserIndex    int  // index of user query (-1 if none)
	SummaryIndex int  // index of final summary (-1 if none)
}

// isLLMToolResultMessage checks if a message is a tool result in LLM format
func isLLMToolResultMessage(m sharedutil.Message) bool {
	return m.Role == "tool"
}

// isLLMToolCallMessage checks if a message is a tool call in LLM format
func isLLMToolCallMessage(m sharedutil.Message) bool {
	return m.Role == "assistant" && len(m.ToolCalls) > 0
}

// isLLMFinalSummaryMessage checks if this is a final summary (assistant message without tool calls)
func isLLMFinalSummaryMessage(m sharedutil.Message) bool {
	return m.Role == "assistant" && len(m.ToolCalls) == 0
}

// findLLMCompleteInteractions identifies complete interactions in LLM message format
// A complete interaction is: user + tool call + tool result + final summary
// Returns interactions from oldest to newest
func (cm *ContextManager) findLLMCompleteInteractions(messages []sharedutil.Message) []llmInteraction {
	var interactions []llmInteraction
	var current *llmInteraction

	for i := range messages {
		m := messages[i]

		if m.Role == "user" {
			// Start new interaction
			if current != nil && current.UserIndex >= 0 {
				interactions = append(interactions, *current)
			}
			current = &llmInteraction{
				UserIndex:    i,
				SummaryIndex: -1,
			}
			current.Indices = append(current.Indices, i)
		} else if isLLMToolCallMessage(m) {
			// Tool call - add to current interaction
			if current == nil {
				current = &llmInteraction{
					UserIndex:    -1,
					SummaryIndex: -1,
				}
			}
			current.Indices = append(current.Indices, i)
		} else if isLLMToolResultMessage(m) {
			// Tool result - add to current interaction
			if current == nil {
				current = &llmInteraction{
					UserIndex:    -1,
					SummaryIndex: -1,
				}
			}
			current.Indices = append(current.Indices, i)
		} else if isLLMFinalSummaryMessage(m) {
			// Final summary
			if current == nil {
				current = &llmInteraction{
					UserIndex:    -1,
					SummaryIndex: -1,
				}
			} else {
				current.SummaryIndex = i
			}
			current.Indices = append(current.Indices, i)
		}
	}

	if current != nil && current.UserIndex >= 0 {
		interactions = append(interactions, *current)
	}

	return interactions
}

// level1CompressLLM compresses LLM message slice by keeping recent N complete interactions
// For old interactions, keeps user query and final summary, discards tool call and tool result
func (cm *ContextManager) level1CompressLLM(messages []sharedutil.Message) ([]sharedutil.Message, int) {
	interactions := cm.findLLMCompleteInteractions(messages)
	if len(interactions) == 0 {
		return messages, 0
	}

	keepCount := cm.config.ToolCallRetention
	if keepCount <= 0 {
		keepCount = DefaultToolCallRetention
	}

	// If interactions within retention limit, no compression needed
	if len(interactions) <= keepCount {
		return messages, 0
	}

	var result []sharedutil.Message
	droppedCount := 0

	for i := range interactions {
		interaction := interactions[i]
		isRecent := i >= len(interactions)-keepCount

		if isRecent {
			// Keep entire interaction intact
			for _, idx := range interaction.Indices {
				result = append(result, messages[idx])
			}
		} else {
			// For old interactions, keep only user query and final summary
			for _, idx := range interaction.Indices {
				m := messages[idx]
				// Keep user messages
				if m.Role == "user" {
					result = append(result, m)
				} else if interaction.SummaryIndex != -1 && idx == interaction.SummaryIndex {
					// Keep final summary
					result = append(result, m)
				} else {
					// Discard tool call and tool result
					droppedCount++
				}
			}
		}
	}

	return result, droppedCount
}