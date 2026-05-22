package agent

import (
	"fmt"
	"strings"

	"github.com/threestoneliu/k8s-agent/pkg/core"
)

// PlanToNaturalLanguage converts a ChangePlan to human-readable Chinese text.
// It formats the plan's ID, summary, risk level, impact, duration, pre-checks,
// steps, and rollback plan into a readable string representation.
func PlanToNaturalLanguage(plan *core.ChangePlan) string {
	if plan == nil {
		return "（空计划）"
	}

	var sb strings.Builder

	sb.WriteString("=== 变更计划 ===\n\n")

	sb.WriteString(fmt.Sprintf("计划ID: %s\n", plan.ID))
	if plan.Summary != "" {
		sb.WriteString(fmt.Sprintf("概要: %s\n", plan.Summary))
	}

	sb.WriteString(fmt.Sprintf("风险等级: %s\n", plan.RiskLevel))

	if plan.Impact != "" {
		sb.WriteString(fmt.Sprintf("影响范围: %s\n", plan.Impact))
	}

	if plan.Duration > 0 {
		sb.WriteString(fmt.Sprintf("预计耗时: %v\n", plan.Duration))
	}

	if len(plan.PreCheck) > 0 {
		sb.WriteString("\n--- 执行前检查 ---\n")
		for _, check := range plan.PreCheck {
			sb.WriteString(fmt.Sprintf("  - %s\n", check))
		}
	}

	if len(plan.Steps) > 0 {
		sb.WriteString("\n--- 变更步骤 ---\n")
		for _, step := range plan.Steps {
			actionDesc := actionToChinese(step.Action)
			targetDesc := targetToChinese(step.Target)
			riskDesc := riskToChinese(step.RiskLevel)

			sb.WriteString(fmt.Sprintf("\n步骤 %d: %s\n", step.Seq, actionDesc))
			sb.WriteString(fmt.Sprintf("  目标: %s\n", targetDesc))
			sb.WriteString(fmt.Sprintf("  风险: %s\n", riskDesc))
			sb.WriteString(fmt.Sprintf("  可回滚: %v\n", step.CanRollback))
			if step.Validate != "" {
				sb.WriteString(fmt.Sprintf("  验证: %s\n", step.Validate))
			}
			if step.Description != "" {
				sb.WriteString(fmt.Sprintf("  描述: %s\n", step.Description))
			}
		}
	}

	if len(plan.RollbackPlan) > 0 {
		sb.WriteString("\n--- 回滚计划 ---\n")
		for _, step := range plan.RollbackPlan {
			actionDesc := actionToChinese(step.Action)
			targetDesc := targetToChinese(step.Target)
			sb.WriteString(fmt.Sprintf("  - 步骤 %d: %s %s\n", step.Seq, actionDesc, targetDesc))
		}
	}

	sb.WriteString("\n================")

	return sb.String()
}

// DiffToNaturalLanguage converts a ResourceDiff to human-readable Chinese text.
// It shows whether there are changes and lists the changed fields with old and new values.
func DiffToNaturalLanguage(diff *core.ResourceDiff) string {
	if diff == nil {
		return "（空差异）"
	}

	var sb strings.Builder

	sb.WriteString("=== 资源差异 ===\n\n")

	if !diff.HasChanges {
		sb.WriteString("状态: 无变化\n")
		return sb.String()
	}

	sb.WriteString("状态: 有变化\n")

	if len(diff.ChangedFields) > 0 {
		sb.WriteString("\n--- 变更的字段 ---\n")
		for _, field := range diff.ChangedFields {
			sb.WriteString(fmt.Sprintf("  - %s\n", field))
		}
	}

	if len(diff.OldValues) > 0 {
		sb.WriteString("\n--- 原值 ---\n")
		for field, value := range diff.OldValues {
			sb.WriteString(fmt.Sprintf("  %s: %v\n", field, formatValue(value)))
		}
	}

	if len(diff.NewValues) > 0 {
		sb.WriteString("\n--- 新值 ---\n")
		for field, value := range diff.NewValues {
			sb.WriteString(fmt.Sprintf("  %s: %v\n", field, formatValue(value)))
		}
	}

	sb.WriteString("\n================")

	return sb.String()
}

// actionToChinese converts an Action to Chinese description.
func actionToChinese(action core.Action) string {
	switch action {
	case core.ActionCreate:
		return "创建资源"
	case core.ActionUpdate:
		return "更新资源"
	case core.ActionDelete:
		return "删除资源"
	case core.ActionInspect:
		return "查看资源"
	default:
		return string(action)
	}
}

// riskToChinese converts a RiskLevel to Chinese description.
func riskToChinese(level core.RiskLevel) string {
	switch level {
	case core.RiskLow:
		return "低风险"
	case core.RiskMedium:
		return "中风险"
	case core.RiskHigh:
		return "高风险"
	case core.RiskCritical:
		return "极高风险"
	default:
		return string(level)
	}
}

// targetToChinese converts a ResourceTarget to Chinese description.
func targetToChinese(target core.ResourceTarget) string {
	desc := fmt.Sprintf("%s/%s", target.Kind, target.Name)
	if target.Namespace != "" {
		desc = fmt.Sprintf("%s/%s (namespace: %s)", target.Kind, target.Name, target.Namespace)
	}
	return desc
}

// formatValue formats a value for display in diff output.
func formatValue(value interface{}) string {
	if value == nil {
		return "<空>"
	}
	switch v := value.(type) {
	case string:
		if v == "" {
			return "<空字符串>"
		}
		return v
	case map[string]interface{}:
		if len(v) == 0 {
			return "{}"
		}
		var parts []string
		for k, val := range v {
			parts = append(parts, fmt.Sprintf("%s=%v", k, formatValue(val)))
		}
		return "{" + strings.Join(parts, ", ") + "}"
	case []interface{}:
		if len(v) == 0 {
			return "[]"
		}
		var parts []string
		for _, val := range v {
			parts = append(parts, fmt.Sprintf("%v", formatValue(val)))
		}
		return "[" + strings.Join(parts, ", ") + "]"
	default:
		return fmt.Sprintf("%v", v)
	}
}