---
title: In-place Point-in-time Recovery using VolumeSnapshot
menu:
  docs_{{ .version }}:
    identifier: volumesnapshot-in-place
    name: In-place Restore (VolumeSnapshot)
    parent: pitr-mysql
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# KubeDB MySQL - In-place Point-in-time Recovery using VolumeSnapshot

Here, we will demonstrate how to restore an **existing** MySQL database to an earlier point in time,
without creating a second database, using a `MySQLOpsRequest` of type `ArchiverRestore`.

This page uses the `VolumeSnapshotter` driver, where the base backup is a CSI `VolumeSnapshot` of the
data volume rather than a file upload. For the same operation with a Restic base backup, see
[In-place Restore (Restic)](/docs/guides/mysql/pitr/restic/in-place-restore.md).

## In-place restore vs. restoring into a new database

[Continuous Archiving and Point-in-time Recovery using VolumeSnapshot](/docs/guides/mysql/pitr/volumesnapshot/archiver.md)
restores into a **new** MySQL object, leaving the original untouched.

An `ArchiverRestore` ops request instead rewinds the database **in place**: the same MySQL object,
name, connection string and secrets. Applications keep pointing at what they already point at.

It is destructive. The request wipes the data volumes and rebuilds them from the snapshot, so
everything written after `recoveryTimestamp` is gone from the live database.

## Before You Begin

- A running Kubernetes cluster with the `KubeDB` and `KubeStash` operators installed.
- A CSI driver that supports snapshots, the `snapshot-controller`, and a **`VolumeSnapshotClass`**.
  Nothing can take a CSI snapshot without one, so check it exists before you rely on this path:

  ```bash
  $ kubectl get volumesnapshotclass
  NAME                      DRIVER               DELETIONPOLICY   AGE
  longhorn-snapshot-vsc     driver.longhorn.io   Delete           1h
  ```

- An archived MySQL database set up as in
  [Continuous Archiving and Point-in-time Recovery using VolumeSnapshot](/docs/guides/mysql/pitr/volumesnapshot/archiver.md),
  with the archiver configured for the VolumeSnapshotter driver:

  ```yaml
  spec:
    fullBackup:
      driver: "VolumeSnapshotter"
      task:
        params:
          volumeSnapshotClassName: "longhorn-snapshot-vsc"
  ```

- At least one **full backup** must have completed, and the binlogs covering your target time must
  have reached the backup storage.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Requirements

The webhook rejects an `ArchiverRestore` request that cannot succeed:

| Field | Rule |
|---|---|
| `spec.archiver` | required |
| `spec.archiver.recoveryTimestamp` | required |
| `spec.archiver.fullDBRepository` | required — a manifest-only restore cannot repopulate a wiped data directory |
| `spec.archiver.replicationStrategy` | one of `none`, `sync`, `fscopy`, `clone` |
| `spec.timeout` | required |
| `spec.apply` | must be `Always` |

and on the database: `spec.storageType` must be `Durable`, and the topology must be `Standalone`,
`GroupReplication` or `InnoDBCluster`.

### Storage size must not be smaller than the snapshot

A PVC created from a `VolumeSnapshot` must request **at least** the snapshot's size. If the database's
`spec.storage` asks for less, provisioning fails and the restore stalls with the database stuck in
`Provisioning`:

```
failed to provision volume with StorageClass "longhorn": ... requested volume size 1073741824
is less than the size 2147483648 for the source snapshot mysql-1786620013
```

This is easy to hit after a volume expansion — the volumes and the new snapshot grow, but a
`spec.storage` that was never updated still names the old size. If a restore sits in `Provisioning`
with nothing obvious on the database, check the PVC's events:

```bash
$ kubectl describe pvc -n demo data-mysql-0 | tail -5
```

## Choose the recovery timestamp

```bash
$ kubectl exec -it -n demo mysql-0 -- bash

mysql> select now();
+---------------------+
| now()               |
+---------------------+
| 2024-12-02 06:38:42 |
+---------------------+
```

`recoveryTimestamp` is RFC 3339 and interpreted as UTC: `2024-12-02T06:38:42Z`. The boundary resolves
to one second.

## ReplicationStrategy

`spec.archiver.replicationStrategy` decides how the other members are brought back once member 0 has
been restored, and applies to group replication only: `none` (each replica restores independently),
`sync` (only pod-0 restores, others clone from it), `fscopy` (only pod-0 restores, others are filled
by file-system copy; no cross-zone support). We use `sync`.

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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/pitr/volumesnapshot/yamls/mysql-inplace-restore.yaml
mysqlopsrequest.ops.kubedb.com/mysql-inplace-restore created

