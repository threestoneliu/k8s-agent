package llm

import (
	"testing"
)

func TestFunctionRegistry(t *testing.T) {
	// Test that registry is populated
	functions := GetFunctions()
	if len(functions) == 0 {
		t.Error("GetFunctions() returned empty, registry not populated")
		return
	}

	t.Logf("Registry has %d functions", len(functions))
	for _, fn := range functions {
		t.Logf("  - %s", fn.Name)
	}

	// Test that handlers are available
	handlerNames := []string{
		"list_pods", "get_pod", "list_services", "list_deployments",
		"list_nodes", "describe_pod", "describe_service", "describe_deployment",
		"get_node_status", "list_namespaces", "create_deployment", "delete_pod",
		"scale_deployment", "update_configmap", "create_scheduled_task",
		"list_scheduled_tasks", "delete_scheduled_task", "use_cluster",
	}

	for _, name := range handlerNames {
		handler, ok := GetHandler(name)
		if !ok {
			t.Errorf("Handler for %s not found in registry", name)
			continue
		}
		if handler == nil {
			t.Errorf("Handler for %s is nil", name)
		} else {
			t.Logf("Handler for %s is registered and not nil", name)
		}
	}
}

func TestK8sFunctionsMatchesRegistry(t *testing.T) {
	functions := GetFunctions()
	k8sFunctions := K8sFunctions

	if len(functions) != len(k8sFunctions) {
		t.Errorf("GetFunctions() count (%d) != K8sFunctions count (%d)",
			len(functions), len(k8sFunctions))
	}

	// Build map of K8sFunctions for comparison
	k8sMap := make(map[string]bool)
	for _, fn := range k8sFunctions {
		k8sMap[fn.Name] = true
	}

	for _, fn := range functions {
		if !k8sMap[fn.Name] {
			t.Errorf("Function %s in GetFunctions() but not in K8sFunctions", fn.Name)
		}
	}
}