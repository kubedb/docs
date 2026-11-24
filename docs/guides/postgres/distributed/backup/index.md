---
title: Backup, Archiving and PITR of a DC-DR Postgres
menu:
  docs_{{ .version }}:
    identifier: guides-postgres-distributed-backup
    name: Backup and PITR
    parent: guides-postgres-distributed
    weight: 74
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Backup, Archiving and PITR of a DC-DR Postgres

A DC-DR Postgres is backed up with KubeStash and a `PostgresArchiver`, like a single-cluster Postgres with [continuous archiving](/docs/guides/postgres/pitr/archiver.md): full backups on a schedule, continuous WAL archiving, and point-in-time recovery (PITR) into a new database. This page explains what is different on DC-DR, and the limitations you must know before you rely on it.

> **Note:** YAML files used in this tutorial are stored in [docs/guides/postgres/distributed/backup/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/postgres/distributed/backup/yamls) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## How it works on DC-DR

KubeStash and the KubeDB operators run on the hub; the database pods run on the spokes. So the work that touches the database runs on the spoke of the **active** data center:

- **Full backup.** The backup Job that KubeStash starts on the hub runs as the ServiceAccount `<db>-backup`. It creates a ManifestWork (`<namespace>-<db>-physical-backup`) in the active data center's cluster namespace, which runs the actual physical backup Job beside the database on that spoke.
- **WAL archiving.** The WAL archiver (wal-g, in a Sidekick named `<db>-sidekick`) runs on the spoke beside the active data center's writable primary. After a switchover or a failover it moves to the new primary. A segment pushed again after a failover with a different content is rejected and retried, so the old and the new primary cannot overwrite each other's WAL in the archive. wal-g does not offer this protection on a `local` storage backend.
- **Secrets.** The spokes have no KubeStash objects, so KubeDB copies the storage credentials, the encryption Secret and the wal-g configuration Secret to every spoke, into the database namespace, as `<db>-backup-storage`, `<db>-backup-encryption` and `<db>-backup-walg-config`.
- **Restore.** A restore runs in the data center that is active when it starts. Only that data center's first pod (ordinal 0) is restored; every other pod, in both data centers, is then created and re-seeded from it through the normal replication path.

## Before You Begin

