package core

import "time"

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
func (s *ChangeSession) Transition(signal SessionSignal) {
	switch signal {
	case SignalConfirm, SignalProceed:
		s.advanceState()
	case SignalAbort:
		s.State = StateFailed
	case SignalModify:
		// Modify typically returns to a previous state
		s.State = StateParsing
	}
}

// advanceState moves the state forward in the happy path.
func (s *ChangeSession) advanceState() {
	switch s.State {
	case StateParsing:
		s.State = StateClarifying
	case StateClarifying:
		s.State = StatePlanning
	case StatePlanning:
		s.State = StateReviewing
	case StateReviewing:
		s.State = StateExecuting
	case StateExecuting:
		s.State = StateCompleted
	default:
		// Already in terminal state, no-op
	}
}

// CreatedAt is the session creation timestamp.
func (s *ChangeSession) CreatedAt() time.Time {
	// Placeholder for future implementation
	return time.Time{}
}