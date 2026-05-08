package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapResource(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Pods
		{"pod singular", "pod", "pods"},
		{"pod plural", "pods", "pods"},
		{"pod mixed case", "POD", "pods"},

		// Deployments
		{"deployment singular", "deployment", "deployments"},
		{"deployment plural", "deployments", "deployments"},
		{"deploy alias", "deploy", "deployments"},

		// Services
		{"service singular", "service", "services"},
		{"service plural", "services", "services"},
		{"svc alias", "svc", "services"},
		{"svcs alias", "svcs", "services"},

		// Namespaces
		{"namespace singular", "namespace", "namespaces"},
		{"namespace plural", "namespaces", "namespaces"},
		{"ns alias", "ns", "namespaces"},

		// Nodes
		{"node singular", "node", "nodes"},
		{"node plural", "nodes", "nodes"},

		// ConfigMaps
		{"configmap singular", "configmap", "configmaps"},
		{"configmap plural", "configmaps", "configmaps"},
		{"cm alias", "cm", "configmaps"},

		// Secrets
		{"secret singular", "secret", "secrets"},
		{"secret plural", "secrets", "secrets"},

		// Ingress
		{"ingress singular", "ingress", "ingresses"},
		{"ingress plural", "ingresses", "ingresses"},
		{"ing alias", "ing", "ingresses"},

		// Unknown resource
		{"unknown resource", "unknown", "unknown"},
		{"case preserved for unknown", "Unknown", "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapResource(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestIsValidResource(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid pod", "pod", true},
		{"valid pods", "pods", true},
		{"valid deployment", "deployment", true},
		{"valid deployments", "deployments", true},
		{"valid service", "service", true},
		{"valid svc", "svc", true},
		{"valid namespace", "namespace", true},
		{"valid ns", "ns", true},
		{"valid node", "node", true},
		{"valid nodes", "nodes", true},
		{"valid configmap", "configmap", true},
		{"valid secret", "secret", true},
		{"valid ingress", "ingress", true},
		{"valid hpa", "hpa", true},
		{"invalid resource", "unknown", false},
		{"empty resource", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsValidResource(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}
