---
title: PostgreSQL Cross-Cluster Branching | KubeDB
description: Branch a PostgreSQL database into a different Kubernetes cluster over Open Cluster Management
menu:
  docs_{{ .version }}:
    identifier: guides-pg-branch-cross-cluster
    name: Cross-Cluster Branching
    parent: guides-pg-branch
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="Cross-cluster branching needs more than the Courier operator: an Open Cluster Management (OCM) hub with both clusters registered as spokes, the courier hub addon manager installed on the hub, and — unlike same-cluster branching — a storage backend that both clusters can resolve a snapshot against. See Prerequisites below before trying this." >}}

# PostgreSQL Cross-Cluster Branching

Everything in [Same-Cluster Branching](/docs/guides/postgres/branch/same-cluster/index.md) and [Customizing a Branch](/docs/guides/postgres/branch/customization/index.md) — the `Branch` CR, its fields, its lifecycle, `spec.postActions`, scheduled refresh — works identically here. The one thing that changes is `spec.target.clusterName`: set it to a cluster other than the source's own, and Courier routes the same `Branch` object through a different code path to build the target database on a different cluster instead of the same one.

This page covers what is different about that path: the architecture, what you need before you can use it, and what to expect operationally. It is a reference, not a hands-on walkthrough — cross-cluster branching needs an OCM hub and two spokes sharing a storage backend, an environment this doc set does not assume you have. Read this alongside the [Overview](/docs/guides/postgres/branch/overview/index.md) and the [Branch CRD reference](/docs/guides/postgres/concepts/branch.md), which this page does not repeat.

## Why Cross-Cluster

Same-cluster branching answers "give me a copy of this database, over there in another namespace." Cross-cluster branching answers the same question when "over there" is a different Kubernetes cluster entirely — a staging cluster fed from a production cluster's data, a per-region dev cluster, or an isolated cluster for a partner or auditor who should never have credentials to the source cluster itself.

## Architecture

A same-cluster branch is one operator doing everything end to end. A cross-cluster branch spans three actors, because the two clusters involved (in OCM terms, two *spokes*) never talk to each other directly — only to the *hub* that manages both:

| Role | Runs on | Responsibility |
|---|---|---|
| **Initiator** | Source cluster | Snapshots the source database's volumes and hands the manifests for the branch (the target `Postgres`, its secrets, the snapshot references) to the hub. |
| **Branch addon manager** | Hub | Relays what the initiator hands it to the creator, and relays the creator's status back. |
| **Creator** | Target cluster | Receives the manifests, clones PVCs from the source's snapshot over the shared storage backend, and lets KubeDB provision the branch. |

The same `kubedb-courier run` binary plays all of Initiator, Creator, and — for a same-cluster branch — Local; it decides which by comparing its own `--cluster-name` to the `Branch`'s `spec.target.clusterName`. `status.mode` on the `Branch` object reports which role that copy of the operator is playing:

| Mode | Meaning |
|---|---|
| `Local` | Same-cluster branch — see [Same-Cluster Branching](/docs/guides/postgres/branch/same-cluster/index.md). |
| `Initiator` | This is the source cluster of a cross-cluster branch. |
| `Creator` | This is the target cluster of a cross-cluster branch. |

Two custom resources carry a cross-cluster branch across the hub, mirroring the way OCM itself ships work to a spoke:

- **`BranchWork`** — created by the initiator in its own namespace on the hub. It names the target cluster and carries the manifests the creator needs: the `Branch` CR, the target `Postgres`, its auth and config Secrets, and the `VolumeSnapshotContent`/`VolumeSnapshot` pair that lets the creator clone from the initiator's snapshot without re-reading the source. A spoke's hub credentials are scoped to its own namespace on the hub, so the initiator cannot write directly into the target's namespace — hence a `BranchWork` on its own side rather than a `ManifestWork` on the target's.
- **`ManifestWork`** — OCM's own delivery mechanism. The branch addon manager translates each `BranchWork` into a `ManifestWork` in the target cluster's namespace on the hub; OCM's work agent on the target spoke applies it. Status flows back the same way it was delivered: the work agent watches the delivered `Branch` (OCM `feedbackRules` with `feedbackScrapeType: Watch`, so status changes are pushed rather than polled), the manager copies that into the `BranchWork`, and the initiator copies it into the source-side `Branch.status` — which is the only object you, as a user, ever need to look at.

