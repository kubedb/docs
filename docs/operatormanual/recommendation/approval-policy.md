---
title: Approval Policy
menu:
  docs_{{ .version }}:
    identifier: approval-policy
    name: Approval Policy
    parent: recommendation
    weight: 70
menu_name: docs_{{ .version }}
section_menu_id: operatormanual
---

> New to KubeDB? Please start [here](/docs/README.md).

# Approval Policy

An `ApprovalPolicy` is a Kubernetes resource that links a [MaintenanceWindow](/docs/operatormanual/recommendation/maintenance-window.md) or [ClusterMaintenanceWindow](/docs/operatormanual/recommendation/cluster-maintenance-window.md) to specific target resources and operations. It enables automatic execution of [Recommendations](/docs/operatormanual/recommendation/recommendation-spec.md) for matching resources during the specified maintenance window without requiring manual approval.

## Overview

ApprovalPolicy automates the recommendation execution workflow by:

1. Matching recommendations against target resources (database kind, group)
2. Associating the matched recommendations with a maintenance window
3. Automatically approving and executing recommendations during that window
4. Supporting fine-grained operation filtering

This eliminates the need for manual approval while maintaining control over when operations run.

## Spec

### Key Fields

**maintenanceWindowRef** (TypedObjectReference)
- Specifies the MaintenanceWindow or ClusterMaintenanceWindow reference for the ApprovalPolicy.
- The referenced window will be used to schedule the execution of matching recommendations.
- Recommendations will be executed in this maintenance window without manual approval.
- Structure:
  ```yaml
  maintenanceWindowRef:
    apiGroup: supervisor.appscode.com  # optional, defaults to supervisor.appscode.com
    kind: MaintenanceWindow             # or ClusterMaintenanceWindow
    name: my-maintenance-window
    namespace: default                  # only for MaintenanceWindow, not for ClusterMaintenanceWindow
  ```

**targets** ([]TargetRef)
- Specifies the list of resources for which the ApprovalPolicy will be effective.
- Supports filtering by resource Kind, Group, and specific Operations.
- Multiple targets can be specified for different resource types.

### TargetRef Structure

**Group** (string)
- The API group of the target resource.
- Examples: `kubedb.com`, `ops.kubedb.com`, `monitoring.coreos.com`

**Kind** (string)
- The Kind of the target resource.
- Examples: `Elasticsearch`, `PostgreSQL`, `MongoDB`, `MySQL`

**operations** ([]Operation, optional)
- Specifies which operations are allowed by this policy.
- If not provided, all operations for the target kind are allowed.
- Each operation is identified by Group and Kind.

### Operation Structure

**Group** (string)
- The API group of the operation.
- Examples: `ops.kubedb.com`, `batch.kubedb.com`

**Kind** (string)
- The Kind of the operation.
- Examples: `ElasticsearchOpsRequest`, `PostgreSQLOpsRequest`, `MongoDBOpsRequest`

## Status

ApprovalPolicy is a simple resource without extensive status tracking. The resource itself acts as the policy definition.

## Examples

### Simple Policy: All Elasticsearch Operations

Create a policy that automatically approves all recommendations for Elasticsearch databases:

```yaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: ApprovalPolicy
metadata:
  name: elasticsearch-policy
  namespace: default
spec:
  maintenanceWindowRef:
    kind: MaintenanceWindow
    name: daily-maintenance
  targets:
    - group: kubedb.com
      kind: Elasticsearch
```

This policy will automatically execute all Elasticsearch recommendations (version updates, TLS rotation, auth rotation, etc.) during the `daily-maintenance` window.

### Policy with Multiple Targets

Create a policy for multiple database types:

```yaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: ApprovalPolicy
metadata:
  name: multi-db-policy
  namespace: default
spec:
  maintenanceWindowRef:
    kind: MaintenanceWindow
    name: weekend-maintenance
  targets:
    - group: kubedb.com
      kind: Elasticsearch
    - group: kubedb.com
      kind: PostgreSQL
    - group: kubedb.com
      kind: MongoDB
    - group: kubedb.com
      kind: MySQL
```

### Policy with Specific Operations

Create a policy that only allows specific operation types:

```yaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: ApprovalPolicy
metadata:
  name: tls-rotation-policy
  namespace: production
spec:
  maintenanceWindowRef:
    kind: MaintenanceWindow
    name: prod-maintenance
  targets:
    - group: kubedb.com
      kind: Elasticsearch
      operations:
        - group: ops.kubedb.com
          kind: ElasticsearchOpsRequest
    - group: kubedb.com
      kind: PostgreSQL
      operations:
        - group: ops.kubedb.com
          kind: PostgreSQLOpsRequest
```

### Cluster-Wide Policy

Create a cluster-wide policy using ClusterMaintenanceWindow:

```yaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: ApprovalPolicy
metadata:
  name: cluster-wide-policy
spec:
  maintenanceWindowRef:
    kind: ClusterMaintenanceWindow
    name: cluster-default-maintenance
  targets:
    - group: kubedb.com
      kind: Elasticsearch
    - group: kubedb.com
      kind: PostgreSQL
    - group: kubedb.com
      kind: MongoDB
```

This policy applies to all namespaces in the cluster.

### Policy with Operation Filtering

Create a policy that allows only version update operations:

```yaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: ApprovalPolicy
metadata:
  name: version-update-policy
  namespace: default
spec:
  maintenanceWindowRef:
    kind: MaintenanceWindow
    name: upgrade-window
  targets:
    - group: kubedb.com
      kind: Elasticsearch
      operations:
        - group: ops.kubedb.com
          kind: ElasticsearchOpsRequest
        - group: ops.kubedb.com
          kind: ElasticsearchRestoreSession
    - group: kubedb.com
      kind: PostgreSQL
      operations:
        - group: ops.kubedb.com
          kind: PostgreSQLOpsRequest
```

