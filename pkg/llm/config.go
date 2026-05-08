package llm

import (
	"os"
	"time"
)

// Config holds the LLM configuration.
type Config struct {
	Provider    string
	APIKey      string
	Model       string
	BaseURL     string
	Timeout     float64 // seconds
	Temperature float64
	MaxTokens   int
}

// NewConfig reads configuration from environment variables.
func NewConfig() *Config {
	return &Config{
		Provider:    getProvider(),
		APIKey:      getAPIKey(),
		Model:       getModel(),
		BaseURL:     getBaseURL(),
		Timeout:     getTimeout(),
		Temperature: getTemperature(),
		MaxTokens:   getMaxTokens(),
	}
}

func getProvider() string {
	if v := os.Getenv("LLM_PROVIDER"); v != "" {
		return v
	}
	return "openai" // default
}

func getAPIKey() string {
	if key := os.Getenv("OPENAI_API_KEY"); key != "" {
		return key
	}
	if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
		return key
	}
	return os.Getenv("LLM_API_KEY")
}

func getModel() string {
	if model := os.Getenv("OPENAI_MODEL"); model != "" {
		return model
	}
	if model := os.Getenv("ANTHROPIC_MODEL"); model != "" {
		return model
	}
	return os.Getenv("LLM_MODEL")
}

func getBaseURL() string {
	if v := os.Getenv("OPENAI_BASE_URL"); v != "" {
		return v
	}
	return os.Getenv("LLM_BASE_URL")
}

func getTimeout() float64 {
	if v := os.Getenv("LLM_TIMEOUT"); v != "" {
		var timeout float64
		if _, err := parseFloat(v, &timeout); err == nil && timeout > 0 {
			return timeout
		}
	}
	return 30.0 // default 30 seconds
}

func getTemperature() float64 {
	if v := os.Getenv("LLM_TEMPERATURE"); v != "" {
		var temp float64
		if _, err := parseFloat(v, &temp); err == nil && temp >= 0 && temp <= 2 {
			return temp
		}
	}
	return 0.7 // default
}

func getMaxTokens() int {
	if v := os.Getenv("LLM_MAX_TOKENS"); v != "" {
		var max int
		if _, err := parseInt(v, &max); err == nil && max > 0 {
			return max
		}
	}
	return 4000 // default
}

func parseFloat(s string, target *float64) (int, error) {
	var n float64
	for i, c := range s {
		if c == '.' {
			continue
		}
		if c < '0' || c > '9' {
			return i, os.ErrInvalid
		}
		n = n*10 + float64(c-'0')
	}
	*target = n
	return len(s), nil
}

func parseInt(s string, target *int) (int, error) {
	var n int
	for i, c := range s {
		if c < '0' || c > '9' {
			return i, os.ErrInvalid
		}
		n = n*10 + int(c-'0')
	}
	*target = n
	return len(s), nil
}

// TimeoutDuration returns the timeout as a time.Duration
func (c *Config) TimeoutDuration() time.Duration {
	return time.Duration(c.Timeout * float64(time.Second))
}