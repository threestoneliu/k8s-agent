package engine

// ParsedOperation represents a structured operation parsed from user input
type ParsedOperation struct {
	Verb      string
	Resource  string
	Name      string
	Namespace string
	Flags     map[string]string
	RawInput  string
}
