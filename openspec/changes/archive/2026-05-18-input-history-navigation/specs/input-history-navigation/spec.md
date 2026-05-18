## ADDED Requirements

### Requirement: Up/Down Key Navigation

The TUI SHALL intercept Up/Down key events to navigate input history when the textinput is focused.

#### Scenarios

- **Scenario: Navigate up**
  - Given user is typing in textinput
  - When user presses Up key
  - Then the previous history item is loaded into the textinput
  - And historyIndex is decremented

- **Scenario: Navigate down**
  - Given user has pressed Up key to browse history
  - When user presses Down key
  - Then the next history item is loaded (or current input restored if at end)

- **Scenario: Enter saves to history**
  - Given user presses Enter with non-empty input
  - When the input is sent to the agent
  - Then the input is also saved to history

### Requirement: History Persistence

The TUI SHALL persist history to `~/.config/k8s-agent/history/history.json` and load it on startup.

#### Scenarios

- **Scenario: Load history on startup**
  - Given TUI starts
  - When the application starts
  - Then load history from `~/.config/k8s-agent/history/history.json`

- **Scenario: Save history after Enter**
  - Given user presses Enter with input "kubectl get pods"
  - When the input is sent to the agent
  - Then append "kubectl get pods" to history and write to file

- **Scenario: History limit 100**
  - Given history has 100 items
  - When user adds a new item
  - Then the oldest item is removed before adding the new one

### Requirement: Shared History

All clusters SHALL share the same history file.

#### Scenarios

- **Scenario: Shared history across clusters**
  - Given user has history items in cluster "dev"
  - When user switches to cluster "prod"
  - Then the same history is available in "prod"

### Requirement: Clear History Command

The TUI SHALL support `/clear-history` command to clear the history.

#### Scenarios

- **Scenario: Clear history command**
  - Given user types `/clear-history`
  - When Enter is pressed
  - Then the history file is deleted
  - And history in memory is cleared

---

## MODIFIED Requirements

None.

---

## REMOVED Requirements

None.

---

## RENAMED Requirements

None.