$ kubectl get mysqlopsrequest -n demo -w
NAME                    TYPE              STATUS       AGE
mysql-inplace-restore   ArchiverRestore   Progressing  30s
mysql-inplace-restore   ArchiverRestore   Successful   6m
```

## What the request does

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

1. **Suspend archiving.** Archiving is stopped for this database before anything is torn down — see
   [Archiving stays suspended](#archiving-stays-suspended).
2. **Retain PV.** The data `PersistentVolume`s are forced to `Retain`, so the pre-restore data
   survives a failure part-way through.
3. **Wipe volumes.** The PetSet, its pods and the data volumes are removed.
4. **Trigger.** The data volume for member 0 is recreated **from the VolumeSnapshot**, and binlogs are
   replayed on top up to `recoveryTimestamp`.
5. **Database ready.** All members rejoin and the database returns to `Ready`.

### What is different from Restic here

Only member 0 is built from the snapshot. Its PVC is created with a `dataSource` pointing at the
`VolumeSnapshot`:

```bash
$ kubectl get pvc -n demo data-mysql-0 -o jsonpath='{.spec.dataSource}'
{"apiGroup":"snapshot.storage.k8s.io","kind":"VolumeSnapshot","name":"mysql-1786620013"}
```

The remaining members get empty volumes and are seeded by Group Replication, which is why they carry
no `dataSource`. The binlog replay on top is identical to the Restic path.

> A PVC's spec is immutable after creation, so this `dataSource` stays on member 0's PVC for the life
> of the claim. It is inert — it is only read when the volume is provisioned.

## retainPV

`spec.archiver.retainPV` decides what happens to the data `PersistentVolume`s the restore replaces.

| | during the restore | after a successful restore |
|---|---|---|
| `true` (default) | forced to `Retain` | kept, and named in an `ArchiverRestoreManualCleanupRequired` condition |
| `false` | forced to `Retain` | deleted |

The volumes are forced to `Retain` in **both** cases, because a failed restore has to be undoable
either way.

> With a CSI driver whose snapshots live inside the source volume rather than in independent storage
> — Longhorn's `type: snap`, for example — deleting the pre-restore volumes can also destroy the
> snapshots taken from them. If you rely on older base backups, prefer `retainPV: true`, or use a
> snapshot type that stores snapshots independently of the volume.

```bash
$ kubectl get mysqlopsrequest -n demo mysql-inplace-restore \
    -o jsonpath='{.status.conditions[?(@.type=="ArchiverRestoreManualCleanupRequired")].message}'
```

## Archiving stays suspended

A restore forks the binlog history: everything written after `recoveryTimestamp` is left behind on an
abandoned branch in the repository. If archiving resumed on its own, the post-restore timeline would
be pushed into a repository that still holds that branch.

So the request suspends archiving **for this database** and never resumes it:

```bash
$ kubectl get mysql -n demo mysql -o jsonpath='{.metadata.annotations.kubedb\.com/suspend-archiver}'
true

$ kubectl get sidekick -n demo
No resources found in demo namespace.

$ kubectl get backupconfiguration -n demo mysql-archiver -o jsonpath='{.spec.paused}'
true
```

The suspension is scoped to the database rather than the archiver, so other databases sharing the same
`MySQLArchiver` keep backing up normally.

**Before resuming, take a fresh full backup**, so that later restores select a base backup taken after
this restore and never look at the abandoned branch:

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

Then resume archiving:

```bash
$ kubectl annotate mysql -n demo mysql kubedb.com/suspend-archiver-
mysql.kubedb.com/mysql annotated
```

## Verify

```bash
$ kubectl exec -it -n demo mysql-0 -- mysql -uroot -p$MYSQL_ROOT_PASSWORD \
    -e "select count(*) from demo.demo_table;"
```

Check every member, and confirm the group reformed:

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

If you restored with `retainPV: true`, the released `PersistentVolume`s are still there on `Retain`.
Delete them by hand once you no longer need the pre-restore data.

## Next Steps

- [Continuous Archiving and Point-in-time Recovery using VolumeSnapshot](/docs/guides/mysql/pitr/volumesnapshot/archiver.md)
- [In-place Restore using Restic](/docs/guides/mysql/pitr/restic/in-place-restore.md)
- Monitor your MySQL database with KubeDB using [Prometheus operator](/docs/guides/mysql/monitoring/prometheus-operator/index.md).
- Detail concepts of [MySQL object](/docs/guides/mysql/concepts/database/index.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
