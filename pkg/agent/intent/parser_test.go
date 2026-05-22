package intent

import (
	"testing"

	"github.com/threestoneliu/k8s-agent/pkg/core"
)

func TestParseToIntent_Basic(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantAct  core.Action
		wantKind string
	}{
		{
			name:    "create deployment",
			input:   "create deployment myapp",
			wantAct: core.ActionCreate,
			wantKind: "Deployment",
		},
		{
			name:    "delete pod",
			input:   "delete pod nginx-pod",
			wantAct: core.ActionDelete,
			wantKind: "Pod",
		},
		{
			name:    "get pods in namespace",
			input:   "get pods in namespace default",
			wantAct: core.ActionInspect,
			wantKind: "Pod",
		},
		{
			name:    "update service",
			input:   "update service my-service",
			wantAct: core.ActionUpdate,
			wantKind: "Service",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseToIntent(tt.input)
			if err != nil {
				t.Fatalf("ParseToIntent() error = %v", err)
			}
			if got == nil {
				t.Fatalf("ParseToIntent() returned nil")
			}
			if got.Action != tt.wantAct {
				t.Errorf("Action = %v, want %v", got.Action, tt.wantAct)
			}
			if got.Target.Kind != tt.wantKind {
				t.Errorf("Kind = %v, want %v", got.Target.Kind, tt.wantKind)
			}
		})
	}
}

func TestParseToIntent_EmptyInput(t *testing.T) {
	got, err := ParseToIntent("")
	if err != nil {
		t.Fatalf("ParseToIntent() error = %v", err)
	}
	if got != nil {
		t.Errorf("ParseToIntent() = %v, want nil for empty input", got)
	}
}

func TestParseToIntent_ActionExtraction(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  core.Action
	}{
		{"create keyword", "create deployment", core.ActionCreate},
		{"add keyword", "add new pod", core.ActionCreate},
		{"delete keyword", "delete the pod", core.ActionDelete},
		{"remove keyword", "remove deployment", core.ActionDelete},
		{"update keyword", "update configmap", core.ActionUpdate},
		{"modify keyword", "modify the service", core.ActionUpdate},
		{"edit keyword", "edit ingress", core.ActionUpdate},
		{"get keyword", "get pods", core.ActionInspect},
		{"list keyword", "list services", core.ActionInspect},
		{"describe keyword", "describe deployment", core.ActionInspect},
		{"show keyword", "show me the nodes", core.ActionInspect},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := ParseToIntent(tt.input)
			if got.Action != tt.want {
				t.Errorf("Action = %v, want %v", got.Action, tt.want)
			}
		})
	}
}

func TestParseToIntent_ChineseKeywords(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  core.Action
	}{
		{"create in Chinese", "创建 deployment", core.ActionCreate},
		{"delete in Chinese", "删除 pod", core.ActionDelete},
		{"update in Chinese", "修改 service", core.ActionUpdate},
		{"inspect in Chinese", "查看 pods", core.ActionInspect},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := ParseToIntent(tt.input)
			if got.Action != tt.want {
				t.Errorf("Action = %v, want %v", got.Action, tt.want)
			}
		})
	}
}

func TestParseToIntent_KindExtraction(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"deployment", "create deployment", "Deployment"},
		{"service alias", "list svc", "Service"},
		{"pod", "get pods", "Pod"},
		{"configmap alias", "get cm", "ConfigMap"},
		{"ingress", "describe ingress", "Ingress"},
		{"namespace", "show ns", "Namespace"},
		{"node", "get nodes", "Node"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := ParseToIntent(tt.input)
			if got.Target.Kind != tt.want {
				t.Errorf("Kind = %v, want %v", got.Target.Kind, tt.want)
			}
		})
	}
}

func TestParseToIntent_NameExtraction(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"named pattern", "get pod named my-pod", "my-pod"},
		{"name pattern", "delete deployment name app", "app"},
		{"chinese named", "创建 configmap 名为 settings", "settings"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := ParseToIntent(tt.input)
			if got.Target.Name != tt.want {
				t.Errorf("Name = %v, want %v", got.Target.Name, tt.want)
			}
		})
	}
}

func TestParseToIntent_NamespaceExtraction(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"in namespace", "get pods in namespace default", "default"},
		{"namespace keyword", "list service namespace kube-system", "kube-system"},
		{"chinese in namespace", "查看 pod 在命名空间 test", "test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := ParseToIntent(tt.input)
			if got.Target.Namespace != tt.want {
				t.Errorf("Namespace = %v, want %v", got.Target.Namespace, tt.want)
			}
		})
	}
}

func TestParseToIntent_RiskLevel(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantRisk core.RiskLevel
	}{
		{"create defaults to low", "create deployment", core.RiskLow},
		{"inspect defaults to low", "get pods", core.RiskLow},
		{"update defaults to medium", "update service", core.RiskMedium},
		{"delete defaults to high", "delete pod", core.RiskHigh},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := ParseToIntent(tt.input)
			if got.RiskLevel != tt.wantRisk {
				t.Errorf("RiskLevel = %v, want %v", got.RiskLevel, tt.wantRisk)
			}
		})
	}
}