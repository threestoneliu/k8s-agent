package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s-agent/pkg/engine"
)

func (rc *RootCommand) newScaleCommand() *cobra.Command {
	scaleCmd := &cobra.Command{
		Use:   "scale <resource> <name> --replicas=<count>",
		Short: "Scale a Kubernetes resource",
		Long:  `Scale a Kubernetes deployment or other scalable resource. This operation requires confirmation.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return rc.executeScale(cmd, args)
		},
	}

	scaleCmd.Flags().StringP("namespace", "n", "default", "Namespace of the resource")
	scaleCmd.Flags().Int("replicas", 1, "Number of replicas")
	scaleCmd.Flags().Bool("force", false, "Skip confirmation")

	return scaleCmd
}

func (rc *RootCommand) executeScale(cmd *cobra.Command, args []string) error {
	resource := args[0]
	name := args[1]

	clusterName := rc.getCurrentCluster()
	if clusterName == "" {
		clusterName = "default"
	}

	namespace, _ := cmd.Flags().GetString("namespace")
	replicas, _ := cmd.Flags().GetInt("replicas")
	force, _ := cmd.Flags().GetBool("force")

	op := &engine.ClassifiedOperation{
		Type:      engine.OperationTypeMutation,
		Verb:      "scale",
		Resource:  resource,
		Name:      name,
		Namespace: namespace,
		Flags: map[string]string{
			"replicas": fmt.Sprintf("%d", replicas),
		},
	}

	// If not force mode, require confirmation
	if !force {
		confirmKey, err := rc.confirmMgr.CreateConfirmation(clusterName, op)
		if err != nil {
			return fmt.Errorf("failed to create confirmation: %w", err)
		}
		return fmt.Errorf("operation requires confirmation. Confirmation key: %s\nUse 'k8s-agent confirm %s' to approve", confirmKey, confirmKey)
	}

	// Execute directly if forced
	result, err := rc.executor.Execute(clusterName, op)
	if err != nil {
		return fmt.Errorf("execution failed: %w", err)
	}

	if result.Success {
		fmt.Println(result.Output)
	} else {
		return fmt.Errorf("operation failed: %v", result.Error)
	}

	return nil
}
