package core

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