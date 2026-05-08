package scheduler

import (
	"testing"
	"time"

	"k8s-agent/pkg/engine"

	"github.com/stretchr/testify/assert"
)

func TestScheduledTask_Defaults(t *testing.T) {
	task := &ScheduledTask{
		ID:            "test-task-1",
		Name:          "Test Task",
		Description:   "A test task",
		CronExpr:      "*/5 * * * *",
		TargetCluster: "cluster1",
		Namespace:     "default",
		Operation:     &engine.ClassifiedOperation{Type: engine.OperationTypeQuery},
		Enabled:       true,
	}

	assert.Equal(t, "test-task-1", task.ID)
	assert.Equal(t, "Test Task", task.Name)
	assert.True(t, task.Enabled)
	assert.NotNil(t, task.Operation)
}

func TestTaskResult_Structure(t *testing.T) {
	now := time.Now()
	result := &TaskResult{
		ExecutedAt: now,
		Success:    true,
		Output:     "pods list retrieved",
		Error:      "",
	}

	assert.Equal(t, now, result.ExecutedAt)
	assert.True(t, result.Success)
	assert.Equal(t, "pods list retrieved", result.Output)
	assert.Empty(t, result.Error)
}

func TestTaskResult_Failure(t *testing.T) {
	now := time.Now()
	result := &TaskResult{
		ExecutedAt: now,
		Success:    false,
		Output:     "",
		Error:      "connection refused",
	}

	assert.Equal(t, now, result.ExecutedAt)
	assert.False(t, result.Success)
	assert.Empty(t, result.Output)
	assert.Equal(t, "connection refused", result.Error)
}

func TestNotificationConfig_Defaults(t *testing.T) {
	config := &NotificationConfig{
		OnSuccess: true,
		OnFailure: true,
		Webhook:   "https://example.com/webhook",
	}

	assert.True(t, config.OnSuccess)
	assert.True(t, config.OnFailure)
	assert.Equal(t, "https://example.com/webhook", config.Webhook)
}

func TestNotificationConfig_SuccessOnly(t *testing.T) {
	config := &NotificationConfig{
		OnSuccess: true,
		OnFailure: false,
		Webhook:   "",
	}

	assert.True(t, config.OnSuccess)
	assert.False(t, config.OnFailure)
	assert.Empty(t, config.Webhook)
}

func TestScheduledTask_WithOperation(t *testing.T) {
	op := &engine.ClassifiedOperation{
		Type:      engine.OperationTypeQuery,
		Verb:      "get",
		Resource:  "pods",
		Name:      "my-pod",
		Namespace: "default",
	}

	task := &ScheduledTask{
		ID:            "task-with-op",
		Name:          "Pod Inspector",
		Description:   "Inspect pod status",
		CronExpr:      "0 * * * *",
		TargetCluster: "prod-cluster",
		Namespace:     "default",
		Operation:     op,
		Enabled:       true,
	}

	assert.NotNil(t, task.Operation)
	assert.Equal(t, engine.OperationTypeQuery, task.Operation.Type)
	assert.Equal(t, "get", task.Operation.Verb)
	assert.Equal(t, "pods", task.Operation.Resource)
}
