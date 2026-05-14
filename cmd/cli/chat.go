package cli

import (
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/threestoneliu/k8s-agent/pkg/agent"
	"github.com/threestoneliu/k8s-agent/pkg/cluster"
	"github.com/threestoneliu/k8s-agent/pkg/llm"
	"github.com/threestoneliu/k8s-agent/pkg/session"
	"github.com/threestoneliu/k8s-agent/pkg/ui"
)

func (rc *RootCommand) newChatCommand() *cobra.Command {
	chatCmd := &cobra.Command{
		Use:   "chat",
		Short: "Start interactive chat mode",
		Long:  `Start an interactive conversation with k8s-agent for Kubernetes operations.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return rc.runChat(cmd, args)
		},
	}

	return chatCmd
}

// clusterListerAdapter wraps *cluster.Registry to implement ListClusters() []string
type clusterListerAdapter struct {
	reg *cluster.Registry
}

func (a *clusterListerAdapter) ListClusters() []string {
	clusters := a.reg.ListClusters()
	names := make([]string, 0, len(clusters))
	for _, c := range clusters {
		names = append(names, c.Name)
	}
	return names
}

// runChat starts the chat with TUI
func (rc *RootCommand) runChat(cmd *cobra.Command, args []string) error {
	// Get current cluster context
	clusterCtx := rc.getCurrentCluster()
	if clusterCtx == "" {
		clusterCtx = rc.currentClusterCtx
	}
	if clusterCtx == "" {
		clusterCtx = "default"
	}

	// Create function executor
	fnExec := llm.NewExecutor(rc.executor)

	// Generate a new session ID for this chat
	sessionID := uuid.New().String()

	// Get session store from manager
	var store session.StoreInterface
	if rc.manager != nil {
		store = rc.manager.GetStore()
	}

	// Create context manager for token-aware message handling
	ctxManager := session.NewContextManager(rc.appConfig.Context)

	// Create the agent
	agentInstance := agent.NewAgent(rc.llmService, fnExec, store, sessionID, clusterCtx, ctxManager)

	// Create channels for agent-UI communication
	inputChan := make(chan ui.Input, 100)
	outputChan := make(chan ui.Output, 100)

	// Create cluster lister adapter
	clusterLister := &clusterListerAdapter{reg: rc.clusterReg}

	// Set cluster lister on agent so it can handle /clusters command
	agentInstance.SetClusterLister(clusterLister)

	// Read and set config content for /config command
	if rc.configPath != "" {
		if configContent, err := os.ReadFile(rc.configPath); err == nil {
			agentInstance.SetConfigContent(string(configContent))
		}
	}

	// Create TUI instance from pkg/ui
	tuiInstance := ui.NewTUI(clusterCtx, clusterLister, &ui.AppConfig{CurrentCluster: clusterCtx})

	// Start agent in background
	go agentInstance.Run(inputChan, outputChan)

	// Run TUI (this blocks until TUI exits)
	if err := tuiInstance.Run(inputChan, outputChan); err != nil {
		return fmt.Errorf("failed to run TUI: %w", err)
	}

	return nil
}