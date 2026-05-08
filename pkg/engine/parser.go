package engine

import (
	"errors"
	"strings"
)

// ErrEmptyInput indicates the input is empty
var ErrEmptyInput = errors.New("input cannot be empty")

// ErrInvalidInput indicates the input format is invalid
var ErrInvalidInput = errors.New("invalid input format")

// Parse parses a natural language command into a structured operation
func Parse(input string) (*ParsedOperation, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, ErrEmptyInput
	}

	parts := strings.Fields(input)
	if len(parts) < 2 {
		return nil, ErrInvalidInput
	}

	op := &ParsedOperation{
		RawInput: input,
		Flags:    make(map[string]string),
	}

	// First word is the verb
	op.Verb = parts[0]

	// Second word is the resource
	op.Resource = parts[1]

	// Parse remaining parts
	i := 2
	for i < len(parts) {
		part := parts[i]

		switch part {
		case "-n", "--namespace":
			if i+1 < len(parts) && !strings.HasPrefix(parts[i+1], "-") {
				op.Namespace = parts[i+1]
				i += 2
			} else {
				op.Namespace = ""
				i++
			}
		case "-l", "--selector":
			if i+1 < len(parts) && !strings.HasPrefix(parts[i+1], "-") {
				op.Flags["l"] = parts[i+1]
				i += 2
			} else {
				op.Flags["l"] = ""
				i++
			}
		case "--image", "--replicas":
			key := strings.TrimPrefix(part, "--")
			if i+1 < len(parts) && !strings.HasPrefix(parts[i+1], "-") {
				op.Flags[key] = parts[i+1]
				i += 2
			} else {
				op.Flags[key] = ""
				i++
			}
		case "--":
			// After --, everything is treated as positional arguments (names)
			i++
			if i < len(parts) {
				op.Name = parts[i]
				i++
			}
		default:
			// If this doesn't look like a flag, it might be the resource name
			if !strings.HasPrefix(part, "-") && op.Name == "" {
				op.Name = part
				i++
			} else if strings.HasPrefix(part, "--") && strings.Contains(part, "=") {
				kv := strings.SplitN(part, "=", 2)
				key := strings.TrimPrefix(kv[0], "--")
				if len(kv) > 1 {
					op.Flags[key] = kv[1]
				} else {
					op.Flags[key] = ""
				}
				i++
			} else if strings.HasPrefix(part, "-") && len(part) > 1 {
				// Short flag like -n
				key := string(part[1])
				if i+1 < len(parts) && !strings.HasPrefix(parts[i+1], "-") {
					op.Flags[key] = parts[i+1]
					i += 2
				} else {
					op.Flags[key] = ""
					i++
				}
			} else {
				i++
			}
		}
	}

	return op, nil
}
