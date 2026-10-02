---
title: Continuous Archiving and Point-in-time Recovery
menu:
  docs_{{ .version }}:
    identifier: pitr-clickhouse-archiver
    name: Overview
    parent: pitr-clickhouse
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# KubeDB ClickHouse - Continuous Archiving and Point-in-time Recovery

Here, we will show you how to use KubeDB to continuously archive a `ClickHouse` database and restore it to a specific point in time.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now,
- install `KubeDB` operator in your cluster following the steps [here](/docs/setup/README.md).
- install `KubeStash` operator in your cluster following the steps [here](https://github.com/kubestash/installer/tree/master/charts/kubestash).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```
> Note: The yaml files used in this tutorial are stored in [docs/guides/clickhouse/pitr/examples](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/clickhouse/pitr/examples) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Continuous Archiving

Continuous archiving involves taking a periodic full backup of the `ClickHouse` cluster and, in between full backups, continuously archiving the incremental changes. To ensure continuous archiving to a remote location we need to prepare a `BackupStorage`, a `RetentionPolicy`, and a `ClickHouseArchiver` for the KubeDB managed `ClickHouse` database.

### Prepare Backend

We are going to store our backed up data into an `S3` bucket. We have to create a `Secret` with necessary credentials and a `BackupStorage` CR to use this backend. If you want to use a different backend, please read the respective backend configuration doc from [here](https://kubestash.com/docs/latest/guides/backends/overview/).

> **Note:** ClickHouse currently supports `S3`, `Azure Blob Storage`, and `Google Cloud Storage` (via S3 compatibility mode) as backup storage backends.

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
      bucket: kubestash
      prefix: clickhouse-pitr
      secretName: s3-secret
      region: us-east-1
      endpoint: http://minio.demo.svc.cluster.local:80
  usagePolicy:
    allowedNamespaces:
      from: All
  default: true
  deletionPolicy: Delete
```

Let's create the BackupStorage we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/clickhouse/pitr/examples/backupstorage.yaml
backupstorage.storage.kubestash.com/s3-storage created
```

**Create RetentionPolicy:**

Now, let's create a `RetentionPolicy` to specify how the old Snapshots should be cleaned up.

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
    last: 2
  usagePolicy:
    allowedNamespaces:
      from: All
```

Let's create the above `RetentionPolicy`,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/clickhouse/pitr/examples/retentionpolicy.yaml
retentionpolicy.storage.kubestash.com/demo-retention created
```

**Create Encryption Secret**

Let's create a secret called `encrypt-secret` with the `Restic` password,

```bash
$ echo -n 'changeit' > RESTIC_PASSWORD
$ kubectl create secret generic -n demo encrypt-secret \
    --from-file=./RESTIC_PASSWORD
secret/encrypt-secret created
```

**Create ClickHouseArchiver CR:**

`ClickHouseArchiver` is a CR provided by KubeDB for managing the continuous archiving (periodic full backup plus incremental backups) of `ClickHouse` databases.

```yaml
apiVersion: archiver.kubedb.com/v1alpha1
kind: ClickHouseArchiver
metadata:
  name: sample-clickhouse-archiver
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
    name: demo-retention
    namespace: demo
  encryptionSecret:
    name: "encrypt-secret"
    namespace: "demo"
  fullBackup:
    driver: "ClickHouseBackup"
    scheduler:
      successfulJobsHistoryLimit: 1
      failedJobsHistoryLimit: 1
      schedule: "*/30 * * * *"
    sessionHistoryLimit: 2
  backupStorage:
    ref:
      name: "s3-storage"
      namespace: "demo"
```

Let's create the above `ClickHouseArchiver`,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/clickhouse/pitr/examples/sample-clickhouse-archiver.yaml
clickhousearchiver.archiver.kubedb.com/sample-clickhouse-archiver created
```

Here,
- `spec.databases.namespaces` and `spec.databases.selector` select the target `ClickHouse` databases this archiver applies to (by namespace and label). A `ClickHouse` CR can alternatively (or additionally) reference an archiver directly via its own `spec.archiver.ref` field, as shown below.
- `spec.fullBackup.scheduler.schedule` specifies how often a new full backup is taken. In between full backups, KubeStash continuously archives incremental changes into the same repository.
- `spec.backupStorage.ref` and `spec.retentionPolicy` reuse the `BackupStorage` and `RetentionPolicy` we created earlier.

> The `KubeDB` provisioner uses `KubeStash` for the periodic full backup and a `sidekick` container for continuous incremental archiving of `ClickHouse` databases, both using the `ClickHouseBackup` driver, which runs ClickHouse's native [`BACKUP`](https://clickhouse.com/docs/operations/backup) command. Incremental backups are taken with the native `BACKUP` command using the latest full backup as the `base_backup`, and the data is restored with the native `RESTORE` command.

### Deploy Sample ClickHouse Database

Below is the YAML of a sample `ClickHouse` CR that we are going to create for this tutorial. Notice the `archiver: "true"` label (matched by the `ClickHouseArchiver`'s `spec.databases.selector`) and the explicit `spec.archiver.ref` pointing to the `ClickHouseArchiver` we created above:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ClickHouse
metadata:
  name: sample-clickhouse
  namespace: demo
  labels:
    archiver: "true"
spec:
  version: 25.7.1
  archiver:
    ref:
      name: sample-clickhouse-archiver
      namespace: demo
  clusterTopology:
    clickHouseKeeper:
      externallyManaged: false
      spec:
        replicas: 3
        storage:
          storageClassName: "local-path"
          accessModes:
            - ReadWriteOnce
          resources:
            requests:
              storage: 1Gi
    cluster:
      name: appscode-cluster
      shards: 2
      replicas: 2
      podTemplate:
        spec:
          containers:
            - name: clickhouse
              resources:
                limits:
                  memory: 4Gi
                requests:
                  cpu: 1
                  memory: 2Gi
          initContainers:
            - name: clickhouse-init
              resources:
                limits:
                  memory: 1Gi
                requests:
                  cpu: 500m
                  memory: 512Mi
      storage:
        storageClassName: "local-path"
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
  deletionPolicy: WipeOut
```

Create the above `ClickHouse` CR,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/clickhouse/pitr/examples/sample-clickhouse.yaml
clickhouse.kubedb.com/sample-clickhouse created
```

Wait until the database goes into `Ready` state,

```bash
$ kubectl get clickhouse -n demo sample-clickhouse
NAME                VERSION   STATUS   AGE
sample-clickhouse   25.7.1    Ready    61s
```

### Insert Initial Data

Since `sample-clickhouse` is deployed with a `clusterTopology`, we create a `ReplicatedMergeTree` table (replicated within a shard via `ClickHouseKeeper`) together with a `Distributed` table on top of it (to read across both shards) — a plain `MergeTree` table would live on a single node only and wouldn't exercise the cluster at all.

```bash
$ kubectl get secret -n demo sample-clickhouse-auth -o jsonpath='{.data.username}' | base64 -d
admin⏎

$ kubectl get secret -n demo sample-clickhouse-auth -o jsonpath='{.data.password}' | base64 -d
psL)PHO!)ryZEcoX⏎

