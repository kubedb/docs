---
title: Application Level Backup & Restore PostgreSQL | KubeStash
description: Application Level Backup and Restore using KubeStash
menu:
  docs_{{ .version }}:
    identifier: guides-application-level-backup-stashv2
    name: Application Level Backup
    parent: guides-rd-backup-stashv2
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Application Level Backup and Restore PostgreSQL database using KubeStash

KubeStash offers application-level backup and restore functionality for `PostgreSQL` databases. It captures both manifest and data backups of any `PostgreSQL` database in a single snapshot. During the restore process, KubeStash first applies the `PostgreSQL` manifest to the cluster and then restores the data into it.

This guide will give you an overview how you can take application-level backup and restore your `PostgreSQL` databases using `Kubestash`.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using `Minikube` or `Kind`.
- Install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md).
- Install `KubeStash` in your cluster following the steps [here](https://kubestash.com/docs/latest/setup/install/kubestash).
- Install KubeStash `kubectl` plugin following the steps [here](https://kubestash.com/docs/latest/setup/install/kubectl-plugin/).
- If you are not familiar with how KubeStash backup and restore PostgreSQL databases, please check the following guide [here](/docs/guides/redis/backup/kubestash/overview/index.md).

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

> **Note:** YAML files used in this tutorial are stored in [docs/guides/redis/backup/kubestash/application-level/examples](https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/redis/backup/kubestash/application-level/examples) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Backup PostgreSQL

KubeStash supports backups for `PostgreSQL` instances across different configurations, including Standalone and HA Cluster setups. In this demonstration, we'll focus on a `PostgreSQL` database using HA cluster configuration. The backup and restore process is similar for Standalone configuration.

This section will demonstrate how to take application-level backup of a `PostgreSQL` database. Here, we are going to deploy a `PostgreSQL` database using KubeDB. Then, we are going to back up the database at the application level to a `GCS` bucket. Finally, we will restore the entire `PostgreSQL` database.

### Deploy Sample PostgreSQL Database

Let's deploy a sample `PostgreSQL` database and insert some data into it.

**Create PostgreSQL CR:**

Below is the YAML of a sample `PostgreSQL` CR that we are going to create for this tutorial:

```yaml
apiVersion: kubedb.com/v1
kind: Redis
metadata:
  name: sample-redis
  namespace: demo
spec:
  version: "16.1"
  replicas: 3
  standbyMode: Hot
  streamingMode: Synchronous
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Create the above `PostgreSQL` CR,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/redis/backup/kubestash/application-level/examples/sample-redis.yaml
redis.kubedb.com/sample-redis created
```

KubeDB will deploy a `PostgreSQL` database according to the above specification. It will also create the necessary `Secrets` and `Services` to access the database.

Let's check if the database is ready to use,

```bash
$ kubectl get rd -n demo sample-redis
NAME              VERSION   STATUS   AGE
sample-redis   16.1      Ready    5m1s
```

The database is `Ready`. Verify that KubeDB has created a `Secret` and a `Service` for this database using the following commands,

```bash
$ kubectl get secret -n demo 
NAME                          TYPE                       DATA   AGE
sample-redis-auth          kubernetes.io/basic-auth   2      5m20s

$ kubectl get service -n demo -l=app.kubernetes.io/instance=sample-redis
NAME                      TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)                      AGE
sample-redis           ClusterIP   10.96.23.177   <none>        5432/TCP,2379/TCP            5m55s
sample-redis-pods      ClusterIP   None           <none>        5432/TCP,2380/TCP,2379/TCP   5m55s
sample-redis-standby   ClusterIP   10.96.26.118   <none>        5432/TCP                     5m55s
```

Here, we have to use service `sample-redis` and secret `sample-redis-auth` to connect with the database. `KubeDB` creates an [AppBinding](/docs/guides/redis/concepts/appbinding.md) CR that holds the necessary information to connect with the database.


**Verify AppBinding:**

Verify that the `AppBinding` has been created successfully using the following command,

```bash
$ kubectl get appbindings -n demo
NAME                       TYPE                  VERSION   AGE
sample-redis            kubedb.com/redis   16.1      9m30s
```

Let's check the YAML of the above `AppBinding`,

```bash
$ kubectl get appbindings -n demo sample-redis -o yaml
```

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1","kind":"Redis","metadata":{"annotations":{},"name":"sample-redis","namespace":"demo"},"spec":{"deletionPolicy":"DoNotTerminate","replicas":3,"standbyMode":"Hot","storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}}},"storageType":"Durable","streamingMode":"Synchronous","version":"16.1"}}
  creationTimestamp: "2024-09-04T10:07:04Z"
  generation: 1
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: sample-redis
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: redises.kubedb.com
  name: sample-redis
  namespace: demo
  ownerReferences:
  - apiVersion: kubedb.com/v1
    blockOwnerDeletion: true
    controller: true
    kind: Redis
    name: sample-redis
    uid: 0810a96c-a2b6-4e8a-a70a-51753660450c
  resourceVersion: "245972"
  uid: 73bdba85-c932-464b-93a8-7f1ba8dfff1b