- A DC-DR installation as in this guide, with KubeStash installed on the hub (see [Install](/docs/guides/postgres/distributed/install/index.md#optional-kubestash)). The Sidekick operator is part of the KubeDB chart.
- The KubeStash `postgres-addon` must have the `distributed-physical-backup-restore` restore task, which ships with the KubeDB catalog of this release:

  ```bash
  $ kubectl --kubeconfig $HUB get addon postgres-addon -o jsonpath='{.spec.restoreTasks[*].name}'
  ```

  The output must contain `distributed-physical-backup-restore`.
- A backup storage that every spoke can reach, for example an S3-compatible bucket. The Jobs and the archiver run on the spokes, not on the hub.
- You should be familiar with [continuous archiving and PITR](/docs/guides/postgres/pitr/archiver.md) for a single-cluster Postgres.

```bash
export HUB=~/dcdr/hub.kubeconfig S1=~/dcdr/spoke-1.kubeconfig S2=~/dcdr/spoke-2.kubeconfig
```

## Set up archiving

Create the `BackupStorage`, its credential Secret, the `RetentionPolicy` and the encryption Secret on the hub, in the database namespace, exactly as in [continuous archiving](/docs/guides/postgres/pitr/archiver.md#continuous-archiving). Then create the `PostgresArchiver` with the **Restic** driver for the full backup. The volumes of a DC-DR database live on the spokes, so the `VolumeSnapshotter` driver, which snapshots PVCs in the hub cluster, cannot be used:

```yaml
apiVersion: archiver.kubedb.com/v1alpha1
kind: PostgresArchiver
metadata:
  name: pg-dcdr-archiver
  namespace: demo
spec:
  pause: false
  databases:
    namespaces:
      from: Selector
      selector:
        matchLabels:
          kubernetes.io/metadata.name: demo
    selector:
      matchLabels:
        archiver: "true"
  retentionPolicy:
    name: postgres-retention-policy
    namespace: demo
  encryptionSecret:
    name: encrypt-secret
    namespace: demo
  fullBackup:
    driver: Restic
    scheduler:
      successfulJobsHistoryLimit: 1
      failedJobsHistoryLimit: 1
      schedule: "0 3 * * *"
    sessionHistoryLimit: 2
  manifestBackup:
    scheduler:
      successfulJobsHistoryLimit: 1
      failedJobsHistoryLimit: 1
      schedule: "0 3 * * *"
    sessionHistoryLimit: 2
  backupStorage:
    ref:
      name: s3-storage
      namespace: demo
```

```bash
$ kubectl --kubeconfig $HUB apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/distributed/backup/yamls/postgresarchiver.yaml
postgresarchiver.archiver.kubedb.com/pg-dcdr-archiver created
```

**Set the archiver when you create the database.** Add the label and the reference to the Postgres of [Provision](/docs/guides/postgres/distributed/provision/index.md) before you apply it:

```yaml
metadata:
  labels:
    archiver: "true"
spec:
  archiver:
    ref:
      name: pg-dcdr-archiver
      namespace: demo
```

Postgres reads its `archive_command` when it starts. If you add the archiver to a database that is already running, the archiver Sidekick runs, but Postgres archives nothing until its pods are restarted: apply a `Restart` PostgresOpsRequest afterwards.

## Check that WAL is archived

KubeDB reports the archiver's state in the `WALArchiving` condition of the Postgres:

```bash
$ kubectl --kubeconfig $HUB -n demo get pg pg-docs -o jsonpath='{range .status.conditions[?(@.type=="WALArchiving")]}{.status} {.reason}: {.message}{"\n"}{end}'
True Archiving: the WAL archiver is running and pushing WAL
```

| Status / reason | Meaning |
|---|---|
| `True` / `Archiving` | The archiver runs beside the primary and has pushed WAL. |
| `False` / `WaitingForFirstPush` | The archiver runs, but no push is recorded yet. If the archiver was added to a running database, it stays here until the pods are restarted. |
| `False` / `WALPushFailing` | The last push failed and no push succeeded after it. The message has the error. |
| `False` / `ArchiverNotOnPrimary` | The archiver runs beside another pod than the writable primary, so the primary's WAL is not archived. |
| `False` / `ArchiverNotRunning` | The archiver Sidekick is not running. |

`True` means that pushes work; it does not tell you up to when you can restore. Check the PITR range of the incremental Snapshot, and that its end advances:

```bash
$ kubectl --kubeconfig $HUB -n demo get snapshot pg-docs-incremental-snapshot -o jsonpath='{.status.components.wal.logStats.start} .. {.status.components.wal.logStats.end}{"\n"}'
```

If a WAL file is missing from the storage and the archiver cannot move past it, KubeDB reports the `WALArchiveGap` condition; see [Accept a WAL archive gap](/docs/guides/postgres/distributed/wal-archive-gap/index.md).

## Restore to a point in time

Restore into a **new** DC-DR database, with its own PlacementPolicy and FailoverGroup, from the repositories the archiver created (`<db>-full` and `<db>-manifest`):

```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: pg-restored
  namespace: demo
spec:
  version: "17.2"
  distributed: true
  replicas: 4
  init:
    archiver:
      recoveryTimestamp: "2026-11-24T08:56:35Z"
      encryptionSecret:
        name: encrypt-secret
        namespace: demo
      fullDBRepository:
        name: pg-docs-full
        namespace: demo
      manifestRepository:
        name: pg-docs-manifest
        namespace: demo
  podTemplate:
    spec:
      podPlacementPolicy:
        name: pg-restored
  ...
```

[restore.yaml](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/postgres/distributed/backup/yamls/restore.yaml) has the complete objects:

```bash
$ kubectl --kubeconfig $HUB apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/distributed/backup/yamls/restore.yaml
failovergroup.apps.k8s.appscode.com/pg-restored created
placementpolicy.apps.k8s.appscode.com/pg-restored created
postgres.kubedb.com/pg-restored created
```

The restore takes longer than a provisioning: the first pod is restored and replays WAL up to the target, then every other pod is re-seeded from it. Wait for `Ready`, `Steady` and `protected`, as after a provisioning.

## Limitations

- **Set the archiver at creation.** An archiver added to a running database archives nothing until the pods are restarted (`WALArchiving` shows `WaitingForFirstPush` meanwhile). Apply a `Restart` PostgresOpsRequest after you add it.
- **A target after the archived WAL is not an error.** A `recoveryTimestamp` later than the last archived WAL restores to the latest archived point, without any error. With no archived WAL at all, you get the full backup only. Check the Snapshot's PITR range (above) before you choose the target time, and check the restored data.
- **Raised server limits.** If, between the full backup and the target time, the source database raised `max_connections`, `max_worker_processes`, `max_wal_senders`, `max_prepared_transactions` or `max_locks_per_transaction`, the restored database must have the same or higher values in its `spec.configuration`. Otherwise recovery pauses when it reaches that change, and the restore does not finish.
- **Restoring again under the same name.** After you delete a restored database to restore it again under the same name, delete its old auth Secret (`<db>-auth`) first. The manifest restore created that Secret, and it is not deleted with the database.
- **WAL archive gaps.** A WAL file missing from the storage stops the PITR range from advancing until the gap is closed or accepted; see [Accept a WAL archive gap](/docs/guides/postgres/distributed/wal-archive-gap/index.md).

## Next Steps

- [Accept a WAL archive gap](/docs/guides/postgres/distributed/wal-archive-gap/index.md).
- [Teardown](/docs/guides/postgres/distributed/teardown/index.md).
