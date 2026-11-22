---
title: Incremental Backup Oracle | KubeStash
description: Continuously back up an Oracle database with LEVEL 0, LEVEL 1 and archived log backups using an OracleArchiver
menu:
  docs_{{ .version }}:
    identifier: guides-oracle-backup-incremental
    name: Incremental Backup
    parent: guides-oracle-backup
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Incremental Backup of Oracle using OracleArchiver

A [base backup](/docs/guides/oracle/backup/base-backup/index.md) can only bring a database back to the moment the backup finished. Everything written between two base backups is lost. Incremental backup closes that gap: KubeDB keeps backing up the changes **between** full backups, so the database can be restored to any point in time.

This guide shows how to turn on incremental backup for an `Oracle` database with an `OracleArchiver`, and how to check that it is working. To restore from these backups, see [Point-in-time Recovery](/docs/guides/oracle/pitr/archiver.md).

## How Incremental Backup Works

An `OracleArchiver` takes three kinds of RMAN backup:

```
time ──────────────────────────────────────────────────────────────────────────────►

LEVEL 0 (full)                                           LEVEL 0 (full)
every fullBackup schedule, e.g. 6h                       next scheduled
│                                                        │
├─ archived logs ─ archived logs ─ archived logs ─ ...   ├─ ...
│  every archivelogInterval, e.g. 1m                     │
│                                                        │
├────────── LEVEL 1 ────────── LEVEL 1 ────────── ...    ├─ ...
│           every backupInterval, e.g. 10m               │
```

| Backup | What it holds | Taken by | How often |
|---|---|---|---|
| LEVEL 0 | A full copy of the database. It is the base the other two build on. | A KubeStash backup job | `spec.fullBackup.scheduler.schedule` |
| Archived redo logs | Every change written to the database since the last cycle | The `<database-name>-sidekick` pod | `spec.incrementalBackup.archivelogInterval` (default `1m`) |
| LEVEL 1 | Only the blocks changed since the LEVEL 0 or the previous LEVEL 1 | The `<database-name>-sidekick` pod | `spec.incrementalBackup.backupInterval` (default `1h`) |

The archived logs are what make every second restorable. The LEVEL 1 backups make a restore faster, because the restore replays less redo.

When you attach an `OracleArchiver` to an `Oracle` database, the KubeDB operator creates:

