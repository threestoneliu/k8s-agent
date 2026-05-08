package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecutor_ExecuteMutation_RequiresConfirmation(t *testing.T) {
	// Mutation operations return confirmation_required without needing a cluster
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
	assert.False(t, result.Success)
	assert.Equal(t, "confirmation_required", result.Output)
	assert.Contains(t, result.Error.Error(), "confirmation required")
	assert.Equal(t, "pod", result.Resource)
	assert.Equal(t, "delete", result.Verb)
}

func TestExecutor_ExecuteMutation_Create(t *testing.T) {
	mockRegistry := new(MockClusterRegistry)
	executor := NewExecutor(mockRegistry)

	op := &ClassifiedOperation{
		Type:      OperationTypeMutation,
		Verb:      "create",
		Resource:  "deployment",
		Name:      "my-app",
		Namespace: "default",
	}

	result, err := executor.Execute("test-cluster", op)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Equal(t, "confirmation_required", result.Output)
	assert.Contains(t, result.Error.Error(), "confirmation required")
}

func TestExecutor_ExecuteMutation_Scale(t *testing.T) {
	mockRegistry := new(MockClusterRegistry)
	executor := NewExecutor(mockRegistry)

	op := &ClassifiedOperation{
		Type:      OperationTypeMutation,
		Verb:      "scale",
		Resource:  "deployment",
		Name:      "my-app",
		Namespace: "default",
	}

	result, err := executor.Execute("test-cluster", op)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Equal(t, "confirmation_required", result.Output)
}

func TestExecutor_ExecuteMutation_Drain(t *testing.T) {
	mockRegistry := new(MockClusterRegistry)
	executor := NewExecutor(mockRegistry)

	op := &ClassifiedOperation{
		Type:      OperationTypeMutation,
		Verb:      "drain",
		Resource:  "node",
		Name:      "worker-1",
	}

	result, err := executor.Execute("test-cluster", op)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Equal(t, "confirmation_required", result.Output)
}

func TestExecutor_ExecuteMutation_Cordon(t *testing.T) {
	mockRegistry := new(MockClusterRegistry)
	executor := NewExecutor(mockRegistry)

	op := &ClassifiedOperation{
		Type:      OperationTypeMutation,
		Verb:      "cordon",
		Resource:  "node",
		Name:      "worker-1",
	}

	result, err := executor.Execute("test-cluster", op)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Equal(t, "confirmation_required", result.Output)
}

func TestExecutor_ExecuteMutation_Patch(t *testing.T) {
	mockRegistry := new(MockClusterRegistry)
	executor := NewExecutor(mockRegistry)

	op := &ClassifiedOperation{
		Type:      OperationTypeMutation,
		Verb:      "patch",
		Resource:  "service",
		Name:      "my-svc",
		Namespace: "default",
	}

	result, err := executor.Execute("test-cluster", op)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Equal(t, "confirmation_required", result.Output)
}

func TestExecutor_Execute_QueryVerbGetPods_Success(t *testing.T) {
	mockRegistry := new(MockClusterRegistry)
	mockRegistry.On("GetCluster", "test-cluster").Return(nil, assert.AnError)

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
	assert.Contains(t, err.Error(), "assert.AnError")
	mockRegistry.AssertExpectations(t)
}

func TestExecutor_Execute_QueryVerbList_Success(t *testing.T) {
	mockRegistry := new(MockClusterRegistry)
	mockRegistry.On("GetCluster", "test-cluster").Return(nil, assert.AnError)

	executor := NewExecutor(mockRegistry)

	op := &ClassifiedOperation{
		Type:      OperationTypeQuery,
		Verb:      "list",
		Resource:  "pods",
		Namespace: "default",
	}

	result, err := executor.Execute("test-cluster", op)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "assert.AnError")
}

func TestExecutor_Execute_QueryVerbDescribe_Success(t *testing.T) {
	mockRegistry := new(MockClusterRegistry)
	mockRegistry.On("GetCluster", "test-cluster").Return(nil, assert.AnError)

	executor := NewExecutor(mockRegistry)

	op := &ClassifiedOperation{
		Type:      OperationTypeQuery,
		Verb:      "describe",
		Resource:  "pod",
		Name:      "nginx",
		Namespace: "default",
	}

	result, err := executor.Execute("test-cluster", op)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "assert.AnError")
}

func TestExecutor_Execute_QueryVerbWatch_Success(t *testing.T) {
	mockRegistry := new(MockClusterRegistry)
	mockRegistry.On("GetCluster", "test-cluster").Return(nil, assert.AnError)

	executor := NewExecutor(mockRegistry)

	op := &ClassifiedOperation{
		Type:      OperationTypeQuery,
		Verb:      "watch",
		Resource:  "pods",
		Namespace: "default",
	}

	result, err := executor.Execute("test-cluster", op)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "assert.AnError")
}

func TestExecutor_Execute_QueryVerbLogs_Success(t *testing.T) {
	mockRegistry := new(MockClusterRegistry)
	mockRegistry.On("GetCluster", "test-cluster").Return(nil, assert.AnError)

	executor := NewExecutor(mockRegistry)

	op := &ClassifiedOperation{
		Type:      OperationTypeQuery,
		Verb:      "logs",
		Resource:  "pod",
		Name:      "nginx",
		Namespace: "default",
	}

	result, err := executor.Execute("test-cluster", op)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "assert.AnError")
}

func TestExecutor_Execute_QueryVerbExec_Success(t *testing.T) {
	mockRegistry := new(MockClusterRegistry)
	mockRegistry.On("GetCluster", "test-cluster").Return(nil, assert.AnError)

	executor := NewExecutor(mockRegistry)

	op := &ClassifiedOperation{
		Type:      OperationTypeQuery,
		Verb:      "exec",
		Resource:  "pod",
		Name:      "nginx",
		Namespace: "default",
	}

	result, err := executor.Execute("test-cluster", op)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "assert.AnError")
}

func TestExecutor_Execute_HighRiskResourceMutation(t *testing.T) {
	mockRegistry := new(MockClusterRegistry)
	executor := NewExecutor(mockRegistry)

	// Even get on nodes is a mutation (high risk)
	op := &ClassifiedOperation{
		Type:      OperationTypeMutation,
		Verb:      "get",
		Resource:  "nodes",
		Name:      "worker-1",
	}

	result, err := executor.Execute("test-cluster", op)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.Success)
	assert.Equal(t, "confirmation_required", result.Output)
}

func TestExecutor_Execute_UnknownResource(t *testing.T) {
	mockRegistry := new(MockClusterRegistry)
	mockRegistry.On("GetCluster", "test-cluster").Return(nil, assert.AnError)

	executor := NewExecutor(mockRegistry)

	op := &ClassifiedOperation{
		Type:      OperationTypeQuery,
		Verb:      "get",
		Resource:  "unknown-resource",
		Name:      "some-name",
	}

	result, err := executor.Execute("test-cluster", op)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "assert.AnError")
}
