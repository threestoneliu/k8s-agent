## Why

当前 `history.json` 使用 JSON 数组格式，用户反馈看起来不舒服。每行一条记录的行分割格式更易读、更易调试。

## What Changes

- 将 `history.json` 改为 `history.txt`
- 从 JSON 数组格式改为每行一条记录
- 使用追加模式保存（不必重写整个文件）
- 添加迁移逻辑：首次运行时转换旧 JSON 文件

## Capabilities

### New Capabilities
- `history-file-format`: 历史文件格式重构

## Impact

- **Affected:** `pkg/ui/tui.go` - loadHistory/saveHistory 函数
- **Migration:** 首次运行自动转换旧 JSON 文件