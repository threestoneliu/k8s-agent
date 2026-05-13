package session

import "strings"

// Interaction represents a user-LLM interaction
type Interaction struct {
	Query            string
	ToolNames        []string
	Summary          string
	Completed        bool
	OriginalMessages []*Message
}

// extractToolName extracts tool name from "[Function Call: k8s_delete]" format
func extractToolName(content string) string {
	// Check for "[Function Call:" first
	if idx := strings.Index(content, "[Function Call:"); idx != -1 {
		start := idx + len("[Function Call:")
		end := strings.Index(content[start:], "]")
		if end == -1 {
			return ""
		}
		name := strings.TrimSpace(content[start : start+end])
		// Remove parentheses if present: "k8s_delete(name='nginx')" -> "k8s_delete"
		if parenIdx := strings.Index(name, "("); parenIdx != -1 {
			name = name[:parenIdx]
		}
		return name
	}
	// Check for "[Tool Call:"
	if idx := strings.Index(content, "[Tool Call:"); idx != -1 {
		start := idx + len("[Tool Call:")
		end := strings.Index(content[start:], "]")
		if end == -1 {
			return ""
		}
		name := strings.TrimSpace(content[start : start+end])
		// Remove parentheses if present
		if parenIdx := strings.Index(name, "("); parenIdx != -1 {
			name = name[:parenIdx]
		}
		return name
	}
	return ""
}

// ParseToInteractions converts a slice of messages into a slice of interactions
func ParseToInteractions(messages []*Message) []Interaction {
	var interactions []Interaction
	var current *Interaction

	for i := range messages {
		msg := messages[i]
		if msg.Role == RoleUser {
			if current != nil {
				interactions = append(interactions, *current)
			}
			current = &Interaction{
				Query: msg.Content,
			}
			current.OriginalMessages = append(current.OriginalMessages, msg)
		} else if msg.Role == RoleAssistant {
			if isToolCallMessage(msg) {
				if current == nil {
					current = &Interaction{}
				}
				if toolName := extractToolName(msg.Content); toolName != "" {
					current.ToolNames = append(current.ToolNames, toolName)
				}
				current.OriginalMessages = append(current.OriginalMessages, msg)
			} else if isToolResultMessage(msg) {
				// Tool result (may appear as assistant message in some formats)
				if current == nil {
					current = &Interaction{}
				}
				current.Completed = true
				current.OriginalMessages = append(current.OriginalMessages, msg)
			} else if current != nil {
				// Final summary (assistant but not tool call or result)
				current.Summary = msg.Content
				current.OriginalMessages = append(current.OriginalMessages, msg)
			}
		} else if isToolResultMessage(msg) {
			if current == nil {
				current = &Interaction{}
			}
			current.Completed = true
			current.OriginalMessages = append(current.OriginalMessages, msg)
		}
	}

	if current != nil {
		interactions = append(interactions, *current)
	}

	return interactions
}