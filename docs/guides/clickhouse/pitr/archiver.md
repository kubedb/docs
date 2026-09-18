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

> The `KubeDB` provisioner uses `KubeStash` for the periodic full backup and a `sidekick` container for continuous incremental archiving of `ClickHouse` databases, both using the `ClickHouseBackup` driver ([clickhouse-backup](https://github.com/Altinity/clickhouse-backup)). Unlike some other KubeDB addons, these jobs run fine with the default (non-root) security context, so no `securityContext` override is required.

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

Let's check the pods which are related to the backup,

```bash
$ kubectl get pods -n demo
NAME                                             READY   STATUS      RESTARTS   AGE
sample-clickhouse-appscode-cluster-shard-0-0    1/1     Running     0          4m12s
sample-clickhouse-appscode-cluster-shard-0-1    1/1     Running     0          4m7s
sample-clickhouse-appscode-cluster-shard-1-0    1/1     Running     0          4m9s
sample-clickhouse-appscode-cluster-shard-1-1    1/1     Running     0          4m4s
sample-clickhouse-archiver-full-backup-1789714189-kps72   0/1   Completed   0   2m59s
sample-clickhouse-keeper-0                       1/1     Running     0          4m14s
sample-clickhouse-keeper-1                       1/1     Running     0          4m8s
sample-clickhouse-keeper-2                       1/1     Running     0          4m3s
sample-clickhouse-sidekick                       1/1     Running     0          44s
```

Here,
- Pod `sample-clickhouse-archiver-full-backup-1789714189-kps72` performed the initial full backup.
- Pod `sample-clickhouse-sidekick` performs continuous incremental archiving. This pod always stays in `Running` phase; it periodically (roughly every minute) archives the changes made since the last full or incremental backup.

### Verify Backup Setup Successful

If everything goes well, the KubeDB provisioner will create a `BackupConfiguration`, and its phase should be `Ready`. The `Ready` phase indicates that the backup setup is successful.

Let's verify the phase of the `BackupConfiguration`,

```bash
$ kubectl get backupconfiguration -n demo
NAME                          PHASE   PAUSED   AGE
sample-clickhouse-archiver    Ready            48s
```

**Verify BackupSession:**

KubeStash triggers an instant full backup as soon as the `BackupConfiguration` is ready. After that, full backups are scheduled according to `spec.fullBackup.scheduler.schedule`, while incremental backups run continuously in between.

```bash
$ kubectl get backupsession -n demo
NAME                                                 INVOKER-TYPE          INVOKER-NAME                  PHASE       DURATION   AGE
sample-clickhouse-archiver-full-backup-1789714189   BackupConfiguration   sample-clickhouse-archiver    Succeeded   33s        53s
```

**Verify Snapshot:**

```bash
$ kubectl get snapshots -n demo -l=kubestash.com/repo-name=sample-clickhouse-full
NAME                                                              REPOSITORY               SESSION       SNAPSHOT-TIME          DELETION-POLICY   PHASE       AGE
sample-clickhouse-full-sample-clarchiver-full-backup-1789714189   sample-clickhouse-full   full-backup   2026-09-18T06:50:00Z   Delete            Succeeded   3m
```

### Insert Some Data

Every successful incremental archiving cycle records the changes made to the database since the last snapshot. To make it obvious that a point-in-time restore really lands on an *older* state (and not just the latest data), we insert data in two batches, separated by an incremental archiving cycle, and only restore up to the first batch.

```bash
$ kubectl get secret -n demo sample-clickhouse-auth -o jsonpath='{.data.username}' | base64 -d
admin⏎

$ kubectl get secret -n demo sample-clickhouse-auth -o jsonpath='{.data.password}' | base64 -d
UVP4L2n_HkUItMOq⏎

$ kubectl exec -it -n demo sample-clickhouse-appscode-cluster-shard-0-0 -- clickhouse-client --user admin --password "UVP4L2n_HkUItMOq"

:) CREATE DATABASE playground;
:) CREATE TABLE playground.equipment (id UInt32, type String, quant UInt32, color String) ENGINE = MergeTree() ORDER BY id;
:) INSERT INTO playground.equipment VALUES (1,'Swing',10,'Red'),(2,'Slide',5,'Blue'),(3,'Monkey Bars',3,'Yellow');

:) SELECT * FROM playground.equipment ORDER BY id;

┌─id─┬─type────────┬─quant─┬─color──┐
│  1 │ Swing       │    10 │ Red    │
│  2 │ Slide       │     5 │ Blue   │
│  3 │ Monkey Bars │     3 │ Yellow │
└────┴─────────────┴───────┴────────┘

:) exit
```

We wait for the `sidekick` pod to pick up this change in its next incremental archiving cycle,

```bash
$ kubectl logs -n demo sample-clickhouse-sidekick --tail=8
...
I0918 06:54:19.019363       1 archiver.go:187] Cluster incremental backup completed successfully for all 2 shards
...
I0918 06:54:19.049946       1 incremental_backup.go:146] Incremental backup cycle completed in 20.322968553s
```

This incremental backup (completed at `06:54:19`) now holds the 3-row state. **This is the point in time we are going to restore to.** Some time later, we insert two more rows into the *same, still-running* database — these rows must **not** appear in our restore:

```bash
$ kubectl exec -it -n demo sample-clickhouse-appscode-cluster-shard-0-0 -- clickhouse-client --user admin --password "UVP4L2n_HkUItMOq"

:) INSERT INTO playground.equipment VALUES (4,'Seesaw',4,'Green'),(5,'Trampoline',2,'Orange');
:) SELECT * FROM playground.equipment ORDER BY id;

┌─id─┬─type────────┬─quant─┬─color──────┐
│  1 │ Swing       │    10 │ Red        │
│  2 │ Slide       │     5 │ Blue       │
│  3 │ Monkey Bars │     3 │ Yellow     │
│  4 │ Seesaw      │     4 │ Green      │
│  5 │ Trampoline  │     2 │ Orange     │
└────┴─────────────┴───────┴────────────┘

:) exit
```

```bash
$ date -u +"%Y-%m-%dT%H:%M:%SZ"
2026-09-18T06:54:31Z
```

The `sidekick` pod archives this second change too, in the *next* incremental cycle (completed at `06:55:18`):

```bash
$ kubectl logs -n demo sample-clickhouse-sidekick --tail=8
...
I0918 06:55:18.939879       1 archiver.go:187] Cluster incremental backup completed successfully for all 2 shards
...
I0918 06:55:18.974402       1 incremental_backup.go:146] Incremental backup cycle completed in 20.246758296s
```

Note that we are **not** touching or dropping anything in the original `sample-clickhouse` database — it keeps running with all 5 rows. This lets us prove, side-by-side, that the restored database ends up with the *older* 3-row state instead of silently picking up the latest data.

## Point-in-time Recovery

Point-In-Time Recovery allows you to restore a `ClickHouse` database to a specific point in time using the continuously archived incremental backups. This is particularly useful in scenarios where you need to recover to a state just before a specific error or unwanted change occurred — without necessarily touching the original database.

We pick a recovery timestamp that falls **after** the first incremental backup (`06:54:19`, holding the 3-row state) but **before** the second one (`06:55:18`, holding the 5-row state). For this demo, we use `2026-09-18T06:54:40Z`.

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
      recoveryTimestamp: "2026-09-18T06:54:40Z"
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
- `spec.init.archiver.recoveryTimestamp` specifies the exact point in time to recover to. KubeDB will restore the latest full backup taken before this timestamp and then replay the incremental backups up to it.
- `spec.init.archiver.encryptionSecret` refers to the same encryption secret used by the `ClickHouseArchiver`.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/clickhouse/pitr/examples/restored-clickhouse-pitr.yaml
clickhouse.kubedb.com/restored-clickhouse-pitr created
```

Let's check the pods which are related to the restore,

```bash
$ kubectl get pods -n demo
NAME                                                  READY   STATUS      RESTARTS   AGE
restored-clickhouse-pitr-appscode-cluster-shard-0-0   1/1     Running     0          3m10s
restored-clickhouse-pitr-appscode-cluster-shard-0-1   1/1     Running     0          3m5s
restored-clickhouse-pitr-appscode-cluster-shard-1-0   1/1     Running     0          3m7s
restored-clickhouse-pitr-appscode-cluster-shard-1-1   1/1     Running     0          3m2s
restored-clickhouse-pitr-inc-backup-restorer-xxxxx    0/1     Completed   0          1m50s
restored-clickhouse-pitr-keeper-0                     1/1     Running     0          3m12s
restored-clickhouse-pitr-keeper-1                     1/1     Running     0          3m6s
restored-clickhouse-pitr-keeper-2                     1/1     Running     0          3m1s
restored-clickhouse-pitr-manifest-restorer-xxxxx      0/1     Completed   0          3m23s
```

Here,
- Pod `restored-clickhouse-pitr-manifest-restorer-xxxxx` is responsible for restoring the database manifest/metadata.
- Pod `restored-clickhouse-pitr-inc-backup-restorer-xxxxx` restores the latest full backup and then replays the incremental backups up to (but not beyond) the requested `recoveryTimestamp`.

> Note: Restore process works sequentially. Manifest Restore --> Full-backup + Incremental Restore.

You can also watch the underlying `RestoreSession` objects,

```bash
$ kubectl get restoresession -n demo
NAME                                            REPOSITORY               PHASE       DURATION   AGE
restored-clickhouse-pitr-inc-backup-restorer   sample-clickhouse-full   Succeeded   23s        110s
restored-clickhouse-pitr-manifest-restorer     sample-clickhouse-full   Succeeded   3s         3m23s
```

#### Verify Restored Data:

At first, check if the database has gone into `Ready` state by the following command,

```bash
$ kubectl get clickhouse -n demo restored-clickhouse-pitr
NAME                       VERSION   STATUS   AGE
restored-clickhouse-pitr   25.7.1    Ready    4m2s
```

Now, let's exec into the pod to verify the restored data. This is the important check: the **original** `sample-clickhouse` database (which we never touched) should still show all 5 rows, while the **restored** `restored-clickhouse-pitr` database — restored to `2026-09-18T06:54:40Z` — should only show the 3 rows that existed at that point in time.

```bash
$ kubectl exec -it -n demo sample-clickhouse-appscode-cluster-shard-0-0 -- clickhouse-client --user admin --password "UVP4L2n_HkUItMOq"

:) SELECT * FROM playground.equipment ORDER BY id;

┌─id─┬─type────────┬─quant─┬─color──────┐
│  1 │ Swing       │    10 │ Red        │
│  2 │ Slide       │     5 │ Blue       │
│  3 │ Monkey Bars │     3 │ Yellow     │
│  4 │ Seesaw      │     4 │ Green      │
│  5 │ Trampoline  │     2 │ Orange     │
└────┴─────────────┴───────┴────────────┘

:) exit
```

```bash
$ kubectl get secret -n demo restored-clickhouse-pitr-auth -o jsonpath='{.data.username}' | base64 -d
admin⏎

