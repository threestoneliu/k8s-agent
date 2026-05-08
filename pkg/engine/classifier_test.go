package engine

import (
	"testing"
)

func TestClassifyQueryOperations(t *testing.T) {
	tests := []struct {
		name     string
		verb     string
		resource string
		want     OperationType
	}{
		{"get pods", "get", "pods", OperationTypeQuery},
		{"list deployments", "list", "deployments", OperationTypeQuery},
		{"describe pod", "describe", "pod", OperationTypeQuery},
		{"watch pods", "watch", "pods", OperationTypeQuery},
		{"logs pod", "logs", "pod", OperationTypeQuery},
		{"exec pod", "exec", "pod", OperationTypeQuery},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyVerb(tt.verb, tt.resource)
			if got != tt.want {
				t.Errorf("ClassifyVerb(%q, %q) = %v, want %v", tt.verb, tt.resource, got, tt.want)
			}
		})
	}
}

func TestClassifyMutationOperations(t *testing.T) {
	tests := []struct {
		name     string
		verb     string
		resource string
		want     OperationType
	}{
		{"create deployment", "create", "deployment", OperationTypeMutation},
		{"update deployment", "update", "deployment", OperationTypeMutation},
		{"patch deployment", "patch", "deployment", OperationTypeMutation},
		{"delete pod", "delete", "pod", OperationTypeMutation},
		{"scale deployment", "scale", "deployment", OperationTypeMutation},
		{"cordon node", "cordon", "node", OperationTypeMutation},
		{"uncordon node", "uncordon", "node", OperationTypeMutation},
		{"drain node", "drain", "node", OperationTypeMutation},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyVerb(tt.verb, tt.resource)
			if got != tt.want {
				t.Errorf("ClassifyVerb(%q, %q) = %v, want %v", tt.verb, tt.resource, got, tt.want)
			}
		})
	}
}

func TestClassifyHighRiskResources(t *testing.T) {
	// High-risk resource mutations require confirmation, but queries do not
	mutationVerbs := []string{"create", "update", "patch", "delete", "scale", "cordon", "uncordon", "drain"}

	for _, verb := range mutationVerbs {
		t.Run(verb+" nodes", func(t *testing.T) {
			got := ClassifyVerb(verb, "nodes")
			if got != OperationTypeMutation {
				t.Errorf("ClassifyVerb(%q, %q) = %v, want %v (high-risk resource mutation)", verb, "nodes", got, OperationTypeMutation)
			}
		})
	}

	// Query operations on high-risk resources do NOT require confirmation
	queryVerbs := []string{"get", "list", "describe", "watch"}

	for _, verb := range queryVerbs {
		t.Run(verb+" nodes", func(t *testing.T) {
			got := ClassifyVerb(verb, "nodes")
			if got != OperationTypeQuery {
				t.Errorf("ClassifyVerb(%q, %q) = %v, want %v (query operation)", verb, "nodes", got, OperationTypeQuery)
			}
		})
	}
}

func TestClassifyUnknownVerb(t *testing.T) {
	tests := []struct {
		name string
		verb string
	}{
		{"unknown verb", "unknown"},
		{"empty verb", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyVerb(tt.verb, "pod")
			if got != OperationTypeUnknown {
				t.Errorf("ClassifyVerb(%q, _) = %v, want %v", tt.verb, got, OperationTypeUnknown)
			}
		})
	}
}

func TestClassifyOperation(t *testing.T) {
	tests := []struct {
		name   string
		input  *ParsedOperation
		expect OperationType
	}{
		{
			name:   "get pods",
			input:  &ParsedOperation{Verb: "get", Resource: "pods", RawInput: "get pods"},
			expect: OperationTypeQuery,
		},
		{
			name:   "delete pod",
			input:  &ParsedOperation{Verb: "delete", Resource: "pod", RawInput: "delete pod"},
			expect: OperationTypeMutation,
		},
		{
			name:   "get nodes (high risk) - query",
			input:  &ParsedOperation{Verb: "get", Resource: "nodes", RawInput: "get nodes"},
			expect: OperationTypeQuery,
		},
		{
			name:   "describe namespace (high risk) - query",
			input:  &ParsedOperation{Verb: "describe", Resource: "namespaces", RawInput: "describe namespace"},
			expect: OperationTypeQuery,
		},
		{
			name:   "delete nodes (high risk) - mutation",
			input:  &ParsedOperation{Verb: "delete", Resource: "nodes", RawInput: "delete nodes"},
			expect: OperationTypeMutation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyOperation(tt.input)
			if got.Type != tt.expect {
				t.Errorf("ClassifyOperation() = %v, want %v", got.Type, tt.expect)
			}
		})
	}
}

func TestClassifyOperationReturnsClassified(t *testing.T) {
	input := &ParsedOperation{
		Verb:      "get",
		Resource:  "pods",
		Name:      "nginx",
		Namespace: "default",
		Flags:     map[string]string{"n": "default"},
		RawInput:  "get pods -n default",
	}

	result := ClassifyOperation(input)

	if result.Type != OperationTypeQuery {
		t.Errorf("expected OperationTypeQuery, got %v", result.Type)
	}
	if result.Verb != input.Verb {
		t.Errorf("expected Verb %q, got %q", input.Verb, result.Verb)
	}
	if result.Resource != input.Resource {
		t.Errorf("expected Resource %q, got %q", input.Resource, result.Resource)
	}
	if result.RawInput != input.RawInput {
		t.Errorf("expected RawInput %q, got %q", input.RawInput, result.RawInput)
	}
}
