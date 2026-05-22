package core

import (
	"testing"
)

func TestStateString(t *testing.T) {
	tests := []struct {
		state    State
		expected string
	}{
		{StateParsing, "PARSING"},
		{StateClarifying, "CLARIFYING"},
		{StatePlanning, "PLANNING"},
		{StateReviewing, "REVIEWING"},
		{StateExecuting, "EXECUTING"},
		{StateCompleted, "COMPLETED"},
		{StateFailed, "FAILED"},
	}

	for _, tc := range tests {
		if got := tc.state.String(); got != tc.expected {
			t.Errorf("State.String() = %v, want %v", got, tc.expected)
		}
	}
}

func TestSessionSignalString(t *testing.T) {
	tests := []struct {
		signal   SessionSignal
		expected string
	}{
		{SignalConfirm, "Confirm"},
		{SignalAbort, "Abort"},
		{SignalModify, "Modify"},
		{SignalProceed, "Proceed"},
	}

	for _, tc := range tests {
		if got := tc.signal.String(); got != tc.expected {
			t.Errorf("SessionSignal.String() = %v, want %v", got, tc.expected)
		}
	}
}

func TestStateConstants(t *testing.T) {
	// Verify state constants are in expected order
	if StateParsing != 0 {
		t.Errorf("StateParsing = %v, want 0", StateParsing)
	}
	if StateCompleted != 5 {
		t.Errorf("StateCompleted = %v, want 5", StateCompleted)
	}
	if StateFailed != 6 {
		t.Errorf("StateFailed = %v, want 6", StateFailed)
	}
}

func TestSignalConstants(t *testing.T) {
	if SignalConfirm != 0 {
		t.Errorf("SignalConfirm = %v, want 0", SignalConfirm)
	}
	if SignalProceed != 3 {
		t.Errorf("SignalProceed = %v, want 3", SignalProceed)
	}
}

// Transition tests

func TestTransitionFromParsing(t *testing.T) {
	tests := []struct {
		name        string
		signal      SessionSignal
		wantState   State
		wantErr     bool
	}{
		{"Confirm -> PLANNING", SignalConfirm, StatePlanning, false},
		{"Modify -> CLARIFYING", SignalModify, StateClarifying, false},
		{"Abort -> FAILED", SignalAbort, StateFailed, false},
		{"Proceed -> invalid", SignalProceed, StateParsing, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			state, err := Transition(StateParsing, tc.signal)
			if tc.wantErr && err == nil {
				t.Errorf("Transition(PARSING, %v) expected error, got nil", tc.signal)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("Transition(PARSING, %v) unexpected error: %v", tc.signal, err)
			}
			if state != tc.wantState {
				t.Errorf("Transition(PARSING, %v) = %v, want %v", tc.signal, state, tc.wantState)
			}
		})
	}
}

func TestTransitionFromClarifying(t *testing.T) {
	tests := []struct {
		name        string
		signal      SessionSignal
		wantState   State
		wantErr     bool
	}{
		{"Confirm -> PARSING", SignalConfirm, StateParsing, false},
		{"Modify -> CLARIFYING", SignalModify, StateClarifying, false},
		{"Abort -> FAILED", SignalAbort, StateFailed, false},
		{"Proceed -> invalid", SignalProceed, StateClarifying, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			state, err := Transition(StateClarifying, tc.signal)
			if tc.wantErr && err == nil {
				t.Errorf("Transition(CLARIFYING, %v) expected error, got nil", tc.signal)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("Transition(CLARIFYING, %v) unexpected error: %v", tc.signal, err)
			}
			if state != tc.wantState {
				t.Errorf("Transition(CLARIFYING, %v) = %v, want %v", tc.signal, state, tc.wantState)
			}
		})
	}
}

func TestTransitionFromPlanning(t *testing.T) {
	tests := []struct {
		name        string
		signal      SessionSignal
		wantState   State
		wantErr     bool
	}{
		{"Confirm -> REVIEWING", SignalConfirm, StateReviewing, false},
		{"Modify -> PLANNING", SignalModify, StatePlanning, false},
		{"Abort -> FAILED", SignalAbort, StateFailed, false},
		{"Proceed -> invalid", SignalProceed, StatePlanning, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			state, err := Transition(StatePlanning, tc.signal)
			if tc.wantErr && err == nil {
				t.Errorf("Transition(PLANNING, %v) expected error, got nil", tc.signal)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("Transition(PLANNING, %v) unexpected error: %v", tc.signal, err)
			}
			if state != tc.wantState {
				t.Errorf("Transition(PLANNING, %v) = %v, want %v", tc.signal, state, tc.wantState)
			}
		})
	}
}

