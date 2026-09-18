---
title: Backup & Restore ClickHouse | KubeStash
description: Backup and Restore ClickHouse database using KubeStash
menu:
  docs_{{ .version }}:
    identifier: guides-clickhouse-logical-backup
    name: Logical Backup
    parent: ch-backup
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Backup and Restore ClickHouse database using KubeStash

KubeStash allows you to backup and restore `ClickHouse` databases. It supports backups for `ClickHouse` instances running in both Standalone and `clusterTopology` (sharded, replicated clusters with `ClickHouseKeeper`) configurations. KubeStash makes managing your `ClickHouse` backups and restorations more straightforward and efficient.

This guide will give you an overview how you can take backup and restore your `ClickHouse` databases using `KubeStash`. Here, we are going to demonstrate the backup and restore process for a `ClickHouse` database using `clusterTopology`. The process is similar for Standalone configuration.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using `Minikube` or `Kind`.
- Install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md).
- Install `KubeStash` in your cluster following the steps [here](https://kubestash.com/docs/latest/setup/install/kubestash).
- Install KubeStash `kubectl` plugin following the steps [here](https://kubestash.com/docs/latest/setup/install/kubectl-plugin/).
- If you are not familiar with how KubeStash backup and restore ClickHouse databases, please check the following guide [here](/docs/guides/clickhouse/backup/overview/index.md).

You should be familiar with the following `KubeStash` concepts:

- [BackupStorage](https://kubestash.com/docs/latest/concepts/crds/backupstorage/)
- [BackupConfiguration](https://kubestash.com/docs/latest/concepts/crds/backupconfiguration/)
- [BackupSession](https://kubestash.com/docs/latest/concepts/crds/backupsession/)
- [RestoreSession](https://kubestash.com/docs/latest/concepts/crds/restoresession/)
- [Addon](https://kubestash.com/docs/latest/concepts/crds/addon/)
- [Function](https://kubestash.com/docs/latest/concepts/crds/function/)
- [Task](https://kubestash.com/docs/latest/concepts/crds/addon/#task-specification)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/guides/clickhouse/backup/logical/examples](/docs/guides/clickhouse/backup/logical/examples) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Backup ClickHouse

This section will demonstrate how to backup a `ClickHouse` database. Here, we are going to deploy a `ClickHouse` database using KubeDB. Then, we are going to backup this database into an `S3` bucket. Finally, we are going to restore the backed up data into another `ClickHouse` database.

### Deploy Sample ClickHouse Database

Below is the YAML of a sample `ClickHouse` CR that we are going to create for this tutorial:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ClickHouse
metadata:
  name: sample-clickhouse
  namespace: demo
spec:
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

Create the above `ClickHouse` CR,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/clickhouse/backup/logical/examples/sample-clickhouse.yaml
clickhouse.kubedb.com/sample-clickhouse created
```

KubeDB will deploy a `ClickHouse` cluster (2 shards x 2 replicas, plus a 3-node `ClickHouseKeeper` ensemble) according to the above specification. It will also create the necessary `Secrets` and `Services` to access the database.

Let's check if the database is ready to use,

```bash
$ kubectl get clickhouse -n demo sample-clickhouse
NAME                VERSION   STATUS   AGE
sample-clickhouse   25.7.1    Ready    3m27s
```

The database is `Ready`. Verify that KubeDB has created a `Secret` and `Services` for this database using the following commands,

```bash
$ kubectl get secret -n demo | grep sample-clickhouse
sample-clickhouse-auth                  kubernetes.io/basic-auth   2      81s
sample-clickhouse-fd9557                Opaque                     3      80s
sample-clickhouse-internal-auth-token   kubernetes.io/basic-auth   1      80s
sample-clickhouse-keeper-config         Opaque                     2      80s

$ kubectl get service -n demo -l=app.kubernetes.io/instance=sample-clickhouse
NAME                            TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)             AGE
sample-clickhouse               ClusterIP   10.43.52.87     <none>        9000/TCP,8123/TCP   81s
sample-clickhouse-keeper        ClusterIP   10.43.166.188   <none>        9181/TCP            80s
sample-clickhouse-keeper-pods   ClusterIP   None            <none>        9234/TCP            80s
sample-clickhouse-pods          ClusterIP   None            <none>        9000/TCP,8123/TCP   81s
```

Here, we have to use service `sample-clickhouse` and secret `sample-clickhouse-auth` to connect with the database. `KubeDB` creates an `AppBinding` CR that holds the necessary information to connect with the database.

**Verify AppBinding:**

Verify that the `AppBinding` has been created successfully using the following command,

```bash
$ kubectl get appbindings -n demo
NAME                TYPE                    VERSION   AGE
sample-clickhouse   kubedb.com/clickhouse   25.7.1    73s
```

Let's check the YAML of the above `AppBinding`,

```bash
$ kubectl get appbindings -n demo sample-clickhouse -o yaml
```

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: sample-clickhouse
  namespace: demo
spec:
  appRef:
    apiGroup: kubedb.com
    kind: ClickHouse
    name: sample-clickhouse
    namespace: demo
  clientConfig:
    service:
      name: sample-clickhouse
      port: 9000
      scheme: http
  secret:
    apiGroup: ""
    kind: Secret
    name: sample-clickhouse-auth
  type: kubedb.com/clickhouse
  version: 25.7.1
```

KubeStash uses the `AppBinding` CR to connect with the target database. It requires the following two fields to be set in AppBinding's `.spec` section.

Here,

- `.spec.clientConfig.service.name` specifies the name of the Service that connects to the database.
- `.spec.secret` specifies the name of the Secret that holds necessary credentials to access the database.
- `.spec.type` specifies the type of the app that this AppBinding is pointing to. KubeDB generated AppBinding follows the following format: `<app group>/<app resource type>`.

**Insert Sample Data:**

Now, we are going to exec into one of the database pod and create some sample data. At first, find out the database `Pod`s using the following command,

```bash
$ kubectl get pods -n demo --selector="app.kubernetes.io/instance=sample-clickhouse"
NAME                                            READY   STATUS    RESTARTS   AGE
sample-clickhouse-appscode-cluster-shard-0-0   1/1     Running   0          119s
sample-clickhouse-appscode-cluster-shard-0-1   1/1     Running   0          22s
sample-clickhouse-appscode-cluster-shard-1-0   1/1     Running   0          116s
sample-clickhouse-appscode-cluster-shard-1-1   1/1     Running   0          22s
```

And copy the username and password of the admin user to access the `clickhouse-client` shell.

```bash
$ kubectl get secret -n demo sample-clickhouse-auth -o jsonpath='{.data.username}' | base64 -d
admin⏎

$ kubectl get secret -n demo sample-clickhouse-auth -o jsonpath='{.data.password}' | base64 -d
fB9sH0(xeg3FBxs7⏎
```

Since `sample-clickhouse` is deployed with a `clusterTopology` (2 shards x 2 replicas), a plain `MergeTree` table would only live on a single node and would **not** be replicated or sharded. To properly use the cluster, we create a `ReplicatedMergeTree` table (for replication within a shard, coordinated through `ClickHouseKeeper`) on every node using `ON CLUSTER`, and a `Distributed` table on top of it (for transparently routing reads/writes across all shards).

Now, let's exec into a `Pod` and create the database and tables,

```bash
$ kubectl exec -it -n demo sample-clickhouse-appscode-cluster-shard-0-0 -- clickhouse-client --user admin --password "fB9sH0(xeg3FBxs7"

# create a database named "playground" on every node of the cluster
:) CREATE DATABASE playground ON CLUSTER 'appscode-cluster';

# create the underlying replicated table on every shard/replica
:) CREATE TABLE playground.equipment_local ON CLUSTER 'appscode-cluster'
   (
       id UInt32,
       type String,
       quant UInt32,
       color String
   )
   ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}/equipment_local', '{replica}')
   ORDER BY id;

# create a Distributed table on top, so we can read/write across all shards through a single table
:) CREATE TABLE playground.equipment ON CLUSTER 'appscode-cluster'
   AS playground.equipment_local
   ENGINE = Distributed('appscode-cluster', 'playground', 'equipment_local', rand());

# insert some rows through the Distributed table
:) INSERT INTO playground.equipment VALUES (1,'Swing',10,'Red'),(2,'Slide',5,'Blue'),(3,'Monkey Bars',3,'Yellow');

# verify that data has been inserted successfully
:) SELECT * FROM playground.equipment ORDER BY id;

┌─id─┬─type────────┬─quant─┬─color──┐
│  1 │ Swing       │    10 │ Red    │
│  2 │ Slide       │     5 │ Blue   │
│  3 │ Monkey Bars │     3 │ Yellow │
└────┴─────────────┴───────┴────────┘

:) exit
```

Here,

- `{shard}` and `{replica}` are macros that KubeDB automatically configures on every `ClickHouse` pod (visible via `SELECT * FROM system.macros`), so the same `CREATE TABLE ... ON CLUSTER` statement creates a correctly-parameterized replica path on each node.
- The `equipment_local` table on the two replicas of a shard (e.g. `shard-0-0` and `shard-0-1`) stays in sync via `ClickHouseKeeper`, while `rand()` in the `Distributed` engine definition spreads rows for the `equipment` table across the two shards.

We can verify this by checking the local table on each shard directly. In this run, all 3 rows happened to land on shard `0` (and were replicated to both of its replicas), while shard `1` has none — this is expected: with a handful of rows, ClickHouse doesn't guarantee an even split across shards.

```bash
$ kubectl exec -it -n demo sample-clickhouse-appscode-cluster-shard-0-0 -- clickhouse-client --user admin --password "fB9sH0(xeg3FBxs7" -q "SELECT * FROM playground.equipment_local ORDER BY id"
1	Swing	10	Red
2	Slide	5	Blue
3	Monkey Bars	3	Yellow

$ kubectl exec -it -n demo sample-clickhouse-appscode-cluster-shard-0-1 -- clickhouse-client --user admin --password "fB9sH0(xeg3FBxs7" -q "SELECT * FROM playground.equipment_local ORDER BY id"
1	Swing	10	Red
2	Slide	5	Blue
3	Monkey Bars	3	Yellow

$ kubectl exec -it -n demo sample-clickhouse-appscode-cluster-shard-1-0 -- clickhouse-client --user admin --password "fB9sH0(xeg3FBxs7" -q "SELECT * FROM playground.equipment_local ORDER BY id"
```

Let's insert a few more rows through the `Distributed` table to see the sharding actually spread the data out,

```bash
$ kubectl exec -n demo sample-clickhouse-appscode-cluster-shard-0-0 -- clickhouse-client --user admin --password "fB9sH0(xeg3FBxs7" -q \
  "INSERT INTO playground.equipment VALUES (4,'Item4',4,'Color4'),(5,'Item5',5,'Color5'),(6,'Item6',6,'Color6'),(7,'Item7',7,'Color7'),(8,'Item8',8,'Color8'),(9,'Item9',9,'Color9'),(10,'Item10',10,'Color10'),(11,'Item11',11,'Color11'),(12,'Item12',12,'Color12'),(13,'Item13',13,'Color13'),(14,'Item14',14,'Color14'),(15,'Item15',15,'Color15')"

$ kubectl exec -n demo sample-clickhouse-appscode-cluster-shard-0-0 -- clickhouse-client --user admin --password "fB9sH0(xeg3FBxs7" -q "SELECT count() FROM playground.equipment_local"
12

$ kubectl exec -n demo sample-clickhouse-appscode-cluster-shard-1-0 -- clickhouse-client --user admin --password "fB9sH0(xeg3FBxs7" -q "SELECT count() FROM playground.equipment_local"
3

$ kubectl exec -n demo sample-clickhouse-appscode-cluster-shard-0-0 -- clickhouse-client --user admin --password "fB9sH0(xeg3FBxs7" -q "SELECT count() FROM playground.equipment"
15
```

Now the data is spread across both shards (12 rows on shard `0`, 3 rows on shard `1`), while the `Distributed` table transparently reports all 15 rows regardless of which shard a client happens to connect to.

Now, we are ready to backup the database.

### Prepare Backend

We are going to store our backed up data into an `S3` bucket. We have to create a `Secret` with necessary credentials and a `BackupStorage` CR to use this backend. If you want to use a different backend, please read the respective backend configuration doc from [here](https://kubestash.com/docs/latest/guides/backends/overview/).

> **Note:** ClickHouse currently supports `S3`, `Azure Blob Storage`, and `Google Cloud Storage` (via S3 compatibility mode) as backup storage backends.

**Create Secret:**

Let's create a secret called `s3-secret` with access credentials to our desired S3 (or S3 compatible, e.g. Minio) bucket,

```bash
$ echo -n '<your-aws-access-key-id-here>' > AWS_ACCESS_KEY_ID
$ echo -n '<your-aws-secret-access-key-here>' > AWS_SECRET_ACCESS_KEY
$ kubectl create secret generic -n demo s3-secret \
    --from-file=./AWS_ACCESS_KEY_ID \
    --from-file=./AWS_SECRET_ACCESS_KEY
secret/s3-secret created
```

**Create BackupStorage:**

Now, create a `BackupStorage` using this secret. Below is the YAML of the `BackupStorage` CR we are going to create,

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
      prefix: clickhouse-backup
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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/clickhouse/backup/logical/examples/backupstorage.yaml
backupstorage.storage.kubestash.com/s3-storage created
```

Now, we are ready to backup our database to our desired backend.

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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/clickhouse/backup/logical/examples/retentionpolicy.yaml
retentionpolicy.storage.kubestash.com/demo-retention created
```

### Backup

We have to create a `BackupConfiguration` targeting our `sample-clickhouse` ClickHouse database. Then, KubeStash will create a `CronJob` for each session to take periodic backup of that database.

Below is the YAML for the `BackupConfiguration` CR to backup the `sample-clickhouse` ClickHouse database that we have deployed earlier,

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: sample-clickhouse-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: ClickHouse
    namespace: demo
    name: sample-clickhouse
  backends:
    - name: s3-backend
      storageRef:
        namespace: demo
        name: s3-storage
      retentionPolicy:
        name: demo-retention
        namespace: demo
  sessions:
    - name: frequent-backup
      scheduler:
        schedule: "*/5 * * * *"
        jobTemplate:
          backoffLimit: 1
      repositories:
        - name: s3-clickhouse-repo
          backend: s3-backend
          directory: /clickhouse
      addon:
        name: clickhouse-addon
        tasks:
          - name: logical-backup
```

- `.spec.target` refers to the targeted `sample-clickhouse` ClickHouse database that we created earlier.
- `.spec.sessions[*].scheduler.schedule` specifies that we want to backup the database at `5 minutes` interval.
- `.spec.sessions[*].addon` refers to the `clickhouse-addon` and the `logical-backup` task, which uses the `ClickHouseBackup` driver ([clickhouse-backup](https://github.com/Altinity/clickhouse-backup)) to take a logical backup of every shard.

> **Note:** Unlike some other KubeDB addons, the `clickhouse-addon` backup and restore jobs run fine with the default (non-root) security context. You don't need to set `spec.sessions[*].addon.jobTemplate.spec.securityContext` (e.g. `runAsUser`/`runAsGroup`/`fsGroup`) for `ClickHouse` backup or restore.

Let's create the `BackupConfiguration` CR that we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/clickhouse/backup/logical/examples/backupconfiguration.yaml
backupconfiguration.core.kubestash.com/sample-clickhouse-backup created
```

**Verify Backup Setup Successful**

If everything goes well, the phase of the `BackupConfiguration` should be `Ready`. The `Ready` phase indicates that the backup setup is successful. Let's verify the `Phase` of the BackupConfiguration,

```bash
$ kubectl get backupconfiguration -n demo
NAME                       PHASE   PAUSED   AGE
sample-clickhouse-backup   Ready            11s
```

Additionally, we can verify that the `Repository` specified in the `BackupConfiguration` has been created using the following command,

```bash
$ kubectl get repo -n demo
NAME                 INTEGRITY   SNAPSHOT-COUNT   SIZE   PHASE   LAST-SUCCESSFUL-BACKUP   AGE
s3-clickhouse-repo               0                0 B    Ready                            16s
```

**Verify CronJob:**

It will also create a `CronJob` with the schedule specified in `spec.sessions[*].scheduler.schedule` field of the `BackupConfiguration` CR.

Verify that the `CronJob` has been created using the following command,

```bash
$ kubectl get cronjob -n demo
NAME                                               SCHEDULE      SUSPEND   ACTIVE   LAST SCHEDULE   AGE
trigger-sample-clickhouse-backup-frequent-backup   */5 * * * *   False     0        <none>          15s
```

**Verify BackupSession:**

KubeStash triggers an instant backup as soon as the `BackupConfiguration` is ready. After that, backups are scheduled according to the specified schedule.

```bash
$ kubectl get backupsession -n demo
NAME                                                 INVOKER-TYPE          INVOKER-NAME               PHASE       DURATION   AGE
sample-clickhouse-backup-frequent-backup-1789709796   BackupConfiguration   sample-clickhouse-backup   Succeeded   26s        30s
```

We can see from the above output that the backup session has succeeded. Now, we are going to verify whether the backed up data has been stored in the backend.

**Verify Backup:**

Once a backup is complete, KubeStash will update the respective `Repository` CR to reflect the backup. Check that the repository `s3-clickhouse-repo` has been updated by the following command,

```bash
$ kubectl get repository -n demo s3-clickhouse-repo
NAME                 INTEGRITY   SNAPSHOT-COUNT   SIZE   PHASE   LAST-SUCCESSFUL-BACKUP   AGE
s3-clickhouse-repo               1                0 B    Ready   2m27s                    2m28s
```

At this moment we have one `Snapshot`. Run the following command to check the respective `Snapshot` which represents the state of a backup run for an application.

```bash
$ kubectl get snapshots -n demo -l=kubestash.com/repo-name=s3-clickhouse-repo
NAME                                                              REPOSITORY           SESSION           SNAPSHOT-TIME          DELETION-POLICY   PHASE       AGE
s3-clickhouse-repo-sample-clickhckup-frequent-backup-1789709796   s3-clickhouse-repo   frequent-backup   2026-09-18T05:36:37Z   Delete            Succeeded   2m28s
```

> Note: KubeStash creates a `Snapshot` with the following labels:
> - `kubestash.com/app-ref-kind: <target-kind>`
> - `kubestash.com/app-ref-name: <target-name>`
> - `kubestash.com/app-ref-namespace: <target-namespace>`
> - `kubestash.com/repo-name: <repository-name>`
>
> These labels can be used to watch only the `Snapshot`s related to our target Database or `Repository`.

If we check the YAML of the `Snapshot`, we can find the information about the backed up components of the database. For a `ClickHouse` cluster, KubeStash records the backup result for the shared metadata as well as for each shard,

```bash
$ kubectl get snapshots -n demo s3-clickhouse-repo-sample-clickhckup-frequent-backup-1789709796 -oyaml
```

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: Snapshot
metadata:
  annotations:
    kubedb.com/db-version: 25.7.1
  labels:
    kubestash.com/app-ref-kind: ClickHouse
    kubestash.com/app-ref-name: sample-clickhouse
    kubestash.com/app-ref-namespace: demo
    kubestash.com/repo-name: s3-clickhouse-repo
  name: s3-clickhouse-repo-sample-clickhckup-frequent-backup-1789709796
  namespace: demo
spec:
  appRef:
    apiGroup: kubedb.com
    kind: ClickHouse
    name: sample-clickhouse
    namespace: demo
  backupSession: sample-clickhouse-backup-frequent-backup-1789709796
  deletionPolicy: Delete
  repository: s3-clickhouse-repo
  session: frequent-backup
  snapshotID: 01M2SG8J9WZV2R65W251ATEGPN
  type: FullBackup
  version: v1
status:
  components:
    dump:
      clickHouseStats:
      - finishTime: "2026-09-18T05:36:40Z"
        host: dump-metadata
        id: 94c88a90-e5f5-46bf-bf21-b389d5b8d626
        startTime: "2026-09-18T05:36:39Z"
        status: SUCCESS
      - finishTime: "2026-09-18T05:36:50Z"
        host: dump-shard-0
        id: d0100062-d9d6-4ef1-bbab-3887137b6733
        startTime: "2026-09-18T05:36:49Z"
        status: SUCCESS
      - finishTime: "2026-09-18T05:36:50Z"
        host: dump-shard-1
        id: fa091258-5c37-4de8-8181-e73b94ceb85a
        startTime: "2026-09-18T05:36:49Z"
        status: SUCCESS
      driver: ClickHouseBackup
      path: repository/v1/frequent-backup/dump/full/1789709796
      phase: Succeeded
  conditions:
  - lastTransitionTime: "2026-09-18T05:36:37Z"
    message: Recent snapshot list updated successfully
    reason: SuccessfullyUpdatedRecentSnapshotList
    status: "True"
    type: RecentSnapshotListUpdated
  - lastTransitionTime: "2026-09-18T05:36:59Z"
    message: Metadata uploaded to backend successfully
    reason: SuccessfullyUploadedSnapshotMetadata
    status: "True"
    type: SnapshotMetadataUploaded
  phase: Succeeded
  snapshotTime: "2026-09-18T05:36:37Z"
  totalComponents: 1
  verificationStatus: NotVerified
```

Here, `status.components.dump.clickHouseStats` shows that a dump was taken for the cluster metadata (`dump-metadata`) as well as for each shard (`dump-shard-0`, `dump-shard-1`).

Now, if we navigate to the S3 bucket, we will see the backed up data stored under the `clickhouse-backup/clickhouse/repository/v1/frequent-backup/dump` directory. KubeStash also keeps the backup for `Repository` and `Snapshot` YAMLs, which can be found in the `clickhouse-backup/clickhouse/repository` and `clickhouse-backup/clickhouse/snapshots` directories respectively.

## Restore

In this section, we are going to restore the database from the backup we have taken in the previous section. We are going to deploy a new database and initialize it from the backup.

Below is the YAML for the `ClickHouse` CR we are going to deploy to restore the backed up data into,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ClickHouse
metadata:
  name: restored-clickhouse
  namespace: demo
spec:
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

Let's create the above database,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/clickhouse/backup/logical/examples/restored-clickhouse.yaml
clickhouse.kubedb.com/restored-clickhouse created
```

Wait until the database goes into `Ready` state,

```bash
$ kubectl get clickhouse -n demo restored-clickhouse
NAME                  VERSION   STATUS   AGE
restored-clickhouse   25.7.1    Ready    3m
```

#### Create RestoreSession:

Now, we need to create a `RestoreSession` CR pointing to the targeted `ClickHouse` database.

Below is the content of the YAML file of the `RestoreSession` object that we are going to create to restore backed up data into the newly created `ClickHouse` database named `restored-clickhouse`.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: sample-clickhouse-restore
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: ClickHouse
    namespace: demo
    name: restored-clickhouse
  dataSource:
    repository: s3-clickhouse-repo
    snapshot: latest
  addon:
    name: clickhouse-addon
    tasks:
      - name: logical-backup-restore
```

Here,

- `.spec.target` refers to the newly created `restored-clickhouse` ClickHouse object to where we want to restore backup data.
- `.spec.dataSource.repository` specifies the Repository object that holds the backed up data.
- `.spec.dataSource.snapshot` specifies to restore from the `latest` Snapshot.
- `.spec.addon` refers to the `clickhouse-addon` and the `logical-backup-restore` task.

> As noted earlier, the `clickhouse-addon` restore job also runs fine with the default security context, so `spec.addon.jobTemplate.spec.securityContext` is not required here either.

Let's create the RestoreSession CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/clickhouse/backup/logical/examples/restoresession.yaml
restoresession.core.kubestash.com/sample-clickhouse-restore created
```

Once you have created the `RestoreSession` object, KubeStash will create a restore Job. Run the following command to watch the phase of the `RestoreSession` object,

```bash
$ watch kubectl get restoresession -n demo
NAME                        REPOSITORY            PHASE       DURATION   AGE
sample-clickhouse-restore   s3-clickhouse-repo    Succeeded   22s        30s
```

The `Succeeded` phase means that the restore process has been completed successfully.

#### Verify Restored Data:

In this section, we are going to verify whether the desired data has been restored successfully. We are going to connect to the database server and check whether the database and the table we created earlier in the original database are restored.

At first, check if the database has gone into `Ready` state by the following command,

```bash
$ kubectl get clickhouse -n demo restored-clickhouse
NAME                  VERSION   STATUS   AGE
restored-clickhouse   25.7.1    Ready    6m
```

Now, find out the database `Pod`s using the following command,

```bash
$ kubectl get pods -n demo --selector="app.kubernetes.io/instance=restored-clickhouse"
NAME                                              READY   STATUS    RESTARTS   AGE
restored-clickhouse-appscode-cluster-shard-0-0   1/1     Running   0          2m46s
restored-clickhouse-appscode-cluster-shard-0-1   1/1     Running   0          2m41s
restored-clickhouse-appscode-cluster-shard-1-0   1/1     Running   0          2m44s
restored-clickhouse-appscode-cluster-shard-1-1   1/1     Running   0          2m40s
```

And copy the username and password of the admin user to access the `clickhouse-client` shell.

```bash
$ kubectl get secret -n demo restored-clickhouse-auth -o jsonpath='{.data.username}' | base64 -d
admin⏎

$ kubectl get secret -n demo restored-clickhouse-auth -o jsonpath='{.data.password}' | base64 -d
1OfTqKc8IzNgoLMi⏎
```

Now, let's exec into the `Pod` and verify the restored data. This time we check more than just the row values — since our `equipment` table is a `Distributed` table on top of a `ReplicatedMergeTree` table, we also verify that both the table engines and the per-shard data distribution were restored correctly.

```bash
$ kubectl exec -it -n demo restored-clickhouse-appscode-cluster-shard-0-0 -- clickhouse-client --user admin --password '1OfTqKc8IzNgoLMi'

:) SHOW DATABASES;

┌─name───────────────┐
│ INFORMATION_SCHEMA  │
│ default             │
│ information_schema  │
│ playground          │
│ system              │
└────────────────────┘

:) SHOW CREATE TABLE playground.equipment_local;

CREATE TABLE playground.equipment_local
(
    `id` UInt32,
    `type` String,
    `quant` UInt32,
    `color` String
)
ENGINE = ReplicatedMergeTree('/clickhouse/tables/{shard}/equipment_local', '{replica}')
ORDER BY id

:) SHOW CREATE TABLE playground.equipment;

CREATE TABLE playground.equipment
(
    `id` UInt32,
    `type` String,
    `quant` UInt32,
    `color` String
)
ENGINE = Distributed('appscode-cluster', 'playground', 'equipment_local', rand())

:) SELECT count() FROM playground.equipment;

15

:) exit
```

The `ReplicatedMergeTree` and `Distributed` table definitions came back exactly as they were, and the `Distributed` table again reports all 15 rows. Let's also confirm the per-shard split survived the restore, matching the original 12/3 split,

```bash
$ kubectl exec -n demo restored-clickhouse-appscode-cluster-shard-0-0 -- clickhouse-client --user admin --password '1OfTqKc8IzNgoLMi' -q "SELECT count() FROM playground.equipment_local"
12

$ kubectl exec -n demo restored-clickhouse-appscode-cluster-shard-1-0 -- clickhouse-client --user admin --password '1OfTqKc8IzNgoLMi' -q "SELECT count() FROM playground.equipment_local"
3

$ kubectl exec -n demo restored-clickhouse-appscode-cluster-shard-0-1 -- clickhouse-client --user admin --password '1OfTqKc8IzNgoLMi' -q "SELECT count() FROM playground.equipment_local"
12
```

So, from the above output, we can see that the `playground` database, the `equipment_local`/`equipment` tables, and the exact per-shard row distribution (including the shard-0 replica) from the original database are all restored successfully.

## Cleanup

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete restoresessions.core.kubestash.com -n demo sample-clickhouse-restore
kubectl delete backupconfigurations.core.kubestash.com -n demo sample-clickhouse-backup
kubectl delete retentionpolicies.storage.kubestash.com -n demo demo-retention
kubectl delete backupstorage -n demo s3-storage
kubectl delete secret -n demo s3-secret
kubectl delete clickhouse -n demo restored-clickhouse
kubectl delete clickhouse -n demo sample-clickhouse
```
