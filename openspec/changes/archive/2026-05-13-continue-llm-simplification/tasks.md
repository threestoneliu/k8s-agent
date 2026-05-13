# Tasks: Continue LLM/Agent/Session Simplification

## 1. Delete Scheduler Module

- [ ] 1.1 Delete `pkg/scheduler/` directory
- [ ] 1.2 Remove scheduled task function handlers from llm/functions.go
- [ ] 1.3 Run `go build ./...` to verify no broken imports
- [ ] 1.4 Run tests to ensure no regressions

## 2. LLM Module Cleanup

- [ ] 2.1 Check for any remaining Provider interface references
- [ ] 2.2 Remove any unused imports in pkg/llm/
- [ ] 2.3 Verify llm.Service uses OpenAISDKProvider directly
- [ ] 2.4 Run `go test ./pkg/llm/...` to verify

## 3. Rename pkg/engine to pkg/k8s

- [ ] 3.1 Rename `pkg/engine/` directory to `pkg/k8s/`
- [ ] 3.2 Update all imports from `k8s-agent/pkg/engine` to `k8s-agent/pkg/k8s`
- [ ] 3.3 Merge executor_query.go into executor.go
- [ ] 3.4 Remove parser.go and classifier.go
- [ ] 3.5 Run `go build ./...` to verify
- [ ] 3.6 Update CLAUDE.md if it references pkg/engine

## 4. Agent Module Split

- [ ] 4.1 Create `pkg/agent/executor.go` with ExecuteFunctionCall
- [ ] 4.2 Create `pkg/agent/session.go` with session management helpers
- [ ] 4.3 Refactor `pkg/agent/agent.go` to use the new files
- [ ] 4.4 Remove global state (SetExecutor/SetSchedulerManager)
- [ ] 4.5 Update agent tests to work with new structure
- [ ] 4.6 Run `go test ./pkg/agent/...` to verify

## 5. Session Compression Simplify

- [ ] 5.1 Merge findCompleteInteractions and findLLMCompleteInteractions
- [ ] 5.2 Verify unified Interaction struct exists once
- [ ] 5.3 Ensure L1/L2/L3 compression triggers still work correctly
- [ ] 5.4 Run `go test ./pkg/session/...` to verify

## 6. Import Update and Build Verification

- [ ] 6.1 Run `go mod tidy` to clean up dependencies
- [ ] 6.2 Run `go build ./...` and fix any remaining errors
- [ ] 6.3 Run `go test ./...` and fix any test failures
- [ ] 6.4 Verify no circular dependencies with `go mod verify`

## 7. Functional Verification

- [ ] 7.1 Build the binary: `go build -o k8s-agent ./cmd/k8s-agent`
- [ ] 7.2 Run chat command with dry-run to verify basic flow
- [ ] 7.3 Verify all subcommands work: chat, cluster
