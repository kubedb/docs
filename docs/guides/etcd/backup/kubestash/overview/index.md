---
title: Backup & Restore Etcd Overview | KubeStash
description: Overview of how KubeDB backs up and restores an Etcd cluster using KubeStash
menu:
  docs_{{ .version }}:
    identifier: guides-etcd-backup-overview-stashv2
    name: Overview
    parent: guides-etcd-backup-stashv2
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

<!-- Drafted from source code and CRD schemas; not yet verified against a live cluster. -->

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="Please install [KubeStash](https://kubestash.com/docs/latest/setup/install/kubestash/) to try this feature. Database backup with KubeStash is already included in the KubeDB license. So, you don't need a separate license for KubeStash." >}}

# Etcd Backup & Restore Overview

KubeDB uses [KubeStash](https://kubestash.com) to back up and restore an `Etcd` cluster. KubeStash is
AppsCode's cloud native data backup and recovery solution for Kubernetes workloads and databases. It
uses [restic](https://github.com/restic/restic) under the hood to store backup data — encrypted and
deduplicated — in any cloud or on-prem storage backend (S3, GCS, Azure Blob, Minio, NetApp, Dell EMC
and so on).

## Snapshot-only: there is no point-in-time recovery

This is the single most important thing to understand before you design a backup strategy for etcd
in KubeDB:

**Backing up an `Etcd` cluster is snapshot-only. There is no continuous archiving and there is no
point-in-time recovery (PITR).**

Databases such as PostgreSQL and MySQL ship a continuous stream of write-ahead-log segments to the
backend alongside their periodic base backups, which lets you replay the log forward and land on an
arbitrary point in time. etcd exposes no comparable primitive — there is no supported way to stream
its raft WAL out of the cluster and replay it into a rebuilt data directory. KubeDB therefore
deliberately scopes etcd backup to whole-cluster snapshots:

- The `EtcdArchiver` CRD has a `spec.fullBackup` section and a `spec.manifestBackup` section, and
  **no `spec.logBackup` section**. The field simply does not exist in the API.
- The generated `BackupConfiguration` has at most two sessions, both of which produce full
  snapshots. There is no third, incremental session.
- Restoring rolls the cluster back to the state captured by one completed snapshot — the most
  recent one, or an older one you select. Any write accepted after that snapshot was taken is lost.

Set your backup schedule accordingly: with snapshot-only backup, your worst-case data loss window is
the interval between two scheduled backups.

## The moving parts

| Object | API group | Who creates it | What it is for |
|---|---|---|---|
| `BackupStorage` | `storage.kubestash.com/v1alpha1` | you | Where the backup data lives (bucket, prefix, credentials `Secret`). |
| `RetentionPolicy` | `storage.kubestash.com/v1alpha1` | you | How many old snapshots to keep. |
| `EtcdArchiver` | `archiver.kubedb.com/v1alpha1` | you | The backup *policy*: which storage, which schedules, which encryption secret, which databases may consume it. |
| `Etcd` | `kubedb.com/v1alpha2` | you | The database. Opts into an archiver through `spec.archiver.ref`. |
| `BackupConfiguration` | `core.kubestash.com/v1alpha1` | the KubeDB provisioner | The concrete, per-database backup definition derived from the archiver. **You do not write this by hand.** |
| `Repository`, `BackupSession`, `Snapshot` | `storage.kubestash.com` / `core.kubestash.com` | the KubeStash operator | The per-run bookkeeping. |
| `RestoreSession` | `core.kubestash.com/v1alpha1` | the KubeDB provisioner | Created at bootstrap time when `spec.init.archiver` is set. **You do not write this by hand either.** |

The important structural difference from a hand-rolled KubeStash setup is that for KubeDB databases
you author an **`EtcdArchiver` policy**, not a `BackupConfiguration`. The KubeDB provisioner owns the
`BackupConfiguration` and keeps it in sync with the archiver.

## How backup is wired up

1. You create a `Secret` holding the backend credentials and a `BackupStorage` that points at the
   bucket.
2. You create an `EtcdArchiver` referencing that `BackupStorage` through
   `spec.backupStorage.ref`, plus a `RetentionPolicy` and an encryption `Secret`.
3. The archiver declares who may consume it in `spec.databases`, which is the standard KubeDB
   double opt-in selector — `spec.databases.namespaces.from` is one of `Same`, `All` or `Selector`,
   optionally narrowed further by `spec.databases.selector`.
4. An `Etcd` object opts in from the other side. Either you set `spec.archiver.ref` explicitly:

    ```yaml
    spec:
      archiver:
        ref:
          name: etcd-archiver
          namespace: demo
    ```

   or you leave `spec.archiver` unset, in which case the provisioner scans every `EtcdArchiver` in
   the cluster, picks the highest-priority one whose `spec.databases` selector admits this database
   (same namespace beats same project namespace beats anything else) and patches its reference into
   `spec.archiver.ref` for you. That is the *double* opt-in: the archiver advertises who may consume
   it, and the database is only ever auto-attached to an archiver that admits it. An explicit
   `spec.archiver.ref` you write yourself is always honoured as-is.
5. Once both sides match, the KubeDB provisioner creates a `BackupConfiguration` named
   `<db-name>-archiver` in the database's namespace, owned by the `Etcd` object, targeting the
   `Etcd` object itself.

### The two sessions

The generated `BackupConfiguration` carries up to two sessions, one per section of the archiver
spec. If a section is omitted from the archiver, its session is not generated; if both are omitted,
no `BackupConfiguration` is created at all.

| Session | Generated from | Addon task(s) | Repository | What it captures |
|---|---|---|---|---|
| `full-backup` | `spec.fullBackup` | `etcd-backup`, then `manifest-backup` | `<db-name>-full` | A live snapshot of the etcd keyspace, plus the Kubernetes manifests. |
| `manifest-backup` | `spec.manifestBackup` | `manifest-backup` | `<db-name>-manifest` | Only the Kubernetes manifests (`Etcd` object, auth `Secret`, and so on). |

Both tasks come from the `etcd-addon` `Addon` object shipped by the KubeDB installer. The
`etcd-backup` task runs the `etcd-restic-plugin` image, which connects to the cluster's client
endpoint on port `2379` using the credentials from the `AppBinding`, streams a consistent snapshot
of the keyspace over the client API, and pipes it straight into the restic repository. Because the
snapshot is streamed rather than staged, the backup `Job` does not need a volume large enough to
hold the whole keyspace.

The `manifest-backup` task is the shared, cross-database `kubedbmanifest-backup` function that every
KubeDB database reuses unchanged — nothing about it is etcd specific.

Each session writes into `<spec.backupStorage.subDir>/<db-namespace>/<db-name>/<session-name>` in
the bucket, and the snapshot data lands under the `dump` component of the resulting `Snapshot`
object.

A few archiver fields worth calling out:

- `spec.pause: true` sets `spec.paused` on the generated `BackupConfiguration`. Scheduled sessions
  stop firing, but the repositories and their snapshots are kept.
- `spec.fullBackup.driver` exists for structural parity with the other KubeDB archivers and is
  **ignored for etcd**. An etcd snapshot is always taken through the client API into a restic
  repository; there is no `VolumeSnapshotter` variant to choose between.
- `spec.deletionPolicy` controls what happens to the created repositories when the
  `BackupConfiguration` goes away. It defaults to `Retain`.

## How restore is wired up

Restore for etcd works differently from most KubeDB databases, and the difference is not cosmetic.

**A snapshot can never be poured into a running member.** etcd's bootstrap semantics require a
member's data directory to be either genuinely empty — so the member can start a brand new cluster
with `--initial-cluster-state=new` — or a fully rebuilt, valid snapshot directory. Injecting data
into a member that is already serving raft traffic is not a supported operation, and there is no
"restore task" you can point at a live cluster.

The path described below is therefore the **bootstrap-time** restore: you do not create a
`RestoreSession`; you create a *new* `Etcd` object with `spec.init.archiver` filled in, and the
provisioner does the rest before the first pod ever starts:

1. You create a new `Etcd` object whose `spec.init.archiver.fullDBRepository` (and optionally
   `.manifestRepository`) points at the repositories produced by the backup.
2. The provisioner sees `spec.init.archiver`, has not yet recorded the
   `DatabaseSuccessfullyRestored` condition, and so **withholds the PetSet entirely**. The `Etcd`
   object reports `status.phase: DataRestoring`.
3. If `manifestRepository` is set, the provisioner first creates a `RestoreSession` named
   `<db-name>-manifest-restorer` running the shared `manifest-restore` task, so that the auth
   `Secret` and friends exist before any pod could reference them.
4. The provisioner then selects the snapshot to restore (see below), pre-creates the seed member's
   `PersistentVolumeClaim` — named `data-<db-name>-0`, exactly what the PetSet's volume claim
   template would have produced for ordinal 0 — and creates a second `RestoreSession` named
   `<db-name>-0-snapshot-restorer`. That session targets the **PVC**, not a pod, and runs the
   `etcd-backup-restore` task, which rebuilds a complete etcd data directory inside the volume.
5. Only after that `RestoreSession` reaches `Succeeded` does the provisioner set the
   `DatabaseSuccessfullyRestored` condition and create the PetSet. Member 0 boots on the restored
   data directory; every other member joins through the normal membership path and streams its copy
   of the data from the leader, so only ordinal 0 is ever restored into.

Once the condition is set the gate is permanently open — the provisioner will not re-run a restore
over a live cluster, which is exactly what you want.

### Restoring into an Etcd that already exists

When the `Etcd` object has to survive — because the `AppBinding`, the connection Secret and every
application pointing at its Service must keep working — the bootstrap path is not available, and
deleting and recreating the object loses everything else about it. That is what the `Restore`
[EtcdOpsRequest](/docs/guides/etcd/concepts/etcdopsrequest.md#specrestore) is for. It obeys the same
rule as above rather than breaking it: the cluster is taken apart down to a single **empty** volume,
the very same `etcd-backup-restore` `RestoreSession` writes the snapshot into that volume while
nothing is running, and the seed member is then started on the restored data directory. Every other
member is discarded and rejoins as a learner.

The trade-off is severe and unavoidable: an in-place restore **replaces the entire keyspace of the
live database**. See [In-place Restore](/docs/guides/etcd/restore/overview.md).

Because `spec.storageType: Ephemeral` gives the provisioner no PVC to restore into ahead of the
pods, restoring from a snapshot requires `spec.storageType: Durable` with `spec.storage` set. The
provisioner rejects the combination explicitly.

### Choosing which snapshot to restore

`spec.init.archiver.recoveryTimestamp` is present on the shared `ArchiverRecovery` type, but for
etcd it does **not** mean "replay to this instant". With no log stream to replay, it degrades to a
selector over the completed snapshots:

- Leave it unset (zero) and the provisioner restores the **newest** successful full `Snapshot` in
  the repository.
- Set it and the provisioner restores the newest successful full `Snapshot` that was completed **at
  or before** that timestamp. If no snapshot qualifies, the restore fails with an explicit error
  rather than silently picking a later one.

"Completed" here is the restic summary end time of the `dump` component where available, because
`Snapshot.status.snapshotTime` records when the session *started*, not when the data was consistent.

### What the restore task rebuilds

The `etcd-backup-restore` task does a server-side data directory rebuild. Along with the snapshot
itself, it writes the bootstrap identity of the seed member into the new data directory — the member
name, the `--initial-cluster` membership, the cluster token and the advertised peer URLs. All four
have sensible defaults derived from the `Etcd` object (the ordinal-0 pod name, the seed member
alone, the offshoot name, and the ordinal-0 pod's peer URL respectively), so you normally do not
have to think about them. It also runs an etcd version-compatibility check and a free-space
pre-flight check against the target volume before touching anything.

## Next Steps

- Walk through a complete backup and restore in
  [Snapshot Backup & Restore of Etcd using KubeStash](/docs/guides/etcd/backup/kubestash/snapshot/index.md).
- Detail concepts of the [EtcdArchiver](/docs/guides/etcd/concepts/etcdarchiver.md) object.
- Detail concepts of the [Etcd](/docs/guides/etcd/concepts/etcd.md) object.
- Detail concepts of the [AppBinding](/docs/guides/etcd/concepts/appbinding.md) object.
