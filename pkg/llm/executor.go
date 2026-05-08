package llm

import (
	"encoding/json"
	"fmt"
	"strings"

	"k8s-agent/pkg/engine"
	"k8s-agent/pkg/scheduler"
)

// Executor executes K8s function calls
type Executor struct {
	engine        *engine.Executor
	schedulerMgr  *scheduler.Manager
}

// NewExecutor creates a new function executor
func NewExecutor(eng *engine.Executor) *Executor {
	SetExecutor(eng)
	return &Executor{engine: eng}
}

// NewExecutorWithScheduler creates a new function executor with scheduler support
func NewExecutorWithScheduler(eng *engine.Executor, schedulerMgr *scheduler.Manager) *Executor {
	SetExecutor(eng)
	SetSchedulerManager(schedulerMgr)
	return &Executor{engine: eng, schedulerMgr: schedulerMgr}
}

// ExecuteConfirmedOperation executes a confirmed mutation operation
func (e *Executor) ExecuteConfirmedOperation(clusterName string, op *engine.ClassifiedOperation) *FunctionCallResult {
	if op == nil {
		return &FunctionCallResult{Error: "operation is nil", Success: false}
	}

	// 直接调用 engine.Execute() 执行，跳过确认检查
	result, err := e.engine.Execute(clusterName, op)
	if err != nil {
		return &FunctionCallResult{Error: err.Error(), Success: false}
	}
	return &FunctionCallResult{Result: result.Output, Success: result.Success}
}

// ExecuteFunctionCall executes a function call and returns the result
func (e *Executor) ExecuteFunctionCall(call *FunctionCall, clusterName string) *FunctionCallResult {
	if call == nil {
		return &FunctionCallResult{
			Error:   "function call is nil",
			Success: false,
		}
	}

	// Parse arguments
	var args map[string]interface{}
	if call.Arguments != "" {
		if err := json.Unmarshal([]byte(call.Arguments), &args); err != nil {
			return &FunctionCallResult{
				Name:    call.Name,
				Error:   fmt.Sprintf("failed to parse arguments: %v", err),
				Success: false,
			}
		}
	}

	// Get cluster name from args if not provided
	if clusterName == "" {
		if cn, ok := args["cluster"].(string); ok && cn != "" {
			clusterName = cn
		}
	}

	// Execute based on function name
	result := e.executeFunction(call.Name, args, clusterName)
	result.Name = call.Name
	return result
}

func (e *Executor) executeFunction(name string, args map[string]interface{}, clusterName string) *FunctionCallResult {
	handler, ok := GetHandler(name)
	if !ok {
		return &FunctionCallResult{
			Error:   fmt.Sprintf("unknown function: %s", name),
			Success: false,
		}
	}
	return handler(args, clusterName)
}

func (e *Executor) listPods(args map[string]interface{}, clusterName string) *FunctionCallResult {
	namespace := getStringArg(args, "namespace", "default")
	op := &engine.ParsedOperation{
		Verb:      "list",
		Resource:  "pods",
		Namespace: namespace,
	}
	classified := engine.ClassifyOperation(op)
	result, err := e.engine.Execute(clusterName, classified)
	if err != nil {
		return &FunctionCallResult{Error: err.Error(), Success: false}
	}
	return &FunctionCallResult{Result: result.Output, Success: result.Success}
}

func (e *Executor) getPod(args map[string]interface{}, clusterName string) *FunctionCallResult {
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
	result, err := e.engine.Execute(clusterName, classified)
	if err != nil {
		return &FunctionCallResult{Error: err.Error(), Success: false}
	}
	return &FunctionCallResult{Result: result.Output, Success: result.Success}
}

func (e *Executor) listServices(args map[string]interface{}, clusterName string) *FunctionCallResult {
	namespace := getStringArg(args, "namespace", "default")
	op := &engine.ParsedOperation{
		Verb:      "list",
		Resource:  "services",
		Namespace: namespace,
	}
	classified := engine.ClassifyOperation(op)
	result, err := e.engine.Execute(clusterName, classified)
	if err != nil {
		return &FunctionCallResult{Error: err.Error(), Success: false}
	}
	return &FunctionCallResult{Result: result.Output, Success: result.Success}
}

func (e *Executor) listDeployments(args map[string]interface{}, clusterName string) *FunctionCallResult {
	namespace := getStringArg(args, "namespace", "default")
	op := &engine.ParsedOperation{
		Verb:      "list",
		Resource:  "deployments",
		Namespace: namespace,
	}
	classified := engine.ClassifyOperation(op)
	result, err := e.engine.Execute(clusterName, classified)
	if err != nil {
		return &FunctionCallResult{Error: err.Error(), Success: false}
	}
	return &FunctionCallResult{Result: result.Output, Success: result.Success}
}

