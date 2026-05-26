---
title: Cluster Maintenance Window
menu:
  docs_{{ .version }}:
    identifier: cluster-maintenance-window
    name: Cluster Maintenance Window
    parent: recommendation
    weight: 60
menu_name: docs_{{ .version }}
section_menu_id: operatormanual
---

> New to KubeDB? Please start [here](/docs/README.md).

# Cluster Maintenance Window

A `ClusterMaintenanceWindow` is a cluster-scoped Kubernetes resource that defines a cluster-wide time period when maintenance operations can be executed. Unlike the namespace-scoped [MaintenanceWindow](/docs/operatormanual/recommendation/maintenance-window.md), ClusterMaintenanceWindow applies to the entire cluster and serves as a fallback when namespace-specific windows are not available.

## Overview

ClusterMaintenanceWindow provides cluster-level scheduling for maintenance operations across all namespaces. It's useful for:

- Setting a cluster-wide default maintenance schedule
- Coordinating maintenance across multiple namespaces
- Providing a fallback window when namespace-specific windows don't exist
- Simplifying cluster operations management

> Note: MaintenanceWindow takes priority over ClusterMaintenanceWindow. If a namespace-scoped MaintenanceWindow is available, it will be used instead.

## Scope

ClusterMaintenanceWindow is a cluster-scoped resource, meaning:
- No namespace is required when creating or referencing it
- Applicable to all namespaces in the cluster (unless overridden by namespace-scoped MaintenanceWindow)
- Can be created and managed at the cluster level

## Spec

ClusterMaintenanceWindow uses the same spec structure as [MaintenanceWindow](/docs/operatormanual/recommendation/maintenance-window.md).

### Key Fields

**isDefault** (boolean, optional)
- Marks this ClusterMaintenanceWindow as the cluster-wide default.
- When a recommendation doesn't specify a maintenance window and no namespace-scoped MaintenanceWindow exists, this window is used.
- Only one default ClusterMaintenanceWindow should exist in the cluster.

**timezone** (string, optional)
- Specifies the timezone for interpreting the times and dates.
- If not set, empty, or "UTC", the given times and dates are considered as UTC.
- If set to "Local", the given times and dates are considered as server local timezone.
- Otherwise, should be a valid IANA Time Zone database location name.
- Examples: `Asia/Dhaka`, `America/New_York`, `Europe/London`

**days** (map[DayOfWeek][]TimeWindow, optional)
- Consists of a map of day-of-week to corresponding list of time windows.
- DayOfWeek values: `Sunday`, `Monday`, `Tuesday`, `Wednesday`, `Thursday`, `Friday`, `Saturday`
- Logical OR relationship between Days and Dates.

**dates** ([]DateWindow, optional)
- Consists of a list of specific dates when maintenance can occur.
- Must be given in UTC format.
- Format: `yyyy-mm-ddThh:mm:ssZ`
- Logical OR relationship between Days and Dates.

## Status

Same structure as [MaintenanceWindow](/docs/operatormanual/recommendation/maintenance-window.md):

**status** (ApprovalStatus)
- Possible values: `Pending`, `Approved`, `Rejected`

**conditions** ([]Condition)
- Applied conditions to the ClusterMaintenanceWindow.

**observedGeneration** (int64, optional)
- Most recent generation observed for this resource.

## Examples

### Cluster-Wide Default Window

Create a cluster-wide default maintenance window for all namespaces:

```yaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: ClusterMaintenanceWindow
metadata:
  name: cluster-default-maintenance
spec:
  isDefault: true
  timezone: UTC
  days:
    Sunday:
      - start: "00:00"
        end: "06:00"
    Saturday:
      - start: "00:00"
        end: "06:00"
```

### Cluster-Wide Weeknight Window

Create a maintenance window for weeknights across the cluster:

```yaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: ClusterMaintenanceWindow
metadata:
  name: cluster-weeknight-maintenance
spec:
  timezone: America/New_York
  days:
    Monday:
      - start: "2:00AM"
        end: "5:00AM"
    Tuesday:
      - start: "2:00AM"
        end: "5:00AM"
    Wednesday:
      - start: "2:00AM"
        end: "5:00AM"
    Thursday:
      - start: "2:00AM"
        end: "5:00AM"
    Friday:
      - start: "2:00AM"
        end: "5:00AM"
```

