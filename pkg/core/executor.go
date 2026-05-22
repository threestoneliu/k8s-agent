package core

import (
	"errors"
	"fmt"
)

// PreCheck represents a pre-execution check that runs before a change plan is executed.
// Pre-checks validate conditions like resource existence, quota, and permissions.
type PreCheck struct {
	// Name identifies the pre-check.
	Name string
	// Run is the function that performs the check.
	Run func(session *ChangeSession, plan *ChangePlan) CheckResult
	// Critical indicates whether failure of this check should abort execution.
	Critical bool
}

// CheckResult represents the result of a pre-check execution.
type CheckResult struct {
	// Passed indicates whether the check passed.
	Passed bool
	// Message is a human-readable description of the result.
	Message string
	// Details provides additional context about the check result.
	Details string
}

// Execute runs pre-checks followed by step-by-step execution of a change plan.
// It logs all operations to the audit log and returns an error if any critical
// pre-check fails or if step execution fails.
// All steps are executed sequentially and each step's result is logged.
func Execute(session *ChangeSession, plan *ChangePlan) error {
	Log(session.ID, "execute_start", "executor", map[string]interface{}{
		"plan_id":    plan.ID,
		"risk_level": plan.RiskLevel,
		"step_count": len(plan.Steps),
	})

	if err := runPreChecks(session, plan); err != nil {
		Log(session.ID, "execute_precheck_failed", "executor", map[string]interface{}{
			"plan_id": plan.ID,
			"error":   err.Error(),
		})
		return err
	}

	Log(session.ID, "execute_prechecks_passed", "executor", map[string]interface{}{
		"plan_id": plan.ID,
	})

	if err := executeSteps(session, plan); err != nil {
		Log(session.ID, "execute_steps_failed", "executor", map[string]interface{}{
			"plan_id": plan.ID,
			"error":   err.Error(),
		})
		return err
	}

	Log(session.ID, "execute_completed", "executor", map[string]interface{}{
		"plan_id": plan.ID,
	})

	return nil
}

// runPreChecks executes all pre-checks for the session and plan.
// Critical checks that fail will cause execution to abort.
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
// If any step fails, execution stops and returns an error.
func executeSteps(session *ChangeSession, plan *ChangePlan) error {
	for _, step := range plan.Steps {
		if err := executeStep(session, plan, &step); err != nil {
			return fmt.Errorf("step %d (%s) failed: %w", step.Seq, step.Action, err)
		}
	}
	return nil
}

// executeStep executes a single change step.
// It logs the step start, executes the appropriate action handler, and logs the result.
func executeStep(session *ChangeSession, plan *ChangePlan, step *ChangeStep) error {
	targetStr := step.Target.Kind + "/" + step.Target.Name
	if step.Target.Namespace != "" {
		targetStr = step.Target.Namespace + "/" + targetStr
	}
	Log(session.ID, "step_start", "executor", map[string]interface{}{
		"plan_id":     plan.ID,
		"step_seq":    step.Seq,
		"step_action": step.Action,
		"target":      targetStr,
		"risk_level":  step.RiskLevel,
	})

	var err error
	switch step.Action {
	case ActionInspect:
		err = executeInspectStep(session, plan, step)
	case ActionCreate:
		err = executeCreateStep(session, plan, step)
	case ActionUpdate:
		err = executeUpdateStep(session, plan, step)
	case ActionDelete:
		err = executeDeleteStep(session, plan, step)
	default:
		err = fmt.Errorf("unknown action: %s", step.Action)
	}

	if err != nil {
		Log(session.ID, "step_failed", "executor", map[string]interface{}{
			"plan_id":     plan.ID,
			"step_seq":    step.Seq,
			"step_action": step.Action,
			"error":       err.Error(),
		})
	} else {
		Log(session.ID, "step_completed", "executor", map[string]interface{}{
			"plan_id":     plan.ID,
			"step_seq":    step.Seq,
			"step_action": step.Action,
		})
	}

	return err
}

// executeInspectStep handles INSPECT action steps.
// Currently a placeholder for actual K8s API inspection logic.
func executeInspectStep(session *ChangeSession, plan *ChangePlan, step *ChangeStep) error {
	return nil
}

// executeCreateStep handles CREATE action steps.
// Currently a placeholder for actual K8s API creation logic.
func executeCreateStep(session *ChangeSession, plan *ChangePlan, step *ChangeStep) error {
	return nil
}

// executeUpdateStep handles UPDATE action steps.
// Currently a placeholder for actual K8s API update logic.
func executeUpdateStep(session *ChangeSession, plan *ChangePlan, step *ChangeStep) error {
	return nil
}

// executeDeleteStep handles DELETE action steps.
// Currently a placeholder for actual K8s API deletion logic.
func executeDeleteStep(session *ChangeSession, plan *ChangePlan, step *ChangeStep) error {
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

// checkResourceExists verifies that the target resource exists or does not exist.
// For CREATE actions, the resource should not exist. For other actions, it should exist.
func checkResourceExists(session *ChangeSession, plan *ChangePlan) CheckResult {
	return CheckResult{
		Passed:  true,
		Message: "resource existence check passed",
		Details: "stub implementation - no K8s API call made",
	}
}

// checkSufficientQuota verifies that there is sufficient quota for the operation.
func checkSufficientQuota(session *ChangeSession, plan *ChangePlan) CheckResult {
	return CheckResult{
		Passed:  true,
		Message: "quota check passed",
		Details: "stub implementation - no K8s API call made",
	}
}

// checkNoConflictingName verifies that there is no naming conflict.
func checkNoConflictingName(session *ChangeSession, plan *ChangePlan) CheckResult {
	return CheckResult{
		Passed:  true,
		Message: "no naming conflict detected",
		Details: "stub implementation - no K8s API call made",
	}
}

// checkBackupSnapshot verifies that backup snapshot is available for rollback.
func checkBackupSnapshot(session *ChangeSession, plan *ChangePlan) CheckResult {
	return CheckResult{
		Passed:  true,
		Message: "backup snapshot check passed",
		Details: "stub implementation - no snapshot verification made",
	}
}

// Pre-check errors.
var (
	// ErrPreCheckFailed is returned when a pre-check fails.
	ErrPreCheckFailed = errors.New("pre-check failed")
	// ErrStepFailed is returned when step execution fails.
	ErrStepFailed = errors.New("step execution failed")
)