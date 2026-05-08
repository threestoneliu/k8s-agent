package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"k8s-agent/pkg/confirmation"
)

func (rc *RootCommand) newConfirmCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "confirm <confirm_key>",
		Short: "Confirm a pending operation",
		Long:  `Approve a pending operation by its confirmation key.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("requires exactly one arg")
			}
			confirmKey := args[0]

			// Validate confirm key format (6 digits)
			if !confirmation.ValidateConfirmKey(confirmKey) {
				return fmt.Errorf("invalid confirmation key format: must be 6 digits")
			}

			// Get the pending operation first
			pending, err := rc.confirmMgr.GetConfirmation(confirmKey)
			if err != nil {
				return fmt.Errorf("failed to get confirmation: %w", err)
			}

			// Approve the confirmation
			err = rc.confirmMgr.ApproveConfirmation(confirmKey)
			if err != nil {
				return fmt.Errorf("failed to confirm operation: %w", err)
			}

			// Execute the confirmed operation
			result := rc.llmExecutor.ExecuteConfirmedOperation(pending.TargetCluster, pending.Operation)
			if !result.Success {
				return fmt.Errorf("operation failed: %s", result.Error)
			}

			fmt.Printf("Operation '%s' confirmed and executed successfully.\n", confirmKey)
			fmt.Printf("Result: %s\n", result.Result)
			return nil
		},
	}
}
