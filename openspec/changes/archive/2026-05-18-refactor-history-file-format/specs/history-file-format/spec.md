## ADDED Requirements

### Requirement: Line-Delimited Text Format

The history file SHALL use line-delimited text format where each line represents a single history entry.

#### Scenarios

- **Scenario: File format**
  - Given history file exists at `~/.config/k8s-agent/history/history.txt`
  - When the file is opened in a text editor
  - Then each history entry appears on its own line
  - And empty lines are ignored

- **Scenario: Entry with special characters**
  - Given user enters `kubectl get pods -n "my namespace"`
  - When saved to history
  - Then the entry appears exactly as entered on a single line

### Requirement: Append-Only Save

The TUI SHALL append new history entries to the file rather than rewriting the entire file.

#### Scenarios

- **Scenario: Append new entry**
  - Given history file has existing entries
  - When user enters a new command and presses Enter
  - Then the new entry is appended to the file
  - And existing entries are preserved

- **Scenario: Large history**
  - Given history has 100 entries
  - When user adds a new entry
  - Then only the new entry is written
  - And file is not rewritten completely

### Requirement: Migration from JSON

The TUI SHALL automatically migrate from the old JSON format to the new line-delimited format on first startup.

#### Scenarios

- **Scenario: Old JSON exists**
  - Given `history.json` exists at `~/.config/k8s-agent/history/`
  - When TUI starts
  - Then the JSON file is read, converted to line-delimited format, and deleted
  - And `history.txt` is created with the same entries

- **Scenario: No old file**
  - Given `history.json` does not exist
  - When TUI starts
  - Then no migration occurs
  - And normal operation continues

---

## MODIFIED Requirements

None.

---

## REMOVED Requirements

None.

---

## RENAMED Requirements

None.