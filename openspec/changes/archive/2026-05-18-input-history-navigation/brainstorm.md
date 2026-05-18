## Design Summary

在 TUI 中添加输入历史导航功能，用户可通过上下键切换历史输入。

**Architecture:**
- 在 `Model` 中添加 `history []string` 和 `historyIndex int`
- 拦截上/下键操作 history
- 新输入按 Enter 后保存到 history
- 持久化到 `~/.config/k8s-agent/history`（每集群独立，限制 100 条）
- 切换集群时清空输入框

**Key Components:**
1. **History Management** — 内存中的 history slice 和 index 管理
2. **Persistence** — 启动时从文件加载，关闭时保存
3. **Key Handling** — 拦截上/下键，添加新命令 `/clear-history`

**Data Flow:**
```
User presses Up → Check historyIndex > 0 → Load history[historyIndex-1] into input → historyIndex--
User presses Down → Check historyIndex < len(history)-1 → Load history[historyIndex+1] into input → historyIndex++
User presses Enter → Save current input to history (if not empty) → Send to agent
```

## Alternatives Considered

### 方案 B：自定义 HistoryManager 类型
- **做法**：创建 `HistoryManager` 类型封装 history 逻辑
- **优点**：更容易测试
- **缺点**：增加复杂度，当前不需要
- **为何未采用**：方案 A 已足够满足需求

### 方案 C：独立 pkg/history 包
- **做法**：创建独立的 history 管理包
- **优点**：可复用
- **缺点**：过度设计，当前只需要 TUI 用
- **为何未采用**：YAGNI

## Agreed Approach

**方案 A：基于现有 textinput 的 history 支持**

- 利用 bubbletea 的 keybinding 机制拦截上/下键
- 在现有 `Model` 结构中添加 history 字段
- 最小化代码改动，不引入新包
- 文件持久化使用 JSON 格式
- 文件路径：`~/.config/k8s-agent/history/<cluster-name>.json`

## Key Decisions

1. **History 存储位置**：`~/.config/k8s-agent/history/<cluster-name>.json`
2. **每集群独立 history**：文件名包含集群名
3. **最大条数**：100 条（超出则删除最早的）
4. **切换集群清空输入框**：同时重置 historyIndex
5. **新命令**：`/clear-history` 清除当前集群的历史

## Open Questions

无