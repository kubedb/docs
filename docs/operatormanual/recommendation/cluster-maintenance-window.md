---
title: Cluster Maintenance Window
menu:
  docs_{{ .version }}:
    identifier: cluster-maintenance-window
    name: Cluster Maintenance Window
    parent: recommendation
    weight: 50
menu_name: docs_{{ .version }}
section_menu_id: operatormanual
---

> New to KubeDB? Please start [here](/docs/README.md).

# Cluster Maintenance Window

A `ClusterMaintenanceWindow` is a **cluster-scoped** custom resource that defines **when recommendations are allowed to execute across the entire cluster**.

> **Use this only when you want a single cluster-wide default.** For per-namespace control, prefer a namespace-scoped [Maintenance Window](/docs/operatormanual/recommendation/maintenance-window.md) — it always takes priority over the cluster default.

Unlike `MaintenanceWindow`, this resource applies globally and acts as a **fallback** when no namespace-specific window is available.

---

## Overview

`ClusterMaintenanceWindow` provides cluster-wide scheduling control for recommendation execution. It works together with:

* `Recommendation` → defines **what to execute**.
* `ApprovalPolicy` → defines **what gets approved**.
* `ClusterMaintenanceWindow` → defines **when execution happens (cluster-wide)**.

---

## Example

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

---

## Scope

* Cluster-scoped resource (no namespace required).
* Applies to all namespaces.
* Used only when a namespace-level `MaintenanceWindow` is not available.

---

## Spec Fields

### `spec.isDefault` (bool, optional)

Marks this as the cluster default window.

* Used when no namespace-level default exists.
* Only one cluster default should exist.

---

### `spec.timezone` (*string, optional)

Specifies the timezone used for evaluation.

Supported values:

* `"UTC"` or empty → treated as UTC.
* `"Local"` → server local timezone.
* IANA timezone → e.g. `Asia/Dhaka`, `America/New_York`.

---

### `spec.days` (map[DayOfWeek][]TimeWindow, optional)

Defines recurring windows by day.

```yaml
spec:
  days:
    Monday:
      - start: "2:00AM"
        end: "5:00AM"
```

* Supports multiple windows per day.
* Logical **OR** with `spec.dates`.

---

### `spec.dates` ([]DateWindow, optional)

Defines specific date-based windows.

```yaml
spec:
  dates:
    - start: 2025-02-01T22:00:00Z
      end: 2025-02-02T04:00:00Z
```

* Must be in UTC format.
* Logical **OR** with `spec.days`.

---

## TimeWindow

Defines a time range within a day.

```yaml
start: "2:00AM"
end: "5:00AM"
```

* `start` (TimeOfDay) — start time.
* `end` (TimeOfDay) — end time.

---

## DateWindow

Defines a full timestamp-based window.

```yaml
start: 2025-02-01T22:00:00Z
end: 2025-02-02T04:00:00Z
```

* `start` (metav1.Time) — start timestamp.
* `end` (metav1.Time) — end timestamp.

---

## Status Fields

* `status.status` (ApprovalStatus) — current state (`Pending`, `Approved`, `Rejected`).
* `status.conditions` ([]Condition, optional) — additional state details.
* `status.observedGeneration` (int64, optional) — latest observed generation.

---

## Window Resolution Order

When a `Recommendation` is created, the Supervisor determines the execution window in this order:

1. Explicit window from the Recommendation (`status.approvedWindow`).
2. Namespace default `MaintenanceWindow`.
3. Cluster default `ClusterMaintenanceWindow`.
4. If none found → remains in `Pending`.

---

## Usage with ApprovalPolicy

```yaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: ApprovalPolicy
metadata:
  name: cluster-db-policy
maintenanceWindowRef:
  kind: ClusterMaintenanceWindow
  name: cluster-default-maintenance
targets:
  - group: kubedb.com
    kind: Elasticsearch
```

### Behavior

* `ApprovalPolicy` → auto-approves the recommendation.
* `ClusterMaintenanceWindow` → controls execution timing.

---

## Best Practices

* Keep only one `spec.isDefault: true` per cluster.
* Schedule windows during off-peak hours.
* Prefer `UTC` for consistency across regions.
* Make windows long enough to comfortably finish typical operations.
* Allow namespace `MaintenanceWindow`s to override the cluster default where teams need finer control.

---

## Comparison with MaintenanceWindow

| Aspect             | MaintenanceWindow    | ClusterMaintenanceWindow |
| ------------------ | -------------------- | ------------------------ |
| Scope              | Namespace            | Cluster                  |
| Priority           | Higher               | Fallback                 |
| Namespace Required | Yes                  | No                       |
| Use Case           | Fine-grained control | Cluster default          |

---

## Example: Multi-Level Setup

**Cluster Default**

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

**Namespace Override**

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

This ensures:

* The cluster has a safe default everywhere.
* Critical namespaces can override the default with tighter windows.
