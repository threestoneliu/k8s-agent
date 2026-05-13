package session

import (
	"errors"
	"time"

	sharedutil "k8s-agent/pkg/shared"
)

// Role constants (OpenAI standard - using shared constants)
const (
	RoleUser      = sharedutil.RoleUser
	RoleAssistant = sharedutil.RoleAssistant
	RoleSystem    = sharedutil.RoleSystem
)

// MessageType constants for UI rendering
const (
	MessageTypeText        = "text"
	MessageTypeThink       = "think"
	MessageTypeUser        = "user"
	MessageTypeToolCall    = "tool_call"
	MessageTypeToolResult  = "tool_result"
)

// Message represents a single message in a conversation
// Embeds shared.Message for Role, Content, ToolCalls, ToolCallID
type Message struct {
	sharedutil.Message
	// UI fields
	MessageType string            `json:"message_type"`
	Think       string            `json:"think,omitempty"`
	Timestamp   time.Time        `json:"timestamp"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// NewMessage creates a new message with the given role and content
func NewMessage(role string, content string, metadata map[string]string) *Message {
	if content == "" {
		panic(errors.New("message content cannot be empty"))
	}

	return &Message{
		Message: sharedutil.Message{
			Role:    role,
			Content: content,
		},
		Timestamp: time.Now(),
		Metadata: metadata,
	}
}

// Conversation represents a conversation session
type Conversation struct {
	ID          string     `json:"id"`
	ClusterName string     `json:"cluster_name"`
	Namespace   string     `json:"namespace"`
	Messages    []*Message `json:"messages"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// NewConversation creates a new conversation with the given ID
func NewConversation(id, clusterName, namespace string) *Conversation {
	now := time.Now()
	return &Conversation{
		ID:          id,
		ClusterName: clusterName,
		Namespace:   namespace,
		Messages:    make([]*Message, 0),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// AddMessage adds a message to the conversation
func (c *Conversation) AddMessage(msg *Message) {
	c.Messages = append(c.Messages, msg)
	c.UpdatedAt = time.Now()
}

// GetLastMessage returns the last message in the conversation
func (c *Conversation) GetLastMessage() *Message {
	if len(c.Messages) == 0 {
		return nil
	}
	return c.Messages[len(c.Messages)-1]
}

// GetMessagesByRole returns all messages with the given role
func (c *Conversation) GetMessagesByRole(role string) []*Message {
	result := make([]*Message, 0, len(c.Messages))
	for _, msg := range c.Messages {
		if msg.Role == role {
			result = append(result, msg)
		}
	}
	return result
}

// GetClusterContext returns the current cluster context
func (c *Conversation) GetClusterContext() string {
	return c.ClusterName
}

// SetClusterContext sets the current cluster context
func (c *Conversation) SetClusterContext(clusterName string) {
	c.ClusterName = clusterName
	c.UpdatedAt = time.Now()
}

// GetNamespace returns the current namespace context
func (c *Conversation) GetNamespace() string {
	return c.Namespace
}

// SetNamespaceContext sets the current namespace context
func (c *Conversation) SetNamespaceContext(namespace string) {
	c.Namespace = namespace
	c.UpdatedAt = time.Now()
}

// MessageCount returns the number of messages in the conversation
func (c *Conversation) MessageCount() int {
	return len(c.Messages)
}
