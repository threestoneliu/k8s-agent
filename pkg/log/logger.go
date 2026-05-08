package log

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
)

// Config holds logger configuration
type Config struct {
	Level  string // debug, info, warn, error
	Format string // json, text
	Output string // stdout, file, both (default: file)
}

// DefaultConfig returns the default logger configuration
func DefaultConfig() *Config {
	return &Config{
		Level:  "info",
		Format: "text",
		Output: "file",
	}
}

// getDefaultLogPath returns the default log file path
func getDefaultLogPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "k8s-agent.log"
	}
	configDir := filepath.Join(home, ".config", "k8s-agent")
	// Ensure directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "k8s-agent.log"
	}
	return filepath.Join(configDir, "app.log")
}

// Logger is the global logger instance
var logger *slog.Logger
var originalOutput io.Writer
var originalLevel slog.Level
var logFilePath string

// Init initializes the global logger with the given configuration
func Init(cfg *Config) {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	// Determine output destinations
	var output io.Writer
	switch strings.ToLower(cfg.Output) {
	case "stdout":
		output = os.Stdout
	case "both":
		// Write to file and stdout
		logPath := logFilePath
		if logPath == "" {
			logPath = getDefaultLogPath()
		}
		f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			f, _ = os.Create(logPath)
		}
		if f != nil {
			output = io.MultiWriter(f, os.Stdout)
		} else {
			output = os.Stdout
		}
	default: // "file" or default
		logPath := getDefaultLogPath()
		logFilePath = logPath
		f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			f, _ = os.Create(logPath)
		}
		if f != nil {
			output = f
		} else {
			output = os.Stdout
		}
	}

	originalOutput = os.Stdout

	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level: parseLevel(cfg.Level),
	}
	originalLevel = parseLevel(cfg.Level)

	switch strings.ToLower(cfg.Format) {
	case "json":
		handler = slog.NewJSONHandler(output, opts)
	default:
		handler = slog.NewTextHandler(output, opts)
	}

	logger = slog.New(handler)
	slog.SetDefault(logger)
}

// Silence redirects logger output to io.Discard
func Silence() {
	if logger != nil {
		handler := slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{})
		logger = slog.New(handler)
		slog.SetDefault(logger)
	}
}

// Restore restores logger output to original stdout
func Restore() {
	if logger != nil && originalOutput != nil {
		handler := slog.NewTextHandler(originalOutput, &slog.HandlerOptions{
			Level: originalLevel,
		})
		logger = slog.New(handler)
		slog.SetDefault(logger)
	}
}

// parseLevel converts a level string to slog.Level
func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// Debug logs a debug message
func Debug(msg string, args ...any) {
	if logger != nil {
		logger.Debug(msg, args...)
	}
}

// Info logs an info message
func Info(msg string, args ...any) {
	if logger != nil {
		logger.Info(msg, args...)
	}
}

// Warn logs a warning message
func Warn(msg string, args ...any) {
	if logger != nil {
		logger.Warn(msg, args...)
	}
}

// Error logs an error message
func Error(msg string, args ...any) {
	if logger != nil {
		logger.Error(msg, args...)
	}
}

// Fatal logs a fatal error and exits
func Fatal(msg string, args ...any) {
	if logger != nil {
		logger.Error(msg, args...)
	}
	os.Exit(1)
}

// WithContext returns a logger with context fields
// Note: Go slog doesn't natively propagate context, so we just return the logger
// In production, you would use ctx.Value() to add context-specific fields
func WithContext(ctx context.Context) *slog.Logger {
	return logger
}

// WithFields returns a logger with additional fields
func WithFields(fields ...any) *slog.Logger {
	return logger.With(fields...)
}