func (e *Executor) listNodes(args map[string]interface{}, clusterName string) *FunctionCallResult {
	op := &engine.ParsedOperation{
		Verb:     "list",
		Resource: "nodes",
	}
	classified := engine.ClassifyOperation(op)
	result, err := e.engine.Execute(clusterName, classified)
	if err != nil {
		return &FunctionCallResult{Error: err.Error(), Success: false}
	}
	return &FunctionCallResult{Result: result.Output, Success: result.Success}
}

func (e *Executor) describePod(args map[string]interface{}, clusterName string) *FunctionCallResult {
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
	result, err := e.engine.Execute(clusterName, classified)
	if err != nil {
		return &FunctionCallResult{Error: err.Error(), Success: false}
	}
	return &FunctionCallResult{Result: result.Output, Success: result.Success}
}

func (e *Executor) describeService(args map[string]interface{}, clusterName string) *FunctionCallResult {
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
	result, err := e.engine.Execute(clusterName, classified)
	if err != nil {
		return &FunctionCallResult{Error: err.Error(), Success: false}
	}
	return &FunctionCallResult{Result: result.Output, Success: result.Success}
}

func (e *Executor) describeDeployment(args map[string]interface{}, clusterName string) *FunctionCallResult {
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
	result, err := e.engine.Execute(clusterName, classified)
	if err != nil {
		return &FunctionCallResult{Error: err.Error(), Success: false}
	}
	return &FunctionCallResult{Result: result.Output, Success: result.Success}
}

func (e *Executor) getNodeStatus(args map[string]interface{}, clusterName string) *FunctionCallResult {
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
	result, err := e.engine.Execute(clusterName, classified)
	if err != nil {
		return &FunctionCallResult{Error: err.Error(), Success: false}
	}
	return &FunctionCallResult{Result: result.Output, Success: result.Success}
}

func (e *Executor) listNamespaces(args map[string]interface{}, clusterName string) *FunctionCallResult {
	op := &engine.ParsedOperation{
		Verb:     "list",
		Resource: "namespaces",
	}
	classified := engine.ClassifyOperation(op)
	result, err := e.engine.Execute(clusterName, classified)
	if err != nil {
		return &FunctionCallResult{Error: err.Error(), Success: false}
	}
	return &FunctionCallResult{Result: result.Output, Success: result.Success}
}

func (e *Executor) createDeployment(args map[string]interface{}, clusterName string) *FunctionCallResult {
	name := getStringArg(args, "name", "")
	image := getStringArg(args, "image", "")
	namespace := getStringArg(args, "namespace", "default")
	replicas := getIntArg(args, "replicas", 1)

	if name == "" || image == "" {
		return &FunctionCallResult{Error: "name and image are required", Success: false}
	}

	op := &engine.ParsedOperation{
		Verb:      "create",
		Resource:  "deployment",
		Name:      name,
		Namespace: namespace,
		Flags: map[string]string{
			"image":    image,
			"replicas": fmt.Sprintf("%d", replicas),
		},
	}
	classified := engine.ClassifyOperation(op)
	result, err := e.engine.Execute(clusterName, classified)
	if err != nil {
		return &FunctionCallResult{Error: err.Error(), Success: false}
	}
	// Check if mutation requires confirmation
	if result.Output == "confirmation_required" {
		return &FunctionCallResult{
			Result:  "confirmation_required:" + name,
			Success: true, // Not an error, needs confirmation
		}
	}
	return &FunctionCallResult{Result: result.Output, Success: result.Success}
}

func (e *Executor) deletePod(args map[string]interface{}, clusterName string) *FunctionCallResult {
	name := getStringArg(args, "name", "")
	namespace := getStringArg(args, "namespace", "default")
	if name == "" {
		return &FunctionCallResult{Error: "name is required", Success: false}
	}
	op := &engine.ParsedOperation{
		Verb:      "delete",
		Resource:  "pod",
		Name:      name,
		Namespace: namespace,
	}
	classified := engine.ClassifyOperation(op)
	result, err := e.engine.Execute(clusterName, classified)
	if err != nil {
		return &FunctionCallResult{Error: err.Error(), Success: false}
	}
	// Check if mutation requires confirmation
	if result.Output == "confirmation_required" {
		return &FunctionCallResult{
			Result:  "confirmation_required:" + name,
			Success: true,
		}
	}
	return &FunctionCallResult{Result: result.Output, Success: result.Success}
}

