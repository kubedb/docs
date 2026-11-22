---
title: Base Backup & Restore Oracle | KubeStash
description: Take a full physical backup of an Oracle database with KubeStash and restore it
menu:
  docs_{{ .version }}:
    identifier: guides-oracle-backup-base
    name: Base Backup
    parent: guides-oracle-backup
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Base Backup and Restore Oracle using KubeStash

A base backup is a full physical backup of an `Oracle` database, taken by RMAN while the database is online. This guide shows how to take a base backup of an `Oracle` database with KubeStash, and how to restore it.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).
- Install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md).
- Install `KubeStash` in your cluster following the steps [here](https://kubestash.com/docs/latest/setup/install/kubestash).
- If you are not familiar with how KubeStash backs up Oracle databases, please check the [overview](/docs/guides/oracle/backup/overview/index.md) first.

You should be familiar with the following `KubeStash` concepts:

- [BackupStorage](https://kubestash.com/docs/latest/concepts/crds/backupstorage/)
- [BackupConfiguration](https://kubestash.com/docs/latest/concepts/crds/backupconfiguration/)
- [BackupSession](https://kubestash.com/docs/latest/concepts/crds/backupsession/)
- [RestoreSession](https://kubestash.com/docs/latest/concepts/crds/restoresession/)
- [Addon](https://kubestash.com/docs/latest/concepts/crds/addon/)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial are stored in [docs/examples/oracle/backup](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/oracle/backup) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

> Oracle images are pulled from `container-registry.oracle.com`. Every Oracle CR must reference an image pull secret (named `orclcred` in this tutorial) through `spec.podTemplate.spec.imagePullSecrets`. If you haven't created one yet, follow the steps [here](/docs/guides/oracle/quickstart#create-oracle-image-pull-secret-important).

## Backup Oracle

### Deploy Sample Oracle Database

Below is the YAML of a sample `Oracle` database that we are going to back up,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Oracle
metadata:
  name: sample-oracle
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
  deletionPolicy: WipeOut
```

Let's create the `Oracle` CR,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/oracle/backup/sample-oracle.yaml
oracle.kubedb.com/sample-oracle created
```

Wait until the database is `Ready`,

```bash
$ kubectl get oracle -n demo sample-oracle
NAME            VERSION   MODE         STATUS   AGE
sample-oracle   21.3.0    Standalone   Ready    9m
```

> **Important:** On the first boot, the Oracle image keeps creating the database for a few minutes after KubeDB reports it `Ready`. Wait until the pod log shows `DATABASE IS READY TO USE!` before you take the first backup:
>
> ```bash
> $ kubectl logs -n demo sample-oracle-0 -c oracle | grep "DATABASE IS READY TO USE"
> DATABASE IS READY TO USE!
> ```

### Insert Sample Data

Let's connect to the database and create a table with a few rows. We will restore this table later.

```bash
$ kubectl exec -it -n demo sample-oracle-0 -c oracle -- sqlplus / as sysdba

SQL> CREATE TABLE emp (id INT PRIMARY KEY, name VARCHAR2(20));

Table created.

SQL> INSERT INTO emp VALUES (1, 'John');
SQL> INSERT INTO emp VALUES (2, 'Mary');
SQL> INSERT INTO emp VALUES (3, 'Alex');
SQL> COMMIT;

Commit complete.

SQL> SELECT * FROM emp;

        ID NAME
---------- --------------------
         1 John
         2 Mary
         3 Alex

SQL> exit
```

### Prepare Backend

We are going to store the backup in an `S3` bucket. KubeStash needs a `Secret` with the bucket credentials and a `BackupStorage` that points to the bucket.

> **Note:** Oracle backups support `S3` and S3 compatible storage (for example MinIO) only. For MinIO, set `endpoint` to the MinIO service URL, including the scheme and port, for example `http://minio.demo.svc.cluster.local:80`.

**Create Secret:**

Let's create a secret called `s3-secret` with access credentials to the bucket,

```bash
$ echo -n '<your-aws-access-key-id-here>' > AWS_ACCESS_KEY_ID
$ echo -n '<your-aws-secret-access-key-here>' > AWS_SECRET_ACCESS_KEY
$ kubectl create secret generic -n demo s3-secret \
    --from-file=./AWS_ACCESS_KEY_ID \
    --from-file=./AWS_SECRET_ACCESS_KEY
secret/s3-secret created
```

**Create BackupStorage:**

Below is the YAML of the `BackupStorage` CR we are going to create,

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
      prefix: oracle
      secretName: s3-secret
      region: us-east-2
      endpoint: https://s3.us-east-2.amazonaws.com
  usagePolicy:
    allowedNamespaces:
      from: All
  default: true
  deletionPolicy: Delete
```

Let's create the `BackupStorage`,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/oracle/backup/backupstorage.yaml
backupstorage.storage.kubestash.com/s3-storage created
```

> **Note:** The `prefix` applies to KubeStash's own metadata only. RMAN writes the backup pieces at the root of the bucket, so use a bucket dedicated to Oracle backups.

**Create RetentionPolicy:**

A `RetentionPolicy` tells KubeStash how many old backups to keep. Below is the YAML of the `RetentionPolicy` we are going to create,

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

Let's create the `RetentionPolicy`,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/oracle/backup/retentionpolicy.yaml
retentionpolicy.storage.kubestash.com/demo-retention created
```

**Create Encryption Secret:**

KubeStash encrypts its backup metadata with this password. Let's create a secret called `encrypt-secret`,

```bash
$ echo -n 'changeit' > RESTIC_PASSWORD
$ kubectl create secret generic -n demo encrypt-secret \
    --from-file=./RESTIC_PASSWORD
secret/encrypt-secret created
```

### Backup

Now, we are going to create a `BackupConfiguration` that backs up `sample-oracle` every 6 hours. Below is the YAML of the `BackupConfiguration`,

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: sample-oracle-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Oracle
    name: sample-oracle
    namespace: demo
  backends:
    - name: s3-backend
      storageRef:
        name: s3-storage
        namespace: demo
      retentionPolicy:
        name: demo-retention
        namespace: demo
  sessions:
    - name: full-backup
      scheduler:
        schedule: "0 */6 * * *"
        jobTemplate:
          backoffLimit: 1
      repositories:
        - name: oracle-repo
          backend: s3-backend
          directory: /sample-oracle
          encryptionSecret:
            name: encrypt-secret
            namespace: demo
      sessionHistoryLimit: 3
      addon:
        name: oracle-addon
        tasks:
          - name: physical-backup
            params:
              targetPod: sample-oracle-0
```

Here,

- `spec.target` points to the `sample-oracle` database we want to back up.
- `spec.sessions[*].scheduler.schedule` is a cron expression. Here, a backup runs every 6 hours.
- `spec.sessions[*].repositories` must list **exactly one** repository. RMAN writes each backup to a single destination.
- `spec.sessions[*].addon.name` is `oracle-addon`, and the task `physical-backup` takes an RMAN full backup.
- `spec.sessions[*].addon.tasks[*].params.targetPod` is the database pod RMAN runs in. For a `Standalone` database, it is `<database-name>-0`.

Let's create the `BackupConfiguration`,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/oracle/backup/backupconfiguration.yaml
backupconfiguration.core.kubestash.com/sample-oracle-backup created
```

**Verify Backup Setup Successful:**

If everything goes well, the phase of the `BackupConfiguration` should be `Ready`,

```bash
$ kubectl get backupconfiguration -n demo
NAME                   PHASE   PAUSED   AGE
sample-oracle-backup   Ready            12s
```

**Verify BackupSession:**

KubeStash triggers a backup as soon as the `BackupConfiguration` is ready. Let's watch the `BackupSession`,

```bash
$ kubectl get backupsession -n demo -w
NAME                                         INVOKER-TYPE          INVOKER-NAME           PHASE       DURATION   AGE
sample-oracle-backup-full-backup-1795154400   BackupConfiguration   sample-oracle-backup   Running                0s
sample-oracle-backup-full-backup-1795154400   BackupConfiguration   sample-oracle-backup   Succeeded   2m35s      2m35s
```

The first backup of a database takes a little longer. Before RMAN starts, the backup job prepares the Oracle wallet that lets RMAN reach the bucket. It creates an `OracleOpsRequest` that applies the wallet to the database, and the database pod restarts once:

```bash
$ kubectl get oracleopsrequest -n demo
NAME                                                                             TYPE          STATUS       AGE
oracle-wallet-create-sample-oracle-sample-oracle-backup-full-backup-1795154400   Reconfigure   Successful   2m40s
```

Later backups reuse the wallet, do not restart the database, and finish faster (about a minute for this database).

**Verify Snapshot:**

Each successful backup creates a `Snapshot`,

```bash
$ kubectl get snapshots -n demo -l=kubestash.com/repo-name=oracle-repo
NAME                                                     REPOSITORY    SESSION       SNAPSHOT-TIME          DELETION-POLICY   PHASE       AGE
oracle-repo-sample-oracle-backup-full-backup-1795154400   oracle-repo   full-backup   2026-11-20T06:00:00Z   Delete            Succeeded   3m
```

The `Snapshot` also carries the details RMAN needs to restore this backup: the database ID, the SCN the backup is consistent at, and the control file and SPFILE pieces.

```bash
$ kubectl get snapshot -n demo oracle-repo-sample-oracle-backup-full-backup-1795154400 -o yaml | grep oracle.kubedb.com
    oracle.kubedb.com/controlfile-handle: 7g575ke3_240_1_1
    oracle.kubedb.com/dbid: "1777412365"
    oracle.kubedb.com/spfile-handle: 7h575ke4_241_1_1
    oracle.kubedb.com/until-scn: "3676964"
```

## Restore

In this section, we are going to simulate a data loss and restore the backup into the same database.

### Simulate Data Loss

Let's drop the `emp` table,

```bash
$ kubectl exec -it -n demo sample-oracle-0 -c oracle -- sqlplus / as sysdba

SQL> DROP TABLE emp PURGE;

Table dropped.

SQL> SELECT * FROM emp;
SELECT * FROM emp
              *
ERROR at line 1:
ORA-00942: table or view does not exist

SQL> exit
```

### Create RestoreSession

Now, we are going to create a `RestoreSession` that restores the `Snapshot` we took earlier. Below is the YAML of the `RestoreSession`,

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: sample-oracle-restore
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Oracle
    name: sample-oracle
    namespace: demo
  dataSource:
    repository: oracle-repo
    snapshot: oracle-repo-sample-oracle-backup-full-backup-1795154400
    encryptionSecret:
      name: encrypt-secret
      namespace: demo
  addon:
    name: oracle-addon
    tasks:
      - name: physical-backup-restore
```

Here,

- `spec.target` is the database to restore into. Here, it is the same `sample-oracle` database.
- `spec.dataSource.snapshot` is the `Snapshot` to restore. Replace it with the name of your own `Snapshot`.
- `spec.addon.tasks[*].name` is `physical-backup-restore`, which runs an RMAN restore and recovery.

By default, the restore brings the database back to exactly the moment the backup finished. The `physical-backup-restore` task also takes these optional `params`:

| Parameter | Default | Description |
|---|---|---|
| `untilTime` | empty | Advanced. Stop recovery at this UTC time (`YYYY-MM-DD HH24:MI:SS`) instead of the end of the backup. RMAN needs archived logs that reach this time, so to restore to any point in time, use an [OracleArchiver](/docs/guides/oracle/pitr/archiver.md) instead. |
| `untilSCN` | empty | Advanced. Stop recovery at this SCN instead of the end of the backup. Do not set it together with `untilTime`. |
| `restoreSPFile` | `false` | When restoring into a different database, also restore the source database's SPFILE (its initialization parameters). |

> **Warning:** The restore shuts down the target database, replaces all of its data, and opens it with `RESETLOGS`. Any change made after the backup is lost.

Let's create the `RestoreSession`,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/oracle/backup/restoresession.yaml
restoresession.core.kubestash.com/sample-oracle-restore created
```

Wait for the `RestoreSession` to succeed,

```bash
$ kubectl get restoresession -n demo -w
NAME                    REPOSITORY    PHASE       DURATION   AGE
sample-oracle-restore   oracle-repo   Running                3s
sample-oracle-restore   oracle-repo   Succeeded   35s        35s
```

### Verify Restored Data

Check that the database is `Ready` again,

```bash
$ kubectl get oracle -n demo sample-oracle
NAME            VERSION   MODE         STATUS   AGE
sample-oracle   21.3.0    Standalone   Ready    31m
```

Now, let's check the `emp` table,

```bash
$ kubectl exec -it -n demo sample-oracle-0 -c oracle -- sqlplus / as sysdba

SQL> SELECT * FROM emp;

        ID NAME
---------- --------------------
         1 John
         2 Mary
         3 Alex

SQL> SELECT open_mode FROM v$database;

OPEN_MODE
--------------------
READ WRITE

SQL> exit
```

The `emp` table is back with all of its rows, so the backup has been restored successfully.

### Restore into a Different Database

To restore a backup into a different `Oracle` database instead, for example to make a copy, first create the new `Oracle` with the same `version` and `edition`. Then set `spec.target.name` of the `RestoreSession` to the new database. The new database gets the data and the database ID of the backed up one, and keeps its own initialization parameters unless you set `restoreSPFile: "true"`.

## Cleanup

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete restoresession -n demo sample-oracle-restore
kubectl delete backupconfiguration -n demo sample-oracle-backup
kubectl delete oracle -n demo sample-oracle
kubectl delete retentionpolicy -n demo demo-retention
kubectl delete backupstorage -n demo s3-storage
kubectl delete secret -n demo s3-secret encrypt-secret
kubectl delete ns demo
```

## Next Steps

- Set up [incremental backup](/docs/guides/oracle/backup/incremental/index.md) so you can restore to any point in time.
- Restore an Oracle database to a [point in time](/docs/guides/oracle/pitr/archiver.md).
