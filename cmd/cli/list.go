package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s-agent/pkg/engine"
)

func (rc *RootCommand) newListCommand() *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list <resource>",
		Short: "List Kubernetes resources",
		Long:  `List Kubernetes resources in a namespace or cluster.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return rc.executeList(cmd, args)
		},
	}

	listCmd.Flags().StringP("namespace", "n", "default", "Namespace to list resources from")

	return listCmd
}

func (rc *RootCommand) executeList(cmd *cobra.Command, args []string) error {
	resource := args[0]

	clusterName := rc.getCurrentCluster()
	if clusterName == "" {
		clusterName = "default"
	}

	namespace, _ := cmd.Flags().GetString("namespace")

	op := &engine.ClassifiedOperation{
		Type:      engine.OperationTypeQuery,
		Verb:      "list",
		Resource:  resource,
		Namespace: namespace,
	}

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