spec:
  appRef:
    apiGroup: kubedb.com
    kind: Redis
    name: sample-redis
    namespace: demo
  clientConfig:
    service:
      name: sample-redis
      path: /
      port: 5432
      query: sslmode=disable
      scheme: redisql
  parameters:
    apiVersion: appcatalog.appscode.com/v1alpha1
    kind: StashAddon
    stash:
      addon:
        backupTask:
          name: redis-backup-16.1
        restoreTask:
          name: redis-restore-16.1
  secret:
    name: sample-redis-auth
  type: kubedb.com/redis
  version: "16.1"
```

KubeStash uses the `AppBinding` CR to connect with the target database. It requires the following two fields to set in AppBinding's `.spec` section.

Here,

- `.spec.clientConfig.service.name` specifies the name of the Service that connects to the database.
- `.spec.secret` specifies the name of the Secret that holds necessary credentials to access the database.
- `.spec.type` specifies the types of the app that this AppBinding is pointing to. KubeDB generated AppBinding follows the following format: `<app group>/<app resource type>`.

**Insert Sample Data:**

Now, we are going to exec into one of the database pod and create some sample data. At first, find out the database `Pod` using the following command,

```bash
$ kubectl get pods -n demo --selector="app.kubernetes.io/instance=sample-redis" 
NAME                READY   STATUS    RESTARTS   AGE
sample-redis-0   2/2     Running   0          16m
sample-redis-1   2/2     Running   0          13m
sample-redis-2   2/2     Running   0          13m
```

Now, let’s exec into the pod and create a table,

```bash
$ kubectl exec -it -n demo sample-redis-0 -- sh

# login as "redis" superuser.
/ $ psql -U redis
psql (16.1)
Type "help" for help.

# list available databases
redis=# \l
                                                        List of databases
     Name      |  Owner   | Encoding | Locale Provider |  Collate   |   Ctype    | ICU Locale | ICU Rules |   Access privileges   
---------------+----------+----------+-----------------+------------+------------+------------+-----------+-----------------------
 kubedb_system | redis | UTF8     | libc            | en_US.utf8 | en_US.utf8 |            |           | 
 redis      | redis | UTF8     | libc            | en_US.utf8 | en_US.utf8 |            |           | 
 template0     | redis | UTF8     | libc            | en_US.utf8 | en_US.utf8 |            |           | =c/redis          +
               |          |          |                 |            |            |            |           | redis=CTc/redis
 template1     | redis | UTF8     | libc            | en_US.utf8 | en_US.utf8 |            |           | =c/redis          +
               |          |          |                 |            |            |            |           | redis=CTc/redis
(4 rows)

