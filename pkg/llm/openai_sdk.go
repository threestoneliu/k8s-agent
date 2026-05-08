package llm

import (
	"context"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared"
	"github.com/openai/openai-go/shared/constant"
)

// OpenAISDKProvider implements Provider using the official OpenAI SDK
type OpenAISDKProvider struct {
	client    openai.Client
	model     string
	maxTokens int
}

// NewOpenAISDKProvider creates a new OpenAI SDK provider
func NewOpenAISDKProvider(cfg *Config) *OpenAISDKProvider {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	maxTokens := cfg.MaxTokens
	if maxTokens <= 0 {
		maxTokens = 2048
	}

	client := openai.NewClient(
		option.WithAPIKey(cfg.APIKey),
		option.WithBaseURL(baseURL),
	)

	return &OpenAISDKProvider{
		client:    client,
		model:     cfg.Model,
		maxTokens: maxTokens,
	}
}

// Chat sends messages to OpenAI API and returns the response
func (p *OpenAISDKProvider) Chat(ctx context.Context, messages []Message) (string, error) {
	openAIMessages := toSDKMessages(messages)

	resp, err := p.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:     p.model,
		Messages:  openAIMessages,
		MaxTokens: openai.Int(int64(p.maxTokens)),
	})
	if err != nil {
		return "", fmt.Errorf("OpenAI API call failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned")
	}

	content := resp.Choices[0].Message.Content
	return content, nil
}

// ChatWithFunctions sends messages with function definitions and returns a function call and/or text response
func (p *OpenAISDKProvider) ChatWithFunctions(ctx context.Context, messages []Message, functions []Function) (string, *FunctionCall, error) {
	openAIMessages := toSDKMessages(messages)

	// Convert our Function type to SDK format using Tools
	sdkTools := make([]openai.ChatCompletionToolParam, len(functions))
	for i, fn := range functions {
		sdkTools[i] = openai.ChatCompletionToolParam{
			Type: constant.ValueOf[constant.Function](),
			Function: shared.FunctionDefinitionParam{
				Name:        fn.Name,
				Description: openai.String(fn.Description),
				Parameters:  shared.FunctionParameters(fn.Parameters),
			},
		}
	}

	resp, err := p.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:     p.model,
		Messages:  openAIMessages,
		Tools:     sdkTools,
		MaxTokens: openai.Int(int64(p.maxTokens)),
	})
	if err != nil {
		return "", nil, fmt.Errorf("OpenAI API call failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", nil, fmt.Errorf("no response choices returned")
	}

	// Check for function call in response (via ToolCalls)
	choice := resp.Choices[0]
	if len(choice.Message.ToolCalls) > 0 {
		toolCall := choice.Message.ToolCalls[0]
		// Return both text content (if any) and function call
		textContent := choice.Message.Content
		return textContent, &FunctionCall{
			ID:        toolCall.ID,
			Name:      toolCall.Function.Name,
			Arguments: toolCall.Function.Arguments,
		}, nil
	}

	// Return text content when no function call
	return choice.Message.Content, nil, nil
}

// Name returns the provider name
func (p *OpenAISDKProvider) Name() string {
	return "openai"
}

// toSDKMessages converts our Message type to SDK format
func toSDKMessages(messages []Message) []openai.ChatCompletionMessageParamUnion {
	result := make([]openai.ChatCompletionMessageParamUnion, len(messages))
	for i, msg := range messages {
		switch msg.Role {
		case "system":
			result[i] = openai.ChatCompletionMessageParamUnion{
				OfSystem: &openai.ChatCompletionSystemMessageParam{
					Content: openai.ChatCompletionSystemMessageParamContentUnion{
						OfString: openai.String(msg.Content),
					},
				},
			}
		case "user":
			result[i] = openai.ChatCompletionMessageParamUnion{
				OfUser: &openai.ChatCompletionUserMessageParam{
					Content: openai.ChatCompletionUserMessageParamContentUnion{
						OfString: openai.String(msg.Content),
					},
				},
			}
		case "assistant":
			assistantMsg := &openai.ChatCompletionAssistantMessageParam{
				Content: openai.ChatCompletionAssistantMessageParamContentUnion{
					OfString: openai.String(msg.Content),
				},
			}
			// If this message has ToolCalls, include them
			if len(msg.ToolCalls) > 0 {
				toolCalls := make([]openai.ChatCompletionMessageToolCallParam, len(msg.ToolCalls))
				for j, tc := range msg.ToolCalls {
					toolCalls[j] = openai.ChatCompletionMessageToolCallParam{
						ID:   tc.ID,
						Type: constant.ValueOf[constant.Function](),
						Function: openai.ChatCompletionMessageToolCallFunctionParam{
							Name:      tc.Name,
							Arguments: tc.Arguments,
						},
					}
				}
				assistantMsg.ToolCalls = toolCalls
			}
			result[i] = openai.ChatCompletionMessageParamUnion{
				OfAssistant: assistantMsg,
			}
		case "tool":
			result[i] = openai.ChatCompletionMessageParamUnion{
				OfTool: &openai.ChatCompletionToolMessageParam{
					Content: openai.ChatCompletionToolMessageParamContentUnion{
						OfString: openai.String(msg.Content),
					},
					Role:       constant.ValueOf[constant.Tool](),
					ToolCallID: msg.ToolCallID,
				},
			}
		default:
			result[i] = openai.ChatCompletionMessageParamUnion{
				OfUser: &openai.ChatCompletionUserMessageParam{
					Content: openai.ChatCompletionUserMessageParamContentUnion{
						OfString: openai.String(msg.Content),
					},
				},
			}
		}
	}
	return result
}