## How ApprovalPolicy Works

When a [Recommendation](/docs/operatormanual/recommendation/recommendation-spec.md) is created:

1. **Matching**: Supervisor controller checks if any ApprovalPolicy matches the recommendation's target
2. **Window Selection**: If matched, the maintenance window from the policy is applied
3. **Auto-Approval**: The recommendation is automatically approved without manual intervention
4. **Execution**: The recommendation's operation is created and executed during the maintenance window

Example flow:

```
Recommendation created for Elasticsearch
    ↓
Find matching ApprovalPolicy for kubedb.com/Elasticsearch
    ↓
ApprovalPolicy found: elasticsearch-policy
    ↓
Get maintenance window: daily-maintenance
    ↓
Check if current time is in maintenance window
    ↓
If yes: Execute recommendation immediately
If no: Wait for next maintenance window
```

## Interaction with Recommendation

### Without ApprovalPolicy

```yaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: Recommendation
metadata:
  name: elastic-version-update
  namespace: default
spec:
  target:
    group: kubedb.com
    kind: Elasticsearch
    name: elastic-cluster
  # ... other fields ...
status:
  approvalStatus: Pending  # Needs manual approval
  phase: Pending           # Waiting for approval
```

### With Matching ApprovalPolicy

```yaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: Recommendation
metadata:
  name: elastic-version-update
  namespace: default
spec:
  target:
    group: kubedb.com
    kind: Elasticsearch
    name: elastic-cluster
  # ... other fields ...
status:
  approvalStatus: Approved  # Auto-approved by policy
  phase: Waiting            # Waiting for maintenance window
```

## Use Cases

### Development Cluster

Run all maintenance operations automatically on weekends:

```yaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: MaintenanceWindow
metadata:
  name: dev-maintenance
  namespace: default
spec:
  isDefault: true
  timezone: UTC
  days:
    Saturday:
      - start: "00:00"
        end: "23:59"
    Sunday:
      - start: "00:00"
        end: "23:59"
---
apiVersion: supervisor.appscode.com/v1alpha1
kind: ApprovalPolicy
metadata:
  name: dev-auto-approve
  namespace: default
spec:
  maintenanceWindowRef:
    kind: MaintenanceWindow
    name: dev-maintenance
  targets:
    - group: kubedb.com
      kind: Elasticsearch
    - group: kubedb.com
      kind: PostgreSQL
```

### Production Cluster

Restrict to critical operations during off-peak hours with explicit operation filtering:

```yaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: ClusterMaintenanceWindow
metadata:
  name: prod-critical-window
spec:
  isDefault: true
  timezone: UTC
  days:
    Sunday:
      - start: "02:00"
        end: "06:00"
---
apiVersion: supervisor.appscode.com/v1alpha1
kind: ApprovalPolicy
metadata:
  name: prod-critical-ops
spec:
  maintenanceWindowRef:
    kind: ClusterMaintenanceWindow
    name: prod-critical-window
  targets:
    - group: kubedb.com
      kind: Elasticsearch
      operations:
        - group: ops.kubedb.com
          kind: ElasticsearchOpsRequest
```

### Multi-Environment

Different policies per namespace:

```yaml
# Production namespace - conservative policy
apiVersion: supervisor.appscode.com/v1alpha1
kind: ApprovalPolicy
metadata:
  name: prod-policy
  namespace: production
spec:
  maintenanceWindowRef:
    kind: MaintenanceWindow
    name: prod-maintenance
  targets:
    - group: kubedb.com
      kind: PostgreSQL
---
# Staging namespace - moderate policy
apiVersion: supervisor.appscode.com/v1alpha1
kind: ApprovalPolicy
metadata:
  name: staging-policy
  namespace: staging
spec:
  maintenanceWindowRef:
    kind: MaintenanceWindow
    name: staging-maintenance
  targets:
    - group: kubedb.com
      kind: Elasticsearch
    - group: kubedb.com
      kind: PostgreSQL
---
# Development namespace - aggressive policy
apiVersion: supervisor.appscode.com/v1alpha1
kind: ApprovalPolicy
metadata:
  name: dev-policy
  namespace: development
spec:
  maintenanceWindowRef:
    kind: MaintenanceWindow
    name: dev-maintenance
  targets:
    - group: kubedb.com
      kind: Elasticsearch
    - group: kubedb.com
      kind: PostgreSQL
    - group: kubedb.com
      kind: MongoDB
    - group: kubedb.com
      kind: MySQL
```

## Best Practices

1. **Namespace Scoping**: Create policies per namespace for better control
2. **Operation Filtering**: Use operation filtering for critical resources
3. **Window Alignment**: Match policy windows with your maintenance schedule
4. **Testing**: Test policies in development before production deployment
5. **Documentation**: Document which policies apply to which resources
6. **Review**: Periodically review and update policies as needs change
7. **Backup Windows**: Create multiple policies for different scenarios
8. **Priority Management**: Use Parallelism settings in recommendations for execution control

## Troubleshooting

**Policy Not Applied**: Ensure the target Kind and Group match exactly
**Wrong Window Used**: Check if multiple ApprovalPolicies match; first match wins
**No Auto-Execution**: Verify the maintenance window is active and properly configured
**Operation Not Allowed**: Check if operation filtering is too restrictive

See [Recommendation Spec](/docs/operatormanual/recommendation/recommendation-spec.md) for debugging recommendation execution.
