---
title: PostgreSQL Branching Overview
description: How KubeDB branches a running PostgreSQL database using CSI volume snapshots
menu:
  docs_{{ .version }}:
    identifier: guides-pg-branch-overview
    name: Overview
    parent: guides-pg-branch
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="Branching is driven by the KubeDB Courier operator, which is not installed by default. Enable it with `--set kubedb-courier.enabled=true` when you [install KubeDB](/docs/setup/README.md). Branching also requires a CSI driver that supports volume snapshots — see Requirements below." >}}

# PostgreSQL Branching Overview

**Branching** creates a full, writable copy of a running KubeDB `PostgreSQL` database in seconds — the way you would branch a Git repository. The copy, called a *branch*, is an ordinary KubeDB `Postgres` database with its own Pods, Service, and credentials. You can write to it, break it, and throw it away without ever touching the database it came from.

A branch is not a logical dump and restore. KubeDB takes a **CSI volume snapshot** of the source database's data volume and clones the branch's volume from that snapshot. On a copy-on-write driver the clone is near-instant and initially shares its blocks with the source, so the snapshot and clone steps take about as long for a 500 GB database as for a 1 GB one, and the branch starts out consuming close to zero extra storage. It only consumes space as it diverges from the source. What remains of the total time is the branched database starting up — the same startup any KubeDB database goes through.

Branching is driven by a single custom resource, `Branch`, in the `courier.kubedb.com` API group, and is reconciled by the **KubeDB Courier** operator.

## Why It Matters

- **Production-like data in seconds.** Developers, QA, and CI get a real copy of the production dataset instead of a stale, hand-curated fixture — without waiting for a multi-hour dump and restore.
- **Cheap.** Copy-on-write means you pay for what changes, not for another full copy of the database.
- **The source is never touched.** KubeDB never pauses, halts, locks, or writes to the source database. It stays online and serving traffic for the whole operation.
- **Fully isolated.** The branch is a separate `Postgres` object with its own volumes, Service, and (optionally) its own root credential. Writes on either side are invisible to the other.
- **Safe to hand out.** Optional *post-actions* run right after the clone comes up, so sensitive columns can be stripped or masked before the branch is ever reported `Ready`.
- **Stays current.** An optional cron schedule re-clones the branch from a fresh source snapshot, so a long-lived dev environment does not drift months behind production.

## Common Use Cases

- Spinning up a dev or QA environment that mirrors production.
- Rehearsing a schema migration or a risky `UPDATE` against real data and real data volume.
- Reproducing a production bug against the data as it stands right now, without touching production. (A branch copies the *current* state — to go back to an earlier point in time, use [Point-in-Time Recovery](/docs/guides/postgres/pitr/archiver.md) instead.)
- Seeding CI pipelines with a realistic database.
- Giving analysts a sandbox that cannot affect production.

## How Branching Works

The following diagram shows how KubeDB creates a branch of a `PostgreSQL` database. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="PostgreSQL Branching Overview" src="/docs/guides/postgres/branch/overview/images/branch_overview.svg">
  <figcaption align="center">Fig: PostgreSQL Branching Overview</figcaption>
</figure>

The branching process consists of the following steps:

1. A user creates a `Branch` custom resource that names the **source** KubeDB `Postgres` database and the **target** database to create — its name, namespace, StorageClass, and resource requests.

2. The KubeDB Courier operator watches for `Branch` objects. When it finds one, it resolves the source `Postgres` database and lists its data PVCs — one for a standalone database, one per replica for an HA cluster.

3. Courier takes a **CSI `VolumeSnapshot` of each source data PVC**. When the CSI driver provides a `VolumeGroupSnapshotClass`, all the volumes of an HA cluster are captured as one crash-consistent group; otherwise Courier falls back to one snapshot per PVC. The source database keeps serving reads and writes throughout.

4. Once every snapshot reports `readyToUse`, Courier **clones the target PVCs** from them, setting each new PVC's `dataSource` to its snapshot. The CSI driver serves the clone from the snapshot; on a copy-on-write driver no data is physically copied.

5. Courier then creates the **target `Postgres` custom resource**. By default its spec is inherited from the source — version, replica count, TLS settings, pod template — and a custom `spec.configSecret` is copied into the target namespace. Three things change that default: the fields you set under `spec.target` (name, namespace, StorageClass, resources, TLS issuer), `spec.resetRootPassword`, which gives the branch a generated credential instead of the source's, and `spec.postActions`, which mutates the cloned data before the branch is `Ready`.

