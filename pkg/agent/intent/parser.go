package intent

import (
	"strings"

	"github.com/threestoneliu/k8s-agent/pkg/core"
)

// ParseToIntent parses user input into a ParsedIntent structure using simple keyword extraction.
// This is a stub implementation that will be enhanced with LLM integration later.
func ParseToIntent(userInput string) (*core.ParsedIntent, error) {
	if userInput == "" {
		return nil, nil
	}

	input := strings.ToLower(strings.TrimSpace(userInput))
	intent := &core.ParsedIntent{
		Params: make(map[string]interface{}),
	}

	// Extract action from keywords
	intent.Action = extractAction(input)

	// Extract resource kind
	intent.Target.Kind = extractKind(input)

	// Extract resource name if present
	intent.Target.Name = extractName(input)

	// Extract namespace if present
	intent.Target.Namespace = extractNamespace(input)

	// Set default risk level based on action
	intent.RiskLevel = core.DefaultRiskLevel(intent)

	return intent, nil
}

// extractAction determines the action from input keywords
func extractAction(input string) core.Action {
	// Check for delete keywords
	if strings.Contains(input, "delete") || strings.Contains(input, "remove") || strings.Contains(input, "删除") {
		return core.ActionDelete
	}
	// Check for update keywords
	if strings.Contains(input, "update") || strings.Contains(input, "modify") || strings.Contains(input, "修改") || strings.Contains(input, "edit") {
		return core.ActionUpdate
	}
	// Check for create keywords
	if strings.Contains(input, "create") || strings.Contains(input, "add") || strings.Contains(input, "创建") || strings.Contains(input, "新建") {
		return core.ActionCreate
	}
	// Check for inspect keywords
	if strings.Contains(input, "get") || strings.Contains(input, "list") || strings.Contains(input, "show") || strings.Contains(input, "describe") || strings.Contains(input, "inspect") || strings.Contains(input, "查看") || strings.Contains(input, "查询") {
		return core.ActionInspect
	}

	// Default to Inspect for queries
	return core.ActionInspect
}

// extractKind extracts the Kubernetes resource kind from input
func extractKind(input string) string {
	// Define common resource kind patterns
	kindPatterns := []string{
		"deployment", "deploy",
		"service", "svc",
		"pod", "pods",
		"configmap", "cm",
		"secret",
		"ingress",
		"statefulset", "sts",
		"daemonset", "ds",
		"job",
		"cronjob",
		"node",
		"namespace", "ns",
		"persistentvolumeclaim", "pvc",
		"replicaset", "rs",
		"endpoint", "ep",
		"serviceaccount", "sa",
		"role",
		"rolebinding",
		"hpa",
	}

	for _, pattern := range kindPatterns {
		if strings.Contains(input, pattern) {
			return normalizeKind(pattern)
		}
	}

	return ""
}

// normalizeKind converts common abbreviations to full kind names
func normalizeKind(pattern string) string {
	kindMap := map[string]string{
		"deploy":       "Deployment",
		"deployment":   "Deployment",
		"svc":          "Service",
		"service":      "Service",
		"pod":          "Pod",
		"pods":         "Pod",
		"cm":           "ConfigMap",
		"configmap":    "ConfigMap",
		"secret":       "Secret",
		"ingress":      "Ingress",
		"sts":          "StatefulSet",
		"statefulset":  "StatefulSet",
		"ds":           "DaemonSet",
		"daemonset":    "DaemonSet",
		"job":          "Job",
		"cronjob":      "CronJob",
		"node":         "Node",
		"ns":           "Namespace",
		"namespace":    "Namespace",
		"pvc":          "PersistentVolumeClaim",
		"persistentvolumeclaim": "PersistentVolumeClaim",
		"rs":           "ReplicaSet",
		"replicaset":   "ReplicaSet",
		"ep":           "Endpoints",
		"endpoint":     "Endpoints",
		"sa":           "ServiceAccount",
		"serviceaccount": "ServiceAccount",
		"role":         "Role",
		"rolebinding":  "RoleBinding",
		"hpa":          "HorizontalPodAutoscaler",
	}

	if full, ok := kindMap[pattern]; ok {
		return full
	}
	// Capitalize first letter for unknown kinds
	return strings.Title(pattern)
}

// extractName extracts resource name from input
func extractName(input string) string {
	// Look for patterns like "named <name>", "name <name>", "<name> in namespace"
	// or standalone words that could be resource names

	// Pattern: "named <name>" or "name <name>"
	if idx := strings.Index(input, "named "); idx != -1 {
		name := strings.TrimSpace(input[idx+5:])
		// Extract just the name part (stop at whitespace or punctuation)
		name = strings.Fields(name)[0]
		return name
	}
	if idx := strings.Index(input, "name "); idx != -1 {
		name := strings.TrimSpace(input[idx+5:])
		name = strings.Fields(name)[0]
		return name
	}

	// Pattern: Chinese "名为" or "名称为"
	// Note: Chinese characters are 3 bytes each in UTF-8
	if idx := strings.Index(input, "名为"); idx != -1 {
		name := strings.TrimSpace(input[idx+6:]) // skip 6 bytes (2 Chinese chars)
		name = strings.Fields(name)[0]
		return name
	}
	if idx := strings.Index(input, "名称为"); idx != -1 {
		name := strings.TrimSpace(input[idx+9:]) // skip 9 bytes (3 Chinese chars)
		name = strings.Fields(name)[0]
		return name
	}

	return ""
}

// extractNamespace extracts namespace from input
func extractNamespace(input string) string {
	// Pattern: "in namespace <ns>" or "namespace <ns>"
	if idx := strings.Index(input, "in namespace "); idx != -1 {
		ns := strings.TrimSpace(input[idx+12:])
		return strings.Fields(ns)[0]
	}
	if idx := strings.Index(input, "namespace "); idx != -1 {
		ns := strings.TrimSpace(input[idx+10:])
		return strings.Fields(ns)[0]
	}
	// Chinese patterns - characters are 3 bytes each in UTF-8
	if idx := strings.Index(input, "在命名空间"); idx != -1 {
		ns := strings.TrimSpace(input[idx+15:]) // skip 15 bytes (5 Chinese chars)
		return strings.Fields(ns)[0]
	}
	if idx := strings.Index(input, "命名空间为"); idx != -1 {
		ns := strings.TrimSpace(input[idx+12:]) // skip 12 bytes (4 Chinese chars)
		return strings.Fields(ns)[0]
	}

	return ""
}