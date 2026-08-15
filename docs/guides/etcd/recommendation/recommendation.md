---
title: Etcd Recommendation Overview
menu:
  docs_{{ .version }}:
    identifier: etcd-recommendation-overview
    name: Recommendation Overview
    parent: etcd-recommendation
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

<!-- Drafted from source code and CRD schemas; not yet verified against a live cluster. -->

> New to KubeDB? Please start [here](/docs/README.md).

# Etcd Recommendation

## Overview

A `Recommendation` is a Kubernetes-native CRD created by the **KubeDB Ops-Manager** and reconciled
by the **KubeDB Supervisor**. For an etcd cluster managed by KubeDB, the Ops-Manager watches the
`Etcd` object's state and emits a Recommendation whenever it detects an action you should take —
a newer version in the catalog, an expiring TLS certificate, or an authentication secret nearing
its rotation deadline.

Nothing runs until the Recommendation is approved — either by you
(`status.approvalStatus: Approved`) or automatically, through the deadline the Ops-Manager set or
through an `ApprovalPolicy` bound to a `MaintenanceWindow`. Once approved, the Supervisor creates
the corresponding `EtcdOpsRequest` and tracks it to completion. **You do not write the
`EtcdOpsRequest` yourself** — the fully-formed request is already carried inside the
Recommendation's `spec.operation`.

This page is the **etcd-specific intro**: which recommendations apply to etcd and which spec
fields trigger them. For prerequisites, Helm flags that control generation timing, and the full
Recommendation lifecycle, see:

- [Recommendation Configuration](/docs/operatormanual/recommendation/configuration.md) — prerequisites, Supervisor CRD install, and all Helm flags.
- [Recommendation Overview](/docs/operatormanual/recommendation) — architecture and lifecycle walkthrough.

<p align="center">
  <img alt="Recommendation Lifecycle" src="/docs/operatormanual/recommendation/images/recommendation-generation.png">
</p>

---

## Relevant KubeDB concepts

- [Etcd](/docs/guides/etcd/concepts/etcd.md)
- [EtcdVersion](/docs/guides/etcd/concepts/etcdversion.md)
- [EtcdOpsRequest](/docs/guides/etcd/concepts/etcdopsrequest.md)
- [Rotate Authentication](/docs/guides/etcd/rotate-authentication/overview.md)
- [Reconfigure TLS](/docs/guides/etcd/tls/overview.md)
- [Update Version](/docs/guides/etcd/update-version/overview.md)

---

## Recommendation types for Etcd

| Type                               | Generated `EtcdOpsRequest` type | Triggered when                                                                    | Walkthrough                                                                                                       |
| ---------------------------------- | ------------------------------- | --------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------- |
| **Version Update**                 | `UpdateVersion`                 | A newer major, minor, or patch `EtcdVersion` becomes available in the catalog      | [Version Update Recommendation](/docs/operatormanual/recommendation/version-update-recommendation.md)              |
| **Same-Version Update**            | `UpdateVersion`                 | The container image for your *current* `EtcdVersion` is refreshed                  | [Version Update Recommendation](/docs/operatormanual/recommendation/version-update-recommendation.md)              |
| **TLS Certificate Rotation**       | `ReconfigureTLS`                | An issued certificate is approaching its expiry threshold                          | [TLS Certificate Rotation Recommendation](/docs/operatormanual/recommendation/rotate-tls-recommendation.md)        |
| **Authentication Secret Rotation** | `RotateAuth`                    | The auth secret is approaching its `rotateAfter` deadline                          | [Authentication Secret Rotation Recommendation](/docs/operatormanual/recommendation/rotate-auth-recommendation.md) |

The engine itself is the same one every KubeDB database uses; there is nothing etcd-specific about
the mechanism. What *is* etcd-specific is which triggers apply, and what the resulting operation
does to a Raft cluster.

Two conditions gate generation for every type:

- **`spec.autoOps.disabled`** on the `Etcd` object. Set it to `true` and no recommendations are
  generated for that cluster at all.
- **No other `EtcdOpsRequest` may be `Progressing`.** If the cluster is already in the middle of
  an operation, the engine waits rather than stacking a second one on top.

---

## Triggers specific to Etcd

This section shows the minimal `Etcd` CR fields that cause each recommendation to be generated.
For deeper, end-to-end walkthroughs use the links in the table above.

### Authentication Secret Rotation

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Etcd
metadata:
  name: etcd-recommendation
  namespace: demo
spec:
  version: "3.6.4"
  replicas: 3
  authSecret:
    name: etcd-recommendation-auth
    rotateAfter: 720h
