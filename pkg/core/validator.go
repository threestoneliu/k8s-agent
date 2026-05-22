package core

import "fmt"

// Action represents the type of Kubernetes operation being performed.
// Actions drive the change management workflow and determine execution behavior.
type Action string

// Action constants for change operations.
const (
	// ActionCreate creates a new Kubernetes resource.
	ActionCreate Action = "CREATE"
	// ActionUpdate modifies an existing Kubernetes resource.
	ActionUpdate Action = "UPDATE"
	// ActionDelete removes a Kubernetes resource.
	ActionDelete Action = "DELETE"
	// ActionInspect reads and displays Kubernetes resource information.
	ActionInspect Action = "INSPECT"
)

// RiskLevel represents the risk level associated with an operation.
// Higher risk levels require additional justification and pre-checks.
type RiskLevel string

// RiskLevel constants.
const (
	RiskLow      RiskLevel = "LOW"      // Safe operations with minimal impact
	RiskMedium   RiskLevel = "MEDIUM"   // Operations with moderate impact
	RiskHigh     RiskLevel = "HIGH"     // Operations with significant impact
	RiskCritical RiskLevel = "CRITICAL" // Operations affecting critical infrastructure
)

// ResourceTarget identifies the target Kubernetes resource for an operation.
// All fields must be properly set for the operation to be valid.
type ResourceTarget struct {
	// Name is the name of the target resource.
	Name string
	// Kind is the Kubernetes resource kind (e.g., "Deployment", "Service").
	Kind string
	// Namespace is the namespace of the resource (empty for cluster-scoped resources).
	Namespace string
	// APIVersion is the API version of the resource (e.g., "apps/v1").
	APIVersion string
}

// ParsedIntent represents a parsed user intent for a Kubernetes operation.
// It contains all the information needed to create a ChangePlan.
type ParsedIntent struct {
	// Action is the type of operation to perform.
	Action Action
	// Target identifies the target resource for the operation.
	Target ResourceTarget
	// Params contains additional operation-specific parameters.
	Params map[string]interface{}
	// RiskLevel is the assessed risk level for the operation.
	RiskLevel RiskLevel
	// Reason provides justification for high-risk operations.
	Reason string
}

// ClarifyQuestion represents a question to clarify incomplete or ambiguous intent.
// When ValidateIntent returns a ClarifyQuestion, the workflow pauses until
// the user provides the missing information.
type ClarifyQuestion struct {
	// Field is the name of the field that needs clarification.
	Field string
	// Question is the text of the question to present to the user.
	Question string
	// Options are available answer choices (empty if free-form input).
	Options []string
	// Required indicates whether an answer is mandatory to proceed.
	Required bool
}

// namespacedKinds contains Kubernetes kinds that are namespaced.
// Namespaced resources require a Namespace field in ResourceTarget.
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
// Returns nil if the intent is valid and ready for planning.
// Validation checks include: valid action, required target fields, namespace requirements,
// and reason requirements for high-risk operations.
func ValidateIntent(intent *ParsedIntent) *ClarifyQuestion {
	if intent == nil {
		return &ClarifyQuestion{
			Field:    "intent",
			Question: "intent is required",
			Required: true,
		}
	}

	if !validActions[intent.Action] {
		return &ClarifyQuestion{
			Field:    "action",
			Question: fmt.Sprintf("invalid action type: %s, valid values are: CREATE, UPDATE, DELETE, INSPECT", intent.Action),
			Required: true,
		}
	}

	if intent.Target.Kind == "" {
		return &ClarifyQuestion{
			Field:    "target.kind",
			Question: "what is the resource kind? (e.g., Deployment, Service, Pod)",
			Required: true,
		}
	}

	if intent.Action != ActionCreate && intent.Target.Name == "" {
		return &ClarifyQuestion{
			Field:    "target.name",
			Question: "what is the resource name?",
			Required: true,
		}
	}

	if namespacedKinds[intent.Target.Kind] && intent.Target.Namespace == "" {
		return &ClarifyQuestion{
			Field:    "target.namespace",
			Question: "which namespace?",
			Required: true,
		}
	}

	if (intent.RiskLevel == RiskHigh || intent.RiskLevel == RiskCritical) && intent.Reason == "" {
		return &ClarifyQuestion{
			Field:    "reason",
			Question: "what is the reason for this operation? (high-risk operations require justification)",
			Required: true,
		}
	}

	return nil
}

// DefaultRiskLevel returns the default risk level for an operation based on action type.
// These defaults can be overridden by explicit risk assessment.
func DefaultRiskLevel(intent *ParsedIntent) RiskLevel {
	if intent == nil {
		return RiskLow
	}

	switch intent.Action {
	case ActionCreate:
		return RiskLow
	case ActionUpdate:
		return RiskMedium
	case ActionDelete:
		return RiskHigh
	case ActionInspect:
		return RiskLow
	default:
		return RiskLow
	}
}

// IsNamespacedKind returns true if the given Kind is a namespaced resource.
// Namespaced resources must include a Namespace in their ResourceTarget.
func IsNamespacedKind(kind string) bool {
	return namespacedKinds[kind]
}

// validActions contains all valid action values.
var validActions = map[Action]bool{
	ActionCreate:  true,
	ActionUpdate:  true,
	ActionDelete: true,
	ActionInspect: true,
}