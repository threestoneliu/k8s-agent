package engine

import (
	"testing"
)

func TestParseBasicCommands(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantVerb    string
		wantResource string
	}{
		{"get pods", "get pods", "get", "pods"},
		{"list deployments", "list deployments", "list", "deployments"},
		{"delete pod", "delete pod", "delete", "pod"},
		{"create deployment", "create deployment", "create", "deployment"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse(%q) returned error: %v", tt.input, err)
			}
			if got.Verb != tt.wantVerb {
				t.Errorf("Parse(%q).Verb = %q, want %q", tt.input, got.Verb, tt.wantVerb)
			}
			if got.Resource != tt.wantResource {
				t.Errorf("Parse(%q).Resource = %q, want %q", tt.input, got.Resource, tt.wantResource)
			}
		})
	}
}

func TestParseWithNamespace(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantNS      string
	}{
		{"list pods -n default", "list pods -n default", "default"},
		{"get pods --namespace kube-system", "get pods --namespace kube-system", "kube-system"},
		{"delete pod nginx -n default", "delete pod nginx -n default", "default"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse(%q) returned error: %v", tt.input, err)
			}
			if got.Namespace != tt.wantNS {
				t.Errorf("Parse(%q).Namespace = %q, want %q", tt.input, got.Namespace, tt.wantNS)
			}
		})
	}
}

func TestParseWithResourceName(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantName  string
	}{
		{"describe pod nginx-xxx", "describe pod nginx-xxx", "nginx-xxx"},
		{"delete pod my-pod", "delete pod my-pod", "my-pod"},
		{"get deployment nginx", "get deployment nginx", "nginx"},
		{"describe deployment my-app", "describe deployment my-app", "my-app"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse(%q) returned error: %v", tt.input, err)
			}
			if got.Name != tt.wantName {
				t.Errorf("Parse(%q).Name = %q, want %q", tt.input, got.Name, tt.wantName)
			}
		})
	}
}

func TestParseWithFlags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		flagKey  string
		flagVal  string
	}{
		{"create deployment with image", "create deployment nginx --image=nginx", "image", "nginx"},
		{"scale with replicas", "scale deployment app --replicas=3", "replicas", "3"},
		{"get pods with label selector", "get pods -l app=nginx", "l", "app=nginx"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse(%q) returned error: %v", tt.input, err)
			}
			if got.Flags[tt.flagKey] != tt.flagVal {
				t.Errorf("Parse(%q).Flags[%q] = %q, want %q", tt.input, tt.flagKey, got.Flags[tt.flagKey], tt.flagVal)
			}
		})
	}
}

func TestParseRawInputPreserved(t *testing.T) {
	inputs := []string{
		"get pods",
		"list deployments -n default",
		"delete pod nginx --namespace kube-system",
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			got, err := Parse(input)
			if err != nil {
				t.Fatalf("Parse(%q) returned error: %v", input, err)
			}
			if got.RawInput != input {
				t.Errorf("Parse(%q).RawInput = %q, want %q", input, got.RawInput, input)
			}
		})
	}
}

func TestParseEmptyInput(t *testing.T) {
	_, err := Parse("")
	if err == nil {
		t.Error("Parse(\"\") expected error, got nil")
	}
}

func TestParseOnlyVerb(t *testing.T) {
	_, err := Parse("get")
	if err == nil {
		t.Error("Parse(\"get\") expected error, got nil")
	}
}

func TestParseUnknownVerb(t *testing.T) {
	result, err := Parse("unknown pods")
	if err != nil {
		t.Errorf("Parse(\"unknown pods\") returned unexpected error: %v", err)
	}
	if result != nil && result.Verb != "unknown" {
		t.Errorf("Parse(\"unknown pods\").Verb = %q, want %q", result.Verb, "unknown")
	}
}

func TestParseWithMultipleFlags(t *testing.T) {
	input := "create deployment nginx --image=nginx --replicas=3 -n default"
	got, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse(%q) returned error: %v", input, err)
	}

	if got.Flags["image"] != "nginx" {
		t.Errorf("Flags[image] = %q, want %q", got.Flags["image"], "nginx")
	}
	if got.Flags["replicas"] != "3" {
		t.Errorf("Flags[replicas] = %q, want %q", got.Flags["replicas"], "3")
	}
	if got.Namespace != "default" {
		t.Errorf("Namespace = %q, want %q", got.Namespace, "default")
	}
}

func TestParseFlagsWithoutValue(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		flagKey     string
		checkNamespace bool
		namespace   string
	}{
		{"image flag without value", "create deployment nginx --image", "image", false, ""},
		{"replicas flag without value", "scale deployment app --replicas", "replicas", false, ""},
		{"namespace flag without value", "get pods --namespace", "namespace", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)
			if err != nil {
				t.Fatalf("Parse(%q) returned error: %v", tt.input, err)
			}
			if tt.checkNamespace {
				if got.Namespace != tt.namespace {
					t.Errorf("Parse(%q).Namespace = %q, want %q", tt.input, got.Namespace, tt.namespace)
				}
			} else {
				if _, ok := got.Flags[tt.flagKey]; !ok {
					t.Errorf("Parse(%q).Flags[%q] not set", tt.input, tt.flagKey)
				}
			}
		})
	}
}

func TestParseShortNamespaceFlag(t *testing.T) {
	input := "get pods -n kube-system"
	got, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse(%q) returned error: %v", input, err)
	}
	if got.Namespace != "kube-system" {
		t.Errorf("Namespace = %q, want %q", got.Namespace, "kube-system")
	}
}

func TestParseShortLabelFlagWithoutValue(t *testing.T) {
	input := "get pods -l"
	got, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse(%q) returned error: %v", input, err)
	}
	if _, ok := got.Flags["l"]; !ok {
		t.Error("Flags[l] should be set even without value")
	}
}

func TestParseDoubleDashSeparator(t *testing.T) {
	input := "get pods -- --custom-flag value"
	got, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse(%q) returned error: %v", input, err)
	}
	if got.Name != "--custom-flag" {
		t.Errorf("Name = %q, want %q", got.Name, "--custom-flag")
	}
}

func TestParseResourceNameAfterFlags(t *testing.T) {
	input := "get pods -n default my-pod"
	got, err := Parse(input)
	if err != nil {
		t.Fatalf("Parse(%q) returned error: %v", input, err)
	}
	if got.Name != "my-pod" {
		t.Errorf("Name = %q, want %q", got.Name, "my-pod")
	}
	if got.Namespace != "default" {
		t.Errorf("Namespace = %q, want %q", got.Namespace, "default")
	}
}
