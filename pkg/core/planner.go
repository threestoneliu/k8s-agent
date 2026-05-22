package core

import (
	"fmt"
	"time"
)

// ChangeStep represents a single step in a change plan.
// Each step defines an action to perform on a target resource.
type ChangeStep struct {
	// Seq is the sequence number of this step in the plan (1-indexed).
	Seq int `json:"seq"`
	// Action is the type of operation to perform.
	Action Action `json:"action"`
	// Target identifies the resource this step operates on.
	Target ResourceTarget `json:"target"`
	// RiskLevel is the risk level of this specific step.
	RiskLevel RiskLevel `json:"riskLevel"`
	// CanRollback indicates whether this step can be rolled back.
	CanRollback bool `json:"canRollback"`
	// Validate is the validation check to run before executing.
	Validate string `json:"validate"`
	// Description is a human-readable description of this step.
	Description string `json:"description"`
}

// ChangePlan represents a complete plan for executing a change operation.
// It contains all steps, pre-checks, and rollback information needed
// to safely execute and potentially undo the change.
type ChangePlan struct {
	// ID uniquely identifies this plan.
	ID string `json:"id"`
	// Summary is a human-readable summary of the change.
	Summary string `json:"summary"`
	// Steps are the ordered list of steps to execute.
	Steps []ChangeStep `json:"steps"`
	// PreCheck lists the pre-execution validation checks to run.
	PreCheck []string `json:"preCheck"`
	// RollbackPlan describes how to undo the change if needed.
	RollbackPlan []ChangeStep `json:"rollbackPlan"`
	// RiskLevel is the overall risk level of the change.
	RiskLevel RiskLevel `json:"riskLevel"`
	// Impact describes the expected impact of the change.
	Impact string `json:"impact"`
	// Duration is the estimated time to complete the change.
	Duration time.Duration `json:"duration"`
}

// ResourceDiff represents the difference between current and desired resource state.
// It is used to display what changes will be made before execution.
type ResourceDiff struct {
	// HasChanges indicates whether any differences exist.
	HasChanges bool `json:"hasChanges"`
	// ChangedFields lists the fields that differ.
	ChangedFields []string `json:"changedFields"`
	// OldValues contains the current values of changed fields.
	OldValues map[string]interface{} `json:"oldValues"`
	// NewValues contains the desired values of changed fields.
	NewValues map[string]interface{} `json:"newValues"`
}

// GeneratePlan creates a ChangePlan from a ParsedIntent.
// The plan includes all steps, pre-checks, rollback plan, and risk assessment.
func GeneratePlan(intent ParsedIntent) *ChangePlan {
	plan := &ChangePlan{
		ID:        generatePlanID(),
		RiskLevel: intent.RiskLevel,
		Impact:    assessImpact(intent),
		Duration:  estimateDuration(intent),
	}

	plan.Steps = generateSteps(intent)
	plan.PreCheck = generatePreChecks(intent)
	plan.RollbackPlan = generateRollbackPlan(intent, plan.Steps)
	plan.Summary = generateSummary(intent, plan.Steps)

	return plan
}

// generatePlanID generates a unique plan ID based on timestamp.
func generatePlanID() string {
	return fmt.Sprintf("plan-%d", time.Now().UnixNano())
}

