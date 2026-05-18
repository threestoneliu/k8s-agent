## 1. Add GetRESTClient to Registry

- [x] 1.1 Add GetRESTClient(clusterName string) (*rest.RESTClient, error) method to Registry in pkg/cluster/registry.go
- [x] 1.2 Return the RESTClient from the cached discovery client or build from cluster config

## 2. Add ListResourcesAsTable method to Executor

- [x] 2.1 Add ListResourcesAsTable method to Executor in pkg/k8s/executor.go
- [x] 2.2 Use RESTClient with Table Accept header: "application/json;as=Table;v=v1;g=meta.k8s.io"
- [x] 2.3 Support LabelSelector and FieldSelector parameters
- [x] 2.4 Parse metav1.Table response with ColumnDefinitions and Rows

## 3. Format Table output as text

- [x] 3.1 Format ColumnDefinitions as header row
- [x] 3.2 Format Rows as aligned text table
- [x] 3.3 Return compact string output suitable for LLM consumption

## 4. Wire up from ListResourcesWithSelectors

- [x] 4.1 Modify ListResourcesWithSelectors to call ListResourcesAsTable internally
- [x] 4.2 Preserve existing API signature for backward compatibility

## 5. Testing

- [x] 5.1 Write unit test for GetRESTClient
- [x] 5.2 Write unit test for ListResourcesAsTable formatting
- [x] 5.3 Run existing tests to ensure no regressions