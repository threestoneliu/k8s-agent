package cli

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetCommand(t *testing.T) {
	cmd := NewRootCommand()
	getCmd, _, err := cmd.Find([]string{"get"})
	require.NoError(t, err)

	assert.Equal(t, "get", getCmd.Name())
	assert.NotNil(t, getCmd.RunE)
}

func TestListCommand(t *testing.T) {
	cmd := NewRootCommand()
	listCmd, _, err := cmd.Find([]string{"list"})
	require.NoError(t, err)

	assert.Equal(t, "list", listCmd.Name())
	assert.NotNil(t, listCmd.RunE)
}

func TestDescribeCommand(t *testing.T) {
	cmd := NewRootCommand()
	describeCmd, _, err := cmd.Find([]string{"describe"})
	require.NoError(t, err)

	assert.Equal(t, "describe", describeCmd.Name())
	assert.NotNil(t, describeCmd.RunE)
}

func TestDeleteCommand(t *testing.T) {
	cmd := NewRootCommand()
	deleteCmd, _, err := cmd.Find([]string{"delete"})
	require.NoError(t, err)

	assert.Equal(t, "delete", deleteCmd.Name())
	assert.NotNil(t, deleteCmd.RunE)
}

func TestCreateCommand(t *testing.T) {
	cmd := NewRootCommand()
	createCmd, _, err := cmd.Find([]string{"create"})
	require.NoError(t, err)

	assert.Equal(t, "create", createCmd.Name())
	assert.NotNil(t, createCmd.RunE)
}

func TestScaleCommand(t *testing.T) {
	cmd := NewRootCommand()
	scaleCmd, _, err := cmd.Find([]string{"scale"})
	require.NoError(t, err)

	assert.Equal(t, "scale", scaleCmd.Name())
	assert.NotNil(t, scaleCmd.RunE)
}

func TestResourceCommandExecution(t *testing.T) {
	cmd := NewRootCommand()

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "get with no args",
			args:    []string{"get"},
			wantErr: true,
		},
		{
			name:    "list with no args",
			args:    []string{"list"},
			wantErr: true,
		},
		{
			name:    "describe with no args",
			args:    []string{"describe"},
			wantErr: true,
		},
		{
			name:    "delete with no args",
			args:    []string{"delete"},
			wantErr: true,
		},
		{
			name:    "create with no args",
			args:    []string{"create"},
			wantErr: true,
		},
		{
			name:    "scale with no args",
			args:    []string{"scale"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			cmd.SetOut(buf)
			cmd.SetErr(buf)
			cmd.SetArgs(tt.args)

			err := cmd.Execute()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// Test resource parsing
func TestParseResourceType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "pods",
			input:    "pods",
			expected: "pods",
		},
		{
			name:     "services",
			input:    "services",
			expected: "services",
		},
		{
			name:     "deployments",
			input:    "deployments",
			expected: "deployments",
		},
		{
			name:     "namespaces",
			input:    "namespaces",
			expected: "namespaces",
		},
		{
			name:     "pods singular",
			input:    "pod",
			expected: "pods",
		},
		{
			name:     "service singular",
			input:    "service",
			expected: "services",
		},
		{
			name:     "deployment singular",
			input:    "deployment",
			expected: "deployments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeResourceType(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper to normalize resource type
func normalizeResourceType(input string) string {
	singularToPlural := map[string]string{
		"pod":         "pods",
		"service":     "services",
		"deployment":  "deployments",
		"namespace":   "namespaces",
		"configmap":   "configmaps",
		"secret":      "secrets",
		"ingress":     "ingresses",
		"statefulset": "statefulsets",
		"daemonset":   "daemonsets",
		"job":         "jobs",
		"cronjob":     "cronjobs",
	}

	if plural, ok := singularToPlural[input]; ok {
		return plural
	}
	return input
}

func TestParseNamespaceFlag(t *testing.T) {
	tests := []struct {
		name      string
		args      []string
		expectNS  string
		expectIdx int
	}{
		{
			name:      "with -n flag",
			args:      []string{"-n", "default", "pods"},
			expectNS:  "default",
			expectIdx: 2,
		},
		{
			name:      "with --namespace flag",
			args:      []string{"--namespace", "kube-system", "pods"},
			expectNS:  "kube-system",
			expectIdx: 2,
		},
		{
			name:      "without namespace flag",
			args:      []string{"pods"},
			expectNS:  "",
			expectIdx: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ns, idx := parseNamespaceFlag(tt.args)
			assert.Equal(t, tt.expectNS, ns)
			assert.Equal(t, tt.expectIdx, idx)
		})
	}
}

// Helper to parse namespace flag
func parseNamespaceFlag(args []string) (string, int) {
	for i, arg := range args {
		if arg == "-n" || arg == "--namespace" {
			if i+1 < len(args) {
				return args[i+1], i + 2
			}
		}
	}
	return "", 0
}

func TestParseResourceName(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected string
	}{
		{
			name:     "simple resource",
			args:     []string{"pods"},
			expected: "",
		},
		{
			name:     "resource with name",
			args:     []string{"pods", "my-pod"},
			expected: "my-pod",
		},
		{
			name:     "resource with namespace and name",
			args:     []string{"-n", "default", "pods", "my-pod"},
			expected: "my-pod",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name := parseResourceName(tt.args)
			assert.Equal(t, tt.expected, name)
		})
	}
}

// Helper to parse resource name from args
func parseResourceName(args []string) string {
	if len(args) == 0 {
		return ""
	}
	// Skip namespace flags
	idx := 0
	for idx < len(args) && (args[idx] == "-n" || args[idx] == "--namespace") {
		idx += 2
	}
	if idx < len(args) {
		idx++ // Skip resource type
	}
	if idx < len(args) {
		return args[idx]
	}
	return ""
}
