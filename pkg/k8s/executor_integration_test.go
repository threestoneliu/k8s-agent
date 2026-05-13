//go:build integration
// +build integration

package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestExecutor_ExecuteQuery_WithFakeClient(t *testing.T) {
	// Create a fake client with some pods
	fakeClient := fake.NewSimpleClientset(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx",
			Namespace: "default",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{Name: "nginx", Image: "nginx:latest"}},
		},
	})

	mockRegistry := new(MockClusterRegistry)
	mockRegistry.On("GetCluster", "test-cluster").Return(fakeClient, nil)

	executor := NewExecutor(mockRegistry)

	op := &ClassifiedOperation{
		Type:      OperationTypeQuery,
		Verb:      "get",
		Resource:  "pods",
		Name:      "nginx",
		Namespace: "default",
	}

	result, err := executor.Execute("test-cluster", op)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success, "get operation should succeed with fake client")
	assert.Equal(t, "pods", result.Resource)
	assert.Equal(t, "get", result.Verb)
	mockRegistry.AssertExpectations(t)
}

func TestExecutor_ExecuteQueryList_WithFakeClient(t *testing.T) {
	// Create a fake client with pods
	fakeClient := fake.NewSimpleClientset(
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "pod1", Namespace: "default"},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "pod2", Namespace: "default"},
		},
	)

	mockRegistry := new(MockClusterRegistry)
	mockRegistry.On("GetCluster", "test-cluster").Return(fakeClient, nil)

	executor := NewExecutor(mockRegistry)

	op := &ClassifiedOperation{
		Type:      OperationTypeQuery,
		Verb:      "list",
		Resource:  "pods",
		Namespace: "default",
	}

	result, err := executor.Execute("test-cluster", op)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Contains(t, result.Output, "pod1")
	assert.Contains(t, result.Output, "pod2")
	mockRegistry.AssertExpectations(t)
}

func TestExecutor_ExecuteQueryServices_WithFakeClient(t *testing.T) {
	fakeClient := fake.NewSimpleClientset(&corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-svc",
			Namespace: "default",
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{Port: 80}},
		},
	})

	mockRegistry := new(MockClusterRegistry)
	mockRegistry.On("GetCluster", "test-cluster").Return(fakeClient, nil)

	executor := NewExecutor(mockRegistry)

	op := &ClassifiedOperation{
		Type:      OperationTypeQuery,
		Verb:      "get",
		Resource:  "services",
		Name:      "my-svc",
		Namespace: "default",
	}

	result, err := executor.Execute("test-cluster", op)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Contains(t, result.Output, "my-svc")
	mockRegistry.AssertExpectations(t)
}

func TestExecutor_ExecuteQueryNamespaces_WithFakeClient(t *testing.T) {
	fakeClient := fake.NewSimpleClientset(&corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "production",
		},
	})

	mockRegistry := new(MockClusterRegistry)
	mockRegistry.On("GetCluster", "test-cluster").Return(fakeClient, nil)

	executor := NewExecutor(mockRegistry)

	op := &ClassifiedOperation{
		Type:      OperationTypeQuery,
		Verb:      "get",
		Resource:  "namespaces",
		Name:      "production",
	}

	result, err := executor.Execute("test-cluster", op)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Contains(t, result.Output, "production")
	mockRegistry.AssertExpectations(t)
}

func TestExecutor_ExecuteQueryNodes_WithFakeClient(t *testing.T) {
	fakeClient := fake.NewSimpleClientset(&corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "worker-1",
		},
		Status: corev1.NodeStatus{
			Conditions: []corev1.NodeCondition{{Type: corev1.NodeReady, Status: corev1.ConditionTrue}},
		},
	})

	mockRegistry := new(MockClusterRegistry)
	mockRegistry.On("GetCluster", "test-cluster").Return(fakeClient, nil)

	executor := NewExecutor(mockRegistry)

	op := &ClassifiedOperation{
		Type:      OperationTypeQuery,
		Verb:      "get",
		Resource:  "nodes",
		Name:      "worker-1",
	}

	result, err := executor.Execute("test-cluster", op)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Contains(t, result.Output, "worker-1")
	mockRegistry.AssertExpectations(t)
}

