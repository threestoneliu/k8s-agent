## Design Summary

重构 history.json 为每行一条记录的行分割格式（history.txt）。

**新格式：**
```
~/.config/k8s-agent/history/history.txt
---
kubectl get pods
kubectl describe pod my-pod
helm install my-chart
```

**实现方式：**
- 加载：`strings.Split(fileContent, "\n")`，过滤空行
- 保存：`strings.Join(history, "\n") + "\n"`
- 追加：直接 Append 到文件末尾

**迁移：** 首次加载时检测旧 JSON 文件，转换为新格式后删除。

## Alternatives Considered

### 方案 A：每行一条记录（采用）
- **做法**：纯文本文件，每行一条历史记录
- **優點**：易读、可直接编辑、解析简单、append 方便
- **缺點**：无法存储元数据（如时间戳）
- **為何未採用**：N/A

### 方案 B：带时间戳的 JSON
- **做法**：`{"entries": [{"text": "...", "timestamp": "..."}]}`
- **優點**：可扩展、支持元数据
- **缺點**：复杂、用户说看起来不舒服
- **為何未採用**：用户明确偏好简单行格式

## Agreed Approach

采用方案 A - 每行一条记录的行分割格式。

用户明确表示 JSON 数组看着不舒服，而纯文本每行一条记录更易读。简单直接的方案也符合 YAGNI 原则。

## Key Decisions

1. 文件名改为 `history.txt`（不再是 .json）
2. 使用行分割格式，每行一条记录
3. 追加模式：每次保存只 append 新记录，不重写整个文件
4. 迁移：首次运行时检测并转换旧 JSON 文件

## Open Questions

无