So a cross-cluster branch still shows up as a single `Branch` object on the source cluster, and its `status` reflects what is happening on the target cluster in near real time, without you needing direct access to the target cluster at all.

## Prerequisites

Everything the [Overview's Requirements](/docs/guides/postgres/branch/overview/index.md#requirements) section lists still applies (CSI snapshot support, a matching `VolumeSnapshotClass`, and so on) — on **both** clusters. On top of that:

- **An OCM hub, with both clusters registered as managed clusters (spokes).** Cross-cluster branching is built on [Open Cluster Management](https://open-cluster-management.io/); this guide assumes the hub and spoke registration already exists rather than walking through setting one up.
- **The courier hub addon manager installed on the hub.** It runs the `kubedb-courier manager` sub-command, registers as an OCM addon, and is what mints each spoke's scoped hub kubeconfig and translates `BranchWork` into `ManifestWork`.
- **The Courier operator running on both clusters** — `kubedb-courier run` needs `--cluster-name` set to that cluster's own OCM cluster name so it can tell Local, Initiator, and Creator apart, plus the `--hub-kubeconfig` and `--hub-namespace` the manager provisioned for it during addon registration.
- **A storage backend both clusters can resolve a snapshot against.** This is the requirement that has no same-cluster equivalent. Same-cluster branching clones a PVC from a `VolumeSnapshot` in the same cluster; cross-cluster branching clones a PVC on the target cluster from a snapshot taken on the source cluster, which only works if the CSI driver's backend (the same storage appliance, or the same cloud storage service) is reachable from — and the snapshot handle is importable by — both clusters. Two clusters each with their own unrelated local storage cannot cross-cluster branch.

{{< notice type="warning" message="Cross-cluster VolumeGroupSnapshot import is not broadly supported by CSI drivers today. A single-PVC source (a standalone database, or one PVC per engine) crosses cleanly. A multi-PVC HA source crosses as one VolumeSnapshotContent import per PVC, which only stays crash-consistent if the backend itself guarantees the group of source snapshots was taken at one consistent point — that guarantee is backend-specific, so verify it with your storage vendor before relying on it for an HA cross-cluster branch." >}}

## The Branch CR

The only difference from a same-cluster `Branch` is one field:

```yaml
apiVersion: courier.kubedb.com/v1alpha1
kind: Branch
metadata:
  name: dev-postgres-remote
  namespace: demo
spec:
  source:
    databaseRef:
      apiGroup: kubedb.com
      kind: Postgres
      name: sample-postgres
    namespace: demo
  target:
    clusterName: dev-cluster    # a different cluster than this one selects cross-cluster
    namespace: demo
    name: dev-postgres-remote
    storageClassName: netapp-trident   # StorageClass in the TARGET cluster
    issuerRef:
      apiGroup: cert-manager.io
      kind: Issuer
      name: dev-ca                     # Issuer in the TARGET cluster
  resetRootPassword: true
  deletionPolicy: Delete
```

This `Branch` is created on the **source** cluster, next to `sample-postgres`, exactly like a same-cluster branch. `spec.target.storageClassName` and `spec.target.issuerRef` name resources in the **target** cluster, not the source — the same rule same-cluster branching already follows for a cross-namespace target, just extended to a different cluster.

## What Happens

1. You create the `Branch` on the source cluster. The operator there compares `spec.target.clusterName` to its own `--cluster-name`, finds them different, and takes the **Initiator** role.
2. The initiator snapshots the source database's volumes, exactly as a same-cluster branch would — the source is never touched beyond that.
3. Instead of cloning locally, the initiator writes a `BranchWork` in its own namespace on the hub, carrying the target `Postgres` manifest, its secrets, and references to the snapshot(s) it just took.
4. The branch addon manager on the hub turns that `BranchWork` into a `ManifestWork` in the target cluster's namespace.
5. OCM delivers the manifests to the target cluster. The Courier operator there sees its own `--cluster-name` matches `spec.target.clusterName` and takes the **Creator** role.
6. The creator clones the target PVCs from the delivered snapshot references, over the shared storage backend — no data crosses the hub itself, only manifests and status.
7. KubeDB provisions the target `Postgres` on the cloned PVCs, in branched mode, the same way it would for a same-cluster branch (see [How Branching Works](/docs/guides/postgres/branch/overview/index.md#how-branching-works)).
8. If set, `spec.resetRootPassword` and `spec.postActions` run on the target exactly as they would locally.
9. The creator's status feeds back through `ManifestWork` to the manager, into the `BranchWork`, and into the source-side `Branch.status` — so `kubectl get branch` on the source cluster shows the same `Ready` phase and conditions a same-cluster branch would, even though the database it describes is running somewhere else.

## Refresh

A scheduled cross-cluster branch (`spec.schedule.cron`) still schedules only on the source: the initiator wakes on the cron tick, takes a fresh snapshot, and ships the new snapshot reference through the same `BranchWork`/`ManifestWork` path with a bumped refresh generation. The creator only halts the target database once the new snapshot has actually arrived, so a delivery failure or a hub outage leaves the branch serving its previous copy rather than sitting halted. The initiator holds onto the previous snapshot until the creator reports the new copy `Ready`, so a broken refresh does not strand the target without any snapshot to fall back to.

## Deletion

Deleting the source-side `Branch` propagates the same way creation did, in reverse, and the source `Branch` keeps its finalizer until the target confirms teardown:

1. The initiator deletes its `BranchWork`.
2. The branch addon manager deletes the corresponding `ManifestWork`, with the delete behavior matching `spec.deletionPolicy` — the delivered resources are removed for `Delete`, or left in place (with ownership metadata stripped, same as a same-cluster `Orphan`) for `Orphan`.
3. The creator tears down the target database and its cloned PVCs (or, for `Orphan`, just detaches them from the branch).
4. Once the target confirms teardown, the initiator releases its retained source snapshot and the source `Branch` CR is finally removed.

Target teardown is confirmed before the source snapshot is released deliberately: on a copy-on-write backend, a snapshot generally cannot be reclaimed while a clone still descends from it, so releasing it before the target's clone is gone would either fail or silently break the target.

## Limitations

- **No shared storage, no cross-cluster branch.** This is the hard requirement; there is no fallback that copies data through the hub.
- **Multi-PVC (HA) sources depend on the backend for cross-cluster group consistency** — see the warning under Prerequisites.
- **The hub is on the critical path for create, refresh, and delete**, though not for the branch's ongoing operation once it is `Ready` — a hub outage pauses new branches, refreshes, and deletions, but a `Ready` cross-cluster branch keeps serving traffic without the hub.
- **Node-local storage cannot be cross-cluster branched**, for the same reason it cannot be branched at all: `hostPath`, `local-path`, and similar volumes have no CSI snapshot object to hand off between clusters.

## Next Steps

- Read the [PostgreSQL Branching Overview](/docs/guides/postgres/branch/overview/index.md) for the mechanics this page builds on.
- Try [Same-Cluster Branching](/docs/guides/postgres/branch/same-cluster/index.md) first — it is the same `Branch` object with none of the OCM prerequisites.
- See every field of the [Branch CRD](/docs/guides/postgres/concepts/branch.md), including `spec.target.clusterName`.
