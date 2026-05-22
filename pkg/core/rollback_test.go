package core

import (
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestResourceIDString(t *testing.T) {
	tests := []struct {
		name     string
		resource ResourceID
		expected string
	}{
		{
			name:     "namespaced resource",
			resource: ResourceID{Name: "my-deploy", Kind: "Deployment", Namespace: "default", APIVersion: "apps/v1"},
			expected: "default/Deployment/my-deploy.apps/v1",
		},
		{
			name:     "cluster-scoped resource",
			resource: ResourceID{Name: "my-node", Kind: "Node", Namespace: "", APIVersion: "v1"},
			expected: "Node/my-node.v1",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.resource.String()
			if got != tc.expected {
				t.Errorf("ResourceID.String() = %v, want %v", got, tc.expected)
			}
		})
	}
}

func TestSnapshotString(t *testing.T) {
	now := time.Date(2026, 5, 22, 10, 30, 0, 0, time.UTC)
	snapshot := &Snapshot{
		ID:         "test-id-123",
		SessionID:  "session-456",
		ResourceID: ResourceID{Name: "my-deploy", Kind: "Deployment", Namespace: "default", APIVersion: "apps/v1"},
		CreatedAt:  now,
	}

	expected := "Snapshot{id=test-id-123, sessionID=session-456, resource=default/Deployment/my-deploy.apps/v1, createdAt=2026-05-22T10:30:00Z}"
	got := snapshot.String()
	if got != expected {
		t.Errorf("Snapshot.String() = %v, want %v", got, expected)
	}
}

func TestCreateSnapshot(t *testing.T) {
	ClearSnapshots() // Ensure clean state
	defer ClearSnapshots()

	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name":      "my-deploy",
				"namespace": "default",
			},
			"spec": map[string]interface{}{
				"replicas": int64(3),
			},
		},
	}

	resourceID := ResourceID{
		Name:       "my-deploy",
		Kind:       "Deployment",
		Namespace:  "default",
		APIVersion: "apps/v1",
	}

	snapshot, err := CreateSnapshot("session-1", resourceID, obj)
	if err != nil {
		t.Fatalf("CreateSnapshot() error = %v, want nil", err)
	}

	if snapshot.ID == "" {
		t.Error("CreateSnapshot() returned snapshot with empty ID")
	}
	if snapshot.SessionID != "session-1" {
		t.Errorf("CreateSnapshot() SessionID = %v, want session-1", snapshot.SessionID)
	}
	if snapshot.ResourceID != resourceID {
		t.Errorf("CreateSnapshot() ResourceID = %v, want %v", snapshot.ResourceID, resourceID)
	}
	if snapshot.Object == nil {
		t.Error("CreateSnapshot() returned nil Object")
	}
	if snapshot.CreatedAt.IsZero() {
		t.Error("CreateSnapshot() returned zero CreatedAt")
	}
}

func TestCreateSnapshotNilObject(t *testing.T) {
	ClearSnapshots()
	defer ClearSnapshots()

	resourceID := ResourceID{Name: "my-deploy", Kind: "Deployment", Namespace: "default", APIVersion: "apps/v1"}

	_, err := CreateSnapshot("session-1", resourceID, nil)
	if err != ErrNilObject {
		t.Errorf("CreateSnapshot(nil) error = %v, want %v", err, ErrNilObject)
	}
}

func TestCreateSnapshotMultiple(t *testing.T) {
	ClearSnapshots()
	defer ClearSnapshots()

	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata":   map[string]interface{}{"name": "my-deploy", "namespace": "default"},
		},
	}

	resourceID := ResourceID{Name: "my-deploy", Kind: "Deployment", Namespace: "default", APIVersion: "apps/v1"}

	// Create multiple snapshots for the same resource
	snap1, _ := CreateSnapshot("session-1", resourceID, obj)
	snap2, _ := CreateSnapshot("session-1", resourceID, obj)
	snap3, _ := CreateSnapshot("session-2", resourceID, obj)

	if snap1.ID == snap2.ID {
		t.Error("CreateSnapshot() created snapshots with duplicate IDs")
	}
	if snap2.ID == snap3.ID {
		t.Error("CreateSnapshot() created snapshots with duplicate IDs")
	}

	// Check snapshot count
	count := SnapshotCount()
	if count != 3 {
		t.Errorf("SnapshotCount() = %v, want 3", count)
	}
}

