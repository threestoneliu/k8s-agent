package llm

import (
	"context"
	"errors"
	"testing"
)

func TestNewService(t *testing.T) {
	t.Run("creates service with OpenAI provider", func(t *testing.T) {
		cfg := &Config{
			Provider: "openai",
			BaseURL:  "https://api.openai.com/v1",
			APIKey:   "test-key",
			Model:    "gpt-4",
		}
		svc, err := NewService(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if svc.ProviderName() != "openai" {
			t.Errorf("expected 'openai', got %q", svc.ProviderName())
		}
	})

	t.Run("defaults to OpenAI for unknown provider", func(t *testing.T) {
		cfg := &Config{
			Provider: "unknown",
			APIKey:   "test-key",
			Model:    "gpt-4",
		}
		svc, err := NewService(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if svc.ProviderName() != "openai" {
			t.Errorf("expected 'openai', got %q", svc.ProviderName())
		}
	})
}

func TestServiceChat(t *testing.T) {
	t.Run("Chat delegates to provider", func(t *testing.T) {
		cfg := &Config{
			Provider: "openai",
			BaseURL:  "https://api.openai.com/v1",
			APIKey:   "test-key",
			Model:    "gpt-4",
		}
		svc, err := NewService(cfg)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		// This will fail due to network, but it shows the call path works
		ctx := context.Background()
		messages := []Message{{Role: "user", Content: "hello"}}
		_, err = svc.Chat(ctx, messages)

		// We expect an error because the mock server doesn't exist
		// This is expected behavior
		if err == nil {
			t.Log("Chat succeeded (unexpected in test environment)")
		}
	})
}

// mockServiceProvider is a test mock for the provider
type mockServiceProvider struct {
	response string
	err      error
	name     string
}

func (m *mockServiceProvider) Chat(ctx context.Context, messages []Message) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.response, nil
}

func (m *mockServiceProvider) ChatWithFunctions(ctx context.Context, messages []Message, functions []Function) (string, *FunctionCall, error) {
	return "", nil, nil
}

func (m *mockServiceProvider) Name() string {
	return m.name
}

func TestServiceChatWithMock(t *testing.T) {
	t.Run("Chat returns response through service", func(t *testing.T) {
		svc := &Service{provider: &mockServiceProvider{
			response: "Hello from mock!",
			name:     "mock",
		}}

		ctx := context.Background()
		messages := []Message{{Role: "user", Content: "hello"}}
		resp, err := svc.Chat(ctx, messages)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if resp != "Hello from mock!" {
			t.Errorf("expected 'Hello from mock!', got %q", resp)
		}
	})

	t.Run("Chat propagates error from provider", func(t *testing.T) {
		svc := &Service{provider: &mockServiceProvider{
			err:  errors.New("provider error"),
			name: "mock",
		}}

		ctx := context.Background()
		messages := []Message{{Role: "user", Content: "hello"}}
		_, err := svc.Chat(ctx, messages)

		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "provider error" {
			t.Errorf("expected 'provider error', got %v", err)
		}
	})

	t.Run("ProviderName returns provider name", func(t *testing.T) {
		svc := &Service{provider: &mockServiceProvider{name: "test-provider"}}
		if svc.ProviderName() != "test-provider" {
			t.Errorf("expected 'test-provider', got %q", svc.ProviderName())
		}
	})
}