```

In this configuration:

* The `rotateAfter` field defines how long the authentication secret remains valid.

KubeDB monitors the configured lifecycle and generates a RotateAuth Recommendation based on the
following conditions:

* If the secret lifespan is greater than one month, a recommendation is generated when less than
  one month of validity remains.

* If the secret lifespan is less than one month, a recommendation is generated when approximately
  one-third of its validity remains.

Note that for etcd this trigger is always live: KubeDB always provisions etcd with its RBAC
enabled and a `root` user, so the engine treats authentication as unconditionally enabled — there
is no "auth is off" case for an etcd cluster the way there is for some other databases.

Once approved, the Supervisor creates an `EtcdOpsRequest` of type `RotateAuth`. For etcd this is
an unusually cheap operation: an etcd RBAC password change is applied live through the
`UserChangePassword` API, so **no pods are restarted** and there is no interruption to the client
API.

### TLS Certificate Rotation

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Etcd
metadata:
  name: etcd-recommendation
  namespace: demo
spec:
  version: "3.6.4"
  replicas: 3
  tls:
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: etcd-ca-issuer
    certificates:
      - alias: server
        duration: 2160h
      - alias: client
        duration: 2160h
      - alias: peer
        duration: 2160h
```

In this configuration:

* The `spec.tls.certificates[].duration` field defines how long each certificate remains valid.
* The engine only tracks certificates when `spec.tls.issuerRef` is set — a cluster without an
  issuer has no certificates for it to watch.

KubeDB monitors the configured lifecycle and generates a RotateTLS Recommendation based on the
following conditions:

* If the certificate duration is greater than one month, a recommendation is generated when less
  than one month of validity remains.

* If the certificate duration is less than one month, a recommendation is generated when
  approximately one-third of its validity remains.

Pay particular attention to the `peer` certificate for etcd. Peer traffic is Raft traffic between
members, and it is always mutually authenticated once TLS is enabled — an expired peer
certificate does not degrade the cluster, it partitions it. That is the main reason to leave TLS
rotation on automatic approval rather than gating it behind a human.

Once approved, the Supervisor creates an `EtcdOpsRequest` of type `ReconfigureTLS` with
`tls.rotateCertificates: true`. The Ops-manager re-issues the certificates through cert-manager
and then performs a leader-aware rolling restart: leadership is moved off the leader before its
pod is recreated, pods are restarted one at a time, and each must be Ready and the cluster
quorum-healthy before the next one is touched.

### Version Update

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Etcd
metadata:
  name: etcd-recommendation
  namespace: demo
spec:
  version: "3.6.4"
  replicas: 3
```

In this configuration:

* KubeDB monitors the running version of the database against the `EtcdVersion` catalog objects
  installed in the cluster.

KubeDB generates a VersionUpdate Recommendation based on the following conditions:

* If a patch version is released, a recommendation is generated.

* If a newer minor or major version becomes available, a recommendation is generated.

* Candidate versions are filtered through the current `EtcdVersion`'s
  `spec.updateConstraints.allowlist` / `denylist`, deprecated versions are skipped, and only
  versions built on the same base OS are considered.

Version update recommendations for etcd are created with `spec.requireExplicitApproval: true` —
they will **not** auto-approve on a deadline. That is deliberate: an etcd version update is
gated by an upgrade-path check (same major, no downgrade, at most one minor step at a time) and
rolls every member, so it is the one recommendation type that always waits for a human or an
`ApprovalPolicy`.

For example: recommending a version update from `3.6.4` to `3.6.5`.

If the `Etcd` object's version changes to something other than what an outstanding recommendation
targeted, the engine marks that recommendation **outdated** rather than leaving it to fire later
against a version you no longer run.

### Same-Version Update

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Etcd
metadata:
  name: etcd-recommendation
  namespace: demo
spec:
  version: "3.6.4"
  replicas: 3
```

In this configuration:

* KubeDB compares the images pinned by the current `EtcdVersion` — the etcd server image, the
  exporter image and, when present, the init container image — against what the running PetSet
  actually references.

* If the container image backing the current version is updated (a security patch or a rebuild
  without a version bump), a recommendation is generated.

Once approved, the Supervisor creates an `EtcdOpsRequest` targeting the same version, which rolls
the members onto the refreshed image.

---

## Listing and inspecting Recommendations

Recommendations are namespaced, and live in the same namespace as the database.

```bash
$ kubectl get recommendation -n demo
NAME                                                     STATUS    OUTDATED   AGE
etcd-recommendation-x-etcd-x-rotate-tls-p4k9zq           Pending   false      12m
etcd-recommendation-x-etcd-x-update-version-h7m2ab       Pending   false      6m
```

The name follows the pattern `<db-name>-x-<db-kind>-x-<recommendation-type>-<random-suffix>`, so
`etcd-recommendation-x-etcd-x-rotate-tls-p4k9zq` is a TLS rotation recommendation for the `Etcd`
object named `etcd-recommendation`.

Every Recommendation is labelled, which makes filtering easy:

```bash
# everything KubeDB generated for one particular Etcd cluster
$ kubectl get recommendation -n demo -l app.kubernetes.io/instance=etcd-recommendation

# only the version-update ones
$ kubectl get recommendation -n demo -l app.kubernetes.io/type=version-update
```

