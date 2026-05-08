package confirmation

import (
	"testing"
	"time"

	"k8s-agent/pkg/engine"
)

func TestConfirmationStatus_Values(t *testing.T) {
	if StatusPending != 0 {
		t.Errorf("StatusPending = %d, want 0", StatusPending)
	}
	if StatusApproved != 1 {
		t.Errorf("StatusApproved = %d, want 1", StatusApproved)
	}
	if StatusRejected != 2 {
		t.Errorf("StatusRejected = %d, want 2", StatusRejected)
	}
	if StatusExpired != 3 {
		t.Errorf("StatusExpired = %d, want 3", StatusExpired)
	}
}

func TestPendingOperation_Fields(t *testing.T) {
	op := &engine.ClassifiedOperation{
		Type:      engine.OperationTypeMutation,
		Verb:      "delete",
		Resource:  "pod",
		Name:      "test-pod",
		Namespace: "default",
	}

	now := time.Now()
	expiresAt := now.Add(5 * time.Minute)

	pending := &PendingOperation{
		ID:            "test-id",
		ConfirmKey:    "123456",
		Operation:     op,
		Status:        StatusPending,
		CreatedAt:     now,
		ExpiresAt:     expiresAt,
		TargetCluster: "test-cluster",
	}

	if pending.ID != "test-id" {
		t.Errorf("ID = %s, want test-id", pending.ID)
	}
	if pending.ConfirmKey != "123456" {
		t.Errorf("ConfirmKey = %s, want 123456", pending.ConfirmKey)
	}
	if pending.Operation != op {
		t.Error("Operation does not match")
	}
	if pending.Status != StatusPending {
		t.Errorf("Status = %v, want %v", pending.Status, StatusPending)
	}
	if !pending.CreatedAt.Equal(now) {
		t.Errorf("CreatedAt = %v, want %v", pending.CreatedAt, now)
	}
	if !pending.ExpiresAt.Equal(expiresAt) {
		t.Errorf("ExpiresAt = %v, want %v", pending.ExpiresAt, expiresAt)
	}
	if pending.TargetCluster != "test-cluster" {
		t.Errorf("TargetCluster = %s, want test-cluster", pending.TargetCluster)
	}
}

func TestPendingOperation_IsExpired(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name      string
		expiresAt time.Time
		expected  bool
	}{
		{"not expired", now.Add(5 * time.Minute), false},
		{"just expired", now.Add(-1 * time.Second), true},
		{"long expired", now.Add(-1 * time.Hour), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pending := &PendingOperation{
				ExpiresAt: tt.expiresAt,
			}
			if got := pending.IsExpired(); got != tt.expected {
				t.Errorf("IsExpired() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPendingOperation_IDGeneration(t *testing.T) {
	// Verify that we can create multiple pending operations with unique IDs
	pending1 := NewPendingOperation("cluster1", &engine.ClassifiedOperation{
		Type:     engine.OperationTypeMutation,
		Verb:     "delete",
		Resource: "pod",
	})

	pending2 := NewPendingOperation("cluster2", &engine.ClassifiedOperation{
		Type:     engine.OperationTypeMutation,
		Verb:     "create",
		Resource: "deployment",
	})

	if pending1.ID == pending2.ID {
		t.Error("Two pending operations should have unique IDs")
	}

	if pending1.ConfirmKey == pending2.ConfirmKey {
		t.Error("Two pending operations should have unique confirm keys")
	}
}
