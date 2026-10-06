---
title: Base Backup and Point-in-time Recovery
menu:
  docs_{{ .version }}:
    identifier: backup-milvus-archiver
    name: Overview
    parent: backup-milvus
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# KubeDB Milvus - Base Backup and Point-in-time Recovery

KubeDB backs up a running Milvus physically: it captures the metadata etcd and the object storage bucket
under one short write fence, and keeps recording every later metadata change and new object version. A
`MilvusArchiver` ties the pieces together. A new `Milvus` can be restored from the archive to the time of any
full backup or to **any point in time** inside the recorded window. This works for Standalone and Distributed
Milvus alike: KubeDB always runs Milvus with the Woodpecker write-ahead log, which lives in the object storage.

## How it works

| Step | What happens |
|------|--------------|
| Full backup | KubeStash runs the `physical-backup` task. The plugin pauses Milvus garbage collection, denies writes (`forceDeny` in the etcd dynamic config, for at most `fenceTimeout`), flushes, dumps the whole etcd keyspace at **one revision** and copies every new object version into the archive. The write fence is lifted right after the dump. |
| Continuous archiving | A `Sidekick` runs `milvus-archiver archive` next to the database. It records every etcd change and every new object version (versioned object pool, tombstones for deletions) and publishes the recoverable window on the `<db>-incremental-snapshot` Snapshot (`status.components.wal.logStats.start/end`). |
| Restore | A new Milvus with `spec.init.archiver` is created. Before any Milvus process starts, a `RestoreSession` rebuilds the etcd image (base dump plus replay to the requested revision) and the objects as of that time. |

The archive lives in the `BackupStorage` of the archiver (S3 compatible storage), below
`<subDir>/<namespace>/<db>`. The change log and the base dumps are encrypted with a key derived from the
archiver's `encryptionSecret`.

## The write-ahead log

Every Milvus that KubeDB creates uses the Woodpecker write-ahead log, stored as objects in the Milvus bucket
(`mq.type: woodpecker`, `woodpecker.storage.type: minio`). There is no option to choose another one, and a
configuration that sets a different `mq.type` or `woodpecker.storage.type: local` is refused, because the
archive must see the write-ahead log in the object storage.

| Milvus | Full backup | Point-in-time recovery | VolumeSnapshot (warm cache) |
|--------|:-----------:|:----------------------:|:---------------------------:|
| Standalone | yes | yes | yes |
| Distributed | yes | yes | no (use the Restic driver) |
| Standalone created by an older KubeDB (RocksMQ) | no | no | no |

### Existing Standalone Milvus (RocksMQ)

A Standalone Milvus created by a KubeDB release before the archiver keeps its write-ahead log in RocksMQ on its
data PVC. Milvus cannot switch the write-ahead log of a running instance without losing data that is not yet
flushed, so KubeDB leaves such a database exactly as it is. The operator marks it with the annotation
`kubedb.com/milvus-wal: rocksmq` (every other Milvus gets `woodpecker`); the annotation cannot be changed.

Such a database cannot be archived: `spec.archiver` is rejected, and if an archiver is attached anyway the Milvus
reports the condition `ArchiverUnsupported` and no backup is configured. To move it to the Woodpecker write-ahead log, take a
logical backup (KubeStash `milvus-addon`, task `logical-backup`) and restore it into a new Milvus
(task `logical-backup-restore`), then switch the clients.

## Before you begin

- KubeDB, KubeStash and Sidekick are installed, together with the KubeStash catalog that provides `milvus-addon`.
- The Milvus must run with `quotaAndLimits.enabled: true` (KubeDB renders this by default): the write fence relies on it.
- `kubectl create ns demo`

## Deploy the archiver

```yaml
apiVersion: archiver.kubedb.com/v1alpha1
kind: MilvusArchiver
metadata:
  name: milvus-archiver
  namespace: demo
spec:
  pause: false
  databases:
    namespaces:
      from: Same
    selector:
      matchLabels:
        archiver: "true"
  retentionPolicy: {name: milvus-retention, namespace: demo}
  encryptionSecret: {name: encrypt-secret, namespace: demo}
  backupStorage:
    ref: {name: s3-storage, namespace: demo}
    subDir: milvus
  deletionPolicy: WipeOut
  fullBackup:
    driver: Restic          # or VolumeSnapshotter (Standalone): also snapshots the data PVC as a warm cache
    task:
      params:
        quiesce: DenyWrites # DenyWrites | Flush | None
        fenceSettle: 10s
        fenceTimeout: 60s
    scheduler:
      schedule: "0 */6 * * *"
    sessionHistoryLimit: 3
  logBackup:
    retentionPeriod: 7d
    retentionSchedule: "0 1 * * *"
  manifestBackup:
    scheduler:
      schedule: "0 */6 * * *"
```