$ kubectl exec -it -n demo sample-clickhouse-appscode-cluster-shard-0-0 -- clickhouse-client --user admin --password 'psL)PHO!)ryZEcoX'

:) CREATE DATABASE IF NOT EXISTS demo_db ON CLUSTER '{cluster}';

:) CREATE TABLE demo_db.users ON CLUSTER '{cluster}'
   (
       id UInt64,
       name String,
       shard_id UInt8
   )
   ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}/demo_db/users', '{replica}')
   ORDER BY id;

:) CREATE TABLE demo_db.users_dist ON CLUSTER '{cluster}' AS demo_db.users
   ENGINE = Distributed('{cluster}', demo_db, users, rand());

:) exit
```

To control exactly which shard each row lives on, we insert directly into the local `demo_db.users` table of one replica of each shard; `ReplicatedMergeTree` then replicates the rows to the other replica of that shard.

Let's exec into the first replica of shard `0` and insert the rows of shard `0`,

```bash
$ kubectl exec -it -n demo sample-clickhouse-appscode-cluster-shard-0-0 -- clickhouse-client --user admin --password 'psL)PHO!)ryZEcoX'

:) INSERT INTO demo_db.users VALUES (1, 'Alice_shard0', 0), (2, 'Bob_shard0', 0);

:) SELECT * FROM demo_db.users ORDER BY id;

   ┌─id─┬─name─────────┬─shard_id─┐
1. │  1 │ Alice_shard0 │        0 │
2. │  2 │ Bob_shard0   │        0 │
   └────┴──────────────┴──────────┘

