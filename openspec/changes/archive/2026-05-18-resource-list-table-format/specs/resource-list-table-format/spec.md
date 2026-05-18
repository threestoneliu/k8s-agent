## ADDED Requirements

### Requirement: resource-list-table-format

ListResourcesWithSelectors SHALL return kubectl-style table output instead of full YAML/JSON when listing Kubernetes resources. The response SHALL be formatted as a compact text table with columns: NAME, READY, STATUS, AGE by default.

#### Scenario: List pods returns table format
- **WHEN** User requests "list pods in default namespace"
- **THEN** Executor SHALL use RESTClient with Table Accept header to request Kubernetes API
- **AND** Response SHALL be formatted as text table with columns NAME, READY, STATUS, AGE
- **AND** Output SHALL be compact (typically under 100 bytes per resource)

#### Scenario: List with label selector returns filtered table
- **WHEN** User requests "list pods with label app=nginx"
- **THEN** Executor SHALL include LabelSelector in API request
- **AND** Table format SHALL be identical to unfiltered request

#### Scenario: List with field selector returns filtered table
- **WHEN** User requests "list pods field-selector status.phase=Running"
- **THEN** Executor SHALL include FieldSelector in API request
- **AND** Table format SHALL be identical to unfiltered request

#### Scenario: Both label and field selectors work together
- **WHEN** User requests "list pods with label app=nginx and field-selector status.phase=Running"
- **THEN** Executor SHALL include both LabelSelector and FieldSelector
- **AND** Results SHALL be filtered by both criteria

---