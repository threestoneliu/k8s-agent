package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s-agent/pkg/engine"
)

func (rc *RootCommand) newDeleteCommand() *cobra.Command {
	deleteCmd := &cobra.Command{
		Use:   "delete <resource> <name>",
		Short: "Delete a Kubernetes resource",
		Long:  `Delete a Kubernetes resource. This operation requires confirmation.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return rc.executeDelete(cmd, args)
		},
	}

	deleteCmd.Flags().StringP("namespace", "n", "default", "Namespace of the resource")
	deleteCmd.Flags().Bool("force", false, "Skip confirmation")

	return deleteCmd
}

func (rc *RootCommand) executeDelete(cmd *cobra.Command, args []string) error {
	resource := args[0]
	name := args[1]

	clusterName := rc.getCurrentCluster()
	if clusterName == "" {
		clusterName = "default"
	}

	namespace, _ := cmd.Flags().GetString("namespace")
	force, _ := cmd.Flags().GetBool("force")

	op := &engine.ClassifiedOperation{
		Type:      engine.OperationTypeMutation,
		Verb:      "delete",
		Resource:  resource,
		Name:      name,
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
