# Test Coverage

## ADDED Requirements

### Requirement: All Tests Pass

All existing tests SHALL pass after the refactor.

#### Scenario: go test ./... passes

Given the codebase after all changes
When running `go test ./...`
Then all tests pass with exit code 0

---

### Requirement: Agent Tests Updated for Split

The `pkg/agent/*_test.go` files SHALL be updated to work with the new file structure.

#### Scenario: Agent tests import correct paths

Given `pkg/agent/*_test.go`
When checking imports
Then they import correct module paths

---

### Requirement: Session Tests Work With Embedded Message

The `pkg/session/*_test.go` files SHALL work with `session.Message` embedding `shared.Message`.

#### Scenario: Session tests pass

Given `pkg/session/` tests
When running `go test ./pkg/session/...`
Then all tests pass