func TestGetLatestSnapshot(t *testing.T) {
	ClearSnapshots()
	defer ClearSnapshots()

	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata":   map[string]interface{}{"name": "my-deploy", "namespace": "default"},
		},
	}

	resourceID := ResourceID{Name: "my-deploy", Kind: "Deployment", Namespace: "default", APIVersion: "apps/v1"}

	// Create snapshots with a small delay between them
	snap1, _ := CreateSnapshot("session-1", resourceID, obj)
	time.Sleep(time.Millisecond * 10)

	// Modify the object
	obj2 := obj.DeepCopy()
	obj2.Object["spec"] = map[string]interface{}{"replicas": int64(5)}
	snap2, _ := CreateSnapshot("session-1", resourceID, obj2)

	latest, err := GetLatestSnapshot(resourceID)
	if err != nil {
		t.Fatalf("GetLatestSnapshot() error = %v, want nil", err)
	}

	if latest.ID != snap2.ID {
		t.Errorf("GetLatestSnapshot() returned snapshot ID %v, want %v", latest.ID, snap2.ID)
	}

	// Verify we can get the older snapshot too
	older, err := GetSnapshot(snap1.ID)
	if err != nil {
		t.Errorf("GetSnapshot() error = %v, want nil", err)
	}
	if older.ID != snap1.ID {
		t.Errorf("GetSnapshot() returned wrong snapshot")
	}
}

func TestGetLatestSnapshotNotFound(t *testing.T) {
	ClearSnapshots()
	defer ClearSnapshots()

	resourceID := ResourceID{Name: "nonexistent", Kind: "Deployment", Namespace: "default", APIVersion: "apps/v1"}

	_, err := GetLatestSnapshot(resourceID)
	if err != ErrNoSnapshotForResource {
		t.Errorf("GetLatestSnapshot() for nonexistent resource error = %v, want %v", err, ErrNoSnapshotForResource)
	}
}

func TestRollback(t *testing.T) {
	ClearSnapshots()
	defer ClearSnapshots()

	originalObj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata":   map[string]interface{}{"name": "my-deploy", "namespace": "default"},
			"spec":       map[string]interface{}{"replicas": int64(3)},
		},
	}

	resourceID := ResourceID{Name: "my-deploy", Kind: "Deployment", Namespace: "default", APIVersion: "apps/v1"}

	// Create snapshot of original state
	CreateSnapshot("session-1", resourceID, originalObj)

	// Modify the object (simulating an update)
	modifiedObj := originalObj.DeepCopy()
	modifiedObj.Object["spec"] = map[string]interface{}{"replicas": int64(10)}

	// Rollback to the snapshot
	restored, err := Rollback("session-1", resourceID)
	if err != nil {
		t.Fatalf("Rollback() error = %v, want nil", err)
	}

	// Verify the restored object has the original replica count
	spec, ok := restored.Object["spec"].(map[string]interface{})
	if !ok {
		t.Fatal("Rollback() returned object without spec")
	}
	replicas, ok := spec["replicas"].(int64)
	if !ok {
		t.Fatal("Rollback() returned object without replicas in spec")
	}
	if replicas != 3 {
		t.Errorf("Rollback() replicas = %v, want 3", replicas)
	}
}

func TestRollbackNoSnapshotForSession(t *testing.T) {
	ClearSnapshots()
	defer ClearSnapshots()

	resourceID := ResourceID{Name: "my-deploy", Kind: "Deployment", Namespace: "default", APIVersion: "apps/v1"}

	_, err := Rollback("nonexistent-session", resourceID)
	if err != ErrNoSnapshotForResource {
		t.Errorf("Rollback() for nonexistent session error = %v, want %v", err, ErrNoSnapshotForResource)
	}
}

