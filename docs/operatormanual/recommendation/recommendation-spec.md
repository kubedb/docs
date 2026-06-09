---
title: Recommendation Spec & Status
menu:
  docs_{{ .version }}:
    identifier: recommendation-spec
    name: Recommendation Spec & Status
    parent: recommendation
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: operatormanual
---

> New to KubeDB? Please start [here](/docs/README.md).

# Recommendation Spec & Status

A `Recommendation` is the Kubernetes-native record of a maintenance action that KubeDB wants to perform on a managed database. The `spec` describes **what** should happen; the `status` reflects **where the action stands** in its lifecycle.

This page is the complete field reference. If you are new to recommendations, read the [Overview](/docs/operatormanual/recommendation) first, then come back here when you need to look up a specific field.

---

## Example Recommendation

```yaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: Recommendation
metadata:
  name: elastic-x-elasticsearch-x-rotate-auth-2juuee
  namespace: es
spec:
  backoffLimit: 5
  deadline: "2025-02-25T09:20:53Z"
  description: Recommending AuthSecret rotation
  operation:
    apiVersion: ops.kubedb.com/v1alpha1
    kind: ElasticsearchOpsRequest
    metadata:
      name: rotate-auth
      namespace: es
    spec:
      databaseRef:
        name: elastic
      type: RotateAuth
  recommender:
    name: kubedb-ops-manager
  rules:
    failed: has(self.status.phase) && self.status.phase == 'Failed'
    inProgress: has(self.status.phase) && self.status.phase == 'Progressing'
    success: has(self.status.phase) && self.status.phase == 'Successful'
  target:
    apiGroup: kubedb.com
    kind: Elasticsearch
    name: elastic
status:
  approvalStatus: Pending
  failedAttempt: 0
  outdated: false
  parallelism: Namespace
  phase: Pending
  reason: WaitingForApproval
```

---

## Spec Fields

* `spec.description` (string, optional) — human-readable reason the recommendation was generated.
* `spec.target` (TypedLocalObjectReference) — the database resource the recommendation acts on (API group, kind, name).
* `spec.operation` (RawExtension) — the Kubernetes resource the Supervisor will create to execute the recommendation (typically an `OpsRequest`).
* `spec.recommender` (ObjectReference) — the component that generated the recommendation (e.g. `kubedb-ops-manager`).
* `spec.deadline` (Time, optional) — execution deadline. After this time the Supervisor will auto-approve and run the operation unless explicit approval is required.
* `spec.requireExplicitApproval` (bool, optional) — if `true`, the recommendation will not run until a human approves it; any matching `ApprovalPolicy` is ignored.
* `spec.backoffLimit` (int, optional) — how many times the Supervisor will retry the operation on failure.
* `spec.rules` (OperationPhaseRules) — CEL expressions that tell the Supervisor how to interpret the operation's status (see below).

---

## OperationPhaseRules

The Supervisor evaluates the rules against the **operation resource** (not the recommendation itself). For example, given this `spec.operation`:

```shell
$ kubectl get recommendation -n es elastic-x-elasticsearch-x-update-version-2juuee \
    -o jsonpath='{.spec.operation}' | yq -y
apiVersion: ops.kubedb.com/v1alpha1
kind: ElasticsearchOpsRequest
metadata:
  name: update-version
  namespace: es
spec:
  databaseRef:
    name: elastic
  type: UpdateVersion
  updateVersion:
    targetVersion: xpack-9.2.3
status: {}
```

the rules are evaluated against the OpsRequest's `status`:

* `spec.rules.success` (string) — CEL expression that returns `true` when the operation has succeeded.
* `spec.rules.inProgress` (string) — CEL expression that returns `true` while the operation is running.
* `spec.rules.failed` (string) — CEL expression that returns `true` when the operation has failed.

> **Notes**
> * `self` refers to the **operation resource created from `spec.operation`** — typically an OpsRequest.
> * Expressions are evaluated against that resource's `status` block.

---

## Status Fields

* `status.approvalStatus` (ApprovalStatus) — `Pending`, `Approved`, or `Rejected`.
* `status.phase` (RecommendationPhase) — current lifecycle phase: `Pending`, `Waiting`, `InProgress`, `Succeeded`, `Failed`, `Skipped`.
* `status.reason` (string) — short message explaining the current state (e.g. `WaitingForApproval`, `StartedExecutingOperation`, `SuccessfullyExecutedOperation`).
* `status.reviewer` (Subject, optional) — who approved or rejected the recommendation.
* `status.comments` (string, optional) — reviewer comments.
* `status.reviewTimestamp` (Time, optional) — when the review happened.
* `status.approvedWindow` (ApprovedWindow, optional) — when execution is allowed (see below).
* `status.parallelism` (Parallelism) — concurrency control: `Namespace`, `Target`, or `TargetAndNamespace`.
* `status.outdated` (bool) — `true` when a newer, more relevant recommendation supersedes this one; outdated recommendations are not executed.
* `status.conditions` ([]Condition, optional) — granular conditions such as `SuccessfullyCreatedOperation`, `SuccessfullyExecutedOperation`.
* `status.createdOperationRef` (LocalObjectReference, optional) — the OpsRequest the Supervisor created from `spec.operation`.
* `status.failedAttempt` (int32) — how many times execution has failed so far.
* `status.observedGeneration` (int64, optional) — the spec generation last observed by the controller.

### Lifecycle phases at a glance

| Phase        | Meaning                                                                 |
| ------------ | ----------------------------------------------------------------------- |
| `Pending`    | Created, not yet approved or scheduled.                                 |
| `Waiting`    | Approved, waiting for the next allowed maintenance window.              |
| `InProgress` | OpsRequest created and running.                                         |
| `Succeeded`  | Operation completed successfully.                                       |
| `Failed`     | Operation failed and the backoff limit was reached.                     |
| `Skipped`    | Recommendation became outdated or was rejected before execution.        |

---

## ApprovedWindow Fields

* `status.approvedWindow.window` (WindowType) — execution strategy: `Immediate`, `NextAvailable`, or `SpecificDates`.
* `status.approvedWindow.maintenanceWindow` (TypedObjectReference, optional) — the MaintenanceWindow the Supervisor will honor.
* `status.approvedWindow.dates` ([]DateWindow, optional) — explicit date ranges (used with `SpecificDates`).

---

For auto-approval configuration, see [Approval Policy](/docs/operatormanual/recommendation/approval-policy.md). For scheduling, see [Maintenance Window](/docs/operatormanual/recommendation/maintenance-window.md) and [Cluster Maintenance Window](/docs/operatormanual/recommendation/cluster-maintenance-window.md).
