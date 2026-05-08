package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"k8s-agent/pkg/engine"
	"k8s-agent/pkg/llm"
	"k8s-agent/pkg/log"
	"k8s-agent/pkg/session"
	"k8s-agent/pkg/ui"
)

// Agent handles the conversation flow between LLM and UI
// It receives user input, builds prompts, executes tool calls, and streams progress to UI
type Agent struct {
	llmSvc      *llm.Service
	fnExec      *llm.Executor
	confirmMgr  interface {
		CreateConfirmation(clusterName string, op *engine.ClassifiedOperation) (string, error)
	}
	store      session.StoreInterface
	ctxManager *session.ContextManager
	sessionID  string
	clusterName string
	messages   []*session.Message
}

// NewAgent creates a new Agent instance
func NewAgent(llmSvc *llm.Service, fnExec *llm.Executor, confirmMgr interface {
	CreateConfirmation(clusterName string, op *engine.ClassifiedOperation) (string, error)
}, store session.StoreInterface, sessionID, clusterName string, ctxManager *session.ContextManager) *Agent {
	var sessionConv *session.Conversation
	if store != nil {
		var err error
		sessionConv, err = store.GetConversation(sessionID)
		if err != nil {
			sessionConv, _ = store.CreateConversation(sessionID, clusterName, "")
		}
		if sessionConv == nil {
			sessionConv, _ = store.CreateConversation(sessionID, clusterName, "")
		}
	}

	return &Agent{
		llmSvc:      llmSvc,
		fnExec:      fnExec,
		confirmMgr:  confirmMgr,
		store:       store,
		ctxManager:  ctxManager,
		sessionID:   sessionID,
		clusterName: clusterName,
		messages:    make([]*session.Message, 0),
	}
}

// ProcessInput processes user input and streams progress to UI
func (a *Agent) ProcessInput(ctx context.Context, input string, uiInterface ui.UI) error {
	if uiInterface == nil {
		return fmt.Errorf("ui interface is required")
	}

	// Add user message to session
	a.addMessageToSession(session.RoleUser, input, nil)
	a.messages = append(a.messages, session.NewMessage(session.RoleUser, input, nil))

	// Build state for processing
	state := NewState(a.clusterName, a.messages)

	// Process with streaming progress
	if err := a.processWithProgress(ctx, state, uiInterface); err != nil {
		return fmt.Errorf("chat processing failed: %w", err)
	}

	return nil
}

