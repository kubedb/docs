---
title: Etcd In-place Restore Overview
menu:
  docs_{{ .version }}:
    identifier: etcd-restore-overview
    name: Overview
    parent: etcd-restore
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

<!-- Drafted from source code and CRD schemas; not yet verified against a live cluster. -->

> New to KubeDB? Please start [here](/docs/README.md).

# Etcd In-place Restore

This guide gives an overview of how the KubeDB Ops-manager operator restores a
[KubeStash](https://kubestash.com/) snapshot into an `Etcd` cluster that **already exists**, using an
`EtcdOpsRequest` of type `Restore`.

> **This operation is fully destructive to the database's current contents.** The entire existing
> keyspace is replaced by the snapshot, and every member's current data directory is discarded. Full
> data loss is not a failure mode here — it is the normal, successful outcome of the operation. There
> is no undo.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Etcd](/docs/guides/etcd/concepts/etcd.md)
  - [EtcdOpsRequest](/docs/guides/etcd/concepts/etcdopsrequest.md)
  - [EtcdArchiver](/docs/guides/etcd/concepts/etcdarchiver.md)
  - [Backup & Restore with KubeStash](/docs/guides/etcd/backup/kubestash/overview/index.md)
  - [Horizontal Scaling Overview](/docs/guides/etcd/scaling/horizontal-scaling/overview.md) — the
    learner-add/promote mechanism this restore relies on to regrow the cluster.

## Two Kinds of Restore

KubeDB can restore an etcd snapshot in two different situations, and they are **not**
interchangeable. Pick the one that matches what you have:

