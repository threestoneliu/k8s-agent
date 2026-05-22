package core

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// TestAuditStateTransition verifies that state transitions are logged.
func TestAuditStateTransition(t *testing.T) {
	sessionID := "test-session-transition"
	auditMutex.Lock()
	delete(auditLog, sessionID)
	auditMutex.Unlock()

	// Create a new session in PARSING state
	session := NewChangeSession(sessionID)

	// Verify initial state
	if session.State != StateParsing {
		t.Errorf("expected initial state PARSING, got %s", session.State)
	}

	// Transition: PARSING -> PLANNING (via SignalConfirm)
	err := session.Transition(SignalConfirm)
	if err != nil {
		t.Fatalf("Transition failed: %v", err)
	}

	// Verify audit log has state_transition entry
	entries := GetAuditLog(sessionID)
	if len(entries) == 0 {
		t.Fatal("expected audit entries for state transition")
	}

	// Find the state_transition entry
	var found bool
	for _, entry := range entries {
		if entry.Action == "state_transition" {
			found = true
			if entry.Actor != "session" {
				t.Errorf("expected actor 'session', got %q", entry.Actor)
			}
			details, ok := entry.Details["from_state"].(string)
			if !ok || details != "PARSING" {
				t.Errorf("expected from_state 'PARSING', got %v", entry.Details["from_state"])
			}
			details, ok = entry.Details["to_state"].(string)
			if !ok || details != "PLANNING" {
				t.Errorf("expected to_state 'PLANNING', got %v", entry.Details["to_state"])
			}
			break
		}
	}
	if !found {
		t.Error("expected to find state_transition audit entry")
	}
}

// TestAuditStateTransitionMultiple verifies multiple state transitions are logged.
func TestAuditStateTransitionMultiple(t *testing.T) {
	sessionID := "test-session-multi-transition"
	auditMutex.Lock()
	delete(auditLog, sessionID)
	auditMutex.Unlock()

	session := NewChangeSession(sessionID)

	// PARSING -> PLANNING (SignalConfirm)
	session.Transition(SignalConfirm)
	// PLANNING -> REVIEWING (SignalConfirm)
	session.Transition(SignalConfirm)
	// REVIEWING -> EXECUTING (SignalConfirm)
	session.Transition(SignalConfirm)
	// EXECUTING -> COMPLETED (SignalConfirm)
	session.Transition(SignalConfirm)

	entries := GetAuditLog(sessionID)

	// Count state_transition entries
	transitionCount := 0
	for _, entry := range entries {
		if entry.Action == "state_transition" {
			transitionCount++
		}
	}

	if transitionCount != 4 {
		t.Errorf("expected 4 state transitions, got %d", transitionCount)
	}
}

// TestAuditExecutorExecute verifies that executor operations are logged.
func TestAuditExecutorExecute(t *testing.T) {
	sessionID := "test-session-execute"
	auditMutex.Lock()
	delete(auditLog, sessionID)
	auditMutex.Unlock()

	session := &ChangeSession{
		ID:    sessionID,
		State: StateExecuting,
	}

	plan := &ChangePlan{
		ID:        "test-plan-1",
		Summary:   "Test plan",
		RiskLevel: RiskLow,
		Steps: []ChangeStep{
			{
				Seq:        1,
				Action:     ActionInspect,
				Target:     ResourceTarget{Name: "test-resource", Kind: "ConfigMap", Namespace: "default"},
				RiskLevel:  RiskLow,
				CanRollback: false,
				Description: "Test step",
			},
		},
	}

	err := Execute(session, plan)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	entries := GetAuditLog(sessionID)

	// Verify we have execute_start entry
	found := false
	for _, entry := range entries {
		if entry.Action == "execute_start" && entry.Actor == "executor" {
			found = true
			if entry.Details["plan_id"] != "test-plan-1" {
				t.Errorf("expected plan_id 'test-plan-1', got %v", entry.Details["plan_id"])
			}
			break
		}
	}
	if !found {
		t.Error("expected to find execute_start audit entry")
	}

	// Verify we have execute_completed entry
	found = false
	for _, entry := range entries {
		if entry.Action == "execute_completed" && entry.Actor == "executor" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find execute_completed audit entry")
	}

	// Verify step entries
	stepStartFound := false
	stepCompletedFound := false
	for _, entry := range entries {
		if entry.Action == "step_start" && entry.Actor == "executor" {
			stepStartFound = true
		}
		if entry.Action == "step_completed" && entry.Actor == "executor" {
			stepCompletedFound = true
		}
	}
	if !stepStartFound {
		t.Error("expected to find step_start audit entry")
	}
	if !stepCompletedFound {
		t.Error("expected to find step_completed audit entry")
	}
}

