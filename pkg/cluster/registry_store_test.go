package cluster

import (
	"os"
	"path/filepath"
	"testing"
)

// TestRegistry_WithStore tests Registry with Store integration
func TestRegistry_WithStore(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")
	clustersPath := filepath.Join(tempDir, "clusters.yaml")

	store, err := NewStore(WithConfigPath(configPath), WithClustersPath(clustersPath))
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	// Create registry with store
	r := NewRegistry(WithStore(store))

	// Add cluster - should persist to store
	err = r.AddCluster("persistent-cluster", "/kubeconfig/path")
	if err != nil {
		t.Fatalf("AddCluster() error = %v", err)
	}

	// Create new registry with same store - should see persisted data
	r2 := NewRegistry(WithStore(store))
	clusters := r2.ListClusters()
	if len(clusters) != 1 {
		t.Errorf("ListClusters() returned %d clusters, want 1", len(clusters))
	}
}

// TestRegistry_AddCluster_WithStore tests AddCluster persists to Store
func TestRegistry_AddCluster_WithStore(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")
	clustersPath := filepath.Join(tempDir, "clusters.yaml")

	store, err := NewStore(WithConfigPath(configPath), WithClustersPath(clustersPath))
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	r := NewRegistry(WithStore(store))

	err = r.AddCluster("test-cluster", "/path/to/kubeconfig")
	if err != nil {
		t.Fatalf("AddCluster() error = %v", err)
	}

	// Verify directly from store
	loaded, err := store.LoadCluster("test-cluster")
	if err != nil {
		t.Fatalf("store.LoadCluster() error = %v", err)
	}

	if loaded.Name != "test-cluster" {
		t.Errorf("store.LoadCluster() name = %v, want test-cluster", loaded.Name)
	}
}

// TestRegistry_RemoveCluster_WithStore tests RemoveCluster persists to Store
func TestRegistry_RemoveCluster_WithStore(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")
	clustersPath := filepath.Join(tempDir, "clusters.yaml")

	store, err := NewStore(WithConfigPath(configPath), WithClustersPath(clustersPath))
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	r := NewRegistry(WithStore(store))

	// Add then remove
	r.AddCluster("to-remove", "/path")
	r.RemoveCluster("to-remove")

	// Verify directly from store
	_, err = store.LoadCluster("to-remove")
	if !os.IsNotExist(err) {
		t.Errorf("store.LoadCluster() after delete should return os.IsNotExist, got %v", err)
	}
}

// TestRegistry_GetCurrentCluster tests GetCurrentCluster method
func TestRegistry_GetCurrentCluster(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")
	clustersPath := filepath.Join(tempDir, "clusters.yaml")

	store, err := NewStore(WithConfigPath(configPath), WithClustersPath(clustersPath))
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	r := NewRegistry(WithStore(store))

	// Set current cluster
	store.SetCurrentCluster("my-cluster")

	// Get current cluster from registry
	current, err := r.GetCurrentCluster()
	if err != nil {
		t.Fatalf("GetCurrentCluster() error = %v", err)
	}

	if current != "my-cluster" {
		t.Errorf("GetCurrentCluster() = %v, want my-cluster", current)
	}
}

// TestRegistry_GetCurrentCluster_NotSet tests GetCurrentCluster when not set
func TestRegistry_GetCurrentCluster_NotSet(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")
	clustersPath := filepath.Join(tempDir, "clusters.yaml")

	store, err := NewStore(WithConfigPath(configPath), WithClustersPath(clustersPath))
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	r := NewRegistry(WithStore(store))

	current, err := r.GetCurrentCluster()
	if err != nil {
		t.Fatalf("GetCurrentCluster() error = %v", err)
	}

	if current != "" {
		t.Errorf("GetCurrentCluster() = %v, want empty string", current)
	}
}
