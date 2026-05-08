package llm

import (
	"context"
	"errors"
	"testing"
)

func TestProviderInterface(t *testing.T) {
	// Test that Provider interface is satisfied
	var _ Provider = &mockProvider{}

	t.Run("Chat returns response", func(t *testing.T) {
		p := &mockProvider{response: "Hello, World!"}
		ctx := context.Background()
		messages := []Message{{Role: "user", Content: "hello"}}

		resp, err := p.Chat(ctx, messages)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if resp != "Hello, World!" {
			t.Errorf("expected 'Hello, World!', got %q", resp)
		}
	})

	t.Run("Chat returns error", func(t *testing.T) {
		p := &mockProvider{err: errors.New("API error")}
		ctx := context.Background()
		messages := []Message{{Role: "user", Content: "hello"}}

		_, err := p.Chat(ctx, messages)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if err.Error() != "API error" {
			t.Errorf("expected 'API error', got %v", err)
		}
	})

	t.Run("Name returns provider name", func(t *testing.T) {
		p := &mockProvider{name: "test-provider"}
		if p.Name() != "test-provider" {
			t.Errorf("expected 'test-provider', got %q", p.Name())
		}
	})
}

type mockProvider struct {
	name     string
	response string
	err      error
}

func (m *mockProvider) Chat(ctx context.Context, messages []Message) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.response, nil
}

func (m *mockProvider) ChatWithFunctions(ctx context.Context, messages []Message, functions []Function) (string, *FunctionCall, error) {
	return "", nil, nil
}

func (m *mockProvider) Name() string {
	return m.name
}