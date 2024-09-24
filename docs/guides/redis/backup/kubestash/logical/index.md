---
title: Backup & Restore Redis | KubeStash
description: Backup ans Restore Redis database using KubeStash
menu:
  docs_{{ .version }}:
    identifier: guides-rd-logical-backup-stashv2
    name: Logical Backup
    parent: guides-rd-backup-stashv2
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Backup and Restore Redis database using KubeStash

KubeStash allows you to backup and restore `Redis` databases. It supports backups for `Redis` instances running in Standalone, Cluster and Sentinel mode. KubeStash makes managing your `Redis` backups and restorations more straightforward and efficient.

This guide will give you an overview how you can take backup and restore your `Redis` databases using `Kubestash`.


## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using `Minikube` or `Kind`.
- Install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md).
- Install `KubeStash` in your cluster following the steps [here](https://kubestash.com/docs/latest/setup/install/kubestash).
- Install KubeStash `kubectl` plugin following the steps [here](https://kubestash.com/docs/latest/setup/install/kubectl-plugin/).
- If you are not familiar with how KubeStash backup and restore Redis databases, please check the following guide [here](/docs/guides/redis/backup/kubestash/overview/index.md).

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

> **Note:** YAML files used in this tutorial are stored in [docs/guides/redis/backup/kubestash/logical/examples](https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/redis/backup/kubestash/logical/examples) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.


## Backup Redis

KubeStash supports backups for `Redis` instances across different configurations, including Standalone, Cluster and Sentinel mode setups. In this demonstration, we'll focus on a `Redis` database in Cluster mode. The backup and restore process is similar for Standalone and Sentinel mode.

This section will demonstrate how to backup a `Redis` database. Here, we are going to deploy a `Redis` database using KubeDB. Then, we are going to backup this database into a `GCS` bucket. Finally, we are going to restore the backup up data into another `Redis` database.


### Deploy Sample Redis Database

Let's deploy a sample `Redis` database and insert some data into it.

**Create Redis CR:**

Below is the YAML of a sample `Redis` CR that we are going to create for this tutorial:

```yaml
apiVersion: kubedb.com/v1
kind: Redis
metadata:
  name: redis-cluster
  namespace: demo
spec:
  version: 7.4.0
  mode: Cluster
  cluster:
    shards: 3
    replicas: 2
  storageType: Durable
  storage:
    storageClassName: "standard"
    resources:
      requests:
        storage: 1Gi
    accessModes:
      - ReadWriteOnce
  deletionPolicy: Delete
```

Create the above `Redis` CR,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/redis/backup/kubestash/logical/examples/redis-cluster.yaml
redis.kubedb.com/redis-cluster created
```

KubeDB will deploy a `Redis` database according to the above specification. It will also create the necessary `Secrets` and `Services` to access the database.

Let's check if the database is ready to use,

```bash
$ kubectl get rd -n demo redis-cluster
NAME            VERSION   STATUS   AGE
redis-cluster   7.4.0     Ready    5m2s
```

The database is `Ready`. Verify that KubeDB has created a `Secret` and a `Service` for this database using the following commands,

```bash
$ kubectl get secret -n demo 
NAME                   TYPE                       DATA   AGE
redis-cluster-auth     kubernetes.io/basic-auth   2      6m16s
redis-cluster-config   Opaque                     1      6m16s


$ kubectl get service -n demo -l=app.kubernetes.io/instance=redis-cluster
NAME                 TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
redis-cluster        ClusterIP   10.96.185.242   <none>        6379/TCP   7m25s
redis-cluster-pods   ClusterIP   None            <none>        6379/TCP   7m25s
```

Here, we have to use service `redis-cluster` and secret `redis-cluster-auth` to connect with the database. `KubeDB` creates an [AppBinding](/docs/guides/redis/concepts/appbinding.md) CR that holds the necessary information to connect with the database.


**Verify AppBinding:**

Verify that the `AppBinding` has been created successfully using the following command,

```bash
$ kubectl get appbindings -n demo
NAME            TYPE               VERSION   AGE
redis-cluster   kubedb.com/redis   7.4.0     7m14s
```

Let's check the YAML of the above `AppBinding`,

```bash
$ kubectl get appbindings -n demo redis-cluster -o yaml
```

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1","kind":"Redis","metadata":{"annotations":{},"name":"redis-cluster","namespace":"demo"},"spec":{"cluster":{"replicas":2,"shards":3},"deletionPolicy":"Delete","mode":"Cluster","storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"},"storageType":"Durable","version":"7.4.0"}}
  creationTimestamp: "2024-09-18T05:29:09Z"
  generation: 1
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: redis-cluster
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: redises.kubedb.com
  name: redis-cluster
  namespace: demo
  ownerReferences:
    - apiVersion: kubedb.com/v1
      blockOwnerDeletion: true
      controller: true
      kind: Redis
      name: redis-cluster
      uid: 089eff8d-81ec-4933-8121-87d5e21d137d
  resourceVersion: "1139825"
  uid: d985a52a-00ef-4857-b597-0ccec62cf838
spec:
  appRef:
    apiGroup: kubedb.com
    kind: Redis
    name: redis-cluster
    namespace: demo
  clientConfig:
    service:
      name: redis-cluster
      port: 6379
      scheme: redis
  parameters:
    apiVersion: config.kubedb.com/v1alpha1
    kind: RedisConfiguration
    stash:
      addon:
        backupTask:
          name: redis-backup-7.0.5
        restoreTask:
          name: redis-restore-7.0.5
  secret:
    name: redis-cluster-auth
  type: kubedb.com/redis
  version: 7.4.0
```

KubeStash uses the `AppBinding` CR to connect with the target database. It requires the following two fields to set in AppBinding's `.spec` section.

Here,

- `.spec.clientConfig.service.name` specifies the name of the Service that connects to the database.
- `.spec.secret` specifies the name of the Secret that holds necessary credentials to access the database.
- `.spec.type` specifies the types of the app that this AppBinding is pointing to. KubeDB generated AppBinding follows the following format: `<app group>/<app resource type>`.


**Insert Sample Data:**

Now, we are going to exec into one of the database pod and create some sample data. At first, find out the database `Pod` using the following command,

```bash
$ kubectl get pods -n demo --selector="app.kubernetes.io/instance=redis-cluster" 
NAME                     READY   STATUS    RESTARTS   AGE
redis-cluster-shard0-0   1/1     Running   0          11m
redis-cluster-shard0-1   1/1     Running   0          11m
redis-cluster-shard1-0   1/1     Running   0          11m
redis-cluster-shard1-1   1/1     Running   0          11m
redis-cluster-shard2-0   1/1     Running   0          10m
redis-cluster-shard2-1   1/1     Running   0          10m
```

#### Connection Information

- Hostname/address: you can use any of these
    - Service: `redis-cluster.demo`
    - Pod IP: (`$ kubectl get pod -n demo -l app.kubernetes.io/name=redises.kubedb.com -o yaml | grep podIP`)
- Port: `6379`
- Username: Run following command to get _username_,

  ```bash
  $ kubectl get secrets -n demo redis-cluster-auth -o jsonpath='{.data.\username}' | base64 -d
  default
  ```

- Password: Run the following command to get _password_,

  ```bash
  $ kubectl get secrets -n demo redis-cluster-auth -o jsonpath='{.data.\password}' | base64 -d
  8UnSPM;(~cXWWs60
  ```
Now, let’s exec into the pod and insert some data,

```bash
$ kubectl exec -it -n demo redis-cluster-shard0-0 -c redis -- bash
redis@redis-cluster-shard0-0:/data$ redis-cli -c
127.0.0.1:6379> auth default 8UnSPM;(~cXWWs60
OK
127.0.0.1:6379> set db redis
OK
127.0.0.1:6379> set name neaj
-> Redirected to slot [5798] located at 10.244.0.48:6379
OK
10.244.0.48:6379> set key value
-> Redirected to slot [12539] located at 10.244.0.52:6379
OK
10.244.0.52:6379> exit
redis@redis-cluster-shard0-0:/data$ exit
exit
```

Now, we are ready to backup the database.

### Prepare Backend

We are going to store our backed up data into a `GCS` bucket. We have to create a `Secret` with necessary credentials and a `BackupStorage` CR to use this backend. If you want to use a different backend, please read the respective backend configuration doc from [here](https://kubestash.com/docs/latest/guides/backends/overview/).

**Create Secret:**

Let's create a secret called `gcs-secret` with access credentials to our desired GCS bucket,

```bash
$ echo -n '<your-project-id>' > GOOGLE_PROJECT_ID
$ cat /path/to/downloaded-sa-key.json > GOOGLE_SERVICE_ACCOUNT_JSON_KEY
$ kubectl create secret generic -n demo gcs-secret \
    --from-file=./GOOGLE_PROJECT_ID \
    --from-file=./GOOGLE_SERVICE_ACCOUNT_JSON_KEY
secret/gcs-secret created
```

**Create BackupStorage:**

Now, create a `BackupStorage` using this secret. Below is the YAML of `BackupStorage` CR we are going to create,

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: BackupStorage
metadata:
  name: gcs-storage
  namespace: demo
spec:
  storage:
    provider: gcs
    gcs:
      bucket: neaj-demo
      prefix: demo
      secretName: gcs-secret
  usagePolicy:
    allowedNamespaces:
      from: All
  deletionPolicy: Delete
```

Let's create the BackupStorage we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/redis/backup/kubestash/logical/examples/backupstorage.yaml
backupstorage.storage.kubestash.com/gcs-storage created
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
  maxRetentionPeriod: 2mo
  successfulSnapshots:
    last: 5
  usagePolicy:
    allowedNamespaces:
      from: All
```

Let’s create the above `RetentionPolicy`,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/redis/backup/kubestash/logical/examples/retentionpolicy.yaml
retentionpolicy.storage.kubestash.com/demo-retention created
```

### Backup

We have to create a `BackupConfiguration` targeting respective `redis-cluster` Redis database. Then, KubeStash will create a `CronJob` for each session to take periodic backup of that database.

At first, we need to create a secret with a Restic password for backup data encryption.

**Create Secret:**

Let's create a secret called `encrypt-secret` with the Restic password,

```bash
$ echo -n 'changeit' > RESTIC_PASSWORD
$ kubectl create secret generic -n demo encrypt-secret \
    --from-file=./RESTIC_PASSWORD
secret "encrypt-secret" created
```

Below is the YAML for `BackupConfiguration` CR to backup the `redis-cluster` database that we have deployed earlier,

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: redis-cluster-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Redis
    namespace: demo
    name: redis-cluster
  backends:
    - name: gcs-backend
      storageRef:
        namespace: demo
        name: gcs-storage
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
        - name: gcs-redis-repo
          backend: gcs-backend
          directory: /redis
          encryptionSecret:
            name: encrypt-secret
            namespace: demo
      addon:
        name: redis-addon
        tasks:
          - name: logical-backup
```

- `.spec.sessions[*].schedule` specifies that we want to backup the database at `5 minutes` interval.
- `.spec.target` refers to the targeted `redis-cluster` Redis database that we created earlier.

Let's create the `BackupConfiguration` CR that we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/redis/kubestash/logical/examples/backupconfiguration.yaml
backupconfiguration.core.kubestash.com/redis-cluster-backup created
```

**Verify Backup Setup Successful**

If everything goes well, the phase of the `BackupConfiguration` should be `Ready`. The `Ready` phase indicates that the backup setup is successful. Let's verify the `Phase` of the BackupConfiguration,

```bash
$ kubectl get backupconfiguration -n demo
NAME                   PHASE   PAUSED   AGE
redis-cluster-backup   Ready            71s
```

Additionally, we can verify that the `Repository` specified in the `BackupConfiguration` has been created using the following command,

```bash
$ kubectl get repo -n demo
NAME              INTEGRITY   SNAPSHOT-COUNT   SIZE        PHASE   LAST-SUCCESSFUL-BACKUP   AGE
gcs-redis-repo                0                0 B         Ready                            2m30s
```

KubeStash keeps the backup for `Repository` YAMLs. If we navigate to the GCS bucket, we will see the `Repository` YAML stored in the `demo/redis` directory.

**Verify CronJob:**

It will also create a `CronJob` with the schedule specified in `spec.sessions[*].scheduler.schedule` field of `BackupConfiguration` CR.

Verify that the `CronJob` has been created using the following command,

```bash
$ kubectl get cronjob -n demo
NAME                                           SCHEDULE      SUSPEND   ACTIVE   LAST SCHEDULE   AGE
trigger-redis-cluster-backup-frequent-backup   */5 * * * *   False     0        45s             2m38s
```

**Verify BackupSession:**

KubeStash triggers an instant backup as soon as the `BackupConfiguration` is ready. After that, backups are scheduled according to the specified schedule.

```bash
$ kubectl get backupsession -n demo -w
NAME                                              INVOKER-TYPE          INVOKER-NAME           PHASE       DURATION   AGE
redis-cluster-backup-frequent-backup-1726651666   BackupConfiguration   redis-cluster-backup   Succeeded   2m25s      2m56s
```

We can see from the above output that the backup session has succeeded. Now, we are going to verify whether the backed up data has been stored in the backend.

**Verify Backup:**

Once a backup is complete, KubeStash will update the respective `Repository` CR to reflect the backup. Check that the repository `gcs-redis-repo` has been updated by the following command,

```bash
$ kubectl get repository -n demo gcs-redis-repo
NAME             INTEGRITY   SNAPSHOT-COUNT   SIZE    PHASE   LAST-SUCCESSFUL-BACKUP   AGE
gcs-redis-repo   true        1                416 B   Ready   4m40s                    5m
```

At this moment we have one `Snapshot`. Run the following command to check the respective `Snapshot` which represents the state of a backup run for an application.

```bash
$ kubectl get snapshots -n demo -l=kubestash.com/repo-name=gcs-redis-repo
NAME                                                             REPOSITORY       SESSION           SNAPSHOT-TIME          DELETION-POLICY   PHASE       AGE
gcs-redis-repo-redis-cluster-backup-frequent-backup-1726651666   gcs-redis-repo   frequent-backup   2024-09-18T09:28:07Z   Delete            Succeeded   5m14s
```

> Note: KubeStash creates a `Snapshot` with the following labels:
> - `kubestash.com/app-ref-kind: <target-kind>`
> - `kubestash.com/app-ref-name: <target-name>`
> - `kubestash.com/app-ref-namespace: <target-namespace>`
> - `kubestash.com/repo-name: <repository-name>`
>
> These labels can be used to watch only the `Snapshot`s related to our target Database or `Repository`.

If we check the YAML of the `Snapshot`, we can find the information about the backed up components of the Database.

```bash
$ kubectl get snapshots -n demo gcs-redis-repo-redis-cluster-backup-frequent-backup-1726651666 -oyaml
```

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: Snapshot
metadata:
  creationTimestamp: "2024-09-18T09:28:07Z"
  finalizers:
    - kubestash.com/cleanup
  generation: 1
  labels:
    kubedb.com/db-version: 7.4.0
    kubestash.com/app-ref-kind: Redis
    kubestash.com/app-ref-name: redis-cluster
    kubestash.com/app-ref-namespace: demo
    kubestash.com/repo-name: gcs-redis-repo
  name: gcs-redis-repo-redis-cluster-backup-frequent-backup-1726651666
  namespace: demo
  ownerReferences:
    - apiVersion: storage.kubestash.com/v1alpha1
      blockOwnerDeletion: true
      controller: true
      kind: Repository
      name: gcs-redis-repo
      uid: 6b4439c8-5c79-443d-af14-a8efd47eb43c
  resourceVersion: "1161141"
  uid: 04110ce9-a015-4d50-a66f-dbc685a4fdff
spec:
  appRef:
    apiGroup: kubedb.com
    kind: Redis
    name: redis-cluster
    namespace: demo
  backupSession: redis-cluster-backup-frequent-backup-1726651666
  deletionPolicy: Delete
  repository: gcs-redis-repo
  session: frequent-backup
  snapshotID: 01J827BRDW8PT3TS9T8QR6KS2S
  type: FullBackup
  version: v1
status:
  components:
    dump:
      driver: Restic
      duration: 30.177171779s
      integrity: true
      path: repository/v1/frequent-backup/dump
      phase: Succeeded
      resticStats:
        - hostPath: dumpfile.resp
          id: dc0c7e16ffea238d80f2f0e23b94d5eee1a598a4b5b9bc3f9edc2e9059e1d9e2
          size: 381 B
          uploaded: 680 B
      size: 416 B
  conditions:
    - lastTransitionTime: "2024-09-18T09:28:07Z"
      message: Recent snapshot list updated successfully
      reason: SuccessfullyUpdatedRecentSnapshotList
      status: "True"
      type: RecentSnapshotListUpdated
    - lastTransitionTime: "2024-09-18T09:30:29Z"
      message: Metadata uploaded to backend successfully
      reason: SuccessfullyUploadedSnapshotMetadata
      status: "True"
      type: SnapshotMetadataUploaded
  integrity: true
  phase: Succeeded
  size: 416 B
  snapshotTime: "2024-09-18T09:28:07Z"
  totalComponents: 1

```

> KubeStash uses [redis-dump-go](https://github.com/yannh/redis-dump-go) to perform backups of target `Redis` databases. Therefore, the component name for logical backups is set as `dump`.

Now, if we navigate to the GCS bucket, we will see the backed up data stored in the `demo/redis/repository/v1/frequent-backup/dump` directory. KubeStash also keeps the backup for `Snapshot` YAMLs, which can be found in the `demo/redis/snapshots` directory.

> Note: KubeStash stores all dumped data encrypted in the backup directory, meaning it remains unreadable until decrypted.

## Restore

In this section, we are going to restore the database from the backup we have taken in the previous section. We are going to deploy a new database and initialize it from the backup.

Now, we have to deploy the restored database similarly as we have deployed the original `redis-cluster` database. However, this time there will be the following differences:

- We are going to specify `.spec.init.waitForInitialRestore` field that tells KubeDB to wait for first restore to complete before marking this database is ready to use.

Below is the YAML for `Redis` CR we are going deploy to initialize from backup,

```yaml
apiVersion: kubedb.com/v1
kind: Redis
metadata:
  name: restored-redis-cluster
  namespace: demo
spec:
  init:
    waitForInitialRestore: true
  version: 7.4.0
  mode: Cluster
  cluster:
    shards: 3
    replicas: 2
  storageType: Durable
  storage:
    storageClassName: "standard"
    resources:
      requests:
        storage: 1Gi
    accessModes:
      - ReadWriteOnce
  deletionPolicy: Delete
```

Let's create the above database,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/redis/backup/kubestash/logical/examples/restored-redis-cluster.yaml
redis.kubedb.com/restore-redis-cluster created
```

If you check the database status, you will see it is stuck in **`Provisioning`** state.

```bash
$ kubectl get redis -n demo restored-redis-cluster
NAME                     VERSION   STATUS         AGE
restored-redis-cluster   7.4.0     Provisioning   2m35s
```

#### Create RestoreSession:

Now, we need to create a `RestoreSession` CR pointing to targeted `Redis` database.

Below, is the contents of YAML file of the `RestoreSession` object that we are going to create to restore backed up data into the newly created `Redis` database named `restored-redis-cluster`.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: redis-cluster-restore
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Redis
    namespace: demo
    name: restored-redis-cluster
  dataSource:
    repository: gcs-redis-repo
    snapshot: latest
    encryptionSecret:
      name: encrypt-secret
      namespace: demo
  addon:
    name: redis-addon
    tasks:
      - name: logical-backup-restore
```

Here,

- `.spec.target` refers to the newly created `restored-redis-cluster` Redis object to where we want to restore backup data.
- `.spec.dataSource.repository` specifies the Repository object that holds the backed up data.
- `.spec.dataSource.snapshot` specifies to restore from latest `Snapshot`.

Let's create the RestoreSession CRD object we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/redis/backup/kubestash/logical/examples/restoresession.yaml
restoresession.core.kubestash.com/redis-cluster-restore created
```

Once, you have created the `RestoreSession` object, KubeStash will create restore Job. Run the following command to watch the phase of the `RestoreSession` object,

```bash
$ watch kubectl get restoresession -n demo
Every 2.0s: kubectl get restoresession -n demo                                                  neaj-desktop: Wed Sep 18 15:53:42 2024
NAME                    REPOSITORY       FAILURE-POLICY   PHASE       DURATION   AGE
redis-cluster-restore   gcs-redis-repo                    Succeeded   1m26s      4m49s
```

The `Succeeded` phase means that the restore process has been completed successfully.

#### Verify Restored Data:

In this section, we are going to verify whether the desired data has been restored successfully. We are going to connect to the database and check whether the data we inserted earlier in the original database are restored.

At first, check if the database has gone into **`Ready`** state by the following command,

```bash
$ kubectl get redis -n demo restored-redis-cluster
NAME                     VERSION   STATUS   AGE
restored-redis-cluster   7.4.0     Ready    8m42s
```

Now, find out the database `Pods` by the following command,

```bash
$ kubectl get pods -n demo --selector="app.kubernetes.io/instance=restored-redis-cluster"
NAME                              READY   STATUS      RESTARTS   AGE
restored-redis-cluster-shard0-0   1/1     Running     0          5m53s
restored-redis-cluster-shard0-1   1/1     Running     0          5m47s
restored-redis-cluster-shard1-0   1/1     Running     0          5m31s
restored-redis-cluster-shard1-1   1/1     Running     0          5m24s
restored-redis-cluster-shard2-0   1/1     Running     0          5m9s
restored-redis-cluster-shard2-1   1/1     Running     0          5m2s
```

Now, lets exec one of the `Pod` and verify restored data.

```bash
$ kubectl exec -it -n demo restored-redis-cluster-shard0-0 -c redis -- bash
redis@restored-redis-cluster-shard0-0:/data$ redis-cli -c
127.0.0.1:6379> auth default lm~;mv7H~eahvZCc
OK
127.0.0.1:6379> get db 
"redis"
127.0.0.1:6379> get name 
-> Redirected to slot [5798] located at 10.244.0.66:6379
"neaj"
10.244.0.66:6379> get key
-> Redirected to slot [12539] located at 10.244.0.70:6379
"value"
10.244.0.70:6379> exit
redis@restored-redis-cluster-shard0-0:/data$ exit
exit
```

So, from the above output, we can see the `redis-cluster` database we had created earlier has been restored in the `restored-redis-cluster` database successfully.

## Cleanup

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete backupconfigurations.core.kubestash.com  -n demo redis-cluster-backup
kubectl delete restoresessions.core.kubestash.com -n demo redis-cluster-restore
kubectl delete retentionpolicies.storage.kubestash.com -n demo demo-retention
kubectl delete backupstorage -n demo gcs-storage
kubectl delete secret -n demo gcs-secret
kubectl delete secret -n demo encrypt-secret
kubectl delete redis -n demo restored-redis-cluster
kubectl delete redis -n demo redis-cluster
```