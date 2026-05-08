package confirmation

import (
	"sync"
	"testing"
	"time"

	"k8s-agent/pkg/engine"
)

func TestCreateConfirmation(t *testing.T) {
	m := NewManager(5 * time.Minute)

	op := &engine.ClassifiedOperation{
		Type:      engine.OperationTypeMutation,
		Verb:      "delete",
		Resource:  "pod",
		Name:      "my-pod",
		Namespace: "default",
	}

	confirmKey, err := m.CreateConfirmation("test-cluster", op)
	if err != nil {
		t.Fatalf("CreateConfirmation() error = %v", err)
	}

	if !ValidateConfirmKey(confirmKey) {
		t.Errorf("CreateConfirmation() returned invalid key format: %s", confirmKey)
	}

	pending, err := m.GetConfirmation(confirmKey)
	if err != nil {
		t.Fatalf("GetConfirmation() error = %v", err)
	}

	if pending == nil {
		t.Fatal("GetConfirmation() returned nil pending operation")
	}

	if pending.Status != StatusPending {
		t.Errorf("Pending operation status = %v, want %v", pending.Status, StatusPending)
	}

	if pending.TargetCluster != "test-cluster" {
		t.Errorf("Pending operation cluster = %s, want %s", pending.TargetCluster, "test-cluster")
	}

	if pending.Operation != op {
		t.Error("Pending operation does not match original operation")
	}

	if pending.ExpiresAt.Before(time.Now().Add(4 * time.Minute)) {
		t.Error("ExpiresAt should be at least 4 minutes in the future for 5 minute TTL")
	}
}

func TestCreateConfirmation_UniqueKeys(t *testing.T) {
	m := NewManager(5 * time.Minute)

	op := &engine.ClassifiedOperation{
		Type:     engine.OperationTypeMutation,
		Verb:     "delete",
		Resource: "pod",
	}

	keys := make(map[string]bool)
	for i := 0; i < 50; i++ {
		key, err := m.CreateConfirmation("test-cluster", op)
		if err != nil {
			t.Fatalf("CreateConfirmation() iteration %d error = %v", i, err)
		}
		if keys[key] {
			t.Errorf("CreateConfirmation() generated duplicate key: %s", key)
		}
		keys[key] = true
	}
}

func TestApproveConfirmation(t *testing.T) {
	m := NewManager(5 * time.Minute)

	op := &engine.ClassifiedOperation{
		Type:     engine.OperationTypeMutation,
		Verb:     "delete",
		Resource: "pod",
	}

	confirmKey, err := m.CreateConfirmation("test-cluster", op)
	if err != nil {
		t.Fatalf("CreateConfirmation() error = %v", err)
	}

	err = m.ApproveConfirmation(confirmKey)
	if err != nil {
		t.Fatalf("ApproveConfirmation() error = %v", err)
	}

	pending, err := m.GetConfirmation(confirmKey)
	if err != nil {
		t.Fatalf("GetConfirmation() error = %v", err)
	}

	if pending.Status != StatusApproved {
		t.Errorf("After approval, status = %v, want %v", pending.Status, StatusApproved)
	}

	if !m.IsConfirmed(confirmKey) {
		t.Error("IsConfirmed() should return true after approval")
	}
}

func TestApproveConfirmation_NotFound(t *testing.T) {
	m := NewManager(5 * time.Minute)

	err := m.ApproveConfirmation("000000")
	if err == nil {
		t.Error("ApproveConfirmation() should return error for non-existent key")
	}
}

func TestRejectConfirmation(t *testing.T) {
	m := NewManager(5 * time.Minute)

	op := &engine.ClassifiedOperation{
		Type:     engine.OperationTypeMutation,
		Verb:     "delete",
		Resource: "pod",
	}

	confirmKey, err := m.CreateConfirmation("test-cluster", op)
	if err != nil {
		t.Fatalf("CreateConfirmation() error = %v", err)
	}

	err = m.RejectConfirmation(confirmKey)
	if err != nil {
		t.Fatalf("RejectConfirmation() error = %v", err)
	}

	pending, err := m.GetConfirmation(confirmKey)
	if err != nil {
		t.Fatalf("GetConfirmation() error = %v", err)
	}

	if pending.Status != StatusRejected {
		t.Errorf("After rejection, status = %v, want %v", pending.Status, StatusRejected)
	}

	if m.IsConfirmed(confirmKey) {
		t.Error("IsConfirmed() should return false after rejection")
	}
}

func TestRejectConfirmation_NotFound(t *testing.T) {
	m := NewManager(5 * time.Minute)

	err := m.RejectConfirmation("000000")
	if err == nil {
		t.Error("RejectConfirmation() should return error for non-existent key")
	}
}

