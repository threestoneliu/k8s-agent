package engine

// OperationType represents the type of operation
type OperationType int

const (
	OperationTypeQuery OperationType = iota
	OperationTypeMutation
	OperationTypeUnknown
)

// ClassifiedOperation represents an operation with its classification
type ClassifiedOperation struct {
	Type      OperationType
	Verb      string
	Resource  string
	Name      string
	Namespace string
	Flags     map[string]string
	RawInput  string
}

// queryVerbs lists verbs that represent query operations
var queryVerbs = map[string]bool{
	"get":      true,
	"list":     true,
	"describe": true,
	"watch":    true,
	"logs":     true,
	"exec":     true,
}

// mutationVerbs lists verbs that represent mutation operations
var mutationVerbs = map[string]bool{
	"create":   true,
	"update":   true,
	"patch":    true,
	"delete":   true,
	"scale":    true,
	"cordon":   true,
	"uncordon": true,
	"drain":    true,
}

// highRiskResources lists resources that are considered high-risk
var highRiskResources = map[string]bool{
	"nodes":           true,
	"persistentvolumes": true,
	"namespaces":      true,
	"storageclasses":  true,
}

// ClassifyVerb classifies a verb and resource combination
func ClassifyVerb(verb, resource string) OperationType {
	verb = normalizeVerb(verb)

	// Query verbs always return Query type (even for high-risk resources)
	if queryVerbs[verb] {
		return OperationTypeQuery
	}

	// Mutation verbs or high-risk resources require mutation confirmation
	if mutationVerbs[verb] || highRiskResources[resource] {
		return OperationTypeMutation
	}

	return OperationTypeUnknown
}

// ClassifyOperation classifies a parsed operation
func ClassifyOperation(op *ParsedOperation) *ClassifiedOperation {
	return &ClassifiedOperation{
		Type:      ClassifyVerb(op.Verb, op.Resource),
		Verb:      op.Verb,
		Resource:  op.Resource,
		Name:      op.Name,
		Namespace: op.Namespace,
		Flags:     op.Flags,
		RawInput:  op.RawInput,
	}
}

// normalizeVerb handles verb variations
func normalizeVerb(verb string) string {
	return verb
}
