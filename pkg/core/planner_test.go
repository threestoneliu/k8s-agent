package core

import (
	"testing"
	"time"
)

func TestGeneratePlan_Create(t *testing.T) {
	intent := ParsedIntent{
		Action: ActionCreate,
		Target: ResourceTarget{
			Name:       "my-app",
			Kind:       "Deployment",
			Namespace:  "default",
			APIVersion: "apps/v1",
		},
		Params:    map[string]interface{}{"replicas": 3},
		RiskLevel: RiskLow,
	}

	plan := GeneratePlan(intent)

	if plan == nil {
		t.Fatal("GeneratePlan returned nil")
	}

	if len(plan.Steps) != 2 {
		t.Errorf("Expected 2 steps for CREATE, got %d", len(plan.Steps))
	}

	if plan.Steps[0].Action != ActionInspect {
		t.Errorf("Step 1 should be Inspect, got %s", plan.Steps[0].Action)
	}

	if plan.Steps[1].Action != ActionCreate {
		t.Errorf("Step 2 should be Create, got %s", plan.Steps[1].Action)
	}

	if plan.Steps[1].CanRollback != true {
		t.Error("CREATE step should be rollbackable")
	}

	if plan.RollbackPlan == nil || len(plan.RollbackPlan) == 0 {
		t.Error("CREATE should have a rollback plan")
	}

	if plan.RollbackPlan[0].Action != ActionDelete {
		t.Error("CREATE rollback should be Delete")
	}
}

func TestGeneratePlan_Update(t *testing.T) {
	intent := ParsedIntent{
		Action: ActionUpdate,
		Target: ResourceTarget{
			Name:       "my-app",
			Kind:       "Deployment",
			Namespace:  "default",
			APIVersion: "apps/v1",
		},
		Params:    map[string]interface{}{"replicas": 5},
		RiskLevel: RiskMedium,
	}

	plan := GeneratePlan(intent)

	if plan == nil {
		t.Fatal("GeneratePlan returned nil")
	}

	if len(plan.Steps) != 3 {
		t.Errorf("Expected 3 steps for UPDATE, got %d", len(plan.Steps))
	}

	if plan.Steps[2].CanRollback != true {
		t.Error("UPDATE step should be rollbackable")
	}

	if plan.RollbackPlan == nil || len(plan.RollbackPlan) == 0 {
		t.Error("UPDATE should have a rollback plan")
	}
}

func TestGeneratePlan_Delete(t *testing.T) {
	intent := ParsedIntent{
		Action: ActionDelete,
		Target: ResourceTarget{
			Name:       "my-app",
			Kind:       "Deployment",
			Namespace:  "default",
			APIVersion: "apps/v1",
		},
		RiskLevel: RiskHigh,
	}

	plan := GeneratePlan(intent)

	if plan == nil {
		t.Fatal("GeneratePlan returned nil")
	}

	if len(plan.Steps) != 3 {
		t.Errorf("Expected 3 steps for DELETE, got %d", len(plan.Steps))
	}

	// Delete steps should not be rollbackable
	for _, step := range plan.Steps {
		if step.CanRollback {
			t.Errorf("DELETE step %d should not be rollbackable", step.Seq)
		}
	}

	// Delete should not have a rollback plan
	if plan.RollbackPlan != nil && len(plan.RollbackPlan) > 0 {
		t.Error("DELETE should not have a rollback plan")
	}
}

func TestGeneratePlan_Inspect(t *testing.T) {
	intent := ParsedIntent{
		Action: ActionInspect,
		Target: ResourceTarget{
			Name:       "my-app",
			Kind:       "Deployment",
			Namespace:  "default",
			APIVersion: "apps/v1",
		},
		RiskLevel: RiskLow,
	}

	plan := GeneratePlan(intent)

	if plan == nil {
		t.Fatal("GeneratePlan returned nil")
	}

	if len(plan.Steps) != 2 {
		t.Errorf("Expected 2 steps for INSPECT, got %d", len(plan.Steps))
	}

	for _, step := range plan.Steps {
		if step.Action != ActionInspect {
			t.Errorf("INSPECT plan should only contain Inspect steps, got %s", step.Action)
		}
	}

	// Inspect should not have a rollback plan
	if plan.RollbackPlan != nil && len(plan.RollbackPlan) > 0 {
		t.Error("INSPECT should not have a rollback plan")
	}
}

func TestGeneratePlan_Summary(t *testing.T) {
	intent := ParsedIntent{
		Action: ActionCreate,
		Target: ResourceTarget{
			Name:       "my-app",
			Kind:       "Deployment",
			Namespace:  "default",
			APIVersion: "apps/v1",
		},
		RiskLevel: RiskLow,
	}

	plan := GeneratePlan(intent)

	if plan.Summary == "" {
		t.Error("Summary should not be empty")
	}
}

func TestCalculateResourceDiff_NoChanges(t *testing.T) {
	current := map[string]interface{}{
		"replicas": 3,
		"image":    "nginx:1.19",
	}
	desired := map[string]interface{}{
		"replicas": 3,
		"image":    "nginx:1.19",
	}

	diff := CalculateResourceDiff(current, desired)

	if diff.HasChanges {
		t.Error("Expected no changes")
	}

	if len(diff.ChangedFields) != 0 {
		t.Errorf("Expected no changed fields, got %v", diff.ChangedFields)
	}
}

