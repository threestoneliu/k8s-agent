package core

import "errors"

// TransitionError represents an invalid state transition error.
// It occurs when a signal is sent to a session in a state that does not
// support that particular transition.
type TransitionError struct {
	FromState State
	Signal    SessionSignal
	Message   string
}

// Error returns the error message describing the invalid transition.
func (e *TransitionError) Error() string {
	return "invalid transition: " + e.Message
}

// Validation errors for state transitions.
var (
	// ErrIntentNotValid indicates the parsed intent does not have all required fields.
	ErrIntentNotValid = errors.New("intent not valid, cannot confirm")
	// ErrIntentIncomplete indicates the parsed intent is missing required information.
	ErrIntentIncomplete = errors.New("intent incomplete, need clarification")
	// ErrInvalidTransition indicates the signal is not valid for the current state.
	ErrInvalidTransition = errors.New("invalid state transition")
	// ErrTerminalState indicates no transitions are allowed from terminal states.
	ErrTerminalState = errors.New("terminal state has no transitions")
)

// Transition validates and performs a state transition based on the signal.
// It returns the new state and an error if the transition is invalid.
// This function implements the state machine's transition logic.
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
// Valid signals: SignalConfirm (-> PLANNING), SignalModify (-> CLARIFYING), SignalAbort (-> FAILED)
func transitionFromParsing(signal SessionSignal) (State, error) {
	switch signal {
	case SignalConfirm:
		return StatePlanning, nil
	case SignalModify:
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
// Valid signals: SignalConfirm (-> PARSING), SignalModify (stay), SignalAbort (-> FAILED)
func transitionFromClarifying(signal SessionSignal) (State, error) {
	switch signal {
	case SignalConfirm:
		return StateParsing, nil
	case SignalModify:
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
// Valid signals: SignalConfirm (-> REVIEWING), SignalModify (stay), SignalAbort (-> FAILED)
func transitionFromPlanning(signal SessionSignal) (State, error) {
	switch signal {
	case SignalConfirm:
		return StateReviewing, nil
	case SignalModify:
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
// Valid signals: SignalConfirm (-> EXECUTING), SignalModify (-> PLANNING), SignalAbort (-> FAILED)
func transitionFromReviewing(signal SessionSignal) (State, error) {
	switch signal {
	case SignalConfirm:
		return StateExecuting, nil
	case SignalModify:
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
// Valid signals: SignalConfirm (-> COMPLETED), SignalAbort (-> FAILED)
func transitionFromExecuting(signal SessionSignal) (State, error) {
	switch signal {
	case SignalConfirm:
		return StateCompleted, nil
	case SignalAbort:
		return StateFailed, nil
	default:
		return StateExecuting, &TransitionError{
			FromState: StateExecuting,
			Signal:    signal,
			Message:   "signal not allowed from EXECUTING",
		}
	}
}