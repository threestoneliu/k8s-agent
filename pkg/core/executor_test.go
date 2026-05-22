package core

import (
	"testing"
	"time"
)

// TestPreCheck tests the PreCheck struct and CheckResult.
func TestPreCheck(t *testing.T) {
	plan := &ChangePlan{
		ID: "test-plan-1",
		Summary: "Test plan",
		Steps: []ChangeStep{
			{
				Seq:        1,
				Action:     ActionInspect,
				RiskLevel:  RiskLow,
				CanRollback: false,
				Description: "Test step",
			},
		},
		PreCheck: []string{"resource_exists"},
	}

	session := &ChangeSession{
		ID:    "test-session-1",
		State: StateExecuting,
	}

	// Test pre-check with stub implementation
	check := PreCheck{
		Name:     "test_check",
		Critical: true,
		Run: func(s *ChangeSession, p *ChangePlan) CheckResult {
			return CheckResult{
				Passed:  true,
				Message: "test passed",
				Details: "details here",
			}
		},
	}

	result := check.Run(session, plan)
	if !result.Passed {
		t.Errorf("expected CheckResult.Passed to be true")
	}
	if result.Message != "test passed" {
		t.Errorf("expected message 'test passed', got %q", result.Message)
	}
}

// TestCheckResult tests the CheckResult struct.
func TestCheckResult(test *testing.T) {
	result := CheckResult{
		Passed:  true,
		Message: "check passed",
		Details: "all good",
	}

	if !result.Passed {
		test.Error("expected Passed to be true")
	}
	if result.Message != "check passed" {
		test.Errorf("expected Message 'check passed', got %q", result.Message)
	}
	if result.Details != "all good" {
		test.Errorf("expected Details 'all good', got %q", result.Details)
	}
}

// TestExecute tests the Execute function.
func TestExecute(test *testing.T) {
	testPlan := &ChangePlan{
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
				Description: "Inspect test resource",
			},
		},
		PreCheck: []string{"resource_exists"},
	}

	session := &ChangeSession{
		ID:    "test-session-1",
		State: StateExecuting,
	}

	err := Execute(session, testPlan)
	if err != nil {
		test.Errorf("Execute returned unexpected error: %v", err)
	}
}

// TestExecuteWithAllStepTypes tests Execute with different action types.
func TestExecuteWithAllStepTypes(t *testing.T) {
	testCases := []struct {
		name   string
		action Action
	}{
		{"inspect_action", ActionInspect},
		{"create_action", ActionCreate},
		{"update_action", ActionUpdate},
		{"delete_action", ActionDelete},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			plan := &ChangePlan{
				ID:        "test-plan-" + string(tc.action),
				Summary:   "Test plan for " + string(tc.action),
				RiskLevel: RiskLow,
				Steps: []ChangeStep{
					{
						Seq:        1,
						Action:     tc.action,
						Target:     ResourceTarget{Name: "test-resource", Kind: "ConfigMap", Namespace: "default"},
						RiskLevel:  RiskLow,
						CanRollback: tc.action != ActionDelete,
						Description: "Test step",
					},
				},
			}

			session := &ChangeSession{
				ID:    "test-session-" + string(tc.action),
				State: StateExecuting,
			}

			err := Execute(session, plan)
			if err != nil {
				t.Errorf("Execute returned unexpected error for action %s: %v", tc.action, err)
			}
		})
	}
}

// TestExecuteWithMultipleSteps tests Execute with multiple steps.
func TestExecuteWithMultipleSteps(test *testing.T) {
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

	session := &ChangeSession{
		ID:    "test-session-multi",
		State: StateExecuting,
	}

	err := Execute(session, plan)
	if err != nil {
		test.Errorf("Execute returned unexpected error: %v", err)
	}
}

// TestPreChecksList verifies that all expected pre-checks are defined.
func TestPreChecksList(test *testing.T) {
	expectedChecks := []string{
		"resource_exists",
		"sufficient_quota",
		"no_conflicting_name",
		"backup_snapshot",
	}

	if len(preChecks) != len(expectedChecks) {
		test.Errorf("expected %d pre-checks, got %d", len(expectedChecks), len(preChecks))
	}

	definedChecks := make(map[string]bool)
	for _, check := range preChecks {
		definedChecks[check.Name] = true
	}

	for _, expected := range expectedChecks {
		if !definedChecks[expected] {
			test.Errorf("expected pre-check %q not found", expected)
		}
	}
}

