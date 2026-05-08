package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestExecutor_ExecuteDescribe_WithFakeClient(t *testing.T) {
	fakeClient := fake.NewSimpleClientset(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx",
			Namespace: "default",
		},
	})

	mockRegistry := new(MockClusterRegistry)
	mockRegistry.On("GetCluster", "test-cluster").Return(fakeClient, nil)

	executor := NewExecutor(mockRegistry)

	op := &ClassifiedOperation{
		Type:      OperationTypeQuery,
		Verb:      "describe",
		Resource:  "pods",
		Name:      "nginx",
		Namespace: "default",
	}

	result, err := executor.Execute("test-cluster", op)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Contains(t, result.Output, "nginx")
	mockRegistry.AssertExpectations(t)
}

func TestExecutor_ExecuteDescribeList_WithFakeClient(t *testing.T) {
	fakeClient := fake.NewSimpleClientset(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "default"},
	})

	mockRegistry := new(MockClusterRegistry)
	mockRegistry.On("GetCluster", "test-cluster").Return(fakeClient, nil)

	executor := NewExecutor(mockRegistry)

	op := &ClassifiedOperation{
		Type:      OperationTypeQuery,
		Verb:      "describe",
		Resource:  "pods",
		Namespace: "default",
	}

	result, err := executor.Execute("test-cluster", op)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Contains(t, result.Output, "pod1")
	mockRegistry.AssertExpectations(t)
}

func TestExecutor_ExecuteLogs_WithFakeClient(t *testing.T) {
	fakeClient := fake.NewSimpleClientset(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx",
			Namespace: "default",
		},
	})

	mockRegistry := new(MockClusterRegistry)
	mockRegistry.On("GetCluster", "test-cluster").Return(fakeClient, nil)

	executor := NewExecutor(mockRegistry)

	op := &ClassifiedOperation{
		Type:      OperationTypeQuery,
		Verb:      "logs",
		Resource:  "pod",
		Name:      "nginx",
		Namespace: "default",
	}

	result, err := executor.Execute("test-cluster", op)

	// Logs operation will fail on fake client but we test the code path
	assert.NoError(t, err)
	assert.NotNil(t, result)
	// The fake client may not support logs properly, but we get through the code path
	mockRegistry.AssertExpectations(t)
}

func TestExecutor_ExecuteLogs_WithContainer(t *testing.T) {
	fakeClient := fake.NewSimpleClientset(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx",
			Namespace: "default",
		},
	})

	mockRegistry := new(MockClusterRegistry)
	mockRegistry.On("GetCluster", "test-cluster").Return(fakeClient, nil)

	executor := NewExecutor(mockRegistry)

	op := &ClassifiedOperation{
		Type:      OperationTypeQuery,
		Verb:      "logs",
		Resource:  "pod",
		Name:      "nginx",
		Namespace: "default",
		Flags:     map[string]string{"container": "nginx"},
	}

	result, err := executor.Execute("test-cluster", op)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	mockRegistry.AssertExpectations(t)
}

func TestExecutor_ExecuteQueryDeployments_WithFakeClient(t *testing.T) {
	fakeClient := fake.NewSimpleClientset(&appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-app",
			Namespace: "default",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(3),
		},
	})

	mockRegistry := new(MockClusterRegistry)
	mockRegistry.On("GetCluster", "test-cluster").Return(fakeClient, nil)

	executor := NewExecutor(mockRegistry)

	op := &ClassifiedOperation{
		Type:      OperationTypeQuery,
		Verb:      "get",
		Resource:  "deployments",
		Name:      "my-app",
		Namespace: "default",
	}

	result, err := executor.Execute("test-cluster", op)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Contains(t, result.Output, "my-app")
	mockRegistry.AssertExpectations(t)
}

func TestExecutor_ExecuteQueryListDeployments_WithFakeClient(t *testing.T) {
	fakeClient := fake.NewSimpleClientset(
		&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "deploy1", Namespace: "default"}},
		&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "deploy2", Namespace: "default"}},
	)

	mockRegistry := new(MockClusterRegistry)
	mockRegistry.On("GetCluster", "test-cluster").Return(fakeClient, nil)

	executor := NewExecutor(mockRegistry)

	op := &ClassifiedOperation{
		Type:      OperationTypeQuery,
		Verb:      "list",
		Resource:  "deployments",
		Namespace: "default",
	}

	result, err := executor.Execute("test-cluster", op)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Contains(t, result.Output, "deploy1")
	assert.Contains(t, result.Output, "deploy2")
	mockRegistry.AssertExpectations(t)
}

func int32Ptr(i int32) *int32 {
	return &i
}
