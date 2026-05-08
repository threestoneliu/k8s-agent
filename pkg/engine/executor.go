package engine

import (
	"context"
	"fmt"

	"k8s.io/client-go/kubernetes"
)

// ClusterRegistryInterface defines the interface for cluster registry operations needed by Executor
type ClusterRegistryInterface interface {
	GetCluster(name string) (kubernetes.Interface, error)
	GetClusterContext(ctx context.Context, name string) (kubernetes.Interface, error)
	ListClusterNames() []string
}

// Executor executes Kubernetes operations
type Executor struct {
	clusterRegistry ClusterRegistryInterface
}

// ExecutionResult holds the result of an operation execution
type ExecutionResult struct {
	Success   bool
	Output    string
	Error     error
	Resource  string
	Verb      string
}

// NewExecutor creates a new Executor
func NewExecutor(clusterRegistry ClusterRegistryInterface) *Executor {
	return &Executor{
		clusterRegistry: clusterRegistry,
	}
}

// Execute performs the operation on the specified cluster
func (e *Executor) Execute(clusterName string, op *ClassifiedOperation) (*ExecutionResult, error) {
	if op == nil {
		return nil, fmt.Errorf("operation is required")
	}

	switch op.Type {
	case OperationTypeQuery:
		return e.executeQuery(clusterName, op)
	case OperationTypeMutation:
		return e.executeMutation(clusterName, op)
	default:
		return nil, fmt.Errorf("unknown operation type: %v", op.Type)
	}
}

// ListClusters returns all registered cluster names
func (e *Executor) ListClusters() []string {
	if e.clusterRegistry == nil {
		return nil
	}
	return e.clusterRegistry.ListClusterNames()
}
