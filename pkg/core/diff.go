package core

import (
	"reflect"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// DetailedFieldChange represents a single field change between two resource states.
// The Path is a list of path components leading to the changed field.
type DetailedFieldChange struct {
	// Path is the JSON path to the changed field as a list of path components.
	Path []string `json:"path"`
	// OldValue is the value before the change (nil if field was added).
	OldValue interface{} `json:"oldValue"`
	// NewValue is the value after the change (nil if field was removed).
	NewValue interface{} `json:"newValue"`
}

// DetailedResourceDiff represents the difference between two versions of a resource.
// It provides field-level change information for detailed comparison.
type DetailedResourceDiff struct {
	// ResourceID identifies the resource this diff applies to.
	ResourceID ResourceID `json:"resourceID"`
	// Before is the resource state before the change.
	Before *unstructured.Unstructured `json:"before"`
	// After is the resource state after the change.
	After *unstructured.Unstructured `json:"after"`
	// Changes is the list of field-level changes detected.
	Changes []DetailedFieldChange `json:"changes"`
}

// CalculateDiff computes the differences between two unstructured Kubernetes objects.
// It performs deep comparison of the object structures and returns a DetailedResourceDiff
// with all field-level changes identified. Either before or after can be nil, but not both.
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
				changes = append(changes, DetailedFieldChange{
					Path:     parsePath(path),
					OldValue: nil,
					NewValue: newField,
				})
			} else if !newExists {
				changes = append(changes, DetailedFieldChange{
					Path:     parsePath(path),
					OldValue: oldField,
					NewValue: nil,
				})
			} else {
				nestedChanges := compareValues(path, oldField, newField)
				changes = append(changes, nestedChanges...)
			}
		}
		return changes
	}

	oldSlice, oldIsSlice := oldVal.([]interface{})
	newSlice, newIsSlice := newVal.([]interface{})

	if oldIsSlice && newIsSlice {
		if len(oldSlice) == len(newSlice) {
			for i := range oldSlice {
				path := prefix + "[" + string(rune('0'+i)) + "]"
				nestedChanges := compareValues(path, oldSlice[i], newSlice[i])
				changes = append(changes, nestedChanges...)
			}
			return changes
		}
		if !reflect.DeepEqual(oldVal, newVal) {
			changes = append(changes, DetailedFieldChange{
				Path:     parsePath(prefix),
				OldValue: oldVal,
				NewValue: newVal,
			})
		}
		return changes
	}

	if (oldIsSlice || newIsSlice) && !reflect.DeepEqual(oldVal, newVal) {
		changes = append(changes, DetailedFieldChange{
			Path:     parsePath(prefix),
			OldValue: oldVal,
			NewValue: newVal,
		})
		return changes
	}

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