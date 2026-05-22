// Package core provides core functionality for the k8s-agent change management system.
// It implements a state machine for managing Kubernetes change operations through
// a structured workflow: Parse -> Clarify -> Plan -> Review -> Execute -> Complete/Failed.
//
// The package provides:
//   - State machine implementation for change session lifecycle
//   - Change session management with state transitions
//   - Audit logging for all state transitions and actions
package core

import "time"

// State represents the state machine's current state in a change session.
// The state machine follows a structured workflow for Kubernetes change operations.
type State int

// State machine states for change management lifecycle.
// Sessions transition through these states based on user signals.
const (
	// StateParsing indicates the session is parsing natural language input into intent.
	StateParsing State = iota
	// StateClarifying indicates the session is requesting clarification from the user.
	StateClarifying
	// StatePlanning indicates the session is generating an execution plan.
	StatePlanning
	// StateReviewing indicates the session is presenting the plan for user review.
	StateReviewing
	// StateExecuting indicates the session is executing approved changes.
	StateExecuting
	// StateCompleted indicates the change was successfully completed.
	StateCompleted
	// StateFailed indicates the change failed or was aborted.
	StateFailed
)

// SessionSignal represents signals that can be sent to a session to trigger
// state transitions during the change management workflow.
type SessionSignal int

// Session signals for controlling state machine transitions.
// These signals drive the session through its lifecycle.
const (
	// SignalConfirm approves the current state and proceeds to the next workflow step.
	SignalConfirm SessionSignal = iota
	// SignalAbort cancels the current session, transitioning to FAILED state.
	SignalAbort
	// SignalModify requests modification of the current state's data.
	SignalModify
	// SignalProceed forces progression to the next state.
	SignalProceed
)

// String returns the human-readable name of the state.
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

// String returns the human-readable name of the signal.
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
// Sessions are created when users initiate a change and track the operation
// through the Parse -> Clarify -> Plan -> Review -> Execute workflow.
type ChangeSession struct {
	// ID uniquely identifies the session.
	ID string
	// State is the current position in the state machine.
	State State
}

// NewChangeSession creates a new change session with the PARSING state.
// The session ID should be a unique identifier (e.g., UUID) provided by the caller.
func NewChangeSession(id string) *ChangeSession {
	return &ChangeSession{
		ID:    id,
		State: StateParsing,
	}
}

// Transition moves the session to a new state based on the provided signal.
// It validates the transition and updates the session state. All transitions
// are logged to the audit log for traceability.
func (s *ChangeSession) Transition(signal SessionSignal) error {
	oldState := s.State
	newState, err := Transition(s.State, signal)
	if err != nil {
		return err
	}
	s.State = newState

	// Log state transition
	Log(s.ID, "state_transition", "session", map[string]interface{}{
		"from_state": oldState.String(),
		"to_state":   newState.String(),
		"signal":     signal.String(),
	})

	return nil
}

// CreatedAt returns the session creation timestamp.
// Note: This is a placeholder - actual timestamp tracking will be implemented.
func (s *ChangeSession) CreatedAt() time.Time {
	return time.Time{}
}