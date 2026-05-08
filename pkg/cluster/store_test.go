package cluster

import (
	"os"
	"path/filepath"
	"testing"
)

// TestStore_NewStore tests Store creation
func TestStore_NewStore(t *testing.T) {
	// Create temp directory for test
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	store, err := NewStore(WithConfigPath(configPath))
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	if store == nil {
		t.Fatal("NewStore() returned nil")
	}
}

// TestStore_SaveCluster tests saving a cluster configuration
func TestStore_SaveCluster(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")
	clustersPath := filepath.Join(tempDir, "clusters.yaml")

	store, err := NewStore(WithConfigPath(configPath), WithClustersPath(clustersPath))
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	cfg := &ClusterConfig{
		Name:       "test-cluster",
		Kubeconfig: "/path/to/kubeconfig",
	}

	err = store.SaveCluster(cfg)
	if err != nil {
		t.Fatalf("SaveCluster() error = %v", err)
	}

	// Verify cluster can be loaded
	loaded, err := store.LoadCluster("test-cluster")
	if err != nil {
		t.Fatalf("LoadCluster() error = %v", err)
	}

	if loaded.Name != cfg.Name {
		t.Errorf("LoadCluster() name = %v, want %v", loaded.Name, cfg.Name)
	}

	if loaded.Kubeconfig != cfg.Kubeconfig {
		t.Errorf("LoadCluster() kubeconfig = %v, want %v", loaded.Kubeconfig, cfg.Kubeconfig)
	}
}

// TestStore_LoadCluster_NotFound tests loading non-existent cluster
func TestStore_LoadCluster_NotFound(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")
	clustersPath := filepath.Join(tempDir, "clusters.yaml")

	store, err := NewStore(WithConfigPath(configPath), WithClustersPath(clustersPath))
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	_, err = store.LoadCluster("non-existent")
	if !os.IsNotExist(err) {
		t.Errorf("LoadCluster() expected os.IsNotExist, got %v", err)
	}
}

// TestStore_ListClusters tests listing all clusters
func TestStore_ListClusters(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")
	clustersPath := filepath.Join(tempDir, "clusters.yaml")

	store, err := NewStore(WithConfigPath(configPath), WithClustersPath(clustersPath))
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	// List when empty
	clusters := store.ListClusters()
	if len(clusters) != 0 {
		t.Errorf("ListClusters() returned %d clusters, want 0", len(clusters))
	}

	// Add clusters
	store.SaveCluster(&ClusterConfig{Name: "cluster1", Kubeconfig: "/path1"})
	store.SaveCluster(&ClusterConfig{Name: "cluster2", Kubeconfig: "/path2"})

	clusters = store.ListClusters()
	if len(clusters) != 2 {
		t.Errorf("ListClusters() returned %d clusters, want 2", len(clusters))
	}
}

// TestStore_DeleteCluster tests deleting a cluster
func TestStore_DeleteCluster(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")
	clustersPath := filepath.Join(tempDir, "clusters.yaml")

	store, err := NewStore(WithConfigPath(configPath), WithClustersPath(clustersPath))
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	// Add then delete
	store.SaveCluster(&ClusterConfig{Name: "to-delete", Kubeconfig: "/path"})

	err = store.DeleteCluster("to-delete")
	if err != nil {
		t.Fatalf("DeleteCluster() error = %v", err)
	}

	// Verify deleted
	_, err = store.LoadCluster("to-delete")
	if !os.IsNotExist(err) {
		t.Errorf("LoadCluster() after delete should return os.IsNotExist, got %v", err)
	}
}

// TestStore_DeleteCluster_NotFound tests deleting non-existent cluster
func TestStore_DeleteCluster_NotFound(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")
	clustersPath := filepath.Join(tempDir, "clusters.yaml")

	store, err := NewStore(WithConfigPath(configPath), WithClustersPath(clustersPath))
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	err = store.DeleteCluster("non-existent")
	if err != nil {
		t.Errorf("DeleteCluster() for non-existent should not return error, got %v", err)
	}
}