// TestCheckResourceExists verifies the resource_exists pre-check.
func TestCheckResourceExists(test *testing.T) {
	session := &ChangeSession{
		ID:    "test-session",
		State: StateExecuting,
	}
	plan := &ChangePlan{
		ID: "test-plan",
	}

	// Find resource_exists check
	var checkResourceExistsFn func(*ChangeSession, *ChangePlan) CheckResult
	for _, check := range preChecks {
		if check.Name == "resource_exists" {
			checkResourceExistsFn = check.Run
			break
		}
	}

	if checkResourceExistsFn == nil {
		test.Fatal("resource_exists pre-check not found")
	}

	result := checkResourceExistsFn(session, plan)
	if !result.Passed {
		test.Error("resource_exists check should pass (stub)")
	}
}

// TestCheckSufficientQuota verifies the sufficient_quota pre-check.
func TestCheckSufficientQuota(test *testing.T) {
	session := &ChangeSession{
		ID:    "test-session",
		State: StateExecuting,
	}
	plan := &ChangePlan{
		ID: "test-plan",
	}

	// Find sufficient_quota check
	var checkSufficientQuotaFn func(*ChangeSession, *ChangePlan) CheckResult
	for _, check := range preChecks {
		if check.Name == "sufficient_quota" {
			checkSufficientQuotaFn = check.Run
			break
		}
	}

	if checkSufficientQuotaFn == nil {
		test.Fatal("sufficient_quota pre-check not found")
	}

	result := checkSufficientQuotaFn(session, plan)
	if !result.Passed {
		test.Error("sufficient_quota check should pass (stub)")
	}
}

// TestCheckNoConflictingName verifies the no_conflicting_name pre-check.
func TestCheckNoConflictingName(test *testing.T) {
	session := &ChangeSession{
		ID:    "test-session",
		State: StateExecuting,
	}
	plan := &ChangePlan{
		ID: "test-plan",
	}

	// Find no_conflicting_name check
	var checkNoConflictingNameFn func(*ChangeSession, *ChangePlan) CheckResult
	for _, check := range preChecks {
		if check.Name == "no_conflicting_name" {
			checkNoConflictingNameFn = check.Run
			break
		}
	}

	if checkNoConflictingNameFn == nil {
		test.Fatal("no_conflicting_name pre-check not found")
	}

	result := checkNoConflictingNameFn(session, plan)
	if !result.Passed {
		test.Error("no_conflicting_name check should pass (stub)")
	}
}

// TestCheckBackupSnapshot verifies the backup_snapshot pre-check.
func TestCheckBackupSnapshot(test *testing.T) {
	session := &ChangeSession{
		ID:    "test-session",
		State: StateExecuting,
	}
	plan := &ChangePlan{
		ID: "test-plan",
	}

	// Find backup_snapshot check
	var checkBackupSnapshotFn func(*ChangeSession, *ChangePlan) CheckResult
	for _, check := range preChecks {
		if check.Name == "backup_snapshot" {
			checkBackupSnapshotFn = check.Run
			break
		}
	}

	if checkBackupSnapshotFn == nil {
		test.Fatal("backup_snapshot pre-check not found")
	}

	result := checkBackupSnapshotFn(session, plan)
	if !result.Passed {
		test.Error("backup_snapshot check should pass (stub)")
	}
}

// TestRunPreChecksCriticalFailure tests that runPreChecks returns error on critical failure.
func TestRunPreChecksCriticalFailure(test *testing.T) {
	// Create a temporary pre-check list with a critical failing check
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

	session := &ChangeSession{
		ID:    "test-session",
		State: StateExecuting,
	}
	plan := &ChangePlan{
		ID: "test-plan",
	}

	// Temporarily replace preChecks
	originalChecks := preChecks
	preChecks = tempChecks
	defer func() { preChecks = originalChecks }()

	err := runPreChecks(session, plan)
	if err == nil {
		test.Error("expected error for critical pre-check failure")
	}
}

