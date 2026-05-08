package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"k8s-agent/pkg/engine"
	"k8s-agent/pkg/scheduler"
)

func (rc *RootCommand) newTaskCommand() *cobra.Command {
	taskCmd := &cobra.Command{
		Use:   "task",
		Short: "Manage scheduled tasks",
		Long:  `Manage scheduled inspection tasks.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			return fmt.Errorf("unknown task command: %s", args[0])
		},
	}

	taskCmd.AddCommand(rc.newTaskListCommand())
	taskCmd.AddCommand(rc.newTaskCreateCommand())
	taskCmd.AddCommand(rc.newTaskDeleteCommand())
	taskCmd.AddCommand(rc.newTaskRunCommand())
	taskCmd.AddCommand(rc.newTaskResultsCommand())

	return taskCmd
}

func (rc *RootCommand) newTaskListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List scheduled tasks",
		Long:  `List all scheduled tasks.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			tasks := rc.schedulerMgr.ListTasks()
			if len(tasks) == 0 {
				fmt.Println("No scheduled tasks.")
				return nil
			}

			fmt.Println("Scheduled tasks:")
			for _, task := range tasks {
				status := "disabled"
				if task.Enabled {
					status = "enabled"
				}
				fmt.Printf("  - %s (%s)\n", task.Name, status)
				fmt.Printf("    ID: %s\n", task.ID)
				fmt.Printf("    Schedule: %s\n", task.CronExpr)
				fmt.Printf("    Cluster: %s\n", task.TargetCluster)
				if !task.NextRunAt.IsZero() {
					fmt.Printf("    Next run: %s\n", task.NextRunAt.Format(time.RFC3339))
				}
			}
			return nil
		},
	}
}

func (rc *RootCommand) newTaskCreateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "create <name> <cron> <operation>",
		Short: "Create a scheduled task",
		Long:  `Create a new scheduled task with a cron expression.`,
		Args:  cobra.MinimumNArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 3 {
				return fmt.Errorf("requires at least 3 args")
			}
			name := args[0]
			cronExpr := args[1]
			operationInput := args[2]

			// Parse the operation
			parsedOp, err := engine.Parse(operationInput)
			if err != nil || parsedOp == nil {
				return fmt.Errorf("invalid operation: %s", operationInput)
			}

			classifiedOp := engine.ClassifyOperation(parsedOp)

			// Create the task
			task := &scheduler.ScheduledTask{
				ID:            name,
				Name:          name,
				CronExpr:      cronExpr,
				TargetCluster: rc.getCurrentCluster(),
				Operation:     classifiedOp,
				Enabled:       true,
			}

			err = rc.schedulerMgr.AddTask(task)
			if err != nil {
				return fmt.Errorf("failed to create task: %w", err)
			}

			fmt.Printf("Task '%s' created successfully.\n", name)
			return nil
		},
	}
}

func (rc *RootCommand) newTaskDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a task",
		Long:  `Delete a scheduled task by ID.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("requires exactly one arg")
			}
			taskID := args[0]

			err := rc.schedulerMgr.RemoveTask(taskID)
			if err != nil {
				return fmt.Errorf("failed to delete task: %w", err)
			}

			fmt.Printf("Task '%s' deleted successfully.\n", taskID)
			return nil
		},
	}
}

func (rc *RootCommand) newTaskRunCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "run <id>",
		Short: "Run a task immediately",
		Long:  `Run a scheduled task immediately without waiting for its next scheduled time.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("requires exactly one arg")
			}
			taskID := args[0]

			err := rc.schedulerMgr.RunTaskManually(taskID)
			if err != nil {
				return fmt.Errorf("failed to run task: %w", err)
			}

			fmt.Printf("Task '%s' executed successfully.\n", taskID)
			return nil
		},
	}
}

func (rc *RootCommand) newTaskResultsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "results <id>",
		Short: "Get task results",
		Long:  `Get the execution results of a task.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("requires exactly one arg")
			}
			taskID := args[0]

			results := rc.schedulerMgr.GetTaskResults(taskID)
			if len(results) == 0 {
				fmt.Printf("No results available for task '%s'.\n", taskID)
				return nil
			}

			for _, result := range results {
				status := "SUCCESS"
				if !result.Success {
					status = "FAILED"
				}
				fmt.Printf("Status: %s\n", status)
				fmt.Printf("Executed at: %s\n", result.ExecutedAt.Format(time.RFC3339))
				if result.Output != "" {
					fmt.Printf("Output: %s\n", result.Output)
				}
				if result.Error != "" {
					fmt.Printf("Error: %s\n", result.Error)
				}
			}

			return nil
		},
	}
}
