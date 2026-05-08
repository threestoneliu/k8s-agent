package cli

import (
	"fmt"

	"github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"k8s-agent/pkg/agent"
	"k8s-agent/pkg/llm"
	"k8s-agent/pkg/session"
)

func (rc *RootCommand) newChatCommand() *cobra.Command {
	chatCmd := &cobra.Command{
		Use:   "chat",
		Short: "Start interactive chat mode",
		Long:  `Start an interactive conversation with k8s-agent for Kubernetes operations.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return rc.runChatTUI(cmd, args)
		},
	}

	return chatCmd
}

// runChatTUI starts the Bubble Tea TUI chat mode
func (rc *RootCommand) runChatTUI(cmd *cobra.Command, args []string) error {
	// Get current cluster context
	clusterCtx := rc.getCurrentCluster()
	if clusterCtx == "" {
		clusterCtx = rc.currentClusterCtx
	}
	if clusterCtx == "" {
		clusterCtx = "default"
	}

	// Create function executor
	fnExec := llm.NewExecutorWithScheduler(rc.executor, rc.schedulerMgr)

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
	agentInstance := agent.NewAgent(rc.llmService, fnExec, rc.confirmMgr, store, sessionID, clusterCtx, ctxManager)

	// Create the TUI model with the agent
	m := newTUIModel(clusterCtx, agentInstance, rc.executor, rc.appConfig)

	// Run the Bubble Tea program with mouse support
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if err := p.Start(); err != nil {
		return fmt.Errorf("failed to start TUI: %w", err)
	}

	return nil
}
