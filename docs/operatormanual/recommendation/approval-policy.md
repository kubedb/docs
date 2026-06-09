---
title: Approval Policy
menu:
  docs_{{ .version }}:
    identifier: approval-policy
    name: Approval Policy
    parent: recommendation
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: operatormanual
---


> New to KubeDB? Please start [here](/docs/README.md).

# Approval Policy

An `ApprovalPolicy` is a Kubernetes custom resource that **automatically approves Recommendations** for selected target resources and binds them to a maintenance window. It lets you say, in one place: *"For every Elasticsearch in this namespace, auto-approve recommendations and run them during the `daily-maintenance` window."*

> **Heads up — no `spec` field.** Unlike most Kubernetes resources, `ApprovalPolicy` puts all configuration **at the top level**. Use `maintenanceWindowRef` and `targets` directly, **not** under `spec:`.

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

`ApprovalPolicy` connects three concepts:

* **Recommendation** → *what* action should be executed.
* **Target resource** → *where* the action applies.
* **Maintenance window** → *when* the action is allowed.

It enables automatic approval of recommendations while keeping strict control over execution timing.

---

## Top-Level Fields

* `maintenanceWindowRef` (TypedObjectReference) — the maintenance window that will schedule execution of approved recommendations.
* `targets` ([]TargetRef, optional) — resources this policy applies to. If omitted, no recommendations are matched.

---

## maintenanceWindowRef Fields

* `maintenanceWindowRef.apiGroup` (string, optional) — API group of the referenced window. Defaults to `supervisor.appscode.com`.
* `maintenanceWindowRef.kind` (string) — `MaintenanceWindow` or `ClusterMaintenanceWindow`.
* `maintenanceWindowRef.name` (string) — name of the referenced window.
* `maintenanceWindowRef.namespace` (string, optional) — namespace of the window. Required for `MaintenanceWindow`, ignored for `ClusterMaintenanceWindow`.

---

## targets (TargetRef)

Each entry matches a database kind and, optionally, filters which operation types are auto-approved.

### TargetRef Fields

* `targets[].group` (string) — API group of the target resource (e.g. `kubedb.com`).
* `targets[].kind` (string) — kind of the target resource (e.g. `Elasticsearch`, `MongoDB`, `PostgreSQL`).
* `targets[].operations` ([]Operation, optional) — operation kinds eligible for auto-approval. If omitted, all operations for that target are eligible.

### Operation Fields

* `targets[].operations[].group` (string) — API group of the operation resource (e.g. `ops.kubedb.com`).
* `targets[].operations[].kind` (string) — kind of the operation resource (e.g. `ElasticsearchOpsRequest`, `MongoDBOpsRequest`).

---

## Behavior

* If `targets` is not specified → no recommendations are matched.
* If `operations` is not specified → all operations for that target are eligible.
* If the matching Recommendation has `spec.requireExplicitApproval: true` → the policy is ignored (a human must approve).

---

## Execution Flow

1. A `Recommendation` is created by the Ops-manager.
2. The Supervisor matches it against `ApprovalPolicy` entries using:
   * `targets[].group`
   * `targets[].kind`
   * `targets[].operations` (if specified)
3. On a match → recommendation is automatically approved.
4. `maintenanceWindowRef` is applied to schedule execution.
5. Execution:
   * Inside the window → runs immediately.
   * Outside the window → waits for the next allowed window.

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
  namespace: default
maintenanceWindowRef:
  kind: ClusterMaintenanceWindow
  name: cluster-default-maintenance
targets:
  - group: kubedb.com
    kind: Elasticsearch
```

See [Recommendation Spec](/docs/operatormanual/recommendation/recommendation-spec.md) for the full Recommendation field reference, and [Maintenance Window](/docs/operatormanual/recommendation/maintenance-window.md) for window scheduling details.
