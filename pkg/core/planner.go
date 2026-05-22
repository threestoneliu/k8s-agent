package core

import (
	"fmt"
	"time"
)

// ChangeStep represents a single step in a change plan.
type ChangeStep struct {
	Seq        int         `json:"seq"`
	Action     Action      `json:"action"`
	Target     ResourceTarget `json:"target"`
	RiskLevel  RiskLevel   `json:"riskLevel"`
	CanRollback bool       `json:"canRollback"`
	Validate   string      `json:"validate"`
	Description string     `json:"description"`
}

// ChangePlan represents a complete plan for executing a change operation.
type ChangePlan struct {
	ID            string      `json:"id"`
	Summary       string      `json:"summary"`
	Steps         []ChangeStep `json:"steps"`
	PreCheck      []string    `json:"preCheck"`
	RollbackPlan  []ChangeStep `json:"rollbackPlan"`
	RiskLevel     RiskLevel   `json:"riskLevel"`
	Impact        string      `json:"impact"`
	Duration      time.Duration `json:"duration"`
}

// ResourceDiff represents the difference between current and desired resource state.
type ResourceDiff struct {
	HasChanges  bool            `json:"hasChanges"`
	ChangedFields []string      `json:"changedFields"`
	OldValues  map[string]interface{} `json:"oldValues"`
	NewValues  map[string]interface{} `json:"newValues"`
}

// GeneratePlan creates a ChangePlan from a ParsedIntent.
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

// generatePlanID generates a unique plan ID.
func generatePlanID() string {
	return fmt.Sprintf("plan-%d", time.Now().UnixNano())
}

// generateSteps generates the steps for the plan based on the intent.
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

// generatePreChecks generates pre-check items for the plan.
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
func generateRollbackPlan(intent ParsedIntent, steps []ChangeStep) []ChangeStep {
	if intent.Action == ActionInspect {
		return nil // Inspect operations don't need rollback
	}

	if intent.Action == ActionDelete {
		return nil // Delete operations typically cannot be rolled back
	}

	// For CREATE: rollback means deleting what was created
	if intent.Action == ActionCreate {
		return []ChangeStep{
			{Seq: 1, Action: ActionDelete, Target: intent.Target, RiskLevel: RiskMedium, CanRollback: false,
				Validate: "confirm-deletion", Description: "Delete the created resource"},
		}
	}

	// For UPDATE: rollback means reverting to captured state
	// Note: In a real implementation, this would include the captured old values
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

	// Check for removed fields
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
func AssessRiskLevel(intent ParsedIntent) RiskLevel {
	risk := intent.RiskLevel
	if risk != "" {
		return risk
	}

	// Apply default risk levels based on action type
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

	// Increase risk for critical resources
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

	// Increase risk for cluster-scoped resources being modified
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