func TestExpiredConfirmation(t *testing.T) {
	m := NewManager(1 * time.Millisecond) // Very short TTL for testing

	op := &engine.ClassifiedOperation{
		Type:     engine.OperationTypeMutation,
		Verb:     "delete",
		Resource: "pod",
	}

	confirmKey, err := m.CreateConfirmation("test-cluster", op)
	if err != nil {
		t.Fatalf("CreateConfirmation() error = %v", err)
	}

	// Wait for expiration
	time.Sleep(10 * time.Millisecond)

	pending, err := m.GetConfirmation(confirmKey)
	if err != nil {
		t.Fatalf("GetConfirmation() error = %v", err)
	}

	if pending.Status != StatusExpired {
		t.Errorf("After expiration, status = %v, want %v", pending.Status, StatusExpired)
	}

	if m.IsConfirmed(confirmKey) {
		t.Error("IsConfirmed() should return false for expired confirmation")
	}
}

func TestApproveExpiredConfirmation(t *testing.T) {
	m := NewManager(1 * time.Millisecond) // Very short TTL for testing

	op := &engine.ClassifiedOperation{
		Type:     engine.OperationTypeMutation,
		Verb:     "delete",
		Resource: "pod",
	}

	confirmKey, err := m.CreateConfirmation("test-cluster", op)
	if err != nil {
		t.Fatalf("CreateConfirmation() error = %v", err)
	}

	// Wait for expiration
	time.Sleep(10 * time.Millisecond)

	err = m.ApproveConfirmation(confirmKey)
	if err == nil {
		t.Error("ApproveConfirmation() should return error for expired confirmation")
	}
}

func TestGetConfirmation_NotFound(t *testing.T) {
	m := NewManager(5 * time.Minute)

	pending, err := m.GetConfirmation("000000")
	if err == nil {
		t.Error("GetConfirmation() should return error for non-existent key")
	}
	if pending != nil {
		t.Error("GetConfirmation() should return nil for non-existent key")
	}
}

func TestGetConfirmation_AfterRejection(t *testing.T) {
	m := NewManager(5 * time.Minute)

	op := &engine.ClassifiedOperation{
		Type:     engine.OperationTypeMutation,
		Verb:     "delete",
		Resource: "pod",
	}

	confirmKey, err := m.CreateConfirmation("test-cluster", op)
	if err != nil {
		t.Fatalf("CreateConfirmation() error = %v", err)
	}

	m.RejectConfirmation(confirmKey)

	pending, err := m.GetConfirmation(confirmKey)
	if err != nil {
		t.Fatalf("GetConfirmation() error = %v", err)
	}

	if pending == nil {
		t.Fatal("GetConfirmation() should return pending operation after rejection")
	}

	if pending.Operation.Type != engine.OperationTypeMutation {
		t.Errorf("Operation type = %v, want %v", pending.Operation.Type, engine.OperationTypeMutation)
	}
}

func TestConcurrentAccess(t *testing.T) {
	m := NewManager(5 * time.Minute)

	op := &engine.ClassifiedOperation{
		Type:     engine.OperationTypeMutation,
		Verb:     "delete",
		Resource: "pod",
	}

	confirmKey, err := m.CreateConfirmation("test-cluster", op)
	if err != nil {
		t.Fatalf("CreateConfirmation() error = %v", err)
	}

	// Concurrent reads
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				m.GetConfirmation(confirmKey)
				m.IsConfirmed(confirmKey)
			}
			done <- true
		}()
	}

	// Concurrent approval attempts
	go func() {
		for j := 0; j < 100; j++ {
			m.ApproveConfirmation(confirmKey)
		}
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 11; i++ {
		<-done
	}

	// Verify final state
	pending, err := m.GetConfirmation(confirmKey)
	if err != nil {
		t.Fatalf("GetConfirmation() error = %v", err)
	}
	if pending.Status != StatusApproved {
		t.Errorf("Final status = %v, want %v", pending.Status, StatusApproved)
	}
}

func TestConcurrentCreate(t *testing.T) {
	m := NewManager(5 * time.Minute)

	op := &engine.ClassifiedOperation{
		Type:     engine.OperationTypeMutation,
		Verb:     "delete",
		Resource: "pod",
	}

	var wg sync.WaitGroup
	keys := make([]string, 500)
	index := 0
	var keysMu sync.Mutex

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				key, err := m.CreateConfirmation("test-cluster", op)
				if err != nil {
					t.Errorf("CreateConfirmation() error = %v", err)
					continue
				}
				keysMu.Lock()
				keys[index] = key
				index++
				keysMu.Unlock()
			}
		}()
	}

	// Wait for all goroutines
	wg.Wait()

	// Verify all keys are unique
	uniqueKeys := make(map[string]bool)
	for _, key := range keys {
		if key == "" {
			continue
		}
		if uniqueKeys[key] {
			t.Errorf("Duplicate key generated: %s", key)
		}
		uniqueKeys[key] = true
	}

	if len(uniqueKeys) != 500 {
		t.Errorf("Expected 500 unique keys, got %d", len(uniqueKeys))
	}
}
