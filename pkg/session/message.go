package session

import (
	"errors"
	"time"
)

// Role represents the role of a message sender
type Role string

// Message roles
const (
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleSystem    Role = "system"
)

// Message represents a single message in a conversation
type Message struct {
	Role      Role              `json:"role"`
	Content   string            `json:"content"`
	Think     string            `json:"think,omitempty"`   // LLM thinking content
	ToolCalls []ToolCall        `json:"tool_calls,omitempty"` // Tool/Function calls
	Timestamp time.Time         `json:"timestamp"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// ToolCall represents a tool/function call made during conversation
type ToolCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// NewMessage creates a new message with the given role and content
func NewMessage(role Role, content string, metadata map[string]string) *Message {
	if content == "" {
		panic(errors.New("message content cannot be empty"))
	}

	return &Message{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
		Metadata:  metadata,
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
func (c *Conversation) GetMessagesByRole(role Role) []*Message {
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
