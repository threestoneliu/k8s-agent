package llm

import (
	"fmt"
	"os"

	k8s "github.com/threestoneliu/k8s-agent/pkg/k8s"
	sharedutil "github.com/threestoneliu/k8s-agent/pkg/shared"
)

// Package-level executor reference
var executor *k8s.Executor

// SetExecutor sets the package-level executor reference
func SetExecutor(e *k8s.Executor) {
	executor = e
}

func init() {
	// Register function handlers
	fmt.Printf("DEBUG: auto_register.init() running, registering %d functions\n", 5)
	RegisterFunction(FunctionDefinition{
		Function: sharedutil.Function{
			Name:        "resource_list",
			Description: "List Kubernetes resources with optional label and field selectors",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"resource": map[string]interface{}{
						"type":        "string",
						"description": "Resource type to list (e.g., pods, services, deployments, nodes)",
					},
					"namespace": map[string]interface{}{
						"type":        "string",
						"description": "Namespace for namespaced resources (optional)",
					},
					"cluster": map[string]interface{}{
						"type":        "string",
						"description": "Cluster name",
					},
					"label_selector": map[string]interface{}{
						"type":        "string",
						"description": "Label selector (e.g., 'app=myapp')",
					},
					"field_selector": map[string]interface{}{
						"type":        "string",
						"description": "Field selector (e.g., 'metadata.name=my-resource')",
					},
				},
				"required": []string{"resource"},
			},
		},
		Handler: resourceListHandler,
	})

	RegisterFunction(FunctionDefinition{
		Function: sharedutil.Function{
			Name:        "resource_get",
			Description: "Get details of a specific Kubernetes resource",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"resource": map[string]interface{}{
						"type":        "string",
						"description": "Resource type (e.g., pod, service, deployment)",
					},
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Resource name",
					},
					"namespace": map[string]interface{}{
						"type":        "string",
						"description": "Namespace (optional for namespaced resources)",
					},
					"cluster": map[string]interface{}{
						"type":        "string",
						"description": "Cluster name",
					},
					"label_selector": map[string]interface{}{
						"type":        "string",
						"description": "Label selector (optional)",
					},
					"field_selector": map[string]interface{}{
						"type":        "string",
						"description": "Field selector (optional)",
					},
				},
				"required": []string{"resource", "name"},
			},
		},
		Handler: resourceGetHandler,
	})

	RegisterFunction(FunctionDefinition{
		Function: sharedutil.Function{
			Name:        "get_apiresources",
			Description: "Get all supported API resource types in the cluster",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"cluster": map[string]interface{}{
						"type":        "string",
						"description": "Cluster name",
					},
				},
			},
		},
		Handler: getAPIResourcesHandler,
	})

	RegisterFunction(FunctionDefinition{
		Function: sharedutil.Function{
			Name:        "use_cluster",
			Description: "Switch the current Kubernetes cluster context",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"cluster": map[string]interface{}{
						"type":        "string",
						"description": "Cluster name to switch to",
					},
				},
				"required": []string{"cluster"},
			},
		},
		Handler: useClusterHandler,
	})

	RegisterFunction(FunctionDefinition{
		Function: sharedutil.Function{
			Name:        "Read",
			Description: "Read the contents of a file from the local filesystem. Use this to read SKILL.md files when you need to follow a skill's workflow.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"file_path": map[string]interface{}{
						"type":        "string",
						"description": "The absolute path to the file to read",
					},
				},
				"required": []string{"file_path"},
			},
		},
		Handler: readFileHandler,
	})
}

func resourceListHandler(args map[string]interface{}, clusterName string) *sharedutil.FunctionResult {
	resource := getStringArg(args, "resource", "")
	if resource == "" {
		return &sharedutil.FunctionResult{Error: "resource type is required", Success: false}
	}

	namespace := getStringArg(args, "namespace", "")
	labelSelector := getStringArg(args, "label_selector", "")
	fieldSelector := getStringArg(args, "field_selector", "")

	if executor == nil {
		return &sharedutil.FunctionResult{Error: "executor not available", Success: false}
	}

	result, err := executor.ListResourcesWithSelectors(clusterName, resource, namespace, labelSelector, fieldSelector)
	if err != nil {
		return &sharedutil.FunctionResult{Error: err.Error(), Success: false}
	}

	if !result.Success {
		return &sharedutil.FunctionResult{Error: result.Output, Success: false}
	}

	return &sharedutil.FunctionResult{Result: result.Output, Success: true}
}

func resourceGetHandler(args map[string]interface{}, clusterName string) *sharedutil.FunctionResult {
	resource := getStringArg(args, "resource", "")
	name := getStringArg(args, "name", "")
	if resource == "" || name == "" {
		return &sharedutil.FunctionResult{Error: "resource type and name are required", Success: false}
	}

	namespace := getStringArg(args, "namespace", "")
	labelSelector := getStringArg(args, "label_selector", "")
	fieldSelector := getStringArg(args, "field_selector", "")

	if executor == nil {
		return &sharedutil.FunctionResult{Error: "executor not available", Success: false}
	}

	result, err := executor.GetResourceWithSelectors(clusterName, resource, name, namespace, labelSelector, fieldSelector)
	if err != nil {
		return &sharedutil.FunctionResult{Error: err.Error(), Success: false}
	}

	if !result.Success {
		return &sharedutil.FunctionResult{Error: result.Output, Success: false}
	}

	return &sharedutil.FunctionResult{Result: result.Output, Success: true}
}

func getAPIResourcesHandler(args map[string]interface{}, clusterName string) *sharedutil.FunctionResult {
	if executor == nil {
		return &sharedutil.FunctionResult{Error: "executor not available", Success: false}
	}

	result, err := executor.GetAPIResources(clusterName)
	if err != nil {
		return &sharedutil.FunctionResult{Error: err.Error(), Success: false}
	}

	if !result.Success {
		return &sharedutil.FunctionResult{Error: result.Output, Success: false}
	}

	return &sharedutil.FunctionResult{Result: result.Output, Success: true}
}

func useClusterHandler(args map[string]interface{}, clusterName string) *sharedutil.FunctionResult {
	targetCluster := getStringArg(args, "cluster", "")
	if targetCluster == "" {
		return &sharedutil.FunctionResult{Error: "cluster name is required", Success: false}
	}

	if executor != nil {
		clusters := executor.ListClusters()
		found := false
		for _, c := range clusters {
			if c == targetCluster {
				found = true
				break
			}
		}
		if !found {
			return &sharedutil.FunctionResult{
				Error:   fmt.Sprintf("cluster '%s' not found. Available clusters: %v", targetCluster, clusters),
				Success: false,
			}
		}
	}

	return &sharedutil.FunctionResult{
		Result:        fmt.Sprintf("Switched to cluster '%s'", targetCluster),
		Success:       true,
		ClusterSwitch: targetCluster,
	}
}

func readFileHandler(args map[string]interface{}, clusterName string) *sharedutil.FunctionResult {
	filePath := getStringArg(args, "file_path", "")
	if filePath == "" {
		return &sharedutil.FunctionResult{Error: "file_path is required", Success: false}
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return &sharedutil.FunctionResult{
			Error:   fmt.Sprintf("failed to read file: %v", err),
			Success: false,
		}
	}

	return &sharedutil.FunctionResult{
		Result: string(content),
		Success: true,
	}
}