// TestAuditExecutorExecuteMultipleSteps verifies executor logs all steps.
func TestAuditExecutorExecuteMultipleSteps(t *testing.T) {
	sessionID := "test-session-multi-step"
	auditMutex.Lock()
	delete(auditLog, sessionID)
	auditMutex.Unlock()

	session := &ChangeSession{
		ID:    sessionID,
		State: StateExecuting,
	}

	plan := &ChangePlan{
		ID:        "test-plan-multi",
		Summary:   "Test plan with multiple steps",
		RiskLevel: RiskMedium,
		Steps: []ChangeStep{
			{
				Seq:        1,
				Action:     ActionInspect,
				Target:     ResourceTarget{Name: "test-resource", Kind: "ConfigMap", Namespace: "default"},
				RiskLevel:  RiskLow,
				CanRollback: false,
				Description: "Inspect step",
			},
			{
				Seq:        2,
				Action:     ActionUpdate,
				Target:     ResourceTarget{Name: "test-resource", Kind: "ConfigMap", Namespace: "default"},
				RiskLevel:  RiskMedium,
				CanRollback: true,
				Description: "Update step",
			},
			{
				Seq:        3,
				Action:     ActionInspect,
				Target:     ResourceTarget{Name: "test-resource", Kind: "ConfigMap", Namespace: "default"},
				RiskLevel:  RiskLow,
				CanRollback: false,
				Description: "Verify step",
			},
		},
	}

	err := Execute(session, plan)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	entries := GetAuditLog(sessionID)

	// Count step_start entries
	stepStartCount := 0
	stepCompletedCount := 0
	for _, entry := range entries {
		if entry.Action == "step_start" && entry.Actor == "executor" {
			stepStartCount++
		}
		if entry.Action == "step_completed" && entry.Actor == "executor" {
			stepCompletedCount++
		}
	}

	if stepStartCount != 3 {
		t.Errorf("expected 3 step_start entries, got %d", stepStartCount)
	}
	if stepCompletedCount != 3 {
		t.Errorf("expected 3 step_completed entries, got %d", stepCompletedCount)
	}
}

// TestAuditRollback verifies that rollback operations are logged.
func TestAuditRollback(t *testing.T) {
	ClearSnapshots()
	defer ClearSnapshots()

	sessionID := "test-session-rollback"
	auditMutex.Lock()
	delete(auditLog, sessionID)
	auditMutex.Unlock()

	// Create a snapshot first
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata":   map[string]interface{}{"name": "my-deploy", "namespace": "default"},
			"spec":       map[string]interface{}{"replicas": int64(3)},
		},
	}

	resourceID := ResourceID{Name: "my-deploy", Kind: "Deployment", Namespace: "default", APIVersion: "apps/v1"}
	CreateSnapshot(sessionID, resourceID, obj)

	// Perform rollback
	restored, err := Rollback(sessionID, resourceID)
	if err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}
	if restored == nil {
		t.Fatal("Rollback returned nil object")
	}

	entries := GetAuditLog(sessionID)

	// Verify rollback entry exists
	found := false
	for _, entry := range entries {
		if entry.Action == "rollback" && entry.Actor == "rollback" {
			found = true
			if entry.Details["resource"] != resourceID.String() {
				t.Errorf("expected resource %q, got %v", resourceID.String(), entry.Details["resource"])
			}
			break
		}
	}
	if !found {
		t.Error("expected to find rollback audit entry")
	}

	// Verify rollback_completed entry
	found = false
	for _, entry := range entries {
		if entry.Action == "rollback_completed" && entry.Actor == "rollback" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find rollback_completed audit entry")
	}
}

// TestAuditRollbackToSnapshot verifies that RollbackToSnapshot is logged.
func TestAuditRollbackToSnapshot(t *testing.T) {
	ClearSnapshots()
	defer ClearSnapshots()

	sessionID := "test-session-rollback-to-snap"
	auditMutex.Lock()
	delete(auditLog, sessionID)
	auditMutex.Unlock()

	// Create a snapshot
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata":   map[string]interface{}{"name": "my-deploy", "namespace": "default"},
		},
	}

	resourceID := ResourceID{Name: "my-deploy", Kind: "Deployment", Namespace: "default", APIVersion: "apps/v1"}
	snap, _ := CreateSnapshot(sessionID, resourceID, obj)

	// Rollback to specific snapshot
	restored, err := RollbackToSnapshot(snap.ID)
	if err != nil {
		t.Fatalf("RollbackToSnapshot failed: %v", err)
	}
	if restored == nil {
		t.Fatal("RollbackToSnapshot returned nil object")
	}

	entries := GetAuditLog(sessionID)

	// Verify rollback_to_snapshot entry
	found := false
	for _, entry := range entries {
		if entry.Action == "rollback_to_snapshot" && entry.Actor == "rollback" {
			found = true
			if entry.Details["snapshot_id"] != snap.ID {
				t.Errorf("expected snapshot_id %q, got %v", snap.ID, entry.Details["snapshot_id"])
			}
			break
		}
	}
	if !found {
		t.Error("expected to find rollback_to_snapshot audit entry")
	}
}