func TestTransitionFromReviewing(t *testing.T) {
	tests := []struct {
		name        string
		signal      SessionSignal
		wantState   State
		wantErr     bool
	}{
		{"Confirm -> EXECUTING", SignalConfirm, StateExecuting, false},
		{"Modify -> PLANNING", SignalModify, StatePlanning, false},
		{"Abort -> FAILED", SignalAbort, StateFailed, false},
		{"Proceed -> invalid", SignalProceed, StateReviewing, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			state, err := Transition(StateReviewing, tc.signal)
			if tc.wantErr && err == nil {
				t.Errorf("Transition(REVIEWING, %v) expected error, got nil", tc.signal)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("Transition(REVIEWING, %v) unexpected error: %v", tc.signal, err)
			}
			if state != tc.wantState {
				t.Errorf("Transition(REVIEWING, %v) = %v, want %v", tc.signal, state, tc.wantState)
			}
		})
	}
}

func TestTransitionFromExecuting(t *testing.T) {
	tests := []struct {
		name        string
		signal      SessionSignal
		wantState   State
		wantErr     bool
	}{
		{"Confirm -> COMPLETED", SignalConfirm, StateCompleted, false},
		{"Abort -> FAILED", SignalAbort, StateFailed, false},
		{"Modify -> invalid", SignalModify, StateExecuting, true},
		{"Proceed -> invalid", SignalProceed, StateExecuting, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			state, err := Transition(StateExecuting, tc.signal)
			if tc.wantErr && err == nil {
				t.Errorf("Transition(EXECUTING, %v) expected error, got nil", tc.signal)
			}
			if !tc.wantErr && err != nil {
				t.Errorf("Transition(EXECUTING, %v) unexpected error: %v", tc.signal, err)
			}
			if state != tc.wantState {
				t.Errorf("Transition(EXECUTING, %v) = %v, want %v", tc.signal, state, tc.wantState)
			}
		})
	}
}

func TestTransitionFromTerminalStates(t *testing.T) {
	tests := []struct {
		name      string
		state     State
		signal    SessionSignal
	}{
		{"COMPLETED + Confirm", StateCompleted, SignalConfirm},
		{"COMPLETED + Modify", StateCompleted, SignalModify},
		{"COMPLETED + Abort", StateCompleted, SignalAbort},
		{"COMPLETED + Proceed", StateCompleted, SignalProceed},
		{"FAILED + Confirm", StateFailed, SignalConfirm},
		{"FAILED + Modify", StateFailed, SignalModify},
		{"FAILED + Abort", StateFailed, SignalAbort},
		{"FAILED + Proceed", StateFailed, SignalProceed},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			state, err := Transition(tc.state, tc.signal)
			// Terminal states should remain unchanged
			if state != tc.state {
				t.Errorf("Transition(%v, %v) = %v, want unchanged state %v", tc.state, tc.signal, state, tc.state)
			}
			// Should always return an error
			if err == nil {
				t.Errorf("Transition(%v, %v) expected error for terminal state, got nil", tc.state, tc.signal)
			}
		})
	}
}

func TestTransitionErrorMessage(t *testing.T) {
	err := &TransitionError{
		FromState: StateParsing,
		Signal:    SignalProceed,
		Message:   "test error",
	}
	if err.Error() != "invalid transition: test error" {
		t.Errorf("TransitionError.Error() = %v, want 'invalid transition: test error'", err.Error())
	}
}

func TestHappyPath(t *testing.T) {
	// Test the complete happy path: PARSING -> CLARIFYING -> PLANNING -> REVIEWING -> EXECUTING -> COMPLETED
	transitions := []struct {
		from  State
		sig   SessionSignal
		to    State
	}{
		{StateParsing, SignalConfirm, StatePlanning},
		{StateClarifying, SignalConfirm, StateParsing},
		{StatePlanning, SignalConfirm, StateReviewing},
		{StateReviewing, SignalConfirm, StateExecuting},
		{StateExecuting, SignalConfirm, StateCompleted},
	}

	for _, tr := range transitions {
		state, err := Transition(tr.from, tr.sig)
		if err != nil {
			t.Errorf("Transition(%v, %v) unexpected error: %v", tr.from, tr.sig, err)
		}
		if state != tr.to {
			t.Errorf("Transition(%v, %v) = %v, want %v", tr.from, tr.sig, state, tr.to)
		}
	}
}

func TestModifyPath(t *testing.T) {
	// Test the modify path: REVIEWING -> PLANNING -> REVIEWING
	transitions := []struct {
		from  State
		sig   SessionSignal
		to    State
	}{
		{StateReviewing, SignalModify, StatePlanning},
		{StatePlanning, SignalConfirm, StateReviewing},
	}

	for _, tr := range transitions {
		state, err := Transition(tr.from, tr.sig)
		if err != nil {
			t.Errorf("Transition(%v, %v) unexpected error: %v", tr.from, tr.sig, err)
		}
		if state != tr.to {
			t.Errorf("Transition(%v, %v) = %v, want %v", tr.from, tr.sig, state, tr.to)
		}
	}
}

func TestAbortPath(t *testing.T) {
	// Test abort from various states
	states := []State{StateParsing, StateClarifying, StatePlanning, StateReviewing, StateExecuting}
	for _, s := range states {
		state, err := Transition(s, SignalAbort)
		if err != nil {
			t.Errorf("Transition(%v, Abort) unexpected error: %v", s, err)
		}
		if state != StateFailed {
			t.Errorf("Transition(%v, Abort) = %v, want FAILED", s, state)
		}
	}
}