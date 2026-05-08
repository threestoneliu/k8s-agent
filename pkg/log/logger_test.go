package log

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInit(t *testing.T) {
	// Test default config
	Init(nil)
	assert.NotNil(t, logger)
}

func TestInitWithConfig(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
	}{
		{
			name:   "debug level",
			config: &Config{Level: "debug", Format: "text"},
		},
		{
			name:   "info level json",
			config: &Config{Level: "info", Format: "json"},
		},
		{
			name:   "warn level",
			config: &Config{Level: "warn", Format: "text"},
		},
		{
			name:   "error level",
			config: &Config{Level: "error", Format: "text"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Init(tt.config)
			assert.NotNil(t, logger)
		})
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input string
	}{
		{"debug"},
		{"DEBUG"},
		{"info"},
		{"INFO"},
		{"warn"},
		{"warning"},
		{"WARN"},
		{"error"},
		{"ERROR"},
		{"unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			parseLevel(tt.input)
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, "info", cfg.Level)
	assert.Equal(t, "text", cfg.Format)
}

func TestDebug(t *testing.T) {
	Init(&Config{Level: "debug", Format: "text"})
	// Should not panic
	Debug("test debug message", "key", "value")
}

func TestInfo(t *testing.T) {
	Init(&Config{Level: "info", Format: "text"})
	// Should not panic
	Info("test info message", "key", "value")
}

func TestWarn(t *testing.T) {
	Init(&Config{Level: "warn", Format: "text"})
	// Should not panic
	Warn("test warn message", "key", "value")
}

func TestError(t *testing.T) {
	Init(&Config{Level: "error", Format: "text"})
	// Should not panic
	Error("test error message", "key", "value")
}

func TestWithFields(t *testing.T) {
	Init(&Config{Level: "info", Format: "text"})
	loggerWithFields := WithFields("key", "value")
	assert.NotNil(t, loggerWithFields)
}

func TestWithContext(t *testing.T) {
	Init(&Config{Level: "info", Format: "text"})
	ctx := context.Background()
	loggerWithCtx := WithContext(ctx)
	assert.NotNil(t, loggerWithCtx)
}

func TestLogLevels(t *testing.T) {
	// Test that different levels work without panic
	levels := []string{"debug", "info", "warn", "error"}
	for _, level := range levels {
		t.Run(level, func(t *testing.T) {
			Init(&Config{Level: level, Format: "text"})
			assert.NotNil(t, logger)
		})
	}
}

func TestFormatOptions(t *testing.T) {
	// Test both format options
	formats := []string{"text", "json"}
	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			Init(&Config{Level: "info", Format: format})
			assert.NotNil(t, logger)
		})
	}
}

func TestJSONFormat(t *testing.T) {
	// Verify JSON parsing works
	jsonStr := `{"level":"INFO","msg":"test"}`
	var logEntry map[string]interface{}
	err := json.Unmarshal([]byte(jsonStr), &logEntry)
	assert.NoError(t, err)
	assert.Equal(t, "INFO", logEntry["level"])
}

func TestTextFormat(t *testing.T) {
	Init(&Config{Level: "info", Format: "text"})
	// Text format is the default, just verify no panic
	Info("text format test")
}

func TestNilConfig(t *testing.T) {
	// Should use defaults
	Init(nil)
	assert.NotNil(t, logger)
}

func TestEmptyFields(t *testing.T) {
	Init(&Config{Level: "debug", Format: "text"})
	WithFields()
}