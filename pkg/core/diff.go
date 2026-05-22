package core

import (
	"reflect"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// DetailedFieldChange represents a single field change between two resource states.
type DetailedFieldChange struct {
	Path     []string      // JSON path to the changed field
	OldValue interface{}
	NewValue interface{}
}

// DetailedResourceDiff represents the difference between two versions of a resource.
// This is a more detailed diff than the basic ResourceDiff used in planning.
type DetailedResourceDiff struct {
	ResourceID ResourceID                     // References ResourceID from rollback.go
	Before     *unstructured.Unstructured
	After      *unstructured.Unstructured
	Changes    []DetailedFieldChange
}

// CalculateDiff computes the differences between two unstructured Kubernetes objects.
// It returns a DetailedResourceDiff with all field-level changes identified.
func CalculateDiff(before, after *unstructured.Unstructured) *DetailedResourceDiff {
	if before == nil && after == nil {
		return &DetailedResourceDiff{Changes: []DetailedFieldChange{}}
	}

	var rid ResourceID
	if before != nil {
		rid = ResourceID{
			Name:       before.GetName(),
			Kind:       before.GetKind(),
			Namespace:  before.GetNamespace(),
			APIVersion: before.GetAPIVersion(),
		}
	} else if after != nil {
		rid = ResourceID{
			Name:       after.GetName(),
			Kind:       after.GetKind(),
			Namespace:  after.GetNamespace(),
			APIVersion: after.GetAPIVersion(),
		}
	}

	if before == nil {
		before = &unstructured.Unstructured{}
	}
	if after == nil {
		after = &unstructured.Unstructured{}
	}

	changes := compareValues("", before.Object, after.Object)

	return &DetailedResourceDiff{
		ResourceID: rid,
		Before:     before,
		After:      after,
		Changes:    changes,
	}
}

// compareValues recursively compares two values and returns a list of field changes.
func compareValues(prefix string, oldVal, newVal interface{}) []DetailedFieldChange {
	var changes []DetailedFieldChange

	oldMap, oldIsMap := oldVal.(map[string]interface{})
	newMap, newIsMap := newVal.(map[string]interface{})

	if oldIsMap && newIsMap {
		// Both are maps - compare keys
		allKeys := make(map[string]bool)
		for k := range oldMap {
			allKeys[k] = true
		}
		for k := range newMap {
			allKeys[k] = true
		}

		for key := range allKeys {
			path := key
			if prefix != "" {
				path = prefix + "." + key
			}

			oldField, oldExists := oldMap[key]
			newField, newExists := newMap[key]

			if !oldExists {
				// Field was added
				changes = append(changes, DetailedFieldChange{
					Path:     parsePath(path),
					OldValue: nil,
					NewValue: newField,
				})
			} else if !newExists {
				// Field was removed
				changes = append(changes, DetailedFieldChange{
					Path:     parsePath(path),
					OldValue: oldField,
					NewValue: nil,
				})
			} else {
				// Both exist - recurse
				nestedChanges := compareValues(path, oldField, newField)
				changes = append(changes, nestedChanges...)
			}
		}
		return changes
	}

	// Handle slices
	oldSlice, oldIsSlice := oldVal.([]interface{})
	newSlice, newIsSlice := newVal.([]interface{})

	if oldIsSlice && newIsSlice {
		// For slices, compare element by element if lengths are the same
		// Otherwise, treat as a complete replacement
		if len(oldSlice) == len(newSlice) {
			for i := range oldSlice {
				path := prefix + "[" + string(rune('0'+i)) + "]"
				nestedChanges := compareValues(path, oldSlice[i], newSlice[i])
				changes = append(changes, nestedChanges...)
			}
			return changes
		}
		// Length changed - treat as complete replacement
		if !reflect.DeepEqual(oldVal, newVal) {
			changes = append(changes, DetailedFieldChange{
				Path:     parsePath(prefix),
				OldValue: oldVal,
				NewValue: newVal,
			})
		}
		return changes
	}

	// Handle case where one is slice and other is not
	if (oldIsSlice || newIsSlice) && !reflect.DeepEqual(oldVal, newVal) {
		changes = append(changes, DetailedFieldChange{
			Path:     parsePath(prefix),
			OldValue: oldVal,
			NewValue: newVal,
		})
		return changes
	}

	// Base case - compare values
	if !reflect.DeepEqual(oldVal, newVal) {
		changes = append(changes, DetailedFieldChange{
			Path:     parsePath(prefix),
			OldValue: oldVal,
			NewValue: newVal,
		})
	}

	return changes
}

// parsePath converts a dot-separated path string to a slice of path components.
func parsePath(path string) []string {
	if path == "" {
		return []string{}
	}
	var result []string
	current := ""
	for _, c := range path {
		if c == '.' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}