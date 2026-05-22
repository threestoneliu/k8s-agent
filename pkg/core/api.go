package core

// ChangeManager defines the interface for managing change sessions.
// It orchestrates the state machine transitions and coordinates
// between intent parsing, planning, and execution.
type ChangeManager interface {
	// CreateSession creates a new change session and returns its ID.
	// The new session starts in the PARSING state.
	CreateSession() (string, error)

	// GetSession retrieves a session by its ID.
	// Returns ErrSessionNotFound if the session does not exist.
	GetSession(id string) (*ChangeSession, error)

	// SignalSession sends a signal to a session to trigger state transition.
	// Returns an error if the signal is invalid for the current state.
	SignalSession(id string, signal SessionSignal) error

	// CloseSession closes a session and cleans up resources.
	// This should be called when a session is complete or aborted.
	CloseSession(id string) error
}