:) exit
```

Now, exec into the first replica of shard `1` and insert the rows of shard `1`,

```bash
$ kubectl exec -it -n demo sample-clickhouse-appscode-cluster-shard-1-0 -- clickhouse-client --user admin --password 'psL)PHO!)ryZEcoX'

:) INSERT INTO demo_db.users VALUES (3, 'Charlie_shard1', 1), (4, 'David_shard1', 1);

:) SELECT * FROM demo_db.users ORDER BY id;

   ┌─id─┬─name───────────┬─shard_id─┐
1. │  3 │ Charlie_shard1 │        1 │
2. │  4 │ David_shard1   │        1 │
   └────┴────────────────┴──────────┘

:) exit
```

Let's verify that the rows have been replicated to the second replica of each shard,

```bash
$ kubectl exec -it -n demo sample-clickhouse-appscode-cluster-shard-0-1 -- clickhouse-client --user admin --password 'psL)PHO!)ryZEcoX'

:) SELECT * FROM demo_db.users ORDER BY id;

   ┌─id─┬─name─────────┬─shard_id─┐
1. │  1 │ Alice_shard0 │        0 │
2. │  2 │ Bob_shard0   │        0 │
   └────┴──────────────┴──────────┘

:) exit
```

```bash
$ kubectl exec -it -n demo sample-clickhouse-appscode-cluster-shard-1-1 -- clickhouse-client --user admin --password 'psL)PHO!)ryZEcoX'

:) SELECT * FROM demo_db.users ORDER BY id;

   ┌─id─┬─name───────────┬─shard_id─┐
1. │  3 │ Charlie_shard1 │        1 │
2. │  4 │ David_shard1   │        1 │
   └────┴────────────────┴──────────┘

:) exit
```

And through the `Distributed` table, which returns the rows from both shards,

```bash
$ kubectl exec -it -n demo sample-clickhouse-appscode-cluster-shard-0-0 -- clickhouse-client --user admin --password 'psL)PHO!)ryZEcoX'

:) SELECT * FROM demo_db.users_dist ORDER BY id;

   ┌─id─┬─name───────────┬─shard_id─┐
1. │  1 │ Alice_shard0   │        0 │
2. │  2 │ Bob_shard0     │        0 │
3. │  3 │ Charlie_shard1 │        1 │
4. │  4 │ David_shard1   │        1 │
   └────┴────────────────┴──────────┘