func (e *Executor) scaleDeployment(args map[string]interface{}, clusterName string) *FunctionCallResult {
	name := getStringArg(args, "name", "")
	replicas := getIntArg(args, "replicas", 0)
	namespace := getStringArg(args, "namespace", "default")
	if name == "" {
		return &FunctionCallResult{Error: "name is required", Success: false}
	}
	if replicas <= 0 {
		return &FunctionCallResult{Error: "replicas must be positive", Success: false}
	}
	op := &engine.ParsedOperation{
		Verb:      "scale",
		Resource:  "deployment",
		Name:      name,
		Namespace: namespace,
		Flags: map[string]string{
			"replicas": fmt.Sprintf("%d", replicas),
		},
	}
	classified := engine.ClassifyOperation(op)
	result, err := e.engine.Execute(clusterName, classified)
	if err != nil {
		return &FunctionCallResult{Error: err.Error(), Success: false}
	}
	// Check if mutation requires confirmation
	if result.Output == "confirmation_required" {
		return &FunctionCallResult{
			Result:  "confirmation_required:" + name,
			Success: true,
		}
	}
	return &FunctionCallResult{Result: result.Output, Success: result.Success}
}

func (e *Executor) updateConfigmap(args map[string]interface{}, clusterName string) *FunctionCallResult {
	name := getStringArg(args, "name", "")
	namespace := getStringArg(args, "namespace", "default")
	data, ok := args["data"].(map[string]interface{})
	if !ok || name == "" {
		return &FunctionCallResult{Error: "name and data are required", Success: false}
	}

	// Convert data map to JSON string for flags
	dataJSON, _ := json.Marshal(data)
	op := &engine.ParsedOperation{
		Verb:      "update",
		Resource:  "configmap",
		Name:      name,
		Namespace: namespace,
		Flags: map[string]string{
			"data": string(dataJSON),
		},
	}
	classified := engine.ClassifyOperation(op)
	result, err := e.engine.Execute(clusterName, classified)
	if err != nil {
		return &FunctionCallResult{Error: err.Error(), Success: false}
	}
	// Check if mutation requires confirmation
	if result.Output == "confirmation_required" {
		return &FunctionCallResult{
			Result:  "confirmation_required:" + name,
			Success: true,
		}
	}
	return &FunctionCallResult{Result: result.Output, Success: result.Success}
}

func (e *Executor) createScheduledTask(args map[string]interface{}, clusterName string) *FunctionCallResult {
	if e.schedulerMgr == nil {
		return &FunctionCallResult{Error: "scheduler not configured", Success: false}
	}

	name := getStringArg(args, "name", "")
	schedule := getStringArg(args, "schedule", "")
	operation := getStringArg(args, "operation", "")
	namespace := getStringArg(args, "namespace", "default")

	if name == "" || schedule == "" || operation == "" {
		return &FunctionCallResult{Error: "name, schedule, and operation are required", Success: false}
	}

	// Parse the operation into a classified operation
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

	if err := e.schedulerMgr.AddTask(task); err != nil {
		return &FunctionCallResult{Error: fmt.Sprintf("failed to create task: %v", err), Success: false}
	}

	return &FunctionCallResult{Result: fmt.Sprintf("Task '%s' created successfully with schedule '%s'", name, schedule), Success: true}
}

func (e *Executor) listScheduledTasks(args map[string]interface{}, clusterName string) *FunctionCallResult {
	if e.schedulerMgr == nil {
		return &FunctionCallResult{Error: "scheduler not configured", Success: false}
	}

	tasks := e.schedulerMgr.ListTasks()
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

func (e *Executor) deleteScheduledTask(args map[string]interface{}, clusterName string) *FunctionCallResult {
	if e.schedulerMgr == nil {
		return &FunctionCallResult{Error: "scheduler not configured", Success: false}
	}

	name := getStringArg(args, "name", "")
	if name == "" {
		return &FunctionCallResult{Error: "name is required", Success: false}
	}

	if err := e.schedulerMgr.RemoveTask(name); err != nil {
		return &FunctionCallResult{Error: fmt.Sprintf("failed to delete task: %v", err), Success: false}
	}

	return &FunctionCallResult{Result: fmt.Sprintf("Task '%s' deleted successfully", name), Success: true}
}

func (e *Executor) useCluster(args map[string]interface{}, clusterName string) *FunctionCallResult {
	targetCluster := getStringArg(args, "cluster", "")
	if targetCluster == "" {
		return &FunctionCallResult{Error: "cluster name is required", Success: false}
	}

	// Validate cluster exists by checking if executor can access it
	// The executor will fail if cluster doesn't exist, but we provide early feedback
	if e.engine != nil {
		// Try to get cluster info to validate it exists
		clusters := e.engine.ListClusters()
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

	// Return cluster switch signal - actual switching handled by caller
	return &FunctionCallResult{
		Result:        fmt.Sprintf("Switched to cluster '%s'", targetCluster),
		Success:       true,
		ClusterSwitch: targetCluster,
	}
}