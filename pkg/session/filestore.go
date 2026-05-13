package session

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	sharedutil "k8s-agent/pkg/shared"
)

const (
	// DefaultMaxCacheSize is the default maximum number of sessions to keep in memory
	DefaultMaxCacheSize = 100
	// DefaultMaxSessions is the default maximum number of session files to retain
	DefaultMaxSessions = 10
)

// FileStore implements persistent storage using JSONL files with LRU cache
type FileStore struct {
	mu         sync.RWMutex
	storageDir string
	maxCache   int
	maxSessions int
	// cacheLRU tracks access order for LRU eviction
	cacheLRU   []string
	cache      map[string]*Conversation
}

// FileMessage represents a message stored in JSONL format
type FileMessage struct {
	Role        string            `json:"role"`
	Content     string            `json:"content"`
	MessageType string            `json:"message_type,omitempty"`
	Think       string            `json:"think,omitempty"`
	ToolCallID  string            `json:"tool_call_id,omitempty"`
	ToolCalls   []FileToolCall    `json:"tool_calls,omitempty"`
	Success     bool              `json:"success,omitempty"`
	Timestamp   string           `json:"timestamp"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// FileToolCall represents a tool call stored in JSONL format
type FileToolCall struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// FileConversation represents a conversation stored in JSONL format
type FileConversation struct {
	ID          string        `json:"id"`
	ClusterName string       `json:"cluster_name"`
	Namespace   string       `json:"namespace"`
	Messages    []FileMessage `json:"messages"`
	CreatedAt   string       `json:"created_at"`
	UpdatedAt   string       `json:"updated_at"`
}

// NewFileStore creates a new file-based store with LRU cache
func NewFileStore(storageDir string) (*FileStore, error) {
	return NewFileStoreWithConfig(storageDir, DefaultMaxCacheSize, DefaultMaxSessions)
}

// NewFileStoreWithMaxCache creates a new file-based store with specified max cache size
func NewFileStoreWithMaxCache(storageDir string, maxCache int) (*FileStore, error) {
	return NewFileStoreWithConfig(storageDir, maxCache, DefaultMaxSessions)
}

// NewFileStoreWithConfig creates a new file-based store with custom cache and session limits
func NewFileStoreWithConfig(storageDir string, maxCache, maxSessions int) (*FileStore, error) {
	// Expand tilde in path
	if strings.HasPrefix(storageDir, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		storageDir = filepath.Join(home, storageDir[1:])
	}

	// Ensure directory exists
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	store := &FileStore{
		storageDir:  storageDir,
		maxCache:    maxCache,
		maxSessions: maxSessions,
		cacheLRU:    make([]string, 0, maxCache),
		cache:       make(map[string]*Conversation, maxCache),
	}

	return store, nil
}

// filePath returns the path to a session file
func (s *FileStore) filePath(sessionID string) string {
	return filepath.Join(s.storageDir, sessionID+".jsonl")
}

// loadConversation loads a single conversation from file in JSONL format
func (s *FileStore) loadConversation(sessionID string) (*Conversation, error) {
	filePath := s.filePath(sessionID)
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	messages := make([]*Message, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var fm FileMessage
		if err := json.Unmarshal(scanner.Bytes(), &fm); err != nil {
			continue // Skip malformed lines
		}

		toolCalls := make([]sharedutil.ToolCall, 0, len(fm.ToolCalls))
		for _, tc := range fm.ToolCalls {
			toolCalls = append(toolCalls, sharedutil.ToolCall{
				ID:        tc.ID,
				Name:      tc.Name,
				Arguments: tc.Arguments,
			})
		}
		messages = append(messages, &Message{
			Message: sharedutil.Message{
				Role:       fm.Role,
				Content:    fm.Content,
				ToolCallID: fm.ToolCallID,
				ToolCalls:  toolCalls,
			},
			MessageType: fm.MessageType,
			Think:       fm.Think,
			Timestamp:   parseTimestamp(fm.Timestamp),
			Metadata:    fm.Metadata,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read conversation: %w", err)
	}

	return &Conversation{
		ID:       sessionID,
		Messages: messages,
	}, nil
}

// saveConversation saves a conversation to file in JSONL format
// Each message is saved as a separate line
func (s *FileStore) saveConversation(conv *Conversation) error {
	filePath := s.filePath(conv.ID)

	// Open file for writing
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	// Write each message as a separate JSON line
	for _, m := range conv.Messages {
		toolCalls := make([]FileToolCall, 0, len(m.ToolCalls))
		for _, tc := range m.ToolCalls {
			toolCalls = append(toolCalls, FileToolCall{
				ID:        tc.ID,
				Name:      tc.Name,
				Arguments: tc.Arguments,
			})
		}

		fileMsg := FileMessage{
			Role:        string(m.Message.Role),
			Content:     m.Message.Content,
			MessageType: m.MessageType,
			Think:       m.Think,
			ToolCallID:  m.Message.ToolCallID,
			ToolCalls:   toolCalls,
			Timestamp:   m.Timestamp.Format("2006-01-02T15:04:05.000Z"),
			Metadata:    m.Metadata,
		}

		data, err := json.Marshal(fileMsg)
		if err != nil {
			return fmt.Errorf("failed to marshal message: %w", err)
		}

		if _, err := file.Write(append(data, '\n')); err != nil {
			return fmt.Errorf("failed to write message: %w", err)
		}
	}

	return nil
}

// touchLRU moves session to end of LRU list (most recently used)
func (s *FileStore) touchLRU(sessionID string) {
	for i, id := range s.cacheLRU {
		if id == sessionID {
			// Move to end
			s.cacheLRU = append(s.cacheLRU[:i], s.cacheLRU[i+1:]...)
			s.cacheLRU = append(s.cacheLRU, sessionID)
			return
		}
	}
	// Not in LRU list, add to end
	s.cacheLRU = append(s.cacheLRU, sessionID)
}

// evictLRU removes the least recently used session from cache
func (s *FileStore) evictLRU() {
	if len(s.cacheLRU) == 0 {
		return
	}
	// Remove first (least recently used)
	evictID := s.cacheLRU[0]
	s.cacheLRU = s.cacheLRU[1:]
	delete(s.cache, evictID)
}

// ensureSpace ensures there's room in cache for a new session
func (s *FileStore) ensureSpace() {
	for len(s.cache) >= s.maxCache {
		s.evictLRU()
	}
}

// GetConversation retrieves a conversation by ID
func (s *FileStore) GetConversation(id string) (*Conversation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check cache first
	if conv, ok := s.cache[id]; ok {
		s.touchLRU(id)
		return conv, nil
	}

	// Load from file
	conv, err := s.loadConversation(id)
	if err != nil {
		return nil, ErrConversationNotFound
	}

	// Ensure space and add to cache
	s.ensureSpace()
	s.cache[id] = conv
	s.touchLRU(id)

	return conv, nil
}

// CreateConversation creates a new conversation in the store
func (s *FileStore) CreateConversation(id, clusterName, namespace string) (*Conversation, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if already exists
	if _, ok := s.cache[id]; ok {
		// Try to load from file
		if _, err := s.loadConversation(id); err == nil {
			return nil, ErrConversationAlreadyExists
		}
	}

	now := time.Now()
	conv := &Conversation{
		ID:          id,
		ClusterName: clusterName,
		Namespace:   namespace,
		Messages:    make([]*Message, 0),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.saveConversation(conv); err != nil {
		return nil, fmt.Errorf("failed to save conversation: %w", err)
	}

	// Ensure space and add to cache
	s.ensureSpace()
	s.cache[id] = conv
	s.touchLRU(id)

	// Cleanup old session files if exceeding limit
	s.cleanupOldSessions()

	return conv, nil
}

// UpdateConversation updates a conversation using the provided update function
func (s *FileStore) UpdateConversation(id string, fn func(*Conversation) error) (*Conversation, error) {
	// Fast path: check cache with read lock first
	s.mu.RLock()
	conv, ok := s.cache[id]
	s.mu.RUnlock()

	if !ok {
		// Slow path: load from file (outside of lock to avoid blocking)
		loaded, err := s.loadConversation(id)
		if err != nil {
			return nil, ErrConversationNotFound
		}

		// Now acquire write lock to update cache
		s.mu.Lock()
		// Double-check after acquiring write lock
		if existing, exists := s.cache[id]; exists {
			s.touchLRU(id)
			s.mu.Unlock()
			conv = existing
		} else {
			s.ensureSpace()
			s.cache[id] = loaded
			s.touchLRU(id)
			s.mu.Unlock()
			conv = loaded
		}
	}

	if err := fn(conv); err != nil {
		return nil, err
	}

	conv.UpdatedAt = time.Now()

	// Save to file (outside of lock to avoid blocking)
	if err := s.saveConversation(conv); err != nil {
		return nil, fmt.Errorf("failed to save conversation: %w", err)
	}

	s.mu.Lock()
	s.touchLRU(id)
	s.mu.Unlock()

	return conv, nil
}

// DeleteConversation deletes a conversation by ID
func (s *FileStore) DeleteConversation(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	filePath := s.filePath(id)
	os.Remove(filePath)

	delete(s.cache, id)
	// Remove from LRU list
	for i, sid := range s.cacheLRU {
		if sid == id {
			s.cacheLRU = append(s.cacheLRU[:i], s.cacheLRU[i+1:]...)
			break
		}
	}

	return nil
}

// cleanupOldSessions removes the oldest session files if exceeding maxSessions limit
func (s *FileStore) cleanupOldSessions() {
	if s.maxSessions <= 0 {
		return
	}

	entries, err := os.ReadDir(s.storageDir)
	if err != nil {
		return
	}

	// Count session files
	var sessionFiles []os.FileInfo
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonl") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		sessionFiles = append(sessionFiles, info)
	}

	// If within limit, no cleanup needed
	if len(sessionFiles) <= s.maxSessions {
		return
	}

	// Sort by modification time (oldest first)
	for i := 0; i < len(sessionFiles)-1; i++ {
		for j := i + 1; j < len(sessionFiles); j++ {
			if sessionFiles[i].ModTime().After(sessionFiles[j].ModTime()) {
				sessionFiles[i], sessionFiles[j] = sessionFiles[j], sessionFiles[i]
			}
		}
	}

	// Delete oldest files to meet the limit
	filesToDelete := len(sessionFiles) - s.maxSessions
	for i := 0; i < filesToDelete; i++ {
		sessionID := strings.TrimSuffix(sessionFiles[i].Name(), ".jsonl")
		filePath := s.filePath(sessionID)
		os.Remove(filePath)
	}
}

// ListConversations returns all conversations from cache
func (s *FileStore) ListConversations() []*Conversation {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return cached conversations only for performance
	// This avoids expensive file system traversal
	result := make([]*Conversation, 0, len(s.cache))
	for _, conv := range s.cache {
		result = append(result, conv)
	}
	return result
}

// parseTimestamp parses a timestamp string in RFC3339 format
func parseTimestamp(ts string) time.Time {
	t, err := time.Parse("2006-01-02T15:04:05.000Z", ts)
	if err != nil {
		return time.Now()
	}
	return t
}
