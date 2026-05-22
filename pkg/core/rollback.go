package core

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// ResourceID identifies a Kubernetes resource for snapshot and rollback operations.
// It is composed of the resource's core identity fields.
type ResourceID struct {
	// Name is the name of the resource.
	Name string `json:"name"`
	// Kind is the Kubernetes resource kind (e.g., "Deployment").
	Kind string `json:"kind"`
	// Namespace is the namespace for namespaced resources (empty for cluster-scoped).
	Namespace string `json:"namespace"`
	// APIVersion is the API version (e.g., "apps/v1").
	APIVersion string `json:"apiVersion"`
}

// String returns a string representation of the ResourceID.
// Format: "namespace/kind/name.apiversion" or "kind/name.apiversion" for cluster-scoped.
func (r ResourceID) String() string {
	if r.Namespace != "" {
		return fmt.Sprintf("%s/%s/%s.%s", r.Namespace, r.Kind, r.Name, r.APIVersion)
	}
	return fmt.Sprintf("%s/%s.%s", r.Kind, r.Name, r.APIVersion)
}

// Snapshot represents a point-in-time capture of a Kubernetes resource state.
// Snapshots are used to enable rollback of changes if something goes wrong.
type Snapshot struct {
	// ID uniquely identifies this snapshot.
	ID string `json:"id"`
	// SessionID is the ID of the session that created this snapshot.
	SessionID string `json:"sessionID"`
	// ResourceID identifies the resource this snapshot captures.
	ResourceID ResourceID `json:"resourceID"`
	// Object is the captured resource state.
	Object *unstructured.Unstructured `json:"object"`
	// CreatedAt is when the snapshot was taken.
	CreatedAt time.Time `json:"createdAt"`
}

// String returns a string representation of the Snapshot.
func (s *Snapshot) String() string {
	return fmt.Sprintf("Snapshot{id=%s, sessionID=%s, resource=%s, createdAt=%s}",
		s.ID, s.SessionID, s.ResourceID.String(), s.CreatedAt.Format(time.RFC3339))
}

// Snapshot errors.
var (
	// ErrSnapshotNotFound is returned when a snapshot ID does not exist.
	ErrSnapshotNotFound = errors.New("snapshot not found")
	// ErrNoSnapshotForResource is returned when no snapshot exists for a resource.
	ErrNoSnapshotForResource = errors.New("no snapshot found for resource")
	// ErrNilObject is returned when attempting to snapshot a nil object.
	ErrNilObject = errors.New("cannot create snapshot of nil object")
)

// CreateSnapshot creates a new snapshot of the given Kubernetes resource.
// It captures the current state of the object for potential rollback later.
// The snapshot is stored in memory and indexed by both session ID and resource ID.
func CreateSnapshot(sessionID string, resource ResourceID, obj *unstructured.Unstructured) (*Snapshot, error) {
	if obj == nil {
		return nil, ErrNilObject
	}

	copiedObj := obj.DeepCopy()

	snapshot := &Snapshot{
		ID:         uuid.New().String(),
		SessionID:  sessionID,
		ResourceID: resource,
		Object:     copiedObj,
		CreatedAt:  time.Now().UTC(),
	}

	store.mu.Lock()
	defer store.mu.Unlock()

	store.snapshots[sessionID] = append(store.snapshots[sessionID], snapshot)
	store.resourceIndex[resource] = append(store.resourceIndex[resource], snapshot)

	return snapshot, nil
}

// GetSnapshot retrieves a snapshot by its ID.
// Returns ErrSnapshotNotFound if no such snapshot exists.
func GetSnapshot(snapshotID string) (*Snapshot, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	for _, snapshots := range store.snapshots {
		for _, s := range snapshots {
			if s.ID == snapshotID {
				return s, nil
			}
		}
	}

	return nil, ErrSnapshotNotFound
}

// GetSnapshotsForSession returns all snapshots associated with a session.
// Returns an empty slice if the session has no snapshots.
func GetSnapshotsForSession(sessionID string) ([]*Snapshot, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	snapshots, exists := store.snapshots[sessionID]
	if !exists {
		return []*Snapshot{}, nil
	}

	result := make([]*Snapshot, len(snapshots))
	copy(result, snapshots)
	return result, nil
}

// GetLatestSnapshot returns the most recent snapshot for a given resource.
// Returns ErrNoSnapshotForResource if no snapshots exist for the resource.
func GetLatestSnapshot(resource ResourceID) (*Snapshot, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	snapshots, exists := store.resourceIndex[resource]
	if !exists || len(snapshots) == 0 {
		return nil, ErrNoSnapshotForResource
	}

	var latest *Snapshot
	for _, s := range snapshots {
		if latest == nil || s.CreatedAt.After(latest.CreatedAt) {
			latest = s
		}
	}

	return latest, nil
}

