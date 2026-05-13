# K8s Rename

## ADDED Requirements

### Requirement: Module Path Is pkg/k8s

The k8s execution module SHALL be located at `pkg/k8s/` (renamed from `pkg/engine/`).

#### Scenario: Module is at pkg/k8s

Given the codebase
When checking module path
Then k8s operations live in pkg/k8s/ not pkg/engine/

---

### Requirement: All Imports Updated

All imports of `pkg/engine/` SHALL be updated to `pkg/k8s/`.

#### Scenario: No imports of pkg/engine

Given all Go files in the codebase
When searching for `"k8s-agent/pkg/engine"`
Then no result is found

#### Scenario: pkg/k8s is imported correctly

Given all Go files that use k8s operations
When checking imports
Then they import `"k8s-agent/pkg/k8s"` not `"k8s-agent/pkg/engine"`

---

### Requirement: Executor Contains Core Operations

The `pkg/k8s/executor.go` SHALL contain: `resource_list`, `resource_get`, `get_apiresources`, `use_cluster` handlers.

#### Scenario: All required handlers exist

Given `pkg/k8s/executor.go`
When calling each handler: resource_list, resource_get, get_apiresources, use_cluster
Then each handler executes without error

---

### Requirement: Parser and Classifier Removed

The `parser.go` and `classifier.go` files SHALL be removed from k8s module.

#### Scenario: No parser.go or classifier.go

Given `pkg/k8s/` directory
When listing files
Then no `parser.go` or `classifier.go` exists

---

### Requirement: ExecutorQuery Merged

The `executor_query.go` logic SHALL be merged into `executor.go`.

#### Scenario: executor_query content is in executor.go

Given `pkg/k8s/executor.go`
When checking for query-related functions
Then all query logic is in executor.go with no executor_query.go reference
