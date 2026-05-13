package cluster

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// AppConfig represents the unified config.yaml structure
type AppConfig struct {
	LLM            LLMConfig     `yaml:"llm"`
	Context        ContextConfig `yaml:"context"`
	Session        SessionConfig `yaml:"session"`
	CurrentCluster string        `yaml:"current-cluster"`
	Log            LogConfig     `yaml:"logging"`
}

// SessionConfig holds session storage configuration
type SessionConfig struct {
	StoragePath  string `yaml:"storage-path"`  // directory for session files
	MaxCacheSize int    `yaml:"max-cache-size"` // max sessions to keep in memory cache
	MaxSessions  int    `yaml:"max-sessions"`  // max session files to retain on disk
}

// LogConfig holds logging configuration
type LogConfig struct {
	Level  string `yaml:"level"`  // debug, info, warn, error
	Format string `yaml:"format"` // json, text
}

// LLMConfig holds LLM provider settings
type LLMConfig struct {
	Provider    string  `yaml:"provider"`     // "openai", "anthropic", "ollama"
	APIKey      string  `yaml:"api-key"`
	Model       string  `yaml:"model"`
	BaseURL     string  `yaml:"base-url"`
	Timeout     float64 `yaml:"timeout"`      // seconds
	Temperature float64 `yaml:"temperature"`
	MaxTokens   int     `yaml:"max-tokens"`
}

// ContextConfig holds context management settings
type ContextConfig struct {
	MaxMessages       int  `yaml:"max-messages"`        // max messages to keep in context
	MaxTokens         int  `yaml:"max-tokens"`          // max tokens in context window
	SummaryEnabled    bool `yaml:"summary-enabled"`    // enable conversation summarization
	SummaryThreshold  int  `yaml:"summary-threshold"`   // trigger summarization when message count reaches this
	ToolCallRetention int  `yaml:"tool-call-retention"`  // number of recent tool calls to retain in level 1 compression
	MaxIterations     int  `yaml:"max-iterations"`      // max function call iterations per turn
}

// DefaultAppConfig returns the default configuration
func DefaultAppConfig() *AppConfig {
	return &AppConfig{
		LLM: LLMConfig{
			Provider:    "",
			APIKey:      "",
			Model:       "",
			BaseURL:     "",
			Timeout:     30.0,
			Temperature: 0.7,
			MaxTokens:   4000,
		},
		Context: ContextConfig{
			MaxMessages:       20,
			MaxTokens:         8000,
			SummaryEnabled:    true,
			SummaryThreshold:  10,
			ToolCallRetention: 10,
			MaxIterations:     10,
		},
		Session: SessionConfig{
			StoragePath:  "~/.config/k8s-agent/sessions",
			MaxCacheSize: 100,
			MaxSessions:  10,
		},
		CurrentCluster: "",
		Log: LogConfig{
			Level:  "info",
			Format: "text",
		},
	}
}

// LoadAppConfig loads configuration from file, with env vars as fallback only
// when config file values are empty. Config file has higher priority.
func LoadAppConfig(path string) (*AppConfig, error) {
	cfg := DefaultAppConfig()

	// Try to load from file first (config file has higher priority)
	if data, err := os.ReadFile(path); err == nil {
		if err := yaml.Unmarshal(data, cfg); err != nil {
			return nil, fmt.Errorf("failed to parse config: %w", err)
		}
	}

	// Apply env vars only when config file value is empty (fallback)
	cfg.LLM.applyEnvFallback()
	cfg.Context.applyEnvFallback()

	return cfg, nil
}

// applyEnvFallback applies environment variable only when the config value is empty
func (c *LLMConfig) applyEnvFallback() {
	if c.Provider == "" {
		if v := os.Getenv("LLM_PROVIDER"); v != "" {
			c.Provider = v
		}
	}
	if c.APIKey == "" {
		if v := os.Getenv("OPENAI_API_KEY"); v != "" {
			c.APIKey = v
		} else if v := os.Getenv("ANTHROPIC_API_KEY"); v != "" {
			c.APIKey = v
		} else if v := os.Getenv("LLM_API_KEY"); v != "" {
			c.APIKey = v
		}
	}
	if c.Model == "" {
		if v := os.Getenv("OPENAI_MODEL"); v != "" {
			c.Model = v
		} else if v := os.Getenv("ANTHROPIC_MODEL"); v != "" {
			c.Model = v
		} else if v := os.Getenv("LLM_MODEL"); v != "" {
			c.Model = v
		}
	}
	if c.BaseURL == "" {
		if v := os.Getenv("OPENAI_BASE_URL"); v != "" {
			c.BaseURL = v
		} else if v := os.Getenv("LLM_BASE_URL"); v != "" {
			c.BaseURL = v
		}
	}
	if c.Timeout == 0 {
		if v := os.Getenv("LLM_TIMEOUT"); v != "" {
			if parsed, err := parseFloatEnv(v); err == nil && parsed > 0 {
				c.Timeout = parsed
			}
		}
	}
	if c.Temperature == 0 {
		if v := os.Getenv("LLM_TEMPERATURE"); v != "" {
			if parsed, err := parseFloatEnv(v); err == nil && parsed >= 0 && parsed <= 2 {
				c.Temperature = parsed
			}
		}
	}
	if c.MaxTokens == 0 {
		if v := os.Getenv("LLM_MAX_TOKENS"); v != "" {
			if parsed, err := parseIntEnv(v); err == nil && parsed > 0 {
				c.MaxTokens = parsed
			}
		}
	}
}

// applyEnvFallback applies environment variable only when the config value is empty
func (c *ContextConfig) applyEnvFallback() {
	if c.MaxMessages == 0 {
		if v := os.Getenv("CONTEXT_MAX_MESSAGES"); v != "" {
			if parsed, err := parseIntEnv(v); err == nil && parsed > 0 {
				c.MaxMessages = parsed
			}
		}
	}
	if c.MaxTokens == 0 {
		if v := os.Getenv("CONTEXT_MAX_TOKENS"); v != "" {
			if parsed, err := parseIntEnv(v); err == nil && parsed > 0 {
				c.MaxTokens = parsed
			}
		}
	}
	if !c.SummaryEnabled {
		if v := os.Getenv("CONTEXT_SUMMARY_ENABLED"); v != "" {
			c.SummaryEnabled = v == "true" || v == "1"
		}
	}
}

func parseFloatEnv(s string) (float64, error) {
	var n float64
	var fraction float64
	var divisor float64
	decimal := false

	for _, c := range s {
		if c == '.' {
			decimal = true
			divisor = 1
			continue
		}
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("invalid character: %c", c)
		}
		if decimal {
			divisor *= 10
			fraction = fraction*10 + float64(c-'0')
		} else {
			n = n*10 + float64(c-'0')
		}
	}
	if decimal {
		n += fraction / divisor
	}
	return n, nil
}

func parseIntEnv(s string) (int, error) {
	var n int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("invalid character: %c", c)
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}

// Save saves the configuration to file
func (c *AppConfig) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}