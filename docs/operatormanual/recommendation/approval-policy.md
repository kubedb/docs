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

## Overview

`ApprovalPolicy` connects three key components:

* **Recommendation** → what action should be executed
* **Target Resource** → where the action applies
* **Maintenance Window** → when the action is allowed

It enables:

* Automatic approval of recommendations
* Controlled execution within maintenance windows
* Optional filtering based on operation types

This eliminates the need for manual approval while still maintaining strict control over **what runs and when it runs**.

---

## Fields

### `maintenanceWindowRef` (Required)

Reference to a Maintenance Window that controls **when approved recommendations will be executed**.

```yaml
maintenanceWindowRef:
  apiGroup: supervisor.appscode.com   # optional
  kind: MaintenanceWindow             # or ClusterMaintenanceWindow
  name: <window-name>
  namespace: <namespace>              # required for namespaced window
```

#### What it does

* Links the policy to a time window
* Determines when approved recommendations are allowed to execute

#### Behavior

* If current time is within the window → execution happens immediately
* If outside → execution is delayed until the next available window

---

### `targets` ([]TargetRef, optional)

Defines which resources this policy applies to.

```yaml
targets:
  - group: kubedb.com
    kind: Elasticsearch
```

#### What it does

* Determines **which recommendations are eligible for auto-approval**
* Matching is done using `group` and `kind`

#### Behavior

* If not specified → policy matches no resources
* If specified → only matching resources are considered

---

## TargetRef

Each entry in `targets` defines a rule to match resources and optionally filter operations.

---

### `group` (string)

Specifies the **API group of the target resource**.

#### What it does

* Used to match the resource targeted by a Recommendation
* Must exactly match the resource’s API group

#### Example

```yaml
group: kubedb.com
```

---

### `kind` (string)

Specifies the **Kind of the target resource**.

#### What it does

* Identifies the type of resource (e.g., database)
* Used together with `group` for matching

#### Example

```yaml
kind: Elasticsearch
```

---

### `operations` ([]Operation, optional)

Defines which operation types are allowed for auto-approval.

#### What it does

* Filters recommendation operations
* Provides fine-grained control over what gets auto-approved

#### Behavior

* If not specified → all operations are allowed
* If specified → only listed operations are allowed
* Non-matching operations → require manual approval

---

## Operation

Defines a specific operation type allowed by the policy.

---

### `group` (string)

Specifies the **API group of the operation resource**.

#### What it does

* Matches the operation created by the Recommendation
* Ensures only operations from this API group are allowed

#### Example

```yaml
group: ops.kubedb.com
```

---

### `kind` (string)

Specifies the **Kind of the operation resource**.

#### What it does

* Identifies the exact operation type (e.g., restart, scaling, TLS rotation)
* Used to determine if the operation is eligible for auto-approval

#### Examples

```yaml
kind: ElasticsearchOpsRequest
kind: PostgreSQLOpsRequest
```

---

## How ApprovalPolicy Works

When a `Recommendation` is created:

1. **Matching**

   * The controller scans all `ApprovalPolicy` resources
   * Matches based on:

     * `target.group`
     * `target.kind`
     * `operation` (if defined)

2. **Approval Decision**

   * If a match is found → recommendation is automatically approved
   * If no match → remains pending (manual approval required)

3. **Override Check**

   * If `requireExplicitApproval: true` in Recommendation
     → **ApprovalPolicy is ignored**

4. **Scheduling**

   * The referenced `maintenanceWindowRef` is applied

5. **Execution**

   * If inside maintenance window → execute immediately
   * If outside → wait for next window



## Examples

### Basic Policy

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
  - group: kubedb.com
    kind: MongoDB
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
