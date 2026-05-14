package session

import (
	"fmt"

	sharedutil "github.com/threestoneliu/k8s-agent/pkg/shared"
)

// ShouldCompress checks if interactions need compression
func ShouldCompress(interactions []Interaction, retentionLimit int) bool {
	return len(interactions) > retentionLimit
}

// ReconstructInteraction rebuilds a compressed interaction as messages
func ReconstructInteraction(inter Interaction) []*Message {
	var messages []*Message
	messages = append(messages, &Message{
		Message: sharedutil.Message{
			Role:    sharedutil.RoleUser,
			Content: inter.Query,
		},
	})
	for _, toolName := range inter.ToolNames {
		messages = append(messages, &Message{
			Message: sharedutil.Message{
				Role:    sharedutil.RoleAssistant,
				Content: "[Tool: " + toolName + "]",
			},
		})
	}
	if inter.Summary != "" {
		messages = append(messages, &Message{
			Message: sharedutil.Message{
				Role:    sharedutil.RoleAssistant,
				Content: inter.Summary,
			},
		})
	}
	return messages
}

// AddPlaceholder adds a placeholder message indicating compression
func AddPlaceholder(messages []*Message, msgDropped, toolCallDropped int) []*Message {
	placeholder := &Message{
		Message: sharedutil.Message{
			Role:    sharedutil.RoleSystem,
			Content: fmt.Sprintf("[%d msgs + %d tool calls condensed]", msgDropped, toolCallDropped),
		},
	}
	return append(messages, placeholder)
}

// CompressInteractions applies interaction-based compression
func CompressInteractions(interactions []Interaction, retentionLimit int) ([]*Message, int) {
	if len(interactions) <= retentionLimit {
		// No compression needed, return original messages
		var result []*Message
		for _, inter := range interactions {
			result = append(result, inter.OriginalMessages...)
		}
		return result, 0
	}

	var result []*Message
	compressedCount := 0

	for i := range interactions {
		isRecent := i >= len(interactions)-retentionLimit
		inter := interactions[i]
		// Keep incomplete interactions intact (they can't be meaningfully compressed)
		if isRecent || !inter.Completed {
			// Keep original messages intact
			result = append(result, inter.OriginalMessages...)
		} else {
			// Compress: rebuild with simplified tool markers
			result = append(result, ReconstructInteraction(inter)...)
			compressedCount++
		}
	}

	return result, compressedCount
}