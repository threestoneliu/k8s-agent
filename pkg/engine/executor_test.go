package engine

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

func (m *MockClusterRegistry) ListClusterNames() []string {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]string)
}

func TestExecutor_Execute_RoutesToQuery(t *testing.T) {
	mockRegistry := new(MockClusterRegistry)
	// For query operations, we need a working kubernetes client
	// Since the actual client is hard to mock, we test the routing logic
	mockRegistry.On("GetCluster", "test-cluster").Return(nil, errors.New("cluster not configured for test"))

	executor := NewExecutor(mockRegistry)

	op := &ClassifiedOperation{
		Type:      OperationTypeQuery,
		Verb:      "get",
		Resource:  "pods",
		Name:      "nginx",
		Namespace: "default",
	}

	result, err := executor.Execute("test-cluster", op)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "cluster not configured for test")
	mockRegistry.AssertExpectations(t)
}

func TestExecutor_Execute_RoutesToMutation(t *testing.T) {
	mockRegistry := new(MockClusterRegistry)
	executor := NewExecutor(mockRegistry)

	op := &ClassifiedOperation{
		Type:      OperationTypeMutation,
		Verb:      "delete",
		Resource:  "pod",
		Name:      "nginx",
		Namespace: "default",
	}

	result, err := executor.Execute("test-cluster", op)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Success, "mutation operations should require confirmation")
	assert.Equal(t, "confirmation_required", result.Output)
	assert.Contains(t, result.Error.Error(), "confirmation required")
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
	assert.Contains(t, err.Error(), "unknown operation type")
}

func TestExecutor_Execute_ClusterNotFound(t *testing.T) {
	mockRegistry := new(MockClusterRegistry)
	mockRegistry.On("GetCluster", "nonexistent").
		Return(nil, errors.New("cluster not found"))

	executor := NewExecutor(mockRegistry)

	op := &ClassifiedOperation{
		Type:     OperationTypeQuery,
		Verb:     "get",
		Resource: "pods",
	}

	result, err := executor.Execute("nonexistent", op)

	assert.Error(t, err)
	assert.Nil(t, result)
	mockRegistry.AssertExpectations(t)
}

func TestExecutor_Execute_NilOperation(t *testing.T) {
	mockRegistry := new(MockClusterRegistry)
	executor := NewExecutor(mockRegistry)

	result, err := executor.Execute("test-cluster", nil)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "operation is required")
}
