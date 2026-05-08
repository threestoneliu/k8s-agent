package llm

import (
	"context"
)

// Service provides LLM operations.
type Service struct {
	provider Provider
}

// NewService creates a new LLM service.
func NewService(cfg *Config) (*Service, error) {
	provider, err := newProvider(cfg)
	if err != nil {
		return nil, err
	}
	return &Service{provider: provider}, nil
}

// Chat sends messages to the LLM and returns the response.
func (s *Service) Chat(ctx context.Context, messages []Message) (string, error) {
	return s.provider.Chat(ctx, messages)
}

// ChatWithFunctions sends messages to the LLM with function definitions
// and returns a function call and/or text response.
func (s *Service) ChatWithFunctions(ctx context.Context, messages []Message, functions []Function) (string, *FunctionCall, error) {
	return s.provider.ChatWithFunctions(ctx, messages, functions)
}

// ProviderName returns the name of the active provider.
func (s *Service) ProviderName() string {
	return s.provider.Name()
}

func newProvider(cfg *Config) (Provider, error) {
	// Default to OpenAI SDK
	return NewOpenAISDKProvider(cfg), nil
}