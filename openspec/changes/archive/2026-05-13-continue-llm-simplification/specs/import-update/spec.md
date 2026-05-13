# Import Update

## ADDED Requirements

### Requirement: No Circular Dependencies

No module SHALL import in a way that creates circular dependencies.

#### Scenario: go mod verify passes

Given all module dependencies
When running `go mod verify`
Then no errors related to cyclic imports

---

### Requirement: All Imports Use Correct Paths

All imports SHALL use correct module paths:
- `k8s-agent/pkg/k8s` (not `pkg/engine`)
- `k8s-agent/pkg/shared` for shared types
- `k8s-agent/pkg/session` for session types

#### Scenario: No imports of old paths

Given all Go files
When searching for `"k8s-agent/pkg/engine"`
Then no result is found

---

### Requirement: Build Succeeds After Refactor

After all changes, `go build ./...` SHALL succeed without errors.

#### Scenario: Build passes

Given the codebase after all changes
When running `go build ./...`
Then exit code is 0 with no errors