# create a database named "demo"
redis=# create database demo;
CREATE DATABASE

# verify that the "demo" database has been created
redis=# \l
                                                        List of databases
     Name      |  Owner   | Encoding | Locale Provider |  Collate   |   Ctype    | ICU Locale | ICU Rules |   Access privileges   
---------------+----------+----------+-----------------+------------+------------+------------+-----------+-----------------------
 demo          | redis | UTF8     | libc            | en_US.utf8 | en_US.utf8 |            |           | 
 kubedb_system | redis | UTF8     | libc            | en_US.utf8 | en_US.utf8 |            |           | 
 redis      | redis | UTF8     | libc            | en_US.utf8 | en_US.utf8 |            |           | 
 template0     | redis | UTF8     | libc            | en_US.utf8 | en_US.utf8 |            |           | =c/redis          +
               |          |          |                 |            |            |            |           | redis=CTc/redis
 template1     | redis | UTF8     | libc            | en_US.utf8 | en_US.utf8 |            |           | =c/redis          +
               |          |          |                 |            |            |            |           | redis=CTc/redis
(5 rows)

# connect to the "demo" database
redis=# \c demo
You are now connected to database "demo" as user "redis".

# create a sample table
demo=# CREATE TABLE COMPANY( NAME TEXT NOT NULL, EMPLOYEE INT NOT NULL);
CREATE TABLE

# verify that the table has been created
demo=# \d
          List of relations
 Schema |  Name   | Type  |  Owner   
--------+---------+-------+----------
 public | company | table | redis
(1 row)

# insert multiple rows of data into the table
demo=# INSERT INTO COMPANY (NAME, EMPLOYEE) VALUES ('TechCorp', 100), ('InnovateInc', 150), ('AlphaTech', 200);
INSERT 0 3

# verify the data insertion
demo=# SELECT * FROM COMPANY;
    name     | employee 
-------------+----------
 TechCorp    |      100
 InnovateInc |      150
 AlphaTech   |      200
(3 rows)

# quit from the database
demo=# \q

# exit from the pod
/ $ exit
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
      bucket: kubestash-qa
      prefix: demo
      secretName: gcs-secret
  usagePolicy:
    allowedNamespaces:
      from: All
  default: true
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

We have to create a `BackupConfiguration` targeting respective `sample-redis` PostgreSQL database. Then, KubeStash will create a `CronJob` for each session to take periodic backup of that database.

At first, we need to create a secret with a Restic password for backup data encryption.

**Create Secret:**

Let's create a secret called `encrypt-secret` with the Restic password,

```bash
$ echo -n 'changeit' > RESTIC_PASSWORD
$ kubectl create secret generic -n demo encrypt-secret \
    --from-file=./RESTIC_PASSWORD \
secret "encrypt-secret" created
```

**Create BackupConfiguration:**

Below is the YAML for `BackupConfiguration` CR to take application-level backup of the `sample-redis` database that we have deployed earlier,

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: sample-redis-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Redis
    namespace: demo
    name: sample-redis
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
          - name: manifest-backup
          - name: logical-backup
```

- `.spec.sessions[*].schedule` specifies that we want to backup at `5 minutes` interval.
- `.spec.target` refers to the targeted `sample-redis` PostgreSQL database that we created earlier.
- `.spec.sessions[*].addon.tasks[*].name[*]` specifies that both the `manifest-backup` and `logical-backup` tasks will be executed.

Let's create the `BackupConfiguration` CR that we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/redis/kubestash/application-level/examples/backupconfiguration.yaml
backupconfiguration.core.kubestash.com/sample-redis-backup created
```

**Verify Backup Setup Successful**

If everything goes well, the phase of the `BackupConfiguration` should be `Ready`. The `Ready` phase indicates that the backup setup is successful. Let's verify the `Phase` of the BackupConfiguration,

```bash
$ kubectl get backupconfiguration -n demo
NAME                     PHASE   PAUSED   AGE
sample-redis-backup   Ready            2m50s
```

