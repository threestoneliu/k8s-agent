package ipc

// Input represents user input from UI to Agent
type Input struct {
	Text        string
	ClusterName string
}

// OutputType represents the type of output message
type OutputType string

const (
	OutputTypeText       OutputType = "text"
	OutputTypeThink      OutputType = "think"
	OutputTypeToolStart  OutputType = "tool_call_start"
	OutputTypeToolResult OutputType = "tool_result"
	OutputTypeDone       OutputType = "done"
	OutputTypeError      OutputType = "error"
)

// Output represents a message sent from Agent to UI
type Output struct {
	Type            OutputType
	Content         string
	ToolName        string
	ToolArgs        string
	ToolResult      string
	ToolSuccess     bool
	ClusterName     string
	MessageType     string
	SessionID       string
	State           string
	Plan            interface{} // *core.ChangePlan
	Diff            interface{} // *core.ResourceDiff
	ClarifyQuestion interface{} // *core.ClarifyQuestion
	RequiresConfirm bool
}