// generateSteps generates the execution steps for the plan based on the intent.
// Different actions (CREATE, UPDATE, DELETE, INSPECT) have different step sequences.
func generateSteps(intent ParsedIntent) []ChangeStep {
	var steps []ChangeStep

	switch intent.Action {
	case ActionCreate:
		steps = []ChangeStep{
			{Seq: 1, Action: ActionInspect, Target: intent.Target, RiskLevel: RiskLow, CanRollback: false,
				Validate: "check-resource-not-exists", Description: "Check target resource does not already exist"},
			{Seq: 2, Action: ActionCreate, Target: intent.Target, RiskLevel: intent.RiskLevel, CanRollback: true,
				Validate: "validate-spec", Description: "Create the resource with validated spec"},
		}
	case ActionUpdate:
		steps = []ChangeStep{
			{Seq: 1, Action: ActionInspect, Target: intent.Target, RiskLevel: RiskLow, CanRollback: false,
				Validate: "check-resource-exists", Description: "Verify target resource exists"},
			{Seq: 2, Action: ActionInspect, Target: intent.Target, RiskLevel: RiskLow, CanRollback: false,
				Validate: "capture-current-state", Description: "Capture current resource state for rollback"},
			{Seq: 3, Action: ActionUpdate, Target: intent.Target, RiskLevel: intent.RiskLevel, CanRollback: true,
				Validate: "validate-spec", Description: "Apply updates to the resource"},
		}
	case ActionDelete:
		steps = []ChangeStep{
			{Seq: 1, Action: ActionInspect, Target: intent.Target, RiskLevel: RiskLow, CanRollback: false,
				Validate: "check-resource-exists", Description: "Verify target resource exists"},
			{Seq: 2, Action: ActionInspect, Target: intent.Target, RiskLevel: RiskLow, CanRollback: false,
				Validate: "capture-current-state", Description: "Capture current resource state for rollback"},
			{Seq: 3, Action: ActionDelete, Target: intent.Target, RiskLevel: intent.RiskLevel, CanRollback: false,
				Validate: "confirm-deletion", Description: "Delete the resource"},
		}
	case ActionInspect:
		steps = []ChangeStep{
			{Seq: 1, Action: ActionInspect, Target: intent.Target, RiskLevel: RiskLow, CanRollback: false,
				Validate: "check-resource-exists", Description: "Verify target resource exists"},
			{Seq: 2, Action: ActionInspect, Target: intent.Target, RiskLevel: RiskLow, CanRollback: false,
				Validate: "retrieve-resource", Description: "Retrieve and display resource details"},
		}
	}

	return steps
}

// generatePreChecks generates pre-check items for the plan based on action type.
func generatePreChecks(intent ParsedIntent) []string {
	checks := []string{
		"validate-kubernetes-connection",
		"check-permissions",
	}

	switch intent.Action {
	case ActionCreate:
		checks = append(checks, "verify-namespace-exists")
	case ActionUpdate:
		checks = append(checks, "verify-resource-exists", "verify-lock-status")
	case ActionDelete:
		checks = append(checks, "verify-resource-exists", "confirm-no-dependents")
	}

	return checks
}

// generateRollbackPlan generates the rollback plan based on the intent and steps.
// INSPECT operations don't need rollback. DELETE operations cannot typically be rolled back.
func generateRollbackPlan(intent ParsedIntent, steps []ChangeStep) []ChangeStep {
	if intent.Action == ActionInspect {
		return nil
	}

	if intent.Action == ActionDelete {
		return nil
	}

	if intent.Action == ActionCreate {
		return []ChangeStep{
			{Seq: 1, Action: ActionDelete, Target: intent.Target, RiskLevel: RiskMedium, CanRollback: false,
				Validate: "confirm-deletion", Description: "Delete the created resource"},
		}
	}

	return []ChangeStep{
		{Seq: 1, Action: ActionUpdate, Target: intent.Target, RiskLevel: RiskMedium, CanRollback: false,
			Validate: "revert-changes", Description: "Revert to previous resource state"},
	}
}

// generateSummary generates a human-readable summary for the plan.
func generateSummary(intent ParsedIntent, steps []ChangeStep) string {
	targetDesc := fmt.Sprintf("%s %s", intent.Target.Kind, intent.Target.Name)
	if intent.Target.Namespace != "" {
		targetDesc = fmt.Sprintf("%s %s in namespace %s", intent.Target.Kind, intent.Target.Name, intent.Target.Namespace)
	}

	return fmt.Sprintf("%s %s (Risk: %s)", intent.Action, targetDesc, intent.RiskLevel)
}

