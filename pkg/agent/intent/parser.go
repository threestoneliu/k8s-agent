package intent

import (
	"strings"

	"github.com/threestoneliu/k8s-agent/pkg/core"
)

// ParseToIntent parses user input into a ParsedIntent structure using simple keyword extraction.
// It recognizes action keywords (create, delete, update, inspect), resource kinds,
// resource names, and namespaces from natural language input.
// This is a stub implementation that will be enhanced with LLM integration later.
func ParseToIntent(userInput string) (*core.ParsedIntent, error) {
	if userInput == "" {
		return nil, nil
	}

	input := strings.ToLower(strings.TrimSpace(userInput))
	intent := &core.ParsedIntent{
		Params: make(map[string]interface{}),
	}

	intent.Action = extractAction(input)
	intent.Target.Kind = extractKind(input)
	intent.Target.Name = extractName(input)
	intent.Target.Namespace = extractNamespace(input)
	intent.RiskLevel = core.DefaultRiskLevel(intent)

	return intent, nil
}

// extractAction determines the action from input keywords.
// It checks for delete, update, create, and inspect keywords.
func extractAction(input string) core.Action {
	if strings.Contains(input, "delete") || strings.Contains(input, "remove") || strings.Contains(input, "删除") {
		return core.ActionDelete
	}
	if strings.Contains(input, "update") || strings.Contains(input, "modify") || strings.Contains(input, "修改") || strings.Contains(input, "edit") {
		return core.ActionUpdate
	}
	if strings.Contains(input, "create") || strings.Contains(input, "add") || strings.Contains(input, "创建") || strings.Contains(input, "新建") {
		return core.ActionCreate
	}
	if strings.Contains(input, "get") || strings.Contains(input, "list") || strings.Contains(input, "show") || strings.Contains(input, "describe") || strings.Contains(input, "inspect") || strings.Contains(input, "查看") || strings.Contains(input, "查询") {
		return core.ActionInspect
	}

	return core.ActionInspect
}

// extractKind extracts the Kubernetes resource kind from input.
// It recognizes common kinds and abbreviations (e.g., "deploy" -> "Deployment").
func extractKind(input string) string {
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

// normalizeKind converts common abbreviations to full kind names.
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
	return strings.Title(pattern)
}

// extractName extracts resource name from input.
// It looks for patterns like "named <name>", "name <name>", and Chinese patterns.
func extractName(input string) string {
	if idx := strings.Index(input, "named "); idx != -1 {
		name := strings.TrimSpace(input[idx+5:])
		name = strings.Fields(name)[0]
		return name
	}
	if idx := strings.Index(input, "name "); idx != -1 {
		name := strings.TrimSpace(input[idx+5:])
		name = strings.Fields(name)[0]
		return name
	}

	if idx := strings.Index(input, "名为"); idx != -1 {
		name := strings.TrimSpace(input[idx+6:])
		name = strings.Fields(name)[0]
		return name
	}
	if idx := strings.Index(input, "名称为"); idx != -1 {
		name := strings.TrimSpace(input[idx+9:])
		name = strings.Fields(name)[0]
		return name
	}

	return ""
}

// extractNamespace extracts namespace from input.
// It looks for patterns like "in namespace <ns>", "namespace <ns>", and Chinese patterns.
func extractNamespace(input string) string {
	if idx := strings.Index(input, "in namespace "); idx != -1 {
		ns := strings.TrimSpace(input[idx+12:])
		return strings.Fields(ns)[0]
	}
	if idx := strings.Index(input, "namespace "); idx != -1 {
		ns := strings.TrimSpace(input[idx+10:])
		return strings.Fields(ns)[0]
	}
	if idx := strings.Index(input, "在命名空间"); idx != -1 {
		ns := strings.TrimSpace(input[idx+15:])
		return strings.Fields(ns)[0]
	}
	if idx := strings.Index(input, "命名空间为"); idx != -1 {
		ns := strings.TrimSpace(input[idx+12:])
		return strings.Fields(ns)[0]
	}

	return ""
}