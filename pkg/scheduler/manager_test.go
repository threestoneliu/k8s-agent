package scheduler

import (
	"errors"
	"sync"
	"testing"
	"time"

	"k8s-agent/pkg/engine"

	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockExecutor is a mock implementation of the executor for testing
type mockExecutor struct {
	executeFunc func(clusterName string, op *engine.ClassifiedOperation) (*engine.ExecutionResult, error)
}

func (m *mockExecutor) Execute(clusterName string, op *engine.ClassifiedOperation) (*engine.ExecutionResult, error) {
	if m.executeFunc != nil {
		return m.executeFunc(clusterName, op)
	}
	return &engine.ExecutionResult{
		Success:  true,
		Output:   "mock output",
		Resource: op.Resource,
		Verb:     op.Verb,
	}, nil
}

func newTestManager() *Manager {
	return &Manager{
		tasks:           make(map[string]*ScheduledTask),
		taskCronEntries: make(map[string]cron.EntryID),
		cron:            cron.New(),
		executor:        nil, // Will be set in tests
	}
}

func newMockExecutor() *mockExecutor {
	return &mockExecutor{}
}

func TestManager_NewManager(t *testing.T) {
	m := NewManager(nil)
	assert.NotNil(t, m)
	assert.NotNil(t, m.tasks)
	assert.NotNil(t, m.cron)
}

func TestManager_AddTask_ValidTask(t *testing.T) {
	m := newTestManager()
	m.executor = newMockExecutor()

	task := &ScheduledTask{
		ID:            "task-1",
		Name:          "Test Task",
		Description:   "A test task",
		CronExpr:      "*/5 * * * *",
		TargetCluster: "cluster1",
		Namespace:     "default",
		Operation:     &engine.ClassifiedOperation{Type: engine.OperationTypeQuery, Verb: "get", Resource: "pods"},
		Enabled:       true,
	}

	err := m.AddTask(task)
	assert.NoError(t, err)

	// Verify task was added
	retrieved, err := m.GetTask("task-1")
	assert.NoError(t, err)
	assert.Equal(t, "task-1", retrieved.ID)
	assert.Equal(t, "Test Task", retrieved.Name)
}

func TestManager_AddTask_InvalidCron(t *testing.T) {
	m := newTestManager()
	m.executor = newMockExecutor()

	task := &ScheduledTask{
		ID:            "task-invalid-cron",
		Name:          "Invalid Cron Task",
		CronExpr:      "invalid cron",
		TargetCluster: "cluster1",
		Operation:     &engine.ClassifiedOperation{Type: engine.OperationTypeQuery, Verb: "get", Resource: "pods"},
		Enabled:       true,
	}

	err := m.AddTask(task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid cron expression")
}

func TestManager_AddTask_DuplicateID(t *testing.T) {
	m := newTestManager()
	m.executor = newMockExecutor()

	task := &ScheduledTask{
		ID:            "duplicate-task",
		Name:          "Task 1",
		CronExpr:      "*/5 * * * *",
		TargetCluster: "cluster1",
		Operation:     &engine.ClassifiedOperation{Type: engine.OperationTypeQuery, Verb: "get", Resource: "pods"},
		Enabled:       true,
	}

	err := m.AddTask(task)
	assert.NoError(t, err)

	// Try to add same ID again
	err = m.AddTask(task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestManager_AddTask_NilOperation(t *testing.T) {
	m := newTestManager()
	m.executor = newMockExecutor()

	task := &ScheduledTask{
		ID:            "task-no-op",
		Name:          "No Operation Task",
		CronExpr:      "*/5 * * * *",
		TargetCluster: "cluster1",
		Enabled:       true,
	}

	err := m.AddTask(task)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "operation is required")
}

func TestManager_RemoveTask(t *testing.T) {
	m := newTestManager()
	m.executor = newMockExecutor()

	task := &ScheduledTask{
		ID:            "task-to-remove",
		Name:          "Task to Remove",
		CronExpr:      "*/5 * * * *",
		TargetCluster: "cluster1",
		Operation:     &engine.ClassifiedOperation{Type: engine.OperationTypeQuery, Verb: "get", Resource: "pods"},
		Enabled:       true,
	}

	err := m.AddTask(task)
	require.NoError(t, err)

	// Remove task
	err = m.RemoveTask("task-to-remove")
	assert.NoError(t, err)

	// Verify task was removed
	_, err = m.GetTask("task-to-remove")
	assert.Error(t, err)
}

func TestManager_RemoveTask_NotFound(t *testing.T) {
	m := newTestManager()
	m.executor = newMockExecutor()

	err := m.RemoveTask("non-existent-task")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestManager_GetTask(t *testing.T) {
	m := newTestManager()
	m.executor = newMockExecutor()

	task := &ScheduledTask{
		ID:            "get-test-task",
		Name:          "Get Test Task",
		CronExpr:      "*/5 * * * *",
		TargetCluster: "cluster1",
		Operation:     &engine.ClassifiedOperation{Type: engine.OperationTypeQuery, Verb: "get", Resource: "pods"},
		Enabled:       true,
	}

	err := m.AddTask(task)
	require.NoError(t, err)

	// Get existing task
	retrieved, err := m.GetTask("get-test-task")
	assert.NoError(t, err)
	assert.Equal(t, "get-test-task", retrieved.ID)

	// Get non-existent task
	_, err = m.GetTask("non-existent")
	assert.Error(t, err)
}

func TestManager_ListTasks(t *testing.T) {
	m := newTestManager()
	m.executor = newMockExecutor()

	// Add multiple tasks
	tasks := []*ScheduledTask{
		{
			ID:            "list-task-1",
			Name:          "List Task 1",
			CronExpr:      "*/5 * * * *",
			TargetCluster: "cluster1",
			Operation:     &engine.ClassifiedOperation{Type: engine.OperationTypeQuery, Verb: "get", Resource: "pods"},
			Enabled:       true,
		},
		{
			ID:            "list-task-2",
			Name:          "List Task 2",
			CronExpr:      "*/10 * * * *",
			TargetCluster: "cluster2",
			Operation:     &engine.ClassifiedOperation{Type: engine.OperationTypeQuery, Verb: "list", Resource: "services"},
			Enabled:       true,
		},
	}

	for _, task := range tasks {
		err := m.AddTask(task)
		require.NoError(t, err)
	}

	// List all tasks
	listed := m.ListTasks()
	assert.Len(t, listed, 2)
}

func TestManager_ListTasks_Empty(t *testing.T) {
	m := newTestManager()
	m.executor = newMockExecutor()

	listed := m.ListTasks()
	assert.Len(t, listed, 0)
}

func TestManager_EnableTask(t *testing.T) {
	m := newTestManager()
	m.executor = newMockExecutor()

	task := &ScheduledTask{
		ID:            "enable-task",
		Name:          "Enable Test Task",
		CronExpr:      "*/5 * * * *",
		TargetCluster: "cluster1",
		Operation:     &engine.ClassifiedOperation{Type: engine.OperationTypeQuery, Verb: "get", Resource: "pods"},
		Enabled:       false,
	}

	err := m.AddTask(task)
	require.NoError(t, err)

	// Verify task is disabled
	retrieved, _ := m.GetTask("enable-task")
	assert.False(t, retrieved.Enabled)

	// Enable task
	err = m.EnableTask("enable-task")
	assert.NoError(t, err)

	// Verify task is now enabled
	retrieved, _ = m.GetTask("enable-task")
	assert.True(t, retrieved.Enabled)
}

func TestManager_EnableTask_NotFound(t *testing.T) {
	m := newTestManager()
	m.executor = newMockExecutor()

	err := m.EnableTask("non-existent-task")
	assert.Error(t, err)
}

func TestManager_DisableTask(t *testing.T) {
	m := newTestManager()
	m.executor = newMockExecutor()

	task := &ScheduledTask{
		ID:            "disable-task",
		Name:          "Disable Test Task",
		CronExpr:      "*/5 * * * *",
		TargetCluster: "cluster1",
		Operation:     &engine.ClassifiedOperation{Type: engine.OperationTypeQuery, Verb: "get", Resource: "pods"},
		Enabled:       true,
	}

	err := m.AddTask(task)
	require.NoError(t, err)

	// Disable task
	err = m.DisableTask("disable-task")
	assert.NoError(t, err)

	// Verify task is now disabled
	retrieved, _ := m.GetTask("disable-task")
	assert.False(t, retrieved.Enabled)
}

func TestManager_DisableTask_NotFound(t *testing.T) {
	m := newTestManager()
	m.executor = newMockExecutor()

	err := m.DisableTask("non-existent-task")
	assert.Error(t, err)
}

func TestManager_GetTaskResults(t *testing.T) {
	m := newTestManager()
	m.executor = newMockExecutor()

	task := &ScheduledTask{
		ID:            "results-task",
		Name:          "Results Test Task",
		CronExpr:      "*/5 * * * *",
		TargetCluster: "cluster1",
		Operation:     &engine.ClassifiedOperation{Type: engine.OperationTypeQuery, Verb: "get", Resource: "pods"},
		Enabled:       true,
	}

	err := m.AddTask(task)
	require.NoError(t, err)

	// Get results (should be empty initially)
	results := m.GetTaskResults("results-task")
	assert.Len(t, results, 0)
}

func TestManager_ConcurrentAccess(t *testing.T) {
	m := newTestManager()
	m.executor = newMockExecutor()

	var wg sync.WaitGroup
	numGoroutines := 10

	// Concurrently add tasks
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			task := &ScheduledTask{
				ID:            string(rune('a' + id)),
				Name:          "Concurrent Task",
				CronExpr:      "*/5 * * * *",
				TargetCluster: "cluster1",
				Operation:     &engine.ClassifiedOperation{Type: engine.OperationTypeQuery, Verb: "get", Resource: "pods"},
				Enabled:       true,
			}
			_ = m.AddTask(task)
		}(i)
	}

	wg.Wait()

	// Verify all tasks were added
	listed := m.ListTasks()
	assert.Len(t, listed, numGoroutines)
}

func TestManager_UpdateTaskNextRun(t *testing.T) {
	m := newTestManager()
	m.executor = newMockExecutor()

	task := &ScheduledTask{
		ID:            "nextrun-task",
		Name:          "NextRun Test Task",
		CronExpr:      "*/5 * * * *",
		TargetCluster: "cluster1",
		Operation:     &engine.ClassifiedOperation{Type: engine.OperationTypeQuery, Verb: "get", Resource: "pods"},
		Enabled:       true,
	}

	err := m.AddTask(task)
	require.NoError(t, err)

	// Verify NextRunAt was set
	retrieved, _ := m.GetTask("nextrun-task")
	assert.False(t, retrieved.NextRunAt.IsZero())
	assert.True(t, retrieved.NextRunAt.After(time.Now()))
}

func TestManager_RunTaskManually(t *testing.T) {
	m := newTestManager()

	executed := false
	mockExec := &mockExecutor{
		executeFunc: func(clusterName string, op *engine.ClassifiedOperation) (*engine.ExecutionResult, error) {
			executed = true
			return &engine.ExecutionResult{
				Success:  true,
				Output:   "test output",
				Resource: op.Resource,
				Verb:     op.Verb,
			}, nil
		},
	}
	m.executor = mockExec

	task := &ScheduledTask{
		ID:            "manual-run-task",
		Name:          "Manual Run Task",
		CronExpr:      "*/5 * * * *",
		TargetCluster: "cluster1",
		Operation:     &engine.ClassifiedOperation{Type: engine.OperationTypeQuery, Verb: "get", Resource: "pods"},
		Enabled:       true,
	}

	err := m.AddTask(task)
	require.NoError(t, err)

	// Run task manually
	err = m.RunTaskManually("manual-run-task")
	assert.NoError(t, err)
	assert.True(t, executed)
}

func TestManager_RunTaskManually_NotFound(t *testing.T) {
	m := newTestManager()
	m.executor = newMockExecutor()

	err := m.RunTaskManually("non-existent-task")
	assert.Error(t, err)
}

func TestManager_RunTaskManually_ExecutorError(t *testing.T) {
	m := newTestManager()

	mockExec := &mockExecutor{
		executeFunc: func(clusterName string, op *engine.ClassifiedOperation) (*engine.ExecutionResult, error) {
			return nil, errors.New("executor error")
		},
	}
	m.executor = mockExec

	task := &ScheduledTask{
		ID:            "exec-error-task",
		Name:          "Exec Error Task",
		CronExpr:      "*/5 * * * *",
		TargetCluster: "cluster1",
		Operation:     &engine.ClassifiedOperation{Type: engine.OperationTypeQuery, Verb: "get", Resource: "pods"},
		Enabled:       true,
	}

	err := m.AddTask(task)
	require.NoError(t, err)

	// Run task should handle executor error
	err = m.RunTaskManually("exec-error-task")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "executor error")
}
