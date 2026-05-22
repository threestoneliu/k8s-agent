package core

import (
	"sort"
	"sync"
	"testing"
	"time"
)

func TestLog(t *testing.T) {
	// Clear any existing entries for this test session
	sessionID := "test-session-log"
	auditMutex.Lock()
	delete(auditLog, sessionID)
	auditMutex.Unlock()

	Log(sessionID, "test-action", "test-actor", map[string]interface{}{"key": "value"})

	entries := GetAuditLog(sessionID)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	if entries[0].SessionID != sessionID {
		t.Errorf("expected SessionID %q, got %q", sessionID, entries[0].SessionID)
	}
	if entries[0].Action != "test-action" {
		t.Errorf("expected Action %q, got %q", "test-action", entries[0].Action)
	}
	if entries[0].Actor != "test-actor" {
		t.Errorf("expected Actor %q, got %q", "test-actor", entries[0].Actor)
	}
	if entries[0].Details["key"] != "value" {
		t.Errorf("expected Details[key] %q, got %q", "value", entries[0].Details["key"])
	}
	if entries[0].Timestamp.IsZero() {
		t.Error("expected Timestamp to be set")
	}
}

func TestLogMultipleEntries(t *testing.T) {
	sessionID := "test-session-multiple"
	auditMutex.Lock()
	delete(auditLog, sessionID)
	auditMutex.Unlock()

	Log(sessionID, "action-1", "actor-1", map[string]interface{}{"step": 1})
	Log(sessionID, "action-2", "actor-2", map[string]interface{}{"step": 2})
	Log(sessionID, "action-3", "actor-3", map[string]interface{}{"step": 3})

	entries := GetAuditLog(sessionID)
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}

	for i, entry := range entries {
		expectedAction := "action-" + string(rune('1'+i))
		if entry.Action != expectedAction {
			t.Errorf("entry %d: expected Action %q, got %q", i, expectedAction, entry.Action)
		}
	}
}

func TestLogMultipleSessions(t *testing.T) {
	session1 := "session-1"
	session2 := "session-2"

	auditMutex.Lock()
	delete(auditLog, session1)
	delete(auditLog, session2)
	auditMutex.Unlock()

	Log(session1, "action-s1", "actor-s1", nil)
	Log(session2, "action-s2", "actor-s2", nil)
	Log(session1, "action-s1-2", "actor-s1-2", nil)

	entries1 := GetAuditLog(session1)
	if len(entries1) != 2 {
		t.Fatalf("expected 2 entries for session1, got %d", len(entries1))
	}

	entries2 := GetAuditLog(session2)
	if len(entries2) != 1 {
		t.Fatalf("expected 1 entry for session2, got %d", len(entries2))
	}
}

func TestGetAuditLogEmpty(t *testing.T) {
	sessionID := "nonexistent-session"
	entries := GetAuditLog(sessionID)
	if entries == nil {
		t.Error("expected non-nil slice for nonexistent session")
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

func TestGetAuditLogReturnsCopy(t *testing.T) {
	sessionID := "test-session-copy"
	auditMutex.Lock()
	delete(auditLog, sessionID)
	auditMutex.Unlock()

	Log(sessionID, "action", "actor", nil)

	entries1 := GetAuditLog(sessionID)
	entries2 := GetAuditLog(sessionID)

	if &entries1[0] == &entries2[0] {
		t.Error("expected GetAuditLog to return independent copies")
	}
}

func TestLogConcurrent(t *testing.T) {
	sessionID := "test-session-concurrent"
	auditMutex.Lock()
	delete(auditLog, sessionID)
	auditMutex.Unlock()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			Log(sessionID, "action", "actor", map[string]interface{}{"i": i})
		}(i)
	}
	wg.Wait()

	entries := GetAuditLog(sessionID)
	if len(entries) != 100 {
		t.Errorf("expected 100 entries after concurrent writes, got %d", len(entries))
	}

	// Verify all entries are present (order may vary)
	values := make([]int, len(entries))
	for i, e := range entries {
		v, ok := e.Details["i"].(int)
		if !ok {
			t.Errorf("entry %d: expected int detail, got %T", i, e.Details["i"])
			continue
		}
		values[i] = v
	}
	sort.Ints(values)
	for i, v := range values {
		if v != i {
			t.Errorf("expected value %d at index %d, got %d", i, i, v)
		}
	}
}

func TestAuditEntryTimestampIsRecent(t *testing.T) {
	before := time.Now().UTC().Add(-time.Second)
	sessionID := "test-session-timestamp"

	auditMutex.Lock()
	delete(auditLog, sessionID)
	auditMutex.Unlock()

	Log(sessionID, "action", "actor", nil)

	after := time.Now().UTC().Add(time.Second)
	entries := GetAuditLog(sessionID)
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}

	if entries[0].Timestamp.Before(before) || entries[0].Timestamp.After(after) {
		t.Errorf("Timestamp %v not in expected range [%v, %v]", entries[0].Timestamp, before, after)
	}
}