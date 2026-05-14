package llm

import (
	sharedutil "github.com/threestoneliu/k8s-agent/pkg/shared"
)

// FunctionHandler is the function handler type
type FunctionHandler func(args map[string]interface{}, clusterName string) *sharedutil.FunctionResult

// FunctionDefinition contains function metadata and handler
type FunctionDefinition struct {
	sharedutil.Function
	Handler FunctionHandler
}

// registry is the function registry
var registry = make(map[string]FunctionDefinition)

// RegisterFunction registers a function
func RegisterFunction(fn FunctionDefinition) {
	registry[fn.Name] = fn
}

// GetFunctions returns all registered function definitions
func getFunctions() []sharedutil.Function {
	result := make([]sharedutil.Function, 0, len(registry))
	for _, fn := range registry {
		result = append(result, fn.Function)
	}
	return result
}

// GetHandler returns the handler for a specific function
func GetHandler(name string) (FunctionHandler, bool) {
	fn, ok := registry[name]
	return fn.Handler, ok
}
