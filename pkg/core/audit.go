// Package core provides core functionality for the k8s-agent change management system.
// It implements a state machine for managing Kubernetes change operations through
// a structured workflow: Parse -> Clarify -> Plan -> Review -> Execute -> Complete/Failed.
//
// The package provides:
//   - State machine implementation for change session lifecycle
//   - Change session management with state transitions
//   - Audit logging for all state transitions and actions
package core

import (
	"sync"
	"time"
)

// AuditEntry represents a single audit log entry.
// All state transitions and significant actions are recorded to the audit log.
type AuditEntry struct {
	// SessionID is the ID of the session this entry belongs to.
	SessionID string
	// Timestamp is when the action occurred (UTC).
	Timestamp time.Time
	// Action is the name of the action (e.g., "state_transition", "execute_start").
	Action string
	// Actor is the component that performed the action (e.g., "session", "executor").
	Actor string
	// Details contains additional context about the action.
	Details map[string]interface{}
}

// auditLog stores audit entries indexed by session ID.
var auditLog = make(map[string][]AuditEntry)
var auditMutex sync.RWMutex

// Log records an audit entry for the given session.
// This function is thread-safe and can be called concurrently.
func Log(sessionID, action, actor string, details map[string]interface{}) {
	auditMutex.Lock()
	defer auditMutex.Unlock()

	entry := AuditEntry{
		SessionID: sessionID,
		Timestamp: time.Now().UTC(),
		Action:    action,
		Actor:     actor,
		Details:   details,
	}

	auditLog[sessionID] = append(auditLog[sessionID], entry)
}

// GetAuditLog returns all audit entries for the given session.
// Returns a copy of the entries to prevent external modification.
func GetAuditLog(sessionID string) []AuditEntry {
	auditMutex.RLock()
	defer auditMutex.RUnlock()

	entries := make([]AuditEntry, len(auditLog[sessionID]))
	copy(entries, auditLog[sessionID])
	return entries
}