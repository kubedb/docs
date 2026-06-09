---
title: Rotate Auth Recommendation
menu:
  docs_{{ .version }}:
    identifier: rotate-auth-recommendation
    name: Rotate Auth
    parent: recommendation
    weight: 60
menu_name: docs_{{ .version }}
section_menu_id: operatormanual
---

> New to KubeDB? Please start [here](/docs/README.md).

# Authentication Rotate Recommendation

Database credentials are a high-value target. Leaving them static — sometimes for years — magnifies the blast radius of any leak: backup tarballs, log lines, abandoned dev pods, or a compromised CI runner can all expose them. Regular rotation limits how long a stolen credential is useful, satisfies compliance audits, exercises the rotation code path itself (so it actually works when you need it), and revokes access for stale humans and services.

KubeDB ships the `RotateAuth` OpsRequest for this — and the Ops-manager generates a `Recommendation` to drive it automatically. Rotation is opt-in: it is only generated when the database CR sets `spec.authSecret.rotateAfter`.

KubeDB raises an auth-rotation recommendation when:

1. AuthSecret lifespan is **more than one month** and **less than one month** of life remains, or
2. AuthSecret lifespan is **less than one month** and **less than one third** of life remains.

> **Note:** Recommendations work across most KubeDB-managed databases. The walkthrough below uses [MongoDB](/docs/guides/mongodb).

## Before you begin

- KubeDB and the **Supervisor** are installed (`--set supervisor.enabled=true`).
- `kubectl` is configured against the cluster.
- The target database CR must set `spec.authSecret.rotateAfter`. Without it, no rotation Recommendation is generated.

## Deploy MongoDB with rotation enabled

For the demo we use an aggressive `rotateAfter: 1h`. In production, pick something realistic like `2160h` (90 days).

```yaml
apiVersion: kubedb.com/v1
kind: MongoDB
metadata:
  name: mg-rarecommendation
  namespace: demo
spec:
  version: "8.0.10"
  authSecret:
    name: mg-rarecommendation-auth
    rotateAfter: 1h
  storage:
    resources:
      requests:
        storage: 500Mi
    storageClassName: local-path
  deletionPolicy: WipeOut
```

Wait until MongoDB reports `Ready`. The time depends on image pull speed.

```bash
$ kubectl get mongodb,pods -n demo
NAME                                     VERSION   STATUS   AGE
mongodb.kubedb.com/mg-rarecommendation   8.0.10    Ready    10m

NAME                        READY   STATUS    RESTARTS   AGE
pod/mg-rarecommendation-0   1/1     Running   0          10m
```

## A rotate-auth Recommendation appears

With `rotateAfter: 1h`, the recommendation engine creates a rotation Recommendation roughly **40 minutes** after the auth secret was created (two-thirds of the lifespan). Once it appears, you will see something like:

```bash
$ kubectl get recommendation -n demo
NAME                                                          STATUS      OUTDATED   AGE
mg-rarecommendation-x-mongodb-x-rotate-auth-<HASH>            <STATUS>    false      <AGE>
mg-rarecommendation-x-mongodb-x-update-version-<HASH>         Pending     false      <AGE>
```

The Recommendation name follows the pattern `<DB-name>-x-<DB-type>-x-<recommendation-type>-<random-suffix>`. Let's look at the full manifest:

```yaml
$ kubectl get recommendation -n demo <ROTATE-AUTH-NAME> -oyaml
<YAML-PLACEHOLDER>
```

What this manifest tells you:

- `spec.description` — explains that the auth secret needs rotation before the secret's expiry timestamp.
- `spec.deadline` — Supervisor will auto-approve and execute at or before this time so rotation finishes before the secret expires.
- `spec.operation` — a `MongoDBOpsRequest` of type `RotateAuth`.
- Notice that `spec.requireExplicitApproval` is **not set**. Auth secret rotation defaults to automatic approval — like TLS, missing the deadline is worse than running unattended.
- When the deadline is reached, Supervisor sets `status.approvalStatus: Approved`, `status.approvedWindow.window: Immediate`, creates the OpsRequest, and reports progress via `status.conditions`.

## Watching the OpsRequest

After auto-approval, an `MongoDBOpsRequest` is created and reaches `Successful`:

```bash
$ kubectl get mongodbopsrequest -n demo
NAME                                              TYPE         STATUS       AGE
mg-rarecommendation-<TIMESTAMP>-rotate-auth-auto  RotateAuth   Successful   <AGE>
```

`RotateAuth` rotates the auth secret with negligible downtime — the database keeps accepting connections throughout the rolling restart.

You can re-check the Recommendation status as JSON:

```bash
$ kubectl get recommendation <ROTATE-AUTH-NAME> \
     -n demo -o json | jq '.status'
```

You will see `phase: Succeeded` and `reason: SuccessfullyExecutedOperation`.

## Rejecting a recommendation

If you need to skip a rotation (for example because you're about to change auth strategy), reject it:

```bash
$ kubectl patch recommendation <ROTATE-AUTH-NAME> \
     -n demo \
     --type merge \
     --subresource='status' \
     -p '{"status":{"approvalStatus":"Rejected"}}'
recommendation.supervisor.appscode.com/<ROTATE-AUTH-NAME> patched
```

## Automating execution with a maintenance window

Auto-approval is great, but `Immediate` execution can still be disruptive at the wrong hour. Pair a `MaintenanceWindow` with an `ApprovalPolicy` to keep rotations off-peak.

### 1. Define a MaintenanceWindow

```yaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: MaintenanceWindow
metadata:
  name: mongo-maintenance
  namespace: demo
spec:
  isDefault: true
  timezone: UTC
  days:
    Sunday:
      - start: "00:00"
        end: "04:00"
    Saturday:
      - start: "00:00"
        end: "04:00"
```

See [Maintenance Window](/docs/operatormanual/recommendation/maintenance-window.md) for the complete reference.

### 2. Auto-approve with an ApprovalPolicy

This policy auto-approves `RotateAuth` (and any other MongoDB ops) and binds them to the window above.

```yaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: ApprovalPolicy
metadata:
  name: mongodb-policy
  namespace: demo
maintenanceWindowRef:
  kind: MaintenanceWindow
  name: mongo-maintenance
  namespace: demo
targets:
  - group: kubedb.com
    kind: MongoDB
    operations:
      - group: ops.kubedb.com
        kind: MongoDBOpsRequest
```

See [Approval Policy](/docs/operatormanual/recommendation/approval-policy.md) for the complete reference.

### 3. Or use a cluster-wide default

If you would rather set one default for the whole cluster, replace the namespace-scoped `MaintenanceWindow` with a `ClusterMaintenanceWindow` and point `maintenanceWindowRef.kind` at it. See [Cluster Maintenance Window](/docs/operatormanual/recommendation/cluster-maintenance-window.md).

> **Important:** Make sure the window opens **before** the recommendation's `spec.deadline`. If the deadline passes first, Supervisor rotates immediately to keep the auth secret from expiring.

---

For the complete field reference for Recommendation, see [Recommendation Spec & Status](/docs/operatormanual/recommendation/recommendation-spec.md).
