package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s-agent/pkg/engine"
)

func (rc *RootCommand) newGetCommand() *cobra.Command {
	getCmd := &cobra.Command{
		Use:   "get <resource> [name]",
		Short: "Get Kubernetes resources",
		Long:  `Get one or more Kubernetes resources.`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return rc.executeGet(cmd, args)
		},
	}

	return getCmd
}

func (rc *RootCommand) executeGet(cmd *cobra.Command, args []string) error {
	resource := args[0]
	var name string
	if len(args) > 1 {
		name = args[1]
	}

	clusterName := rc.getCurrentCluster()
	if clusterName == "" {
		clusterName = "default"
	}

	// Parse namespace flag
	namespace, _ := cmd.Flags().GetString("namespace")
	if namespace == "" {
		namespace = "default"
	}

	// Build the operation
	op := &engine.ClassifiedOperation{
		Type:      engine.OperationTypeQuery,
		Verb:      "get",
		Resource:  resource,
		Name:      name,
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
