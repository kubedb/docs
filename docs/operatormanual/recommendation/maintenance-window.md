---
title: Maintenance Window
menu:
  docs_{{ .version }}:
    identifier: maintenance-window
    name: Maintenance Window
    parent: recommendation
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: operatormanual
---

> New to KubeDB? Please start [here](/docs/README.md).

# Maintenance Window

A `MaintenanceWindow` defines **when** an approved `Recommendation` is allowed to execute inside a namespace. It ensures that potentially disruptive `OpsRequest`s — version upgrades, TLS rotations, auth secret rotations — only run inside predefined time ranges, so they never collide with peak usage.

Pair a `MaintenanceWindow` with an [ApprovalPolicy](/docs/operatormanual/recommendation/approval-policy.md) to fully automate routine maintenance.

## When to use

- You want recommended OpsRequests to run only during off-peak hours.
- Your team has a fixed change-management window (e.g. Sunday 2–4 AM).
- You need per-namespace scheduling control (different teams, different windows).

For a single cluster-wide default schedule, use a [Cluster Maintenance Window](/docs/operatormanual/recommendation/cluster-maintenance-window.md) instead.

---

## Example

```yaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: MaintenanceWindow
metadata:
  name: daily-maintenance
  namespace: default
spec:
  isDefault: true
  timezone: Asia/Dhaka
  days:
    Sunday:
      - start: "2:00AM"
        end: "4:00AM"
    Monday:
      - start: "2:00AM"
        end: "4:00AM"
status:
  status: Pending
```

---

## Spec Fields

* `spec.isDefault` (bool, optional) — marks this window as the default for the namespace. Used automatically when a recommendation does not explicitly reference a window.
* `spec.timezone` (*string, optional) — timezone used to interpret time values.
* `spec.days` (map[DayOfWeek][]TimeWindow, optional) — recurring weekly windows.
* `spec.dates` ([]DateWindow, optional) — specific date-based windows.

---

## DayOfWeek Values

* `Sunday`
* `Monday`
* `Tuesday`
* `Wednesday`
* `Thursday`
* `Friday`
* `Saturday`

---

## TimeWindow Fields

* `spec.days.<day>[].start` (TimeOfDay) — start time of the window.
* `spec.days.<day>[].end` (TimeOfDay) — end time of the window.

Example:

```yaml
days:
  Friday:
    - start: "10:00PM"
      end: "11:00PM"
```

---

## DateWindow Fields

* `spec.dates[].start` (Time) — start timestamp (UTC).
* `spec.dates[].end` (Time) — end timestamp (UTC).

Example:

```yaml
dates:
  - start: 2025-02-01T22:00:00Z
    end: 2025-02-02T04:00:00Z
```

---

## Behavior

* `spec.days` and `spec.dates` are combined with logical **OR** — execution is allowed if either condition matches.
* If both are empty → no execution window exists and approved recommendations stay in `Waiting`.
* If `spec.isDefault: true` → the window is used automatically when a Recommendation does not specify one.
* Timezone interpretation:
  * `""` or `UTC` → treated as UTC.
  * `Local` → the server's local timezone.
  * Otherwise → must be a valid IANA timezone (e.g. `Asia/Dhaka`, `America/New_York`).

---

## Status Fields

* `status.status` (ApprovalStatus) — current state (`Pending`, `Approved`, `Rejected`).
* `status.observedGeneration` (int64, optional) — latest observed spec generation.
* `status.conditions` ([]Condition, optional) — applied conditions.

---

## Execution Flow

When a `Recommendation` is processed:

1. The Supervisor selects a MaintenanceWindow:
   * Uses an explicitly referenced window if provided.
   * Otherwise uses the namespace default (`spec.isDefault: true`).
   * Otherwise falls back to a cluster-wide default ([ClusterMaintenanceWindow](/docs/operatormanual/recommendation/cluster-maintenance-window.md)).
2. Time evaluation:
   * If the current time is inside the window → execute immediately.
   * If outside → wait until the next valid window.
3. Execution:
   * The OpsRequest is created and runs only within the allowed time range.

---

## Usage with ApprovalPolicy

A `MaintenanceWindow` controls **when** an action runs; an `ApprovalPolicy` controls **what gets auto-approved**. Use them together:

```yaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: ApprovalPolicy
metadata:
  name: es-policy
  namespace: default
maintenanceWindowRef:
  kind: MaintenanceWindow
  name: daily-maintenance
targets:
  - group: kubedb.com
    kind: Elasticsearch
```

* `ApprovalPolicy` → handles automatic approval.
* `MaintenanceWindow` → controls execution timing.

See [Approval Policy](/docs/operatormanual/recommendation/approval-policy.md) for the full reference.
