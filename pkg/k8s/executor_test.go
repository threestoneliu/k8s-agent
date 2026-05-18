package engine

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/threestoneliu/k8s-agent/pkg/cluster"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
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

func (m *MockClusterRegistry) GetRESTClient(name string, gvr schema.GroupVersionResource) (*rest.RESTClient, error) {
	args := m.Called(name, gvr)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*rest.RESTClient), args.Error(1)
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

func TestFormatTable(t *testing.T) {
	tests := []struct {
		name     string
		table    *metav1.Table
		expected string
	}{
		{
			name:     "nil table",
			table:    nil,
			expected: "",
		},
		{
			name:     "empty table",
			table:    &metav1.Table{},
			expected: "",
		},
		{
			name: "table with headers but no rows",
			table: &metav1.Table{
				ColumnDefinitions: []metav1.TableColumnDefinition{
					{Name: "NAME", Type: "string"},
					{Name: "READY", Type: "string"},
					{Name: "STATUS", Type: "string"},
					{Name: "AGE", Type: "string"},
				},
				Rows: []metav1.TableRow{},
			},
			expected: "NAME READY STATUS AGE\n",
		},
		{
			name: "table with headers and rows",
			table: &metav1.Table{
				ColumnDefinitions: []metav1.TableColumnDefinition{
					{Name: "NAME", Type: "string"},
					{Name: "READY", Type: "string"},
					{Name: "STATUS", Type: "string"},
					{Name: "AGE", Type: "string"},
				},
				Rows: []metav1.TableRow{
					{Cells: []interface{}{"pod-1", "1/1", "Running", "1d"}},
					{Cells: []interface{}{"pod-2", "0/1", "Pending", "2d"}},
				},
			},
			expected: "NAME READY STATUS AGE\npod-1 1/1 Running 1d\npod-2 0/1 Pending 2d\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTable(tt.table)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStringPointer(t *testing.T) {
	// Test is covered implicitly via the main functionality tests
	// stringPointer was removed in favor of direct Param calls
}