`quiesce` selects how writes are frozen for the crash-consistent capture:

- `DenyWrites` (default): writes and DDL are rejected for the length of the dump (about ten seconds including the settle time).
- `Flush`: no fence. The capture is only as consistent as a flush; choose it when `quotaAndLimits` cannot be enabled.
- `None`: no quiescing at all. Not recommended.

If `quotaAndLimits.enabled` is `false` the full backup fails with an explanatory message unless
`allowQuiesceDowngrade: "true"` is set, which falls back to `Flush`.

## Opt a Milvus in

A database opts in by carrying the archiver's label (double opt-in).

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Milvus
metadata:
  name: milvus
  namespace: demo
  labels:
    archiver: "true"
spec:
  version: "2.6.11"
  topology:
    mode: Standalone
  objectStorage:
    configSecret:
      name: milvus-storage
  storageType: Durable
  storage:
    accessModes: [ReadWriteOnce]
    resources:
      requests:
        storage: 5Gi
  deletionPolicy: WipeOut
```

KubeDB adds `spec.archiver.ref`, creates the `BackupConfiguration` and, after the **first successful full
backup**, the `Sidekick` and the incremental Snapshot:

```bash
$ kubectl get backupconfiguration,sidekick,snapshot -n demo
```

```bash
$ kubectl get snapshot -n demo milvus-incremental-snapshot -o jsonpath='{.status.components.wal.logStats}'
{"end":"2026-09-30T16:23:04Z","lsn":"rev:552","start":"2026-09-30T16:17:37Z"}
```

`start` is the fence time of the oldest full backup that is still kept and `end` is the latest point that can
be restored. `lsn` is the last archived etcd revision. The Milvus conditions `LogBackupLagging`,
`LogBackupDegraded` and `LogBackupGap` mirror the state of the archiver.

Changes to the archive policy do not interrupt the database: `spec.archiver.pause: true` on the Milvus (or
`spec.pause` on the archiver) stops the sidekick and pauses the schedules.

To stop archiving, remove the archiver label and `spec.archiver` from the Milvus. KubeDB then deletes the sidekick
and the retention CronJob and pauses the `BackupConfiguration`; the repositories and Snapshots stay, so the existing
backups can still be restored. Attaching an archiver again resumes the schedules. Deleting the `MilvusArchiver`
itself instead removes the `BackupConfiguration`, and the archiver's `deletionPolicy` then decides whether the
backups are deleted.

## Write fence safety

A full backup briefly denies writes (the fence). The backup lifts it itself, also when its pod is terminated
(SIGTERM) or when `fenceTimeout` expires. If the pod is killed without a chance to clean up, the fence keeps a
marker with a deadline (`fenceTimeout` plus one minute): the sidekick lifts an expired fence within 30 seconds, and
the next backup lifts it before it starts. Databases without a sidekick (VolumeSnapshotter archivers) recover at
the next scheduled backup.

## Restore

Create a new Milvus from the archive. The object storage of the new Milvus must be **empty** and use the same
`rootPath` as the source (the statistics of text indexes embed full object keys); a different bucket is fine.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Milvus
metadata:
  name: milvus-restored
  namespace: demo
spec:
  version: "2.6.11"
  topology:
    mode: Standalone
  objectStorage:
    configSecret:
      name: milvus-restore-storage   # empty bucket, same rootPath as the source
  init:
    archiver:
      recoveryTimestamp: "2026-09-30T16:22:29Z"
      encryptionSecret: {name: encrypt-secret, namespace: demo}
      fullDBRepository: {name: milvus-full, namespace: demo}
      manifestRepository: {name: milvus-manifest, namespace: demo}   # optional: keep the source's credentials
  storageType: Durable
  storage:
    accessModes: [ReadWriteOnce]
    resources:
      requests:
        storage: 5Gi
  deletionPolicy: WipeOut
```

KubeDB picks the newest full backup at or before `recoveryTimestamp`, replays the change log up to that time
and starts Milvus on the result. The condition `ArchiverRecoveryPlanned` names the chosen base. Distributed
targets use `topology.mode: Distributed` with the same sizes of the source.

