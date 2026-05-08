package llm

import "k8s-agent/pkg/engine"

// Function represents a function that can be called by the LLM
type Function struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// FunctionCall represents a call to a function by the LLM
type FunctionCall struct {
	ID        string
	Name      string
	Arguments string // JSON string
}

// FunctionCallResult represents the result of executing a function call
type FunctionCallResult struct {
	Name          string
	Result        string                    // 执行结果
	Error         string                    // 错误信息
	Success       bool
	ClusterSwitch string                    // 集群切换标记，如 "prod"
	Operation     *engine.ClassifiedOperation // 待确认的操作信息
}

// FunctionHandler 是函数处理函数类型
type FunctionHandler func(args map[string]interface{}, clusterName string) *FunctionCallResult

// FunctionDefinition 包含函数元数据和处理方法
type FunctionDefinition struct {
	Function
	Handler FunctionHandler
}

// registry 是函数注册表
var registry = make(map[string]FunctionDefinition)

// RegisterFunction 注册函数
func RegisterFunction(fn FunctionDefinition) {
	registry[fn.Name] = fn
}

// GetFunctions 返回所有已注册函数的定义列表
func GetFunctions() []Function {
	result := make([]Function, 0, len(registry))
	for _, fn := range registry {
		result = append(result, fn.Function)
	}
	return result
}

// GetHandler 返回指定函数的处理函数
func GetHandler(name string) (FunctionHandler, bool) {
	fn, ok := registry[name]
	return fn.Handler, ok
}

// K8sFunctions returns the list of Kubernetes functions available for LLM calling
// Deprecated: Use GetFunctions() instead
// Note: This is lazily initialized to ensure all init() functions have run
var K8sFunctions []Function

func init() {
	// Give auto_register.go a chance to populate the registry first
	// This ensures K8sFunctions is populated after all registrations
	K8sFunctions = GetFunctions()
}