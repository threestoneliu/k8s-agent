# Input History Navigation Implementation Plan

> **For agentic workers:** Use superpowers:subagent-driven-development to implement this plan task-by-task.

**Goal:** Add up/down key navigation for input history in TUI, with shared persistent storage.

**Architecture:** Add history fields to Model, intercept up/down keys, save to ~/.config/k8s-agent/history/history.json

**Tech Stack:** Go, k8s-agent codebase, charmbracelet/bubbletea

---

## Task 1: Add history fields to Model

**Files:**
- Modify: `pkg/ui/tui.go`

- [ ] **Step 1: Add fields to Model struct**

```go
type Model struct {
    // ... existing fields ...
    history       []string  // 输入历史
    historyIndex  int       // 当前浏览位置，-1 表示不在浏览历史
}
```

Run: `go build ./pkg/ui/...`

- [ ] **Step 2: Commit**

```bash
git add pkg/ui/tui.go
git commit -m "feat(ui): add history fields to Model"
```

---

## Task 2: Implement history load/save functions

**Files:**
- Modify: `pkg/ui/tui.go`

- [ ] **Step 1: Add history file path constant**

```go
const historyFile = "~/.config/k8s-agent/history/history.json"
```

- [ ] **Step 2: Add loadHistory function**

```go
func loadHistory() ([]string, error) {
    path := expandPath(historyFile)
    data, err := os.ReadFile(path)
    if err != nil {
        if os.IsNotExist(err) {
            return []string{}, nil
        }
        return nil, err
    }
    var history []string
    if err := json.Unmarshal(data, &history); err != nil {
        return []string{}, nil  // 损坏的文件，返回空
    }
    return history, nil
}
```

- [ ] **Step 3: Add saveHistory function**

```go
func saveHistory(history []string) error {
    path := expandPath(historyFile)
    data, err := json.Marshal(history)
    if err != nil {
        return err
    }
    return os.WriteFile(path, data, 0644)
}
```

- [ ] **Step 4: Initialize history in Model**

在 `New` 函数或初始化时调用 `loadHistory`

Run: `go build ./pkg/ui/...`

- [ ] **Step 5: Commit**

```bash
git add pkg/ui/tui.go
git commit -m "feat(ui): add history load/save functions"
```

---

## Task 3: Modify key handling for up/down

**Files:**
- Modify: `pkg/ui/tui.go`

- [ ] **Step 1: Add key handling in Update method**

在 `Update` 方法中，当收到 `tea.KeyMsg` 时：
- 如果是 `key.Up`：浏览上一条历史
- 如果是 `key.Down`：浏览下一条历史

```go
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case tea.KeyMsg:
        switch msg.Type {
        case key.Up:
            if len(m.history) > 0 && m.historyIndex < len(m.history)-1 {
                if m.historyIndex == -1 {
                    m.tempInput = m.textinput.Value()
                }
                m.historyIndex++
                m.textinput.SetValue(m.history[m.historyIndex])
            }
        case key.Down:
            if m.historyIndex == -1 {
                // 不在浏览历史，不处理
            } else if m.historyIndex == len(m.history)-1 {
                // 已到末尾，恢复 tempInput
                m.textinput.SetValue(m.tempInput)
                m.historyIndex = -1
            } else {
                m.historyIndex++
                m.textinput.SetValue(m.history[m.historyIndex])
            }
        }
    }
    // ... rest of Update ...
}
```

Run: `go build ./pkg/ui/...`

- [ ] **Step 2: Commit**

```bash
git add pkg/ui/tui.go
git commit -m "feat(ui): add up/down key handling for history navigation"
```

---

## Task 4: Save to history on Enter

**Files:**
- Modify: `pkg/ui/tui.go`

- [ ] **Step 1: Modify Enter handling**

当用户按 Enter 时，如果输入非空，保存到 history：

```go
// 在发送输入到 agent 之前
if userInput != "" && m.historyIndex == -1 {
    m.history = append(m.history, userInput)
    if len(m.history) > 100 {
        m.history = m.history[1:]
    }
    go saveHistory(m.history)  // 异步保存
}
```

Run: `go build ./pkg/ui/...`

- [ ] **Step 2: Commit**

```bash
git add pkg/ui/tui.go
git commit -m "feat(ui): save input to history on Enter"
```

---

## Task 5: Handle cluster switch

**Files:**
- Modify: `pkg/ui/tui.go`

- [ ] **Step 1: Clear input and reset historyIndex on cluster switch**

当处理 `/cluster <name>` 命令时：
1. 清空 textinput
2. 重置 historyIndex = -1

历史文件是共享的，不需要额外处理。

- [ ] **Step 2: Commit**

```bash
git add pkg/ui/tui.go
git commit -m "feat(ui): handle history on cluster switch"
```

---

## Task 6: Add /clear-history command

**Files:**
- Modify: `pkg/ui/tui.go`

- [ ] **Step 1: Add /clear-history handling**

```go
if userInput == "/clear-history" {
    m.history = []string{}
    os.Remove(expandPath(historyFile))
    m.textinput.Reset()
    m.historyIndex = -1
    m.tempInput = ""
    // 继续等待下一个输入
}
```

- [ ] **Step 2: Commit**

```bash
git add pkg/ui/tui.go
git commit -m "feat(ui): add /clear-history command"
```

---

## Task 7: Final verification

- [ ] **Step 1: Run build**

```bash
go build ./...
```

- [ ] **Step 2: Run tests**

```bash
go test ./...
```

- [ ] **Step 3: Commit**

```bash
git add -A
git commit -m "chore: final verification for input-history-navigation"
```