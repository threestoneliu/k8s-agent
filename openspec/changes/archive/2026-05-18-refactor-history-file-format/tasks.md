## 1. 更新常量

- [ ] 1.1 将 `historyFile` 常量从 `history.json` 改为 `history.txt`

## 2. 重写 loadHistory 函数

- [ ] 2.1 修改 loadHistory 使用行分割格式读取
- [ ] 2.2 过滤空行

## 3. 重写 saveHistory 函数

- [ ] 3.1 修改 saveHistory 使用追加模式
- [ ] 3.2 使用 Append 模式写入新记录

## 4. 添加迁移函数

- [ ] 4.1 创建 migrateFromJSON 函数
- [ ] 4.2 在 newModel 中调用迁移函数

## 5. 测试

- [ ] 5.1 运行 `go build ./pkg/ui/...` 确保编译通过
- [ ] 5.2 手动测试历史记录功能