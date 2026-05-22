package core

import (
	"errors"
	"fmt"
)

// PreCheck represents a pre-execution check for a change operation.
type PreCheck struct {
	Name     string
	Run      func(session *ChangeSession, plan *ChangePlan) CheckResult
	Critical bool
}

// CheckResult represents the result of a pre-check execution.
type CheckResult struct {
	Passed  bool
	Message string
	Details string
}

// Execute runs pre-checks followed by step-by-step execution of a change plan.
// It returns an error if any critical pre-check fails or if step execution fails.
func Execute(session *ChangeSession, plan *ChangePlan) error {
	// Run pre-checks first
	if err := runPreChecks(session, plan); err != nil {
		return err
	}

	// Execute steps one by one
	if err := executeSteps(session, plan); err != nil {
		return err
	}

	return nil
}

// runPreChecks executes all pre-checks for the session and plan.
// It returns an error if any critical pre-check fails.
func runPreChecks(session *ChangeSession, plan *ChangePlan) error {
	for _, check := range preChecks {
		result := check.Run(session, plan)
		if !result.Passed && check.Critical {
			return fmt.Errorf("critical pre-check %q failed: %s", check.Name, result.Message)
		}
	}
	return nil
}

// executeSteps executes each step in the change plan sequentially.
func executeSteps(session *ChangeSession, plan *ChangePlan) error {
	for _, step := range plan.Steps {
		if err := executeStep(session, plan, &step); err != nil {
			return fmt.Errorf("step %d (%s) failed: %w", step.Seq, step.Action, err)
		}
	}
	return nil
}

// executeStep executes a single change step.
func executeStep(session *ChangeSession, plan *ChangePlan, step *ChangeStep) error {
	// Placeholder for actual step execution logic
	// In a real implementation, this would call K8s API based on step.Action
	switch step.Action {
	case ActionInspect:
		return executeInspectStep(session, plan, step)
	case ActionCreate:
		return executeCreateStep(session, plan, step)
	case ActionUpdate:
		return executeUpdateStep(session, plan, step)
	case ActionDelete:
		return executeDeleteStep(session, plan, step)
	default:
		return fmt.Errorf("unknown action: %s", step.Action)
	}
}

// executeInspectStep handles INSPECT action steps.
func executeInspectStep(session *ChangeSession, plan *ChangePlan, step *ChangeStep) error {
	// Placeholder for inspect logic
	return nil
}

// executeCreateStep handles CREATE action steps.
func executeCreateStep(session *ChangeSession, plan *ChangePlan, step *ChangeStep) error {
	// Placeholder for create logic
	return nil
}

// executeUpdateStep handles UPDATE action steps.
func executeUpdateStep(session *ChangeSession, plan *ChangePlan, step *ChangeStep) error {
	// Placeholder for update logic
	return nil
}

// executeDeleteStep handles DELETE action steps.
func executeDeleteStep(session *ChangeSession, plan *ChangePlan, step *ChangeStep) error {
	// Placeholder for delete logic
	return nil
}

// preChecks contains all pre-checks to be run before executing a change plan.
var preChecks = []PreCheck{
	{
		Name:     "resource_exists",
		Critical: false,
		Run:      checkResourceExists,
	},
	{
		Name:     "sufficient_quota",
		Critical: false,
		Run:      checkSufficientQuota,
	},
	{
		Name:     "no_conflicting_name",
		Critical: false,
		Run:      checkNoConflictingName,
	},
	{
		Name:     "backup_snapshot",
		Critical: false,
		Run:      checkBackupSnapshot,
	},
}

// checkResourceExists verifies that the target resource exists or does not exist
// depending on the action type (CREATE should not exist, others should exist).
func checkResourceExists(session *ChangeSession, plan *ChangePlan) CheckResult {
	// Stub implementation - returns Passed=true
	// In real implementation, this would query K8s API to check resource existence
	return CheckResult{
		Passed:  true,
		Message: "resource existence check passed",
		Details: "stub implementation - no K8s API call made",
	}
}

// checkSufficientQuota verifies that there is sufficient quota for the operation.
func checkSufficientQuota(session *ChangeSession, plan *ChangePlan) CheckResult {
	// Stub implementation - returns Passed=true
	// In real implementation, this would check namespace resource quota
	return CheckResult{
		Passed:  true,
		Message: "quota check passed",
		Details: "stub implementation - no K8s API call made",
	}
}

// checkNoConflictingName verifies that there is no naming conflict.
func checkNoConflictingName(session *ChangeSession, plan *ChangePlan) CheckResult {
	// Stub implementation - returns Passed=true
	// In real implementation, this would check for conflicting names
	return CheckResult{
		Passed:  true,
		Message: "no naming conflict detected",
		Details: "stub implementation - no K8s API call made",
	}
}

// checkBackupSnapshot verifies that backup snapshot is available for rollback.
func checkBackupSnapshot(session *ChangeSession, plan *ChangePlan) CheckResult {
	// Stub implementation - returns Passed=true
	// In real implementation, this would verify snapshots exist for rollback
	return CheckResult{
		Passed:  true,
		Message: "backup snapshot check passed",
		Details: "stub implementation - no snapshot verification made",
	}
}

// Pre-check errors.
var (
	ErrPreCheckFailed = errors.New("pre-check failed")
	ErrStepFailed     = errors.New("step execution failed")
)