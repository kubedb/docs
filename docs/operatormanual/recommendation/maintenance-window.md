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

A `MaintenanceWindow` defines when a `Recommendation` is allowed to execute. It ensures that `opsrequest` run only during predefined time ranges, preventing disruption during peak usage periods.

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

* `spec.isDefault` (bool | optional) – marks this window as the default for the namespace.

* `spec.timezone` (*string | optional) – specifies timezone for interpreting time values.

* `spec.days` (map[DayOfWeek][]TimeWindow | optional) – defines recurring weekly maintenance windows.

* `spec.dates` ([]DateWindow | optional) – defines specific date-based maintenance windows.

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

* `spec.days.<day>.start` (TimeOfDay) – start time of the maintenance window.

* `spec.days.<day>.end` (TimeOfDay) – end time of the maintenance window.

Example:

```yaml
days:
  Friday:
    - start: "10:00PM"
      end: "11:00PM"
```

---

## DateWindow Fields

* `spec.dates[].start` (Time) – start timestamp (UTC).

* `spec.dates[].end` (Time) – end timestamp (UTC).

Example:

```yaml
dates:
  - start: 2025-02-01T22:00:00Z
    end: 2025-02-02T04:00:00Z
```

---

## Behavior

* `spec.days` and `spec.dates` follow a logical OR relationship

  → execution is allowed if either condition matches

* If both are empty  no execution window exists

* If `spec.isDefault=true` automatically used when a Recommendation does not specify a window

* Time interpretation:

  * `""` or `UTC` → treated as UTC
  * `Local` → server local timezone
  * otherwise → must be valid IANA timezone (e.g., `Asia/Dhaka`)

---

## Status Fields

* `status.status` (ApprovalStatus) – current state (`Pending`, `Approved`, `Rejected`)

* `status.observedGeneration` (int64 | optional) – latest observed generation

* `status.conditions` ([]Condition | optional) – applied conditions

---

## Execution Flow

When a `Recommendation` is processed:

1. Controller selects a MaintenanceWindow:

   * Uses explicitly referenced window if provided
   * Otherwise uses default window (`spec.isDefault: true`)

2. Time evaluation:

   * If current time is within window → execute immediately
   * If outside → wait until next valid window

3. Execution:

   * Operation runs only within allowed time range



## Usage with ApprovalPolicy

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

* `ApprovalPolicy` → handles automatic approval
* `MaintenanceWindow` → controls execution timing

