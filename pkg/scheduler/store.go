package scheduler

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

// DefaultTasksPath is the default path for storing tasks
var DefaultTasksPath = filepath.Join(os.Getenv("HOME"), ".config", "k8s-agent", "tasks.yaml")

// Store provides persistent storage for scheduled tasks
type Store struct {
	tasksPath string
	mu       sync.RWMutex
}

// NewStore creates a new task store
func NewStore(tasksPath string) (*Store, error) {
	if tasksPath == "" {
		tasksPath = DefaultTasksPath
	}

	s := &Store{
		tasksPath: tasksPath,
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(tasksPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create tasks directory: %w", err)
	}

	return s, nil
}

// SaveTasks saves all tasks to file
func (s *Store) SaveTasks(tasks []*ScheduledTask) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := yaml.Marshal(tasks)
	if err != nil {
		return fmt.Errorf("failed to marshal tasks: %w", err)
	}

	if err := os.WriteFile(s.tasksPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write tasks file: %w", err)
	}

	return nil
}

// LoadTasks loads all tasks from file
func (s *Store) LoadTasks() ([]*ScheduledTask, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := os.ReadFile(s.tasksPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []*ScheduledTask{}, nil
		}
		return nil, fmt.Errorf("failed to read tasks file: %w", err)
	}

	if len(data) == 0 {
		return []*ScheduledTask{}, nil
	}

	var tasks []*ScheduledTask
	if err := yaml.Unmarshal(data, &tasks); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tasks: %w", err)
	}

	return tasks, nil
}
