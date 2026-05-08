package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (rc *RootCommand) newClusterCommand() *cobra.Command {
	clusterCmd := &cobra.Command{
		Use:   "cluster",
		Short: "Manage Kubernetes clusters",
		Long:  `Manage Kubernetes cluster configurations.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			return fmt.Errorf("unknown cluster command: %s", args[0])
		},
	}

	clusterCmd.AddCommand(rc.newClusterListCommand())
	clusterCmd.AddCommand(rc.newClusterAddCommand())
	clusterCmd.AddCommand(rc.newClusterUseCommand())
	clusterCmd.AddCommand(rc.newClusterRemoveCommand())

	return clusterCmd
}

func (rc *RootCommand) newClusterListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List configured clusters",
		Long:  `List all configured Kubernetes clusters.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			clusters := rc.clusterReg.ListClusters()
			if len(clusters) == 0 {
				fmt.Println("No clusters configured.")
				return nil
			}
			fmt.Println("Configured clusters:")
			for _, c := range clusters {
				fmt.Printf("  - %s\n", c.Name)
			}
			return nil
		},
	}
}

func (rc *RootCommand) newClusterAddCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "add <name> <kubeconfig>",
		Short: "Add a cluster",
		Long:  `Add a new Kubernetes cluster configuration.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 2 {
				return fmt.Errorf("requires exactly two args")
			}
			name := args[0]
			kubeconfig := args[1]

			err := rc.clusterReg.AddCluster(name, kubeconfig)
			if err != nil {
				return fmt.Errorf("failed to add cluster: %w", err)
			}

			fmt.Printf("Cluster '%s' added successfully.\n", name)
			return nil
		},
	}
}

func (rc *RootCommand) newClusterUseCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "use <name>",
		Short: "Set active cluster",
		Long:  `Set the active cluster context.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("requires exactly one arg")
			}
			name := args[0]

			// Verify cluster exists by trying to get it
			_, err := rc.clusterReg.GetCluster(name)
			if err != nil {
				return fmt.Errorf("cluster not found: %s", name)
			}

			// Save the current cluster to config
			err = rc.clusterReg.SetCurrentCluster(name)
			if err != nil {
				return fmt.Errorf("failed to set current cluster: %w", err)
			}

			fmt.Printf("Active cluster set to '%s'\n", name)
			return nil
		},
	}
}

func (rc *RootCommand) newClusterRemoveCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <name>",
		Short: "Remove a cluster",
		Long:  `Remove a Kubernetes cluster configuration.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("requires exactly one arg")
			}
			name := args[0]

			err := rc.clusterReg.RemoveCluster(name)
			if err != nil {
				return fmt.Errorf("failed to remove cluster: %w", err)
			}

			fmt.Printf("Cluster '%s' removed successfully.\n", name)
			return nil
		},
	}
}
