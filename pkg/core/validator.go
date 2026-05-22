package core

import "fmt"

// Action represents the type of operation being performed.
type Action string

// Action constants for change operations.
const (
	ActionCreate  Action = "CREATE"
	ActionUpdate  Action = "UPDATE"
	ActionDelete  Action = "DELETE"
	ActionInspect Action = "INSPECT"
)

// RiskLevel represents the risk level of an operation.
type RiskLevel string

// RiskLevel constants.
const (
	RiskLow      RiskLevel = "LOW"
	RiskMedium   RiskLevel = "MEDIUM"
	RiskHigh     RiskLevel = "HIGH"
	RiskCritical RiskLevel = "CRITICAL"
)

// ResourceTarget identifies the target resource for an operation.
type ResourceTarget struct {
	Name       string
	Kind       string
	Namespace  string
	APIVersion string
}

// ParsedIntent represents a parsed user intent for a Kubernetes operation.
type ParsedIntent struct {
	Action    Action
	Target    ResourceTarget
	Params    map[string]interface{}
	RiskLevel RiskLevel
	Reason    string
}

// ClarifyQuestion represents a question to clarify incomplete intent.
type ClarifyQuestion struct {
	Field    string
	Question string
	Options  []string
	Required bool
}

// validActions contains all valid action values.
var validActions = map[Action]bool{
	ActionCreate:  true,
	ActionUpdate:  true,
	ActionDelete: true,
	ActionInspect: true,
}

// validRiskLevels contains all valid risk level values.
var validRiskLevels = map[RiskLevel]bool{
	RiskLow:      true,
	RiskMedium:   true,
	RiskHigh:     true,
	RiskCritical: true,
}

// namespacedKinds contains Kubernetes kinds that are namespaced.
var namespacedKinds = map[string]bool{
	"Deployment":             true,
	"Service":                true,
	"ConfigMap":              true,
	"Secret":                 true,
	"Pod":                    true,
	"ReplicaSet":             true,
	"StatefulSet":            true,
	"DaemonSet":              true,
	"Job":                    true,
	"CronJob":                true,
	"Ingress":                true,
	"ServiceAccount":         true,
	"Role":                   true,
	"RoleBinding":            true,
	"PersistentVolumeClaim":  true,
	"Endpoints":              true,
	"LimitRange":             true,
	"ResourceQuota":          true,
	"HorizontalPodAutoscaler": true,
}

// ValidateIntent validates a ParsedIntent and returns a ClarifyQuestion if validation fails.
// Returns nil if the intent is valid and ready for processing.
func ValidateIntent(intent *ParsedIntent) *ClarifyQuestion {
	// Validate Action
	if !validActions[intent.Action] {
		return &ClarifyQuestion{
			Field:    "action",
			Question: fmt.Sprintf("无效的操作类型: %s，有效值为: CREATE, UPDATE, DELETE, INSPECT", intent.Action),
			Required: true,
		}
	}

	// Validate Target.Kind is required
	if intent.Target.Kind == "" {
		return &ClarifyQuestion{
			Field:    "target.kind",
			Question: "目标资源的 Kind 是什么？（例如：Deployment, Service, Pod）",
			Required: true,
		}
	}

	// Target.Name is required for UPDATE/DELETE/INSPECT
	if intent.Action != ActionCreate && intent.Target.Name == "" {
		return &ClarifyQuestion{
			Field:    "target.name",
			Question: "目标资源的名称是什么？",
			Required: true,
		}
	}

	// Target.Namespace is required if Kind is namespaced
	if namespacedKinds[intent.Target.Kind] && intent.Target.Namespace == "" {
		return &ClarifyQuestion{
			Field:    "target.namespace",
			Question: "目标 namespace 是哪个？",
			Required: true,
		}
	}

	// Reason is required if RiskLevel >= HIGH
	if (intent.RiskLevel == RiskHigh || intent.RiskLevel == RiskCritical) && intent.Reason == "" {
		return &ClarifyQuestion{
			Field:    "reason",
			Question: "这个操作的原因是什么？（高风险操作需要明确原因）",
			Required: true,
		}
	}

	return nil
}

// DefaultRiskLevel returns the default risk level for an operation based on action and params.
func DefaultRiskLevel(intent *ParsedIntent) RiskLevel {
	switch intent.Action {
	case ActionCreate:
		// CREATE with standard params defaults to LOW
		return RiskLow
	case ActionUpdate:
		// UPDATE to status/scale or spec/label/annotation defaults to MEDIUM
		return RiskMedium
	case ActionDelete:
		// DELETE any resource defaults to HIGH
		return RiskHigh
	case ActionInspect:
		// INSPECT is typically low risk
		return RiskLow
	default:
		return RiskLow
	}
}

// IsNamespacedKind returns true if the given Kind is a namespaced resource.
func IsNamespacedKind(kind string) bool {
	return namespacedKinds[kind]
}