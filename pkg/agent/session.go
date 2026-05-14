package agent

import (
	"fmt"
	"time"

	"github.com/threestoneliu/k8s-agent/pkg/session"
	sharedutil "github.com/threestoneliu/k8s-agent/pkg/shared"
)

// NewMessageWithType creates a session message with explicit MessageType
func NewMessageWithType(role, content string, msgType string, metadata map[string]string) *session.Message {
	return &session.Message{
		Message: sharedutil.Message{
			Role:    role,
			Content: content,
		},
		MessageType: msgType,
		Timestamp:   time.Now(),
		Metadata:    metadata,
	}
}

// NewToolCallMessage creates a session message for a tool call
func NewToolCallMessage(fnCall *sharedutil.FunctionCall, metadata map[string]string) *session.Message {
	return &session.Message{
		Message: sharedutil.Message{
			Role:       sharedutil.RoleAssistant,
			Content:    fmt.Sprintf("执行工具: %s(%s)", fnCall.Name, fnCall.Arguments),
			ToolCallID: fnCall.ID,
		},
		MessageType: session.MessageTypeToolCall,
		Timestamp:   time.Now(),
		Metadata:    metadata,
	}
}

// NewToolResultMessage creates a session message for a tool result
func NewToolResultMessage(success bool, result, errorMsg string, toolCallID string) *session.Message {
	content := "工具执行成功"
	if !success {
		content = fmt.Sprintf("工具执行失败: %s", errorMsg)
	}
	return &session.Message{
		Message: sharedutil.Message{
			Role:       sharedutil.RoleAssistant,
			Content:    content,
			ToolCallID: toolCallID,
		},
		MessageType: session.MessageTypeToolResult,
		Timestamp:   time.Now(),
	}
}