:) exit
```

### Verify Backup Setup Successful

If everything goes well, the KubeDB provisioner will create a `BackupConfiguration` for the database, and its phase should be `Ready`. The `Ready` phase indicates that the backup setup is successful.

```bash
$ kubectl get backupconfiguration -n demo
NAME                         PHASE   PAUSED   AGE
sample-clickhouse-archiver   Ready            15m
```

**Verify BackupSession:**

KubeStash triggers an instant full backup as soon as the `BackupConfiguration` is ready. After that, full backups are taken according to `spec.fullBackup.scheduler.schedule` (every 30 minutes here), while incremental backups are taken continuously in between.

> **Note:** You can also trigger a full backup manually at any time, but we are not doing that here. In this tutorial, we simply rely on the instant backup, the full backup schedule and the continuous incremental backups.

```bash
$ kubectl get backupsession -n demo
NAME                                                INVOKER-TYPE          INVOKER-NAME                 PHASE       DURATION   AGE
sample-clickhouse-archiver-full-backup-1790745995   BackupConfiguration   sample-clickhouse-archiver   Succeeded   34s        14m
sample-clickhouse-archiver-full-backup-1790746201   BackupConfiguration   sample-clickhouse-archiver   Succeeded   30s        11m
```

Here, the first `BackupSession` is the instant full backup taken right after the `BackupConfiguration` became ready, and the second one was taken by the `*/30 * * * *` schedule.

**Verify Snapshot:**

```bash
$ kubectl get snapshots -n demo -l=kubestash.com/repo-name=sample-clickhouse-full
NAME                                                              REPOSITORY               SESSION       SNAPSHOT-TIME          DELETION-POLICY   PHASE       AGE
sample-clickhouse-full-sample-clarchiver-full-backup-1790745995   sample-clickhouse-full   full-backup   2026-09-30T05:26:46Z   Delete            Succeeded   14m
sample-clickhouse-full-sample-clarchiver-full-backup-1790746201   sample-clickhouse-full   full-backup   2026-09-30T05:30:01Z   Delete            Succeeded   11m
```

**Verify Incremental Backup:**

Let's check the pods which are related to the backup,

```bash
$ kubectl get pods -n demo
NAME                                                              READY   STATUS      RESTARTS   AGE
retention-policy-sample-clickhouse-archiver-full-bac-179079knjw   0/1     Completed   0          11m
retention-policy-sample-clickhouse-archiver-full-bac-17907z92h8   0/1     Completed   0          14m
sample-clickhouse-appscode-cluster-shard-0-0                      1/1     Running     0          16m
sample-clickhouse-appscode-cluster-shard-0-1                      1/1     Running     0          16m
sample-clickhouse-appscode-cluster-shard-1-0                      1/1     Running     0          16m
sample-clickhouse-appscode-cluster-shard-1-1                      1/1     Running     0          16m
sample-clickhouse-archiver-full-backup-1790745995-9lh4m           0/1     Completed   0          14m
sample-clickhouse-archiver-full-backup-1790746201-ggjd9           0/1     Completed   0          11m
sample-clickhouse-keeper-0                                        1/1     Running     0          16m
sample-clickhouse-keeper-1                                        1/1     Running     0          16m
sample-clickhouse-keeper-2                                        1/1     Running     0          16m
sample-clickhouse-sidekick                                        1/1     Running     0          3m37s
trigger-sample-clickhouse-archiver-full-backup-29845770-ff2nw     0/1     Completed   0          11m
```

Here,
- Pods `sample-clickhouse-archiver-full-backup-*` performed the full backups.
- Pod `sample-clickhouse-sidekick` performs continuous incremental archiving. It stays in `Running` phase and runs an incremental backup cycle roughly every minute, archiving the changes made since the latest full backup.

Each incremental cycle backs up the metadata and every shard. You can follow the cycles in the `sidekick` logs,

```bash
$ kubectl logs -n demo sample-clickhouse-sidekick --tail=12
I0930 05:41:08.574880       1 helpers.go:347] Latest successful snapshot:  sample-clickhouse-full-sample-clarchiver-full-backup-1790746201
I0930 05:41:08.629451       1 output.go:290] Backup started with OperationID=dce60f64-7a38-4f87-9c10-3c01a5b0cd8d InitialStatus=CREATING_BACKUP
I0930 05:41:18.704574       1 output.go:337] Backup daua1t1t3d5c73dcv5b0-metadata completed successfully!
I0930 05:41:18.704594       1 archiver.go:225] Metadata incremental backup completed successfully
I0930 05:41:18.758046       1 output.go:58] Backup started with OperationID=d697a0c8-497f-4a1c-af38-f8ea2d66b34d InitialStatus=CREATING_BACKUP
I0930 05:41:18.758548       1 output.go:58] Backup started with OperationID=a2bf8757-8080-43be-810d-5db522f00d64 InitialStatus=CREATING_BACKUP
I0930 05:41:28.815165       1 output.go:105] Backup daua1t1t3d5c73dcv5b0-shard-1 completed successfully!
I0930 05:41:28.815189       1 archiver.go:274] Data incremental backup completed for shard 1
I0930 05:41:28.815619       1 output.go:105] Backup daua1t1t3d5c73dcv5b0-shard-0 completed successfully!
I0930 05:41:28.815643       1 archiver.go:274] Data incremental backup completed for shard 0
I0930 05:41:28.815656       1 archiver.go:187] Cluster incremental backup completed successfully for all 2 shards
I0930 05:41:28.838303       1 incremental_backup.go:146] Incremental backup cycle completed in 20.271855756s
```

### Insert Incremental Data

Now, let's insert more data into the database over time. Each batch is picked up by the next incremental backup cycle of the `sidekick`.

Insert the second batch,

```bash
$ kubectl exec -it -n demo sample-clickhouse-appscode-cluster-shard-0-0 -- clickhouse-client --user admin --password 'psL)PHO!)ryZEcoX'

:) INSERT INTO demo_db.users VALUES (5, 'Eve_shard0', 0), (6, 'Frank_shard0', 0);

:) exit
```

```bash
$ kubectl exec -it -n demo sample-clickhouse-appscode-cluster-shard-1-0 -- clickhouse-client --user admin --password 'psL)PHO!)ryZEcoX'

