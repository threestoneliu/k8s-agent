## Design Summary

使用 RESTClient + Table Accept Header 获取 k8s 资源列表，直接返回 kubectl 表格格式。

**方案：**
- RESTClient 请求 k8s API
- Accept header 设置为 Table 格式
- API Server 直接返回表格结构
- 支持 LabelSelector/FieldSelector
- Table 格式已足够紧凑，不需要分页

## Alternatives Considered

### 方案 A：DiscoveryClient + 自己转换（未采用）
- **做法**：用 DiscoveryClient 获取列定义，用 dynamic client 获取数据，自己转换
- **優點**：不改 Registry
- **缺點**：需要自己实现字段提取逻辑，复杂
- **為何未採用**：方案 B 更简单

### 方案 B：RESTClient + Table Accept（采用）
- **做法**：用 RESTClient 请求，设置 Accept header，API 直接返回 Table
- **優點**：API 直接返回表格，不需要自己转换
- **缺點**：需要修改 Registry 添加 GetRESTClient
- **為何採用**：实现更简单，代码更少

## Agreed Approach

采用 RESTClient + Table Accept Header 方案：
- 复用现有 selector 支持
- API 直接返回表格，不需要客户端转换
- Table 格式已足够紧凑，不需要分页

## Key Decisions

1. **RESTClient 获取**：从 Registry 获取 RESTClient，复用 cluster config
2. **Table Accept Header**：`"application/json;as=Table;v=v1;g=meta.k8s.io"`
3. **Selector 支持**：与 Accept header 独立，不冲突
4. **分页**：不需要，Table 格式已足够紧凑

## Open Questions

无