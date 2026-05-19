---
name: k8s-inspection
description: K8s cluster inspection workflow. Use when user wants to inspect cluster health, check node status, or perform routine maintenance checks.
license: Apache-2.0
compatibility: k8s-agent
metadata:
  author: k8s-agent
  version: "1.0"
---

# K8s Inspection

## Workflows

### 巡检流程 (Inspection Workflow)

1. **检查节点状态**：调用 `resource_list(resource="nodes")`
2. **分析节点状态**：确认所有节点都是 Ready 状态
3. **检查集群所有分区下的 Pod 状态**：使用 list-all-resources skill 获取所有 namespace 的 Pod 列表
4. **分析 Pod 状态**：识别 Error/Pending/CrashLoopBackOff 状态的 Pod
5. **检查 Events**：调用 `resource_list(resource="events")`
6. **生成巡检报告**：汇总节点、Pod、Events 的状态，输出巡检结论

## 输出格式

巡检完成后，输出以下格式的摘要：

```
## 巡检报告

### 节点状态
- 总节点数: X
- Ready: Y
- NotReady: Z

### Pod 状态
- 总 Pods: X
- Running: Y
- Error: Z
- Pending: W

### 关键事件
- [列出最近的关键事件]

### 结论
[给出巡检结论和建议]
```

## 触发方式

- 显式: `/skill k8s-inspection` 或 `用巡检skill`
- 隐式: "巡检一下", "检查集群健康", "cluster inspection"