:) INSERT INTO demo_db.users VALUES (7, 'Grace_shard1', 1), (8, 'Henry_shard1', 1);

:) exit
```

Insert the third batch,

```bash
$ kubectl exec -it -n demo sample-clickhouse-appscode-cluster-shard-0-0 -- clickhouse-client --user admin --password 'psL)PHO!)ryZEcoX'

:) INSERT INTO demo_db.users VALUES (9, 'Ivy_shard0', 0), (10, 'Jack_shard0', 0), (11, 'Kate_shard0', 0);

:) exit
```

```bash
$ kubectl exec -it -n demo sample-clickhouse-appscode-cluster-shard-1-0 -- clickhouse-client --user admin --password 'psL)PHO!)ryZEcoX'

:) INSERT INTO demo_db.users VALUES (12, 'Leo_shard1', 1), (13, 'Mia_shard1', 1);

:) exit
```

At this point, shard `0` holds ids `1, 2, 5, 6, 9, 10, 11` and shard `1` holds ids `3, 4, 7, 8, 12, 13` (13 rows in total). **This is the state we are going to restore to later.** Wait for the next incremental backup cycle to complete (check the `sidekick` logs as shown above); any time after that cycle and before the next batch is inserted is a valid recovery point for this state. In this run, the third batch was inserted at `05:33:07Z` and was archived by the incremental cycle that started at `05:33:49Z` and completed at `05:34:09Z`.

Some time later (the fourth batch was inserted at `05:35:09Z`), we insert two more batches into the *same, still-running* database — these rows must **not** appear in our restore.

Insert the fourth batch,

```bash
$ kubectl exec -it -n demo sample-clickhouse-appscode-cluster-shard-0-0 -- clickhouse-client --user admin --password 'psL)PHO!)ryZEcoX'

:) INSERT INTO demo_db.users VALUES (14, 'Noah_shard0', 0), (15, 'Olivia_shard0', 0);

:) exit
```

```bash
$ kubectl exec -it -n demo sample-clickhouse-appscode-cluster-shard-1-0 -- clickhouse-client --user admin --password 'psL)PHO!)ryZEcoX'

:) INSERT INTO demo_db.users VALUES (16, 'Paul_shard1', 1), (17, 'Quinn_shard1', 1), (18, 'Riley_shard1', 1);

:) exit
```

Insert the fifth batch,

```bash
$ kubectl exec -it -n demo sample-clickhouse-appscode-cluster-shard-0-0 -- clickhouse-client --user admin --password 'psL)PHO!)ryZEcoX'

:) INSERT INTO demo_db.users VALUES (19, 'Sophia_shard0', 0), (20, 'Thomas_shard0', 0), (21, 'Uma_shard0', 0);

:) exit
```

```bash
$ kubectl exec -it -n demo sample-clickhouse-appscode-cluster-shard-1-0 -- clickhouse-client --user admin --password 'psL)PHO!)ryZEcoX'

:) INSERT INTO demo_db.users VALUES (22, 'Victor_shard1', 1), (23, 'Willow_shard1', 1);

:) exit
```

The original database now has all 23 rows,

```bash
$ kubectl exec -it -n demo sample-clickhouse-appscode-cluster-shard-0-0 -- clickhouse-client --user admin --password 'psL)PHO!)ryZEcoX'

:) SELECT * FROM demo_db.users ORDER BY id;

    ┌─id─┬─name──────────┬─shard_id─┐
 1. │  1 │ Alice_shard0  │        0 │
 2. │  2 │ Bob_shard0    │        0 │
 3. │  5 │ Eve_shard0    │        0 │
 4. │  6 │ Frank_shard0  │        0 │
 5. │  9 │ Ivy_shard0    │        0 │
 6. │ 10 │ Jack_shard0   │        0 │
 7. │ 11 │ Kate_shard0   │        0 │
 8. │ 14 │ Noah_shard0   │        0 │
 9. │ 15 │ Olivia_shard0 │        0 │
10. │ 19 │ Sophia_shard0 │        0 │
11. │ 20 │ Thomas_shard0 │        0 │
12. │ 21 │ Uma_shard0    │        0 │
    └────┴───────────────┴──────────┘

:) exit
```

```bash
$ kubectl exec -it -n demo sample-clickhouse-appscode-cluster-shard-1-0 -- clickhouse-client --user admin --password 'psL)PHO!)ryZEcoX'

