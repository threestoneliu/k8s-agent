# Simplified K8s Module

## ADDED Requirements

### Requirement: K8s Module Name

The module SHALL be named `pkg/k8s/` (renamed from `pkg/engine/`).

#### Scenario: Module directory is pkg/k8s/

Given the codebase
When checking module path
Then k8s operations live in pkg/k8s/ not pkg/engine/

---

### Requirement: Core Operations

K8s executor SHALL support these operations:
- `resource_list`: List resources with namespace, label selector, field selector
- `resource_get`: Get specific resource (name, namespace for namespaced resources)
- `get_apiresources`: Discover available API resource types
- `use_cluster`: Switch current cluster context

#### Scenario: resource_list returns k8s resources

Given cluster has pods in "default" namespace
When calling resource_list with namespace="default", resource="pods"
Then list of pods is returned

#### Scenario: resource_get returns single resource

Given cluster has nginx pod in "default" namespace
When calling resource_get with name="nginx", namespace="default", resource="pods"
Then pod details are returned

#### Scenario: get_apiresources returns available types

When calling get_apiresources
Then list of available API resource types is returned

---

### Requirement: Simplified Resource Normalization

Resource name normalization SHALL be simplified, removing mapper.go complexity.

#### Scenario: "po" resolves to "pods"

Given resource name "po"
When normalizing for API call
Then it resolves to "pods"
And correct API endpoint is called

---

### Requirement: Cluster Registry Integration

K8s executor SHALL use cluster.Registry for client access.

#### Scenario: Executor uses Registry for client

Given cluster "dev" is registered
When executor performs operation on "dev" cluster
Then client is obtained from Registry
And correct kubeconfig context is used

---

## REMOVED Requirements

### Requirement: Complex Resource Mapper

The mapper.go resource normalization logic SHALL be simplified.

### Requirement: Parser and Classifier

parser.go and classifier.go SHALL be removed as separate files.

### Requirement: Executor Query Split

executor_query.go SHALL be merged into executor.go.

### Requirement: Scheduled Task Handler

createScheduledTaskHandler SHALL be removed (scheduler module deleted).

---

## Implementation Notes

**Files:**
- `pkg/k8s/executor.go` — Core k8s operations (merged from executor.go + executor_query.go)
- `pkg/k8s/registry.go` — Cluster client management
- `pkg/k8s/types.go` — Simple type definitions

**Removed:**
- mapper.go, parser.go, classifier.go, executor_query.go
- pkg/scheduler/ entire module

**Dependencies:**
- imports pkg/cluster/ for Registry