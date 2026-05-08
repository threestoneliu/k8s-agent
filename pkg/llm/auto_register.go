package llm

import (
	"encoding/json"
	"fmt"
	"strings"

	"k8s-agent/pkg/engine"
	"k8s-agent/pkg/scheduler"
)

func init() {
	RegisterFunction(FunctionDefinition{
		Function: Function{
			Name:        "list_pods",
			Description: "List pods in a namespace",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"namespace": map[string]interface{}{
						"type":    "string",
						"default": "default",
					},
					"cluster": map[string]interface{}{
						"type":        "string",
						"description": "Cluster name",
					},
				},
			},
		},
		Handler: listPodsHandler,
	})

	RegisterFunction(FunctionDefinition{
		Function: Function{
			Name:        "get_pod",
			Description: "Get details of a specific pod",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Pod name",
					},
					"namespace": map[string]interface{}{
						"type":    "string",
						"default": "default",
					},
					"cluster": map[string]interface{}{
						"type":        "string",
						"description": "Cluster name",
					},
				},
				"required": []string{"name"},
			},
		},
		Handler: getPodHandler,
	})

	RegisterFunction(FunctionDefinition{
		Function: Function{
			Name:        "list_services",
			Description: "List services in a namespace",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"namespace": map[string]interface{}{
						"type":    "string",
						"default": "default",
					},
					"cluster": map[string]interface{}{
						"type":        "string",
						"description": "Cluster name",
					},
				},
			},
		},
		Handler: listServicesHandler,
	})

	RegisterFunction(FunctionDefinition{
		Function: Function{
			Name:        "list_deployments",
			Description: "List deployments in a namespace",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"namespace": map[string]interface{}{
						"type":    "string",
						"default": "default",
					},
					"cluster": map[string]interface{}{
						"type":        "string",
						"description": "Cluster name",
					},
				},
			},
		},
		Handler: listDeploymentsHandler,
	})

	RegisterFunction(FunctionDefinition{
		Function: Function{
			Name:        "list_nodes",
			Description: "List all nodes in the cluster",
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
		Handler: listNodesHandler,
	})

	RegisterFunction(FunctionDefinition{
		Function: Function{
			Name:        "describe_pod",
			Description: "Get detailed information about a pod including events and conditions",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Pod name",
					},
					"namespace": map[string]interface{}{
						"type":    "string",
						"default": "default",
					},
					"cluster": map[string]interface{}{
						"type":        "string",
						"description": "Cluster name",
					},
				},
				"required": []string{"name"},
			},
		},
		Handler: describePodHandler,
	})

	RegisterFunction(FunctionDefinition{
		Function: Function{
			Name:        "describe_service",
			Description: "Get detailed information about a service",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Service name",
					},
					"namespace": map[string]interface{}{
						"type":    "string",
						"default": "default",
					},
					"cluster": map[string]interface{}{
						"type":        "string",
						"description": "Cluster name",
					},
				},
				"required": []string{"name"},
			},
		},
		Handler: describeServiceHandler,
	})

	RegisterFunction(FunctionDefinition{
		Function: Function{
			Name:        "describe_deployment",
			Description: "Get detailed information about a deployment",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Deployment name",
					},
					"namespace": map[string]interface{}{
						"type":    "string",
						"default": "default",
					},
					"cluster": map[string]interface{}{
						"type":        "string",
						"description": "Cluster name",
					},
				},
				"required": []string{"name"},
			},
		},
		Handler: describeDeploymentHandler,
	})

	RegisterFunction(FunctionDefinition{
		Function: Function{
			Name:        "get_node_status",
			Description: "Get status information about a node",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"node_name": map[string]interface{}{
						"type":        "string",
						"description": "Node name",
					},
					"cluster": map[string]interface{}{
						"type":        "string",
						"description": "Cluster name",
					},
				},
				"required": []string{"node_name"},
			},
		},
		Handler: getNodeStatusHandler,
	})

	RegisterFunction(FunctionDefinition{
		Function: Function{
			Name:        "list_namespaces",
			Description: "List all namespaces in the cluster",
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
		Handler: listNamespacesHandler,
	})

	RegisterFunction(FunctionDefinition{
		Function: Function{
			Name:        "create_deployment",
			Description: "Create a new deployment (requires confirmation)",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Deployment name",
					},
					"image": map[string]interface{}{
						"type":        "string",
						"description": "Container image",
					},
					"namespace": map[string]interface{}{
						"type":    "string",
						"default": "default",
					},
					"replicas": map[string]interface{}{
						"type":        "integer",
						"description": "Number of replicas",
						"default":     1,
					},
					"cluster": map[string]interface{}{
						"type":        "string",
						"description": "Cluster name",
					},
				},
				"required": []string{"name", "image"},
			},
		},
		Handler: createDeploymentHandler,
	})

	RegisterFunction(FunctionDefinition{
		Function: Function{
			Name:        "delete_pod",
			Description: "Delete a pod (requires confirmation)",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Pod name",
					},
					"namespace": map[string]interface{}{
						"type":    "string",
						"default": "default",
					},
					"cluster": map[string]interface{}{
						"type":        "string",
						"description": "Cluster name",
					},
				},
				"required": []string{"name"},
			},
		},
		Handler: deletePodHandler,
	})

	RegisterFunction(FunctionDefinition{
		Function: Function{
			Name:        "scale_deployment",
			Description: "Scale a deployment to a specific number of replicas (requires confirmation)",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Deployment name",
					},
					"replicas": map[string]interface{}{
						"type":        "integer",
						"description": "Number of replicas",
					},
					"namespace": map[string]interface{}{
						"type":    "string",
						"default": "default",
					},
					"cluster": map[string]interface{}{
						"type":        "string",
						"description": "Cluster name",
					},
				},
				"required": []string{"name", "replicas"},
			},
		},
		Handler: scaleDeploymentHandler,
	})

	RegisterFunction(FunctionDefinition{
		Function: Function{
			Name:        "update_configmap",
			Description: "Update a ConfigMap (requires confirmation)",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "ConfigMap name",
					},
					"data": map[string]interface{}{
						"type":        "object",
						"description": "ConfigMap data key-value pairs",
					},
					"namespace": map[string]interface{}{
						"type":    "string",
						"default": "default",
					},
					"cluster": map[string]interface{}{
						"type":        "string",
						"description": "Cluster name",
					},
				},
				"required": []string{"name", "data"},
			},
		},
		Handler: updateConfigmapHandler,
	})

	RegisterFunction(FunctionDefinition{
		Function: Function{
			Name:        "create_scheduled_task",
			Description: "Create a scheduled task for periodic Kubernetes operations",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Task name (unique identifier)",
					},
					"schedule": map[string]interface{}{
						"type":        "string",
						"description": "Cron expression (e.g., '*/5 * * * *' for every 5 minutes, '0 * * * *' for hourly, '0 0 * * *' for daily)",
					},
					"operation": map[string]interface{}{
						"type":        "string",
						"description": "Operation to execute: list_pods, list_services, list_deployments, list_nodes, describe_pod, etc.",
					},
					"namespace": map[string]interface{}{
						"type":        "string",
						"description": "Namespace for the operation (optional, default: default)",
					},
					"cluster": map[string]interface{}{
						"type":        "string",
						"description": "Cluster name",
					},
				},
				"required": []string{"name", "schedule", "operation"},
			},
		},
		Handler: createScheduledTaskHandler,
	})

	RegisterFunction(FunctionDefinition{
		Function: Function{
			Name:        "list_scheduled_tasks",
			Description: "List all scheduled tasks",
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
		Handler: listScheduledTasksHandler,
	})

	RegisterFunction(FunctionDefinition{
		Function: Function{
			Name:        "delete_scheduled_task",
			Description: "Delete a scheduled task by name",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Task name to delete",
					},
				},
				"required": []string{"name"},
			},
		},
		Handler: deleteScheduledTaskHandler,
	})

	RegisterFunction(FunctionDefinition{
		Function: Function{
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
}

// Handler functions

func listPodsHandler(args map[string]interface{}, clusterName string) *FunctionCallResult {
	namespace := getStringArg(args, "namespace", "default")
	op := &engine.ParsedOperation{
		Verb:      "list",
		Resource:  "pods",
		Namespace: namespace,
	}
	classified := engine.ClassifyOperation(op)
	result, err := getExecutor().Execute(clusterName, classified)
	if err != nil {
		return &FunctionCallResult{Error: err.Error(), Success: false}
	}
	return &FunctionCallResult{Result: result.Output, Success: result.Success}
}

func getPodHandler(args map[string]interface{}, clusterName string) *FunctionCallResult {
	name := getStringArg(args, "name", "")
	namespace := getStringArg(args, "namespace", "default")
	if name == "" {
		return &FunctionCallResult{Error: "name is required", Success: false}
	}
	op := &engine.ParsedOperation{
		Verb:      "get",
		Resource:  "pod",
		Name:      name,
		Namespace: namespace,
	}
	classified := engine.ClassifyOperation(op)
	result, err := getExecutor().Execute(clusterName, classified)
	if err != nil {
		return &FunctionCallResult{Error: err.Error(), Success: false}
	}
	return &FunctionCallResult{Result: result.Output, Success: result.Success}
}

func listServicesHandler(args map[string]interface{}, clusterName string) *FunctionCallResult {
	namespace := getStringArg(args, "namespace", "default")
	op := &engine.ParsedOperation{
		Verb:      "list",
		Resource:  "services",
		Namespace: namespace,
	}
	classified := engine.ClassifyOperation(op)
	result, err := getExecutor().Execute(clusterName, classified)
	if err != nil {
		return &FunctionCallResult{Error: err.Error(), Success: false}
	}
	return &FunctionCallResult{Result: result.Output, Success: result.Success}
}

func listDeploymentsHandler(args map[string]interface{}, clusterName string) *FunctionCallResult {
	namespace := getStringArg(args, "namespace", "default")
	op := &engine.ParsedOperation{
		Verb:      "list",
		Resource:  "deployments",
		Namespace: namespace,
	}
	classified := engine.ClassifyOperation(op)
	result, err := getExecutor().Execute(clusterName, classified)
	if err != nil {
		return &FunctionCallResult{Error: err.Error(), Success: false}
	}
	return &FunctionCallResult{Result: result.Output, Success: result.Success}
}

func listNodesHandler(args map[string]interface{}, clusterName string) *FunctionCallResult {
	op := &engine.ParsedOperation{
		Verb:     "list",
		Resource: "nodes",
	}
	classified := engine.ClassifyOperation(op)
	result, err := getExecutor().Execute(clusterName, classified)
	if err != nil {
		return &FunctionCallResult{Error: err.Error(), Success: false}
	}
	return &FunctionCallResult{Result: result.Output, Success: result.Success}
}

func describePodHandler(args map[string]interface{}, clusterName string) *FunctionCallResult {
	name := getStringArg(args, "name", "")
	namespace := getStringArg(args, "namespace", "default")
	if name == "" {
		return &FunctionCallResult{Error: "name is required", Success: false}
	}
	op := &engine.ParsedOperation{
		Verb:      "describe",
		Resource:  "pod",
		Name:      name,
		Namespace: namespace,
	}
	classified := engine.ClassifyOperation(op)
	result, err := getExecutor().Execute(clusterName, classified)
	if err != nil {
		return &FunctionCallResult{Error: err.Error(), Success: false}
	}
	return &FunctionCallResult{Result: result.Output, Success: result.Success}
}

func describeServiceHandler(args map[string]interface{}, clusterName string) *FunctionCallResult {
	name := getStringArg(args, "name", "")
	namespace := getStringArg(args, "namespace", "default")
	if name == "" {
		return &FunctionCallResult{Error: "name is required", Success: false}
	}
	op := &engine.ParsedOperation{
		Verb:      "describe",
		Resource:  "service",
		Name:      name,
		Namespace: namespace,
	}
	classified := engine.ClassifyOperation(op)
	result, err := getExecutor().Execute(clusterName, classified)
	if err != nil {
		return &FunctionCallResult{Error: err.Error(), Success: false}
	}
	return &FunctionCallResult{Result: result.Output, Success: result.Success}
}

func describeDeploymentHandler(args map[string]interface{}, clusterName string) *FunctionCallResult {
	name := getStringArg(args, "name", "")
	namespace := getStringArg(args, "namespace", "default")
	if name == "" {
		return &FunctionCallResult{Error: "name is required", Success: false}
	}
	op := &engine.ParsedOperation{
		Verb:      "describe",
		Resource:  "deployment",
		Name:      name,
		Namespace: namespace,
	}
	classified := engine.ClassifyOperation(op)
	result, err := getExecutor().Execute(clusterName, classified)
	if err != nil {
		return &FunctionCallResult{Error: err.Error(), Success: false}
	}
	return &FunctionCallResult{Result: result.Output, Success: result.Success}
}

func getNodeStatusHandler(args map[string]interface{}, clusterName string) *FunctionCallResult {
	nodeName := getStringArg(args, "node_name", "")
	if nodeName == "" {
		return &FunctionCallResult{Error: "node_name is required", Success: false}
	}
	op := &engine.ParsedOperation{
		Verb:     "get",
		Resource: "node",
		Name:     nodeName,
	}
	classified := engine.ClassifyOperation(op)
	result, err := getExecutor().Execute(clusterName, classified)
	if err != nil {
		return &FunctionCallResult{Error: err.Error(), Success: false}
	}
	return &FunctionCallResult{Result: result.Output, Success: result.Success}
}

func listNamespacesHandler(args map[string]interface{}, clusterName string) *FunctionCallResult {
	op := &engine.ParsedOperation{
		Verb:     "list",
		Resource: "namespaces",
	}
	classified := engine.ClassifyOperation(op)
	result, err := getExecutor().Execute(clusterName, classified)
	if err != nil {
		return &FunctionCallResult{Error: err.Error(), Success: false}
	}
	return &FunctionCallResult{Result: result.Output, Success: result.Success}
}

func createDeploymentHandler(args map[string]interface{}, clusterName string) *FunctionCallResult {
	name := getStringArg(args, "name", "")
	image := getStringArg(args, "image", "")
	namespace := getStringArg(args, "namespace", "default")
	replicas := getIntArg(args, "replicas", 1)

	if name == "" || image == "" {
		return &FunctionCallResult{Error: "name and image are required", Success: false}
	}

	// 直接构建 ClassifiedOperation，不执行
	classified := &engine.ClassifiedOperation{
		Type:      engine.OperationTypeMutation,
		Verb:      "create",
		Resource:  "deployment",
		Name:      name,
		Namespace: namespace,
		Flags: map[string]string{
			"image":    image,
			"replicas": fmt.Sprintf("%d", replicas),
		},
	}

	return &FunctionCallResult{
		Result:    "confirmation_required:" + name,
		Success:   true,
		Operation: classified,
	}
}

func deletePodHandler(args map[string]interface{}, clusterName string) *FunctionCallResult {
	name := getStringArg(args, "name", "")
	namespace := getStringArg(args, "namespace", "default")
	if name == "" {
		return &FunctionCallResult{Error: "name is required", Success: false}
	}

	// 直接构建 ClassifiedOperation，不执行
	classified := &engine.ClassifiedOperation{
		Type:      engine.OperationTypeMutation,
		Verb:      "delete",
		Resource:  "pod",
		Name:      name,
		Namespace: namespace,
	}

	return &FunctionCallResult{
		Result:    "confirmation_required:" + name,
		Success:   true,
		Operation: classified,
	}
}

func scaleDeploymentHandler(args map[string]interface{}, clusterName string) *FunctionCallResult {
	name := getStringArg(args, "name", "")
	replicas := getIntArg(args, "replicas", 0)
	namespace := getStringArg(args, "namespace", "default")
	if name == "" {
		return &FunctionCallResult{Error: "name is required", Success: false}
	}
	if replicas <= 0 {
		return &FunctionCallResult{Error: "replicas must be positive", Success: false}
	}

	// 直接构建 ClassifiedOperation，不执行
	classified := &engine.ClassifiedOperation{
		Type:      engine.OperationTypeMutation,
		Verb:      "scale",
		Resource:  "deployment",
		Name:      name,
		Namespace: namespace,
		Flags: map[string]string{
			"replicas": fmt.Sprintf("%d", replicas),
		},
	}

	return &FunctionCallResult{
		Result:    "confirmation_required:" + name,
		Success:   true,
		Operation: classified,
	}
}

func updateConfigmapHandler(args map[string]interface{}, clusterName string) *FunctionCallResult {
	name := getStringArg(args, "name", "")
	namespace := getStringArg(args, "namespace", "default")
	data, ok := args["data"].(map[string]interface{})
	if !ok || name == "" {
		return &FunctionCallResult{Error: "name and data are required", Success: false}
	}

	dataJSON, _ := json.Marshal(data)

	// 直接构建 ClassifiedOperation，不执行
	classified := &engine.ClassifiedOperation{
		Type:      engine.OperationTypeMutation,
		Verb:      "update",
		Resource:  "configmap",
		Name:      name,
		Namespace: namespace,
		Flags: map[string]string{
			"data": string(dataJSON),
		},
	}

	return &FunctionCallResult{
		Result:    "confirmation_required:" + name,
		Success:   true,
		Operation: classified,
	}
}

func createScheduledTaskHandler(args map[string]interface{}, clusterName string) *FunctionCallResult {
	schedulerMgr := getSchedulerMgr()
	if schedulerMgr == nil {
		return &FunctionCallResult{Error: "scheduler not configured", Success: false}
	}

	name := getStringArg(args, "name", "")
	schedule := getStringArg(args, "schedule", "")
	operation := getStringArg(args, "operation", "")
	namespace := getStringArg(args, "namespace", "default")

	if name == "" || schedule == "" || operation == "" {
		return &FunctionCallResult{Error: "name, schedule, and operation are required", Success: false}
	}

	parsedOp, err := engine.Parse(operation)
	if err != nil || parsedOp == nil {
		return &FunctionCallResult{Error: fmt.Sprintf("invalid operation: %s", operation), Success: false}
	}

	classifiedOp := engine.ClassifyOperation(parsedOp)
	classifiedOp.Namespace = namespace

	task := &scheduler.ScheduledTask{
		ID:            name,
		Name:          name,
		CronExpr:      schedule,
		TargetCluster: clusterName,
		Operation:     classifiedOp,
		Enabled:       true,
	}

	if err := schedulerMgr.AddTask(task); err != nil {
		return &FunctionCallResult{Error: fmt.Sprintf("failed to create task: %v", err), Success: false}
	}

	return &FunctionCallResult{Result: fmt.Sprintf("Task '%s' created successfully with schedule '%s'", name, schedule), Success: true}
}

func listScheduledTasksHandler(args map[string]interface{}, clusterName string) *FunctionCallResult {
	schedulerMgr := getSchedulerMgr()
	if schedulerMgr == nil {
		return &FunctionCallResult{Error: "scheduler not configured", Success: false}
	}

	tasks := schedulerMgr.ListTasks()
	if len(tasks) == 0 {
		return &FunctionCallResult{Result: "No scheduled tasks.", Success: true}
	}

	var result strings.Builder
	result.WriteString("Scheduled tasks:\n")
	for _, task := range tasks {
		status := "disabled"
		if task.Enabled {
			status = "enabled"
		}
		result.WriteString(fmt.Sprintf("  - %s (%s)\n", task.Name, status))
		result.WriteString(fmt.Sprintf("    ID: %s\n", task.ID))
		result.WriteString(fmt.Sprintf("    Schedule: %s\n", task.CronExpr))
		result.WriteString(fmt.Sprintf("    Cluster: %s\n", task.TargetCluster))
		if !task.NextRunAt.IsZero() {
			result.WriteString(fmt.Sprintf("    Next run: %s\n", task.NextRunAt.Format("2006-01-02 15:04:05")))
		}
	}

	return &FunctionCallResult{Result: result.String(), Success: true}
}

func deleteScheduledTaskHandler(args map[string]interface{}, clusterName string) *FunctionCallResult {
	schedulerMgr := getSchedulerMgr()
	if schedulerMgr == nil {
		return &FunctionCallResult{Error: "scheduler not configured", Success: false}
	}

	name := getStringArg(args, "name", "")
	if name == "" {
		return &FunctionCallResult{Error: "name is required", Success: false}
	}

	if err := schedulerMgr.RemoveTask(name); err != nil {
		return &FunctionCallResult{Error: fmt.Sprintf("failed to delete task: %v", err), Success: false}
	}

	return &FunctionCallResult{Result: fmt.Sprintf("Task '%s' deleted successfully", name), Success: true}
}

func useClusterHandler(args map[string]interface{}, clusterName string) *FunctionCallResult {
	targetCluster := getStringArg(args, "cluster", "")
	if targetCluster == "" {
		return &FunctionCallResult{Error: "cluster name is required", Success: false}
	}

	exec := getExecutor()
	if exec != nil {
		clusters := exec.ListClusters()
		found := false
		for _, c := range clusters {
			if c == targetCluster {
				found = true
				break
			}
		}
		if !found {
			return &FunctionCallResult{
				Error:   fmt.Sprintf("cluster '%s' not found. Available clusters: %v", targetCluster, clusters),
				Success: false,
			}
		}
	}

	return &FunctionCallResult{
		Result:        fmt.Sprintf("Switched to cluster '%s'", targetCluster),
		Success:       true,
		ClusterSwitch: targetCluster,
	}
}

// Helper to get executor from package-level variable
func getExecutor() *engine.Executor {
	return executor
}

// Helper to get scheduler manager from package-level variable
func getSchedulerMgr() *scheduler.Manager {
	return schedulerMgr
}

// Package-level references set by executor.go
var (
	executor     *engine.Executor
	schedulerMgr *scheduler.Manager
)

// SetExecutor sets the package-level executor reference
func SetExecutor(e *engine.Executor) {
	executor = e
}

// SetSchedulerManager sets the package-level scheduler manager reference
func SetSchedulerManager(m *scheduler.Manager) {
	schedulerMgr = m
}

func getStringArg(args map[string]interface{}, key, defaultVal string) string {
	if val, ok := args[key].(string); ok {
		return val
	}
	return defaultVal
}

func getIntArg(args map[string]interface{}, key string, defaultVal int) int {
	if val, ok := args[key].(float64); ok {
		return int(val)
	}
	return defaultVal
}