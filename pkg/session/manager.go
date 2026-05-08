package session

import "errors"

// Manager error variables
var (
	ErrMessageEmpty = errors.New("message content cannot be empty")
)

// StoreInterface defines the interface for conversation storage
type StoreInterface interface {
	CreateConversation(id, clusterName, namespace string) (*Conversation, error)
	GetConversation(id string) (*Conversation, error)
	UpdateConversation(id string, fn func(*Conversation) error) (*Conversation, error)
	DeleteConversation(id string) error
	ListConversations() []*Conversation
}

// Manager provides conversation management operations
type Manager struct {
	store StoreInterface
}

// NewManager creates a new conversation manager with in-memory store
func NewManager() *Manager {
	return &Manager{
		store: NewStore(),
	}
}

// NewManagerWithFileStore creates a new conversation manager with file-based store
func NewManagerWithFileStore(storagePath string) (*Manager, error) {
	return NewManagerWithFileStoreAndCache(storagePath, 100)
}

// NewManagerWithFileStoreAndCache creates a new conversation manager with file-based store and custom cache size
func NewManagerWithFileStoreAndCache(storagePath string, maxCache int) (*Manager, error) {
	store, err := NewFileStoreWithMaxCache(storagePath, maxCache)
	if err != nil {
		return nil, err
	}
	return &Manager{store: store}, nil
}

// NewManagerWithFileStoreAndLimits creates a new conversation manager with file-based store and custom limits
func NewManagerWithFileStoreAndLimits(storagePath string, maxCache, maxSessions int) (*Manager, error) {
	store, err := NewFileStoreWithConfig(storagePath, maxCache, maxSessions)
	if err != nil {
		return nil, err
	}
	return &Manager{store: store}, nil
}

// GetStore returns the underlying store
func (m *Manager) GetStore() StoreInterface {
	return m.store
}

// CreateConversation creates a new conversation session
func (m *Manager) CreateConversation(id, clusterName, namespace string) (*Conversation, error) {
	return m.store.CreateConversation(id, clusterName, namespace)
}

// GetConversation retrieves a conversation by ID
func (m *Manager) GetConversation(id string) (*Conversation, error) {
	return m.store.GetConversation(id)
}

// AddMessage adds a message to a conversation
func (m *Manager) AddMessage(conversationID string, role Role, content string, metadata map[string]string) error {
	if content == "" {
		return ErrMessageEmpty
	}

	_, err := m.store.UpdateConversation(conversationID, func(c *Conversation) error {
		msg := NewMessage(role, content, metadata)
		c.AddMessage(msg)
		return nil
	})
	return err
}

// SetClusterContext sets the cluster context for a conversation
func (m *Manager) SetClusterContext(conversationID, clusterName string) error {
	_, err := m.store.UpdateConversation(conversationID, func(c *Conversation) error {
		c.SetClusterContext(clusterName)
		return nil
	})
	return err
}

// SetNamespaceContext sets the namespace context for a conversation
func (m *Manager) SetNamespaceContext(conversationID, namespace string) error {
	_, err := m.store.UpdateConversation(conversationID, func(c *Conversation) error {
		c.SetNamespaceContext(namespace)
		return nil
	})
	return err
}

// ListConversations lists all conversations
func (m *Manager) ListConversations() []*Conversation {
	return m.store.ListConversations()
}

// DeleteConversation deletes a conversation
func (m *Manager) DeleteConversation(id string) error {
	return m.store.DeleteConversation(id)
}