// TestStore_SetCurrentCluster tests setting the current cluster
func TestStore_SetCurrentCluster(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	store, err := NewStore(WithConfigPath(configPath))
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	// Set current cluster
	err = store.SetCurrentCluster("my-cluster")
	if err != nil {
		t.Fatalf("SetCurrentCluster() error = %v", err)
	}

	// Verify current cluster
	current, err := store.GetCurrentCluster()
	if err != nil {
		t.Fatalf("GetCurrentCluster() error = %v", err)
	}

	if current != "my-cluster" {
		t.Errorf("GetCurrentCluster() = %v, want %v", current, "my-cluster")
	}
}

// TestStore_GetCurrentCluster_NotSet tests getting current cluster when not set
func TestStore_GetCurrentCluster_NotSet(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	store, err := NewStore(WithConfigPath(configPath))
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	_, err = store.GetCurrentCluster()
	if err != nil {
		t.Errorf("GetCurrentCluster() when not set should not return error, got %v", err)
	}
}

// TestStore_Persistence tests that data persists across Store instances
func TestStore_Persistence(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "clusters.yaml")
	appConfigPath := filepath.Join(tempDir, "config.yaml")

	// First store - save data
	store1, err := NewStore(
		WithConfigPath(appConfigPath),
		WithClustersPath(configPath),
	)
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	store1.SaveCluster(&ClusterConfig{Name: "persist-cluster", Kubeconfig: "/path"})
	store1.SetCurrentCluster("persist-cluster")

	// Second store - load data
	store2, err := NewStore(
		WithConfigPath(appConfigPath),
		WithClustersPath(configPath),
	)
	if err != nil {
		t.Fatalf("NewStore() second instance error = %v", err)
	}

	// Verify data persisted
	loaded, err := store2.LoadCluster("persist-cluster")
	if err != nil {
		t.Fatalf("LoadCluster() after persistence error = %v", err)
	}

	if loaded.Name != "persist-cluster" {
		t.Errorf("LoadCluster() name = %v, want persist-cluster", loaded.Name)
	}

	current, _ := store2.GetCurrentCluster()
	if current != "persist-cluster" {
		t.Errorf("GetCurrentCluster() = %v, want persist-cluster", current)
	}
}

// TestStore_UpdateCluster tests updating an existing cluster
func TestStore_UpdateCluster(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")
	clustersPath := filepath.Join(tempDir, "clusters.yaml")

	store, err := NewStore(WithConfigPath(configPath), WithClustersPath(clustersPath))
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	// Add initial cluster
	store.SaveCluster(&ClusterConfig{Name: "update-cluster", Kubeconfig: "/old-path"})

	// Update with new kubeconfig
	store.SaveCluster(&ClusterConfig{Name: "update-cluster", Kubeconfig: "/new-path"})

	loaded, _ := store.LoadCluster("update-cluster")
	if loaded.Kubeconfig != "/new-path" {
		t.Errorf("LoadCluster() kubeconfig = %v, want /new-path", loaded.Kubeconfig)
	}
}

// TestStore_ConcurrentAccess tests concurrent reads and writes
func TestStore_ConcurrentAccess(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")
	clustersPath := filepath.Join(tempDir, "clusters.yaml")

	store, err := NewStore(WithConfigPath(configPath), WithClustersPath(clustersPath))
	if err != nil {
		t.Fatalf("NewStore() error = %v", err)
	}

	// Add initial cluster
	store.SaveCluster(&ClusterConfig{Name: "concurrent-cluster", Kubeconfig: "/path"})

	// Concurrent reads
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				store.ListClusters()
				store.LoadCluster("concurrent-cluster")
			}
			done <- true
		}()
	}

	// Concurrent writes
	for i := 0; i < 10; i++ {
		go func(i int) {
			for j := 0; j < 100; j++ {
				store.SaveCluster(&ClusterConfig{Name: "concurrent-cluster", Kubeconfig: "/path"})
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}
}
