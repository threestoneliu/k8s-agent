package agent

import (
	"strings"
	"testing"
)

func TestBuildSystemPrompt(t *testing.T) {
	tests := []struct {
		name        string
		clusterName string
		wantContain string
	}{
		{
			name:        "default cluster",
			clusterName: "",
			wantContain: "default",
		},
		{
			name:        "specific cluster",
			clusterName: "prod-cluster",
			wantContain: "prod-cluster",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt := BuildSystemPrompt(tt.clusterName, nil)
			if prompt == "" {
				t.Fatal("BuildSystemPrompt returned empty string")
			}
			if !strings.Contains(prompt, tt.wantContain) {
				t.Errorf("expected prompt to contain '%s', got '%s'", tt.wantContain, prompt)
			}
			if !strings.Contains(prompt, "Kubernetes") {
				t.Errorf("expected prompt to contain 'Kubernetes'")
			}
		})
	}
}