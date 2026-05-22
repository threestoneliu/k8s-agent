package core

// ChangeManager defines the interface for managing change sessions.
// It orchestrates the state machine transitions and coordinates
// between intent parsing, planning, and execution.
type ChangeManager interface {
	// CreateSession creates a new change session and returns its ID.
	CreateSession() (string, error)

	// GetSession retrieves a session by its ID.
	GetSession(id string) (*ChangeSession, error)

	// SignalSession sends a signal to a session to trigger state transition.
	SignalSession(id string, signal SessionSignal) error

	// CloseSession closes a session and cleans up resources.
	CloseSession(id string) error
}