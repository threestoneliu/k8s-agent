## Why

当前 ListResources 返回完整的 YAML/JSON 格式，当资源数量很大时（如 1000+ pods）会超过 LLM 上下文限制。需要改为 kubectl 表格风格输出，大幅减少数据量。

## What Changes

- 使用 RESTClient + Table Accept Header 请求 k8s API
- API 直接返回 Table 格式，无需客户端转换
- 支持现有的 LabelSelector 和 FieldSelector
- 表格格式紧凑，大幅减少 LLM 数据传输量

## Capabilities

### New Capabilities
- `resource-list-table-format`: K8s 资源列表表格格式输出

## Impact

- **Affected:** `pkg/cluster/registry.go` - 添加 GetRESTClient() 方法
- **Affected:** `pkg/k8s/executor.go` - 修改 ListResourcesWithSelectors 使用 Table 格式