:) SELECT * FROM demo_db.users ORDER BY id;

    ┌─id─┬─name───────────┬─shard_id─┐
 1. │  3 │ Charlie_shard1 │        1 │
 2. │  4 │ David_shard1   │        1 │
 3. │  7 │ Grace_shard1   │        1 │
 4. │  8 │ Henry_shard1   │        1 │
 5. │ 12 │ Leo_shard1     │        1 │
 6. │ 13 │ Mia_shard1     │        1 │
 7. │ 16 │ Paul_shard1    │        1 │
 8. │ 17 │ Quinn_shard1   │        1 │
 9. │ 18 │ Riley_shard1   │        1 │
10. │ 22 │ Victor_shard1  │        1 │
11. │ 23 │ Willow_shard1  │        1 │
    └────┴────────────────┴──────────┘

:) exit
```

```bash
$ kubectl exec -it -n demo sample-clickhouse-appscode-cluster-shard-0-0 -- clickhouse-client --user admin --password 'psL)PHO!)ryZEcoX'

:) SELECT count() FROM demo_db.users_dist;

   ┌─count()─┐
1. │      23 │
   └─────────┘

:) exit
```

Note that we are **not** touching or dropping anything in the original `sample-clickhouse` database — it keeps running with all 23 rows. This lets us prove, side-by-side, that the restored database ends up with the *older* 13-row state (still correctly split across shards) instead of the latest data.

## Point-in-time Recovery

Point-In-Time Recovery allows you to restore a `ClickHouse` database to a specific point in time using the continuously archived incremental backups. This is particularly useful in scenarios where you need to recover to a state just before a specific error or unwanted change occurred — without necessarily touching the original database.

We use the recovery timestamp `2026-09-30T05:34:00Z`, which falls **after** the third batch was archived (incremental cycle started at `05:33:49Z`) and **before** the fourth batch was inserted (`05:35:09Z`).

**Create Restored ClickHouse CR:**

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ClickHouse
metadata:
  name: restored-clickhouse-pitr
  namespace: demo
spec:
  init:
    archiver:
      encryptionSecret:
        name: encrypt-secret
        namespace: demo
      fullDBRepository:
        name: sample-clickhouse-full
        namespace: demo
      recoveryTimestamp: "2026-09-30T05:34:00Z"
  version: 25.7.1
  clusterTopology:
    clickHouseKeeper:
      externallyManaged: false
      spec:
        replicas: 3
        storage:
          storageClassName: "local-path"
          accessModes:
            - ReadWriteOnce
          resources:
            requests:
              storage: 1Gi
    cluster:
      name: appscode-cluster
      shards: 2
      replicas: 2
      podTemplate:
        spec:
          containers:
            - name: clickhouse
              resources:
                limits:
                  memory: 4Gi
                requests:
                  cpu: 1
                  memory: 2Gi
          initContainers:
            - name: clickhouse-init
              resources:
                limits:
                  memory: 1Gi
                requests:
                  cpu: 500m
                  memory: 512Mi
      storage:
        storageClassName: "local-path"
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
  deletionPolicy: WipeOut
```

