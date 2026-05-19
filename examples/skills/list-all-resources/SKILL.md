---
name: list-all-resources
description: List Kubernetes resources across all namespaces. Use when user wants to query resources of a specific type (like rolebindings, pods, services, configmaps) or all resource types across the entire cluster. Triggers when user says "查看所有"、"查询全部"、"across all namespaces"、"in all namespaces" or similar expressions indicating cluster-wide scope.
license: Apache-2.0
compatibility: k8s-agent
metadata:
  author: k8s-agent
  version: "1.0"
---

# List All Resources

## Workflows

### 列出所有资源 (List All Resources Workflow)

**使用场景**: 用户想要查看集群中指定类型或所有类型的资源。

**前置判断**: 用户请求涉及以下意图时触发：
- "查看所有 rolebinding"
- "查询所有 pods"
- "集群总览"
- "查询全部资源"
- "列出所有 xxx"
- "在所有 namespace 中查询"
- "across all namespaces"

**执行步骤**:

1. **确定要查询的资源类型**
   - 如果用户指定了具体资源类型（如 rolebindings, pods），使用该类型
   - 如果用户想查所有资源类型，调用 `get_apiresources()` 获取支持的所有 API 资源类型列表

2. **识别资源作用域**
   - 调用 `get_apiresources()` 查看资源的 `namespaced` 属性
   - cluster-scoped 资源：nodes, persistentvolumes, componentstatuses, namespaces, storageclasses, ingressclasses, rolebindings, clusterrolebindings 等
   - namespace-scoped 资源：pods, services, deployments, statefulsets, daemonsets, jobs, cronjobs, configmaps, secrets, ingresses, rolebindings 等

3. **查询 Cluster-scoped 资源**
   - 对于 cluster-scoped 资源，直接调用 `resource_list(resource="<resource>")`
   - 记录返回的表格结果

4. **查询 Namespace-scoped 资源**
   - 对于 namespace-scoped 资源：
     a. 先调用 `resource_list(resource="namespaces")` 获取所有 namespace 列表
     b. 解析 namespace 列表，提取 namespace 名称
     c. 对于每个 namespace，调用 `resource_list(resource="<resource>", namespace="<namespace>")`
     d. 如果资源数量很多，可以按 namespace 分组展示

5. **汇总输出**
   - 按类别分组展示结果：
     - Cluster-scoped 资源（节点、PV、StorageClass 等）
     - Namespace-scoped 资源（按 namespace 分组）

## 输出格式

```
## 集群资源总览

### Cluster-Scoped 资源

#### Nodes (<count>)
| NAME | STATUS | AGE |
|------|--------|-----|
| ... | ... | ... |

#### PersistentVolumes (<count>)
| NAME | CAPACITY | STATUS |
|------|----------|--------|
| ... | ... | ... |

### Namespace-Scoped 资源

#### Namespace: <namespace-name>

##### Pods (<count>)
| NAME | READY | STATUS | AGE |
|------|-------|--------|-----|
| ... | ... | ... | ... |

##### Services (<count>)
| NAME | TYPE | CLUSTER-IP | AGE |
|------|------|-----------|-----|
| ... | ... | ... | ... |

[... 其他资源类型 ...]
```

## 关键决策点

### 3.1 判断资源类型是否为 Cluster-scoped

资源类型的 scope 可通过以下方式判断：
- 调用 `get_apiresources()` 返回的资源列表中，某些资源的 `namespaced` 字段为 `false`
- 或者根据经验判断：
  - **通常是 Cluster-scoped**: `nodes`, `persistentvolumes`, `namespaces`, `storageclasses`, `ingressclasses`, `componentstatuses`, `nodes`, `persistentvolumeclaims`
  - **通常是 Namespace-scoped**: `pods`, `services`, `deployments`, `replicasets`, `statefulsets`, `daemonsets`, `jobs`, `cronjobs`, `configmaps`, `secrets`, `ingresses`, `endpoints`, `events`

### 3.2 大量 namespace 时的处理

如果 namespace 数量很多（如 > 50），遍历所有 namespace 可能会很慢：
- 可以限制只查询最活跃的 N 个 namespace（按资源数量排序）
- 或者提示用户指定具体的 namespace

### 3.3 错误处理

如果某个资源的查询失败（如权限不足）：
- 记录错误，继续查询其他资源
- 在输出中标记失败的资源：`[Error: <reason>]`

## 触发方式

- 显式: `/skill list-all-resources`
- 隐式: "查看所有 rolebinding", "查询所有 pods", "查看所有 configmap", "查询全部资源", "列出集群中所有资源", "集群总览", "查看整个集群的资源", "在所有 namespace 中查询"