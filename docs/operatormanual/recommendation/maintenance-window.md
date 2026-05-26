---
title: Maintenance Window
menu:
  docs_{{ .version }}:
    identifier: maintenance-window
    name: Maintenance Window
    parent: recommendation
    weight: 50
menu_name: docs_{{ .version }}
section_menu_id: operatormanual
---

> New to KubeDB? Please start [here](/docs/README.md).

# Maintenance Window

A `MaintenanceWindow` is a Kubernetes resource that defines a time period when maintenance operations, such as recommended operations, can be executed in a namespace. This allows you to schedule automatic operations during preferred times when your infrastructure is idle or traffic is at the lowest.

## Overview

MaintenanceWindow enables automatic execution of [Recommendations](/docs/operatormanual/recommendation/recommendation-spec.md) without manual intervention. When a recommendation is created, the supervisor controller checks if a suitable MaintenanceWindow exists and executes the recommended operation within that window.

> Note: For cluster-wide maintenance windows, see [Cluster Maintenance Window](/docs/operatormanual/recommendation/cluster-maintenance-window.md).

## Spec

### Key Fields

**isDefault** (boolean, optional)
- Marks this MaintenanceWindow as the default for the namespace.
- When a recommendation doesn't specify a maintenance window, the default window is used automatically.
- Only one default MaintenanceWindow should exist per namespace.

**timezone** (string, optional)
- Specifies the timezone for interpreting the times and dates.
- If not set, empty, or "UTC", the given times and dates are considered as UTC.
- If set to "Local", the given times and dates are considered as server local timezone.
- Otherwise, should be a valid IANA Time Zone database location name.
- Examples: `Asia/Dhaka`, `America/New_York`, `Europe/London`
- Ref: https://www.iana.org/time-zones and https://en.wikipedia.org/wiki/List_of_tz_database_time_zones

**days** (map[DayOfWeek][]TimeWindow, optional)
- Consists of a map of day-of-week to corresponding list of time windows.
- There is a logical OR relationship between Days and Dates (operation can run on specified days OR dates).
- DayOfWeek values: `Sunday`, `Monday`, `Tuesday`, `Wednesday`, `Thursday`, `Friday`, `Saturday`
- Each TimeWindow contains Start and End times.

**dates** ([]DateWindow, optional)
- Consists of a list of specific dates when maintenance can occur.
- Dates must always be given in UTC format.
- Format: `yyyy-mm-ddThh:mm:ssZ` (where Z stands for UTC/GMT +0000)
- There is a logical OR relationship between Days and Dates.

### TimeWindow Structure

A TimeWindow specifies the start and end time within a day.

```yaml
start: "10:40AM"
end: "7:00PM"
```

Supported time formats: `HH:MMAM/PM` or `HH:MM` (24-hour format)

### DateWindow Structure

A DateWindow specifies the start and end timestamps for specific date ranges.

```yaml
start: 2025-01-25T00:00:18Z
end: 2025-01-25T23:41:18Z
```

## Status

### Key Fields

**status** (ApprovalStatus)
- Specifies the current status of the MaintenanceWindow.
- Possible values: `Pending`, `Approved`, `Rejected`
- Default: `Pending`

**conditions** ([]Condition)
- Conditions applied to the MaintenanceWindow.

**observedGeneration** (int64, optional)
- The most recent generation observed for this resource.

## Examples

### Daily Maintenance Window

Create a MaintenanceWindow that allows maintenance every day from 2:00 AM to 4:00 AM in Asia/Dhaka timezone:

```yaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: MaintenanceWindow
metadata:
  name: daily-maintenance
  namespace: default
spec:
  timezone: Asia/Dhaka
  days:
    Sunday:
      - start: "2:00AM"
        end: "4:00AM"
    Monday:
      - start: "2:00AM"
        end: "4:00AM"
    Tuesday:
      - start: "2:00AM"
        end: "4:00AM"
    Wednesday:
      - start: "2:00AM"
        end: "4:00AM"
    Thursday:
      - start: "2:00AM"
        end: "4:00AM"
    Friday:
      - start: "2:00AM"
        end: "4:00AM"
    Saturday:
      - start: "2:00AM"
        end: "4:00AM"
```

### Weekend Maintenance Window

Create a MaintenanceWindow for weekend maintenance only:

```yaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: MaintenanceWindow
metadata:
  name: weekend-maintenance
  namespace: default
spec:
  isDefault: true
  timezone: UTC
  days:
    Saturday:
      - start: "00:00"
        end: "06:00"
    Sunday:
      - start: "00:00"
        end: "06:00"
```

### Specific Dates Maintenance Window

Create a MaintenanceWindow for specific maintenance periods:

```yaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: MaintenanceWindow
metadata:
  name: scheduled-maintenance
  namespace: default
spec:
  timezone: UTC
  dates:
    - start: 2025-02-01T22:00:00Z
      end: 2025-02-02T04:00:00Z
    - start: 2025-03-01T22:00:00Z
      end: 2025-03-02T04:00:00Z
```

### Multiple Time Windows Per Day

Create a MaintenanceWindow with multiple time slots:

```yaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: MaintenanceWindow
metadata:
  name: multi-slot-maintenance
  namespace: default
spec:
  timezone: America/New_York
  days:
    Wednesday:
      - start: "2:00AM"
        end: "3:00AM"
      - start: "10:00PM"
        end: "11:00PM"
    Friday:
      - start: "2:00AM"
        end: "3:00AM"
      - start: "10:00PM"
        end: "11:00PM"
```

## Usage with Recommendations

Once a MaintenanceWindow is created, recommendations can use it in several ways:

1. **Default Window**: If marked as `isDefault: true`, recommendations automatically use it
2. **Specific Reference**: Recommendations can explicitly reference this window
3. **NextAvailable**: Recommendations can request execution in the next available window

See [Recommendation Spec](/docs/operatormanual/recommendation/recommendation-spec.md) for more details on how recommendations use MaintenanceWindows.

## Using with ApprovalPolicy

MaintenanceWindows are typically used with [ApprovalPolicy](/docs/operatormanual/recommendation/approval-policy.md) resources to automate recommendation execution:

```yaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: ApprovalPolicy
metadata:
  name: es-policy
  namespace: default
spec:
  maintenanceWindowRef:
    name: daily-maintenance
  targets:
    - group: kubedb.com
      kind: Elasticsearch
```

This policy will execute all Elasticsearch recommendations within the `daily-maintenance` window automatically.

## Best Practices

1. **Choose Off-Peak Hours**: Schedule maintenance during times with lowest traffic
2. **Set Adequate Duration**: Allow enough time for operations to complete
3. **Default Window**: Create a default MaintenanceWindow in each namespace
4. **Timezone Awareness**: Use appropriate timezone for your infrastructure
5. **Test Duration**: Monitor how long operations typically take and adjust window duration accordingly
6. **Explicit Approval**: Use `requireExplicitApproval` in recommendations that need manual review

## Comparison with ClusterMaintenanceWindow

| Aspect | MaintenanceWindow | ClusterMaintenanceWindow |
|--------|-------------------|-------------------------|
| Scope | Namespace-scoped | Cluster-wide |
| Priority | Higher (checked first) | Lower (fallback) |
| Use Case | Namespace-specific scheduling | Cluster-default scheduling |
