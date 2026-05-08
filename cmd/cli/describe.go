package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s-agent/pkg/engine"
)

func (rc *RootCommand) newDescribeCommand() *cobra.Command {
	describeCmd := &cobra.Command{
		Use:   "describe <resource> <name>",
		Short: "Describe a Kubernetes resource",
		Long:  `Show detailed information about a specific Kubernetes resource.`,
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return rc.executeDescribe(cmd, args)
		},
	}

	describeCmd.Flags().StringP("namespace", "n", "default", "Namespace of the resource")

	return describeCmd
}

func (rc *RootCommand) executeDescribe(cmd *cobra.Command, args []string) error {
	resource := args[0]
	name := args[1]

	clusterName := rc.getCurrentCluster()
	if clusterName == "" {
		clusterName = "default"
	}

	namespace, _ := cmd.Flags().GetString("namespace")

	op := &engine.ClassifiedOperation{
		Type:      engine.OperationTypeQuery,
		Verb:      "describe",
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
