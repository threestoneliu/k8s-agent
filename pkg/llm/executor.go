package llm

import (
	"encoding/json"
	"fmt"

	k8s "github.com/threestoneliu/k8s-agent/pkg/k8s"
	sharedutil "github.com/threestoneliu/k8s-agent/pkg/shared"
)

// Executor executes K8s function calls
type Executor struct {
	engine *k8s.Executor
}

// NewExecutor creates a new function executor
func NewExecutor(eng *k8s.Executor) *Executor {
	SetExecutor(eng)
	return &Executor{engine: eng}
}

// ExecuteConfirmedOperation executes a confirmed mutation operation
func (e *Executor) ExecuteConfirmedOperation(clusterName string, op *k8s.ClassifiedOperation) *sharedutil.FunctionResult {
	if op == nil {
		return &sharedutil.FunctionResult{Error: "operation is nil", Success: false}
	}

	result, err := e.engine.Execute(clusterName, op)
	if err != nil {
		return &sharedutil.FunctionResult{Error: err.Error(), Success: false}
	}
	return &sharedutil.FunctionResult{Result: result.Output, Success: result.Success}
}

// ExecuteFunctionCall executes a function call and returns the result
func (e *Executor) ExecuteFunctionCall(call *sharedutil.FunctionCall, clusterName string) *sharedutil.FunctionResult {
	if call == nil {
		return &sharedutil.FunctionResult{
			Error:   "function call is nil",
			Success: false,
		}
	}

	var args map[string]interface{}
	if call.Arguments != "" {
		if err := json.Unmarshal([]byte(call.Arguments), &args); err != nil {
			return &sharedutil.FunctionResult{
				Name:    call.Name,
				Error:   fmt.Sprintf("failed to parse arguments: %v", err),
				Success: false,
			}
		}
	}

	if clusterName == "" {
		if cn, ok := args["cluster"].(string); ok && cn != "" {
			clusterName = cn
		}
	}

	result := e.executeFunction(call.Name, args, clusterName)
	result.Name = call.Name
	return result
}

func (e *Executor) executeFunction(name string, args map[string]interface{}, clusterName string) *sharedutil.FunctionResult {
	handler, ok := GetHandler(name)
	if !ok {
		return &sharedutil.FunctionResult{
			Error:   fmt.Sprintf("unknown function: %s", name),
			Success: false,
		}
	}
	return handler(args, clusterName)
}

func (e *Executor) useCluster(args map[string]interface{}, clusterName string) *sharedutil.FunctionResult {
	targetCluster := getStringArg(args, "cluster", "")
	if targetCluster == "" {
		return &sharedutil.FunctionResult{Error: "cluster name is required", Success: false}
	}

	if e.engine != nil {
		clusters := e.engine.ListClusters()
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
	}
}

func getStringArg(args map[string]interface{}, key, defaultVal string) string {
	if v, ok := args[key].(string); ok {
		return v
	}
	return defaultVal
}