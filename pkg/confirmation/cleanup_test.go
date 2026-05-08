package confirmation

import (
	"testing"
	"time"

	"k8s-agent/pkg/engine"
)

func TestCleanupExpired(t *testing.T) {
	m := NewManager(1 * time.Millisecond) // Very short TTL for testing

	op := &engine.ClassifiedOperation{
		Type:     engine.OperationTypeMutation,
		Verb:     "delete",
		Resource: "pod",
	}

	// Create multiple confirmations
	keys := make([]string, 5)
	for i := 0; i < 5; i++ {
		key, err := m.CreateConfirmation("test-cluster", op)
		if err != nil {
			t.Fatalf("CreateConfirmation() error = %v", err)
		}
		keys[i] = key
	}

	// Wait for expiration
	time.Sleep(10 * time.Millisecond)

	// Run cleanup
	cleaned := m.CleanupExpired()
	if cleaned != 5 {
		t.Errorf("CleanupExpired() cleaned %d, want 5", cleaned)
	}

	// Verify all are expired
	for _, key := range keys {
		pending, _ := m.GetConfirmation(key)
		if pending.Status != StatusExpired {
			t.Errorf("Key %s status = %v, want %v", key, pending.Status, StatusExpired)
		}
	}
}

func TestCleanupExpired_NoneToClean(t *testing.T) {
	m := NewManager(5 * time.Minute)

	op := &engine.ClassifiedOperation{
		Type:     engine.OperationTypeMutation,
		Verb:     "delete",
		Resource: "pod",
	}

	// Create confirmations that are not expired
	for i := 0; i < 3; i++ {
		_, err := m.CreateConfirmation("test-cluster", op)
		if err != nil {
			t.Fatalf("CreateConfirmation() error = %v", err)
		}
	}

	cleaned := m.CleanupExpired()
	if cleaned != 0 {
		t.Errorf("CleanupExpired() cleaned %d, want 0", cleaned)
	}
}

func TestCleanupExpired_PartialCleanup(t *testing.T) {
	// This test is tricky because we can't easily test partial cleanup
	// without time manipulation. We verify that the method works correctly.
	m := NewManager(5 * time.Minute)

	op := &engine.ClassifiedOperation{
		Type:     engine.OperationTypeMutation,
		Verb:     "delete",
		Resource: "pod",
	}

	// Create some confirmations
	for i := 0; i < 3; i++ {
		_, err := m.CreateConfirmation("test-cluster", op)
		if err != nil {
			t.Fatalf("CreateConfirmation() error = %v", err)
		}
	}

	// Cleanup should not affect non-expired items
	m.CleanupExpired()

	if m.CountPending() != 3 {
		t.Errorf("After CleanupExpired(), pending count = %d, want 3", m.CountPending())
	}
}

func TestStartCleanupRoutine(t *testing.T) {
	m := NewManager(10 * time.Millisecond)

	op := &engine.ClassifiedOperation{
		Type:     engine.OperationTypeMutation,
		Verb:     "delete",
		Resource: "pod",
	}

	// Create a confirmation
	confirmKey, err := m.CreateConfirmation("test-cluster", op)
	if err != nil {
		t.Fatalf("CreateConfirmation() error = %v", err)
	}

	// Start cleanup routine
	stopCh := m.StartCleanupRoutine(5 * time.Millisecond)
	defer close(stopCh)

	// Wait for cleanup to run
	time.Sleep(30 * time.Millisecond)

	// The confirmation should be expired and cleaned
	pending, err := m.GetConfirmation(confirmKey)
	if err != nil {
		t.Fatalf("GetConfirmation() error = %v", err)
	}

	if pending.Status != StatusExpired {
		t.Errorf("Status = %v, want %v", pending.Status, StatusExpired)
	}
}

func TestStartCleanupRoutine_StopChannel(t *testing.T) {
	m := NewManager(10 * time.Millisecond)

	op := &engine.ClassifiedOperation{
		Type:     engine.OperationTypeMutation,
		Verb:     "delete",
		Resource: "pod",
	}

	// Create a confirmation
	_, err := m.CreateConfirmation("test-cluster", op)
	if err != nil {
		t.Fatalf("CreateConfirmation() error = %v", err)
	}

	// Start cleanup routine and immediately stop
	stopCh := m.StartCleanupRoutine(5 * time.Millisecond)
	time.Sleep(1 * time.Millisecond)
	close(stopCh)

	// Give some time for the cleanup to potentially run
	time.Sleep(10 * time.Millisecond)

	// The Manager should still be functional
	confirmKey, err := m.CreateConfirmation("test-cluster", op)
	if err != nil {
		t.Fatalf("CreateConfirmation() after stop error = %v", err)
	}

	pending, err := m.GetConfirmation(confirmKey)
	if err != nil {
		t.Fatalf("GetConfirmation() error = %v", err)
	}

	if pending.Status != StatusPending {
		t.Errorf("New confirmation status = %v, want %v", pending.Status, StatusPending)
	}
}

func TestCountPending(t *testing.T) {
	m := NewManager(5 * time.Minute)

	if m.CountPending() != 0 {
		t.Errorf("Initial CountPending() = %d, want 0", m.CountPending())
	}

	op := &engine.ClassifiedOperation{
		Type:     engine.OperationTypeMutation,
		Verb:     "delete",
		Resource: "pod",
	}

	// Create 3 confirmations
	for i := 0; i < 3; i++ {
		_, err := m.CreateConfirmation("test-cluster", op)
		if err != nil {
			t.Fatalf("CreateConfirmation() error = %v", err)
		}
	}

	if m.CountPending() != 3 {
		t.Errorf("After creating 3, CountPending() = %d, want 3", m.CountPending())
	}

	// Approve one
	keys, _ := m.ListConfirmations()
	m.ApproveConfirmation(keys[0])

	if m.CountPending() != 2 {
		t.Errorf("After approving 1, CountPending() = %d, want 2", m.CountPending())
	}
}

func TestListConfirmations(t *testing.T) {
	m := NewManager(5 * time.Minute)

	op := &engine.ClassifiedOperation{
		Type:     engine.OperationTypeMutation,
		Verb:     "delete",
		Resource: "pod",
	}

	// Create confirmations for different clusters
	m.CreateConfirmation("cluster1", op)
	m.CreateConfirmation("cluster2", op)
	m.CreateConfirmation("cluster3", op)

	keys, clusters := m.ListConfirmations()
	if len(keys) != 3 {
		t.Errorf("ListConfirmations() returned %d keys, want 3", len(keys))
	}

	if len(clusters) != 3 {
		t.Errorf("ListConfirmations() returned %d clusters, want 3", len(clusters))
	}
}
