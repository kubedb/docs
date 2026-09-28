---
title: Reconfiguring Etcd TLS/SSL
menu:
  docs_{{ .version }}:
    identifier: etcd-reconfigure-tls-overview
    name: Overview
    parent: etcd-reconfigure-tls
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

<!-- Drafted from source code and CRD schemas; not yet verified against a live cluster. -->

> New to KubeDB? Please start [here](/docs/README.md).

# Reconfiguring TLS of Etcd

This guide gives an overview of how the KubeDB Ops-manager operator reconfigures the TLS configuration of an `Etcd` cluster — adding TLS, removing TLS, changing the `Issuer`/`ClusterIssuer` or certificate specs, and rotating the certificates.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
    - [Etcd](/docs/guides/etcd/concepts/etcd.md)
    - [EtcdOpsRequest](/docs/guides/etcd/concepts/etcdopsrequest.md)
    - [Etcd TLS/SSL overview](/docs/guides/etcd/tls/overview.md)

## The three mutually exclusive operations

`spec.tls` on an `EtcdOpsRequest` of type `ReconfigureTLS` describes exactly **one** operation per request. The admission webhook counts them and rejects a request that names more than one, with `only one TLS reconfiguration operation is allowed at a time`:

| Operation             | Field(s)                        | Meaning                                                                                                                              |
|-----------------------|---------------------------------|--------------------------------------------------------------------------------------------------------------------------------------|
| **Remove TLS**        | `remove: true`                  | Clears `spec.tls` on the `Etcd` object, re-renders the members without any TLS flags, and deletes the now-orphaned `Certificate` objects. |
| **Rotate**            | `rotateCertificates: true`      | Asks cert-manager to re-issue every existing certificate, keeping the same issuer and the same certificate specs.                       |
| **Issue / update**    | `issuerRef` and/or `certificates` | Points the cluster at a different `Issuer`/`ClusterIssuer`, or changes the per-alias certificate specs. This is also how you add TLS to a cluster that does not have it yet. |

Two further validations are worth knowing before you write the request:

- `rotateCertificates: true` requires TLS to be enabled already — the webhook rejects it with `rotateCertificates requires TLS to already be enabled with issuerRef on Etcd`. There is nothing to rotate otherwise.
- When you supply only `certificates` (no `issuerRef`), the `Etcd` object must already carry an `issuerRef`, otherwise the request is rejected with `tls.issuerRef is required for Etcd ReconfigureTLS`.

## How the reconfiguring process works

1. A user creates an `Etcd` Custom Resource Object (CRO).

2. The `KubeDB` Provisioner operator watches the `Etcd` CRO and creates the `PetSet` and the necessary secrets, services, etc.

3. In order to change the TLS configuration, the user creates an `EtcdOpsRequest` CR of type `ReconfigureTLS`.

4. The `KubeDB` Ops-manager operator watches the `EtcdOpsRequest` CR.

5. When it finds one, it **pauses** the referenced `Etcd` object, so the Provisioner operator does not fight it for ownership of the `PetSet` during the operation.

6. The operator folds the request into a copy of the `Etcd` object and reconciles the cert-manager `Certificate` objects from it — creating, updating or (for `remove`) preparing to delete them. For a rotation it puts cert-manager's `Issuing` condition on each `Certificate`, which is what makes cert-manager re-issue it.

7. It then waits for cert-manager to finish, recording the `CertificateSynced` condition on the ops request once every certificate secret carries the new material.

8. It re-renders the `PetSet` (condition `UpdateEtcdPetSet`). The TLS flags, volumes and volume mounts on the etcd container are all derived from `spec.tls`, so this single step covers adding, changing and removing TLS.

9. It restarts the members (condition `RestartEtcdPods`). This is the **leader-aware rolling restart**: the operator identifies the current Raft leader, restarts every follower first — one at a time, waiting for each pod to become Ready and for the cluster to report quorum before moving to the next — and restarts the leader **last**, after handing its leadership to another voter with etcd's `MoveLeader` RPC. The ordering matters for etcd in a way it does not for a leader/follower SQL database: evicting the leader outright triggers an election, and an election makes the cluster unavailable for writes for an election timeout, whereas a deliberate transfer is near instantaneous. Peer traffic is mutually authenticated, so a member restarted with new material also has to rejoin the quorum before the next one is taken down.

10. It writes the new `spec.tls` back onto the `Etcd` object (condition `UpdateDatabase`) and deletes any `Certificate` objects that are no longer referenced.

11. Finally it resumes the `Etcd` object so the Provisioner operator returns to its usual duties, and marks the ops request `Successful`.

Because step 9 restarts every member one at a time and waits for quorum in between, the cluster stays available for reads and writes throughout — a `ReconfigureTLS` is not a downtime operation for a 3-member (or larger) cluster. A single-member cluster (`replicas: 1`) has no quorum to preserve and will be briefly unavailable while its only member restarts.

In the [next](/docs/guides/etcd/reconfigure-tls/reconfigure-tls.md) doc, we show a step-by-step guide for each of the four scenarios using the `EtcdOpsRequest` CRD.
