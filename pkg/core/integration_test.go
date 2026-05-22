package core

import (
	"fmt"
	"sync"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// TestIntegration_CompleteFlow tests the complete happy path:
// PARSING → CLARIFYING → PLANNING → REVIEWING → EXECUTING → COMPLETED
func TestIntegration_CompleteFlow(t *testing.T) {
	// Start with PARSING state
	session := NewChangeSession("test-session-complete")
	if session.State != StateParsing {
		t.Errorf("expected initial state PARSING, got %s", session.State)
	}

	// PARSING → CLARIFYING (via Modify signal when intent is incomplete)
	err := session.Transition(SignalModify)
	if err != nil {
		t.Fatalf("PARSING → CLARIFYING failed: %v", err)
	}
	if session.State != StateClarifying {
		t.Errorf("expected state CLARIFYING, got %s", session.State)
	}

	// CLARIFYING → PARSING (via Confirm after clarification)
	err = session.Transition(SignalConfirm)
	if err != nil {
		t.Fatalf("CLARIFYING → PARSING failed: %v", err)
	}
	if session.State != StateParsing {
		t.Errorf("expected state PARSING, got %s", session.State)
	}

	// PARSING → PLANNING (via Confirm when intent is valid)
	err = session.Transition(SignalConfirm)
	if err != nil {
		t.Fatalf("PARSING → PLANNING failed: %v", err)
	}
	if session.State != StatePlanning {
		t.Errorf("expected state PLANNING, got %s", session.State)
	}

	// PLANNING → REVIEWING (via Confirm after plan review)
	err = session.Transition(SignalConfirm)
	if err != nil {
		t.Fatalf("PLANNING → REVIEWING failed: %v", err)
	}
	if session.State != StateReviewing {
		t.Errorf("expected state REVIEWING, got %s", session.State)
	}

	// REVIEWING → EXECUTING (via Confirm after reviewing changes)
	err = session.Transition(SignalConfirm)
	if err != nil {
		t.Fatalf("REVIEWING → EXECUTING failed: %v", err)
	}
	if session.State != StateExecuting {
		t.Errorf("expected state EXECUTING, got %s", session.State)
	}

	// Execute the plan
	plan := GeneratePlan(ParsedIntent{
		Action: ActionInspect,
		Target: ResourceTarget{
			Name:      "test-resource",
			Kind:      "ConfigMap",
			Namespace: "default",
		},
		RiskLevel: RiskLow,
	})

	err = Execute(session, plan)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// EXECUTING → COMPLETED (via Confirm after successful execution)
	err = session.Transition(SignalConfirm)
	if err != nil {
		t.Fatalf("EXECUTING → COMPLETED failed: %v", err)
	}
	if session.State != StateCompleted {
		t.Errorf("expected state COMPLETED, got %s", session.State)
	}
}

// TestIntegration_AbortFlow tests abort from various states:
// User sends abort signal → FAILED state
func TestIntegration_AbortFlow(t *testing.T) {
	testCases := []struct {
		name        string
		initialState State
		setup       func(*ChangeSession)
	}{
		{
			name:        "abort_from_parsing",
			initialState: StateParsing,
			setup:       func(s *ChangeSession) {},
		},
		{
			name:        "abort_from_clarifying",
			initialState: StateClarifying,
			setup:       func(s *ChangeSession) {},
		},
		{
			name:        "abort_from_planning",
			initialState: StatePlanning,
			setup:       func(s *ChangeSession) {},
		},
		{
			name:        "abort_from_reviewing",
			initialState: StateReviewing,
			setup:       func(s *ChangeSession) {},
		},
		{
			name:        "abort_from_executing",
			initialState: StateExecuting,
			setup:       func(s *ChangeSession) {},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			session := &ChangeSession{
				ID:    "test-session-abort",
				State: tc.initialState,
			}

			// Send abort signal
			err := session.Transition(SignalAbort)
			if err != nil {
				t.Fatalf("Abort transition failed: %v", err)
			}

			// Verify FAILED state
			if session.State != StateFailed {
				t.Errorf("expected state FAILED, got %s", session.State)
			}

			// Verify terminal state: no further transitions allowed
			err = session.Transition(SignalConfirm)
			if err == nil {
				t.Error("expected error when sending signal to terminal FAILED state")
			}
		})
	}
}

