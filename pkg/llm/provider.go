package llm

import "context"

// Message represents a chat message.
type Message struct {
	Role       string
	Content    string
	ToolCallID string
	// For assistant messages with tool_calls
	ToolCalls []struct {
		ID        string
		Name      string
		Arguments string
	}
}

// Provider defines the interface for LLM providers.
type Provider interface {
	// Chat sends messages to the LLM and returns the response.
	Chat(ctx context.Context, messages []Message) (string, error)
	// ChatWithFunctions sends messages to the LLM with function definitions
	// and returns a function call and/or text response.
	ChatWithFunctions(ctx context.Context, messages []Message, functions []Function) (string, *FunctionCall, error)
	// Name returns the provider name.
	Name() string
}