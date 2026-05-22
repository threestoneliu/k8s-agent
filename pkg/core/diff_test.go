package core

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestCalculateDiff_IdenticalObjects(t *testing.T) {
	before := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name":      "test-config",
				"namespace": "default",
			},
			"data": map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
		},
	}

	after := before.DeepCopy()

	diff := CalculateDiff(before, after)

	if len(diff.Changes) != 0 {
		t.Errorf("Expected no changes, got %d changes", len(diff.Changes))
	}

	if diff.ResourceID.Name != "test-config" {
		t.Errorf("Expected ResourceID.Name to be 'test-config', got '%s'", diff.ResourceID.Name)
	}
	if diff.ResourceID.Kind != "ConfigMap" {
		t.Errorf("Expected ResourceID.Kind to be 'ConfigMap', got '%s'", diff.ResourceID.Kind)
	}
	if diff.ResourceID.Namespace != "default" {
		t.Errorf("Expected ResourceID.Namespace to be 'default', got '%s'", diff.ResourceID.Namespace)
	}
}

func TestCalculateDiff_SimpleFieldChange(t *testing.T) {
	before := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name":      "test-deploy",
				"namespace": "default",
			},
			"spec": map[string]interface{}{
				"replicas": int64(3),
			},
		},
	}

	after := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name":      "test-deploy",
				"namespace": "default",
			},
			"spec": map[string]interface{}{
				"replicas": int64(5),
			},
		},
	}

	diff := CalculateDiff(before, after)

	if len(diff.Changes) != 1 {
		t.Errorf("Expected 1 change, got %d changes", len(diff.Changes))
	}

	change := diff.Changes[0]
	if len(change.Path) != 2 || change.Path[0] != "spec" || change.Path[1] != "replicas" {
		t.Errorf("Expected path ['spec', 'replicas'], got %v", change.Path)
	}
	if change.OldValue != int64(3) {
		t.Errorf("Expected OldValue to be 3, got %v", change.OldValue)
	}
	if change.NewValue != int64(5) {
		t.Errorf("Expected NewValue to be 5, got %v", change.NewValue)
	}
}

func TestCalculateDiff_NestedFieldChange(t *testing.T) {
	before := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Pod",
			"metadata": map[string]interface{}{
				"name":      "test-pod",
				"namespace": "default",
			},
			"spec": map[string]interface{}{
				"containers": []interface{}{
					map[string]interface{}{
						"name":  "nginx",
						"image": "nginx:1.19",
					},
				},
			},
		},
	}

	after := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Pod",
			"metadata": map[string]interface{}{
				"name":      "test-pod",
				"namespace": "default",
			},
			"spec": map[string]interface{}{
				"containers": []interface{}{
					map[string]interface{}{
						"name":  "nginx",
						"image": "nginx:1.20",
					},
				},
			},
		},
	}

	diff := CalculateDiff(before, after)

	if len(diff.Changes) != 1 {
		t.Errorf("Expected 1 change, got %d changes: %v", len(diff.Changes), diff.Changes)
	}

	change := diff.Changes[0]
	if len(change.Path) < 1 || change.Path[0] != "spec" {
		t.Errorf("Expected path to start with 'spec', got %v", change.Path)
	}
	if change.OldValue != "nginx:1.19" {
		t.Errorf("Expected OldValue to be 'nginx:1.19', got %v", change.OldValue)
	}
	if change.NewValue != "nginx:1.20" {
		t.Errorf("Expected NewValue to be 'nginx:1.20', got %v", change.NewValue)
	}
}

func TestCalculateDiff_AddedField(t *testing.T) {
	before := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name":      "test-config",
				"namespace": "default",
			},
			"data": map[string]interface{}{
				"key1": "value1",
			},
		},
	}

	after := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name":      "test-config",
				"namespace": "default",
			},
			"data": map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
		},
	}

	diff := CalculateDiff(before, after)

	if len(diff.Changes) != 1 {
		t.Errorf("Expected 1 change, got %d changes", len(diff.Changes))
	}

	change := diff.Changes[0]
	if len(change.Path) != 2 || change.Path[0] != "data" || change.Path[1] != "key2" {
		t.Errorf("Expected path ['data', 'key2'], got %v", change.Path)
	}
	if change.OldValue != nil {
		t.Errorf("Expected OldValue to be nil for added field, got %v", change.OldValue)
	}
	if change.NewValue != "value2" {
		t.Errorf("Expected NewValue to be 'value2', got %v", change.NewValue)
	}
}