6. The KubeDB provisioner reconciles that `Postgres` object in *branched* mode: instead of provisioning empty volumes, it **adopts the PVCs Courier already cloned**, and PostgreSQL starts on top of the cloned data directory. For an HA source, each cloned volume is matched to its ordinal (`0 → 0`, `1 → 1`, `2 → 2`) and the cluster re-forms as a primary with streaming standbys.

7. If `spec.postActions` is set, Courier runs those containers as Jobs against the branch — with the branch's connection details injected as environment variables — and only marks the branch `Ready` after every one of them has succeeded.

When the `Branch` reaches `Ready`, the branched database is an ordinary KubeDB `Postgres`: it has its own `AppBinding`, Services, and auth Secret, and it can be scaled, reconfigured, backed up, and monitored like any other.

## Same-Cluster Branching

This guide covers **same-cluster** branching: the source database and the branch live in the same Kubernetes cluster. The branch may be created in the same namespace as the source or in a different one — when the namespaces differ, Courier mirrors the source snapshot into the target namespace so the clone can reference it.

Courier reports this in `status.mode`:

| Mode | Meaning |
|---|---|
| `Local` | Same-cluster branch — this operator runs the whole flow. This is the mode covered here. |
| `Initiator` / `Creator` | Cross-cluster branching roles. |

Leaving `spec.target.clusterName` empty selects a same-cluster branch.

## What the Branch Inherits

| Carried over from the source | Controlled by the `Branch` |
|---|---|
| PostgreSQL version | Target name and namespace (`spec.target.name`, `spec.target.namespace`) |
| Replica count and topology | StorageClass (`spec.target.storageClassName`) |
| Custom configuration (`spec.configSecret`) — copied into the target namespace | CPU/memory requests and limits (`spec.target.resources`) |
| TLS settings | TLS issuer for the branch (`spec.target.issuerRef`) — a branch cannot reuse the source's certificates |
| Pod template, monitoring, and other spec fields | Root credential: reuse the source's, or generate a fresh one (`spec.resetRootPassword`) |
| The data itself, as of the snapshot | Post-clone data treatment (`spec.postActions`) |

The branch's `deletionPolicy` is always `Delete` on the target `Postgres` object — the `Branch` owns it. What happens when you delete the `Branch` itself is governed by `spec.deletionPolicy` on the `Branch` (see below).

The target database is stamped with a `kubedb.com/branched-from` annotation recording the source it came from:

```bash
$ kubectl get pg -n demo dev-postgres -o jsonpath='{.metadata.annotations.kubedb\.com/branched-from}'
{"cluster":"","source":"demo/sample-postgres"}
```

## Branch Lifecycle

A `Branch` reports its progress through `status.phase`:

| Phase | What is happening |
|---|---|
| `Pending` | The `Branch` has been accepted and is about to start. |
| `Snapshotting` | Courier is taking the CSI snapshot(s) of the source data volumes. |
| `Cloning` | The target PVCs are being cloned from those snapshots. |
| `Provisioning` | The target `Postgres` has been created and KubeDB is bringing it up on the cloned volumes. |
| `ActionsRunning` | The `spec.postActions` Jobs are running against the branch. |
| `Ready` | The branch is up and usable. |
| `Refreshing` | A scheduled refresh is re-cloning the branch from a new source snapshot. |
| `Failed` | Something went wrong; see `status.conditions` and `status.history`. |
| `Deleting` | Teardown is in progress. |

Alongside the phase, `status.conditions` records each milestone:

`SnapshotReady` → `TargetCreated` → `TargetReady` → [`RootPasswordReset`] → [`PostActionsCompleted`] → `Ready`

`status.resources` lists everything the branch owns (its auth Secret, cloned PVCs, and current post-action Job), and `status.snapshot` records the snapshot set backing the current copy — the strategy used, and one entry per source volume with its size and readiness.

## Keeping a Branch Fresh

A branch is a point-in-time copy: it does not follow the source. To keep a long-lived branch from drifting, set a cron schedule:

```yaml
spec:
  schedule:
    cron: "0 2 * * *"   # re-clone from the source every night at 02:00
```

At each tick Courier takes a **fresh source snapshot first**, then halts the branch, discards its cloned volumes, re-clones them from the new snapshot, brings the branch back up, and re-runs any post-actions on the new data. Taking the snapshot before halting anything means a failure at that stage leaves the branch running on its previous copy.