func TestRollbackNoSnapshotForResource(t *testing.T) {
	ClearSnapshots()
	defer ClearSnapshots()

	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata":   map[string]interface{}{"name": "other-deploy", "namespace": "default"},
		},
	}
	resourceID := ResourceID{Name: "other-deploy", Kind: "Deployment", Namespace: "default", APIVersion: "apps/v1"}

	CreateSnapshot("session-1", resourceID, obj)

	// Try to rollback a different resource
	differentResourceID := ResourceID{Name: "my-deploy", Kind: "Deployment", Namespace: "default", APIVersion: "apps/v1"}
	_, err := Rollback("session-1", differentResourceID)
	if err != ErrNoSnapshotForResource {
		t.Errorf("Rollback() for nonexistent resource error = %v, want %v", err, ErrNoSnapshotForResource)
	}
}

func TestGetSnapshotsForSession(t *testing.T) {
	ClearSnapshots()
	defer ClearSnapshots()

	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata":   map[string]interface{}{"name": "my-deploy", "namespace": "default"},
		},
	}

	resourceID := ResourceID{Name: "my-deploy", Kind: "Deployment", Namespace: "default", APIVersion: "apps/v1"}

	// Create multiple snapshots for the same session
	CreateSnapshot("session-1", resourceID, obj)
	time.Sleep(time.Millisecond * 10)
	CreateSnapshot("session-1", resourceID, obj)
	time.Sleep(time.Millisecond * 10)
	CreateSnapshot("session-1", resourceID, obj)

	// Create snapshots for a different session
	CreateSnapshot("session-2", resourceID, obj)

	snapshots, err := GetSnapshotsForSession("session-1")
	if err != nil {
		t.Fatalf("GetSnapshotsForSession() error = %v, want nil", err)
	}

	if len(snapshots) != 3 {
		t.Errorf("GetSnapshotsForSession() returned %v snapshots, want 3", len(snapshots))
	}
}

func TestGetSnapshotsForSessionNotFound(t *testing.T) {
	ClearSnapshots()
	defer ClearSnapshots()

	snapshots, err := GetSnapshotsForSession("nonexistent-session")
	if err != nil {
		t.Fatalf("GetSnapshotsForSession() error = %v, want nil", err)
	}

	if len(snapshots) != 0 {
		t.Errorf("GetSnapshotsForSession() returned %v snapshots, want 0", len(snapshots))
	}
}

func TestGetSnapshotsForResource(t *testing.T) {
	ClearSnapshots()
	defer ClearSnapshots()

	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata":   map[string]interface{}{"name": "my-deploy", "namespace": "default"},
		},
	}

	resourceID := ResourceID{Name: "my-deploy", Kind: "Deployment", Namespace: "default", APIVersion: "apps/v1"}

	// Create snapshots from different sessions
	CreateSnapshot("session-1", resourceID, obj)
	time.Sleep(time.Millisecond * 10)
	CreateSnapshot("session-2", resourceID, obj)
	time.Sleep(time.Millisecond * 10)
	CreateSnapshot("session-3", resourceID, obj)

	snapshots, err := GetSnapshotsForResource(resourceID)
	if err != nil {
		t.Fatalf("GetSnapshotsForResource() error = %v, want nil", err)
	}

	if len(snapshots) != 3 {
		t.Errorf("GetSnapshotsForResource() returned %v snapshots, want 3", len(snapshots))
	}

	// Verify they are sorted newest first
	for i := 0; i < len(snapshots)-1; i++ {
		if snapshots[i].CreatedAt.Before(snapshots[i+1].CreatedAt) {
			t.Error("GetSnapshotsForResource() returned snapshots not sorted by CreatedAt descending")
		}
	}
}

