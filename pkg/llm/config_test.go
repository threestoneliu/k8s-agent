package llm

import (
	"os"
	"testing"
)

func TestNewConfig(t *testing.T) {
	t.Run("reads from environment variables", func(t *testing.T) {
		os.Setenv("LLM_PROVIDER", "openai")
		os.Setenv("OPENAI_API_KEY", "test-key")
		os.Setenv("OPENAI_MODEL", "gpt-4")
		os.Setenv("OPENAI_BASE_URL", "https://api.openai.com/v1")
		defer func() {
			os.Unsetenv("LLM_PROVIDER")
			os.Unsetenv("OPENAI_API_KEY")
			os.Unsetenv("OPENAI_MODEL")
			os.Unsetenv("OPENAI_BASE_URL")
		}()

		cfg := NewConfig()

		if cfg.Provider != "openai" {
			t.Errorf("expected provider 'openai', got %q", cfg.Provider)
		}
		if cfg.APIKey != "test-key" {
			t.Errorf("expected API key 'test-key', got %q", cfg.APIKey)
		}
		if cfg.Model != "gpt-4" {
			t.Errorf("expected model 'gpt-4', got %q", cfg.Model)
		}
		if cfg.BaseURL != "https://api.openai.com/v1" {
			t.Errorf("expected base URL 'https://api.openai.com/v1', got %q", cfg.BaseURL)
		}
	})

	t.Run("defaults to openai when no env vars set", func(t *testing.T) {
		os.Unsetenv("LLM_PROVIDER")
		os.Unsetenv("OPENAI_API_KEY")
		os.Unsetenv("OPENAI_MODEL")
		os.Unsetenv("OPENAI_BASE_URL")
		os.Unsetenv("ANTHROPIC_API_KEY")
		os.Unsetenv("ANTHROPIC_MODEL")
		os.Unsetenv("LLM_API_KEY")
		os.Unsetenv("LLM_MODEL")
		os.Unsetenv("LLM_BASE_URL")
		os.Unsetenv("LLM_TIMEOUT")
		os.Unsetenv("LLM_TEMPERATURE")
		os.Unsetenv("LLM_MAX_TOKENS")

		cfg := NewConfig()

		if cfg.Provider != "openai" {
			t.Errorf("expected provider 'openai', got %q", cfg.Provider)
		}
		if cfg.APIKey != "" {
			t.Errorf("expected empty API key, got %q", cfg.APIKey)
		}
		if cfg.Model != "" {
			t.Errorf("expected empty model, got %q", cfg.Model)
		}
		if cfg.BaseURL != "" {
			t.Errorf("expected empty base URL, got %q", cfg.BaseURL)
		}
		if cfg.Timeout != 30.0 {
			t.Errorf("expected timeout 30.0, got %f", cfg.Timeout)
		}
		if cfg.Temperature != 0.7 {
			t.Errorf("expected temperature 0.7, got %f", cfg.Temperature)
		}
		if cfg.MaxTokens != 4000 {
			t.Errorf("expected max tokens 4000, got %d", cfg.MaxTokens)
		}
	})
}