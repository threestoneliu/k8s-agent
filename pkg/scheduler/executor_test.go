package scheduler

import (
	"errors"
	"testing"
	"time"

	"k8s-agent/pkg/engine"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunTask_Success(t *testing.T) {
	m := newTestManager()

	executed := false
	mockExec := &mockExecutor{
		executeFunc: func(clusterName string, op *engine.ClassifiedOperation) (*engine.ExecutionResult, error) {
			executed = true
			assert.Equal(t, "test-cluster", clusterName)
			assert.Equal(t, "pods", op.Resource)
			return &engine.ExecutionResult{
				Success:  true,
				Output:   "pod list retrieved",
				Resource: op.Resource,
				Verb:     op.Verb,
			}, nil
		},
	}
	m.executor = mockExec

	task := &ScheduledTask{
		ID:            "run-success-task",
		Name:          "Run Success Task",
		CronExpr:      "*/5 * * * *",
		TargetCluster: "test-cluster",
		Operation:     &engine.ClassifiedOperation{Type: engine.OperationTypeQuery, Verb: "list", Resource: "pods"},
		Enabled:       true,
	}

	err := m.AddTask(task)
	require.NoError(t, err)

	// Simulate task execution
	m.runTask("run-success-task")

	assert.True(t, executed)

	// Verify result was recorded
	taskResult := m.GetTaskResults("run-success-task")
	require.Len(t, taskResult, 1)
	assert.True(t, taskResult[0].Success)
	assert.Equal(t, "pod list retrieved", taskResult[0].Output)
}

func TestRunTask_ExecutorError(t *testing.T) {
	m := newTestManager()

	mockExec := &mockExecutor{
		executeFunc: func(clusterName string, op *engine.ClassifiedOperation) (*engine.ExecutionResult, error) {
			return nil, errors.New("connection refused")
		},
	}
	m.executor = mockExec

	task := &ScheduledTask{
		ID:            "run-error-task",
		Name:          "Run Error Task",
		CronExpr:      "*/5 * * * *",
		TargetCluster: "test-cluster",
		Operation:     &engine.ClassifiedOperation{Type: engine.OperationTypeQuery, Verb: "list", Resource: "pods"},
		Enabled:       true,
	}

	err := m.AddTask(task)
	require.NoError(t, err)

	// Simulate task execution
	m.runTask("run-error-task")

	// Verify result was recorded with error
	taskResult := m.GetTaskResults("run-error-task")
	require.Len(t, taskResult, 1)
	assert.False(t, taskResult[0].Success)
	assert.Contains(t, taskResult[0].Error, "connection refused")
}

func TestRunTask_NonExistent(t *testing.T) {
	m := newTestManager()
	m.executor = newMockExecutor()

	// Running a non-existent task should not panic
	m.runTask("non-existent-task")
}

func TestRunTask_NoExecutor(t *testing.T) {
	m := newTestManager()
	// No executor set

	task := &ScheduledTask{
		ID:            "no-exec-task",
		Name:          "No Executor Task",
		CronExpr:      "*/5 * * * *",
		TargetCluster: "test-cluster",
		Operation:     &engine.ClassifiedOperation{Type: engine.OperationTypeQuery, Verb: "list", Resource: "pods"},
		Enabled:       true,
	}

	err := m.AddTask(task)
	require.NoError(t, err)

	// Running without executor should not panic
	m.runTask("no-exec-task")
}

func TestRunTask_UpdatesNextRunAt(t *testing.T) {
	m := newTestManager()

	mockExec := &mockExecutor{
		executeFunc: func(clusterName string, op *engine.ClassifiedOperation) (*engine.ExecutionResult, error) {
			return &engine.ExecutionResult{
				Success:  true,
				Output:   "test",
				Resource: op.Resource,
				Verb:     op.Verb,
			}, nil
		},
	}
	m.executor = mockExec

	task := &ScheduledTask{
		ID:            "next-run-update-task",
		Name:          "Next Run Update Task",
		CronExpr:      "*/5 * * * *",
		TargetCluster: "test-cluster",
		Operation:     &engine.ClassifiedOperation{Type: engine.OperationTypeQuery, Verb: "list", Resource: "pods"},
		Enabled:       true,
	}

	err := m.AddTask(task)
	require.NoError(t, err)

	// Get initial NextRunAt
	initialTask, _ := m.GetTask("next-run-update-task")
	initialNextRun := initialTask.NextRunAt

	// Simulate some time passing
	time.Sleep(10 * time.Millisecond)

	// Run task
	m.runTask("next-run-update-task")

	// Verify NextRunAt was updated
	updatedTask, _ := m.GetTask("next-run-update-task")
	assert.True(t, updatedTask.NextRunAt.After(initialNextRun) || updatedTask.NextRunAt.Equal(initialNextRun))
}

func TestRunTask_UpdatesLastRunAt(t *testing.T) {
	m := newTestManager()

	mockExec := &mockExecutor{
		executeFunc: func(clusterName string, op *engine.ClassifiedOperation) (*engine.ExecutionResult, error) {
			return &engine.ExecutionResult{
				Success:  true,
				Output:   "test",
				Resource: op.Resource,
				Verb:     op.Verb,
			}, nil
		},
	}
	m.executor = mockExec

	task := &ScheduledTask{
		ID:            "last-run-task",
		Name:          "Last Run Task",
		CronExpr:      "*/5 * * * *",
		TargetCluster: "test-cluster",
		Operation:     &engine.ClassifiedOperation{Type: engine.OperationTypeQuery, Verb: "list", Resource: "pods"},
		Enabled:       true,
	}

	err := m.AddTask(task)
	require.NoError(t, err)

	// Initially LastRunAt should be zero
	initialTask, _ := m.GetTask("last-run-task")
	assert.True(t, initialTask.LastRunAt.IsZero())

	// Run task
	m.runTask("last-run-task")

	// Verify LastRunAt was updated
	updatedTask, _ := m.GetTask("last-run-task")
	assert.False(t, updatedTask.LastRunAt.IsZero())
}

func TestRunTask_RecordsOutput(t *testing.T) {
	m := newTestManager()

	expectedOutput := "deployment.apps/nginx created"
	mockExec := &mockExecutor{
		executeFunc: func(clusterName string, op *engine.ClassifiedOperation) (*engine.ExecutionResult, error) {
			return &engine.ExecutionResult{
				Success:  true,
				Output:   expectedOutput,
				Resource: op.Resource,
				Verb:     op.Verb,
			}, nil
		},
	}
	m.executor = mockExec

	task := &ScheduledTask{
		ID:            "output-task",
		Name:          "Output Task",
		CronExpr:      "*/5 * * * *",
		TargetCluster: "test-cluster",
		Operation:     &engine.ClassifiedOperation{Type: engine.OperationTypeMutation, Verb: "create", Resource: "deployments"},
		Enabled:       true,
	}

	err := m.AddTask(task)
	require.NoError(t, err)

	// Run task
	m.runTask("output-task")

	// Verify output was recorded
	taskResult := m.GetTaskResults("output-task")
	require.Len(t, taskResult, 1)
	assert.Equal(t, expectedOutput, taskResult[0].Output)
}

func TestRunTask_RecordsError(t *testing.T) {
	m := newTestManager()

	expectedError := "pods \"nonexistent\" not found"
	mockExec := &mockExecutor{
		executeFunc: func(clusterName string, op *engine.ClassifiedOperation) (*engine.ExecutionResult, error) {
			return &engine.ExecutionResult{
				Success:  false,
				Output:   "",
				Error:    errors.New(expectedError),
				Resource: op.Resource,
				Verb:     op.Verb,
			}, nil
		},
	}
	m.executor = mockExec

	task := &ScheduledTask{
		ID:            "error-task",
		Name:          "Error Task",
		CronExpr:      "*/5 * * * *",
		TargetCluster: "test-cluster",
		Operation:     &engine.ClassifiedOperation{Type: engine.OperationTypeQuery, Verb: "get", Resource: "pods"},
		Enabled:       true,
	}

	err := m.AddTask(task)
	require.NoError(t, err)

	// Run task
	m.runTask("error-task")

	// Verify error was recorded
	taskResult := m.GetTaskResults("error-task")
	require.Len(t, taskResult, 1)
	assert.False(t, taskResult[0].Success)
	assert.Equal(t, expectedError, taskResult[0].Error)
}
