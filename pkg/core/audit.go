// Package core provides core functionality for the k8s-agent.
package core

import (
	"sync"
	"time"
)

// AuditEntry represents a single audit log entry.
type AuditEntry struct {
	SessionID string
	Timestamp time.Time
	Action    string
	Actor     string
	Details   map[string]interface{}
}

// auditLog stores audit entries indexed by session ID.
var auditLog = make(map[string][]AuditEntry)
var auditMutex sync.RWMutex

// Log records an audit entry for the given session.
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
func GetAuditLog(sessionID string) []AuditEntry {
	auditMutex.RLock()
	defer auditMutex.RUnlock()

	entries := make([]AuditEntry, len(auditLog[sessionID]))
	copy(entries, auditLog[sessionID])
	return entries
}