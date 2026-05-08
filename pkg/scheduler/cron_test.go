package scheduler

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseCronExpression_Valid(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		wantErr bool
	}{
		{
			name:    "every minute",
			expr:    "* * * * *",
			wantErr: false,
		},
		{
			name:    "every 5 minutes",
			expr:    "*/5 * * * *",
			wantErr: false,
		},
		{
			name:    "every hour at minute 30",
			expr:    "30 * * * *",
			wantErr: false,
		},
		{
			name:    "every day at midnight",
			expr:    "0 0 * * *",
			wantErr: false,
		},
		{
			name:    "every Monday at 9am",
			expr:    "0 9 * * 1",
			wantErr: false,
		},
		{
			name:    "every weekday at 9am",
			expr:    "0 9 * * 1-5",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseCronExpression(tt.expr)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseCronExpression_Invalid(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		wantErr bool
	}{
		{
			name:    "empty string",
			expr:    "",
			wantErr: true,
		},
		{
			name:    "invalid format",
			expr:    "invalid",
			wantErr: true,
		},
		{
			name:    "too many fields",
			expr:    "* * * * * *",
			wantErr: true,
		},
		{
			name:    "too few fields",
			expr:    "* * *",
			wantErr: true,
		},
		{
			name:    "invalid minute value",
			expr:    "60 * * * *",
			wantErr: true,
		},
		{
			name:    "invalid hour value",
			expr:    "* 25 * * *",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseCronExpression(tt.expr)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetNextRun(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name    string
		expr    string
		wantErr bool
		checkFn func(t *testing.T, next time.Time)
	}{
		{
			name:    "every minute",
			expr:    "* * * * *",
			wantErr: false,
			checkFn: func(t *testing.T, next time.Time) {
				// Next run should be within the next minute
				assert.True(t, next.After(now) || next.Equal(now))
				assert.True(t, next.Before(now.Add(2*time.Minute)))
			},
		},
		{
			name:    "every hour",
			expr:    "0 * * * *",
			wantErr: false,
			checkFn: func(t *testing.T, next time.Time) {
				// Next run should be at the next hour
				assert.True(t, next.After(now) || next.Equal(now))
				assert.True(t, next.Before(now.Add(2*time.Hour)))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schedule, err := ParseCronExpression(tt.expr)
			require.NoError(t, err)

			next := GetNextRun(schedule)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				tt.checkFn(t, next)
			}
		})
	}
}

func TestGetNextRun_InvalidSchedule(t *testing.T) {
	next := GetNextRun(nil)
	assert.True(t, next.IsZero())
}

func TestParseCronExpression_Standard5Fields(t *testing.T) {
	// robfig/cron v3 uses standard 5-field cron by default
	schedule, err := ParseCronExpression("*/10 * * * *")
	require.NoError(t, err)
	assert.NotNil(t, schedule)

	// Schedule should produce next run times
	next := schedule.Next(time.Now())
	assert.False(t, next.IsZero())
	assert.True(t, next.After(time.Now()))
}