// TestAuditRollbackFailed verifies that failed rollbacks are logged.
func TestAuditRollbackFailed(t *testing.T) {
	ClearSnapshots()
	defer ClearSnapshots()

	sessionID := "test-session-rollback-failed"
	auditMutex.Lock()
	delete(auditLog, sessionID)
	auditMutex.Unlock()

	resourceID := ResourceID{Name: "nonexistent", Kind: "Deployment", Namespace: "default", APIVersion: "apps/v1"}

	// Attempt rollback with no snapshot
	_, err := Rollback(sessionID, resourceID)
	if err != ErrNoSnapshotForResource {
		t.Fatalf("expected ErrNoSnapshotForResource, got %v", err)
	}

	entries := GetAuditLog(sessionID)

	// Verify rollback_failed entry
	found := false
	for _, entry := range entries {
		if entry.Action == "rollback_failed" && entry.Actor == "rollback" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find rollback_failed audit entry")
	}
}

// TestAuditExecutorFailed verifies that executor failures are logged.
func TestAuditExecutorFailed(t *testing.T) {
	sessionID := "test-session-execute-failed"
	auditMutex.Lock()
	delete(auditLog, sessionID)
	auditMutex.Unlock()

	session := &ChangeSession{
		ID:    sessionID,
		State: StateExecuting,
	}

	// Use a plan with an unknown action to force failure
	plan := &ChangePlan{
		ID:        "test-plan-fail",
		Summary:   "Test plan",
		RiskLevel: RiskLow,
		Steps: []ChangeStep{
			{
				Seq:        1,
				Action:     "UNKNOWN_ACTION",
				Target:     ResourceTarget{Name: "test-resource", Kind: "ConfigMap", Namespace: "default"},
				RiskLevel:  RiskLow,
				CanRollback: false,
				Description: "Unknown step",
			},
		},
	}

	err := Execute(session, plan)
	if err == nil {
		t.Fatal("expected Execute to fail")
	}

	entries := GetAuditLog(sessionID)

	// Verify execute_steps_failed entry
	found := false
	for _, entry := range entries {
		if entry.Action == "execute_steps_failed" && entry.Actor == "executor" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find execute_steps_failed audit entry")
	}
}

// TestAuditExecutorPreCheckFailed verifies pre-check failures are logged.
func TestAuditExecutorPreCheckFailed(t *testing.T) {
	sessionID := "test-session-precheck-failed"
	auditMutex.Lock()
	delete(auditLog, sessionID)
	auditMutex.Unlock()

	session := &ChangeSession{
		ID:    sessionID,
		State: StateExecuting,
	}

	// Temporarily replace preChecks with a failing critical check
	tempChecks := []PreCheck{
		{
			Name:     "critical_failing",
			Critical: true,
			Run: func(s *ChangeSession, p *ChangePlan) CheckResult {
				return CheckResult{
					Passed:  false,
					Message: "critical failure",
					Details: "this check must fail",
				}
			},
		},
	}

	originalChecks := preChecks
	preChecks = tempChecks
	defer func() { preChecks = originalChecks }()

	plan := &ChangePlan{
		ID:        "test-plan-precheck-fail",
		Summary:   "Test plan",
		RiskLevel: RiskLow,
		Steps:     []ChangeStep{},
	}

	err := Execute(session, plan)
	if err == nil {
		t.Fatal("expected Execute to fail due to pre-check failure")
	}

	entries := GetAuditLog(sessionID)

	// Verify execute_precheck_failed entry
	found := false
	for _, entry := range entries {
		if entry.Action == "execute_precheck_failed" && entry.Actor == "executor" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected to find execute_precheck_failed audit entry")
	}
}

// TestAuditStepFailed verifies that individual step failures are logged.
func TestAuditStepFailed(t *testing.T) {
	sessionID := "test-session-step-failed"
	auditMutex.Lock()
	delete(auditLog, sessionID)
	auditMutex.Unlock()

	session := &ChangeSession{
		ID:    sessionID,
		State: StateExecuting,
	}

	plan := &ChangePlan{
		ID:        "test-plan-step-fail",
		Summary:   "Test plan",
		RiskLevel: RiskLow,
		Steps: []ChangeStep{
			{
				Seq:        1,
				Action:     "UNKNOWN_ACTION",
				Target:     ResourceTarget{Name: "test-resource", Kind: "ConfigMap", Namespace: "default"},
				RiskLevel:  RiskLow,
				CanRollback: false,
				Description: "Unknown step",
			},
		},
	}

	err := Execute(session, plan)
	if err == nil {
		t.Fatal("expected Execute to fail")
	}

	entries := GetAuditLog(sessionID)

	// Verify step_failed entry
	found := false
	for _, entry := range entries {
		if entry.Action == "step_failed" && entry.Actor == "executor" {
			found = true
			if entry.Details["step_seq"] != 1 {
				t.Errorf("expected step_seq 1, got %v", entry.Details["step_seq"])
			}
			break
		}
	}
	if !found {
		t.Error("expected to find step_failed audit entry")
	}
}
