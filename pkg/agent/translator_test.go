package agent

import (
	"strings"
	"testing"
	"time"

	"github.com/threestoneliu/k8s-agent/pkg/core"
)

func TestPlanToNaturalLanguage(t *testing.T) {
	tests := []struct {
		name     string
		plan     *core.ChangePlan
		contains []string
	}{
		{
			name:     "nil plan",
			plan:     nil,
			contains: []string{"（空计划）"},
		},
		{
			name: "basic create plan",
			plan: &core.ChangePlan{
				ID:      "plan-123",
				Summary: "创建 Deployment",
				Steps: []core.ChangeStep{
					{
						Seq:        1,
						Action:     core.ActionInspect,
						Target:     core.ResourceTarget{Name: "my-deploy", Kind: "Deployment", Namespace: "default"},
						RiskLevel:  core.RiskLow,
						CanRollback: false,
						Validate:   "check-resource-not-exists",
						Description: "检查目标资源不存在",
					},
					{
						Seq:        2,
						Action:     core.ActionCreate,
						Target:     core.ResourceTarget{Name: "my-deploy", Kind: "Deployment", Namespace: "default"},
						RiskLevel:  core.RiskLow,
						CanRollback: true,
						Validate:   "validate-spec",
						Description: "创建资源",
					},
				},
				PreCheck:     []string{"validate-kubernetes-connection", "check-permissions"},
				RollbackPlan: nil,
				RiskLevel:    core.RiskLow,
				Impact:       "Creates a new Deployment resource",
				Duration:     30 * time.Second,
			},
			contains: []string{
				"计划ID: plan-123",
				"创建 Deployment",
				"风险等级: LOW",
				"步骤 1",
				"创建资源",
				"namespace: default",
			},
		},
		{
			name: "update plan with rollback",
			plan: &core.ChangePlan{
				ID:      "plan-456",
				Summary: "更新 ConfigMap",
				Steps: []core.ChangeStep{
					{
						Seq:        1,
						Action:     core.ActionInspect,
						Target:     core.ResourceTarget{Name: "my-config", Kind: "ConfigMap", Namespace: "default"},
						RiskLevel:  core.RiskLow,
						CanRollback: false,
						Validate:   "check-resource-exists",
						Description: "验证资源存在",
					},
					{
						Seq:        2,
						Action:     core.ActionUpdate,
						Target:     core.ResourceTarget{Name: "my-config", Kind: "ConfigMap", Namespace: "default"},
						RiskLevel:  core.RiskMedium,
						CanRollback: true,
						Validate:   "validate-spec",
						Description: "更新资源",
					},
				},
				PreCheck: []string{"validate-kubernetes-connection", "check-permissions"},
				RollbackPlan: []core.ChangeStep{
					{
						Seq:    1,
						Action: core.ActionUpdate,
						Target: core.ResourceTarget{Name: "my-config", Kind: "ConfigMap", Namespace: "default"},
					},
				},
				RiskLevel: core.RiskMedium,
				Impact:    "Modifies existing ConfigMap resource",
				Duration:  20 * time.Second,
			},
			contains: []string{
				"计划ID: plan-456",
				"更新 ConfigMap",
				"回滚计划",
				"步骤 1: 查看资源",
				"步骤 2: 更新资源",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PlanToNaturalLanguage(tt.plan)
			for _, check := range tt.contains {
				if !strings.Contains(result, check) {
					t.Errorf("PlanToNaturalLanguage() output should contain %q, got:\n%s", check, result)
				}
			}
		})
	}
}

func TestDiffToNaturalLanguage(t *testing.T) {
	tests := []struct {
		name     string
		diff     *core.ResourceDiff
		contains []string
	}{
		{
			name:     "nil diff",
			diff:     nil,
			contains: []string{"（空差异）"},
		},
		{
			name: "no changes",
			diff: &core.ResourceDiff{
				HasChanges:   false,
				ChangedFields: []string{},
				OldValues:    map[string]interface{}{},
				NewValues:    map[string]interface{}{},
			},
			contains: []string{"状态: 无变化"},
		},
		{
			name: "with changes",
			diff: &core.ResourceDiff{
				HasChanges: true,
				ChangedFields: []string{"spec.replicas", "spec.template.spec.containers[0].image"},
				OldValues: map[string]interface{}{
					"spec.replicas": 3,
				},
				NewValues: map[string]interface{}{
					"spec.replicas": 5,
				},
			},
			contains: []string{
				"状态: 有变化",
				"spec.replicas",
				"原值",
				"新值",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DiffToNaturalLanguage(tt.diff)
			for _, check := range tt.contains {
				if !strings.Contains(result, check) {
					t.Errorf("DiffToNaturalLanguage() output should contain %q, got:\n%s", check, result)
				}
			}
		})
	}
}

func TestActionToChinese(t *testing.T) {
	tests := []struct {
		action   core.Action
		expected string
	}{
		{core.ActionCreate, "创建资源"},
		{core.ActionUpdate, "更新资源"},
		{core.ActionDelete, "删除资源"},
		{core.ActionInspect, "查看资源"},
	}

	for _, tt := range tests {
		t.Run(string(tt.action), func(t *testing.T) {
			result := actionToChinese(tt.action)
			if result != tt.expected {
				t.Errorf("actionToChinese(%s) = %s, want %s", tt.action, result, tt.expected)
			}
		})
	}
}

func TestRiskToChinese(t *testing.T) {
	tests := []struct {
		level    core.RiskLevel
		expected string
	}{
		{core.RiskLow, "低风险"},
		{core.RiskMedium, "中风险"},
		{core.RiskHigh, "高风险"},
		{core.RiskCritical, "极高风险"},
	}

	for _, tt := range tests {
		t.Run(string(tt.level), func(t *testing.T) {
			result := riskToChinese(tt.level)
			if result != tt.expected {
				t.Errorf("riskToChinese(%s) = %s, want %s", tt.level, result, tt.expected)
			}
		})
	}
}

func TestTargetToChinese(t *testing.T) {
	tests := []struct {
		name     string
		target   core.ResourceTarget
		contains string
	}{
		{
			name:     "cluster scoped",
			target:   core.ResourceTarget{Name: "my-node", Kind: "Node"},
			contains: "Node/my-node",
		},
		{
			name:     "namespaced",
			target:   core.ResourceTarget{Name: "my-deploy", Kind: "Deployment", Namespace: "default"},
			contains: "namespace: default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := targetToChinese(tt.target)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("targetToChinese() = %s, want contains %s", result, tt.contains)
			}
		})
	}
}

func TestFormatValue(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		contains string
	}{
		{"nil value", nil, "<空>"},
		{"empty string", "", "<空字符串>"},
		{"normal string", "hello", "hello"},
		{"empty map", map[string]interface{}{}, "{}"},
		{"map with values", map[string]interface{}{"key": "value"}, "key=value"},
		{"empty slice", []interface{}{}, "[]"},
		{"slice with values", []interface{}{1, 2, 3}, "1, 2, 3"},
		{"int value", 42, "42"},
		{"bool value", true, "true"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatValue(tt.value)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("formatValue(%v) = %s, want contains %s", tt.value, result, tt.contains)
			}
		})
	}
}