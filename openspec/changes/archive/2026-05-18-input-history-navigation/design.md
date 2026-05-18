# Input History Navigation Design

## Context

TUI 目前使用 `charmbracelet/bubbletea` 的 `textinput` 组件处理用户输入。用户输入历史需要通过上下键进行切换。目前没有历史记录功能。

## Goals / Non-Goals

**Goals:**
- 用户可以通过上/下键浏览历史输入
- 历史持久化到文件
- 所有集群共享同一份历史
- 支持 `/clear-history` 命令

**Non-Goals:**
- 不实现 bash/zsh 那样的复杂历史搜索（如 Ctrl+R）
- 不实现按集群独立的历史

## Decisions

### 1. History 存储格式

```json
// ~/.config/k8s-agent/history/history.json
["first command", "second command", ...]
```

**为何**：JSON 格式可读性好，易于调试，且 Go 标准库支持良好。

### 2. 内存数据结构

```go
type Model struct {
    // ... existing fields ...
    history         []string  // 历史记录
    historyIndex    int       // 当前浏览位置，-1 表示不在浏览历史
}
```

**为何**：直接在 Model 中管理，避免引入新类型导致复杂度增加。

### 3. Key Handling 逻辑

```
上键按下：
1. 如果 historyIndex == -1（不在浏览历史），保存当前输入到 temp
2. 如果 historyIndex > 0，historyIndex--
3. 加载 history[historyIndex] 到 textinput
4. 将 cursor 移到末尾

下键按下：
1. 如果 historyIndex == -1，不处理
2. 如果 historyIndex < len(history)-1，historyIndex++
3. 加载 history[historyIndex] 到 textinput
4. 如果 historyIndex 已到末尾，恢复 temp（如果有）

Enter 按下：
1. 如果 historyIndex != -1，说明用户在浏览历史
2. 保存当前输入到 history（如果非空）
3. 重置 historyIndex = -1
4. 正常发送输入到 agent
```

### 4. 持久化时机

- **启动时**：从文件加载 history 到内存
- **保存时**：每次用户按 Enter 且有有效输入时，更新内存并写入文件

**为何**：实时保存确保 Ctrl+C 等异常退出时不会丢失太多历史。

### 5. 切换集群处理

当用户执行 `/cluster <name>` 命令时：
1. 清空 textinput
2. 重置 historyIndex = -1

历史文件是共享的，不需要处理切换。

## File Structure

```
pkg/ui/
├── tui.go          # 修改：添加 history 字段和 key handling
├── tui_test.go     # 修改：添加 history 相关测试
└── ui.go           # 无变更
```

## Data Flow

```
┌─────────────────────────────────────────┐
│  User presses Up                        │
│  ↓                                      │
│  Check historyIndex > 0                 │
│  ↓                                      │
│  historyIndex--                         │
│  ↓                                      │
│  Load history[historyIndex]             │
│  ↓                                      │
│  Update textinput.Value()               │
│  ↓                                      │
│  Update viewport if needed              │
└─────────────────────────────────────────┘
```

## Risks / Trade-offs

| 风险 | 影响 | 缓解 |
|------|------|------|
| 历史文件损坏 | 启动失败 | 启动时检测文件完整性，损坏则忽略 |
| 历史过大 | 内存占用 | 限制 100 条，超出删除最旧的 |