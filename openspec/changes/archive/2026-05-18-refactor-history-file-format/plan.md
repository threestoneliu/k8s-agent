# Implementation Plan

> **For agentic workers:** Use superpowers:subagent-driven-development to implement this plan task-by-task.

**Goal:** Refactor history file from JSON array to line-delimited text format.

**Files:**
- Modify: `pkg/ui/tui.go`

---

## Task 1: Update constant

**Files:**
- Modify: `pkg/ui/tui.go`

- [ ] **Step 1: Change historyFile constant**

```go
const historyFile = "~/.config/k8s-agent/history/history.txt"
```

Run: `go build ./pkg/ui/...`

---

## Task 2: Rewrite loadHistory

**Files:**
- Modify: `pkg/ui/tui.go`

- [ ] **Step 1: Replace loadHistory function**

```go
func loadHistory() ([]string, error) {
    path := expandPath(historyFile)
    data, err := os.ReadFile(path)
    if err != nil {
        if os.IsNotExist(err) {
            return []string{}, nil
        }
        return []string{}, err
    }
    lines := strings.Split(strings.TrimSpace(string(data)), "\n")
    var history []string
    for _, line := range lines {
        if line != "" {
            history = append(history, line)
        }
    }
    return history, nil
}
```

Run: `go build ./pkg/ui/...`

---

## Task 3: Rewrite saveHistory for append mode

**Files:**
- Modify: `pkg/ui/tui.go`

- [ ] **Step 1: Replace saveHistory with append mode**

```go
func saveHistory(history []string) error {
    path := expandPath(historyFile)
    dir := filepath.Dir(path)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return err
    }
    // Use append mode - each entry is one line
    f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        return err
    }
    defer f.Close()
    for _, entry := range history {
        if _, err := fmt.Fprintln(f, entry); err != nil {
            return err
        }
    }
    return nil
}
```

- [ ] **Step 2: Update call site to use append mode**

Instead of saving all history on each change, append only new entries:

```go
// In Enter key handling, replace full save with append
go appendHistory(userInput)
```

- [ ] **Step 3: Add appendHistory helper**

```go
func appendHistory(entry string) error {
    path := expandPath(historyFile)
    dir := filepath.Dir(path)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return err
    }
    f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        return err
    }
    defer f.Close()
    _, err = fmt.Fprintln(f, entry)
    return err
}
```

Run: `go build ./pkg/ui/...`

---

## Task 4: Add JSON migration

**Files:**
- Modify: `pkg/ui/tui.go`

- [ ] **Step 1: Add migrateFromJSON function**

```go
func migrateFromJSON() error {
    jsonPath := expandPath("~/.config/k8s-agent/history/history.json")
    txtPath := expandPath(historyFile)

    if _, err := os.Stat(jsonPath); err != nil {
        // JSON file doesn't exist, no migration needed
        return nil
    }

    // Read old JSON
    data, err := os.ReadFile(jsonPath)
    if err != nil {
        return err
    }

    var history []string
    if err := json.Unmarshal(data, &history); err != nil {
        return err
    }

    // Write new format
    f, err := os.Create(txtPath)
    if err != nil {
        return err
    }
    defer f.Close()

    for _, entry := range history {
        if _, err := fmt.Fprintln(f, entry); err != nil {
            return err
        }
    }

    // Delete old file
    os.Remove(jsonPath)
    return nil
}
```

- [ ] **Step 2: Call migrateFromJSON in newModel**

```go
func (t *TUI) newModel(...) tea.Model {
    // Migrate old JSON if exists
    migrateFromJSON()

    history, _ := loadHistory()
    // ...
}
```

Run: `go build ./pkg/ui/...`

---

## Task 5: Test

- [ ] **Step 1: Run build**

```bash
go build ./...
```

- [ ] **Step 2: Run tests**

```bash
go test ./pkg/ui/...
```

- [ ] **Step 3: Manual test**

1. Delete existing history file
2. Run TUI
3. Enter some commands
4. Check `~/.config/k8s-agent/history/history.txt` contains one entry per line