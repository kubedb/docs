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
full backup or, for a Milvus that keeps its write-ahead log on the object storage, to **any point in time**
inside the recorded window.

## How it works

| Step | What happens |
|------|--------------|
| Full backup | KubeStash runs the `physical-backup` task. The plugin pauses Milvus garbage collection, denies writes (`forceDeny` in the etcd dynamic config, for at most `fenceTimeout`), flushes, dumps the whole etcd keyspace at **one revision** and copies every new object version into the archive. The write fence is lifted right after the dump. |
| Continuous archiving | A `Sidekick` runs `milvus-archiver archive` next to the database. It records every etcd change and every new object version (versioned object pool, tombstones for deletions) and publishes the recoverable window on the `<db>-incremental-snapshot` Snapshot (`status.components.wal.logStats.start/end`). |
| Restore | A new Milvus with `spec.init.archiver` is created. Before any Milvus process starts, a `RestoreSession` rebuilds the etcd image (base dump plus replay to the requested revision) and the objects as of that time. |

The archive lives in the `BackupStorage` of the archiver (S3 compatible storage), below
`<subDir>/<namespace>/<db>`. The change log and the base dumps are encrypted with a key derived from the
archiver's `encryptionSecret`.

## Point-in-time recovery needs the Woodpecker WAL

| Milvus | WAL | Base backup | Point-in-time recovery |
|--------|-----|:-----------:|:----------------------:|
| Standalone (default) | RocksMQ on the data PVC | yes | no (only the time of a full backup) |
| Standalone with `spec.wal.type: Woodpecker` | objects in the bucket | yes | yes |
| Distributed | Woodpecker | yes | yes |

For a RocksMQ standalone Milvus the RocksMQ directories are captured inside the write fence together with
the metadata (tar stream by the `physical-backup` task, or a CSI `VolumeSnapshot` with the `volume-snapshot`
task). The archiver records the condition `PITRUnsupported` on such a database.

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
    driver: Restic          # or VolumeSnapshotter for the data PVC of a RocksMQ standalone Milvus
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

A database opts in by carrying the archiver's label (double opt-in). For point-in-time recovery of a
standalone Milvus set the WAL type at creation; it cannot be changed later.

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
  wal:
    type: Woodpecker
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
  wal:
    type: Woodpecker
  objectStorage:
    configSecret:
      name: milvus-restore-storage   # empty bucket, same rootPath as the source
  init:
    archiver:
      recoveryTimestamp: "2026-09-30T16:22:29Z"
      encryptionSecret: {name: encrypt-secret, namespace: demo}
      fullDBRepository: {name: milvus-full, namespace: demo}
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

What the operator refuses, with a clear message on the `SuccessfullyDataRestored` condition:

| Situation | Behaviour |
|-----------|-----------|
| time before the oldest kept full backup | rejected, the earliest restorable time is reported |
| time after the end of the window | clamped to the end of the window (logged in the restore Job) |
| time inside a **log gap** (see below) | rejected, the nearest restorable times before and after the gap are reported |
| target bucket or etcd not empty | rejected, nothing is overwritten |
| target `rootPath`, `dmlChannelNum` or channel prefix differs from the source | rejected |
| target Milvus of another minor version, or an older patch, than the backup | rejected |

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

Set `fullBackup.driver: VolumeSnapshotter` and `fullBackup.task.params.volumeSnapshotClassName` to snapshot
the data PVC of a RocksMQ standalone Milvus inside the write fence, instead of copying the RocksMQ directories.
Metadata and objects are archived exactly as with the Restic driver. A restore creates the data PVC from the
VolumeSnapshot.

## Limitations

- Only S3 compatible `BackupStorage` providers can hold the archive.
- The restored Milvus must use the same `rootPath` and an empty bucket.
- Collections, indexes and users, roles and privileges are restored as they were; nothing is selectable.
- Backing up a Milvus that uses an external etcd with TLS/authentication requires the client secrets to be
  readable by the archiver service account.
- The write fence denies writes for about ten seconds per full backup.

## Cleanup

```bash
kubectl delete milvus -n demo milvus milvus-restored
kubectl delete milvusarchiver -n demo milvus-archiver
```