Additionally, we can verify that the `Repository` specified in the `BackupConfiguration` has been created using the following command,

```bash
$ kubectl get repo -n demo
NAME                  INTEGRITY   SNAPSHOT-COUNT   SIZE     PHASE   LAST-SUCCESSFUL-BACKUP   AGE
gcs-redis-repo                 0                0 B      Ready                            3m
```

KubeStash keeps the backup for `Repository` YAMLs. If we navigate to the GCS bucket, we will see the `Repository` YAML stored in the `demo/redis` directory.

**Verify CronJob:**

It will also create a `CronJob` with the schedule specified in `spec.sessions[*].scheduler.schedule` field of `BackupConfiguration` CR.

Verify that the `CronJob` has been created using the following command,

```bash
$ kubectl get cronjob -n demo
NAME                                             SCHEDULE      SUSPEND   ACTIVE   LAST SCHEDULE   AGE
trigger-sample-redis-backup-frequent-backup   */5 * * * *             0        2m45s           3m25s
```

**Verify BackupSession:**

KubeStash triggers an instant backup as soon as the `BackupConfiguration` is ready. After that, backups are scheduled according to the specified schedule.

```bash
$ kubectl get backupsession -n demo -w
NAME                                                INVOKER-TYPE          INVOKER-NAME              PHASE       DURATION   AGE
sample-redis-backup-frequent-backup-1725449400   BackupConfiguration   sample-redis-backup    Succeeded              7m22s
```

We can see from the above output that the backup session has succeeded. Now, we are going to verify whether the backed up data has been stored in the backend.

**Verify Backup:**

Once a backup is complete, KubeStash will update the respective `Repository` CR to reflect the backup. Check that the repository `sample-redis-backup` has been updated by the following command,

```bash
$ kubectl get repository -n demo gcs-redis-repo
NAME                       INTEGRITY   SNAPSHOT-COUNT   SIZE    PHASE   LAST-SUCCESSFUL-BACKUP   AGE
gcs-redis-repo          true        1                806 B   Ready   8m27s                    9m18s
```

At this moment we have one `Snapshot`. Run the following command to check the respective `Snapshot` which represents the state of a backup run for an application.

```bash
$ kubectl get snapshots -n demo -l=kubestash.com/repo-name=gcs-redis-repo
NAME                                                                  REPOSITORY          SESSION           SNAPSHOT-TIME          DELETION-POLICY   PHASE       AGE
gcs-redis-repo-sample-redis-backup-frequent-backup-1725449400   gcs-redis-repo   frequent-backup   2024-01-23T13:10:54Z   Delete            Succeeded   16h
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
$ kubectl get snapshots -n demo gcs-redis-repo-sample-redis-backup-frequent-backup-1725449400 -oyaml
```

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: Snapshot
metadata:
  creationTimestamp: "2024-09-05T09:08:03Z"
  finalizers:
  - kubestash.com/cleanup
  generation: 1
  labels:
    kubestash.com/app-ref-kind: Redis
    kubestash.com/app-ref-name: sample-redis
    kubestash.com/app-ref-namespace: demo
    kubestash.com/repo-name: gcs-redis-repo
  annotations:
    kubedb.com/db-version: "16.1"
  name: gcs-redis-repo-sample-redis-backup-frequent-backup-1725449400
  namespace: demo
  ownerReferences:
  - apiVersion: storage.kubestash.com/v1alpha1
    blockOwnerDeletion: true
    controller: true
    kind: Repository
    name: gcs-redis-repo
    uid: fa9086e5-285a-4b4a-9096-072bf7dbe2f7
  resourceVersion: "289843"
  uid: 43f17a3f-4ac7-443c-a139-151f2e5bf462
