package cli

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s-agent/pkg/engine"
	"k8s-agent/pkg/scheduler"
)

func TestTaskListCommand(t *testing.T) {
	cmd := NewRootCommand()
	listCmd, _, err := cmd.Find([]string{"task", "list"})
	require.NoError(t, err)

	assert.Equal(t, "list", listCmd.Name())
	assert.NotNil(t, listCmd.RunE)
}

func TestTaskCreateCommand(t *testing.T) {
	cmd := NewRootCommand()
	createCmd, _, err := cmd.Find([]string{"task", "create"})
	require.NoError(t, err)

	assert.Equal(t, "create", createCmd.Name())

	// Test with missing args
	buf := &bytes.Buffer{}
	createCmd.SetOut(buf)
	createCmd.SetErr(buf)

	err = createCmd.RunE(createCmd, []string{}) // Missing all args
	assert.Error(t, err)
}

func TestTaskDeleteCommand(t *testing.T) {
	cmd := NewRootCommand()
	deleteCmd, _, err := cmd.Find([]string{"task", "delete"})
	require.NoError(t, err)

	assert.Equal(t, "delete", deleteCmd.Name())

	// Test with missing args
	buf := &bytes.Buffer{}
	deleteCmd.SetOut(buf)
	deleteCmd.SetErr(buf)

	err = deleteCmd.RunE(deleteCmd, []string{}) // Missing task ID
	assert.Error(t, err)
}

func TestTaskRunCommand(t *testing.T) {
	cmd := NewRootCommand()
	runCmd, _, err := cmd.Find([]string{"task", "run"})
	require.NoError(t, err)

	assert.Equal(t, "run", runCmd.Name())

	// Test with missing args
	buf := &bytes.Buffer{}
	runCmd.SetOut(buf)
	runCmd.SetErr(buf)

	err = runCmd.RunE(runCmd, []string{}) // Missing task ID
	assert.Error(t, err)
}

func TestTaskResultsCommand(t *testing.T) {
	cmd := NewRootCommand()
	resultsCmd, _, err := cmd.Find([]string{"task", "results"})
	require.NoError(t, err)

	assert.Equal(t, "results", resultsCmd.Name())

	// Test with missing args
	buf := &bytes.Buffer{}
	resultsCmd.SetOut(buf)
	resultsCmd.SetErr(buf)

	err = resultsCmd.RunE(resultsCmd, []string{}) // Missing task ID
	assert.Error(t, err)
}

// MockSchedulerManager is a mock for testing
type MockSchedulerManager struct {
	tasks map[string]*scheduler.ScheduledTask
}

func NewMockSchedulerManager() *MockSchedulerManager {
	return &MockSchedulerManager{
		tasks: make(map[string]*scheduler.ScheduledTask),
	}
}

func (m *MockSchedulerManager) AddTask(task *scheduler.ScheduledTask) error {
	if task.ID == "" {
		return nil
	}
	if _, exists := m.tasks[task.ID]; exists {
		return nil
	}
	m.tasks[task.ID] = task
	return nil
}

func (m *MockSchedulerManager) ListTasks() []*scheduler.ScheduledTask {
	result := make([]*scheduler.ScheduledTask, 0, len(m.tasks))
	for _, task := range m.tasks {
		result = append(result, task)
	}
	return result
}

func (m *MockSchedulerManager) GetTask(taskID string) (*scheduler.ScheduledTask, error) {
	if task, ok := m.tasks[taskID]; ok {
		return task, nil
	}
	return nil, nil
}

func (m *MockSchedulerManager) RemoveTask(taskID string) error {
	if _, exists := m.tasks[taskID]; exists {
		delete(m.tasks, taskID)
		return nil
	}
	return nil
}

func (m *MockSchedulerManager) RunTaskManually(taskID string) error {
	if _, exists := m.tasks[taskID]; !exists {
		return nil
	}
	return nil
}

func TestSchedulerManager_MockAddListTasks(t *testing.T) {
	mock := NewMockSchedulerManager()

	// Add a task
	task := &scheduler.ScheduledTask{
		ID:            "task-1",
		Name:          "Test Task",
		CronExpr:      "*/5 * * * *",
		TargetCluster: "test-cluster",
		Operation: &engine.ClassifiedOperation{
			Type:     engine.OperationTypeQuery,
			Verb:     "get",
			Resource: "pods",
		},
		Enabled: true,
	}
	err := mock.AddTask(task)
	require.NoError(t, err)

	// List tasks
	tasks := mock.ListTasks()
	assert.Len(t, tasks, 1)
	assert.Equal(t, "task-1", tasks[0].ID)
}

