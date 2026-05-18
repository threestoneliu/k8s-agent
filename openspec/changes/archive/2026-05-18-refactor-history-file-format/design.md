# Refactor History File Format

## Context

历史记录目前存储为 JSON 数组格式（`history.json`），用户反馈看起来不舒服。需要改为每行一条记录的行分割格式（`history.txt`）。

## Goals / Non-Goals

**Goals:**
- 改用更易读的行分割格式
- 简化加载/保存逻辑
- 支持追加模式（不必每次重写整个文件）
- 迁移旧 JSON 文件到新格式

**Non-Goals:**
- 不添加时间戳等元数据
- 不改变历史记录的功能逻辑

## Decisions

### 1. 文件格式

采用每行一条记录的行分割格式：

```
~/.config/k8s-agent/history/history.txt
---
kubectl get pods
kubectl describe pod my-pod
helm install my-chart
```

**为何**：简单易读、用户明确偏好

### 2. 文件名

从 `history.json` 改为 `history.txt`

**为何**：明确表示是纯文本格式，避免与 JSON 格式混淆

### 3. 加载逻辑

```go
func loadHistory() ([]string, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        if os.IsNotExist(err) {
            return []string{}, nil
        }
        return nil, err
    }
    lines := strings.Split(strings.TrimSpace(string(data)), "\n")
    // 过滤空行
    var history []string
    for _, line := range lines {
        if line != "" {
            history = append(history, line)
        }
    }
    return history, nil
}
```

### 4. 保存逻辑

追加模式：只追加新记录，不重写整个文件

```go
func appendHistory(entry string) error {
    f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        return err
    }
    defer f.Close()
    _, err = fmt.Fprintln(f, entry)
    return err
}
```

### 5. 迁移策略

首次运行时检测 `history.json` 是否存在：
- 存在则读取、转换为新格式、删除旧文件
- 不存在则直接使用新文件

```go
func migrateIfNeeded() error {
    jsonPath := expandPath("~/.config/k8s-agent/history/history.json")
    if _, err := os.Stat(jsonPath); err == nil {
        // 读取旧文件，转换，删除
        oldData, _ := os.ReadFile(jsonPath)
        var history []string
        json.Unmarshal(oldData, &history)
        // 写入新格式
        saveHistory(history)
        os.Remove(jsonPath)
    }
}
```

## Risks / Trade-offs

| 风险 | 影响 | 缓解 |
|------|------|------|
| 迁移失败 | 历史丢失 | 先备份再迁移 |
| 空行问题 | 解析异常 | 过滤空行 |

## Migration Plan

1. 部署新代码
2. 首次运行时检测旧 JSON 文件
3. 转换为新格式（每行一条）
4. 删除旧 JSON 文件
5. 验证历史记录正常