$ kubectl get secret -n demo restored-clickhouse-pitr-auth -o jsonpath='{.data.password}' | base64 -d
K7bQmZ2xT9nRpLsW⏎

$ kubectl exec -it -n demo restored-clickhouse-pitr-appscode-cluster-shard-0-0 -- clickhouse-client --user admin --password "K7bQmZ2xT9nRpLsW"

:) SHOW DATABASES;

┌─name───────────────┐
│ INFORMATION_SCHEMA  │
│ default             │
│ information_schema  │
│ playground          │
│ system              │
└────────────────────┘

:) SELECT * FROM playground.equipment ORDER BY id;

┌─id─┬─type────────┬─quant─┬─color──────┐
│  1 │ Swing       │    10 │ Red        │
│  2 │ Slide       │     5 │ Blue       │
│  3 │ Monkey Bars │     3 │ Yellow     │
└────┴─────────────┴───────┴────────────┘

:) exit
```

As shown above, `restored-clickhouse-pitr` only has the 3 rows that existed at `2026-09-18T06:54:40Z` — rows `4` (`Seesaw`) and `5` (`Trampoline`), which were inserted afterward at `06:54:31` and archived in the *second* incremental backup, are correctly **not** present. Meanwhile, the original `sample-clickhouse` database — which we never modified or restored — still has all 5 rows. This confirms that `spec.init.archiver.recoveryTimestamp` genuinely recovers to the specified historical point rather than simply reconstructing the latest available state.

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
