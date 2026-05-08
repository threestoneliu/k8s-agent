package scheduler

import (
	"fmt"
	"sync"
	"time"

	"k8s-agent/pkg/engine"
	"k8s-agent/pkg/log"

	"github.com/robfig/cron/v3"
)

// ExecutorInterface defines the interface for executing operations
type ExecutorInterface interface {
	Execute(clusterName string, op *engine.ClassifiedOperation) (*engine.ExecutionResult, error)
}

// Manager manages scheduled tasks
type Manager struct {
	tasks           map[string]*ScheduledTask
	executor        ExecutorInterface
	cron            *cron.Cron
	mu              sync.RWMutex
	taskCronEntries map[string]cron.EntryID
	store           *Store
}

// NewManager creates a new scheduler manager
func NewManager(executor ExecutorInterface) *Manager {
	m := &Manager{
		tasks:           make(map[string]*ScheduledTask),
		executor:        executor,
		cron:            cron.New(),
		taskCronEntries: make(map[string]cron.EntryID),
	}
	return m
}

// NewManagerWithStore creates a new scheduler manager with persistence
func NewManagerWithStore(executor ExecutorInterface, store *Store) *Manager {
	m := NewManager(executor)
	m.store = store

	// Load tasks from store
	if store != nil {
		if tasks, err := store.LoadTasks(); err == nil && tasks != nil {
			for _, task := range tasks {
				m.tasks[task.ID] = task
				// Re-add to cron scheduler if enabled
				if task.Enabled && task.CronExpr != "" {
					m.addToCronScheduler(task)
				}
			}
		}
	}

	return m
}

// AddTask adds a new scheduled task
func (m *Manager) AddTask(task *ScheduledTask) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	log.Info("adding scheduled task", "id", task.ID, "name", task.Name, "cron", task.CronExpr)

	// Validate task first
	if task.Operation == nil {
		log.Error("task validation failed", "id", task.ID, "reason", "operation is nil")
		return fmt.Errorf("operation is required for task %s", task.ID)
	}

	if task.CronExpr == "" {
		log.Error("task validation failed", "id", task.ID, "reason", "cron expr is empty")
		return fmt.Errorf("cron expression is required for task %s", task.ID)
	}

	// Validate cron expression
	schedule, err := ParseCronExpression(task.CronExpr)
	if err != nil {
		log.Error("task validation failed", "id", task.ID, "reason", "invalid cron", "error", err)
		return fmt.Errorf("invalid cron expression for task %s: %w", task.ID, err)
	}

	// Check if task ID already exists
	if _, exists := m.tasks[task.ID]; exists {
		log.Warn("task already exists", "id", task.ID)
		return fmt.Errorf("task %s already exists", task.ID)
	}

	// Set NextRunAt
	task.NextRunAt = schedule.Next(time.Now())
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()

	// Store task
	m.tasks[task.ID] = task

	// Persist to store
	if m.store != nil {
		m.saveTasks()
	}

	// If task is enabled, add to cron scheduler
	if task.Enabled {
		m.addToCronScheduler(task)
	}

	return nil
}

// RemoveTask removes a task by ID
func (m *Manager) RemoveTask(taskID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, exists := m.tasks[taskID]
	if !exists {
		return fmt.Errorf("task %s not found", taskID)
	}

	// Remove from cron scheduler if enabled
	if task.Enabled {
		if entryID, ok := m.taskCronEntries[taskID]; ok {
			m.cron.Remove(entryID)
			delete(m.taskCronEntries, taskID)
		}
	}

	delete(m.tasks, taskID)

	// Persist to store
	if m.store != nil {
		m.saveTasks()
	}

	return nil
}

// GetTask retrieves a task by ID
func (m *Manager) GetTask(taskID string) (*ScheduledTask, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	task, exists := m.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task %s not found", taskID)
	}

	return task, nil
}

// ListTasks returns all tasks
func (m *Manager) ListTasks() []*ScheduledTask {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tasks := make([]*ScheduledTask, 0, len(m.tasks))
	for _, task := range m.tasks {
		tasks = append(tasks, task)
	}

	return tasks
}

// EnableTask enables a task
func (m *Manager) EnableTask(taskID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, exists := m.tasks[taskID]
	if !exists {
		return fmt.Errorf("task %s not found", taskID)
	}

	if task.Enabled {
		return nil // Already enabled
	}

	task.Enabled = true
	task.UpdatedAt = time.Now()
	m.addToCronScheduler(task)

	// Persist to store
	if m.store != nil {
		m.saveTasks()
	}

	return nil
}