spec:
  appRef:
    apiGroup: kubedb.com
    kind: Redis
    name: sample-redis
    namespace: demo
  backupSession: sample-redis-backup-frequent-backup-1725527283
  deletionPolicy: Delete
  repository: gcs-redis-repo
  session: frequent-backup
  snapshotID: 01J70Q1NT6FW11YBBARRFJ6SYB
  type: FullBackup
  version: v1
status:
  components:
    dump:
      driver: Restic
      duration: 6.684476865s
      integrity: true
      path: repository/v1/frequent-backup/dump
      phase: Succeeded
      resticStats:
      - hostPath: dumpfile.sql
        id: 4b820700710f9f7b6a8d5b052367b51875e68dcd9052c749a686506db6a66374
        size: 3.345 KiB
        uploaded: 3.634 KiB
      size: 1.135 KiB
    manifest:
      driver: Restic
      duration: 7.477728298s
      integrity: true
      path: repository/v1/frequent-backup/manifest
      phase: Succeeded
      resticStats:
      - hostPath: /kubestash-tmp/manifest
        id: 9da4d1b7df6dd946e15a8a0d2a2a3c14776351e27926156770530ca03f6f8002
        size: 3.064 KiB
        uploaded: 1.443 KiB
      size: 2.972 KiB
  conditions:
  - lastTransitionTime: "2024-09-05T09:08:03Z"
    message: Recent snapshot list updated successfully
    reason: SuccessfullyUpdatedRecentSnapshotList
    status: "True"
    type: RecentSnapshotListUpdated
  - lastTransitionTime: "2024-09-05T09:08:49Z"
    message: Metadata uploaded to backend successfully
    reason: SuccessfullyUploadedSnapshotMetadata
    status: "True"
    type: SnapshotMetadataUploaded
  integrity: true
  phase: Succeeded
  size: 4.106 KiB
  snapshotTime: "2024-09-05T09:08:03Z"
  totalComponents: 2
```

> KubeStash uses `rd_dump` or `rd_dumpall` to perform backups of target `PostgreSQL` databases. Therefore, the component name for logical backups is set as `dump`.

> KubeStash set component name as `manifest` for the `manifest backup` of PostgreSQL databases.

Now, if we navigate to the GCS bucket, we will see the backed up data stored in the `demo/popstgres/repository/v1/frequent-backup/dump` directory. KubeStash also keeps the backup for `Snapshot` YAMLs, which can be found in the `demo/redis/snapshots` directory.

> Note: KubeStash stores all dumped data encrypted in the backup directory, meaning it remains unreadable until decrypted.

## Restore

In this section, we are going to restore the entire database from the backup that we have taken in the previous section.

For this tutorial, we will restore the database in a separate namespace called `dev`.

First, create the namespace by running the following command:

```bash
$ kubectl create ns dev
namespace/dev created
```

#### Create RestoreSession:

We need to create a RestoreSession CR.

Below, is the contents of YAML file of the `RestoreSession` CR that we are going to create to restore the entire database.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: restore-sample-redis
  namespace: demo
spec:
  manifestOptions:
    restoreNamespace: dev
    redis:
      db: true
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
      - name: manifest-restore
```

Here,

- `.spec.manifestOptions.redis.db` specifies whether to restore the DB manifest or not.
- `.spec.dataSource.repository` specifies the Repository object that holds the backed up data.
- `.spec.dataSource.snapshot` specifies to restore from latest `Snapshot`.
- `.spec.addon.tasks[*]` specifies that both the `manifest-restore` and `logical-backup-restore` tasks.

