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

An `ApprovalPolicy` is a Kubernetes custom resource that defines **automatic approval rules for Recommendations** based on target resources and a specified maintenance window.

Unlike most Kubernetes resources, `ApprovalPolicy` does **not use a `spec` field**. All configuration is defined at the **top level**.

---

## Example ApprovalPolicy

```yaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: ApprovalPolicy
metadata:
  name: elasticsearch-policy
  namespace: default
maintenanceWindowRef:
  kind: MaintenanceWindow
  name: daily-maintenance
targets:
  - group: kubedb.com
    kind: Elasticsearch
```

---

## Overview

`ApprovalPolicy` connects three key components:

* **Recommendation** → what action should be executed
* **Target Resource** → where the action applies
* **Maintenance Window** → when the action is allowed

It enables automatic approval of recommendations while maintaining strict control over execution timing.

---

## Top-Level Fields

* `maintenanceWindowRef` (TypedObjectReference) – specifies the maintenance window used to schedule execution of approved recommendations.

* `targets` ([]TargetRef | optional) – specifies the list of resources for which this policy will apply.

---

## maintenanceWindowRef Fields

* `maintenanceWindowRef.apiGroup` (string | optional) – specifies the API group of the referenced window. Defaults to `supervisor.appscode.com`.

* `maintenanceWindowRef.kind` (string) – specifies the type of window resource (`MaintenanceWindow` or `ClusterMaintenanceWindow`).

* `maintenanceWindowRef.name` (string) – specifies the name of the referenced maintenance window.

* `maintenanceWindowRef.namespace` (string | optional) – specifies the namespace of the window (required for `MaintenanceWindow`, not used for `ClusterMaintenanceWindow`).

---

## targets (TargetRef)

Each entry defines a rule for matching resources and optionally filtering operations.

---

### TargetRef Fields

* `targets[].group` (string) – specifies the API group of the target resource (e.g., `kubedb.com`).

* `targets[].kind` (string) – specifies the kind of the target resource (e.g., `Elasticsearch`, `PostgreSQL`).

* `targets[].operations` ([]Operation | optional) – specifies allowed operation types for auto-approval.

---

## operations (Operation)

Defines which operation types are eligible for automatic approval.

---

### Operation Fields

* `targets[].operations[].group` (string) – specifies the API group of the operation resource (e.g., `ops.kubedb.com`).

* `targets[].operations[].kind` (string) – specifies the kind of the operation resource (e.g., `ElasticsearchOpsRequest`).

---

## Behavior

* If `targets` is not specified → no recommendations are matched
* If `operations` is not specified → all operations for that target are allowed
* If `requireExplicitApproval: true` in Recommendation → ApprovalPolicy is ignored

---

## Execution Flow

1. A `Recommendation` is created
2. Controller matches it with `ApprovalPolicy` using:

   * `targets[].group`
   * `targets[].kind`
   * `targets[].operations` (if defined)
3. If matched → automatically approved
4. `maintenanceWindowRef` is applied
5. Execution:

   * Inside window → runs immediately
   * Outside window → waits for next window

---

## Examples

### Multiple Targets

```yaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: ApprovalPolicy
metadata:
  name: multi-db-policy
  namespace: default
maintenanceWindowRef:
  kind: MaintenanceWindow
  name: weekend-maintenance
targets:
  - group: kubedb.com
    kind: Elasticsearch
  - group: kubedb.com
    kind: PostgreSQL
```

---

### Operation Filtering

```yaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: ApprovalPolicy
metadata:
  name: tls-policy
  namespace: default
maintenanceWindowRef:
  kind: MaintenanceWindow
  name: prod-maintenance
targets:
  - group: kubedb.com
    kind: Elasticsearch
    operations:
      - group: ops.kubedb.com
        kind: ElasticsearchOpsRequest
```

---

### Cluster-Wide Policy

```yaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: ApprovalPolicy
metadata:
  name: cluster-policy
maintenanceWindowRef:
  kind: ClusterMaintenanceWindow
  name: cluster-default-maintenance
targets:
  - group: kubedb.com
    kind: Elasticsearch
```

See [Recommendation Spec](/docs/operatormanual/recommendation/recommendation-spec.md) for more details.
