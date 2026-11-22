---
title: Continuous Archiving and Point-in-time Recovery
menu:
  docs_{{ .version }}:
    identifier: guides-oracle-pitr-archiver
    name: Overview
    parent: guides-oracle-pitr
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# KubeDB Oracle - Continuous Archiving and Point-in-time Recovery

Here, we will show you how to use KubeDB to continuously archive an `Oracle` database and restore it to a specific point in time.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now,
- install `KubeDB` operator in your cluster following the steps [here](/docs/setup/README.md).
- install `KubeStash` operator in your cluster following the steps [here](https://kubestash.com/docs/latest/setup/install/kubestash).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: The yaml files used in this tutorial are stored in [docs/examples/oracle/pitr](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/oracle/pitr) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

> Oracle images are pulled from `container-registry.oracle.com`. Every Oracle CR must reference an image pull secret (named `orclcred` in this tutorial) through `spec.podTemplate.spec.imagePullSecrets`. If you haven't created one yet, follow the steps [here](/docs/guides/oracle/quickstart#create-oracle-image-pull-secret-important).

## Continuous Archiving

Continuous archiving means taking a full backup of the `Oracle` database on a schedule and, in between full backups, continuously backing up the changes. For Oracle, KubeDB does this with RMAN:

- a **LEVEL 0** full backup on a schedule,
- the **archived redo logs** every minute,
- a **LEVEL 1** incremental backup on a fixed interval.

The [Incremental Backup](/docs/guides/oracle/backup/incremental/index.md) guide explains these in more detail. To set up continuous archiving, we need a `BackupStorage`, a `RetentionPolicy`, and an `OracleArchiver`.

### Prepare Backend

We are going to store the backed up data in an `S3` bucket. We have to create a `Secret` with the necessary credentials and a `BackupStorage` CR to use this backend.

> **Note:** Oracle supports `S3` and S3 compatible storage (for example MinIO) as backup storage. For MinIO, set `endpoint` to the MinIO service URL, including the scheme and port, for example `http://minio.demo.svc.cluster.local:80`.

**Create Secret:**

Let's create a secret called `s3-secret` with access credentials to our desired S3 bucket,

```bash
$ echo -n '<your-aws-access-key-id-here>' > AWS_ACCESS_KEY_ID
$ echo -n '<your-aws-secret-access-key-here>' > AWS_SECRET_ACCESS_KEY
$ kubectl create secret generic -n demo s3-secret \
    --from-file=./AWS_ACCESS_KEY_ID \
    --from-file=./AWS_SECRET_ACCESS_KEY
secret/s3-secret created
```

**Create BackupStorage:**

Now, create a `BackupStorage` that references this secret. Below is the YAML of the `BackupStorage` CR we are going to create,

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

Let's create the BackupStorage we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/oracle/pitr/backupstorage.yaml
backupstorage.storage.kubestash.com/s3-storage created
```

**Create RetentionPolicy:**

Now, let's create a `RetentionPolicy` to specify how the old Snapshots should be cleaned up. When a full backup is deleted, its LEVEL 1 and archived log backups are deleted with it, so this also decides how far back you can restore.

Below is the YAML of the `RetentionPolicy` object that we are going to create,

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

Let's create the above `RetentionPolicy`,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/oracle/pitr/retentionpolicy.yaml
retentionpolicy.storage.kubestash.com/demo-retention created
```

**Create Encryption Secret:**

Let's create a secret called `encrypt-secret` with the `Restic` password,

```bash
$ echo -n 'changeit' > RESTIC_PASSWORD
$ kubectl create secret generic -n demo encrypt-secret \
    --from-file=./RESTIC_PASSWORD
secret/encrypt-secret created
```

**Create OracleArchiver CR:**

`OracleArchiver` is a CR provided by KubeDB for managing the continuous archiving (scheduled full backups plus incremental backups) of `Oracle` databases.

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

Let's create the above `OracleArchiver`,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/oracle/pitr/oracle-archiver.yaml
oraclearchiver.archiver.kubedb.com/oracle-archiver created
```

Here,
- `spec.databases.namespaces` and `spec.databases.selector` select the target `Oracle` databases this archiver applies to (by namespace and label). Here, every `Oracle` in the `demo` namespace with the label `archiver: "true"`.
- `spec.fullBackup.scheduler.schedule` specifies how often a new LEVEL 0 full backup is taken. Here, every 6 hours.
- `spec.incrementalBackup.archivelogInterval` specifies how often the archived redo logs are shipped. Here, every minute. This is roughly how much data you can lose.
- `spec.incrementalBackup.backupInterval` specifies how often a LEVEL 1 incremental backup is taken. Here, every 10 minutes.
- `spec.backupStorage.ref` and `spec.retentionPolicy` reuse the `BackupStorage` and `RetentionPolicy` we created earlier.

> The `KubeDB` provisioner uses `KubeStash` for the scheduled full backups and a `sidekick` pod for the incremental backups, both using the `OracleBackup` driver, which runs Oracle's [RMAN](https://docs.oracle.com/en/database/oracle/oracle-database/21/bradv/) inside the database pod. A point-in-time restore uses RMAN too: it restores the full backup, applies the LEVEL 1 backups, and replays the archived logs up to the time you choose.

### Deploy Sample Oracle Database

Below is the YAML of a sample `Oracle` CR that we are going to create for this tutorial. Notice the `archiver: "true"` label, which the `OracleArchiver`'s `spec.databases.selector` matches:

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

Create the above `Oracle` CR,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/oracle/pitr/sample-oracle.yaml
oracle.kubedb.com/sample-oracle created
```

Wait until the database goes into `Ready` state,

```bash
$ kubectl get oracle -n demo sample-oracle
NAME            VERSION   MODE         STATUS   AGE
sample-oracle   21.3.0    Standalone   Ready    9m
```

The KubeDB operator attaches the `OracleArchiver` to the database and records it in `spec.archiver.ref`,

```bash
$ kubectl get oracle -n demo sample-oracle -o jsonpath='{.spec.archiver}'
{"ref":{"name":"oracle-archiver","namespace":"demo"}}
```

### Verify Backup Setup Successful

If everything goes well, the KubeDB provisioner will create a `BackupConfiguration` for the database, and its phase should be `Ready`. The `Ready` phase indicates that the backup setup is successful.

```bash
$ kubectl get backupconfiguration -n demo
NAME                     PHASE   PAUSED   AGE
sample-oracle-archiver   Ready            3m
```

**Verify BackupSession:**

KubeStash triggers an instant full backup as soon as the `BackupConfiguration` is ready. After that, full backups are taken according to `spec.fullBackup.scheduler.schedule` (every 6 hours here).

```bash
$ kubectl get backupsession -n demo
NAME                                            INVOKER-TYPE          INVOKER-NAME             PHASE       DURATION   AGE
sample-oracle-archiver-full-backup-1795151405   BackupConfiguration   sample-oracle-archiver   Succeeded   2m31s      3m
```

> The first backup of a database prepares an Oracle wallet that lets RMAN reach the bucket. It restarts the database pod once.

**Verify Snapshot:**

```bash
$ kubectl get snapshots -n demo -l=kubestash.com/repo-name=sample-oracle-full
NAME                                                               REPOSITORY           SESSION       SNAPSHOT-TIME          DELETION-POLICY   PHASE       AGE
sample-oracle-full-sample-oracle-archiver-full-backup-1795151405   sample-oracle-full   full-backup   2026-11-20T05:10:05Z   Delete            Succeeded   3m
sample-oracle-incremental-snapshot                                 sample-oracle-full   full-backup                          Delete            Pending     3m
```

Here,
- `sample-oracle-full-sample-oracle-archiver-full-backup-*` is the LEVEL 0 full backup.
- `sample-oracle-incremental-snapshot` is not a backup. It records the history of the incremental backups.

**Verify Incremental Backup:**

Let's check the pods which are related to the backup,

```bash
$ kubectl get pods -n demo
NAME                                                                READY   STATUS      RESTARTS   AGE
retention-policy-sample-oracle-archiver-full-backup-17951514wx8fk   0/1     Completed   0          10m
sample-oracle-0                                                     1/1     Running     0          8m
sample-oracle-archiver-full-backup-1795151405-n4m4t                 0/1     Completed   0          10m
sample-oracle-sidekick                                              1/1     Running     0          11m
```

Here,
- Pod `sample-oracle-archiver-full-backup-*` performed the full backup.
- Pod `sample-oracle-sidekick` performs the incremental backups. It stays in `Running` phase, ships the archived logs every minute, and takes a LEVEL 1 backup every 10 minutes.

You can follow the cycles in the `sidekick` logs. Each archived log cycle tells you the latest point the database can now be restored to,

```bash
$ kubectl logs -n demo sample-oracle-sidekick | grep -E "Initiating|shipped"
I1120 05:12:50.377566       1 incremental_backup.go:153] Initiating incremental backup loop for BackupConfiguration demo/sample-oracle-archiver: LEVEL 1 every 10m0s, archived logs every 1m0s...
I1120 05:14:01.551623       1 archivelog_backup.go:176] shipped 1 archived log(s) under RMAN tag KI-D8480D9C-A20261120T051350; snapshot sample-oracle-full-sample-oracle-archiver-full-backup-1795151405 is restorable up to 2026-11-20T05:13:41Z (SCN 3712408)
I1120 05:15:12.633007       1 archivelog_backup.go:176] shipped 2 archived log(s) under RMAN tag KI-D8480D9C-A20261120T051501; snapshot sample-oracle-full-sample-oracle-archiver-full-backup-1795151405 is restorable up to 2026-11-20T05:14:52Z (SCN 3712986)
```

### Insert Data

Now, let's create a table and insert some data. We will connect to the database with `sqlplus`.

```bash
$ kubectl exec -it -n demo sample-oracle-0 -c oracle -- sqlplus / as sysdba

SQL> CREATE TABLE orders (id INT PRIMARY KEY, item VARCHAR2(30));

Table created.

SQL> INSERT INTO orders VALUES (1, 'keyboard');
SQL> INSERT INTO orders VALUES (2, 'mouse');
SQL> INSERT INTO orders VALUES (3, 'monitor');
SQL> COMMIT;

Commit complete.
```

A little later, insert a second batch,

```bash
SQL> INSERT INTO orders VALUES (4, 'laptop');
SQL> INSERT INTO orders VALUES (5, 'headset');
SQL> COMMIT;

Commit complete.

SQL> SELECT * FROM orders ORDER BY id;

        ID ITEM
---------- ------------------------------
         1 keyboard
         2 mouse
         3 monitor
         4 laptop
         5 headset
```

**This is the state we are going to restore to later.** Let's note the current time in UTC. We will use it as the recovery timestamp,

```bash
SQL> SELECT TO_CHAR(SYS_EXTRACT_UTC(SYSTIMESTAMP), 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS now_utc FROM dual;

NOW_UTC
--------------------
2026-11-20T05:41:30Z
```

Now, let's make a mistake. We insert two more orders, and then delete order `1` by accident,

```bash
SQL> INSERT INTO orders VALUES (6, 'webcam');
SQL> INSERT INTO orders VALUES (7, 'speaker');
SQL> DELETE FROM orders WHERE id = 1;
SQL> COMMIT;

Commit complete.

SQL> SELECT * FROM orders ORDER BY id;

        ID ITEM
---------- ------------------------------
         2 mouse
         3 monitor
         4 laptop
         5 headset
         6 webcam
         7 speaker

SQL> exit
```

The original database now has 6 orders and has lost order `1`. We are **not** going to touch the original `sample-oracle` database. Instead, we will restore a new database to `2026-11-20T05:41:30Z`, which should have orders `1` to `5` only.

Before restoring, make sure the backups already reach the recovery timestamp. Wait for the next archived log cycle, then check that the `restorable up to` time in the `sidekick` logs is **later** than `2026-11-20T05:41:30Z`,

```bash
$ kubectl logs -n demo sample-oracle-sidekick --tail=1
I1120 05:43:01.233510       1 archivelog_backup.go:176] shipped 1 archived log(s) under RMAN tag KI-D8480D9C-A20261120T054250; snapshot sample-oracle-full-sample-oracle-archiver-full-backup-1795151405 is restorable up to 2026-11-20T05:42:41Z (SCN 3725894)
```

You can also read the restorable window from the full backup `Snapshot`. `oracle.kubedb.com/until-time` is the earliest point you can restore to, and `oracle.kubedb.com/log-until-time` is the latest,

```bash
$ kubectl get snapshot -n demo sample-oracle-full-sample-oracle-archiver-full-backup-1795151405 -o yaml | grep -E "until-time"
    oracle.kubedb.com/log-until-time: "2026-11-20T05:42:41Z"
    oracle.kubedb.com/until-time: "2026-11-20T05:13:00Z"
```

## Point-in-time Recovery

Point-In-Time Recovery allows you to restore an `Oracle` database to a specific point in time using the continuously archived backups. This is useful when you need to recover the state just before a mistake, like the accidental delete above, without touching the original database.

We use the recovery timestamp `2026-11-20T05:41:30Z`, which falls **after** the second batch was committed and **before** the accidental delete.

**Create Restored Oracle CR:**

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Oracle
metadata:
  name: restored-oracle
  namespace: demo
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
  init:
    archiver:
      recoveryTimestamp: "2026-11-20T05:41:30Z"
      fullDBRepository:
        name: sample-oracle-full
        namespace: demo
      encryptionSecret:
        name: encrypt-secret
        namespace: demo
  deletionPolicy: WipeOut
```

Here,
- `spec.init.archiver.recoveryTimestamp` specifies the point in time to recover to, in UTC. KubeDB picks the full backup whose restorable window covers this time, restores it, and replays every change committed before it.
- `spec.init.archiver.fullDBRepository` refers to the `Repository` created by the `OracleArchiver` (named `<Oracle-name>-full` by convention, i.e. `sample-oracle-full`).
- `spec.init.archiver.encryptionSecret` refers to the same encryption secret used by the `OracleArchiver`.
- The restored database must be a `Standalone` database with the same `version` and `edition` as the original, and at least as much storage.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/oracle/pitr/restored-oracle.yaml
oracle.kubedb.com/restored-oracle created
```

Let's watch the database. The restore starts once the new database has finished its first-time setup, which takes a few minutes,

```bash
$ kubectl get oracle -n demo restored-oracle -w
NAME              VERSION   MODE         STATUS          AGE
restored-oracle   21.3.0    Standalone   Provisioning    10s
restored-oracle   21.3.0    Standalone   DataRestoring   2m
restored-oracle   21.3.0    Standalone   Ready           11m
```

Let's check the pods which are related to the restore,

```bash
$ kubectl get pods -n demo | grep restored-oracle
restored-oracle-0                                                   1/1     Running     0          11m
restored-oracle-archiver-restore-flgxp                              0/1     Completed   0          9m
```

Here,
- Pod `restored-oracle-archiver-restore-*` restores the full backup whose window covers `recoveryTimestamp`, applies its LEVEL 1 backups, and replays the archived logs up to `recoveryTimestamp`.

Let's check the underlying `RestoreSession`, which the KubeDB operator created for us,

```bash
$ kubectl get restoresession -n demo
NAME                               REPOSITORY           PHASE       DURATION   AGE
restored-oracle-archiver-restore   sample-oracle-full   Succeeded   7m10s      9m
```

The restore job first waits for the new database to finish its first-time setup, then prepares its Oracle wallet (which restarts the database pod once), and then runs RMAN. You can see which backup was picked and where the recovery stopped in the restore pod's logs,

```bash
$ kubectl logs -n demo restored-oracle-archiver-restore-flgxp | grep -E "restoring to|recovering|completed successfully|recovered to"
I1120 05:47:18.105812       1 restore.go:170] restoring to 2026-11-20T05:41:30Z from snapshot sample-oracle-full-sample-oracle-archiver-full-backup-1795151405, restorable from 2026-11-20T05:13:00Z to 2026-11-20T05:46:52Z; RMAN target time 2026-11-20 05:41:30 in the database clock
I1120 05:53:44.248145       1 restore.go:265] recovering to the requested point in time 2026-11-20 05:41:30
I1120 05:54:24.527892       1 executor.go:184] oracle physical restore completed successfully on pod demo/restored-oracle-0
I1120 05:54:31.856646       1 restore.go:207] database recovered to SCN 3725511 (2026-11-20T05:41:30Z)
```

> **Note:** If the restore fails, for example because the recovery timestamp is outside the restorable window, the `RestoreSession` stays `Failed`. Fix the cause, then delete the `RestoreSession`; KubeDB creates it again and retries.

#### Verify Restored Data:

At first, check if the database has gone into `Ready` state by the following command,

```bash
$ kubectl get oracle -n demo restored-oracle
NAME              VERSION   MODE         STATUS   AGE
restored-oracle   21.3.0    Standalone   Ready    11m
```

Now, let's exec into the pod to verify the restored data. The **original** `sample-oracle` database still has orders `2` to `7`, while the **restored** `restored-oracle` database, restored to `2026-11-20T05:41:30Z`, should have orders `1` to `5`.

```bash
$ kubectl exec -it -n demo restored-oracle-0 -c oracle -- sqlplus / as sysdba

SQL> SELECT * FROM orders ORDER BY id;

        ID ITEM
---------- ------------------------------
         1 keyboard
         2 mouse
         3 monitor
         4 laptop
         5 headset

SQL> SELECT open_mode FROM v$database;

OPEN_MODE
--------------------
READ WRITE

SQL> exit
```

As shown above, `restored-oracle` has exactly the 5 orders that existed at `2026-11-20T05:41:30Z`. Order `1`, which was deleted after the recovery timestamp, is back, and orders `6` and `7`, which were inserted after it, are correctly **not** present. The original `sample-oracle` database is unchanged. This confirms that `spec.init.archiver.recoveryTimestamp` recovers the database to the specified point in time, rather than to the latest backup.

> **Note:** The restored database does not have the `archiver: "true"` label, so it is not backed up. To back it up, add the label once it is `Ready`.

### Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete oracle -n demo restored-oracle
kubectl delete oracle -n demo sample-oracle
kubectl delete oraclearchiver -n demo oracle-archiver
kubectl delete repository -n demo sample-oracle-full
kubectl delete retentionpolicies.storage.kubestash.com -n demo demo-retention
kubectl delete backupstorage -n demo s3-storage
kubectl delete secret -n demo s3-secret encrypt-secret
```

## Next Steps

- Learn how the [incremental backups](/docs/guides/oracle/backup/incremental/index.md) behind point-in-time recovery work, and how to tune them.
- Take a [base backup](/docs/guides/oracle/backup/base-backup/index.md) of an Oracle database and restore it.
