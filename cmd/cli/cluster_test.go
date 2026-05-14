package cli

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/threestoneliu/k8s-agent/pkg/cluster"
)

func TestClusterListCommand(t *testing.T) {
	cmd := NewRootCommand()
	listCmd, _, err := cmd.Find([]string{"cluster", "list"})
	require.NoError(t, err)

	assert.Equal(t, "list", listCmd.Name())
	assert.NotNil(t, listCmd.RunE)

	// Test execution
	buf := &bytes.Buffer{}
	listCmd.SetOut(buf)
	listCmd.SetErr(buf)

	// Should not panic
	err = listCmd.RunE(listCmd, []string{})
	assert.NoError(t, err) // No clusters configured initially
}

func TestClusterAddCommand(t *testing.T) {
	cmd := NewRootCommand()
	addCmd, _, err := cmd.Find([]string{"cluster", "add"})
	require.NoError(t, err)

	assert.Equal(t, "add", addCmd.Name())

	// Test with missing args - should error
	buf := &bytes.Buffer{}
	addCmd.SetOut(buf)
	addCmd.SetErr(buf)

	err = addCmd.RunE(addCmd, []string{}) // Missing args
	assert.Error(t, err)
}

func TestClusterUseCommand(t *testing.T) {
	cmd := NewRootCommand()
	useCmd, _, err := cmd.Find([]string{"cluster", "use"})
	require.NoError(t, err)

	assert.Equal(t, "use", useCmd.Name())

	// Test with missing args
	buf := &bytes.Buffer{}
	useCmd.SetOut(buf)
	useCmd.SetErr(buf)

	err = useCmd.RunE(useCmd, []string{}) // Missing cluster name
	assert.Error(t, err)
}

func TestClusterRemoveCommand(t *testing.T) {
	cmd := NewRootCommand()
	removeCmd, _, err := cmd.Find([]string{"cluster", "remove"})
	require.NoError(t, err)

	assert.Equal(t, "remove", removeCmd.Name())

	// Test with missing args
	buf := &bytes.Buffer{}
	removeCmd.SetOut(buf)
	removeCmd.SetErr(buf)

	err = removeCmd.RunE(removeCmd, []string{}) // Missing cluster name
	assert.Error(t, err)
}

// MockClusterRegistry is a mock for testing
type MockClusterRegistry struct {
	clusters map[string]*cluster.ClusterConfig
}

func NewMockClusterRegistry() *MockClusterRegistry {
	return &MockClusterRegistry{
		clusters: make(map[string]*cluster.ClusterConfig),
	}
}

func (m *MockClusterRegistry) AddCluster(name, kubeconfig string) error {
	if name == "" {
		return cluster.ErrInvalidClusterName
	}
	if _, exists := m.clusters[name]; exists {
		return cluster.ErrClusterAlreadyExists
	}
	m.clusters[name] = &cluster.ClusterConfig{
		Name:       name,
		Kubeconfig: kubeconfig,
	}
	return nil
}

func (m *MockClusterRegistry) ListClusters() []*cluster.ClusterConfig {
	result := make([]*cluster.ClusterConfig, 0, len(m.clusters))
	for _, cfg := range m.clusters {
		result = append(result, cfg)
	}
	return result
}

func (m *MockClusterRegistry) RemoveCluster(name string) error {
	if _, exists := m.clusters[name]; !exists {
		return cluster.ErrClusterNotFound
	}
	delete(m.clusters, name)
	return nil
}

func (m *MockClusterRegistry) GetCluster(name string) (interface{}, error) {
	if cfg, ok := m.clusters[name]; ok {
		return cfg, nil
	}
	return nil, cluster.ErrClusterNotFound
}

func TestClusterRegistry_MockAddListRemove(t *testing.T) {
	mock := NewMockClusterRegistry()

	// Add a cluster
	err := mock.AddCluster("test-cluster", "/path/to/kubeconfig")
	require.NoError(t, err)

	// List clusters
	clusters := mock.ListClusters()
	assert.Len(t, clusters, 1)
	assert.Equal(t, "test-cluster", clusters[0].Name)

	// Remove cluster
	err = mock.RemoveCluster("test-cluster")
	require.NoError(t, err)

	// Verify removed
	clusters = mock.ListClusters()
	assert.Len(t, clusters, 0)
}

func TestClusterRegistry_MockDuplicateAdd(t *testing.T) {
	mock := NewMockClusterRegistry()

	// Add a cluster
	err := mock.AddCluster("test-cluster", "/path/to/kubeconfig")
	require.NoError(t, err)

	// Try to add same cluster again
	err = mock.AddCluster("test-cluster", "/path/to/kubeconfig")
	assert.Error(t, err)
	assert.Equal(t, cluster.ErrClusterAlreadyExists, err)
}

func TestClusterRegistry_MockRemoveNonExistent(t *testing.T) {
	mock := NewMockClusterRegistry()

	// Try to remove non-existent cluster
	err := mock.RemoveCluster("non-existent")
	assert.Error(t, err)
	assert.Equal(t, cluster.ErrClusterNotFound, err)
}

func TestClusterRegistry_MockEmptyName(t *testing.T) {
	mock := NewMockClusterRegistry()

	// Try to add cluster with empty name
	err := mock.AddCluster("", "/path/to/kubeconfig")
	assert.Error(t, err)
	assert.Equal(t, cluster.ErrInvalidClusterName, err)
}
