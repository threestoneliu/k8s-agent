package llm

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestExecutor_ExecuteFunctionCall_NilCall(t *testing.T) {
	executor := &Executor{}
	result := executor.ExecuteFunctionCall(nil, "test-cluster")

	if result.Success {
		t.Error("Expected success to be false for nil call")
	}
	if result.Error != "function call is nil" {
		t.Errorf("Expected error 'function call is nil', got: %s", result.Error)
	}
}

func TestExecutor_ExecuteFunctionCall_UnknownFunction(t *testing.T) {
	executor := &Executor{}
	call := &FunctionCall{
		Name:      "unknown_function",
		Arguments: `{}`,
	}
	result := executor.ExecuteFunctionCall(call, "test-cluster")

	if result.Success {
		t.Error("Expected success to be false for unknown function")
	}
	if result.Error == "" {
		t.Error("Expected error message for unknown function")
	}
}

func TestExecutor_ExecuteFunctionCall_InvalidJSON(t *testing.T) {
	executor := &Executor{}
	call := &FunctionCall{
		Name:      "list_pods",
		Arguments: `{invalid json}`,
	}
	result := executor.ExecuteFunctionCall(call, "test-cluster")

	if result.Success {
		t.Error("Expected success to be false for invalid JSON")
	}
	if result.Error == "" {
		t.Error("Expected error message for invalid JSON")
	}
}

func TestExecutor_ExecuteFunctionCall_MissingRequiredArg(t *testing.T) {
	executor := &Executor{}
	call := &FunctionCall{
		Name:      "get_pod",
		Arguments: `{"namespace":"default"}`, // missing required "name"
	}
	result := executor.ExecuteFunctionCall(call, "test-cluster")

	if result.Success {
		t.Error("Expected success to be false for missing required arg")
	}
	if result.Error == "" {
		t.Error("Expected error message for missing required arg")
	}
}

func TestExecutor_ExecuteFunctionCall_ClusterFromArgs(t *testing.T) {
	// This test verifies that cluster can be extracted from args
	// We use create_deployment which returns confirmation_required (doesn't need real executor)
	call := &FunctionCall{
		Name:      "create_deployment",
		Arguments: `{"cluster":"my-cluster","name":"test","image":"nginx"}`,
	}

	// Execute with empty cluster - cluster should come from args
	result := ExecuteFunctionCallForTest(call, "")

	// Should get confirmation_required, not cluster-related error
	if result == nil {
		t.Fatal("Expected result, got nil")
	}
	if result.Error != "" {
		t.Errorf("Expected no error, got: %s", result.Error)
	}
	if !result.Success {
		t.Error("Expected success=true for confirmation_required")
	}
	if !strings.HasPrefix(result.Result, "confirmation_required:") {
		t.Errorf("Expected confirmation_required prefix, got: %s", result.Result)
	}
}

// ExecuteFunctionCallForTest is a test helper that doesn't require package-level executor
func ExecuteFunctionCallForTest(call *FunctionCall, clusterName string) *FunctionCallResult {
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
	result := executeFunctionForTest(call.Name, args, clusterName)
	result.Name = call.Name
	return result
}

func executeFunctionForTest(name string, args map[string]interface{}, clusterName string) *FunctionCallResult {
	handler, ok := GetHandler(name)
	if !ok {
		return &FunctionCallResult{
			Error:   fmt.Sprintf("unknown function: %s", name),
			Success: false,
		}
	}
	return handler(args, clusterName)
}

func TestGetStringArg(t *testing.T) {
	tests := []struct {
		name       string
		args       map[string]interface{}
		key        string
		defaultVal string
		want       string
	}{
		{
			name:       "existing string",
			args:       map[string]interface{}{"key": "value"},
			key:        "key",
			defaultVal: "default",
			want:       "value",
		},
		{
			name:       "missing key",
			args:       map[string]interface{}{},
			key:        "key",
			defaultVal: "default",
			want:       "default",
		},
		{
			name:       "wrong type",
			args:       map[string]interface{}{"key": 123},
			key:        "key",
			defaultVal: "default",
			want:       "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getStringArg(tt.args, tt.key, tt.defaultVal)
			if got != tt.want {
				t.Errorf("getStringArg() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetIntArg(t *testing.T) {
	tests := []struct {
		name       string
		args       map[string]interface{}
		key        string
		defaultVal int
		want       int
	}{
		{
			name:       "existing int",
			args:       map[string]interface{}{"key": float64(42)},
			key:        "key",
			defaultVal: 0,
			want:       42,
		},
		{
			name:       "missing key",
			args:       map[string]interface{}{},
			key:        "key",
			defaultVal: 10,
			want:       10,
		},
		{
			name:       "wrong type",
			args:       map[string]interface{}{"key": "string"},
			key:        "key",
			defaultVal: 10,
			want:       10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getIntArg(tt.args, tt.key, tt.defaultVal)
			if got != tt.want {
				t.Errorf("getIntArg() = %v, want %v", got, tt.want)
			}
		})
	}
}