// processWithProgress handles LLM function calling with progress callbacks to UI
func (a *Agent) processWithProgress(ctx context.Context, state *State, uiInterface ui.UI) error {
	systemPrompt := BuildSystemPrompt(state.ClusterName)

	var messages []llm.Message
	if a.ctxManager != nil {
		messages = a.ctxManager.BuildContextMessages(systemPrompt, state.SessionMessages, "")
	} else {
		messages = []llm.Message{{Role: "system", Content: systemPrompt}}
		for _, m := range state.SessionMessages {
			messages = append(messages, llm.Message{
				Role:    string(m.Role),
				Content: m.Content,
			})
		}
	}

	log.Info("agent process started", "cluster", state.ClusterName, "messageCount", len(messages))

	maxIterations := 10

	for i := range maxIterations {
		log.Info("LLM iteration", "iteration", i, "messageCount", len(messages))

		textResp, fnCall, err := a.llmSvc.ChatWithFunctions(ctx, messages, llm.K8sFunctions)
		if err != nil {
			log.Error("LLM call failed", "error", err)
			return fmt.Errorf("LLM call failed: %w", err)
		}

		log.Info("LLM response received", "fnCall", fnCall, "textResp", textResp)

		// Send text response if any
		if textResp != "" {
			uiInterface.SendProgress(ui.Progress{
				Type:      "text",
				Content:   textResp,
				Timestamp: time.Now(),
			})
		}

		// If no function call, we're done
		if fnCall == nil {
			// Final response
			uiInterface.SendMessage(&session.Message{
				Role:      session.RoleAssistant,
				Content:   textResp,
				Timestamp: time.Now(),
			})
			uiInterface.Done()
			return nil
		}

		// Report tool call start
		uiInterface.SendProgress(ui.Progress{
			Type:      "tool_call_start",
			ToolName:  fnCall.Name,
			ToolArgs:  fnCall.Arguments,
			Timestamp: time.Now(),
		})

		// Execute function call
		log.Info("Executing function call", "name", fnCall.Name, "args", fnCall.Arguments)
		result := a.fnExec.ExecuteFunctionCall(fnCall, state.ClusterName)
		log.Info("Function execution result", "success", result.Success, "hasClusterSwitch", result.ClusterSwitch != "")

		// Report tool result
		uiInterface.SendProgress(ui.Progress{
			Type:        "tool_result",
			ToolName:    fnCall.Name,
			ToolArgs:    fnCall.Arguments,
			ToolResult:  result.Result,
			ToolSuccess: result.Success,
			Timestamp:   time.Now(),
		})

		// Handle cluster switch
		if result.ClusterSwitch != "" {
			state.ClusterName = result.ClusterSwitch
			uiInterface.SetClusterName(result.ClusterSwitch)
			systemPrompt = BuildSystemPrompt(state.ClusterName)
			if a.ctxManager != nil {
				messages = a.ctxManager.BuildContextMessages(systemPrompt, state.SessionMessages, "")
			} else {
				messages = []llm.Message{{Role: "system", Content: systemPrompt}}
				for j := 1; j < len(state.SessionMessages); j++ {
					messages = append(messages, llm.Message{
						Role:    string(state.SessionMessages[j].Role),
						Content: state.SessionMessages[j].Content,
					})
				}
			}
			fnCallMsg := llm.Message{
				Role:    "assistant",
				Content: fmt.Sprintf("[Function Call: %s]\nArguments: %s", fnCall.Name, fnCall.Arguments),
				ToolCalls: []struct {
					ID        string
					Name      string
					Arguments string
				}{{ID: fnCall.ID, Name: fnCall.Name, Arguments: fnCall.Arguments}},
			}
			messages = append(messages, fnCallMsg)
			messages = append(messages, llm.Message{
				Role:       "tool",
				Content:    result.Result,
				ToolCallID: fnCall.ID,
			})
			continue
		}

		// Add function call message
		fnCallMsg := llm.Message{
			Role:    "assistant",
			Content: fmt.Sprintf("[Function Call: %s]\nArguments: %s", fnCall.Name, fnCall.Arguments),
			ToolCalls: []struct {
				ID        string
				Name      string
				Arguments string
			}{{ID: fnCall.ID, Name: fnCall.Name, Arguments: fnCall.Arguments}},
		}
		messages = append(messages, fnCallMsg)

		// Add function result as tool message
		var toolContent string
		if result.Success {
			if strings.HasPrefix(result.Result, "confirmation_required:") {
				confirmKey, err := a.confirmMgr.CreateConfirmation(state.ClusterName, result.Operation)
				if err != nil {
					toolContent = fmt.Sprintf("Error creating confirmation: %s", err.Error())
				} else {
					toolContent = fmt.Sprintf("Confirmation required. Use 'k8s-agent confirm %s' to approve.", confirmKey)
				}
			} else {
				toolContent = result.Result
			}
		} else {
			toolContent = fmt.Sprintf("Error: %s", result.Error)
		}
		messages = append(messages, llm.Message{
			Role:       "tool",
			Content:    toolContent,
			ToolCallID: fnCall.ID,
		})

		// Check if we should re-apply context compression
		if a.ctxManager != nil && len(messages) > a.ctxManager.MessageCount()*2 {
			tempMessages := make([]*session.Message, 0, len(messages))
			for _, m := range messages[1:] {
				tempMessages = append(tempMessages, &session.Message{
					Role:    session.Role(m.Role),
					Content: m.Content,
				})
			}
			state.SessionMessages = tempMessages
			messages = a.ctxManager.BuildContextMessages(systemPrompt, state.SessionMessages, "")
		}
	}

	return fmt.Errorf("maximum function call iterations reached")
}

// addMessageToSession adds a message to the session and persists it
func (a *Agent) addMessageToSession(role session.Role, content string, metadata map[string]string) {
	if a.store == nil {
		return
	}

	a.store.UpdateConversation(a.sessionID, func(conv *session.Conversation) error {
		conv.AddMessage(session.NewMessage(role, content, metadata))
		return nil
	})
}

// GetMessages returns the conversation history
func (a *Agent) GetMessages() []*session.Message {
	return a.messages
}

// GetClusterName returns the current cluster name
func (a *Agent) GetClusterName() string {
	return a.clusterName
}

// SetClusterName sets the current cluster name
func (a *Agent) SetClusterName(clusterName string) {
	a.clusterName = clusterName
	if a.store != nil {
		a.store.UpdateConversation(a.sessionID, func(conv *session.Conversation) error {
			conv.SetClusterContext(clusterName)
			return nil
		})
	}
}

// State tracks the conversation state
type State struct {
	ClusterName     string
	SessionMessages []*session.Message
}

// NewState creates a new conversation state
func NewState(clusterName string, messages []*session.Message) *State {
	return &State{
		ClusterName:     clusterName,
		SessionMessages: messages,
	}
}