// DisableTask disables a task
func (m *Manager) DisableTask(taskID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, exists := m.tasks[taskID]
	if !exists {
		return fmt.Errorf("task %s not found", taskID)
	}

	if !task.Enabled {
		return nil // Already disabled
	}

	// Remove from cron scheduler
	if entryID, ok := m.taskCronEntries[taskID]; ok {
		m.cron.Remove(entryID)
		delete(m.taskCronEntries, taskID)
	}

	task.Enabled = false
	task.UpdatedAt = time.Now()

	// Persist to store
	if m.store != nil {
		m.saveTasks()
	}

	return nil
}

// GetTaskResults returns the execution results for a task
func (m *Manager) GetTaskResults(taskID string) []*TaskResult {
	m.mu.RLock()
	defer m.mu.RUnlock()

	task, exists := m.tasks[taskID]
	if !exists {
		return nil
	}

	if task.LastResult == nil {
		return []*TaskResult{}
	}

	return []*TaskResult{task.LastResult}
}

// RunTaskManually runs a task immediately
func (m *Manager) RunTaskManually(taskID string) error {
	m.mu.RLock()
	task, exists := m.tasks[taskID]
	if !exists {
		m.mu.RUnlock()
		return fmt.Errorf("task %s not found", taskID)
	}
	executor := m.executor
	m.mu.RUnlock()

	if executor == nil {
		return fmt.Errorf("executor not configured")
	}

	// Execute the task
	result, err := executor.Execute(task.TargetCluster, task.Operation)
	if err != nil {
		return fmt.Errorf("failed to execute task: %w", err)
	}

	// Record result
	m.mu.Lock()
	defer m.mu.Unlock()

	task.LastRunAt = time.Now()
	task.LastResult = &TaskResult{
		ExecutedAt: task.LastRunAt,
		Success:    result.Success,
		Output:     result.Output,
		Error:      "",
	}
	if result.Error != nil {
		task.LastResult.Error = result.Error.Error()
	}

	// Update NextRunAt if scheduled
	if task.Enabled && task.CronExpr != "" {
		schedule, err := ParseCronExpression(task.CronExpr)
		if err == nil {
			task.NextRunAt = schedule.Next(time.Now())
		}
	}

	return nil
}

// addToCronScheduler adds a task to the cron scheduler
func (m *Manager) addToCronScheduler(task *ScheduledTask) {
	_, err := ParseCronExpression(task.CronExpr)
	if err != nil {
		return
	}

	taskID := task.ID
	entryID, err := m.cron.AddFunc(task.CronExpr, func() {
		m.runTask(taskID)
	})
	if err != nil {
		return
	}

	m.taskCronEntries[taskID] = entryID
}

// runTask executes a scheduled task
func (m *Manager) runTask(taskID string) {
	m.mu.RLock()
	task, exists := m.tasks[taskID]
	executor := m.executor
	m.mu.RUnlock()

	if !exists || executor == nil {
		return
	}

	// Execute the operation (context handled by executor)
	result, err := executor.Execute(task.TargetCluster, task.Operation)

	// Record result
	m.mu.Lock()
	defer m.mu.Unlock()

	task.LastRunAt = time.Now()
	task.LastResult = &TaskResult{
		ExecutedAt: task.LastRunAt,
		Success:    false,
		Error:      "",
	}

	if err != nil {
		task.LastResult.Error = err.Error()
	} else if result != nil {
		task.LastResult.Success = result.Success
		task.LastResult.Output = result.Output
		if result.Error != nil {
			task.LastResult.Error = result.Error.Error()
		}
	}

	// Update NextRunAt
	schedule, err := ParseCronExpression(task.CronExpr)
	if err == nil {
		task.NextRunAt = schedule.Next(time.Now())
	}

	task.UpdatedAt = time.Now()
}

// saveTasks persists tasks to store
func (m *Manager) saveTasks() {
	if m.store == nil {
		return
	}
	tasks := make([]*ScheduledTask, 0, len(m.tasks))
	for _, task := range m.tasks {
		tasks = append(tasks, task)
	}
	m.store.SaveTasks(tasks)
}

// Start starts the cron scheduler
func (m *Manager) Start() error {
	m.cron.Start()
	return nil
}

// Stop stops the cron scheduler
func (m *Manager) Stop() {
	ctx := m.cron.Stop()
	<-ctx.Done()
}
