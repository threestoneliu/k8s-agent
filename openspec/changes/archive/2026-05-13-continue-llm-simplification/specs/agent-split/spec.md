# Agent Split

## ADDED Requirements

### Requirement: Agent Split Into Three Files

The `pkg/agent/` module SHALL be split into three focused files:
- `agent.go`: Main agent loop, IPC handling, LLM interaction
- `executor.go`: Tool call execution, result formatting
- `session.go`: Session management, message creation

#### Scenario: Three files exist with correct responsibilities

Given `pkg/agent/` directory
When listing Go files
Then agent.go, executor.go, session.go all exist

#### Scenario: agent.go contains main loop

Given `pkg/agent/agent.go`
When checking content
Then it contains `Agent` struct and `Run()` method

#### Scenario: executor.go contains tool execution

Given `pkg/agent/executor.go`
When checking content
Then it contains `ExecuteFunctionCall()` function

#### Scenario: session.go contains session management

Given `pkg/agent/session.go`
When checking content
Then it contains session creation and message helpers

---

### Requirement: Dependency Injection Instead of Global State

The `pkg/agent` module SHALL use dependency injection for all dependencies, not global `SetExecutor()` or `SetSchedulerManager()`.

#### Scenario: No global state setters

Given `pkg/agent/` module
When searching for `SetExecutor` or `SetSchedulerManager`
Then no result is found

#### Scenario: Dependencies injected via constructor

Given `NewAgent(llmSvc, executor, sessionMgr, ...)`
When creating agent
Then all dependencies are passed via constructor

---

### Requirement: Agent Communicates via IPC Channels

Agent SHALL receive input via `inputChan <-chan ui.Input` and send output via `outputChan chan<- ui.Output`.

#### Scenario: Run method signature is correct

Given `Agent.Run()` method
When checking signature
Then it matches: `Run(inputChan <-chan ui.Input, outputChan chan<- ui.Output)`