// GetSnapshotsForResource returns all snapshots for a given resource, newest first.
func GetSnapshotsForResource(resource ResourceID) ([]*Snapshot, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	snapshots, exists := store.resourceIndex[resource]
	if !exists {
		return []*Snapshot{}, nil
	}

	result := make([]*Snapshot, len(snapshots))
	copy(result, snapshots)

	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if result[j].CreatedAt.After(result[i].CreatedAt) {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result, nil
}

// Rollback restores a resource to its latest snapshot state within a session.
// It returns the restored object for the caller to apply.
// Logs the rollback operation to the audit log.
func Rollback(sessionID string, resource ResourceID) (*unstructured.Unstructured, error) {
	Log(sessionID, "rollback", "rollback", map[string]interface{}{
		"resource": resource.String(),
	})

	store.mu.RLock()
	defer store.mu.RUnlock()

	snapshots, exists := store.snapshots[sessionID]
	if !exists {
		Log(sessionID, "rollback_failed", "rollback", map[string]interface{}{
			"resource": resource.String(),
			"error":    "no snapshots for session",
		})
		return nil, ErrNoSnapshotForResource
	}

	var latest *Snapshot
	for _, s := range snapshots {
		if s.ResourceID == resource {
			if latest == nil || s.CreatedAt.After(latest.CreatedAt) {
				latest = s
			}
		}
	}

	if latest == nil {
		Log(sessionID, "rollback_failed", "rollback", map[string]interface{}{
			"resource": resource.String(),
			"error":    "no snapshot found for resource",
		})
		return nil, ErrNoSnapshotForResource
	}

	Log(sessionID, "rollback_completed", "rollback", map[string]interface{}{
		"resource":   resource.String(),
		"snapshot_id": latest.ID,
	})

	return latest.Object.DeepCopy(), nil
}

// RollbackToSnapshot restores a resource to a specific snapshot by its ID.
// Returns the restored object for the caller to apply.
func RollbackToSnapshot(snapshotID string) (*unstructured.Unstructured, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	for sessionID, snapshots := range store.snapshots {
		for _, s := range snapshots {
			if s.ID == snapshotID {
				Log(sessionID, "rollback_to_snapshot", "rollback", map[string]interface{}{
					"snapshot_id": snapshotID,
					"resource":    s.ResourceID.String(),
				})
				return s.Object.DeepCopy(), nil
			}
		}
	}

	Log("", "rollback_to_snapshot_failed", "rollback", map[string]interface{}{
		"snapshot_id": snapshotID,
		"error":       "snapshot not found",
	})

	return nil, ErrSnapshotNotFound
}

// DeleteSnapshotsForSession removes all snapshots associated with a session.
// This is typically called when a session is closed to free memory.
func DeleteSnapshotsForSession(sessionID string) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	snapshots, exists := store.snapshots[sessionID]
	if !exists {
		return nil
	}

	for _, s := range snapshots {
		if resourceSnapshots, exists := store.resourceIndex[s.ResourceID]; exists {
			var filtered []*Snapshot
			for _, rs := range resourceSnapshots {
				if rs.SessionID != sessionID {
					filtered = append(filtered, rs)
				}
			}
			if len(filtered) == 0 {
				delete(store.resourceIndex, s.ResourceID)
			} else {
				store.resourceIndex[s.ResourceID] = filtered
			}
		}
	}

	delete(store.snapshots, sessionID)

	return nil
}

// SnapshotCount returns the total number of snapshots in the store.
// This is useful for testing and monitoring.
func SnapshotCount() int {
	store.mu.RLock()
	defer store.mu.RUnlock()

	count := 0
	for _, snapshots := range store.snapshots {
		count += len(snapshots)
	}
	return count
}

// ClearSnapshots clears all snapshots from the store.
// This should only be used in tests.
func ClearSnapshots() {
	store.mu.Lock()
	defer store.mu.Unlock()

	store.snapshots = make(map[string][]*Snapshot)
	store.resourceIndex = make(map[ResourceID][]*Snapshot)
}

// snapshotStore provides in-memory storage for snapshots.
type snapshotStore struct {
	mu          sync.RWMutex
	snapshots   map[string][]*Snapshot       // key: sessionID
	resourceIndex map[ResourceID][]*Snapshot // key: resourceID
}

// Global snapshot store instance.
var store = &snapshotStore{
	snapshots:     make(map[string][]*Snapshot),
	resourceIndex: make(map[ResourceID][]*Snapshot),
}