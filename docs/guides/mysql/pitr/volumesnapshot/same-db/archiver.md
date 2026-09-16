---
title: Restore in the Same Database using VolumeSnapshot
menu:
  docs_{{ .version }}:
    identifier: volumesnapshot-same-db
    name: Restore in the Same Database
    parent: pitr-volumesnapshot-mysql
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# KubeDB MySQL - Restore in the Same Database using VolumeSnapshot

Here, we will demonstrate the whole process: archiving a MySQL database continuously with the
`VolumeSnapshotter` driver, and then rewinding **that same database** to an earlier point in time using
a `MySQLOpsRequest` of type `ArchiverRestore`.

With this driver the base backup is a CSI `VolumeSnapshot` of the data volume rather than a file
upload. For the same operation with a Restic base backup, see
[Restore in the Same Database Using Restic Driver](/docs/guides/mysql/pitr/restic/same-db/archiver.md).

## Same database or a different one?

**[Restore in a different database](/docs/guides/mysql/pitr/volumesnapshot/different-db/archiver.md)**
creates a *new* MySQL object from the archive and leaves the original untouched. Choose this when the
original is still serving traffic, or when you want to inspect the recovered data first.

**Restore in the same database** — this page — rewinds the existing MySQL object in place. The name,
connection string, secrets and service all stay the same, so applications keep pointing at what they
already point at.

It is destructive: the request wipes the data volumes and rebuilds them from the snapshot, so
everything written after `recoveryTimestamp` is gone from the live database.

## Before You Begin