With `manifestRepository`, KubeDB first restores the source's auth secret as the new Milvus's own auth secret
(`milvus-restored-auth` above, or the name set in `spec.authSecret.name`), before it would generate one. The
restored Milvus therefore accepts the source's credentials, and the source's secret is left untouched, so the
restore can run in the source's namespace. If that auth secret already exists when the restore starts, it is
kept and the source's credentials are not restored. Only the auth secret is restored. The rendered configuration is
generated for the new Milvus, and a custom configuration secret must be referenced in the new Milvus's spec.

What the operator refuses, with a clear message on the `SuccessfullyDataRestored` condition:

| Situation | Behaviour |
|-----------|-----------|
| time before the oldest kept full backup | rejected, the earliest restorable time is reported |
| time after the end of the window | clamped to the end of the window (logged in the restore Job) |
| time inside a **log gap** (see below) | rejected, the nearest restorable times before and after the gap are reported |
| target bucket or etcd not empty | rejected, nothing is overwritten |
| target `rootPath`, `dmlChannelNum` or channel prefix differs from the source | rejected |
| target Milvus of another minor version, or an older patch, than the backup | rejected |
| target Milvus runs another Woodpecker version than the backup | rejected |

Restored Milvus nodes register themselves again: the session keys of the source are not carried over and the
streaming node assignment of every channel is reset, so the restored Milvus never talks to the old pods.

## Retention

`logBackup.retentionPeriod` (`<n>[dwmy]`) bounds the archive. KubeDB runs it as the CronJob `<db>-log-retention` on `logBackup.retentionSchedule`. The retention job keeps the newest full backup
at or before the horizon and removes everything no restore at a later time needs: change-log segments, time
index files, superseded object versions and the older full backups with their Snapshots. A restore to a time
before the kept full backup is no longer possible afterwards.

## Log gaps

The change log is continuous only while every etcd revision is recorded. If the archiver is stopped for so
long that etcd compacts revisions it has not yet recorded, the archiver notices (`ErrCompacted`), records the
gap, annotates the incremental Snapshot (`archiver.kubedb.com/log-gap`), sets the `LogBackupGap` condition and
asks KubeStash for a new full backup. Restores to a time after that full backup work again; times inside the
gap cannot be restored. Increase the archiver resources or etcd's `--auto-compaction-retention` if gaps occur.

## VolumeSnapshot backups

Set `fullBackup.driver: VolumeSnapshotter` and `fullBackup.task.params.volumeSnapshotClassName` to also take a
CSI `VolumeSnapshot` of the data PVC of a **Standalone** Milvus. The snapshot is taken inside the same write
fence, right after the metadata dump. Metadata and objects are archived exactly as with the Restic driver, so
the archive alone is a complete backup.

The data PVC holds only the local cache of Milvus (loaded segments, mmap files and disk indexes). A restore from
such a backup creates the data PVC of the new Milvus from the VolumeSnapshot, so it starts with a **warm cache**
instead of downloading everything again. The condition `WarmCacheRestored` reports whether the snapshot was used.
If it cannot be used (deleted, not ready, or in another namespace than the restored Milvus) the restore still
succeeds and Milvus starts with a cold cache.

Requirements:

- the Milvus data PVC uses a CSI `StorageClass` (`spec.storage.storageClassName`) and a matching
  `VolumeSnapshotClass` exists;
- a failed snapshot fails the full backup session, so the backup you asked for is never silently incomplete;
- a Distributed Milvus has no PVC worth snapshotting: attached to an archiver with the VolumeSnapshotter driver it
  reports the condition `ArchiverUnsupported` and is not backed up; use the Restic driver.

## Limitations

- Only S3 compatible `BackupStorage` providers can hold the archive.
- The restored Milvus must use the same `rootPath` and an empty bucket.
- Collections, indexes and users, roles and privileges are restored as they were; nothing is selectable.
- Backing up a Milvus that uses an external etcd with TLS/authentication requires the client secrets to be
  readable by the archiver service account.
- The write fence denies writes for about ten seconds per full backup.
- A Standalone Milvus created by a KubeDB release before the archiver (RocksMQ write-ahead log) cannot be
  archived; migrate it with a logical backup.

## Cleanup

```bash
kubectl delete milvus -n demo milvus milvus-restored
kubectl delete milvusarchiver -n demo milvus-archiver
```
