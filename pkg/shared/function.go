package shared

// Function represents a callable function definition
type Function struct {
	Name        string
	Description string
	Parameters  map[string]interface{}
}

// FunctionCall represents a function call made during conversation
type FunctionCall struct {
	ID        string
	Name      string
	Arguments string
}

// FunctionResult represents the result of executing a function
type FunctionResult struct {
	Name          string
	Result        string
	Error         string
	Success       bool
	ClusterSwitch string
}
