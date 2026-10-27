---
title: Etcd Vertical Scaling Overview
menu:
  docs_{{ .version }}:
    identifier: etcd-vertical-scaling-overview
    name: Overview
    parent: etcd-vertical-scaling
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

<!-- Drafted from source code and CRD schemas; not yet verified against a live cluster. -->

> New to KubeDB? Please start [here](/docs/README.md).

# Etcd Vertical Scaling

This guide gives an overview of how the KubeDB Ops-manager operator updates the compute resources
(CPU and Memory) of the `etcd` container of an `Etcd` cluster.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Etcd](/docs/guides/etcd/concepts/etcd.md)
  - [EtcdOpsRequest](/docs/guides/etcd/concepts/etcdopsrequest.md)

## How Vertical Scaling Process Works

The vertical scaling process consists of the following steps:

1. At first, a user creates an `Etcd` Custom Resource (CR).

2. The `KubeDB` Provisioner operator watches the `Etcd` CR.

3. When the operator finds an `Etcd` CR, it creates a `PetSet` and the related resources — Secrets,
   Services, RBAC, and the cluster-state ConfigMap.

4. Then, in order to update the resources of the `Etcd` cluster, the user creates an `EtcdOpsRequest`
   CR of type `VerticalScaling` with the desired resources.

5. The `KubeDB` Ops-manager operator watches the `EtcdOpsRequest` CR.

6. When it finds a `VerticalScaling` `EtcdOpsRequest`, it pauses the referenced `Etcd` object, so the
   `KubeDB` Provisioner operator does not reconcile it while the scaling is in progress.

7. The Ops-manager operator patches the `PetSet` pod template with the new resources, so any pod
   recreated from that point on — by this ops request or by a later failure — carries them.

8. It also patches `spec.podTemplate.spec.containers[]` on the `Etcd` object, so the Provisioner
   operator re-renders exactly the same template once the database is resumed.

9. It then actuates the change on the running pods, in the mode selected by
   `spec.verticalScaling.mode` (see below).

10. Finally, the Ops-manager operator resumes the `Etcd` object so that the Provisioner operator
    resumes its usual operations, and marks the `EtcdOpsRequest` `Successful`.

## Vertical Scaling Modes

KubeDB actuates vertical scaling in one of two modes, selected through the
`spec.verticalScaling.mode` field of the `EtcdOpsRequest`:

- **`Restart`** (default): The operator restarts the member pods so that they come back with the new
  CPU and Memory. The restart is **leader-aware and sequential** — this is the part that is specific
  to etcd:
  - The operator first resolves the current Raft leader through etcd's `Status()` RPC.
  - Every **follower** is evicted first, one at a time. After each eviction the operator waits for
    the pod to become `Ready` *and* for the cluster to report a healthy quorum again before it
    touches the next member. Without that gate, a second eviction in a 3-member cluster would take
    quorum down.
  - The **leader is restarted last**, and only after its leadership has been handed to another
    voting member with etcd's `MoveLeader` RPC. A deliberate leadership transfer is near
    instantaneous, whereas evicting the leader outright triggers an election, during which the whole
    cluster is unavailable for writes for an election timeout.
  - A single-member cluster has no one to hand leadership to, so it is simply restarted.

- **`InPlace`**: The operator resizes the running containers using the Kubernetes
  [in-place Pod resize](https://kubernetes.io/docs/tasks/configure-pod-container/resize-container-resources/)
  (`pods/resize` subresource) — no pod restart, no leadership change, and no Raft election risk at
  all. If a Node cannot accommodate the new request (the kubelet marks the resize `Infeasible`), the
  operator automatically falls back to the `Restart` behavior **for those pods only**, using the same
  leader-last ordering described above.

If `spec.verticalScaling.mode` is omitted, it defaults to `Restart`.

> **Note:** `InPlace` mode relies on the Kubernetes `InPlacePodVerticalScaling` feature gate, which
> is enabled by default from Kubernetes v1.33. On older clusters, or when the feature gate is
> disabled, use `Restart` mode.

## Scalable Containers

`spec.verticalScaling` for etcd has no node-selection or topology sub-fields — etcd has exactly one
container role, so there is a single `etcd` entry:

| Field                            | Container | Notes                                                                |
|----------------------------------|-----------|----------------------------------------------------------------------|
| `spec.verticalScaling.etcd`      | `etcd`    | The etcd member container. This is the field you normally want.       |
| `spec.verticalScaling.exporter`  | `exporter`| Present for structural parity with other KubeDB databases only.        |

> **Important:** KubeDB does **not** deploy a metrics exporter sidecar for etcd — etcd serves its own
> Prometheus metrics natively on the metrics listener (port `2381`), so there is nothing extra to
> resize. `spec.verticalScaling.exporter` exists in the `EtcdOpsRequest` API for structural parity
> with the other KubeDB databases, and setting it has **no effect** unless you have added a container
> literally named `exporter` yourself through `spec.podTemplate`. Do not rely on it.

In the [next](/docs/guides/etcd/scaling/vertical-scaling/vertical-scaling.md) doc, we are going to
show a step-by-step guide on updating the resources of an Etcd cluster using the `EtcdOpsRequest` CRD.
