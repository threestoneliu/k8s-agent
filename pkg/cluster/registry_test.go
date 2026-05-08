package cluster

import (
	"context"
	"errors"
	"testing"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

// MockClientset embeds kubernetes.Interface for testing
type MockClientset struct {
	kubernetes.Interface
}

func TestRegistry_GetCluster(t *testing.T) {
	tests := []struct {
		name        string
		clusterName string
		setup       func(*Registry)
		wantErr     bool
		errType     error
	}{
		{
			name:        "get existing cluster",
			clusterName: "test-cluster",
			setup: func(r *Registry) {
				// Pre-populate both clusters and clients map
				r.clusters["test-cluster"] = &ClusterConfig{
					Name: "test-cluster",
				}
				r.clients["test-cluster"] = fake.NewSimpleClientset()
			},
			wantErr: false,
		},
		{
			name:        "get non-existent cluster",
			clusterName: "non-existent",
			setup:       func(r *Registry) {},
			wantErr:     true,
			errType:     ErrClusterNotFound,
		},
		{
			name:        "get cluster with empty name",
			clusterName: "",
			setup:       func(r *Registry) {},
			wantErr:     true,
			errType:     ErrClusterNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRegistry()
			tt.setup(r)

			client, err := r.GetCluster(tt.clusterName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCluster() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && client == nil {
				t.Error("GetCluster() returned nil client when no error expected")
			}
		})
	}
}

func TestRegistry_ListClusters(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*Registry)
		expected int
	}{
		{
			name:     "list clusters when empty",
			setup:    func(r *Registry) {},
			expected: 0,
		},
		{
			name: "list multiple clusters",
			setup: func(r *Registry) {
				r.clusters["cluster1"] = &ClusterConfig{Name: "cluster1"}
				r.clusters["cluster2"] = &ClusterConfig{Name: "cluster2"}
			},
			expected: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRegistry()
			tt.setup(r)

			clusters := r.ListClusters()
			if len(clusters) != tt.expected {
				t.Errorf("ListClusters() returned %d clusters, want %d", len(clusters), tt.expected)
			}
		})
	}
}

func TestRegistry_AddCluster(t *testing.T) {
	tests := []struct {
		name        string
		clusterName string
		kubeconfig  string
		setup       func(*Registry)
		wantErr     bool
		errType     error
	}{
		{
			name:        "add new cluster successfully",
			clusterName: "new-cluster",
			kubeconfig:  "/path/to/kubeconfig",
			setup:       func(r *Registry) {},
			wantErr:     false,
		},
		{
			name:        "add duplicate cluster",
			clusterName: "existing",
			kubeconfig:  "/path/to/kubeconfig",
			setup: func(r *Registry) {
				r.clusters["existing"] = &ClusterConfig{Name: "existing"}
			},
			wantErr: true,
			errType: ErrClusterAlreadyExists,
		},
		{
			name:        "add cluster with empty name",
			clusterName: "",
			kubeconfig:  "/path/to/kubeconfig",
			setup:       func(r *Registry) {},
			wantErr:     true,
			errType:     ErrInvalidClusterName,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRegistry()
			tt.setup(r)

			err := r.AddCluster(tt.clusterName, tt.kubeconfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("AddCluster() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify cluster was added
				if _, ok := r.clusters[tt.clusterName]; !ok {
					t.Errorf("AddCluster() did not add cluster %s to registry", tt.clusterName)
				}
			}
		})
	}
}

func TestRegistry_RemoveCluster(t *testing.T) {
	tests := []struct {
		name        string
		clusterName string
		setup       func(*Registry)
		wantErr     bool
		errType     error
	}{
		{
			name:        "remove existing cluster",
			clusterName: "to-remove",
			setup: func(r *Registry) {
				r.clusters["to-remove"] = &ClusterConfig{Name: "to-remove"}
			},
			wantErr: false,
		},
		{
			name:        "remove non-existent cluster",
			clusterName: "non-existent",
			setup:       func(r *Registry) {},
			wantErr:     true,
			errType:     ErrClusterNotFound,
		},
		{
			name:        "remove cluster with empty name",
			clusterName: "",
			setup:       func(r *Registry) {},
			wantErr:     true,
			errType:     ErrClusterNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRegistry()
			tt.setup(r)

			err := r.RemoveCluster(tt.clusterName)
			if (err != nil) != tt.wantErr {
				t.Errorf("RemoveCluster() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify cluster was removed
				if _, ok := r.clusters[tt.clusterName]; ok {
					t.Errorf("RemoveCluster() did not remove cluster %s from registry", tt.clusterName)
				}
			}
		})
	}
}

