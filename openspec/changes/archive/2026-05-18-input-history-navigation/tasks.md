## 1. 添加 history 相关字段到 Model

- [ ] 1.1 在 `Model` 结构体中添加 `history []string` 字段
- [ ] 1.2 在 `Model` 结构体中添加 `historyIndex int` 字段
- [ ] 1.3 在 `Model` 结构体中添加 `tempInput string` 字段（保存浏览历史时的当前输入）

## 2. 实现 history 加载和保存

- [ ] 2.1 创建 `loadHistory() ([]string, error)` 函数
- [ ] 2.2 创建 `saveHistory(history []string) error` 函数
- [ ] 2.3 在 TUI 初始化时调用 `loadHistory` 加载历史
- [ ] 2.4 在 TUI 关闭时调用 `saveHistory` 保存历史

## 3. 修改 key handling

- [ ] 3.1 在 `Update` 方法中拦截上/下键
- [ ] 3.2 实现上键逻辑：浏览上一条历史
- [ ] 3.3 实现下键逻辑：浏览下一条历史或恢复当前输入

## 4. 修改 Enter 处理

- [ ] 4.1 在 Enter 处理中添加保存历史逻辑
- [ ] 4.2 当 historyIndex != -1 时，不要重复保存

## 5. 处理集群切换

- [ ] 5.1 切换集群时清空 textinput
- [ ] 5.2 重置 historyIndex = -1

## 6. 添加 /clear-history 命令

- [ ] 6.1 在命令处理中添加 `/clear-history` 分支
- [ ] 6.2 删除历史文件并清空内存中的 history

## 7. 测试

- [ ] 7.1 运行 `go build ./pkg/ui/...` 确保编译通过
- [ ] 7.2 运行 `go test ./pkg/ui/...` 确保测试通过