// TestIntegration_ModifyAndRetry tests modify path in review state:
// REVIEWING → PLANNING (modify) → REVIEWING (confirm) → EXECUTING
func TestIntegration_ModifyAndRetry(t *testing.T) {
	session := NewChangeSession("test-session-modify")
	if session.State != StateParsing {
		t.Errorf("expected initial state PARSING, got %s", session.State)
	}

	// Quick path to REVIEWING
	_ = session.Transition(SignalConfirm) // → PLANNING
	_ = session.Transition(SignalConfirm) // → REVIEWING

	if session.State != StateReviewing {
		t.Fatalf("expected state REVIEWING, got %s", session.State)
	}

	// REVIEWING → PLANNING (via Modify when changes are needed)
	err := session.Transition(SignalModify)
	if err != nil {
		t.Fatalf("REVIEWING → PLANNING (Modify) failed: %v", err)
	}
	if session.State != StatePlanning {
		t.Errorf("expected state PLANNING, got %s", session.State)
	}

	// PLANNING → REVIEWING (via Confirm after replanning)
	err = session.Transition(SignalConfirm)
	if err != nil {
		t.Fatalf("PLANNING → REVIEWING (Confirm) failed: %v", err)
	}
	if session.State != StateReviewing {
		t.Errorf("expected state REVIEWING, got %s", session.State)
	}

	// REVIEWING → EXECUTING (via Confirm)
	err = session.Transition(SignalConfirm)
	if err != nil {
		t.Fatalf("REVIEWING → EXECUTING failed: %v", err)
	}
	if session.State != StateExecuting {
		t.Errorf("expected state EXECUTING, got %s", session.State)
	}

	// Execute and complete
	plan := GeneratePlan(ParsedIntent{
		Action: ActionUpdate,
		Target: ResourceTarget{
			Name:      "test-resource",
			Kind:      "ConfigMap",
			Namespace: "default",
		},
		RiskLevel: RiskMedium,
	})

	err = Execute(session, plan)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	err = session.Transition(SignalConfirm) // → COMPLETED
	if err != nil {
		t.Fatalf("EXECUTING → COMPLETED failed: %v", err)
	}
	if session.State != StateCompleted {
		t.Errorf("expected state COMPLETED, got %s", session.State)
	}
}

// TestIntegration_RollbackFlow tests Execute → Rollback flow.
func TestIntegration_RollbackFlow(t *testing.T) {
	// Clear any existing snapshots
	ClearSnapshots()

	session := NewChangeSession("test-session-rollback")

	// Progress to EXECUTING state
	_ = session.Transition(SignalConfirm) // → PLANNING
	_ = session.Transition(SignalConfirm) // → REVIEWING
	_ = session.Transition(SignalConfirm) // → EXECUTING

	if session.State != StateExecuting {
		t.Fatalf("expected state EXECUTING, got %s", session.State)
	}

	// Create a snapshot before executing
	resourceID := ResourceID{
		Name:       "test-deployment",
		Kind:       "Deployment",
		Namespace:  "default",
		APIVersion: "apps/v1",
	}

	// Create a test object
	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name":      "test-deployment",
				"namespace": "default",
			},
			"spec": map[string]interface{}{
				"replicas": int64(3),
			},
		},
	}

	// Create snapshot for rollback
	snapshot, err := CreateSnapshot(session.ID, resourceID, obj)
	if err != nil {
		t.Fatalf("CreateSnapshot failed: %v", err)
	}
	if snapshot == nil {
		t.Fatal("expected non-nil snapshot")
	}

	// Execute the plan
	plan := GeneratePlan(ParsedIntent{
		Action: ActionUpdate,
		Target: ResourceTarget{
			Name:      "test-deployment",
			Kind:      "Deployment",
			Namespace: "default",
		},
		RiskLevel: RiskMedium,
	})

	err = Execute(session, plan)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	// Rollback the resource
	restoredObj, err := Rollback(session.ID, resourceID)
	if err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}
	if restoredObj == nil {
		t.Fatal("expected non-nil restored object")
	}

	// Verify the restored object has the snapshot's spec
	spec, found, err := unstructured.NestedMap(restoredObj.Object, "spec")
	if err != nil {
		t.Fatalf("failed to get spec from restored object: %v", err)
	}
	if !found {
		t.Fatal("spec not found in restored object")
	}
	if spec["replicas"] != int64(3) {
		t.Errorf("expected replicas 3, got %v", spec["replicas"])
	}

	// Cleanup
	DeleteSnapshotsForSession(session.ID)
}

