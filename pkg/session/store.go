package session

import (
	"errors"
	"sync"
)

// Store error variables
var (
	ErrConversationNotFound      = errors.New("conversation not found")
	ErrConversationAlreadyExists = errors.New("conversation already exists")
	ErrInvalidConversationID     = errors.New("invalid conversation ID")
)

// Store provides in-memory storage for conversations
type Store struct {
	mu      sync.RWMutex
	convMap map[string]*Conversation
}

// NewStore creates a new conversation store
func NewStore() *Store {
	return &Store{
		convMap: make(map[string]*Conversation),
	}
}

// CreateConversation creates a new conversation in the store
func (s *Store) CreateConversation(id, clusterName, namespace string) (*Conversation, error) {
	if id == "" {
		return nil, ErrInvalidConversationID
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.convMap[id]; exists {
		return nil, ErrConversationAlreadyExists
	}

	conv := NewConversation(id, clusterName, namespace)
	s.convMap[id] = conv
	return conv, nil
}

// GetConversation retrieves a conversation by ID
func (s *Store) GetConversation(id string) (*Conversation, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	conv, ok := s.convMap[id]
	if !ok {
		return nil, ErrConversationNotFound
	}
	return conv, nil
}

// DeleteConversation deletes a conversation by ID
func (s *Store) DeleteConversation(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.convMap[id]; !exists {
		return ErrConversationNotFound
	}

	delete(s.convMap, id)
	return nil
}

// ListConversations returns all conversations
func (s *Store) ListConversations() []*Conversation {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*Conversation, 0, len(s.convMap))
	for _, conv := range s.convMap {
		result = append(result, conv)
	}
	return result
}

// UpdateConversation updates a conversation using the provided update function
func (s *Store) UpdateConversation(id string, fn func(*Conversation) error) (*Conversation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	conv, ok := s.convMap[id]
	if !ok {
		return nil, ErrConversationNotFound
	}

	if err := fn(conv); err != nil {
		return nil, err
	}

	return conv, nil
}
