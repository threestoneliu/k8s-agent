# Operation Classification

k8s-agent classifies all operations into two categories: Query (read-only) and Mutation (potentially destructive). This classification determines whether an operation executes immediately or requires human confirmation.

## Classification Rules

### Query Operations

Query operations are read-only and execute immediately without confirmation.

| Verb | Description | Example |
|------|-------------|---------|
| `get` | Retrieve specific resource(s) | `get pods` |
| `list` | List all resources of a type | `list services` |
| `describe` | Show detailed resource information | `describe pod nginx` |
| `watch` | Watch for changes in real-time | `watch pods` |
| `logs` | Retrieve pod logs | `logs my-pod` |
| `exec` | Execute command in a pod | `exec my-pod -- ls` |

### Mutation Operations

Mutation operations modify resources and require explicit confirmation before execution.

| Verb | Description | Example |
|------|-------------|---------|
| `create` | Create new resources | `create pod nginx` |
| `update` | Update existing resources | `update deployment myapp` |
| `patch` | Partially update resources | `patch pod nginx` |
| `delete` | Delete resources | `delete pod nginx` |
| `scale` | Scale deployment replicas | `scale deployment myapp --replicas=5` |
| `cordon` | Mark node as unschedulable | `cordon node-1` |
| `uncordon` | Mark node as schedulable | `uncordon node-1` |
| `drain` | Drain node for maintenance | `drain node-1` |

### High-Risk Resources

Operations on these resources are always classified as Mutation, regardless of verb:

| Resource | Description |
|----------|-------------|
| `nodes` | Node-level operations can affect cluster stability |
| `persistentvolumes` | PV operations can cause data loss |
| `namespaces` | Namespace operations affect all resources within |
| `storageclasses` | Storage class changes affect persistent storage |

## Classification Algorithm

```go
func ClassifyVerb(verb, resource string) OperationType {
    // Query verbs on non-high-risk resources
    if queryVerbs[verb] && !highRiskResources[resource] {
        return OperationTypeQuery
    }

    // Mutation verbs OR high-risk resources
    if mutationVerbs[verb] || highRiskResources[resource] {
        return OperationTypeMutation
    }

    return OperationTypeUnknown
}
```

## Resource Mapping

The following aliases are recognized and mapped to their canonical Kubernetes API resource names:

| Alias | Canonical Resource |
|-------|-------------------|
| `pod` / `pods` | `pods` |
| `deployment` / `deployments` / `deploy` | `deployments` |
| `service` / `services` / `svc` / `svcs` | `services` |
| `namespace` / `namespaces` / `ns` | `namespaces` |
| `node` / `nodes` | `nodes` |
| `configmap` / `configmaps` / `cm` | `configmaps` |
| `secret` / `secrets` | `secrets` |
| `ingress` / `ingresses` / `ing` | `ingresses` |
| `pv` / `persistentvolume` / `persistentvolumes` | `persistentvolumes` |
| `pvc` / `persistentvolumeclaim` / `persistentvolumeclaims` | `persistentvolumeclaims` |
| `sc` / `storageclass` / `storageclasses` | `storageclasses` |
| `endpoint` / `endpoints` / `ep` | `endpoints` |
| `event` / `events` | `events` |
| `quota` / `resourcequota` / `resourcequotas` | `resourcequotas` |
| `limits` / `limitrange` / `limitranges` | `limitranges` |
| `hpa` / `horizontalpodautoscaler` / `horizontalpodautoscalers` | `horizontalpodautoscalers` |
| `pdb` / `poddisruptionbudget` / `poddisruptionbudgets` | `poddisruptionbudgets` |

## Confirmation Flow

### For Mutation Operations

1. User enters mutation command (e.g., `delete pod nginx`)
2. System parses and classifies as Mutation
3. System creates pending confirmation with unique key
4. System returns confirmation key to user
5. User approves with `k8s-agent confirm <key>`
6. System executes operation after approval

### Confirmation TTL

- Default TTL: 5 minutes
- Expired confirmations cannot be approved
- Background cleanup routine removes expired entries

## Examples

### Example 1: Query with Alias

**Input:** `get svc`
**Classification:** Query (get verb, services resource)
**Execution:** Immediate

1. Parser extracts: verb="get", resource="svc"
2. Mapper converts: "svc" → "services"
3. Classifier determines: Query (get on non-high-risk resource)
4. Executor runs: direct query against API

### Example 2: Mutation on High-Risk Resource

**Input:** `get nodes`
**Classification:** Mutation (high-risk resource)
**Execution:** Requires confirmation

1. Parser extracts: verb="get", resource="nodes"
2. Classifier sees "nodes" in highRiskResources
3. System creates confirmation, returns key
4. User must confirm before execution

### Example 3: Mutation with Standard Verb

**Input:** `delete deployment myapp`
**Classification:** Mutation (delete verb)
**Execution:** Requires confirmation

1. Parser extracts: verb="delete", resource="deployment", name="myapp"
2. Mapper converts: "deployment" → "deployments"
3. Classifier determines: Mutation (delete verb)
4. System creates confirmation, returns key

## Edge Cases

### Empty Input
Returns error: "input cannot be empty"

### Invalid Format
Returns error: "invalid input format" (requires at least verb + resource)

### Unknown Verb
Returns `OperationTypeUnknown` - currently executes directly (should prompt confirmation in future)

### Ambiguous Commands
For commands like `scale deployment myapp to 3`, the parser extracts:
- verb: "scale"
- resource: "deployment"
- name: "myapp"
- flags: {"replicas": "3"}

The classifier handles "scale" as a mutation verb requiring confirmation.