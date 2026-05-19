package agent

import "github.com/threestoneliu/k8s-agent/pkg/llm"

// BuildSystemPrompt creates the system prompt for function calling mode
// Functions are injected via OpenAI API parameters, not text
// fnExec may be nil (e.g., during session restoration)
func BuildSystemPrompt(clusterName string, fnExec *llm.Executor) string {
	if clusterName == "" {
		clusterName = "default"
	}

	prompt := `你是 Kubernetes 集群管理助手。

角色：Kubernetes 运维专家
职责：
- 查询集群资源状态（pods, services, deployments, nodes 等）

约束：
- 当前集群上下文: ` + clusterName + `
- 使用中文回复`

	// Append skill prompt if executor has skills
	if fnExec != nil {
		skillXML := fnExec.GetAvailableSkillsXML()
		skillInstruction := fnExec.GetProgressiveDisclosurePrompt()
		if skillXML != "" {
			prompt += "\n\n" + skillInstruction + "\n\n" + skillXML
		}
	}

	return prompt
}
