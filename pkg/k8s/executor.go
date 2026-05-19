package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/threestoneliu/k8s-agent/pkg/cluster"
	"github.com/threestoneliu/k8s-agent/pkg/log"
)

// ClusterRegistryInterface defines the interface for cluster registry operations needed by Executor
type ClusterRegistryInterface interface {
	GetCluster(name string) (kubernetes.Interface, error)
	GetClusterContext(ctx context.Context, name string) (kubernetes.Interface, error)
	GetDynamicCluster(name string) (dynamic.Interface, error)
	GetRESTClient(name string, gvr schema.GroupVersionResource) (*rest.RESTClient, error)
	ListClusterNames() []string
	GetResourceCache() *cluster.ResourceCache
}

// Executor executes Kubernetes operations
type Executor struct {
	clusterRegistry ClusterRegistryInterface
	resourceCache   *cluster.ResourceCache
}

// ExecutionResult holds the result of an operation execution
type ExecutionResult struct {
	Success   bool
	Output    string
	Error     error
	Resource  string
	Verb      string
}

// OperationType represents the type of operation
type OperationType int

const (
	OperationTypeQuery OperationType = iota
	OperationTypeMutation
	OperationTypeUnknown
)

// ClassifiedOperation represents an operation with its classification
type ClassifiedOperation struct {
	Type      OperationType
	Verb      string
	Resource  string
	Name      string
	Namespace string
	Flags     map[string]string
	RawInput  string
}

// NewExecutor creates a new Executor
func NewExecutor(clusterRegistry ClusterRegistryInterface) *Executor {
	return &Executor{
		clusterRegistry: clusterRegistry,
	}
}

// getResourceCache returns the resource cache, initializing it lazily if needed
func (e *Executor) getResourceCache() *cluster.ResourceCache {
	if e.resourceCache == nil && e.clusterRegistry != nil {
		e.resourceCache = e.clusterRegistry.GetResourceCache()
	}
	return e.resourceCache
}

// Execute performs the operation on the specified cluster
func (e *Executor) Execute(clusterName string, op *ClassifiedOperation) (*ExecutionResult, error) {
	if op == nil {
		return nil, fmt.Errorf("operation is required")
	}

	return nil, fmt.Errorf("unsupported operation type: %v (use ListResourcesWithSelectors or GetResourceWithSelectors for queries)", op.Type)
}

// ListClusters returns all registered cluster names
func (e *Executor) ListClusters() []string {
	if e.clusterRegistry == nil {
		return nil
	}
	return e.clusterRegistry.ListClusterNames()
}

// normalizeResource maps various resource name forms to canonical form using dynamic discovery
// It first tries exact match, then case-insensitive match, then discovers singular/plural variations
func (e *Executor) normalizeResource(resource string) string {
	cache := e.getResourceCache()
	if cache == nil {
		// Fallback to lowercase if cache not available
		return strings.ToLower(resource)
	}

	clusterNames := e.clusterRegistry.ListClusterNames()
	if len(clusterNames) == 0 {
		return strings.ToLower(resource)
	}
	clusterName := clusterNames[0]

	// First try exact match (the cache already indexes by plural form)
	if gvr, ok := cache.GetGVR(clusterName, resource); ok {
		return gvr.Resource
	}

	// Try case-insensitive match
	lowerResource := strings.ToLower(resource)
	if gvr, ok := cache.GetGVR(clusterName, lowerResource); ok {
		return gvr.Resource
	}

	// Try singular/plural variations by checking all known resources
	apiResources := cache.GetAPIResources(clusterName)
	if apiResources == nil {
		return lowerResource
	}

	// Build a map of lowercase resource names for matching
	lowerToCanonical := make(map[string]string)
	for _, ar := range apiResources {
		lowerName := strings.ToLower(ar.Name)
		if lowerToCanonical[lowerName] == "" {
			lowerToCanonical[lowerName] = ar.Name
		}
		// Also index by singular form (remove trailing 's')
		if strings.HasSuffix(lowerName, "s") && len(lowerName) > 1 {
			singular := strings.TrimSuffix(lowerName, "s")
			if lowerToCanonical[singular] == "" {
				lowerToCanonical[singular] = ar.Name
			}
		}
		// Index by singular form without 's'
		if !strings.HasSuffix(lowerName, "s") {
			plural := lowerName + "s"
			if lowerToCanonical[plural] == "" {
				lowerToCanonical[plural] = ar.Name
			}
		}
	}

	// Try to find matching resource
	if canonical, ok := lowerToCanonical[lowerResource]; ok {
		return canonical
	}

	// No match found, return lowercase input as last resort
	return lowerResource
}

