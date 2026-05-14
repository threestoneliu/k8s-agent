package llm

import (
	"context"

	sharedutil "github.com/threestoneliu/k8s-agent/pkg/shared"
)

// Service provides LLM operations with direct OpenAI calls
type Service struct {
	client          *OpenAISDKProvider
	functions       []sharedutil.Function
	responseParser  ResponseParser
}

// NewService creates a new LLM service
func NewService(cfg *LLMConfig) *Service {
	provider := NewOpenAISDKProvider(cfg)
	return &Service{
		client:          provider,
		functions:       getFunctions(),
		responseParser:   &OpenAIResponseParser{},
	}
}

// Chat sends messages to the LLM and returns the response
func (s *Service) Chat(ctx context.Context, messages []sharedutil.Message) (string, error) {
	return s.client.Chat(ctx, messages)
}

// ChatWithFunctions sends messages with function definitions
func (s *Service) ChatWithFunctions(ctx context.Context, messages []sharedutil.Message, functions []sharedutil.Function) (string, *sharedutil.FunctionCall, error) {
	return s.client.ChatWithFunctions(ctx, messages, functions)
}

// GetFunctions returns registered function definitions
func (s *Service) GetFunctions() []sharedutil.Function {
	return s.functions
}

// ResponseParser returns the response parser for this service
func (s *Service) ResponseParser() ResponseParser {
	return s.responseParser
}