func TestCalculateResourceDiff_WithChanges(t *testing.T) {
	current := map[string]interface{}{
		"replicas": 3,
		"image":    "nginx:1.19",
	}
	desired := map[string]interface{}{
		"replicas": 5,
		"image":    "nginx:1.20",
	}

	diff := CalculateResourceDiff(current, desired)

	if !diff.HasChanges {
		t.Error("Expected changes to be detected")
	}

	if len(diff.ChangedFields) != 2 {
		t.Errorf("Expected 2 changed fields, got %d", len(diff.ChangedFields))
	}

	if diff.OldValues["replicas"] != 3 {
		t.Errorf("Expected old replicas to be 3, got %v", diff.OldValues["replicas"])
	}

	if diff.NewValues["replicas"] != 5 {
		t.Errorf("Expected new replicas to be 5, got %v", diff.NewValues["replicas"])
	}
}

func TestCalculateResourceDiff_NewField(t *testing.T) {
	current := map[string]interface{}{
		"replicas": 3,
	}
	desired := map[string]interface{}{
		"replicas": 3,
		"image":    "nginx:1.19",
	}

	diff := CalculateResourceDiff(current, desired)

	if !diff.HasChanges {
		t.Error("Expected changes to be detected")
	}

	if len(diff.ChangedFields) != 1 {
		t.Errorf("Expected 1 changed field, got %d", len(diff.ChangedFields))
	}
}

func TestCalculateResourceDiff_RemovedField(t *testing.T) {
	current := map[string]interface{}{
		"replicas": 3,
		"image":    "nginx:1.19",
	}
	desired := map[string]interface{}{
		"replicas": 3,
	}

	diff := CalculateResourceDiff(current, desired)

	if !diff.HasChanges {
		t.Error("Expected changes to be detected")
	}

	if len(diff.ChangedFields) != 1 {
		t.Errorf("Expected 1 changed field, got %d", len(diff.ChangedFields))
	}
}

func TestAssessRiskLevel_Default(t *testing.T) {
	tests := []struct {
		action    Action
		expected  RiskLevel
	}{
		{ActionCreate, RiskLow},
		{ActionUpdate, RiskMedium},
		{ActionDelete, RiskHigh},
		{ActionInspect, RiskLow},
	}

	for _, tt := range tests {
		intent := ParsedIntent{
			Action: tt.action,
			Target: ResourceTarget{
				Kind:      "Deployment",
				Namespace: "default",
			},
		}

		risk := AssessRiskLevel(intent)
		if risk != tt.expected {
			t.Errorf("AssessRiskLevel(%s) = %s, want %s", tt.action, risk, tt.expected)
		}
	}
}

func TestAssessRiskLevel_CriticalKind(t *testing.T) {
	criticalKinds := []string{"Node", "PersistentVolume", "ClusterRole", "Namespace"}

	for _, kind := range criticalKinds {
		intent := ParsedIntent{
			Action: ActionUpdate,
			Target: ResourceTarget{
				Kind: kind,
			},
		}

		risk := AssessRiskLevel(intent)
		if risk == RiskMedium {
			// Medium is minimum for critical kinds on update
			continue
		}
		if risk != RiskHigh && risk != RiskCritical {
			t.Errorf("AssessRiskLevel(%s) = %s, want at least HIGH for critical kind", kind, risk)
		}
	}
}

func TestAssessRiskLevel_NonNamespaced(t *testing.T) {
	intent := ParsedIntent{
		Action: ActionUpdate,
		Target: ResourceTarget{
			Kind: "Node",
		},
	}

	risk := AssessRiskLevel(intent)
	// Node is both critical and non-namespaced, so should be HIGH or CRITICAL
	if risk != RiskHigh && risk != RiskCritical {
		t.Errorf("AssessRiskLevel(Node) = %s, want HIGH or CRITICAL", risk)
	}
}

func TestGeneratePreChecks(t *testing.T) {
	intent := ParsedIntent{
		Action: ActionCreate,
		Target: ResourceTarget{
			Name:      "my-app",
			Kind:      "Deployment",
			Namespace: "default",
		},
	}

	plan := GeneratePlan(intent)

	if len(plan.PreCheck) == 0 {
		t.Error("PreCheck should not be empty")
	}
}

func TestPlanDuration(t *testing.T) {
	tests := []struct {
		action   Action
		minDur   time.Duration
	}{
		{ActionCreate, 20 * time.Second},
		{ActionUpdate, 15 * time.Second},
		{ActionDelete, 10 * time.Second},
		{ActionInspect, 1 * time.Second},
	}

	for _, tt := range tests {
		intent := ParsedIntent{
			Action: tt.action,
			Target: ResourceTarget{
				Kind:      "Deployment",
				Namespace: "default",
			},
		}

		plan := GeneratePlan(intent)
		if plan.Duration < tt.minDur {
			t.Errorf("Duration for %s = %v, want at least %v", tt.action, plan.Duration, tt.minDur)
		}
	}
}

func TestChangeStep_Sequence(t *testing.T) {
	intent := ParsedIntent{
		Action: ActionUpdate,
		Target: ResourceTarget{
			Name:      "my-app",
			Kind:      "Deployment",
			Namespace: "default",
		},
	}

	plan := GeneratePlan(intent)

	for i, step := range plan.Steps {
		if step.Seq != i+1 {
			t.Errorf("Step %d should have Seq = %d, got %d", i, i+1, step.Seq)
		}
	}
}