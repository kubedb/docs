---
title: Restore in the Same Database Using Restic Driver
menu:
  docs_{{ .version }}:
    identifier: restic-same-db
    name: Restore in the Same Database
    parent: pitr-restic-mysql
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# KubeDB MySQL - Restore in the Same Database Using Restic Driver

Here, we will demonstrate the whole process: archiving a MySQL database continuously with the
`Restic` driver, and then rewinding **that same database** to an earlier point in time using a
`MySQLOpsRequest` of type `ArchiverRestore`.

This uses [Percona XtraBackup](https://www.percona.com/mysql/software/percona-xtrabackup) for the base
backup. For the same operation with a CSI VolumeSnapshot base backup, see
[Restore in the Same Database using VolumeSnapshot](/docs/guides/mysql/pitr/volumesnapshot/same-db/archiver.md).

## Same database or a different one?

There are two ways to recover to a point in time, and they suit different situations.

**[Restore in a different database](/docs/guides/mysql/pitr/restic/different-db/archiver.md)** creates a
*new* MySQL object from the archive and leaves the original untouched. Choose this when the original is
still serving traffic, or when you want to inspect the recovered data before committing to it.

**Restore in the same database** — this page — rewinds the existing MySQL object in place. The name,
the connection string, the secrets and the service all stay the same, so applications keep pointing at
what they already point at, and nothing has to be re-wired.

The trade-off is that it is destructive. The request wipes the database's data volumes and rebuilds
them from the archive, so everything written after `recoveryTimestamp` is gone from the live database.
That is the intent of the operation, but it is worth being explicit about.

## Before You Begin

You need a Kubernetes cluster with `kubectl` configured. If you don't have one, you can create one
with [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Install the `KubeDB` operator following the steps [here](/docs/setup/README.md), and the `KubeStash`
operator following the steps [here](https://github.com/kubestash/installer/tree/master/charts/kubestash).

This tutorial uses a separate namespace called `demo`.

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: The yaml files used in this tutorial are stored in [docs/guides/mysql/pitr/restic/same-db/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/mysql/pitr/restic/same-db/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Continuous Archiving

Before anything can be restored, the database has to be archived. That needs a `BackupStorage`, a
`RetentionPolicy`, an `EncryptionSecret` and a `MySQLArchiver`.

### BackupStorage

`BackupStorage` is a CR provided by KubeStash that can manage storage from various providers like GCS,
S3 and more. Here we use an AWS S3 bucket.

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
      bucket: mysql-xtrabackup
      region: us-east-1
      prefix: my-demo
      secretName: s3-secret
  usagePolicy:
    allowedNamespaces:
      from: All
  deletionPolicy: WipeOut
```

Note: verify the bucket already exists on your provider before applying this.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/pitr/restic/same-db/yamls/backupstorage.yaml
backupstorage.storage.kubestash.com/storage created
```

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

```bash
$ kubectl apply -f storage-secret.yaml
secret/s3-secret created
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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/pitr/restic/same-db/yamls/retention-policy.yaml
retentionpolicy.storage.kubestash.com/mysql-retention-policy created
```

> Keep the binlog retention at least as long as your base-backup retention. If binlogs expire sooner
> than the full backups do, an older base backup survives with no binlogs to replay onto it, and it can
> no longer be used for point-in-time recovery.

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

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/pitr/restic/same-db/yamls/encryptionSecret.yaml
secret/encrypt-secret created
```

### MySQLArchiver

The archiver selects the databases it backs up by label, takes the base backups, and runs the sidekick
that pushes binlogs continuously.

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
    driver: "Restic"
    jobTemplate:
      spec:
        securityContext:
          runAsUser: 999
          runAsGroup: 0
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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/pitr/restic/same-db/yamls/mysqlarchiver.yaml
mysqlarchiver.archiver.kubedb.com/mysqlarchiver-sample created
```

## Deploy MySQL

The `archiver: "true"` label is what makes the archiver adopt this database.

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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/pitr/restic/same-db/yamls/mysql.yaml
mysql.kubedb.com/mysql created
```

Wait for the database to be `Ready`, and for the first full backup to complete. The sidekick that
pushes binlogs is deliberately held back until a full backup succeeds — binlogs with no base backup to
replay onto cannot be restored from.

```bash
$ kubectl get mysql -n demo -w
NAME    VERSION   STATUS   AGE
mysql   9.7.1     Ready    5m

$ kubectl get backupsession -n demo
NAME                                    INVOKER-TYPE          INVOKER-NAME     PHASE       AGE
mysql-archiver-full-backup-1734156000   BackupConfiguration   mysql-archiver   Succeeded   2m

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
```

Now suppose the table is dropped by accident:

```bash
mysql> drop table demo_table;
mysql> flush logs;
```

`06:38:42` is the moment we want back. `recoveryTimestamp` is RFC 3339 and interpreted as UTC, so it
becomes `2024-12-02T06:38:42Z`.

> The boundary resolves to one second. Transactions committed within the same second as
> `recoveryTimestamp` are either all included or all excluded — they cannot be split.

Make sure the binlogs covering that moment have actually reached the backup storage before restoring.

## Requirements for an ArchiverRestore

The webhook rejects a request that cannot succeed, rather than letting it fail half-way:

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
> request Pending until the database is `Ready`. This request wipes the database, so on a rerun against
> one a previous attempt already wiped, `Ready` never arrives and the request waits forever.

### ReplicationStrategy

`spec.archiver.replicationStrategy` decides how the other members are brought back once member 0 has
been restored. It applies to group replication only.

***none*** — every replica restores the base backup and binlogs independently, then joins the group.

***sync*** — only pod-0 restores. The other replicas then clone from pod-0 using the MySQL clone plugin.

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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/pitr/restic/same-db/yamls/mysql-inplace-restore.yaml
mysqlopsrequest.ops.kubedb.com/mysql-inplace-restore created

$ kubectl get mysqlopsrequest -n demo -w
NAME                    TYPE              STATUS        AGE
mysql-inplace-restore   ArchiverRestore   Progressing   30s
mysql-inplace-restore   ArchiverRestore   Successful    6m
```

## What the request does

Each step records a condition, so `kubectl describe` tells you exactly where a restore got to:

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

1. **Suspend archiving.** Archiving is stopped for this database before anything is torn down — the
   sidekick is deleted and the `BackupConfiguration` paused. See
   [Archiving is disabled, and stays disabled](#archiving-is-disabled-and-stays-disabled).
2. **Retain PV.** The data `PersistentVolume`s are forced to the `Retain` reclaim policy, so the
   pre-restore data survives even if the restore fails part-way.
3. **Wipe volumes.** The PetSet, its pods and the data volumes are removed.
4. **Trigger.** The database is prepared for recovery, and the provisioner rebuilds it from the base
   backup and replays binlogs up to `recoveryTimestamp`.
5. **Database ready.** All members rejoin and the database returns to `Ready`.

The database is unavailable between steps 3 and 5.

## retainPV

`spec.archiver.retainPV` decides what happens to the data `PersistentVolume`s the restore replaces.

| | during the restore | after a successful restore |
|---|---|---|
| `true` (default) | forced to `Retain` | kept, and named in an `ArchiverRestoreManualCleanupRequired` condition |
| `false` | forced to `Retain` | deleted |

The volumes are forced to `Retain` in **both** cases, because a failed restore has to be undoable
either way. The flag only decides what survives a restore that *succeeded*.

With `retainPV: true`, the released volumes are listed for you:

```bash
$ kubectl get mysqlopsrequest -n demo mysql-inplace-restore \
    -o jsonpath='{.status.conditions[?(@.type=="ArchiverRestoreManualCleanupRequired")].message}'
```

They are yours to delete once you are satisfied with the restore. KubeDB deliberately does not remove
them for you — they hold the only copy of the pre-restore data.

## Archiving is disabled, and stays disabled

**The restore disables archiving for this database, on purpose, and does not turn it back on.**

A restore forks the binlog history. Everything the database wrote after `recoveryTimestamp` is still in
the repository, on a branch the restore abandoned. If archiving simply resumed, the new post-restore
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

- The suspension applies to **this database only**, not to the `MySQLArchiver`. Other databases
  selected by the same archiver keep backing up normally.
- `spec.archiver.ref` is left in place, so the database still records which archiver it belongs to.

### Re-enabling archiving

**Only do this once you have checked the restored data and are satisfied it is correct.** While
archiving is off there is no new backup coverage, so do not leave it off indefinitely either.

**First, take a fresh full backup.** This is the step that matters: a base backup taken *after* the
restore means every later restore selects it and never looks at the abandoned branch.

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

Wait for it to succeed, then remove the annotation:

```bash
$ kubectl annotate mysql -n demo mysql kubedb.com/suspend-archiver-
mysql.kubedb.com/mysql annotated
```

The sidekick returns within seconds and the `BackupConfiguration` un-pauses:

```bash
$ kubectl get sidekick -n demo
NAME             STATUS    AGE
mysql-sidekick   Current   5s
```

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

The dropped table is back with its ten rows. Check **every** member, not just pod-0 — a restore that
rebuilt pod-0 correctly but re-seeded the others badly is exactly what per-member verification catches:

```bash
$ kubectl exec -it -n demo mysql-1 -- mysql -uroot -p$MYSQL_ROOT_PASSWORD \
    -e "select count(*) from demo.demo_table;"

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

If you restored with `retainPV: true`, the released `PersistentVolume`s are still there and still on
`Retain`. Delete them by hand when you no longer need the pre-restore data.

## Next Steps

- [Restore in a different database using Restic](/docs/guides/mysql/pitr/restic/different-db/archiver.md)
- [Restore in the same database using VolumeSnapshot](/docs/guides/mysql/pitr/volumesnapshot/same-db/archiver.md)
- Monitor your MySQL database with KubeDB using [Prometheus operator](/docs/guides/mysql/monitoring/prometheus-operator/index.md).
- Detail concepts of [MySQL object](/docs/guides/mysql/concepts/database/index.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
