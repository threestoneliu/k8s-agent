package engine

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/threestoneliu/k8s-agent/pkg/cluster"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

// MockClusterRegistry is a mock implementation of ClusterRegistryInterface
type MockClusterRegistry struct {
	mock.Mock
}

func (m *MockClusterRegistry) GetCluster(name string) (kubernetes.Interface, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(kubernetes.Interface), args.Error(1)
}

func (m *MockClusterRegistry) GetClusterContext(ctx context.Context, name string) (kubernetes.Interface, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(kubernetes.Interface), args.Error(1)
}

func (m *MockClusterRegistry) GetDynamicCluster(name string) (dynamic.Interface, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(dynamic.Interface), args.Error(1)
}

func (m *MockClusterRegistry) ListClusterNames() []string {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]string)
}

func (m *MockClusterRegistry) GetResourceCache() *cluster.ResourceCache {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*cluster.ResourceCache)
}

func TestExecutor_Execute_UnknownOperationType(t *testing.T) {
	mockRegistry := new(MockClusterRegistry)
	executor := NewExecutor(mockRegistry)

	op := &ClassifiedOperation{
		Type:     OperationTypeUnknown,
		Verb:     "unknown",
		Resource: "pod",
	}

	result, err := executor.Execute("test-cluster", op)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "unsupported operation type")
}

func TestExecutor_Execute_NilOperation(t *testing.T) {
	mockRegistry := new(MockClusterRegistry)
	executor := NewExecutor(mockRegistry)

	result, err := executor.Execute("test-cluster", nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "operation is required")
}