The label values are `rotate-tls`, `rotate-auth` and `version-update`, and every KubeDB-generated
Recommendation also carries `app.kubernetes.io/managed-by: kubedb.com`.

To see what a Recommendation actually proposes, read its `spec.operation` — it holds the complete
`EtcdOpsRequest` the Supervisor will create:

```bash
$ kubectl get recommendation -n demo etcd-recommendation-x-etcd-x-rotate-tls-p4k9zq -o yaml
apiVersion: supervisor.appscode.com/v1alpha1
kind: Recommendation
metadata:
  creationTimestamp: "2026-02-11T13:41:20Z"
  generation: 1
  labels:
    app.kubernetes.io/instance: etcd-recommendation
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/type: rotate-tls
  name: etcd-recommendation-x-etcd-x-rotate-tls-p4k9zq
  namespace: demo
  resourceVersion: "80311"
  uid: c5d2e9b1-7a04-4f38-9e22-1b3c4d5e6f70
spec:
  deadline: "2026-03-06T09:00:00Z"
  description: Recommending TLS certificate rotation, etcd-recommendation-peer-cert
    Certificate is going to be expire on 2026-03-06 09:20:00 +0000 UTC
  operation:
    apiVersion: ops.kubedb.com/v1alpha1
    kind: EtcdOpsRequest
    metadata:
      name: rotate-tls
      namespace: demo
    spec:
      databaseRef:
        name: etcd-recommendation
      tls:
        rotateCertificates: true
      type: ReconfigureTLS
    status: {}
  recommender:
    name: kubedb-ops-manager
  rules:
    failed: has(self.status) && has(self.status.phase) && self.status.phase == 'Failed'
    inProgress: has(self.status) && has(self.status.phase) && self.status.phase == 'Progressing'
    success: has(self.status) && has(self.status.phase) && self.status.phase == 'Successful'
  target:
    apiGroup: kubedb.com
    kind: Etcd
    name: etcd-recommendation
status:
  outdated: false
  phase: Pending
```

A version update recommendation looks the same except for its `spec.operation` and the explicit
approval flag:

```yaml
spec:
  requireExplicitApproval: true
  operation:
    apiVersion: ops.kubedb.com/v1alpha1
    kind: EtcdOpsRequest
    metadata:
      name: update-version
      namespace: demo
    spec:
      databaseRef:
        name: etcd-recommendation
      type: UpdateVersion
      updateVersion:
        targetVersion: 3.6.5
    status: {}
```

---

## Acting on a Recommendation

Approve it. That is the whole interaction — the Supervisor turns the approval into the
`EtcdOpsRequest` for you.

```bash
$ kubectl patch recommendation etcd-recommendation-x-etcd-x-update-version-h7m2ab \
     -n demo \
     --type merge \
     --subresource='status' \
     -p '{"status":{"approvalStatus":"Approved","approvedWindow":{"window":"Immediate"}}}'
recommendation.supervisor.appscode.com/etcd-recommendation-x-etcd-x-update-version-h7m2ab patched
```

Watch the Recommendation pick up the created operation:

```bash
$ kubectl get recommendation -n demo etcd-recommendation-x-etcd-x-update-version-h7m2ab -o jsonpath='{.status}'
```

`status.createdOperationRef.name` names the `EtcdOpsRequest` the Supervisor created, which you can
then follow exactly like any hand-written one:

```bash
$ kubectl get etcdopsrequest -n demo
NAME                                                  TYPE            STATUS       AGE
etcd-recommendation-1780937248-update-version-auto    UpdateVersion   Successful   6m
```

If you want to skip a recommendation instead — for example because you are about to swap issuers,
or you are pinning the version deliberately — reject it:

```bash
$ kubectl patch recommendation etcd-recommendation-x-etcd-x-update-version-h7m2ab \
     -n demo \
     --type merge \
     --subresource='status' \
     -p '{"status":{"approvalStatus":"Rejected"}}'
recommendation.supervisor.appscode.com/etcd-recommendation-x-etcd-x-update-version-h7m2ab patched
```

A rejected version-update recommendation is remembered: the engine will not immediately re-create
an identical one for the same target version.

To take the approval decision off your hands entirely, pair an `ApprovalPolicy` with a
`MaintenanceWindow` so that recommendations are auto-approved but only execute off-peak. That is
the recommended setup for etcd, where the version update rolls every member of the Raft cluster.
See [Approval Policy](/docs/operatormanual/recommendation/approval-policy.md) and
[Maintenance Window](/docs/operatormanual/recommendation/maintenance-window.md).

---

For prerequisites, Helm configuration flags, and the full cross-database Recommendation lifecycle,
see the [Recommendation Configuration](/docs/operatormanual/recommendation/configuration.md) and
[Recommendation Overview](/docs/operatormanual/recommendation) in the operator manual.

## Next Steps

- [Autoscale the compute resources of an Etcd cluster](/docs/guides/etcd/autoscaler/compute/compute-autoscale.md).
- [Etcd Maintenance Operations](/docs/guides/etcd/maintenance/overview.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