Here,
- `spec.init.archiver.fullDBRepository` refers to the `Repository` created by the `ClickHouseArchiver` (named `<ClickHouse-name>-full` by convention, i.e. `sample-clickhouse-full`).
- `spec.init.archiver.recoveryTimestamp` specifies the point in time to recover to. KubeDB will restore the latest full backup taken before this timestamp together with the latest incremental backup taken before it.
- `spec.init.archiver.encryptionSecret` refers to the same encryption secret used by the `ClickHouseArchiver`.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/clickhouse/pitr/examples/restored-clickhouse-pitr.yaml
clickhouse.kubedb.com/restored-clickhouse-pitr created
```

Let's check the pods which are related to the restore,

```bash
$ kubectl get pods -n demo | grep restored-clickhouse-pitr
restored-clickhouse-pitr-appscode-cluster-shard-0-0               1/1     Running     0          110s
restored-clickhouse-pitr-appscode-cluster-shard-0-1               1/1     Running     0          106s
restored-clickhouse-pitr-appscode-cluster-shard-1-0               1/1     Running     0          107s
restored-clickhouse-pitr-appscode-cluster-shard-1-1               1/1     Running     0          102s
restored-clickhouse-pitr-inc-backup-restorer-2fnpv                0/1     Completed   0          48s
restored-clickhouse-pitr-keeper-0                                 1/1     Running     0          113s
restored-clickhouse-pitr-keeper-1                                 1/1     Running     0          107s
restored-clickhouse-pitr-keeper-2                                 1/1     Running     0          103s
restored-clickhouse-pitr-manifest-restorer-hc6sp                  0/1     Completed   0          2m6s
```

Here,
- Pod `restored-clickhouse-pitr-manifest-restorer-*` is responsible for restoring the database manifest/metadata.
- Pod `restored-clickhouse-pitr-inc-backup-restorer-*` restores the latest full backup taken before `recoveryTimestamp` together with the latest incremental backup taken before `recoveryTimestamp`.

> Note: Restore process works sequentially. Manifest Restore --> Full-backup + Incremental Restore.

Let's watch the underlying `RestoreSession` objects,

```bash
$ kubectl get restoresession -n demo
NAME                                           REPOSITORY               PHASE       DURATION   AGE
restored-clickhouse-pitr-inc-backup-restorer   sample-clickhouse-full   Succeeded   23s        49s
restored-clickhouse-pitr-manifest-restorer     sample-clickhouse-full   Succeeded   2s         2m6s
```

You can also see which backups were picked in the logs of the incremental restorer pod,

```bash
$ kubectl logs -n demo restored-clickhouse-pitr-inc-backup-restorer-2fnpv | grep -E "Starting restore|Selected Backup|completed successfully"
I0930 05:40:56.162546       1 pitr_restorer.go:115] Starting restore for snapshot: sample-clickhouse-full-sample-clarchiver-full-backup-1790746201
I0930 05:40:56.694446       1 helpers.go:495] Selected Backup For Restore: dau9uf9t3d5c73fgt720
I0930 05:41:16.907283       1 pitr_restorer.go:130] Restore completed successfully for snapshot: sample-clickhouse-full-sample-clarchiver-full-backup-1790746201
```

Here, the full backup `sample-clickhouse-full-sample-clarchiver-full-backup-1790746201` (taken at `05:30:01Z`) was restored together with the incremental backup `dau9uf9t3d5c73fgt720`, which is the incremental backup cycle started at `05:33:49Z` — the last one before our recovery timestamp `05:34:00Z`.

#### Verify Restored Data:

At first, check if the database has gone into `Ready` state by the following command,

```bash
$ kubectl get clickhouse -n demo restored-clickhouse-pitr
NAME                       VERSION   STATUS   AGE
restored-clickhouse-pitr   25.7.1    Ready    2m7s
```

Now, let's exec into the pods to verify the restored data. The **original** `sample-clickhouse` database (which we never touched) still has all 23 rows, while the **restored** `restored-clickhouse-pitr` database — restored to `2026-09-30T05:34:00Z` — should only have the 13 rows that existed at that point in time, on the same shards as before.

Since the manifest restorer also restores the auth secret of the original database, the restored database uses the same credentials as `sample-clickhouse`,

```bash
$ kubectl get secret -n demo restored-clickhouse-pitr-auth -o jsonpath='{.data.username}' | base64 -d
admin⏎

$ kubectl get secret -n demo restored-clickhouse-pitr-auth -o jsonpath='{.data.password}' | base64 -d
psL)PHO!)ryZEcoX⏎
```

Let's check the data on every replica of each shard,

```bash
$ kubectl exec -it -n demo restored-clickhouse-pitr-appscode-cluster-shard-0-0 -- clickhouse-client --user admin --password 'psL)PHO!)ryZEcoX'

:) SELECT * FROM demo_db.users ORDER BY id;

   ┌─id─┬─name─────────┬─shard_id─┐
1. │  1 │ Alice_shard0 │        0 │
2. │  2 │ Bob_shard0   │        0 │
3. │  5 │ Eve_shard0   │        0 │
4. │  6 │ Frank_shard0 │        0 │
5. │  9 │ Ivy_shard0   │        0 │
6. │ 10 │ Jack_shard0  │        0 │
7. │ 11 │ Kate_shard0  │        0 │
   └────┴──────────────┴──────────┘

:) exit
```

```bash
$ kubectl exec -it -n demo restored-clickhouse-pitr-appscode-cluster-shard-0-1 -- clickhouse-client --user admin --password 'psL)PHO!)ryZEcoX'