// TestIntegration_MultiSession tests two sessions running concurrently.
func TestIntegration_MultiSession(t *testing.T) {
	// Clear any existing snapshots
	ClearSnapshots()

	var wg sync.WaitGroup
	errors := make(chan error, 2)

	// Session 1: Complete flow for UPDATE
	wg.Add(1)
	go func() {
		defer wg.Done()

		session1 := NewChangeSession("session-1-update")
		if session1.State != StateParsing {
			errors <- fmt.Errorf("session1: expected PARSING, got %s", session1.State)
			return
		}

		// Progress to EXECUTING
		if err := session1.Transition(SignalConfirm); err != nil {
			errors <- fmt.Errorf("session1: PARSING→PLANNING failed: %v", err)
			return
		}
		if session1.State != StatePlanning {
			errors <- fmt.Errorf("session1: expected PLANNING, got %s", session1.State)
			return
		}

		if err := session1.Transition(SignalConfirm); err != nil {
			errors <- fmt.Errorf("session1: PLANNING→REVIEWING failed: %v", err)
			return
		}
		if session1.State != StateReviewing {
			errors <- fmt.Errorf("session1: expected REVIEWING, got %s", session1.State)
			return
		}

		if err := session1.Transition(SignalConfirm); err != nil {
			errors <- fmt.Errorf("session1: REVIEWING→EXECUTING failed: %v", err)
			return
		}
		if session1.State != StateExecuting {
			errors <- fmt.Errorf("session1: expected EXECUTING, got %s", session1.State)
			return
		}

		// Execute plan
		plan1 := GeneratePlan(ParsedIntent{
			Action: ActionUpdate,
			Target: ResourceTarget{
				Name:      "deployment-1",
				Kind:      "Deployment",
				Namespace: "default",
			},
			RiskLevel: RiskMedium,
		})

		if err := Execute(session1, plan1); err != nil {
			errors <- fmt.Errorf("session1: Execute failed: %v", err)
			return
		}

		// Complete
		if err := session1.Transition(SignalConfirm); err != nil {
			errors <- fmt.Errorf("session1: EXECUTING→COMPLETED failed: %v", err)
			return
		}
		if session1.State != StateCompleted {
			errors <- fmt.Errorf("session1: expected COMPLETED, got %s", session1.State)
			return
		}
	}()

	// Session 2: INSPECT flow
	wg.Add(1)
	go func() {
		defer wg.Done()

		session2 := NewChangeSession("session-2-inspect")
		if session2.State != StateParsing {
			errors <- fmt.Errorf("session2: expected PARSING, got %s", session2.State)
			return
		}

		// Progress to EXECUTING
		if err := session2.Transition(SignalConfirm); err != nil {
			errors <- fmt.Errorf("session2: PARSING→PLANNING failed: %v", err)
			return
		}
		if session2.State != StatePlanning {
			errors <- fmt.Errorf("session2: expected PLANNING, got %s", session2.State)
			return
		}

		if err := session2.Transition(SignalConfirm); err != nil {
			errors <- fmt.Errorf("session2: PLANNING→REVIEWING failed: %v", err)
			return
		}
		if session2.State != StateReviewing {
			errors <- fmt.Errorf("session2: expected REVIEWING, got %s", session2.State)
			return
		}

		if err := session2.Transition(SignalConfirm); err != nil {
			errors <- fmt.Errorf("session2: REVIEWING→EXECUTING failed: %v", err)
			return
		}
		if session2.State != StateExecuting {
			errors <- fmt.Errorf("session2: expected EXECUTING, got %s", session2.State)
			return
		}

		// Execute plan
		plan2 := GeneratePlan(ParsedIntent{
			Action: ActionInspect,
			Target: ResourceTarget{
				Name:      "deployment-2",
				Kind:      "Deployment",
				Namespace: "default",
			},
			RiskLevel: RiskLow,
		})

		if err := Execute(session2, plan2); err != nil {
			errors <- fmt.Errorf("session2: Execute failed: %v", err)
			return
		}

		// Complete
		if err := session2.Transition(SignalConfirm); err != nil {
			errors <- fmt.Errorf("session2: EXECUTING→COMPLETED failed: %v", err)
			return
		}
		if session2.State != StateCompleted {
			errors <- fmt.Errorf("session2: expected COMPLETED, got %s", session2.State)
			return
		}
	}()

	// Wait for both sessions to complete
	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Error(err)
	}
}

