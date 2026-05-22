package core

import "time"

// State represents the state machine's current state in a change session.
type State int

// State machine states for change management lifecycle.
const (
	StateParsing    State = iota // PARSING - parsing natural language input
	StateClarifying              // CLARIFYING - clarifying ambiguous intent
	StatePlanning                // PLANNING - planning execution steps
	StateReviewing               // REVIEWING - reviewing proposed changes
	StateExecuting               // EXECUTING - executing approved changes
	StateCompleted               // COMPLETED - change successfully completed
	StateFailed                  // FAILED - change failed or aborted
)

// SessionSignal represents signals that can be sent to a session.
type SessionSignal int

// Session signals for controlling state machine transitions.
const (
	SignalConfirm SessionSignal = iota // Confirm - approve current state
	SignalAbort                         // Abort - abort the session
	SignalModify                        // Modify - request modification
	SignalProceed                       // Proceed - proceed to next state
)

// String returns the string representation of a State.
func (s State) String() string {
	switch s {
	case StateParsing:
		return "PARSING"
	case StateClarifying:
		return "CLARIFYING"
	case StatePlanning:
		return "PLANNING"
	case StateReviewing:
		return "REVIEWING"
	case StateExecuting:
		return "EXECUTING"
	case StateCompleted:
		return "COMPLETED"
	case StateFailed:
		return "FAILED"
	default:
		return "UNKNOWN"
	}
}

// String returns the string representation of a SessionSignal.
func (s SessionSignal) String() string {
	switch s {
	case SignalConfirm:
		return "Confirm"
	case SignalAbort:
		return "Abort"
	case SignalModify:
		return "Modify"
	case SignalProceed:
		return "Proceed"
	default:
		return "UNKNOWN"
	}
}

// ChangeSession represents a single change management session.
// It tracks the state machine state and metadata for a change operation.
type ChangeSession struct {
	ID    string
	State State
	// TODO: Add intent parsing, planning, and execution fields in future tasks
}

// NewChangeSession creates a new change session with the PARSING state.
func NewChangeSession(id string) *ChangeSession {
	return &ChangeSession{
		ID:    id,
		State: StateParsing,
	}
}

// Transition moves the session to a new state based on the signal.
// It uses the standalone Transition function and updates the session state.
func (s *ChangeSession) Transition(signal SessionSignal) error {
	newState, err := Transition(s.State, signal)
	if err != nil {
		return err
	}
	s.State = newState
	return nil
}

// CreatedAt is the session creation timestamp.
func (s *ChangeSession) CreatedAt() time.Time {
	// Placeholder for future implementation
	return time.Time{}
}