### Cluster-Wide Monthly Maintenance

Create a cluster-wide maintenance window for the first Sunday of each month:

```yaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: ClusterMaintenanceWindow
metadata:
  name: cluster-monthly-maintenance
spec:
  timezone: UTC
  dates:
    - start: 2025-02-02T20:00:00Z
      end: 2025-02-03T06:00:00Z
    - start: 2025-03-02T20:00:00Z
      end: 2025-03-03T06:00:00Z
    - start: 2025-04-06T20:00:00Z
      end: 2025-04-07T06:00:00Z
```

## Window Resolution Order

When a recommendation is created, the supervisor controller resolves the execution window in this order:

1. **Explicit Window in Recommendation**: If specified in `ApprovedWindow`
2. **Namespace-Scoped MaintenanceWindow**: If a default exists in the same namespace
3. **ClusterMaintenanceWindow**: If a default exists at cluster level
4. **Pending**: If no window is found, recommendation stays pending

Example flow:

```
Recommendation created
    ↓
Has explicit window? → Use it → Done
    ↓ No
Has namespace default MaintenanceWindow? → Use it → Done
    ↓ No
Has cluster default ClusterMaintenanceWindow? → Use it → Done
    ↓ No
Stay in Pending state
```

## Using with ApprovalPolicy

ClusterMaintenanceWindows can be referenced in [ApprovalPolicy](/docs/operatormanual/recommendation/approval-policy.md) for cluster-wide automation:

```yaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: ApprovalPolicy
metadata:
  name: cluster-db-policy
spec:
  maintenanceWindowRef:
    name: cluster-default-maintenance
  targets:
    - group: kubedb.com
      kind: Elasticsearch
    - group: kubedb.com
      kind: PostgreSQL
    - group: kubedb.com
      kind: MongoDB
```

This policy will automatically execute recommendations for all three database types during the cluster maintenance window.

## Best Practices

1. **Single Default**: Only one default ClusterMaintenanceWindow should exist
2. **Off-Peak Hours**: Schedule during lowest cluster activity
3. **Namespace Overrides**: Allow namespaces to override with their own MaintenanceWindows for specific needs
4. **Adequate Duration**: Ensure window is long enough for all expected operations
5. **Timezone Consistency**: Use UTC for cluster-wide windows to avoid timezone confusion
6. **Monitoring**: Monitor operation execution times to optimize window duration

## Comparison with MaintenanceWindow

| Aspect | MaintenanceWindow | ClusterMaintenanceWindow |
|--------|-------------------|-------------------------|
| Scope | Namespace | Cluster-wide |
| API Group | supervisor.appscode.com/v1alpha1 | supervisor.appscode.com/v1alpha1 |
| Default Required | Optional per namespace | One per cluster (optional) |
| Priority | Higher | Fallback |
| Metadata | Namespace-scoped | Cluster-scoped |
| Use Case | Namespace-specific scheduling | Cluster default scheduling |
| Recommendation Scope | Applies to recommendations in same namespace | Applies to all namespaces without their own window |

## Example: Multi-Level Maintenance Windows

For a production cluster with different requirements:

**Cluster Level (Default for all):**
```yaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: ClusterMaintenanceWindow
metadata:
  name: cluster-default
spec:
  isDefault: true
  timezone: UTC
  days:
    Sunday:
      - start: "00:00"
        end: "06:00"
```

**Production Namespace (Override for critical services):**
```yaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: MaintenanceWindow
metadata:
  name: prod-maintenance
  namespace: production
spec:
  isDefault: true
  timezone: UTC
  days:
    Sunday:
      - start: "02:00"
        end: "04:00"
```

**Development Namespace (Looser schedule):**
```yaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: MaintenanceWindow
metadata:
  name: dev-maintenance
  namespace: development
spec:
  isDefault: true
  timezone: UTC
  days:
    Wednesday:
      - start: "20:00"
        end: "23:00"
    Sunday:
      - start: "00:00"
        end: "06:00"
```

This setup allows the cluster to have a safe default while letting namespaces customize based on their needs.
