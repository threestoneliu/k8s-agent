package llm

import (
	"encoding/json"
	"testing"
)

func TestFunctionCall_JSON(t *testing.T) {
	call := FunctionCall{
		Name:      "list_pods",
		Arguments: `{"namespace":"default"}`,
	}

	// Test JSON marshaling
	data, err := json.Marshal(call)
	if err != nil {
		t.Fatalf("failed to marshal FunctionCall: %v", err)
	}

	// Test JSON unmarshaling
	var decoded FunctionCall
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal FunctionCall: %v", err)
	}

	if decoded.Name != call.Name {
		t.Errorf("Name mismatch: got %v, want %v", decoded.Name, call.Name)
	}
	if decoded.Arguments != call.Arguments {
		t.Errorf("Arguments mismatch: got %v, want %v", decoded.Arguments, call.Arguments)
	}
}

func TestFunctionCallResult_Success(t *testing.T) {
	result := FunctionCallResult{
		Name:    "list_pods",
		Result:  `{"pods":["nginx-abc","redis-xyz"]}`,
		Success: true,
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal FunctionCallResult: %v", err)
	}

	var decoded FunctionCallResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal FunctionCallResult: %v", err)
	}

	if decoded.Name != result.Name {
		t.Errorf("Name mismatch: got %v, want %v", decoded.Name, result.Name)
	}
	if decoded.Result != result.Result {
		t.Errorf("Result mismatch: got %v, want %v", decoded.Result, result.Result)
	}
	if decoded.Success != result.Success {
		t.Errorf("Success mismatch: got %v, want %v", decoded.Success, result.Success)
	}
	if decoded.Error != "" {
		t.Errorf("Error should be empty for success case, got: %v", decoded.Error)
	}
}

func TestFunctionCallResult_Error(t *testing.T) {
	result := FunctionCallResult{
		Name:    "list_pods",
		Result:  "",
		Error:   "namespace not found",
		Success: false,
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("failed to marshal FunctionCallResult: %v", err)
	}

	var decoded FunctionCallResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal FunctionCallResult: %v", err)
	}

	if decoded.Success != false {
		t.Errorf("Success should be false, got: %v", decoded.Success)
	}
	if decoded.Error != result.Error {
		t.Errorf("Error mismatch: got %v, want %v", decoded.Error, result.Error)
	}
}

func TestFunction_Schema(t *testing.T) {
	fn := Function{
		Name:        "list_pods",
		Description: "List pods in a namespace",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"namespace": map[string]interface{}{
					"type":    "string",
					"default": "default",
				},
			},
		},
	}

	// Test JSON schema representation
	data, err := json.Marshal(fn)
	if err != nil {
		t.Fatalf("failed to marshal Function: %v", err)
	}

	var decoded Function
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to unmarshal Function: %v", err)
	}

	if decoded.Name != fn.Name {
		t.Errorf("Name mismatch: got %v, want %v", decoded.Name, fn.Name)
	}
	if decoded.Description != fn.Description {
		t.Errorf("Description mismatch: got %v, want %v", decoded.Description, fn.Description)
	}
}

func TestK8sFunctions_NotEmpty(t *testing.T) {
	if len(K8sFunctions) == 0 {
		t.Error("K8sFunctions should not be empty")
	}
}

func TestK8sFunctions_ValidSchema(t *testing.T) {
	for _, fn := range K8sFunctions {
		if fn.Name == "" {
			t.Error("Function name should not be empty")
		}
		if fn.Description == "" {
			t.Error("Function description should not be empty")
		}
		if fn.Parameters == nil {
			t.Error("Function parameters should not be nil")
		}

		// Verify parameters is valid JSON
		data, err := json.Marshal(fn.Parameters)
		if err != nil {
			t.Errorf("Function %s has invalid parameters: %v", fn.Name, err)
		}

		// Verify it can be unmarshaled back
		var params map[string]interface{}
		if err := json.Unmarshal(data, &params); err != nil {
			t.Errorf("Function %s parameters cannot be unmarshaled: %v", fn.Name, err)
		}
	}
}

func TestK8sFunctions_HasRequiredFunctions(t *testing.T) {
	expectedFunctions := map[string]bool{
		"list_pods":        true,
		"get_pod":          true,
		"list_services":    true,
		"list_deployments": true,
		"list_nodes":       true,
		"describe_pod":     true,
		"create_deployment": true,
		"delete_pod":       true,
		"scale_deployment": true,
	}

	found := make(map[string]bool)
	for _, fn := range K8sFunctions {
		if expectedFunctions[fn.Name] {
			found[fn.Name] = true
		}
	}

	for name := range expectedFunctions {
		if !found[name] {
			t.Errorf("Missing expected function: %s", name)
		}
	}
}