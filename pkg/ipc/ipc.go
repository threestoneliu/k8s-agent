package ipc

// Input represents user input from UI to Agent.
// It carries the user's text command and the target cluster name.
type Input struct {
	// Text is the user's natural language input.
	Text string
	// ClusterName is the target cluster for this command.
	ClusterName string
}

// OutputType represents the type of output message sent from Agent to UI.
type OutputType string

// Output message types.
const (
	// OutputTypeText is a plain text response.
	OutputTypeText OutputType = "text"
	// OutputTypeThink indicates the agent is thinking/processing.
	OutputTypeThink OutputType = "think"
	// OutputTypeToolStart indicates a tool call is starting.
	OutputTypeToolStart OutputType = "tool_call_start"
	// OutputTypeToolResult contains the result of a tool call.
	OutputTypeToolResult OutputType = "tool_result"
	// OutputTypeDone indicates the response is complete.
	OutputTypeDone OutputType = "done"
	// OutputTypeError indicates an error occurred.
	OutputTypeError OutputType = "error"
)

// Output represents a message sent from Agent to UI.
// It contains various fields that may be populated depending on the output type.
type Output struct {
	// Type identifies the type of output message.
	Type OutputType
	// Content is the text content for text output.
	Content string
	// ToolName is the name of the tool for tool call outputs.
	ToolName string
	// ToolArgs is the arguments for the tool call.
	ToolArgs string
	// ToolResult is the result returned from the tool.
	ToolResult string
	// ToolSuccess indicates whether the tool call succeeded.
	ToolSuccess bool
	// ClusterName is the cluster the output relates to.
	ClusterName string
	// MessageType is an additional message type identifier.
	MessageType string
	// SessionID is the change session ID if applicable.
	SessionID string
	// State is the current state of the change session.
	State string
	// Plan is the change plan (core.ChangePlan) for review outputs.
	Plan interface{}
	// Diff is the resource diff (core.ResourceDiff) for diff outputs.
	Diff interface{}
	// ClarifyQuestion is a clarification question (core.ClarifyQuestion) if intent is incomplete.
	ClarifyQuestion interface{}
	// RequiresConfirm indicates the user needs to confirm before proceeding.
	RequiresConfirm bool
}