Let's create the RestoreSession CR object we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/redis/backup/kubestash/application-level/examples/restoresession.yaml
restoresession.core.kubestash.com/restore-sample-redis created
```

Once, you have created the `RestoreSession` object, KubeStash will create restore Job. Run the following command to watch the phase of the `RestoreSession` object,

```bash
$ watch kubectl get restoresession -n demo
Every 2.0s: kubectl get restores... AppsCode-PC-03: Wed Aug 21 10:44:05 2024
NAME                      REPOSITORY            FAILURE-POLICY   PHASE       DURATION   AGE
restore-sample-redis   gcs-redis-repo                      Succeeded   3s         53s
```

The `Succeeded` phase means that the restore process has been completed successfully.


#### Verify Restored PostgreSQL Manifest:

In this section, we will verify whether the desired `PostgreSQL` database manifest has been successfully applied to the cluster.

```bash
$ kubectl get redis -n dev 
NAME              VERSION   STATUS   AGE
sample-redis   16.1      Ready    9m46s
```

The output confirms that the `PostgreSQL` database has been successfully created with the same configuration as it had at the time of backup.


#### Verify Restored Data:

In this section, we are going to verify whether the desired data has been restored successfully. We are going to connect to the database server and check whether the database and the table we created earlier in the original database are restored.

At first, check if the database has gone into **`Ready`** state by the following command,

```bash
$ kubectl get redis -n dev sample-redis
NAME              VERSION   STATUS   AGE
sample-redis   16.1      Ready    9m46s
```

Now, find out the database `Pod` by the following command,

```bash
$ kubectl get pods -n dev --selector="app.kubernetes.io/instance=sample-redis"
NAME                READY   STATUS    RESTARTS   AGE
sample-redis-0   2/2     Running   0          12m
sample-redis-1   2/2     Running   0          12m
sample-redis-2   2/2     Running   0          12m
```


Now, lets exec one of the Pod and verify restored data.

```bash
$ kubectl exec -it -n dev sample-redis-0 -- /bin/sh
# login as "redis" superuser.
/ # psql -U redis
psql (11.11)
Type "help" for help.

# verify that the "demo" database has been restored
redis=# \l
                                                        List of databases
     Name      |  Owner   | Encoding | Locale Provider |  Collate   |   Ctype    | ICU Locale | ICU Rules |   Access privileges   
---------------+----------+----------+-----------------+------------+------------+------------+-----------+-----------------------
 demo          | redis | UTF8     | libc            | en_US.utf8 | en_US.utf8 |            |           | 
 kubedb_system | redis | UTF8     | libc            | en_US.utf8 | en_US.utf8 |            |           | 
 redis      | redis | UTF8     | libc            | en_US.utf8 | en_US.utf8 |            |           | 
 template0     | redis | UTF8     | libc            | en_US.utf8 | en_US.utf8 |            |           | =c/redis          +
               |          |          |                 |            |            |            |           | redis=CTc/redis
 template1     | redis | UTF8     | libc            | en_US.utf8 | en_US.utf8 |            |           | =c/redis          +
               |          |          |                 |            |            |            |           | redis=CTc/redis
(5 rows)

# connect to the "demo" database
redis=# \c demo
You are now connected to database "demo" as user "redis".

# verify that the sample table has been restored
demo=# \d
          List of relations
 Schema |  Name   | Type  |  Owner   
--------+---------+-------+----------
 public | company | table | redis
(1 row)

# Verify that the sample data has been restored 
demo=# SELECT * FROM COMPANY;
    name     | employee 
-------------+----------
 TechCorp    |      100
 InnovateInc |      150
 AlphaTech   |      200
(3 rows)

# disconnect from the database
demo=# \q

# exit from the pod
/ # exit
```

So, from the above output, we can see the `demo` database we had created in the original database `sample-redis` has been restored successfully.

## Cleanup

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete backupconfigurations.core.kubestash.com  -n demo sample-redis-backup
kubectl delete retentionpolicies.storage.kubestash.com -n demo demo-retention
kubectl delete restoresessions.core.kubestash.com -n demo restore-sample-redis
kubectl delete backupstorage -n demo gcs-storage
kubectl delete secret -n demo gcs-secret
kubectl delete secret -n demo encrypt-secret
kubectl delete redis -n demo sample-redis
kubectl delete redis -n dev sample-redis
```