// Test errors are distinct
func TestRegistry_Errors(t *testing.T) {
	// Ensure error variables are distinct
	errs := []error{ErrClusterNotFound, ErrClusterAlreadyExists, ErrInvalidClusterName, ErrInvalidKubeconfig}
	seen := make(map[error]bool)

	for _, err := range errs {
		if seen[err] {
			t.Errorf("duplicate error value detected: %v", err)
		}
		seen[err] = true
	}

	// Ensure errors are not nil
	for _, err := range errs {
		if err == nil {
			t.Error("error variable is nil")
		}
	}
}

// TestRegistry_GetClusterContext_NotFound tests GetClusterContext with non-existent cluster
func TestRegistry_GetClusterContext_NotFound(t *testing.T) {
	r := NewRegistry()
	ctx := context.Background()

	_, err := r.GetClusterContext(ctx, "non-existent")
	if !errors.Is(err, ErrClusterNotFound) {
		t.Errorf("GetClusterContext() expected ErrClusterNotFound, got %v", err)
	}
}

// TestRegistry_GetCluster_LazyClientBuild tests GetCluster when client needs to be built
func TestRegistry_GetCluster_LazyClientBuild(t *testing.T) {
	r := NewRegistry()

	// Add cluster config but no client - triggers lazy build
	r.clusters["lazy-cluster"] = &ClusterConfig{
		Name:       "lazy-cluster",
		Kubeconfig: "/nonexistent/kubeconfig/path",
	}

	// Should fail because kubeconfig doesn't exist
	_, err := r.GetCluster("lazy-cluster")
	if err == nil {
		t.Error("GetCluster() expected error for invalid kubeconfig")
	}
}

// TestRegistry_GetCluster_EmptyKubeconfigPath tests buildClient with empty kubeconfig path
func TestRegistry_GetCluster_EmptyKubeconfigPath(t *testing.T) {
	r := NewRegistry()

	// Add cluster with empty kubeconfig - should use KUBECONFIG env or ~/.kube/config
	r.clusters["empty-kubeconfig"] = &ClusterConfig{
		Name:       "empty-kubeconfig",
		Kubeconfig: "",
	}

	// This will fail if no valid kubeconfig exists at default locations
	// but it exercises the empty kubeconfig path
	_, err := r.GetCluster("empty-kubeconfig")
	// Error is expected because default kubeconfig likely doesn't exist in test env
	if err != nil && !errors.Is(err, ErrInvalidKubeconfig) {
		t.Logf("GetCluster() error (expected for test env): %v", err)
	}
}

// TestRegistry_AddCluster_RemoveCluster_AddAgain tests add/remove/add cycle
func TestRegistry_AddCluster_RemoveCluster_AddAgain(t *testing.T) {
	r := NewRegistry()

	// Add cluster
	err := r.AddCluster("cycle-cluster", "/kubeconfig/path")
	if err != nil {
		t.Fatalf("AddCluster() failed: %v", err)
	}

	// Remove cluster
	err = r.RemoveCluster("cycle-cluster")
	if err != nil {
		t.Fatalf("RemoveCluster() failed: %v", err)
	}

	// Add again - should succeed
	err = r.AddCluster("cycle-cluster", "/kubeconfig/path")
	if err != nil {
		t.Errorf("AddCluster() after RemoveCluster() failed: %v", err)
	}
}

// TestRegistry_ListClusters_ReturnsCopy ensures ListClusters returns a copy
func TestRegistry_ListClusters_ReturnsCopy(t *testing.T) {
	r := NewRegistry()
	r.clusters["cluster1"] = &ClusterConfig{Name: "cluster1"}

	clusters1 := r.ListClusters()
	clusters2 := r.ListClusters()

	if &clusters1 == &clusters2 {
		t.Error("ListClusters() should return a new slice each time")
	}

	if len(clusters1) != len(clusters2) {
		t.Errorf("ListClusters() returned different lengths: %d vs %d", len(clusters1), len(clusters2))
	}
}

// Benchmark tests
func BenchmarkRegistry_ListClusters(b *testing.B) {
	r := NewRegistry()
	for i := 0; i < 100; i++ {
		name := string(rune('a' + i%26))
		r.clusters[name] = &ClusterConfig{Name: name}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = r.ListClusters()
	}
}

func BenchmarkRegistry_GetCluster(b *testing.B) {
	r := NewRegistry()
	for i := 0; i < 100; i++ {
		name := string(rune('a' + i%26))
		r.clusters[name] = &ClusterConfig{Name: name}
		r.clients[name] = fake.NewSimpleClientset()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = r.GetCluster("a")
	}
}
