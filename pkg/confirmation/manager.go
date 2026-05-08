package confirmation

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"k8s-agent/pkg/engine"
)

var (
	ErrNotFound       = errors.New("confirmation not found")
	ErrAlreadyHandled = errors.New("confirmation already approved or rejected")
	ErrExpired        = errors.New("confirmation has expired")
)

// Manager manages pending confirmation requests
type Manager struct {
	pending map[string]*PendingOperation
	ttl     time.Duration
	mu      sync.RWMutex
}

// NewManager creates a new confirmation manager with the specified TTL
func NewManager(ttl time.Duration) *Manager {
	return &Manager{
		pending: make(map[string]*PendingOperation),
		ttl:     ttl,
	}
}

// CreateConfirmation creates a new pending confirmation for an operation
func (m *Manager) CreateConfirmation(clusterName string, op *engine.ClassifiedOperation) (confirmKey string, err error) {
	pending := NewPendingOperation(clusterName, op)
	pending.ExpiresAt = time.Now().Add(m.ttl)

	m.mu.Lock()
	defer m.mu.Unlock()

	// Ensure unique confirm key
	for m.pending[pending.ConfirmKey] != nil {
		pending.ConfirmKey, _ = GenerateSecureConfirmKey()
	}

	m.pending[pending.ConfirmKey] = pending
	return pending.ConfirmKey, nil
}

// ApproveConfirmation approves a pending confirmation
func (m *Manager) ApproveConfirmation(confirmKey string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	pending, ok := m.pending[confirmKey]
	if !ok {
		return ErrNotFound
	}

	// Check if already handled
	if pending.Status != StatusPending {
		return ErrAlreadyHandled
	}

	// Check if expired
	if pending.IsExpired() {
		pending.Status = StatusExpired
		return ErrExpired
	}

	pending.Status = StatusApproved
	return nil
}

// RejectConfirmation rejects a pending confirmation
func (m *Manager) RejectConfirmation(confirmKey string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	pending, ok := m.pending[confirmKey]
	if !ok {
		return ErrNotFound
	}

	// Check if already handled
	if pending.Status != StatusPending {
		return ErrAlreadyHandled
	}

	pending.Status = StatusRejected
	return nil
}

// GetConfirmation retrieves a pending operation by its confirm key
func (m *Manager) GetConfirmation(confirmKey string) (*PendingOperation, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	pending, ok := m.pending[confirmKey]
	if !ok {
		return nil, ErrNotFound
	}

	// Check and update expiration status
	if pending.IsExpired() && pending.Status == StatusPending {
		pending.Status = StatusExpired
	}

	return pending, nil
}

// IsConfirmed returns true if the confirmation has been approved
func (m *Manager) IsConfirmed(confirmKey string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	pending, ok := m.pending[confirmKey]
	if !ok {
		return false
	}

	if pending.IsExpired() && pending.Status == StatusPending {
		return false
	}

	return pending.Status == StatusApproved
}

// CleanupExpired marks all expired pending operations as expired
// Returns the number of operations cleaned up
func (m *Manager) CleanupExpired() int {
	m.mu.Lock()
	defer m.mu.Unlock()

	count := 0
	now := time.Now()
	for _, pending := range m.pending {
		if pending.Status == StatusPending && now.After(pending.ExpiresAt) {
			pending.Status = StatusExpired
			count++
		}
	}
	return count
}

// StartCleanupRoutine starts a background goroutine that periodically cleans up expired operations
// Returns a stop channel that should be closed to stop the routine
func (m *Manager) StartCleanupRoutine(interval time.Duration) chan struct{} {
	stopCh := make(chan struct{})

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				m.CleanupExpired()
			case <-stopCh:
				return
			}
		}
	}()

	return stopCh
}

// CountPending returns the number of pending (not yet approved/rejected) operations
func (m *Manager) CountPending() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	count := 0
	for _, pending := range m.pending {
		if pending.Status == StatusPending {
			count++
		}
	}
	return count
}

// ListConfirmations returns all confirmation keys and their target clusters
func (m *Manager) ListConfirmations() ([]string, []string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	keys := make([]string, 0, len(m.pending))
	clusters := make([]string, 0, len(m.pending))

	for key, pending := range m.pending {
		keys = append(keys, key)
		clusters = append(clusters, pending.TargetCluster)
	}

	return keys, clusters
}

// GetPendingByCluster returns all pending operations for a specific cluster
func (m *Manager) GetPendingByCluster(clusterName string) []*PendingOperation {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*PendingOperation, 0)
	for _, pending := range m.pending {
		if pending.TargetCluster == clusterName && pending.Status == StatusPending {
			result = append(result, pending)
		}
	}

	return result
}

// String implements the Stringer interface for debugging
func (m *Manager) String() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return fmt.Sprintf("Manager{pending: %d}", len(m.pending))
}
