package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s-agent/pkg/cluster"
	"k8s-agent/pkg/k8s"
	"k8s-agent/pkg/llm"
	"k8s-agent/pkg/log"
	"k8s-agent/pkg/session"
)

// RootCommand is the root command structure
type RootCommand struct {
	manager           *session.Manager
	executor          *engine.Executor
	llmExecutor       *llm.Executor
	clusterReg        *cluster.Registry
	llmService        *llm.Service
	currentClusterCtx string
	appConfig         *cluster.AppConfig
	configPath        string
}

// NewRootCommand creates the root command
func NewRootCommand() *cobra.Command {
	// Load unified configuration from config file + env vars
	// Priority: --config flag > CONFIG_PATH env > default path
	configPath := viper.GetString("config")
	if configPath == "" {
		configPath = os.Getenv("K8S_AGENT_CONFIG")
	}
	if configPath == "" {
		configPath = filepath.Join(os.Getenv("HOME"), ".config", "k8s-agent", "config.yaml")
	}

	appCfg, err := cluster.LoadAppConfig(configPath)
	if err != nil {
		// Use defaults if config loading fails
		appCfg = cluster.DefaultAppConfig()
	}

	// Initialize logger
	log.Init(&log.Config{
		Level:  appCfg.Log.Level,
		Format: appCfg.Log.Format,
	})

	// Initialize store for persistence
	store, err := cluster.NewStore()
	if err != nil {
		// Fall back to in-memory only if store creation fails
		store = nil
	}

	// Initialize registry with store for persistence
	var clusterReg *cluster.Registry
	if store != nil {
		clusterReg = cluster.NewRegistry(cluster.WithStore(store))
	} else {
		clusterReg = cluster.NewRegistry()
	}

	exec := engine.NewExecutor(clusterReg)

	// Create LLM executor for confirmed operations
	llmExec := llm.NewExecutor(exec)

	// Initialize session manager with file-based storage if configured
	var sessionMgr *session.Manager
	if appCfg.Session.StoragePath != "" {
		maxCache := appCfg.Session.MaxCacheSize
		if maxCache <= 0 {
			maxCache = 100
		}
		maxSessions := appCfg.Session.MaxSessions
		if maxSessions <= 0 {
			maxSessions = 10
		}
		sessionMgr, err = session.NewManagerWithFileStoreAndLimits(appCfg.Session.StoragePath, maxCache, maxSessions)
		if err != nil {
			log.Warn("failed to create file-based session store, using in-memory", "error", err)
			sessionMgr = session.NewManager()
		}
	} else {
		sessionMgr = session.NewManager()
	}

	// Initialize LLM service with config
	var llmSvc *llm.Service
	llmCfg := &llm.LLMConfig{
		APIKey:      appCfg.LLM.APIKey,
		Model:       appCfg.LLM.Model,
		BaseURL:     appCfg.LLM.BaseURL,
		MaxTokens:   appCfg.LLM.MaxTokens,
	}
	llmSvc = llm.NewService(llmCfg)

	// Get current cluster context
	currentCluster := ""
	if clusterReg != nil {
		current, err := clusterReg.GetCurrentCluster()
		if err == nil {
			currentCluster = current
		}
	}

	if currentCluster == "" && appCfg.CurrentCluster != "" {
		currentCluster = appCfg.CurrentCluster
	}

	rc := &RootCommand{
		manager:           sessionMgr,
		executor:          exec,
		llmExecutor:       llmExec,
		clusterReg:        clusterReg,
		llmService:        llmSvc,
		currentClusterCtx: currentCluster,
		appConfig:         appCfg,
		configPath:        configPath,
	}

	// Create the root command
	rootCmd := &cobra.Command{
		Use:   "k8s-agent",
		Short: "k8s-agent is a Kubernetes management agent",
		Long:  `k8s-agent is a CLI tool for managing Kubernetes clusters with natural language interface.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			return fmt.Errorf("unknown command: %s", args[0])
		},
	}

	// Add --config flag
	rootCmd.PersistentFlags().StringP("config", "c", "", "Path to config file (default: ~/.config/k8s-agent/config.yaml)")

	// Store viper binding for --config flag
	viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))

	// Add subcommands
	rootCmd.AddCommand(rc.newChatCommand())
	rootCmd.AddCommand(rc.newClusterCommand())

	// Configure viper
	viper.AutomaticEnv()

	return rootCmd
}

// GetAppConfig returns the app configuration
func (rc *RootCommand) GetAppConfig() *cluster.AppConfig {
	return rc.appConfig
}
func (rc *RootCommand) getCurrentCluster() string {
	if rc.clusterReg == nil {
		return ""
	}
	current, err := rc.clusterReg.GetCurrentCluster()
	if err != nil {
		return ""
	}
	return current
}