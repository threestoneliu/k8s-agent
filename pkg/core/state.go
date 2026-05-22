package core

import "errors"

// TransitionError represents an invalid state transition error.
type TransitionError struct {
	FromState State
	Signal    SessionSignal
	Message   string
}

func (e *TransitionError) Error() string {
	return "invalid transition: " + e.Message
}

// Validation errors for state transitions.
var (
	ErrIntentNotValid    = errors.New("intent not valid, cannot confirm")
	ErrIntentIncomplete  = errors.New("intent incomplete, need clarification")
	ErrInvalidTransition = errors.New("invalid state transition")
	ErrTerminalState     = errors.New("terminal state has no transitions")
)

// Transition validates and performs a state transition based on the signal.
// It returns the new state and an error if the transition is invalid.
func Transition(currentState State, signal SessionSignal) (State, error) {
	switch currentState {
	case StateParsing:
		return transitionFromParsing(signal)
	case StateClarifying:
		return transitionFromClarifying(signal)
	case StatePlanning:
		return transitionFromPlanning(signal)
	case StateReviewing:
		return transitionFromReviewing(signal)
	case StateExecuting:
		return transitionFromExecuting(signal)
	case StateCompleted, StateFailed:
		return currentState, &TransitionError{
			FromState: currentState,
			Signal:   signal,
			Message:  "terminal state has no transitions",
		}
	default:
		return currentState, ErrInvalidTransition
	}
}

// transitionFromParsing handles transitions from PARSING state.
func transitionFromParsing(signal SessionSignal) (State, error) {
	switch signal {
	case SignalConfirm:
		// Confirm → PLANNING (if intent valid)
		return StatePlanning, nil
	case SignalModify:
		// Modify → CLARIFYING (if intent incomplete)
		return StateClarifying, nil
	case SignalAbort:
		return StateFailed, nil
	default:
		return StateParsing, &TransitionError{
			FromState: StateParsing,
			Signal:    signal,
			Message:   "signal not allowed from PARSING",
		}
	}
}

// transitionFromClarifying handles transitions from CLARIFYING state.
func transitionFromClarifying(signal SessionSignal) (State, error) {
	switch signal {
	case SignalConfirm:
		// Confirm → PARSING (with clarified intent)
		return StateParsing, nil
	case SignalModify:
		// Modify stays in clarifying for more clarification
		return StateClarifying, nil
	case SignalAbort:
		return StateFailed, nil
	default:
		return StateClarifying, &TransitionError{
			FromState: StateClarifying,
			Signal:    signal,
			Message:   "signal not allowed from CLARIFYING",
		}
	}
}

// transitionFromPlanning handles transitions from PLANNING state.
func transitionFromPlanning(signal SessionSignal) (State, error) {
	switch signal {
	case SignalConfirm:
		// Confirm → REVIEWING
		return StateReviewing, nil
	case SignalModify:
		// Modify stays in planning for revision
		return StatePlanning, nil
	case SignalAbort:
		return StateFailed, nil
	default:
		return StatePlanning, &TransitionError{
			FromState: StatePlanning,
			Signal:    signal,
			Message:   "signal not allowed from PLANNING",
		}
	}
}

// transitionFromReviewing handles transitions from REVIEWING state.
func transitionFromReviewing(signal SessionSignal) (State, error) {
	switch signal {
	case SignalConfirm:
		// Confirm → EXECUTING
		return StateExecuting, nil
	case SignalModify:
		// Modify → PLANNING
		return StatePlanning, nil
	case SignalAbort:
		return StateFailed, nil
	default:
		return StateReviewing, &TransitionError{
			FromState: StateReviewing,
			Signal:    signal,
			Message:   "signal not allowed from REVIEWING",
		}
	}
}

// transitionFromExecuting handles transitions from EXECUTING state.
func transitionFromExecuting(signal SessionSignal) (State, error) {
	switch signal {
	case SignalConfirm:
		// Confirm → COMPLETED
		return StateCompleted, nil
	case SignalAbort:
		// Abort → FAILED
		return StateFailed, nil
	default:
		return StateExecuting, &TransitionError{
			FromState: StateExecuting,
			Signal:    signal,
			Message:   "signal not allowed from EXECUTING",
		}
	}
}