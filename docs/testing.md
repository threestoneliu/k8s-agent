# Testing Guide

## Running Tests

### All Tests

```bash
go test ./...
```

### With Coverage

```bash
go test -cover ./...
```

### With Race Detection

```bash
go test -race ./...
```

### Specific Package

```bash
go test ./pkg/engine/...
go test ./cmd/cli/...
```

### E2E Tests

```bash
go test ./tests/e2e/...
```

### Integration Tests

Integration tests require a real Kubernetes cluster connection and are tagged with `integration` build tag.

```bash
# Run all tests including integration
go test -tags=integration ./...

# Run only integration tests
go test -tags=integration ./pkg/engine/...
```

**Note**: Integration tests will fail if no Kubernetes cluster is available. They use the current kubeconfig context.

## Test Coverage

Current coverage by package:

| Package | Coverage |
|---------|----------|
| cmd/cli | ~27% |
| pkg/cluster | ~96% |
| pkg/confirmation | ~83% |
| pkg/engine | ~83% |
| pkg/scheduler | ~88% |
| pkg/session | 100% |

To generate detailed coverage report:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## Test Structure

### Unit Tests
Located alongside source files with `_test.go` suffix:
- `pkg/engine/parser_test.go`
- `pkg/confirmation/manager_test.go`
- `pkg/scheduler/task_test.go`

### Integration Tests
Requires Kubernetes cluster, marked with build tags:
- `pkg/engine/executor_integration_test.go` (+build integration)

### E2E Tests
End-to-end flow tests in dedicated directory:
- `tests/e2e/e2e_test.go`

## Writing Tests

### Table-Driven Tests

```go
func TestExample(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {
            name:     "basic case",
            input:    "test input",
            expected: "expected output",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := functionUnderTest(tt.input)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

### Using MockExecutor

For testing operation flows without real K8s cluster:

```go
mockExec := &MockExecutor{
    ExecuteFunc: func(clusterName string, op *engine.ClassifiedOperation) (*engine.ExecutionResult, error) {
        return &engine.ExecutionResult{
            Success: true,
            Output:  "mock output",
        }, nil
    },
}
```

### Testing Confirmation Flow

```go
confirmMgr := confirmation.NewManager(5 * time.Second)

classifiedOp := &engine.ClassifiedOperation{
    Type:     engine.OperationTypeMutation,
    Verb:     "delete",
    Resource: "pods",
    Name:     "nginx",
}

// Create confirmation
confirmKey, err := confirmMgr.CreateConfirmation("test-cluster", classifiedOp)
require.NoError(t, err)

// Verify pending
pending, err := confirmMgr.GetConfirmation(confirmKey)
assert.Equal(t, confirmation.StatusPending, pending.Status)

// Approve
err = confirmMgr.ApproveConfirmation(confirmKey)
require.NoError(t, err)

// Verify approved
pending, err = confirmMgr.GetConfirmation(confirmKey)
assert.Equal(t, confirmation.StatusApproved, pending.Status)
```