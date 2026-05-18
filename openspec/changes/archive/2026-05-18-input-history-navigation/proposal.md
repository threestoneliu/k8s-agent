## Why

用户在使用 TUI 时经常需要重复执行相似的命令（如 `kubectl get pods`）。目前没有历史记录功能，用户需要手动重新输入。添加上下键导航历史可以显著提升用户体验。

## What Changes

**New Feature: Input History Navigation**

- 在 TUI 中添加 history 字段，存储用户输入历史
- 拦截上/下键，用户可以浏览历史输入
- 新输入保存到 history
- 持久化到 `~/.config/k8s-agent/history/history.json`
- 所有集群共享历史
- 限制 100 条，超出删除最旧的
- 添加 `/clear-history` 命令清除历史

## Capabilities

- `input-history-navigation`: TUI 输入历史导航，支持上下键切换，持久化存储

## Impact

**受影响代码：**
- `pkg/ui/tui.go`: 添加 history 字段，修改 key handling
- `pkg/ui/tui_test.go`: 添加 history 相关测试

**无影响：**
- 其他包无变更
- API 无破坏性变更