// assessImpact assesses the impact of the operation based on the intent.
func assessImpact(intent ParsedIntent) string {
	switch intent.Action {
	case ActionCreate:
		return "Creates a new " + intent.Target.Kind + " resource"
	case ActionUpdate:
		return "Modifies existing " + intent.Target.Kind + " resource"
	case ActionDelete:
		return "Permanently removes " + intent.Target.Kind + " resource"
	case ActionInspect:
		return "Reads " + intent.Target.Kind + " resource information"
	default:
		return "Unknown impact"
	}
}

// estimateDuration estimates the expected duration for the operation.
func estimateDuration(intent ParsedIntent) time.Duration {
	switch intent.Action {
	case ActionCreate:
		return 30 * time.Second
	case ActionUpdate:
		return 20 * time.Second
	case ActionDelete:
		return 15 * time.Second
	case ActionInspect:
		return 5 * time.Second
	default:
		return 10 * time.Second
	}
}

// CalculateResourceDiff calculates the difference between current and desired state.
func CalculateResourceDiff(current, desired map[string]interface{}) ResourceDiff {
	diff := ResourceDiff{
		HasChanges:   false,
		ChangedFields: []string{},
		OldValues:   make(map[string]interface{}),
		NewValues:   make(map[string]interface{}),
	}

	for key, newVal := range desired {
		oldVal, exists := current[key]
		if !exists || !valuesEqual(oldVal, newVal) {
			diff.HasChanges = true
			diff.ChangedFields = append(diff.ChangedFields, key)
			if exists {
				diff.OldValues[key] = oldVal
			}
			diff.NewValues[key] = newVal
		}
	}

	for key, oldVal := range current {
		if _, exists := desired[key]; !exists {
			diff.HasChanges = true
			diff.ChangedFields = append(diff.ChangedFields, key)
			diff.OldValues[key] = oldVal
		}
	}

	return diff
}

// valuesEqual compares two values for equality.
func valuesEqual(a, b interface{}) bool {
	switch a.(type) {
	case string:
		if bStr, ok := b.(string); ok {
			return a.(string) == bStr
		}
	case int:
		if bInt, ok := b.(int); ok {
			return a.(int) == bInt
		}
	case bool:
		if bBool, ok := b.(bool); ok {
			return a.(bool) == bBool
		}
	case map[string]interface{}:
		if bMap, ok := b.(map[string]interface{}); ok {
			return mapsEqual(a.(map[string]interface{}), bMap)
		}
	}
	return false
}

// mapsEqual compares two maps for equality.
func mapsEqual(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for key, valA := range a {
		valB, exists := b[key]
		if !exists || !valuesEqual(valA, valB) {
			return false
		}
	}
	return true
}

// AssessRiskLevel determines the risk level based on action and target.
// It increases risk for critical resources (Node, PV, etc.) and cluster-scoped resources.
func AssessRiskLevel(intent ParsedIntent) RiskLevel {
	risk := intent.RiskLevel
	if risk != "" {
		return risk
	}

	switch intent.Action {
	case ActionCreate:
		risk = RiskLow
	case ActionUpdate:
		risk = RiskMedium
	case ActionDelete:
		risk = RiskHigh
	case ActionInspect:
		risk = RiskLow
	}

	criticalKinds := map[string]bool{
		"Node":                   true,
		"PersistentVolume":       true,
		"ClusterRole":           true,
		"ClusterRoleBinding":     true,
		"Namespace":              true,
		"StorageClass":           true,
		"VolumeAttachment":       true,
	}

	if criticalKinds[intent.Target.Kind] {
		switch risk {
		case RiskLow:
			risk = RiskMedium
		case RiskMedium:
			risk = RiskHigh
		case RiskHigh:
			risk = RiskCritical
		}
	}

	if !IsNamespacedKind(intent.Target.Kind) && intent.Action != ActionInspect {
		switch risk {
		case RiskLow:
			risk = RiskMedium
		case RiskMedium:
			risk = RiskHigh
		}
	}

	return risk
}