- a `BackupConfiguration` named `<database-name>-archiver`, which takes the LEVEL 0 backups on schedule,
- a `Repository` named `<database-name>-full`, which holds all the backups of the database,
- a `Snapshot` named `<database-name>-incremental-snapshot`, which records the history of the incremental cycles,
- a `Sidekick` named `<database-name>-sidekick`, which runs the incremental loop.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).
- Install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md).
- Install `KubeStash` in your cluster following the steps [here](https://kubestash.com/docs/latest/setup/install/kubestash).
- If you are not familiar with how KubeStash backs up Oracle databases, please check the [overview](/docs/guides/oracle/backup/overview/index.md) first.

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial are stored in [docs/examples/oracle/pitr](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/oracle/pitr) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

> Oracle images are pulled from `container-registry.oracle.com`. Every Oracle CR must reference an image pull secret (named `orclcred` in this tutorial) through `spec.podTemplate.spec.imagePullSecrets`. If you haven't created one yet, follow the steps [here](/docs/guides/oracle/quickstart#create-oracle-image-pull-secret-important).

## Prepare Backend

We are going to store the backups in an `S3` bucket. Oracle backups support `S3` and S3 compatible storage (for example MinIO) only.

**Create Secret:**

```bash
$ echo -n '<your-aws-access-key-id-here>' > AWS_ACCESS_KEY_ID
$ echo -n '<your-aws-secret-access-key-here>' > AWS_SECRET_ACCESS_KEY
$ kubectl create secret generic -n demo s3-secret \
    --from-file=./AWS_ACCESS_KEY_ID \
    --from-file=./AWS_SECRET_ACCESS_KEY
secret/s3-secret created
```

**Create BackupStorage:**

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: BackupStorage
metadata:
  name: s3-storage
  namespace: demo
spec:
  storage:
    provider: s3
    s3:
      bucket: kubedb-oracle
      prefix: oracle-pitr
      secretName: s3-secret
      region: us-east-2
      endpoint: https://s3.us-east-2.amazonaws.com
  usagePolicy:
    allowedNamespaces:
      from: All
  default: true
  deletionPolicy: Delete
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/oracle/pitr/backupstorage.yaml
backupstorage.storage.kubestash.com/s3-storage created
```

**Create RetentionPolicy:**

The `RetentionPolicy` decides how many full backups to keep. When a full backup is deleted, its LEVEL 1 and archived log backups are deleted with it. So the retention policy also decides how far back in time you can restore.

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: RetentionPolicy
metadata:
  name: demo-retention
  namespace: demo
spec:
  default: true
  failedSnapshots:
    last: 2
  successfulSnapshots:
    last: 3
  usagePolicy:
    allowedNamespaces:
      from: All
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/oracle/pitr/retentionpolicy.yaml
retentionpolicy.storage.kubestash.com/demo-retention created
```

With a LEVEL 0 every 6 hours and the last 3 kept, you can restore to any point in roughly the last 12 to 18 hours.

**Create Encryption Secret:**

```bash
$ echo -n 'changeit' > RESTIC_PASSWORD
$ kubectl create secret generic -n demo encrypt-secret \
    --from-file=./RESTIC_PASSWORD
secret/encrypt-secret created
```

## Create OracleArchiver

`OracleArchiver` is a CR provided by KubeDB to manage continuous archiving (scheduled full backups plus incremental backups) of `Oracle` databases. Below is the YAML of the `OracleArchiver` we are going to create,

```yaml
apiVersion: archiver.kubedb.com/v1alpha1
kind: OracleArchiver
metadata:
  name: oracle-archiver
  namespace: demo
spec:
  pause: false
  databases:
    namespaces:
      from: Same
    selector:
      matchLabels:
        archiver: "true"
  retentionPolicy:
    name: demo-retention
    namespace: demo
  encryptionSecret:
    name: encrypt-secret
    namespace: demo
  fullBackup:
    scheduler:
      schedule: "0 */6 * * *"
      successfulJobsHistoryLimit: 2
      failedJobsHistoryLimit: 2
    sessionHistoryLimit: 3
  incrementalBackup:
    backupInterval: 10m
    archivelogInterval: 1m
  backupStorage:
    ref:
      name: s3-storage
      namespace: demo
  deletionPolicy: Retain
```

Here,

- `spec.databases` selects the `Oracle` databases this archiver backs up. Here, it picks every `Oracle` in the same namespace with the label `archiver: "true"`.
- `spec.fullBackup.scheduler.schedule` is how often a LEVEL 0 full backup is taken. Here, every 6 hours.
- `spec.incrementalBackup.backupInterval` is how often a LEVEL 1 backup is taken. Here, every 10 minutes.
- `spec.incrementalBackup.archivelogInterval` is how often the archived redo logs are shipped. Here, every minute.
- `spec.backupStorage.ref`, `spec.retentionPolicy` and `spec.encryptionSecret` reuse the objects we created above.
- `spec.deletionPolicy: Retain` keeps the `Repository` and its backups if the archiver is deleted.

Let's create the `OracleArchiver`,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/oracle/pitr/oracle-archiver.yaml
oraclearchiver.archiver.kubedb.com/oracle-archiver created
```

## Deploy Sample Oracle Database

Below is the YAML of the `Oracle` database we are going to back up. Notice the `archiver: "true"` label, which the `OracleArchiver` selects,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Oracle
metadata:
  name: sample-oracle
  namespace: demo
  labels:
    archiver: "true"
spec:
  version: "21.3.0"
  edition: enterprise
  mode: Standalone
  replicas: 1
  storageType: Durable
  storage:
    storageClassName: "local-path"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 10Gi
  podTemplate:
    spec:
      imagePullSecrets:
        - name: orclcred
  deletionPolicy: WipeOut
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/oracle/pitr/sample-oracle.yaml
oracle.kubedb.com/sample-oracle created
```

Wait until the database is `Ready`,

```bash
$ kubectl get oracle -n demo sample-oracle
NAME            VERSION   MODE         STATUS   AGE
sample-oracle   21.3.0    Standalone   Ready    9m
```

> You can also add the label to an `Oracle` that is already running: `kubectl label oracle -n demo sample-oracle archiver=true`. Incremental backup does not need a new database.

## Verify Incremental Backup

**Verify the archiver is attached:**

The KubeDB operator finds the `OracleArchiver` that selects the database and records it in `spec.archiver.ref`,

```bash
$ kubectl get oracle -n demo sample-oracle -o jsonpath='{.spec.archiver}'
{"ref":{"name":"oracle-archiver","namespace":"demo"}}
```

**Verify the backup objects:**

```bash
$ kubectl get backupconfiguration,repository,sidekick -n demo
NAME                                                            PHASE   PAUSED   AGE
backupconfiguration.core.kubestash.com/sample-oracle-archiver   Ready            2m

NAME                                                  INTEGRITY   SNAPSHOT-COUNT   SIZE   PHASE   LAST-SUCCESSFUL-BACKUP   AGE
repository.storage.kubestash.com/sample-oracle-full                0                0 B    Ready                            2m

NAME                                                    STATUS    AGE
sidekick.apps.k8s.appscode.com/sample-oracle-sidekick   Current   2m
```

**Verify the full (LEVEL 0) backup:**

KubeStash takes the first LEVEL 0 as soon as the `BackupConfiguration` is ready. After that, it follows the schedule.

```bash
$ kubectl get backupsession -n demo
NAME                                            INVOKER-TYPE          INVOKER-NAME             PHASE       DURATION   AGE
sample-oracle-archiver-full-backup-1795151405   BackupConfiguration   sample-oracle-archiver   Succeeded   2m31s      3m

$ kubectl get snapshots -n demo -l=kubestash.com/repo-name=sample-oracle-full
NAME                                                               REPOSITORY           SESSION       SNAPSHOT-TIME          DELETION-POLICY   PHASE       AGE
sample-oracle-full-sample-oracle-archiver-full-backup-1795151405   sample-oracle-full   full-backup   2026-11-20T05:10:05Z   Delete            Succeeded   3m
sample-oracle-incremental-snapshot                                 sample-oracle-full   full-backup                          Delete            Pending     3m
```

Here,

- `sample-oracle-full-sample-oracle-archiver-full-backup-*` is the LEVEL 0 full backup. A restore starts from one of these.
- `sample-oracle-incremental-snapshot` is not a backup. It records the history of the incremental cycles. Its `PHASE` column is not used; check its history as shown below.

> As with a base backup, the first backup of a database creates an Oracle wallet and restarts the database pod once.

**Verify the incremental loop:**

The `sample-oracle-sidekick` pod stays `Running` and runs the incremental cycles,

```bash
$ kubectl get pods -n demo
NAME                                                                READY   STATUS      RESTARTS   AGE
retention-policy-sample-oracle-archiver-full-backup-17951514wx8fk   0/1     Completed   0          10m
sample-oracle-0                                                     1/1     Running     0          8m
sample-oracle-archiver-full-backup-1795151405-n4m4t                 0/1     Completed   0          10m
sample-oracle-sidekick                                              1/1     Running     0          11m
```

You can follow the cycles in its logs. Each archived log cycle tells you how far the database can now be restored,

```bash
$ kubectl logs -n demo sample-oracle-sidekick | grep -E "Initiating|shipped|incremental backup wrote"
I1120 05:12:50.377566       1 incremental_backup.go:153] Initiating incremental backup loop for BackupConfiguration demo/sample-oracle-archiver: LEVEL 1 every 10m0s, archived logs every 1m0s...
I1120 05:14:01.551623       1 archivelog_backup.go:176] shipped 1 archived log(s) under RMAN tag KI-D8480D9C-A20261120T051350; snapshot sample-oracle-full-sample-oracle-archiver-full-backup-1795151405 is restorable up to 2026-11-20T05:13:41Z (SCN 3712408)
I1120 05:15:12.633007       1 archivelog_backup.go:176] shipped 2 archived log(s) under RMAN tag KI-D8480D9C-A20261120T051501; snapshot sample-oracle-full-sample-oracle-archiver-full-backup-1795151405 is restorable up to 2026-11-20T05:14:52Z (SCN 3712986)
I1120 05:23:05.118820       1 incremental.go:219] oracle incremental backup wrote 5 piece(s), 41.3 MiB, under RMAN tag KI-D8480D9C-I20261120T052250 (LEVEL 1, parent SCN 3708101); chain recoverable to SCN 3716240 in incarnation 2
```

**Verify the restorable window:**

The full backup `Snapshot` records the time range you can restore to. `oracle.kubedb.com/until-time` is the earliest point, and `oracle.kubedb.com/log-until-time` is the latest. The latest point moves forward with every archived log cycle.

```bash
$ kubectl get snapshot -n demo sample-oracle-full-sample-oracle-archiver-full-backup-1795151405 -o yaml | grep -E "until-time"
    oracle.kubedb.com/log-until-time: "2026-11-20T05:22:41Z"
    oracle.kubedb.com/until-time: "2026-11-20T05:13:00Z"
```

The `sample-oracle-incremental-snapshot` keeps a summary of all the cycles, across all the retained full backups,

```bash
$ kubectl get snapshot -n demo sample-oracle-incremental-snapshot -o jsonpath='{.status.components.log.logStats}' | jq '{start, end, totalSucceededCount, totalFailedCount}'
{
  "start": "2026-11-20T05:13:00.000000Z",
  "end": "2026-11-20T05:22:41.000000Z",
  "totalSucceededCount": 10,
  "totalFailedCount": null
}
```

Here, `start` and `end` are the oldest and newest points in time you can restore to right now.

## Configuration

All the settings of the incremental loop live under `spec.incrementalBackup` of the `OracleArchiver`. They all have defaults, so you only need to set the ones you want to change.

| Field | Default | Description |
|---|---|---|
| `backupInterval` | `1h` | Time between two LEVEL 1 backups. Shorter means faster restores, at the cost of more backups. |
| `archivelogInterval` | `1m` | Time between two archived log cycles. This is roughly how much data you can lose. `0` turns the cycles off, so only the LEVEL 1 points can be restored. |
| `archiveLagTarget` | `60s` | Sets Oracle's `ARCHIVE_LAG_TARGET`, so the database switches to a new redo log at least this often while it is being written to. Either `0` (leave the database setting alone) or between `60s` and `2h`. |
| `forceLogging` | `true` | Turns on `FORCE LOGGING`, so even `NOLOGGING` operations write redo. Without it, those changes are corrupt after a point-in-time restore. |
| `blockChangeTracking` | `true` | Turns on Oracle Block Change Tracking, so a LEVEL 1 reads only the changed blocks instead of every datafile. Enterprise edition only. |
| `cumulative` | `false` | When `true`, every LEVEL 1 holds all the changes since the LEVEL 0, not just since the previous LEVEL 1. |
| `minControlFileRecordKeepDays` | `0` | Raises `CONTROL_FILE_RECORD_KEEP_TIME` to at least this many days. Keep it longer than your retention window. `0` leaves the database setting alone. |
| `rmanTimeout` | `6h` | The longest one RMAN run of the loop may take. |
| `successfulLogHistoryLimit` | `5` | How many successful cycles the incremental snapshot keeps in its history. |
| `failedLogHistoryLimit` | `5` | How many failed cycles the incremental snapshot keeps in its history. |

> A full backup and an incremental cycle never run at the same time. If a cycle is due while a full backup is running, the cycle is skipped. A full backup that starts while a cycle is running waits for it to finish.

## Pause and Resume

To pause the backups of every database the archiver selects, set `spec.pause` to `true`. KubeDB pauses the `BackupConfiguration` and stops the sidekick.

```bash
$ kubectl patch oraclearchiver -n demo oracle-archiver --type=merge -p '{"spec":{"pause":true}}'
oraclearchiver.archiver.kubedb.com/oracle-archiver patched
```

To resume, set it back to `false`,

```bash
$ kubectl patch oraclearchiver -n demo oracle-archiver --type=merge -p '{"spec":{"pause":false}}'
oraclearchiver.archiver.kubedb.com/oracle-archiver patched
```

To stop incremental backup completely, delete the `OracleArchiver`. KubeDB removes the `BackupConfiguration` and the sidekick. With `deletionPolicy: Retain`, the `Repository` and the existing backups are kept, so you can still restore from them.

## Next Steps

- Restore an Oracle database to a [point in time](/docs/guides/oracle/pitr/archiver.md) from these backups.

## Cleanup

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete oracle -n demo sample-oracle
kubectl delete oraclearchiver -n demo oracle-archiver
kubectl delete repository -n demo sample-oracle-full
kubectl delete retentionpolicy -n demo demo-retention
kubectl delete backupstorage -n demo s3-storage
kubectl delete secret -n demo s3-secret encrypt-secret
kubectl delete ns demo
```