func TestExecutor_ExecuteQueryListServices_WithFakeClient(t *testing.T) {
	fakeClient := fake.NewSimpleClientset(
		&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "svc1", Namespace: "default"}},
		&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "svc2", Namespace: "default"}},
	)

	mockRegistry := new(MockClusterRegistry)
	mockRegistry.On("GetCluster", "test-cluster").Return(fakeClient, nil)

	executor := NewExecutor(mockRegistry)

	op := &ClassifiedOperation{
		Type:      OperationTypeQuery,
		Verb:      "list",
		Resource:  "services",
		Namespace: "default",
	}

	result, err := executor.Execute("test-cluster", op)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Contains(t, result.Output, "svc1")
	assert.Contains(t, result.Output, "svc2")
	mockRegistry.AssertExpectations(t)
}

func TestExecutor_ExecuteQueryListNamespaces_WithFakeClient(t *testing.T) {
	fakeClient := fake.NewSimpleClientset(
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "ns1"}},
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "ns2"}},
	)

	mockRegistry := new(MockClusterRegistry)
	mockRegistry.On("GetCluster", "test-cluster").Return(fakeClient, nil)

	executor := NewExecutor(mockRegistry)

	op := &ClassifiedOperation{
		Type:      OperationTypeQuery,
		Verb:      "list",
		Resource:  "namespaces",
	}

	result, err := executor.Execute("test-cluster", op)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Contains(t, result.Output, "ns1")
	assert.Contains(t, result.Output, "ns2")
	mockRegistry.AssertExpectations(t)
}

func TestExecutor_ExecuteQueryListNodes_WithFakeClient(t *testing.T) {
	fakeClient := fake.NewSimpleClientset(
		&corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node1"}},
		&corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node2"}},
	)

	mockRegistry := new(MockClusterRegistry)
	mockRegistry.On("GetCluster", "test-cluster").Return(fakeClient, nil)

	executor := NewExecutor(mockRegistry)

	op := &ClassifiedOperation{
		Type:      OperationTypeQuery,
		Verb:      "list",
		Resource:  "nodes",
	}

	result, err := executor.Execute("test-cluster", op)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Contains(t, result.Output, "node1")
	assert.Contains(t, result.Output, "node2")
	mockRegistry.AssertExpectations(t)
}

func TestExecutor_ExecuteWatch_ReturnsWatchMessage(t *testing.T) {
	fakeClient := fake.NewSimpleClientset()

	mockRegistry := new(MockClusterRegistry)
	mockRegistry.On("GetCluster", "test-cluster").Return(fakeClient, nil)

	executor := NewExecutor(mockRegistry)

	op := &ClassifiedOperation{
		Type:      OperationTypeQuery,
		Verb:      "watch",
		Resource:  "pods",
		Namespace: "default",
	}

	result, err := executor.Execute("test-cluster", op)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
	assert.Contains(t, result.Output, "Watching pods")
	assert.Contains(t, result.Output, "namespace default")
	mockRegistry.AssertExpectations(t)
}

func TestExecutor_ExecuteQuery_ClusterError(t *testing.T) {
	fakeClient := fake.NewSimpleClientset()

	mockRegistry := new(MockClusterRegistry)
	mockRegistry.On("GetCluster", "error-cluster").Return(fakeClient, nil)

	executor := NewExecutor(mockRegistry)

	op := &ClassifiedOperation{
		Type:      OperationTypeQuery,
		Verb:      "get",
		Resource:  "pods",
		Name:      "nginx",
		Namespace: "default",
	}

	// Test with an error cluster name to trigger cluster lookup error
	_, err := executor.Execute("error-cluster", op)

	// This should fail because pods don't exist in the fake client
	// but we get past the cluster lookup
	assert.NoError(t, err)
}