// TestIntegration_ConcurrentAbort tests aborting one session while another runs.
func TestIntegration_ConcurrentAbort(t *testing.T) {
	// Clear any existing snapshots
	ClearSnapshots()

	var wg sync.WaitGroup
	errors := make(chan error, 2)

	// Session 1: Will be aborted
	wg.Add(1)
	go func() {
		defer wg.Done()

		session1 := NewChangeSession("session-to-abort")

		// Progress to PLANNING
		if err := session1.Transition(SignalConfirm); err != nil {
			errors <- fmt.Errorf("session1: PARSING→PLANNING failed: %v", err)
			return
		}
		if session1.State != StatePlanning {
			errors <- fmt.Errorf("session1: expected PLANNING, got %s", session1.State)
			return
		}

		// Abort from PLANNING
		if err := session1.Transition(SignalAbort); err != nil {
			errors <- fmt.Errorf("session1: PLANNING→FAILED failed: %v", err)
			return
		}
		if session1.State != StateFailed {
			errors <- fmt.Errorf("session1: expected FAILED, got %s", session1.State)
			return
		}
	}()

	// Session 2: Completes normally
	wg.Add(1)
	go func() {
		defer wg.Done()

		session2 := NewChangeSession("session-completes")

		// Progress to EXECUTING
		if err := session2.Transition(SignalConfirm); err != nil {
			errors <- fmt.Errorf("session2: PARSING→PLANNING failed: %v", err)
			return
		}
		if err := session2.Transition(SignalConfirm); err != nil {
			errors <- fmt.Errorf("session2: PLANNING→REVIEWING failed: %v", err)
			return
		}
		if err := session2.Transition(SignalConfirm); err != nil {
			errors <- fmt.Errorf("session2: REVIEWING→EXECUTING failed: %v", err)
			return
		}
		if session2.State != StateExecuting {
			errors <- fmt.Errorf("session2: expected EXECUTING, got %s", session2.State)
			return
		}

		// Execute and complete
		plan := GeneratePlan(ParsedIntent{
			Action: ActionInspect,
			Target: ResourceTarget{
				Name:      "test-resource",
				Kind:      "ConfigMap",
				Namespace: "default",
			},
			RiskLevel: RiskLow,
		})

		if err := Execute(session2, plan); err != nil {
			errors <- fmt.Errorf("session2: Execute failed: %v", err)
			return
		}

		if err := session2.Transition(SignalConfirm); err != nil {
			errors <- fmt.Errorf("session2: EXECUTING→COMPLETED failed: %v", err)
			return
		}
		if session2.State != StateCompleted {
			errors <- fmt.Errorf("session2: expected COMPLETED, got %s", session2.State)
			return
		}
	}()

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Error(err)
	}
}

