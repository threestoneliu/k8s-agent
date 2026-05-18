# Resource List Table Format Design

## Context

当前 `ListResourcesWithSelectors` 返回完整的 YAML/JSON 格式，当资源数量很大时会超过 LLM 上下文限制。需要改为 kubectl 表格风格输出，减少数据传输量。

## Goals / Non-Goals

**Goals:**
- kubectl 表格风格输出
- 支持 LabelSelector 和 FieldSelector
- 减少 LLM 数据量

**Non-Goals:**
- 不实现分页（Table 格式已足够紧凑）
- 不实现复杂列格式化
- 不实现列排序

## Decisions

### 1. 使用 RESTClient + Table Accept Header

使用 RESTClient 请求 k8s API，设置 Accept header 为 Table 格式：

```go
req.Header().Set("Accept", "application/json;as=Table;v=v1;g=meta.k8s.io")
```

API Server 直接返回表格结构，不需要客户端转换。

### 2. 支持现有 Selector

LabelSelector 和 FieldSelector 与 Accept header 独立，互不冲突：

```go
req := r.restClient.Get().
    Resource(gvr.Resource).
    Namespace(namespace).
    VersionedParams(&metav1.ListOptions{
        LabelSelector: labelSelector,
        FieldSelector: fieldSelector,
    }, metav1.ParameterCodec)

req.Header().Set("Accept", "application/json;as=Table;v=v1;g=meta.k8s.io")
```

### 3. Table 响应格式

```json
{
  "kind": "Table",
  "columnDefinitions": [
    {"name": "NAME", "type": "string"},
    {"name": "READY", "type": "string"},
    {"name": "STATUS", "type": "string"},
    {"name": "AGE", "type": "string"}
  ],
  "rows": [
    {"cells": ["my-pod-1", "1/1", "Running", "10d"]},
    {"cells": ["my-pod-2", "1/1", "Running", "5d"]}
  ]
}
```

### 4. Registry 添加 GetRESTClient

需要 Registry 添加获取 RESTClient 的方法：

```go
func (r *Registry) GetRESTClient(clusterName string) (*rest.RESTClient, error)
```

### 5. 格式化输出

```
NAME                     READY   STATUS    AGE
my-pod-1                1/1     Running   10d
my-pod-2                1/1     Running   5d

Showing 100 of 1523 items
```

## Architecture

```
User: "list pods with label app=nginx"
         ↓
LLM: parses intent, calls ListResourcesWithSelectors
         ↓
Executor: builds RESTClient request with Table Accept header
         ↓
K8s API: returns Table format directly
         ↓
Executor: parses Table, formats as text
         ↓
Output: formatted table (compact, LLM-friendly)
```

## Files to Modify

1. **pkg/cluster/registry.go** - Add `GetRESTClient()` method
2. **pkg/k8s/executor.go** - Add `ListResourcesAsTable()` method or modify existing
3. **pkg/k8s/types.go** - Add Table-related types if needed

## Implementation Steps

1. Add `GetRESTClient()` to Registry
2. Add `ListResourcesAsTable()` method to Executor
3. Use `metav1.Table` for parsing response
4. Format output as text table
5. Update LLM prompts to understand table output

## Risks / Trade-offs

| 风险 | 影响 | 缓解 |
|------|------|------|
| RESTClient 需要额外配置 | 增加复杂度 | 复用 cluster config |

**注：** 几乎所有 k8s 资源（包括 CRDs）都支持 Table 格式，不需要 fallback