func TestSchedulerManager_MockRemoveTask(t *testing.T) {
	mock := NewMockSchedulerManager()

	// Add a task
	task := &scheduler.ScheduledTask{
		ID:       "task-1",
		Name:     "Test Task",
		CronExpr: "*/5 * * * *",
	}
	err := mock.AddTask(task)
	require.NoError(t, err)

	// Remove task
	err = mock.RemoveTask("task-1")
	require.NoError(t, err)

	// Verify removed
	tasks := mock.ListTasks()
	assert.Len(t, tasks, 0)
}

func TestSchedulerManager_MockGetTask(t *testing.T) {
	mock := NewMockSchedulerManager()

	// Add a task
	task := &scheduler.ScheduledTask{
		ID:       "task-1",
		Name:     "Test Task",
		CronExpr: "*/5 * * * *",
	}
	err := mock.AddTask(task)
	require.NoError(t, err)

	// Get task
	retrieved, err := mock.GetTask("task-1")
	require.NoError(t, err)
	assert.Equal(t, "task-1", retrieved.ID)
	assert.Equal(t, "Test Task", retrieved.Name)
}

func TestSchedulerManager_MockGetNonExistentTask(t *testing.T) {
	mock := NewMockSchedulerManager()

	// Get non-existent task
	retrieved, err := mock.GetTask("non-existent")
	assert.NoError(t, err)
	assert.Nil(t, retrieved)
}

func TestSchedulerManager_MockRunTask(t *testing.T) {
	mock := NewMockSchedulerManager()

	// Add a task
	task := &scheduler.ScheduledTask{
		ID:       "task-1",
		Name:     "Test Task",
		CronExpr: "*/5 * * * *",
	}
	err := mock.AddTask(task)
	require.NoError(t, err)

	// Run task
	err = mock.RunTaskManually("task-1")
	assert.NoError(t, err)

	// Run non-existent task should not error
	err = mock.RunTaskManually("non-existent")
	assert.NoError(t, err)
}

func TestTaskSubCommandExecution(t *testing.T) {
	cmd := NewRootCommand()

	// Test task list with no tasks
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"task", "list"})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestParseCronExpression(t *testing.T) {
	tests := []struct {
		name      string
		expr      string
		shouldErr bool
	}{
		{
			name:      "valid expression - every 5 minutes",
			expr:      "*/5 * * * *",
			shouldErr: false,
		},
		{
			name:      "valid expression - daily at midnight",
			expr:      "0 0 * * *",
			shouldErr: false,
		},
		{
			name:      "valid expression - weekly",
			expr:      "0 0 * * 0",
			shouldErr: false,
		},
		{
			name:      "invalid expression",
			expr:      "invalid",
			shouldErr: true,
		},
		{
			name:      "empty expression",
			expr:      "",
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := scheduler.ParseCronExpression(tt.expr)
			if tt.shouldErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBuildTaskFromArgs(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		expectID   string
		expectCron string
		expectErr  bool
	}{
		{
			name:       "valid args",
			args:       []string{"my-task", "*/5 * * * *", "get pods"},
			expectID:   "my-task",
			expectCron: "*/5 * * * *",
		},
		{
			name:      "missing name",
			args:      []string{"*/5 * * * *", "get pods"},
			expectErr: true,
		},
		{
			name:      "missing cron",
			args:      []string{"my-task", "get pods"},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task, err := buildTaskFromArgs(tt.args)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectID, task.ID)
				assert.Equal(t, tt.expectCron, task.CronExpr)
			}
		})
	}
}

// Helper to build task from CLI args
func buildTaskFromArgs(args []string) (*scheduler.ScheduledTask, error) {
	if len(args) < 3 {
		return nil, fmt.Errorf("requires 3 args")
	}
	return &scheduler.ScheduledTask{
		ID:       args[0],
		CronExpr: args[1],
		Name:     args[0],
	}, nil
}