// ListResourcesWithSelectors lists Kubernetes resources with label and field selectors, returning table format
func (e *Executor) ListResourcesWithSelectors(clusterName, resource, namespace, labelSelector, fieldSelector string) (*ExecutionResult, error) {
	// Delegate to ListResourcesAsTable for table output
	return e.ListResourcesAsTable(clusterName, resource, namespace, labelSelector, fieldSelector)
}

// GetResourceWithSelectors gets a specific Kubernetes resource with label and field selectors using dynamic client
func (e *Executor) GetResourceWithSelectors(clusterName, resource, name, namespace, labelSelector, fieldSelector string) (*ExecutionResult, error) {
	log.Debug("GetResourceWithSelectors: START", "cluster", clusterName, "resource", resource, "name", name)

	if name == "" {
		return &ExecutionResult{Success: false, Output: "resource name is required"}, nil
	}

	// Normalize resource name
	resource = e.normalizeResource(resource)

	dynClient, err := e.clusterRegistry.GetDynamicCluster(clusterName)
	if err != nil {
		return &ExecutionResult{Success: false, Output: fmt.Sprintf("failed to get cluster %s: %v", clusterName, err)}, nil
	}

	k8sClient, err := e.clusterRegistry.GetCluster(clusterName)
	if err != nil {
		return &ExecutionResult{Success: false, Output: fmt.Sprintf("failed to get cluster %s: %v", clusterName, err)}, nil
	}

	if err := e.getResourceCache().LazyRefresh(clusterName, k8sClient); err != nil {
		return &ExecutionResult{Success: false, Output: fmt.Sprintf("failed to refresh resource cache: %v", err)}, nil
	}

	gvr, ok := e.getResourceCache().GetGVR(clusterName, resource)
	if !ok {
		return &ExecutionResult{Success: false, Output: fmt.Sprintf("unsupported resource type: %s", resource)}, nil
	}

	ctx := context.Background()
	var item *unstructured.Unstructured

	if e.getResourceCache().IsNamespaced(clusterName, resource) {
		if namespace == "" {
			namespace = "default"
		}
		item, err = dynClient.Resource(gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	} else {
		item, err = dynClient.Resource(gvr).Get(ctx, name, metav1.GetOptions{})
	}

	if err != nil {
		return &ExecutionResult{Success: false, Output: err.Error()}, nil
	}

	jsonData, _ := json.MarshalIndent(item.Object, "", "  ")
	return &ExecutionResult{Success: true, Output: string(jsonData), Resource: resource, Verb: "get"}, nil
}

// GetAPIResources returns all supported API resource types in the cluster using DiscoveryClient
func (e *Executor) GetAPIResources(clusterName string) (*ExecutionResult, error) {
	client, err := e.clusterRegistry.GetCluster(clusterName)
	if err != nil {
		return &ExecutionResult{Success: false, Output: fmt.Sprintf("failed to get cluster %s: %v", clusterName, err)}, nil
	}

	k8sClient, ok := client.(kubernetes.Interface)
	if !ok {
		return &ExecutionResult{Success: false, Output: "invalid kubernetes client type"}, nil
	}

	if err := e.getResourceCache().LazyRefresh(clusterName, k8sClient); err != nil {
		return &ExecutionResult{Success: false, Output: fmt.Sprintf("failed to get API resources: %v", err)}, nil
	}

	apiResources := e.getResourceCache().GetAPIResources(clusterName)
	output := fmt.Sprintf("Found %d API resources", len(apiResources))
	return &ExecutionResult{Success: true, Output: output, Resource: "apiresources", Verb: "get"}, nil
}

// ListResourcesAsTable lists Kubernetes resources and returns a formatted table using RESTClient with Table accept header
func (e *Executor) ListResourcesAsTable(clusterName, resource, namespace, labelSelector, fieldSelector string) (*ExecutionResult, error) {
	log.Debug("ListResourcesAsTable: START", "cluster", clusterName, "resource", resource)

	// Normalize resource name (singular to plural)
	resource = e.normalizeResource(resource)

	k8sClient, err := e.clusterRegistry.GetCluster(clusterName)
	if err != nil {
		return &ExecutionResult{Success: false, Output: fmt.Sprintf("failed to get cluster %s: %v", clusterName, err)}, nil
	}

	if err := e.getResourceCache().LazyRefresh(clusterName, k8sClient); err != nil {
		return &ExecutionResult{Success: false, Output: fmt.Sprintf("failed to refresh resource cache: %v", err)}, nil
	}

	cache := e.getResourceCache()
	gvr, ok := cache.GetGVR(clusterName, resource)
	if !ok {
		log.Warn("GetGVR failed for resource %s", resource)
		return &ExecutionResult{Success: false, Output: fmt.Sprintf("unsupported resource type: %s", resource)}, nil
	}
	groupStr := gvr.Group
	if groupStr == "" {
		groupStr = "(core)"
	}
	log.Debug("GVR details: Group=%s, Version=%s, Resource=%s", groupStr, gvr.Version, gvr.Resource)

	isNs := cache.IsNamespaced(clusterName, resource)
	gvrStr := fmt.Sprintf("Group=%s, Version=%s, Resource=%s", groupStr, gvr.Version, gvr.Resource)
	log.Debug("ListResourcesAsTable: gvr=%s, isNamespaced=%v", gvrStr, isNs)

	restClient, err := e.clusterRegistry.GetRESTClient(clusterName, gvr)
	if err != nil {
		return &ExecutionResult{Success: false, Output: fmt.Sprintf("failed to get REST client: %v", err)}, nil
	}

	// Build the request path
	var path string
	if isNs {
		if namespace == "" {
			namespace = "default"
		}
		if gvr.Group == "" {
			path = fmt.Sprintf("/api/%s/namespaces/%s/%s", gvr.Version, namespace, gvr.Resource)
		} else {
			path = fmt.Sprintf("/apis/%s/%s/namespaces/%s/%s", gvr.Group, gvr.Version, namespace, gvr.Resource)
		}
	} else {
		if gvr.Group == "" {
			path = fmt.Sprintf("/api/%s/%s", gvr.Version, gvr.Resource)
		} else {
			path = fmt.Sprintf("/apis/%s/%s/%s", gvr.Group, gvr.Version, gvr.Resource)
		}
	}

	log.Debug("ListResourcesAsTable: path=%s", path)

	// Execute request with Table accept header
	table := &metav1.Table{}
	err = restClient.Get().
		AbsPath(path).
		SetHeader("Accept", "application/json;as=Table;v=v1;g=meta.k8s.io").
		Param("labelSelector", labelSelector).
		Param("fieldSelector", fieldSelector).
		Do(context.Background()).
		Into(table)

	if err != nil {
		log.Error("ListResourcesAsTable: REST client call failed", "error", err, "path", path)
		return &ExecutionResult{Success: false, Output: err.Error()}, nil
	}

	// Format the table output similar to kubectl get
	output := formatTable(table)
	return &ExecutionResult{Success: true, Output: output, Resource: resource, Verb: "list"}, nil
}

// formatTable formats a metav1.Table into a kubectl-style string
func formatTable(table *metav1.Table) string {
	if table == nil {
		return ""
	}

	var sb strings.Builder

	// Print column headers
	if len(table.ColumnDefinitions) > 0 {
		headers := make([]string, len(table.ColumnDefinitions))
		for i, col := range table.ColumnDefinitions {
			headers[i] = col.Name
		}
		sb.WriteString(strings.Join(headers, " "))
		sb.WriteString("\n")
	}

	// Print rows
	for _, row := range table.Rows {
		cells := make([]string, len(row.Cells))
		for i, cell := range row.Cells {
			cells[i] = fmt.Sprintf("%v", cell)
		}
		sb.WriteString(strings.Join(cells, " "))
		sb.WriteString("\n")
	}

	return sb.String()
}