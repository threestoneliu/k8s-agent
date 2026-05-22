package core

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// ResourceID identifies a Kubernetes resource for snapshot/rollback operations.
// It is derived from the resource's identity fields.
type ResourceID struct {
	Name       string `json:"name"`
	Kind       string `json:"kind"`
	Namespace  string `json:"namespace"`
	APIVersion string `json:"apiVersion"`
}

// String returns a string representation of the ResourceID.
func (r ResourceID) String() string {
	if r.Namespace != "" {
		return fmt.Sprintf("%s/%s/%s.%s", r.Namespace, r.Kind, r.Name, r.APIVersion)
	}
	return fmt.Sprintf("%s/%s.%s", r.Kind, r.Name, r.APIVersion)
}

// Snapshot represents a point-in-time capture of a Kubernetes resource state.
type Snapshot struct {
	ID         string                    `json:"id"`
	SessionID  string                    `json:"sessionID"`
	ResourceID ResourceID                `json:"resourceID"`
	Object     *unstructured.Unstructured `json:"object"`
	CreatedAt  time.Time                `json:"createdAt"`
}

// String returns a string representation of the Snapshot.
func (s *Snapshot) String() string {
	return fmt.Sprintf("Snapshot{id=%s, sessionID=%s, resource=%s, createdAt=%s}",
		s.ID, s.SessionID, s.ResourceID.String(), s.CreatedAt.Format(time.RFC3339))
}

// snapshotStore provides in-memory storage for snapshots.
type snapshotStore struct {
	mu         sync.RWMutex
	snapshots  map[string][]*Snapshot // key: sessionID, value: list of snapshots for that session
	resourceIndex map[ResourceID][]*Snapshot // key: resourceID, value: list of snapshots for that resource
}

// Global snapshot store instance.
var store = &snapshotStore{
	snapshots:     make(map[string][]*Snapshot),
	resourceIndex: make(map[ResourceID][]*Snapshot),
}

// Snapshot errors.
var (
	ErrSnapshotNotFound     = errors.New("snapshot not found")
	ErrNoSnapshotForResource = errors.New("no snapshot found for resource")
	ErrNilObject            = errors.New("cannot create snapshot of nil object")
)

// CreateSnapshot creates a new snapshot of the given Kubernetes resource.
// It captures the current state of the object for potential rollback later.
func CreateSnapshot(sessionID string, resource ResourceID, obj *unstructured.Unstructured) (*Snapshot, error) {
	if obj == nil {
		return nil, ErrNilObject
	}

	// Deep copy the object to preserve its state
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

	// Add to session-based index
	store.snapshots[sessionID] = append(store.snapshots[sessionID], snapshot)

	// Add to resource-based index for lookup
	store.resourceIndex[resource] = append(store.resourceIndex[resource], snapshot)

	return snapshot, nil
}

// GetSnapshot retrieves a snapshot by its ID.
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
func GetSnapshotsForSession(sessionID string) ([]*Snapshot, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	snapshots, exists := store.snapshots[sessionID]
	if !exists {
		return []*Snapshot{}, nil
	}

	// Return a copy to avoid external modifications
	result := make([]*Snapshot, len(snapshots))
	copy(result, snapshots)
	return result, nil
}

// GetLatestSnapshot returns the most recent snapshot for a given resource.
func GetLatestSnapshot(resource ResourceID) (*Snapshot, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	snapshots, exists := store.resourceIndex[resource]
	if !exists || len(snapshots) == 0 {
		return nil, ErrNoSnapshotForResource
	}

	// Find the latest snapshot
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

	// Sort by CreatedAt descending (newest first)
	result := make([]*Snapshot, len(snapshots))
	copy(result, snapshots)

	// Simple bubble sort for small slices (snapshots per resource should be few)
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if result[j].CreatedAt.After(result[i].CreatedAt) {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result, nil
}

// Rollback restores a resource to its latest snapshot state.
// It returns the restored object for the caller to apply.
func Rollback(sessionID string, resource ResourceID) (*unstructured.Unstructured, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	// Find the latest snapshot for this resource in this session
	snapshots, exists := store.snapshots[sessionID]
	if !exists {
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
		return nil, ErrNoSnapshotForResource
	}

	// Return a deep copy of the object for restoration
	return latest.Object.DeepCopy(), nil
}

// RollbackToSnapshot restores a resource to a specific snapshot by ID.
func RollbackToSnapshot(snapshotID string) (*unstructured.Unstructured, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	for _, snapshots := range store.snapshots {
		for _, s := range snapshots {
			if s.ID == snapshotID {
				return s.Object.DeepCopy(), nil
			}
		}
	}

	return nil, ErrSnapshotNotFound
}

// DeleteSnapshotsForSession removes all snapshots associated with a session.
// This is typically called when a session is closed.
func DeleteSnapshotsForSession(sessionID string) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	snapshots, exists := store.snapshots[sessionID]
	if !exists {
		return nil
	}

	// Remove from resource index
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

	// Remove from session index
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