package confirmation

import (
	"fmt"
	"time"

	"k8s-agent/pkg/engine"
)

// ConfirmationStatus represents the status of a confirmation request
type ConfirmationStatus int

const (
	StatusPending ConfirmationStatus = iota
	StatusApproved
	StatusRejected
	StatusExpired
)

// String returns the string representation of ConfirmationStatus
func (s ConfirmationStatus) String() string {
	switch s {
	case StatusPending:
		return "pending"
	case StatusApproved:
		return "approved"
	case StatusRejected:
		return "rejected"
	case StatusExpired:
		return "expired"
	default:
		return fmt.Sprintf("unknown(%d)", s)
	}
}

// PendingOperation represents a pending operation awaiting confirmation
type PendingOperation struct {
	ID            string
	ConfirmKey    string // 6-digit confirmation code
	Operation     *engine.ClassifiedOperation
	Status        ConfirmationStatus
	CreatedAt     time.Time
	ExpiresAt     time.Time // TTL 5 minutes
	TargetCluster string
}

// IsExpired checks if the pending operation has expired
func (p *PendingOperation) IsExpired() bool {
	return time.Now().After(p.ExpiresAt)
}

// NewPendingOperation creates a new pending operation with a generated ID and confirm key
func NewPendingOperation(clusterName string, op *engine.ClassifiedOperation) *PendingOperation {
	confirmKey, _ := GenerateSecureConfirmKey() // Error ignored as GenerateSecureConfirmKey always succeeds
	return &PendingOperation{
		ID:            generateID(),
		ConfirmKey:    confirmKey,
		Operation:     op,
		Status:        StatusPending,
		CreatedAt:     time.Now(),
		ExpiresAt:     time.Now().Add(defaultTTL),
		TargetCluster: clusterName,
	}
}

// defaultTTL is the default time-to-live for confirmations
const defaultTTL = 5 * time.Minute