:) SELECT * FROM demo_db.users ORDER BY id;

   ┌─id─┬─name─────────┬─shard_id─┐
1. │  1 │ Alice_shard0 │        0 │
2. │  2 │ Bob_shard0   │        0 │
3. │  5 │ Eve_shard0   │        0 │
4. │  6 │ Frank_shard0 │        0 │
5. │  9 │ Ivy_shard0   │        0 │
6. │ 10 │ Jack_shard0  │        0 │
7. │ 11 │ Kate_shard0  │        0 │
   └────┴──────────────┴──────────┘

:) exit
```

```bash
$ kubectl exec -it -n demo restored-clickhouse-pitr-appscode-cluster-shard-1-0 -- clickhouse-client --user admin --password 'psL)PHO!)ryZEcoX'

:) SELECT * FROM demo_db.users ORDER BY id;

   ┌─id─┬─name───────────┬─shard_id─┐
1. │  3 │ Charlie_shard1 │        1 │
2. │  4 │ David_shard1   │        1 │
3. │  7 │ Grace_shard1   │        1 │
4. │  8 │ Henry_shard1   │        1 │
5. │ 12 │ Leo_shard1     │        1 │
6. │ 13 │ Mia_shard1     │        1 │
   └────┴────────────────┴──────────┘

:) exit
```

```bash
$ kubectl exec -it -n demo restored-clickhouse-pitr-appscode-cluster-shard-1-1 -- clickhouse-client --user admin --password 'psL)PHO!)ryZEcoX'

:) SELECT * FROM demo_db.users ORDER BY id;

   ┌─id─┬─name───────────┬─shard_id─┐
1. │  3 │ Charlie_shard1 │        1 │
2. │  4 │ David_shard1   │        1 │
3. │  7 │ Grace_shard1   │        1 │
4. │  8 │ Henry_shard1   │        1 │
5. │ 12 │ Leo_shard1     │        1 │
6. │ 13 │ Mia_shard1     │        1 │
   └────┴────────────────┴──────────┘

:) exit
```

And through the `Distributed` table,

```bash
$ kubectl exec -it -n demo restored-clickhouse-pitr-appscode-cluster-shard-0-0 -- clickhouse-client --user admin --password 'psL)PHO!)ryZEcoX'

:) SELECT count() FROM demo_db.users_dist;

   ┌─count()─┐
1. │      13 │
   └─────────┘

:) exit
```

Let's also verify that both table definitions were restored,

```bash
$ kubectl exec -it -n demo restored-clickhouse-pitr-appscode-cluster-shard-0-0 -- clickhouse-client --user admin --password 'psL)PHO!)ryZEcoX'

:) SHOW CREATE TABLE demo_db.users;

CREATE TABLE demo_db.users
(
    `id` UInt64,
    `name` String,
    `shard_id` UInt8
)
ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}/demo_db/users', '{replica}')
ORDER BY id
SETTINGS index_granularity = 8192

:) SHOW CREATE TABLE demo_db.users_dist;

CREATE TABLE demo_db.users_dist
(
    `id` UInt64,
    `name` String,
    `shard_id` UInt8
)
ENGINE = Distributed('{cluster}', 'demo_db', 'users', rand())

:) exit
```

As shown above, `restored-clickhouse-pitr` has exactly the 13 rows that existed at `2026-09-30T05:34:00Z` — ids `1, 2, 5, 6, 9, 10, 11` on both replicas of shard `0` and ids `3, 4, 7, 8, 12, 13` on both replicas of shard `1`. The rows of the fourth and fifth batches (ids `14`-`23`), which were inserted after the recovery timestamp, are correctly **not** present, while the original `sample-clickhouse` database still has all 23 rows. This confirms that `spec.init.archiver.recoveryTimestamp` recovers both the schema (`ReplicatedMergeTree` + `Distributed` tables) and the exact per-shard data to the specified historical point, rather than the latest available state.

### Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete clickhouse -n demo restored-clickhouse-pitr
kubectl delete clickhouse -n demo sample-clickhouse
kubectl delete clickhousearchiver -n demo sample-clickhouse-archiver
kubectl delete retentionpolicies.storage.kubestash.com -n demo demo-retention
kubectl delete backupstorage -n demo s3-storage
kubectl delete secret -n demo s3-secret encrypt-secret
```