| | Bootstrap-time restore | In-place restore (this guide) |
|---|---|---|
| Where you declare it | `spec.init.archiver` on a **new** `Etcd` object | an `EtcdOpsRequest` of type `Restore` |
| When it works | only while the `Etcd` object is being created | on an `Etcd` that already exists, healthy or degraded |
| What happens to existing data | there is none | **all of it is replaced by the snapshot** |
| Guide | [Snapshot Backup & Restore](/docs/guides/etcd/backup/kubestash/snapshot/index.md#restore) | [In-place Restore walkthrough](/docs/guides/etcd/restore/restore.md) |

If you are standing up a **new** cluster from a backup — a copy in another namespace, a clone for
testing, a rebuild after the original object was deleted — use the bootstrap path. It is strictly
safer, because there is nothing to lose.

Use the in-place restore when the `Etcd` object has to survive: everything that references it (the
`AppBinding`, the connection Secret, the archiver, the autoscaler, the applications that resolve its
Service name) keeps working, because the object itself never goes away. Before this ops request
existed, the only way to restore into an existing cluster was to delete the `Etcd` object and
recreate it from the backup — which loses everything else about it.

## When to Use It

- The keyspace itself is the problem: data was deleted or corrupted by an application, and you want
  a known-good copy back.
- The cluster is degraded past the point of rescue — for example every member's volume is gone, so
  there is no surviving member to rebuild from and
  [`RecoverFromQuorumLoss`](/docs/guides/etcd/recover-from-quorum-loss/overview.md) has nothing to
  work with.

Do **not** reach for it when the data is fine and only the cluster is unhealthy. A cluster that has
lost its quorum but still has one member holding good data should be rebuilt with
[`RecoverFromQuorumLoss`](/docs/guides/etcd/recover-from-quorum-loss/overview.md), which keeps the
survivor's data instead of discarding it and does not roll you back to the last backup.

## Why the Cluster Has to Be Taken Apart

There is no way to "restore into" a running etcd. A member's data directory has to *already be* a
valid restored snapshot **before** the etcd process starts — restoring a snapshot is a filesystem
operation on the data directory, not an RPC the cluster serves.

So the restore deliberately reuses the bootstrap mechanism rather than inventing a second one: the
cluster is taken apart down to a single empty volume, the snapshot is written into that volume by
exactly the same KubeStash `RestoreSession` the bootstrap path uses, and the cluster is started
again as a one-member cluster.

**Only the seed member (ordinal 0) is ever restored into.** The other members are not restored
independently — they are discarded and rejoin as fresh
[learners](/docs/guides/etcd/scaling/horizontal-scaling/overview.md#learners), streaming the restored
keyspace from the seed through etcd's own membership API. That pulls the snapshot out of the
repository exactly once, touches one volume instead of all of them, and reuses a mechanism the
membership reconciliation already runs on every scale up. Keeping their old data directories would
be worse than useless: they describe a keyspace the restored one no longer relates to, so they could
not rejoin anyway.

## How the Restore Process Works

Every step is gated by its own condition on the ops request, so an operator restart resumes exactly
where it left off rather than starting over.

1. The user creates an `EtcdOpsRequest` of type `Restore` with `spec.restore.fullDBRepository` and a
   `spec.timeout`.

2. The Ops-manager operator pauses the `Etcd` object, so the Provisioner operator does not fight the
   restore.

3. The `PetSet` is deleted **with an orphan propagation policy** — it would otherwise recreate every
   pod the restore deletes and re-provision every claim it discards.

4. **Every member except ordinal 0 is discarded**, one per pass: the pod is taken down (tolerating
   members whose pod object is already gone, which a degraded cluster often has) and its claim is
   dropped.

5. **The seed volume is wiped.** Ordinal 0 is taken down, its claim is deleted, the operator waits
   for the claim to actually disappear (a claim still held by the `pvc-protection` finalizer would
   be adopted — data and all — instead of replaced), and a fresh empty claim is created under the
   same name, specced exactly as the `PetSet`'s own volume claim template would have produced it.

6. **The snapshot is applied.** The operator resolves which `Snapshot` to use from the repository,
   creates a KubeStash `RestoreSession` named `<ops-request-name>-snapshot-restorer` against the
   empty seed claim, and waits for it. The session is created once and never rewritten, so a backup
   that completes while the restore is running cannot retarget a restore already under way.

7. The cluster-state ConfigMap is rewritten for a **single member** cluster — the same call the
   provisioner makes for a bootstrap restore, because the situation it has to describe is the
   identical one.

8. The `PetSet` is recreated with one replica, which brings the seed member back on top of the
   restored volume, and the operator waits until it is `Ready` and answering.

9. The reclaim policies parked while the old volumes were being discarded are restored — **this is
   what finally destroys the data the restore replaced**, and it is deliberately left until the
   restored member is up and healthy, so a crash or a failed `RestoreSession` has thrown nothing
   away that a human could still go back to.

10. The `Etcd` object is resumed and the request is marked `Successful`.

Nothing in the restore regrows the cluster. Once the database is resumed, the ordinary membership
reconciliation sees one member against `spec.replicas` and adds the rest back through **exactly the
same learner-add/promote path an ordinary scale up uses**.

## Preconditions

- **`spec.restore.fullDBRepository` is required, and has no default.** The bootstrap restore can
  afford to be lenient because it has nothing to lose; this one replaces the whole keyspace of a
  live database, so there is no safe repository to guess at. The request fails immediately without
  it.
- **`spec.timeout` is required.** Several steps — waiting for the `RestoreSession`, waiting for the
  restored member to come up — have no other deadline at all. Size it against how large the snapshot
  is and how fast the backend storage is, not against how long a pod takes to restart.
- `spec.storageType` must be `Durable` and `spec.storage` must be set. There is no PVC to restore
  into before the seed member starts on an ephemeral cluster.
- The repository must hold at least one **successful full** `Snapshot` for this database. Snapshots
  are matched in the database's own namespace by repository name.
- `spec.restore.recoveryTimestamp` is **optional**. Leave it out and the latest available snapshot
  is used; set it and the operator picks the newest snapshot taken **at or before** that instant.
- `spec.restore.encryptionSecret` must name the Secret holding the encryption key the snapshot was
  backed up with, if the backup was encrypted.

In the [next](/docs/guides/etcd/restore/restore.md) doc, we are going to show a step-by-step guide on
restoring a snapshot into an existing Etcd cluster using the `EtcdOpsRequest` CRD.