// TestIntegration_ValidateIntentInFlow tests ValidateIntent integration with state machine.
func TestIntegration_ValidateIntentInFlow(t *testing.T) {
	// Test valid intent passes validation
	validIntent := &ParsedIntent{
		Action: ActionCreate,
		Target: ResourceTarget{
			Kind:      "Deployment",
			Name:      "my-app",
			Namespace: "default",
		},
		RiskLevel: RiskLow,
	}

	clarifyQ := ValidateIntent(validIntent)
	if clarifyQ != nil {
		t.Errorf("expected nil for valid intent, got %+v", clarifyQ)
	}

	// Test invalid intent returns clarification question
	invalidIntent := &ParsedIntent{
		Action: ActionDelete,
		Target: ResourceTarget{
			Kind: "Deployment",
			// Missing name for DELETE
		},
		RiskLevel: RiskHigh,
		// Missing reason for HIGH risk
	}

	clarifyQ = ValidateIntent(invalidIntent)
	if clarifyQ == nil {
		t.Fatal("expected ClarifyQuestion for invalid intent, got nil")
	}

	// Should get target.name first (before reason) since it's required for DELETE
	if clarifyQ.Field != "target.name" {
		t.Errorf("expected field 'target.name', got %s", clarifyQ.Field)
	}
}

// TestIntegration_PlanGenerationInFlow tests plan generation as part of flow.
func TestIntegration_PlanGenerationInFlow(t *testing.T) {
	session := NewChangeSession("test-session-plan-gen")

	// Progress to PLANNING state
	_ = session.Transition(SignalConfirm) // → PLANNING

	if session.State != StatePlanning {
		t.Fatalf("expected PLANNING, got %s", session.State)
	}

	// Generate plan in PLANNING state
	intent := ParsedIntent{
		Action: ActionUpdate,
		Target: ResourceTarget{
			Name:      "test-deployment",
			Kind:      "Deployment",
			Namespace: "default",
		},
		RiskLevel: RiskMedium,
		Reason:    "Scaling for traffic increase",
	}

	plan := GeneratePlan(intent)
	if plan == nil {
		t.Fatal("GeneratePlan returned nil")
	}

	// Verify plan structure
	if len(plan.Steps) == 0 {
		t.Error("plan should have steps")
	}

	if plan.RiskLevel != RiskMedium {
		t.Errorf("expected risk level MEDIUM, got %s", plan.RiskLevel)
	}

	// Verify rollback plan exists for UPDATE
	if plan.RollbackPlan == nil || len(plan.RollbackPlan) == 0 {
		t.Error("UPDATE should have rollback plan")
	}

	// Move to REVIEWING
	_ = session.Transition(SignalConfirm) // → REVIEWING

	if session.State != StateReviewing {
		t.Fatalf("expected REVIEWING, got %s", session.State)
	}
}

// TestIntegration_TerminalStateIsolation tests that terminal states are isolated.
func TestIntegration_TerminalStateIsolation(t *testing.T) {
	// Test COMPLETED state
	completedSession := &ChangeSession{
		ID:    "test-completed",
		State: StateCompleted,
	}

	// All signals should error
	for _, signal := range []SessionSignal{SignalConfirm, SignalAbort, SignalModify, SignalProceed} {
		err := completedSession.Transition(signal)
		if err == nil {
			t.Errorf("COMPLETED + %s: expected error, got nil", signal)
		}
		// State should remain COMPLETED
		if completedSession.State != StateCompleted {
			t.Errorf("COMPLETED + %s: expected state COMPLETED, got %s", signal, completedSession.State)
		}
	}

	// Test FAILED state
	failedSession := &ChangeSession{
		ID:    "test-failed",
		State: StateFailed,
	}

	// All signals should error
	for _, signal := range []SessionSignal{SignalConfirm, SignalAbort, SignalModify, SignalProceed} {
		err := failedSession.Transition(signal)
		if err == nil {
			t.Errorf("FAILED + %s: expected error, got nil", signal)
		}
		// State should remain FAILED
		if failedSession.State != StateFailed {
			t.Errorf("FAILED + %s: expected state FAILED, got %s", signal, failedSession.State)
		}
	}
}