// TestRunPreChecksNonCriticalFailure tests that runPreChecks continues on non-critical failure.
func TestRunPreChecksNonCriticalFailure(test *testing.T) {
	// Create a temporary pre-check list with a non-critical failing check
	tempChecks := []PreCheck{
		{
			Name:     "non_critical_failing",
			Critical: false,
			Run: func(s *ChangeSession, p *ChangePlan) CheckResult {
				return CheckResult{
					Passed:  false,
					Message: "non-critical failure",
					Details: "this check can fail",
				}
			},
		},
	}

	session := &ChangeSession{
		ID:    "test-session",
		State: StateExecuting,
	}
	plan := &ChangePlan{
		ID: "test-plan",
	}

	// Temporarily replace preChecks
	originalChecks := preChecks
	preChecks = tempChecks
	defer func() { preChecks = originalChecks }()

	err := runPreChecks(session, plan)
	if err != nil {
		test.Errorf("expected no error for non-critical pre-check failure, got: %v", err)
	}
}

// TestExecuteStepsEmptyPlan tests Execute with an empty plan.
func TestExecuteStepsEmptyPlan(test *testing.T) {
	plan := &ChangePlan{
		ID:        "empty-plan",
		Summary:   "Empty plan",
		RiskLevel: RiskLow,
		Steps:     []ChangeStep{},
	}

	session := &ChangeSession{
		ID:    "test-session",
		State: StateExecuting,
	}

	err := Execute(session, plan)
	if err != nil {
		test.Errorf("Execute with empty plan should not error: %v", err)
	}
}

// TestExecuteWithRiskLevels tests Execute with different risk levels.
func TestExecuteWithRiskLevels(t *testing.T) {
	riskLevels := []RiskLevel{RiskLow, RiskMedium, RiskHigh, RiskCritical}

	for _, risk := range riskLevels {
		t.Run(string(risk), func(t *testing.T) {
			plan := &ChangePlan{
				ID:        "test-plan-" + string(risk),
				Summary:   "Test plan",
				RiskLevel: risk,
				Steps: []ChangeStep{
					{
						Seq:        1,
						Action:     ActionInspect,
						Target:     ResourceTarget{Name: "test-resource", Kind: "ConfigMap", Namespace: "default"},
						RiskLevel:  risk,
						CanRollback: false,
						Description: "Test step",
					},
				},
			}

			session := &ChangeSession{
				ID:    "test-session-" + string(risk),
				State: StateExecuting,
			}

			err := Execute(session, plan)
			if err != nil {
				t.Errorf("Execute returned unexpected error for risk %s: %v", risk, err)
			}
		})
	}
}

// TestExecuteWithPlanMetadata tests that Execute handles plan metadata correctly.
func TestExecuteWithPlanMetadata(test *testing.T) {
	plan := &ChangePlan{
		ID:            "test-plan-meta",
		Summary:       "Test plan with metadata",
		RiskLevel:     RiskMedium,
		Impact:        "Modifies ConfigMap resource",
		Duration:      30 * time.Second,
		RollbackPlan:  []ChangeStep{},
		Steps: []ChangeStep{
			{
				Seq:        1,
				Action:     ActionUpdate,
				Target:     ResourceTarget{Name: "test-configmap", Kind: "ConfigMap", Namespace: "default"},
				RiskLevel:  RiskMedium,
				CanRollback: true,
				Validate:   "validate-spec",
				Description: "Update ConfigMap",
			},
		},
		PreCheck: []string{"resource_exists", "sufficient_quota"},
	}

	session := &ChangeSession{
		ID:    "test-session-meta",
		State: StateExecuting,
	}

	err := Execute(session, plan)
	if err != nil {
		test.Errorf("Execute returned unexpected error: %v", err)
	}

	if session.State != StateExecuting {
		test.Errorf("session state should remain Executing, got %s", session.State)
	}
}

// TestExecuteStepUnknownAction tests executeStep with an unknown action.
func TestExecuteStepUnknownAction(test *testing.T) {
	plan := &ChangePlan{
		ID: "test-plan-unknown",
	}
	step := &ChangeStep{
		Seq:    1,
		Action: "UNKNOWN_ACTION",
	}
	session := &ChangeSession{
		ID:    "test-session",
		State: StateExecuting,
	}

	err := executeStep(session, plan, step)
	if err == nil {
		test.Error("expected error for unknown action")
	}
}