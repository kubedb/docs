---
title: In-place Point-in-time Recovery Using Restic Driver
menu:
  docs_{{ .version }}:
    identifier: restic-in-place
    name: In-place Restore (Restic)
    parent: pitr-mysql
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# KubeDB MySQL - In-place Point-in-time Recovery Using Restic Driver

Here, we will demonstrate how to restore an **existing** MySQL database to an earlier point in
time, without creating a second database, using a `MySQLOpsRequest` of type `ArchiverRestore`.

This page uses the `Restic` driver, where the base backup is taken with
[Percona XtraBackup](https://www.percona.com/mysql/software/percona-xtrabackup) and uploaded to
your backup storage. For the same operation with a CSI VolumeSnapshot base backup, see
[In-place Restore (VolumeSnapshot)](/docs/guides/mysql/pitr/volumesnapshot/in-place-restore.md).

## In-place restore vs. restoring into a new database

[Continuous Archiving and Point-in-time Recovery Using Restic Driver](/docs/guides/mysql/pitr/restic/archiver.md)
restores into a **new** MySQL object, leaving the original untouched. That is the safer choice when
you want to compare the two, or when the original is still serving traffic.

An `ArchiverRestore` ops request instead rewinds the database **in place**: the same MySQL object,
the same name, the same connection string, the same secrets. Applications keep pointing at what they
already point at.

The trade-off is that it is destructive. The request wipes the database's data volumes and rebuilds
them from the repository. Everything written after `recoveryTimestamp` is gone from the live
database — that is the point of the operation, but it is worth saying plainly.

## Before You Begin

- You need a running Kubernetes cluster, and the `KubeDB` and `KubeStash` operators installed.
- You need an archived MySQL database. This page assumes the setup from
  [Continuous Archiving and Point-in-time Recovery Using Restic Driver](/docs/guides/mysql/pitr/restic/archiver.md):
  a `BackupStorage`, a `RetentionPolicy`, an `EncryptionSecret`, a `MySQLArchiver` with
  `spec.fullBackup.driver: Restic`, and a MySQL database labelled so the archiver adopts it.
- At least one **full backup** must have completed, and the binlogs covering your target time must
  have reached the backup storage. A restore replays binlogs on top of a base backup; without a base
  backup there is nothing to replay onto.

To keep the examples copy-pasteable we use the `demo` namespace.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Requirements

The webhook rejects an `ArchiverRestore` request that cannot succeed, rather than letting it fail
half-way. The rules are:

| Field | Rule |
|---|---|
| `spec.archiver` | required |
| `spec.archiver.recoveryTimestamp` | required |
| `spec.archiver.fullDBRepository` | required — a manifest-only restore cannot repopulate a wiped data directory |
| `spec.archiver.replicationStrategy` | one of `none`, `sync`, `fscopy`, `clone` |
| `spec.timeout` | required — the default per-step budget is unrelated to how long a restore takes |
| `spec.apply` | must be `Always` |

and on the database itself:

- `spec.storageType` must be `Durable`. An ephemeral database has no data volumes to restore into.
- The topology must be `Standalone`, `GroupReplication` or `InnoDBCluster`. `SemiSync` is rejected
  because its entrypoint has no `PITR_RESTORE` gate, and a remote replica is rejected because it is
  seeded from its source — restore the source instead.

> `spec.apply: Always` is required rather than merely recommended. The default, `IfReady`, holds the
> request Pending until the database reaches `Ready`. This request wipes the database, so on a rerun
> against one a previous attempt already wiped, `Ready` never arrives and the request waits forever.

## Choose the recovery timestamp

Say a table was dropped by accident and you want the database back as it was just before that:

```bash
$ kubectl exec -it -n demo mysql-0 -- bash

mysql> select now();
+---------------------+
| now()               |
+---------------------+
| 2024-12-02 06:38:42 |
+---------------------+

mysql> drop table demo_table;
```

`recoveryTimestamp` is in RFC 3339 and interpreted as UTC, so `2024-12-02 06:38:42` becomes
`2024-12-02T06:38:42Z`.

> The boundary resolves to one second. Transactions committed inside the same second as
> `recoveryTimestamp` are either all included or all excluded — you cannot split them.

## ReplicationStrategy

`spec.archiver.replicationStrategy` decides how the other members are brought back once member 0 has
been restored. It applies to group replication only.

***none*** — every replica restores the base backup and binlogs independently, then joins the group.

***sync*** — only pod-0 restores. The other replicas then clone from pod-0 using the MySQL clone
plugin.

***fscopy*** — only pod-0 restores, and the other replicas' data directories are filled by copying
pod-0's. Does not support cross-zone operation.

We use `sync` below.

## Restore in place

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MySQLOpsRequest
metadata:
  name: mysql-inplace-restore
  namespace: demo
spec:
  type: ArchiverRestore
  databaseRef:
    name: mysql
  apply: Always
  timeout: 30m
  archiver:
    recoveryTimestamp: "2024-12-02T06:38:42Z"
    encryptionSecret:
      name: encrypt-secret
      namespace: demo
    fullDBRepository:
      name: mysql-full
      namespace: demo
    replicationStrategy: sync
    retainPV: true
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/pitr/restic/yamls/mysql-inplace-restore.yaml
mysqlopsrequest.ops.kubedb.com/mysql-inplace-restore created
```

Watch it progress:

```bash
$ kubectl get mysqlopsrequest -n demo -w
NAME                    TYPE              STATUS       AGE
mysql-inplace-restore   ArchiverRestore   Progressing  30s
mysql-inplace-restore   ArchiverRestore   Successful   6m
```

## What the request does

Each step records a condition, so `kubectl describe` tells you where a restore got to:

```bash
$ kubectl get mysqlopsrequest -n demo mysql-inplace-restore -o jsonpath='{range .status.conditions[*]}{.type}{"\n"}{end}'
ArchiverRestoreSuspendArchiver
ArchiverRestoreRetainPV
ArchiverRestoreWipeVolumes
ArchiverRestoreTriggered
ArchiverRestoreDataRestored
DatabaseReady
Successful
```

1. **Suspend archiving.** Before anything is torn down, archiving is stopped for this database: the
   archiver sidekick is deleted and the `BackupConfiguration` paused. See
   [Archiving stays suspended](#archiving-stays-suspended) below.
2. **Retain PV.** The data `PersistentVolume`s are forced to the `Retain` reclaim policy, so the
   pre-restore data survives even if the restore fails part-way.
3. **Wipe volumes.** The PetSet, its pods and the data volumes are removed.
4. **Trigger.** The database is prepared for recovery, and the provisioner rebuilds it from the base
   backup and replays binlogs up to `recoveryTimestamp`.
5. **Database ready.** All members rejoin and the database returns to `Ready`.

## retainPV

`spec.archiver.retainPV` decides what happens to the data `PersistentVolume`s the restore replaces.

| | during the restore | after a successful restore |
|---|---|---|
| `true` (default) | forced to `Retain` | kept, and named in an `ArchiverRestoreManualCleanupRequired` condition |
| `false` | forced to `Retain` | deleted |

The volumes are forced to `Retain` in **both** cases, because a failed restore has to be undoable
either way. The flag only decides what survives a restore that succeeded.

With `retainPV: true`, the released volumes are listed for you:

```bash
$ kubectl get mysqlopsrequest -n demo mysql-inplace-restore \
    -o jsonpath='{.status.conditions[?(@.type=="ArchiverRestoreManualCleanupRequired")].message}'
```

They are yours to delete once you are satisfied with the restore. KubeDB deliberately does not remove
them for you — they hold the only copy of the pre-restore data.

## Archiving stays suspended

A restore forks the binlog history: everything the database wrote after `recoveryTimestamp` is left
behind on an abandoned branch in the repository. If archiving resumed on its own, the post-restore
timeline would be pushed into a repository that still holds that branch.

So the request suspends archiving **for this database** and never resumes it:

```bash
$ kubectl get mysql -n demo mysql -o jsonpath='{.metadata.annotations.kubedb\.com/suspend-archiver}'
true

$ kubectl get sidekick -n demo
No resources found in demo namespace.

$ kubectl get backupconfiguration -n demo mysql-archiver -o jsonpath='{.spec.paused}'
true
```

The suspension is scoped to the database, not to the archiver, so other databases sharing the same
`MySQLArchiver` keep backing up normally. `spec.archiver.ref` is left in place, so the database still
records which archiver it belongs to.

**Before resuming, take a fresh full backup.** A base backup taken *after* the restore means later
restores select it and never look at the abandoned branch:

```bash
$ kubectl apply -f - <<EOF
apiVersion: core.kubestash.com/v1alpha1
kind: BackupSession
metadata:
  name: post-restore-full-1
  namespace: demo
spec:
  invoker:
    apiGroup: core.kubestash.com
    kind: BackupConfiguration
    name: mysql-archiver
  session: full-backup
EOF
```

> The BackupSession name must end in `-<digits>`.

Then resume archiving by removing the annotation:

```bash
$ kubectl annotate mysql -n demo mysql kubedb.com/suspend-archiver-
mysql.kubedb.com/mysql annotated
```

The sidekick returns within seconds and the `BackupConfiguration` un-pauses.

## Verify

```bash
$ kubectl exec -it -n demo mysql-0 -- mysql -uroot -p$MYSQL_ROOT_PASSWORD

mysql> select count(*) from demo_table;
```

Check every member, not just pod-0 — a restore that rebuilt pod-0 correctly but re-seeded the others
badly is exactly what per-member verification catches.

```bash
$ kubectl exec -it -n demo mysql-0 -- mysql -uroot -p$MYSQL_ROOT_PASSWORD \
    -e "select member_host, member_state from performance_schema.replication_group_members;"
```

## Cleaning up

```bash
$ kubectl delete mysqlopsrequest -n demo mysql-inplace-restore
$ kubectl delete mysql -n demo mysql
$ kubectl delete mysqlarchiver -n demo mysqlarchiver-sample
$ kubectl delete backupstorage -n demo storage
$ kubectl delete ns demo
```

If you restored with `retainPV: true`, the released `PersistentVolume`s are still there and still on
`Retain`. Delete them by hand when you no longer need the pre-restore data.

## Next Steps

- [Continuous Archiving and Point-in-time Recovery Using Restic Driver](/docs/guides/mysql/pitr/restic/archiver.md)
- [In-place Restore using VolumeSnapshot](/docs/guides/mysql/pitr/volumesnapshot/in-place-restore.md)
- Monitor your MySQL database with KubeDB using [Prometheus operator](/docs/guides/mysql/monitoring/prometheus-operator/index.md).
- Detail concepts of [MySQL object](/docs/guides/mysql/concepts/database/index.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
