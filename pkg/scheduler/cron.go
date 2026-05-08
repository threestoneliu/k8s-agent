package scheduler

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
)

// ParseCronExpression parses a standard 5-field cron expression
// Format: minute hour day month weekday
func ParseCronExpression(expr string) (cron.Schedule, error) {
	if expr == "" {
		return nil, fmt.Errorf("cron expression cannot be empty")
	}

	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	schedule, err := parser.Parse(expr)
	if err != nil {
		return nil, fmt.Errorf("invalid cron expression: %w", err)
	}

	return schedule, nil
}

// GetNextRun returns the next run time for a schedule
func GetNextRun(schedule cron.Schedule) time.Time {
	if schedule == nil {
		return time.Time{}
	}
	return schedule.Next(time.Now())
}

// GetNextRunAt returns the next run time for a cron expression
func GetNextRunAt(expr string) (time.Time, error) {
	schedule, err := ParseCronExpression(expr)
	if err != nil {
		return time.Time{}, err
	}
	return schedule.Next(time.Now()), nil
}
