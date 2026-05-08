package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s-agent/pkg/engine"
)

func (rc *RootCommand) newCreateCommand() *cobra.Command {
	createCmd := &cobra.Command{
		Use:   "create <resource> <spec>",
		Short: "Create a Kubernetes resource",
		Long:  `Create a Kubernetes resource from a specification. This operation requires confirmation.`,
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return rc.executeCreate(cmd, args)
		},
	}

	createCmd.Flags().StringP("namespace", "n", "default", "Namespace for the resource")
	createCmd.Flags().Bool("force", false, "Skip confirmation")

	return createCmd
}

func (rc *RootCommand) executeCreate(cmd *cobra.Command, args []string) error {
	resource := args[0]
	spec := args[1]

	clusterName := rc.getCurrentCluster()
	if clusterName == "" {
		clusterName = "default"
	}

	namespace, _ := cmd.Flags().GetString("namespace")
	force, _ := cmd.Flags().GetBool("force")

	op := &engine.ClassifiedOperation{
		Type:      engine.OperationTypeMutation,
		Verb:      "create",
		Resource:  resource,
		Name:      spec,
		Namespace: namespace,
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