func TestCalculateDiff_RemovedField(t *testing.T) {
	before := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name":      "test-config",
				"namespace": "default",
			},
			"data": map[string]interface{}{
				"key1": "value1",
				"key2": "value2",
			},
		},
	}

	after := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name":      "test-config",
				"namespace": "default",
			},
			"data": map[string]interface{}{
				"key1": "value1",
			},
		},
	}

	diff := CalculateDiff(before, after)

	if len(diff.Changes) != 1 {
		t.Errorf("Expected 1 change, got %d changes", len(diff.Changes))
	}

	change := diff.Changes[0]
	if len(change.Path) != 2 || change.Path[0] != "data" || change.Path[1] != "key2" {
		t.Errorf("Expected path ['data', 'key2'], got %v", change.Path)
	}
	if change.OldValue != "value2" {
		t.Errorf("Expected OldValue to be 'value2', got %v", change.OldValue)
	}
	if change.NewValue != nil {
		t.Errorf("Expected NewValue to be nil for removed field, got %v", change.NewValue)
	}
}

func TestCalculateDiff_NilBefore(t *testing.T) {
	after := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name":      "test-config",
				"namespace": "default",
			},
			"data": map[string]interface{}{
				"key1": "value1",
			},
		},
	}

	diff := CalculateDiff(nil, after)

	if diff.ResourceID.Name != "test-config" {
		t.Errorf("Expected ResourceID.Name to be 'test-config', got '%s'", diff.ResourceID.Name)
	}
}

func TestCalculateDiff_NilAfter(t *testing.T) {
	before := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "ConfigMap",
			"metadata": map[string]interface{}{
				"name":      "test-config",
				"namespace": "default",
			},
			"data": map[string]interface{}{
				"key1": "value1",
			},
		},
	}

	diff := CalculateDiff(before, nil)

	if diff.ResourceID.Name != "test-config" {
		t.Errorf("Expected ResourceID.Name to be 'test-config', got '%s'", diff.ResourceID.Name)
	}
}

func TestCalculateDiff_MultipleChanges(t *testing.T) {
	before := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name":      "test-deploy",
				"namespace": "default",
			},
			"spec": map[string]interface{}{
				"replicas": int64(3),
				"selector": map[string]interface{}{
					"matchLabels": map[string]interface{}{
						"app": "nginx",
					},
				},
			},
		},
	}

	after := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name":      "test-deploy",
				"namespace": "default",
			},
			"spec": map[string]interface{}{
				"replicas": int64(5),
				"selector": map[string]interface{}{
					"matchLabels": map[string]interface{}{
						"app": "nginx",
					},
				},
			},
		},
	}

	diff := CalculateDiff(before, after)

	if len(diff.Changes) != 1 {
		t.Errorf("Expected 1 change (replicas), got %d changes: %v", len(diff.Changes), diff.Changes)
	}
}

func TestParsePath(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"spec.replicas", []string{"spec", "replicas"}},
		{"spec.containers[0].image", []string{"spec", "containers[0]", "image"}},
		{"metadata.name", []string{"metadata", "name"}},
		{"", []string{}},
		{"single", []string{"single"}},
	}

	for _, tt := range tests {
		result := parsePath(tt.input)
		if len(result) != len(tt.expected) {
			t.Errorf("parsePath(%q) = %v, expected %v", tt.input, result, tt.expected)
			continue
		}
		for i := range result {
			if result[i] != tt.expected[i] {
				t.Errorf("parsePath(%q) = %v, expected %v", tt.input, result, tt.expected)
				break
			}
		}
	}
}