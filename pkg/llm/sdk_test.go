package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOpenAISDK_ToolCalling(t *testing.T) {
	// Create a test server that returns a tool call response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Log what we received
		var reqBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err == nil {
			t.Logf("Received request: tools=%v", reqBody["tools"])
		}

		// Return a tool call response (OpenAI format)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id": "chatcmpl-123",
			"object": "chat.completion",
			"created": 1234567890,
			"model": "gpt-4o",
			"choices": []map[string]interface{}{
				{
					"index": 0,
					"message": map[string]interface{}{
						"role": "assistant",
						"content": nil,
						"tool_calls": []map[string]interface{}{
							{
								"id": "call_123",
								"type": "function",
								"function": map[string]interface{}{
									"name": "list_pods",
									"arguments": `{"namespace":"default"}`,
								},
							},
						},
					},
					"finish_reason": "tool_calls",
				},
			},
		})
	}))
	defer server.Close()

	cfg := &Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
		Model:   "gpt-4o",
	}
	provider := NewOpenAISDKProvider(cfg)

	messages := []Message{
		{Role: "user", Content: "显示default命名空间的pod"},
	}
	functions := []Function{
		{
			Name:        "list_pods",
			Description: "List pods in a namespace",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"namespace": map[string]interface{}{
						"type":    "string",
						"default": "default",
					},
				},
			},
		},
	}

	_, fnCall, err := provider.ChatWithFunctions(context.Background(), messages, functions)
	if err != nil {
		t.Fatalf("ChatWithFunctions failed: %v", err)
	}

	if fnCall == nil {
		t.Fatal("Expected function call, got nil")
	}

	if fnCall.Name != "list_pods" {
		t.Errorf("Expected function name 'list_pods', got '%s'", fnCall.Name)
	}

	if fnCall.Arguments != `{"namespace":"default"}` {
		t.Errorf("Expected arguments '{\"namespace\":\"default\"}', got '%s'", fnCall.Arguments)
	}

	t.Logf("Successfully received function call: %s(%s)", fnCall.Name, fnCall.Arguments)
}

func TestOpenAISDK_TextResponse(t *testing.T) {
	// Create a test server that returns text (no tool call)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id": "chatcmpl-123",
			"object": "chat.completion",
			"created": 1234567890,
			"model": "gpt-4o",
			"choices": []map[string]interface{}{
				{
					"index": 0,
					"message": map[string]interface{}{
						"role": "assistant",
						"content": "I'll list the pods for you now.",
					},
					"finish_reason": "stop",
				},
			},
		})
	}))
	defer server.Close()

	cfg := &Config{
		BaseURL: server.URL,
		APIKey:  "test-key",
		Model:   "gpt-4o",
	}
	provider := NewOpenAISDKProvider(cfg)

	messages := []Message{
		{Role: "user", Content: "帮我查看一下集群里有哪些pod"},
	}
	functions := GetFunctions()

	_, fnCall, err := provider.ChatWithFunctions(context.Background(), messages, functions)
	if err != nil {
		t.Fatalf("ChatWithFunctions failed: %v", err)
	}

	// Text response means no function call
	if fnCall != nil {
		t.Error("Expected nil function call for text response")
	}

	t.Log("Text response handled correctly (no tool call)")
}