// TestIntegration_SnapshotIsolation tests that snapshots are isolated per session.
func TestIntegration_SnapshotIsolation(t *testing.T) {
	ClearSnapshots()

	resourceID := ResourceID{
		Name:       "shared-resource",
		Kind:       "Deployment",
		Namespace:  "default",
		APIVersion: "apps/v1",
	}

	// Session 1 creates a snapshot
	session1 := NewChangeSession("session-1")
	obj1 := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"spec": map[string]interface{}{"replicas": int64(1)},
		},
	}
	snapshot1, err := CreateSnapshot(session1.ID, resourceID, obj1)
	if err != nil {
		t.Fatalf("CreateSnapshot for session1 failed: %v", err)
	}

	// Session 2 creates a snapshot for same resource
	session2 := NewChangeSession("session-2")
	obj2 := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"spec": map[string]interface{}{"replicas": int64(3)},
		},
	}
	snapshot2, err := CreateSnapshot(session2.ID, resourceID, obj2)
	if err != nil {
		t.Fatalf("CreateSnapshot for session2 failed: %v", err)
	}

	// Verify snapshots are different
	if snapshot1.ID == snapshot2.ID {
		t.Error("snapshots should have different IDs")
	}

	// Session 1 rolls back → should get session1's snapshot (replicas=1)
	restored1, err := Rollback(session1.ID, resourceID)
	if err != nil {
		t.Fatalf("Rollback for session1 failed: %v", err)
	}
	spec1, _, _ := unstructured.NestedMap(restored1.Object, "spec")
	if spec1["replicas"] != int64(1) {
		t.Errorf("session1: expected replicas 1, got %v", spec1["replicas"])
	}

	// Session 2 rolls back → should get session2's snapshot (replicas=3)
	restored2, err := Rollback(session2.ID, resourceID)
	if err != nil {
		t.Fatalf("Rollback for session2 failed: %v", err)
	}
	spec2, _, _ := unstructured.NestedMap(restored2.Object, "spec")
	if spec2["replicas"] != int64(3) {
		t.Errorf("session2: expected replicas 3, got %v", spec2["replicas"])
	}

	// Cleanup
	DeleteSnapshotsForSession(session1.ID)
	DeleteSnapshotsForSession(session2.ID)
}

// TestIntegration_ClarifyPath tests CLARIFYING → PARSING → CLARIFYING loop.
func TestIntegration_ClarifyPath(t *testing.T) {
	session := NewChangeSession("test-session-clarify")

	// PARSING → CLARIFYING (intent incomplete)
	err := session.Transition(SignalModify)
	if err != nil {
		t.Fatalf("PARSING → CLARIFYING failed: %v", err)
	}
	if session.State != StateClarifying {
		t.Errorf("expected CLARIFYING, got %s", session.State)
	}

	// Stay in CLARIFYING with Modify
	err = session.Transition(SignalModify)
	if err != nil {
		t.Fatalf("CLARIFYING (stay) with Modify failed: %v", err)
	}
	if session.State != StateClarifying {
		t.Errorf("expected CLARIFYING (stay), got %s", session.State)
	}

	// CLARIFYING → PARSING (intent clarified)
	err = session.Transition(SignalConfirm)
	if err != nil {
		t.Fatalf("CLARIFYING → PARSING failed: %v", err)
	}
	if session.State != StateParsing {
		t.Errorf("expected PARSING, got %s", session.State)
	}

	// Can go back to CLARIFYING again if needed
	err = session.Transition(SignalModify)
	if err != nil {
		t.Fatalf("PARSING → CLARIFYING (2nd pass) failed: %v", err)
	}
	if session.State != StateClarifying {
		t.Errorf("expected CLARIFYING (2nd pass), got %s", session.State)
	}
}