You need a Kubernetes cluster with `kubectl` configured, the `KubeDB` operator
([setup](/docs/setup/README.md)) and the `KubeStash` operator
([setup](https://github.com/kubestash/installer/tree/master/charts/kubestash)).

You also need a CSI driver that supports snapshots, the `snapshot-controller`, and a
**`VolumeSnapshotClass`**. Nothing can take a CSI snapshot without one, so confirm it exists before
relying on this path:

```bash
$ kubectl get volumesnapshotclass
NAME                    DRIVER               DELETIONPOLICY   AGE
longhorn-snapshot-vsc   driver.longhorn.io   Delete           1h
```

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: The yaml files used in this tutorial are stored in [docs/guides/mysql/pitr/volumesnapshot/same-db/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/mysql/pitr/volumesnapshot/same-db/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Continuous Archiving

### VolumeSnapshotClass

```yaml
apiVersion: snapshot.storage.k8s.io/v1
kind: VolumeSnapshotClass
metadata:
  name: longhorn-snapshot-vsc
driver: driver.longhorn.io
deletionPolicy: Delete
parameters:
  type: snap
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/pitr/volumesnapshot/same-db/yamls/voluemsnapshotclass.yaml
volumesnapshotclass.snapshot.storage.k8s.io/longhorn-snapshot-vsc created
```

> Where the snapshots are stored matters later. Some drivers keep a snapshot *inside* the volume it was
> taken from — Longhorn's `type: snap` is one — so deleting that volume also destroys the snapshots
> taken from it. See [retainPV](#retainpv) below.

### BackupStorage

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: BackupStorage
metadata:
  name: storage
  namespace: demo
spec:
  storage:
    provider: s3
    s3:
      endpoint: s3.amazonaws.com
      bucket: mysql-archiver
      region: us-east-1
      prefix: my-demo
      secretName: s3-secret
  usagePolicy:
    allowedNamespaces:
      from: All
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/pitr/volumesnapshot/same-db/yamls/backupstorage.yaml
backupstorage.storage.kubestash.com/storage created
```

The binlogs still go to this storage even though the base backup is a snapshot — only the base backup
changes with this driver.

### Secret for backup storage

```yaml
apiVersion: v1
kind: Secret
type: Opaque
metadata:
  name: s3-secret
  namespace: demo
stringData:
  AWS_ACCESS_KEY_ID: "*************26CX"
  AWS_SECRET_ACCESS_KEY: "************jj3lp"
  AWS_ENDPOINT: s3.amazonaws.com
```

### Retention policy

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: RetentionPolicy
metadata:
  name: mysql-retention-policy
  namespace: demo
spec:
  maxRetentionPeriod: "30d"
  successfulSnapshots:
    last: 10
  failedSnapshots:
    last: 2
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/pitr/volumesnapshot/same-db/yamls/retentionPolicy.yaml
retentionpolicy.storage.kubestash.com/mysql-retention-policy created
```

> Keep the binlog retention at least as long as your base-backup retention. If binlogs expire sooner
> than the base backups do, an older base backup survives with no binlogs to replay onto it, and can no
> longer be used for point-in-time recovery.

### EncryptionSecret

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: encrypt-secret
  namespace: demo
stringData:
  RESTIC_PASSWORD: "changeit"
```

### MySQLArchiver

The driver is `VolumeSnapshotter`, and the class name is passed to the snapshot task:

```yaml
apiVersion: archiver.kubedb.com/v1alpha1
kind: MySQLArchiver
metadata:
  name: mysqlarchiver-sample
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
    name: mysql-retention-policy
    namespace: demo
  encryptionSecret:
    name: "encrypt-secret"
    namespace: "demo"
  fullBackup:
    driver: "VolumeSnapshotter"
    task:
      params:
        volumeSnapshotClassName: "longhorn-snapshot-vsc"
    scheduler:
      successfulJobsHistoryLimit: 1
      failedJobsHistoryLimit: 1
      schedule: "0 0 * * *"
    sessionHistoryLimit: 2
  manifestBackup:
    scheduler:
      successfulJobsHistoryLimit: 1
      failedJobsHistoryLimit: 1
      schedule: "0 0 * * *"
    sessionHistoryLimit: 2
  backupStorage:
    ref:
      name: "storage"
      namespace: "demo"
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/pitr/volumesnapshot/same-db/yamls/mysqlarchiver.yaml
mysqlarchiver.archiver.kubedb.com/mysqlarchiver-sample created
```

## Deploy MySQL

```yaml
apiVersion: kubedb.com/v1
kind: MySQL
metadata:
  name: mysql
  namespace: demo
  labels:
    archiver: "true"
spec:
  version: "9.7.1"
  replicas: 3
  topology:
    mode: GroupReplication
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  archiver:
    ref:
      name: mysqlarchiver-sample
      namespace: demo
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/pitr/volumesnapshot/same-db/yamls/mysql.yaml
mysql.kubedb.com/mysql created
```

Wait for the database to be `Ready` and the first full backup to complete. The binlog sidekick is held
back until a full backup succeeds, because binlogs with no base backup to replay onto cannot be
restored from.

```bash
$ kubectl get volumesnapshot -n demo
NAME                READYTOUSE   SOURCEPVC      RESTORESIZE   SNAPSHOTCLASS           AGE
mysql-1734156013    true         data-mysql-1   1Gi           longhorn-snapshot-vsc   2m

$ kubectl get sidekick -n demo
NAME             STATUS    AGE
mysql-sidekick   Current   2m
```

## Insert data and note the time

```bash
$ kubectl exec -it -n demo mysql-0 -- bash

mysql> create database demo;
mysql> use demo;
mysql> create table demo_table(id int);
mysql> insert into demo_table values (1),(2),(3),(4),(5),(6),(7),(8),(9),(10);

mysql> select now();
+---------------------+
| now()               |
+---------------------+
| 2024-12-02 06:38:42 |
+---------------------+

mysql> drop table demo_table;
mysql> flush logs;
```

`recoveryTimestamp` is RFC 3339 and interpreted as UTC: `2024-12-02T06:38:42Z`. The boundary resolves
to one second, so transactions inside that second are all in or all out.

## Requirements for an ArchiverRestore

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

> `spec.apply: Always` is required. The default, `IfReady`, waits for the database to be `Ready`, which
> never happens once a previous attempt has wiped its volumes.

### Storage size must not be smaller than the snapshot

A PVC created from a `VolumeSnapshot` must request **at least** the snapshot's size. If `spec.storage`
asks for less, provisioning fails and the restore stalls with the database stuck in `Provisioning`:

```
failed to provision volume with StorageClass "longhorn": ... requested volume size 1073741824
is less than the size 2147483648 for the source snapshot mysql-1734156013
```

This is easy to hit after a volume expansion — the volumes and the new snapshots grow, but a
`spec.storage` that was never updated still names the old size. The database's own status only says
that replicas are not ready, so check the PVC events when a restore sits in `Provisioning`:

```bash
$ kubectl describe pvc -n demo data-mysql-0 | tail -5
```

### ReplicationStrategy

`spec.archiver.replicationStrategy` decides how the other members are brought back once member 0 has
been restored, and applies to group replication only: `none` (each replica restores independently),
`sync` (only pod-0 restores, others clone from it), `fscopy` (only pod-0 restores, others filled by
file-system copy; no cross-zone support). We use `sync`.

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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/pitr/volumesnapshot/same-db/yamls/mysql-inplace-restore.yaml
mysqlopsrequest.ops.kubedb.com/mysql-inplace-restore created

$ kubectl get mysqlopsrequest -n demo -w
NAME                    TYPE              STATUS        AGE
mysql-inplace-restore   ArchiverRestore   Progressing   30s
mysql-inplace-restore   ArchiverRestore   Successful    6m
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
   [Archiving is disabled, and stays disabled](#archiving-is-disabled-and-stays-disabled).
2. **Retain PV.** The data `PersistentVolume`s are forced to `Retain`, so the pre-restore data survives
   a failure part-way through.
3. **Wipe volumes.** The PetSet, its pods and the data volumes are removed.
4. **Trigger.** Member 0's data volume is recreated **from the VolumeSnapshot**, and binlogs are
   replayed on top up to `recoveryTimestamp`.
5. **Database ready.** All members rejoin and the database returns to `Ready`.

The database is unavailable between steps 3 and 5.

### What is different from Restic here

Members are built from the snapshot rather than from a restic repository, so their PVCs are created with a
`dataSource` naming that snapshot. How many carry it depends on `replicationStrategy`:

| strategy | PVCs created with a `dataSource` |
|---|---|
| `sync` | member 0 only — the rest get empty volumes and are seeded by Group Replication |
| `none` | every member, because each one is restored independently |

While the restore is running you can see it:

```bash
$ kubectl get pvc -n demo data-mysql-0 -o jsonpath='{.spec.dataSource}'
{"apiGroup":"snapshot.storage.k8s.io","kind":"VolumeSnapshot","name":"mysql-1734156013"}
```

The binlog replay on top is identical to the Restic path.

### The dataSource is removed before the restore finishes

A PVC's spec is immutable, so the reference cannot be patched away — the claim has to be rebuilt. The
operator does this for you, and by the time the restore completes every data PVC is clean:

```bash
$ kubectl get pvc -n demo data-mysql-0 -o jsonpath='{.spec.dataSource}'
                                  # empty
```

It happens after the binlog replay and before the database is started, one member at a time. Nothing is
copied: the volume already holds the restored data, so the claim is deleted and an identical one without
the `dataSource` is bound straight back to the same `PersistentVolume`. The volume is forced to `Retain`
for the duration and its original reclaim policy is put back afterwards.

Why it is worth doing at all:

- **The reference stops being true.** Your retention policy deletes the `VolumeSnapshot` in time, leaving
  the claim permanently naming an object that no longer exists.
- **Some CSI drivers refuse to expand a volume whose claim carries one.** This has been reported on
  `azuredisk`. It is not universal — on Longhorn the same claim expands online with the field present — so
  treat it as driver-specific rather than a rule of Kubernetes.

### Skipping it

Set `kubedb.com/strip-pvc-datasource: "false"` on the database to leave the claims exactly as the restore
made them:

```yaml
apiVersion: kubedb.com/v1
kind: MySQL
metadata:
  name: mysql
  namespace: demo
  annotations:
    kubedb.com/strip-pvc-datasource: "false"
```

The annotation is read on the database being restored. Any other value — and its absence — means the
rebuild runs.

Two reasons you might want it off. Each rebuilt claim restarts its member, so a `none` restore of a
three-member cluster restarts the cluster three times, once per claim; that is cheap on a small database
and less so on a large one. And on a single-replica database the rebuild never runs regardless of this
annotation, because it has no peer to fall back on if the swap fails.

## retainPV

`spec.archiver.retainPV` decides what happens to the data `PersistentVolume`s the restore replaces.

| | during the restore | after a successful restore |
|---|---|---|
| `true` (default) | forced to `Retain` | kept, and named in an `ArchiverRestoreManualCleanupRequired` condition |
| `false` | forced to `Retain` | deleted |

The volumes are forced to `Retain` in **both** cases, because a failed restore has to be undoable either
way. The flag only decides what survives a restore that *succeeded*.

> This matters more with the VolumeSnapshotter driver than with Restic. If your CSI driver stores
> snapshots inside the source volume — Longhorn's `type: snap`, for instance — then deleting the
> pre-restore volumes can destroy the snapshots taken from them, and base backups that depended on those
> volumes stop being restorable. Prefer `retainPV: true`, or a snapshot type that stores snapshots
> independently of the volume.

```bash
$ kubectl get mysqlopsrequest -n demo mysql-inplace-restore \
    -o jsonpath='{.status.conditions[?(@.type=="ArchiverRestoreManualCleanupRequired")].message}'
```

## Archiving is disabled, and stays disabled

**The restore disables archiving for this database, on purpose, and does not turn it back on.**

A restore forks the binlog history. Everything written after `recoveryTimestamp` is still in the
repository, on a branch the restore abandoned. If archiving simply resumed, the new post-restore
timeline would be pushed into a repository that still holds that abandoned branch, and a later restore
could replay transactions you deliberately rolled back.

So the ops request suspends archiving and leaves the decision to re-enable it with you:

```bash
$ kubectl get mysql -n demo mysql -o jsonpath='{.metadata.annotations.kubedb\.com/suspend-archiver}'
true

$ kubectl get sidekick -n demo
No resources found in demo namespace.

$ kubectl get backupconfiguration -n demo mysql-archiver -o jsonpath='{.spec.paused}'
true
```

Two things worth knowing about the scope:

- The suspension applies to **this database only**, not to the `MySQLArchiver`. Other databases selected
  by the same archiver keep backing up normally.
- `spec.archiver.ref` is left in place, so the database still records which archiver it belongs to.

### Re-enabling archiving

**Only do this once you have checked the restored data and are satisfied it is correct.** While
archiving is off there is no new backup coverage, so do not leave it off indefinitely either.

> **The backup below is required, not advisory.** A restore forks the binlog history, and until
> a base backup exists that is newer than the fork, a *later* restore of this database can
> replay the branch this one abandoned — returning rows you deliberately discarded and dropping
> rows that were committed, with every status object reporting success. Completing these steps
> is part of completing the restore.

**Order matters here, and it is the opposite of what seems natural.**

Re-enable archiving **first**, then take the full backup:

```bash
$ kubectl annotate mysql -n demo mysql kubedb.com/suspend-archiver-
mysql.kubedb.com/mysql annotated
```

Then wait for the sidekick to actually **push a binlog** — not merely for its pod to be
Running:

```bash
$ kubectl logs mysql-sidekick -n demo | grep "Archiving binlog"
INFO: 2026/01/02 03:04:05.123456 Archiving binlog.000004
```

and only then trigger a full backup:

```bash
$ kubectl create job -n demo post-restore-backup --from=cronjob/trigger-mysql-archiver-full-backup
```

The push is what matters because the restore discovers timelines by listing objects in the
repository: until a binlog has been uploaded there is nothing under the new timeline's prefix,
so the restore cannot see it, and a timeline it cannot see cannot be the successor that lets
the superseded one be dropped.

On a busy database the push follows the resume within seconds, because the archiver uploads
the already-closed binlogs in the data directory as soon as it starts. On a quiet one there may
be nothing closed to send, and the timeline will not appear until the next rotation — up to one
`logRotateInterval`. That is exactly the case where waiting on the pod rather than the log line
would mislead you.

The reason is the rule the restore uses to discard a superseded timeline: it drops one only
when **both that timeline and the timeline after it** were created before the base backup.
Resuming archiving is what creates that following timeline. So a base backup taken while
archiving is still suspended is *older* than the timeline that comes next, the pair is not
discarded, and the abandoned history is replayed after all. Taken after the resume, both
fall behind it and are dropped.

Treat the gap between the two commands as a window in which restores are not safe, and keep
it short.

> This protects recovery points **after** the new base backup. Recovering to a moment between
> the restore and that backup still selects the older base and still replays the abandoned
> history, until the older base backups age out of retention.

## Verify

```bash
$ kubectl exec -it -n demo mysql-0 -- mysql -uroot -p$MYSQL_ROOT_PASSWORD \
    -e "select count(*) from demo.demo_table;"
+----------+
| count(*) |
+----------+
|       10 |
+----------+
```

Check **every** member, not just pod-0, and confirm the group reformed:

```bash
$ kubectl exec -it -n demo mysql-0 -- mysql -uroot -p$MYSQL_ROOT_PASSWORD \
    -e "select member_host, member_state from performance_schema.replication_group_members;"
```

All three members should report `ONLINE`.

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

- [Restore in a different database using VolumeSnapshot](/docs/guides/mysql/pitr/volumesnapshot/different-db/archiver.md)
- [Restore in the same database using Restic](/docs/guides/mysql/pitr/restic/same-db/archiver.md)
- Monitor your MySQL database with KubeDB using [Prometheus operator](/docs/guides/mysql/monitoring/prometheus-operator/index.md).
- Detail concepts of [MySQL object](/docs/guides/mysql/concepts/database/index.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
