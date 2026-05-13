package llm

import (
	"testing"
)

func TestNewService(t *testing.T) {
	t.Run("creates service with config", func(t *testing.T) {
		cfg := &LLMConfig{
			BaseURL:  "https://api.openai.com/v1",
			APIKey:   "test-key",
			Model:    "gpt-4",
			MaxTokens: 2048,
		}
		svc := NewService(cfg)
		if svc == nil {
			t.Fatal("expected non-nil service")
		}
	})
}

func TestLLMConfig_Fields(t *testing.T) {
	cfg := &LLMConfig{
		APIKey:   "test-key",
		Model:    "gpt-4",
		BaseURL:  "https://api.openai.com/v1",
		MaxTokens: 4096,
	}

	if cfg.APIKey != "test-key" {
		t.Errorf("expected APIKey 'test-key', got %q", cfg.APIKey)
	}
	if cfg.Model != "gpt-4" {
		t.Errorf("expected Model 'gpt-4', got %q", cfg.Model)
	}
	if cfg.MaxTokens != 4096 {
		t.Errorf("expected MaxTokens 4096, got %d", cfg.MaxTokens)
	}
}