func TestRollbackToSnapshot(t *testing.T) {
	ClearSnapshots()
	defer ClearSnapshots()

	originalObj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata":   map[string]interface{}{"name": "my-deploy", "namespace": "default"},
			"spec":       map[string]interface{}{"replicas": int64(3)},
		},
	}

	resourceID := ResourceID{Name: "my-deploy", Kind: "Deployment", Namespace: "default", APIVersion: "apps/v1"}

	snap, _ := CreateSnapshot("session-1", resourceID, originalObj)

	// Rollback to specific snapshot
	restored, err := RollbackToSnapshot(snap.ID)
	if err != nil {
		t.Fatalf("RollbackToSnapshot() error = %v, want nil", err)
	}

	spec := restored.Object["spec"].(map[string]interface{})
	replicas := spec["replicas"].(int64)
	if replicas != 3 {
		t.Errorf("RollbackToSnapshot() replicas = %v, want 3", replicas)
	}
}

func TestRollbackToSnapshotNotFound(t *testing.T) {
	ClearSnapshots()
	defer ClearSnapshots()

	_, err := RollbackToSnapshot("nonexistent-id")
	if err != ErrSnapshotNotFound {
		t.Errorf("RollbackToSnapshot() for nonexistent ID error = %v, want %v", err, ErrSnapshotNotFound)
	}
}

func TestDeleteSnapshotsForSession(t *testing.T) {
	ClearSnapshots()
	defer ClearSnapshots()

	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata":   map[string]interface{}{"name": "my-deploy", "namespace": "default"},
		},
	}

	resourceID := ResourceID{Name: "my-deploy", Kind: "Deployment", Namespace: "default", APIVersion: "apps/v1"}

	// Create snapshots for two sessions
	CreateSnapshot("session-1", resourceID, obj)
	CreateSnapshot("session-1", resourceID, obj)
	CreateSnapshot("session-2", resourceID, obj)

	if SnapshotCount() != 3 {
		t.Errorf("SnapshotCount() = %v, want 3", SnapshotCount())
	}

	// Delete snapshots for session-1
	err := DeleteSnapshotsForSession("session-1")
	if err != nil {
		t.Fatalf("DeleteSnapshotsForSession() error = %v, want nil", err)
	}

	if SnapshotCount() != 1 {
		t.Errorf("After DeleteSnapshotsForSession(), SnapshotCount() = %v, want 1", SnapshotCount())
	}

	// Verify session-2 snapshots still exist
	snapshots, _ := GetSnapshotsForSession("session-2")
	if len(snapshots) != 1 {
		t.Errorf("GetSnapshotsForSession(session-2) returned %v snapshots, want 1", len(snapshots))
	}
}

func TestDeleteSnapshotsForSessionNotFound(t *testing.T) {
	ClearSnapshots()
	defer ClearSnapshots()

	// Should not error when deleting non-existent session
	err := DeleteSnapshotsForSession("nonexistent-session")
	if err != nil {
		t.Fatalf("DeleteSnapshotsForSession() for nonexistent session error = %v, want nil", err)
	}
}

func TestSnapshotIndependence(t *testing.T) {
	// Verify that modifying a snapshot's Object does not affect the stored object
	ClearSnapshots()
	defer ClearSnapshots()

	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata":   map[string]interface{}{"name": "my-deploy", "namespace": "default"},
			"spec":       map[string]interface{}{"replicas": int64(3)},
		},
	}

	resourceID := ResourceID{Name: "my-deploy", Kind: "Deployment", Namespace: "default", APIVersion: "apps/v1"}

	snap, _ := CreateSnapshot("session-1", resourceID, obj)

	// Modify the returned object's spec
	snap.Object.Object["spec"] = map[string]interface{}{"replicas": int64(999)}

	// Create another snapshot
	snap2, _ := CreateSnapshot("session-1", resourceID, obj)

	// Verify the new snapshot has the original value
	spec := snap2.Object.Object["spec"].(map[string]interface{})
	replicas := spec["replicas"].(int64)
	if replicas != 3 {
		t.Errorf("Second snapshot has replicas = %v, want 3 (modification should not affect store)", replicas)
	}
}