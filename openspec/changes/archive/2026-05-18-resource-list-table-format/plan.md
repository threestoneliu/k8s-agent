# Resource List Table Format Implementation Plan

> **For agentic workers:** Use superpowers:subagent-driven-development
> to implement this plan task-by-task.

**Goal:** Modify ListResourcesWithSelectors to return kubectl-style table format instead of full YAML/JSON.

**Architecture:** Use RESTClient with Table Accept header to request Kubernetes API. API Server returns Table format directly which is then formatted as compact text output.

**Tech Stack:** Go, Kubernetes client-go, RESTClient

---

## Task 1: Add GetRESTClient to Registry

- [ ] **Step 1:** Read pkg/cluster/registry.go to understand existing Registry structure
- [ ] **Step 2:** Add GetRESTClient(clusterName string) (*rest.RESTClient, error) method
- [ ] **Step 3:** Extract RESTClient from cluster config or cached client
- [ ] **Step 4:** Write unit test in pkg/cluster/registry_test.go
- [ ] **Step 5:** Run tests to verify

## Task 2: Add ListResourcesAsTable to Executor

- [ ] **Step 1:** Read pkg/k8s/executor.go to understand existing ListResourcesWithSelectors
- [ ] **Step 2:** Add ListResourcesAsTable method with Table Accept header
- [ ] **Step 3:** Use "application/json;as=Table;v=v1;g=meta.k8s.io" Accept header
- [ ] **Step 4:** Parse metav1.Table response with ColumnDefinitions and Rows
- [ ] **Step 5:** Return formatted string output
- [ ] **Step 6:** Write unit test in pkg/k8s/executor_test.go
- [ ] **Step 7:** Run tests to verify

## Task 3: Wire ListResourcesAsTable into ListResourcesWithSelectors

- [ ] **Step 1:** Modify ListResourcesWithSelectors to call ListResourcesAsTable internally
- [ ] **Step 2:** Preserve existing API signature for backward compatibility
- [ ] **Step 3:** Run full test suite to ensure no regressions

## Task 4: Integration Verification

- [ ] **Step 1:** Run `go build ./...` to verify compilation
- [ ] **Step 2:** Run `go test ./...` to verify all tests pass
- [ ] **Step 3:** Manual test with actual k8s cluster if available