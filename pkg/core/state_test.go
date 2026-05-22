package core

import "testing"

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

func TestNewChangeSession(t *testing.T) {
	session := NewChangeSession("test-id")
	if session.ID != "test-id" {
		t.Errorf("NewChangeSession().ID = %v, want test-id", session.ID)
	}
	if session.State != StateParsing {
		t.Errorf("NewChangeSession().State = %v, want StateParsing", session.State)
	}
}

func TestChangeSessionTransition(t *testing.T) {
	session := NewChangeSession("test")

	// Test Proceed advances state
	session.Transition(SignalProceed)
	if session.State != StateClarifying {
		t.Errorf("After Proceed: State = %v, want StateClarifying", session.State)
	}

	// Test Abort sets to Failed
	session.Transition(SignalAbort)
	if session.State != StateFailed {
		t.Errorf("After Abort: State = %v, want StateFailed", session.State)
	}

	// Test Modify resets to Parsing
	session.State = StateReviewing
	session.Transition(SignalModify)
	if session.State != StateParsing {
		t.Errorf("After Modify: State = %v, want StateParsing", session.State)
	}
}

func TestAdvanceState(t *testing.T) {
	session := NewChangeSession("test")

	// Happy path: parse -> clarify -> plan -> review -> execute -> completed
	transitions := []State{
		StateParsing, StateClarifying, StatePlanning,
		StateReviewing, StateExecuting, StateCompleted,
	}

	current := StateParsing
	for _, expected := range transitions {
		if current != expected {
			t.Errorf("State = %v, want %v", current, expected)
		}
		if current != StateCompleted {
			session.advanceState()
			current = session.State
		}
	}

	// Completed state should not advance further
	session.advanceState()
	if session.State != StateCompleted {
		t.Errorf("After advance from COMPLETED: State = %v, want StateCompleted", session.State)
	}
}