{{< notice type="warning" message="A refresh replaces the branch's data. Anything written to the branch since the last refresh is discarded, and the branch is briefly unavailable while it is re-cloned. Omit `spec.schedule` for a one-shot branch that is never refreshed." >}}

`status.lastSuccessfulRefreshTime` records how current the data is, and is surfaced as the `FRESHNESS` column:

```bash
$ kubectl get branch -n demo
NAME         PHASE   MODE    TARGET         FRESHNESS   AGE
dev-branch   Ready   Local   dev-postgres   3m12s       12m
```

`spec.historyLimit` bounds the run history kept in `status.history` (3 successful and 2 failed runs by default).

## Deleting a Branch

`spec.deletionPolicy` decides what happens to the branched database when the `Branch` object is deleted:

| Policy | Effect |
|---|---|
| `Delete` (default) | Tears the branch down completely: the target `Postgres`, its cloned PVCs, the auth Secret, the snapshots, and any post-action Jobs. The source is untouched. |
| `Orphan` | Keeps the branched database as a standalone KubeDB `Postgres`. Its ownership metadata is stripped and its auth Secret is retained, so it keeps working — it is simply no longer managed by a `Branch`. |

## Failure Handling

If something fails, the `Branch` goes to `Failed` with a `Failed` condition and an entry in `status.history`. Whether it retries depends on what failed:

- **A post-action failure is terminal.** The same container would fail the same way on an unchanged spec, so Courier suspends the Job — keeping the failed Pod and its logs for you to read — and waits. Fix the `postActions` spec and re-apply; the edit triggers a retry.
- **A one-shot branch (no `spec.schedule`) holds `Failed`.** Nothing external will fix it, so it waits for you to correct the spec.
- **A scheduled branch retries transient failures on its own** about a minute later, starting again from `Snapshotting`. This matters because a refresh can fail after the branch has been halted, and abandoning it would leave the database down.

## Requirements

Branching is built on CSI volume snapshots, so the storage layer has to support them:

- The **KubeDB Courier** operator must be installed (`--set kubedb-courier.enabled=true`). Branching currently supports KubeDB `PostgreSQL` databases.
- The source database's volumes must be provisioned by a **CSI driver that supports volume snapshots** (for example AWS EBS, GCE PD, Ceph RBD, TopoLVM with thin provisioning, Longhorn, or Rook). The [external-snapshotter](https://github.com/kubernetes-csi/external-snapshotter) CRDs and controller must be installed.
- A **`VolumeSnapshotClass`** must exist for that driver. Courier uses the driver's default class unless you pin one with `spec.volumeSnapshotClassName`.
- The target StorageClass must be backed by the **same CSI driver** as the source — a snapshot can only be restored by the driver that created it.
- Branching an HA cluster as one crash-consistent group additionally requires a **`VolumeGroupSnapshotClass`** for the driver. Without one, Courier falls back to per-PVC snapshots taken back to back.

{{< notice type="warning" message="Snapshot support is a property of the CSI driver, not of Kubernetes. `hostPath`, `local-path`, and NFS-backed StorageClasses generally cannot be snapshotted and cannot be branched." >}}

## Branching vs. Backup vs. Migration

KubeDB has three features that copy data, and they answer different questions:

| Feature | Question it answers | Copy is | Source |
|---|---|---|---|
| **Branching** | "Give me a throwaway copy of production to develop against." | A live, writable database, ready in seconds | A KubeDB-managed database in your cluster |
| [Backup & Restore](/docs/guides/postgres/backup/kubestash/overview/index.md) | "If we lose this database, how do we get it back?" | Data at rest in object storage, restorable on demand | Any KubeDB-managed database |
| [Migration](/docs/operatormanual/migration/) | "How do we move an existing database into KubeDB?" | The same database, relocated, with minimal downtime | An external instance (RDS, CNPG, self-hosted, …) |

Branching is not a backup: the clone shares storage blocks with the source, so it is not an independent durable copy and it does not protect you from losing the source. Use [KubeStash backups](/docs/guides/postgres/backup/kubestash/overview/index.md) for durability, and branching for disposable copies.

## Next Steps

- Follow the hands-on walkthrough: [Branch a PostgreSQL Database in the Same Cluster](/docs/guides/postgres/branch/same-cluster/index.md).
- Then customize it: [Customizing a PostgreSQL Branch](/docs/guides/postgres/branch/customization/index.md) — separate credential, post-actions, scheduled refresh, cross-namespace targets, HA, snapshot-class pinning.
- See every field of the [Branch CRD](/docs/guides/postgres/concepts/branch.md).
