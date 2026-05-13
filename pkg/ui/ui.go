package ui

import "k8s-agent/pkg/ipc"

// Re-export IPC types for backward compatibility
type Input = ipc.Input
type Output = ipc.Output
type OutputType = ipc.OutputType

// Keep constants
const (
	OutputTypeText       = ipc.OutputTypeText
	OutputTypeThink      = ipc.OutputTypeThink
	OutputTypeToolStart  = ipc.OutputTypeToolStart
	OutputTypeToolResult = ipc.OutputTypeToolResult
	OutputTypeDone       = ipc.OutputTypeDone
	OutputTypeError      = ipc.OutputTypeError
)