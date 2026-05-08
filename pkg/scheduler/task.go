package scheduler

import (
	"time"

	"k8s-agent/pkg/engine"
)

// ScheduledTask represents a scheduled inspection task
type ScheduledTask struct {
	ID            string
	Name          string
	Description   string
	CronExpr      string // cron expression
	TargetCluster string // target cluster
	Namespace     string // namespace
	Operation     *engine.ClassifiedOperation
	Enabled       bool
	LastRunAt     time.Time
	NextRunAt     time.Time
	LastResult    *TaskResult
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// TaskResult holds the result of a task execution
type TaskResult struct {
	ExecutedAt time.Time
	Success    bool
	Output     string
	Error      string
}

// NotificationConfig configures notifications for task events
type NotificationConfig struct {
	OnSuccess bool
	OnFailure bool
	Webhook   string
}
