---
title: Backup & Restore PostgreSQL | KubeStash
description: Backup ans Restore PostgreSQL database using KubeStash
menu:
  docs_{{ .version }}:
    identifier: guides-pg-logical-backup-stashv2
    name: Logical Backup
    parent: guides-pg-backup-stashv2
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Backup and Restore PostgreSQL database using KubeStash

KubeStash allows you to backup and restore `PostgreSQL` databases. It supports backups for `PostgreSQL` instances running in Standalone,  and HA cluster configurations. KubeStash makes managing your `PostgreSQL` backups and restorations more straightforward and efficient.

This guide will give you an overview how you can take backup and restore your `PostgreSQL` databases using `Kubestash`.


## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using `Minikube` or `Kind`.
- Install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md).
- Install `KubeStash` in your cluster following the steps [here](https://kubestash.com/docs/latest/setup/install/kubestash).
- Install KubeStash `kubectl` plugin following the steps [here](https://kubestash.com/docs/latest/setup/install/kubectl-plugin/).
- If you are not familiar with how KubeStash backup and restore PostgreSQL databases, please check the following guide [here](/docs/guides/postgres/backup/kubestash/overview/index.md).

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

> **Note:** YAML files used in this tutorial are stored in [docs/guides/postgres/backup/kubestash/logical/examples](https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/backup/kubestash/logical/examples) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.


## Backup PostgreSQL

KubeStash supports backups for `PostgreSQL` instances across different configurations, including Standalone and HA Cluster setups. In this demonstration, we'll focus on a `PostgreSQL` database using HA cluster configuration. The backup and restore process is similar for Standalone configuration.

This section will demonstrate how to backup a `PostgreSQL` database. Here, we are going to deploy a `PostgreSQL` database using KubeDB. Then, we are going to backup this database into a `GCS` bucket. Finally, we are going to restore the backup up data into another `PostgreSQL` database.


### Deploy Sample PostgreSQL Database

Let's deploy a sample `PostgreSQL` database and insert some data into it.

**Create PostgreSQL CR:**

Below is the YAML of a sample `PostgreSQL` CR that we are going to create for this tutorial:

```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: sample-postgres
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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/backup/kubestash/logical/examples/sample-postgres.yaml
postgres.kubedb.com/sample-postgres created
```

KubeDB will deploy a `PostgreSQL` database according to the above specification. It will also create the necessary `Secrets` and `Services` to access the database.

Let's check if the database is ready to use,

```bash
$ kubectl get pg -n demo sample-postgres
NAME              VERSION   STATUS   AGE
sample-postgres   16.1      Ready    5m1s
```

The database is `Ready`. Verify that KubeDB has created a `Secret` and a `Service` for this database using the following commands,

```bash
$ kubectl get secret -n demo 
NAME                          TYPE                       DATA   AGE
sample-postgres-auth          kubernetes.io/basic-auth   2      5m20s

$ kubectl get service -n demo -l=app.kubernetes.io/instance=sample-postgres
NAME                      TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)                      AGE
sample-postgres           ClusterIP   10.96.23.177   <none>        5432/TCP,2379/TCP            5m55s
sample-postgres-pods      ClusterIP   None           <none>        5432/TCP,2380/TCP,2379/TCP   5m55s
sample-postgres-standby   ClusterIP   10.96.26.118   <none>        5432/TCP                     5m55s
```

Here, we have to use service `sample-postgres` and secret `sample-postgres-auth` to connect with the database. `KubeDB` creates an [AppBinding](/docs/guides/postgres/concepts/appbinding.md) CR that holds the necessary information to connect with the database.


**Verify AppBinding:**

Verify that the `AppBinding` has been created successfully using the following command,

```bash
$ kubectl get appbindings -n demo
NAME                       TYPE                  VERSION   AGE
sample-postgres            kubedb.com/postgres   16.1      9m30s
```

Let's check the YAML of the above `AppBinding`,

```bash
$ kubectl get appbindings -n demo sample-postgres -o yaml
```

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1","kind":"Postgres","metadata":{"annotations":{},"name":"sample-postgres","namespace":"demo"},"spec":{"deletionPolicy":"DoNotTerminate","replicas":3,"standbyMode":"Hot","storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}}},"storageType":"Durable","streamingMode":"Synchronous","version":"16.1"}}
  creationTimestamp: "2024-09-04T10:07:04Z"
  generation: 1
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: sample-postgres
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: postgreses.kubedb.com
  name: sample-postgres
  namespace: demo
  ownerReferences:
  - apiVersion: kubedb.com/v1
    blockOwnerDeletion: true
    controller: true
    kind: Postgres
    name: sample-postgres
    uid: 0810a96c-a2b6-4e8a-a70a-51753660450c
  resourceVersion: "245972"
  uid: 73bdba85-c932-464b-93a8-7f1ba8dfff1b
spec:
  appRef:
    apiGroup: kubedb.com
    kind: Postgres
    name: sample-postgres
    namespace: demo
  clientConfig:
    service:
      name: sample-postgres
      path: /
      port: 5432
      query: sslmode=disable
      scheme: postgresql
  parameters:
    apiVersion: appcatalog.appscode.com/v1alpha1
    kind: StashAddon
    stash:
      addon:
        backupTask:
          name: postgres-backup-16.1
        restoreTask:
          name: postgres-restore-16.1
  secret:
    name: sample-postgres-auth
  type: kubedb.com/postgres
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
$ kubectl get pods -n demo --selector="app.kubernetes.io/instance=sample-postgres" 
NAME                READY   STATUS    RESTARTS   AGE
sample-postgres-0   2/2     Running   0          16m
sample-postgres-1   2/2     Running   0          13m
sample-postgres-2   2/2     Running   0          13m
```

Now, let’s exec into the pod and create a table,

```bash
$ kubectl exec -it -n demo sample-postgres-0 -- sh

# login as "postgres" superuser.
/ $ psql -U postgres
psql (16.1)
Type "help" for help.

# list available databases
postgres=# \l
                                                        List of databases
     Name      |  Owner   | Encoding | Locale Provider |  Collate   |   Ctype    | ICU Locale | ICU Rules |   Access privileges   
---------------+----------+----------+-----------------+------------+------------+------------+-----------+-----------------------
 kubedb_system | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |            |           | 
 postgres      | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |            |           | 
 template0     | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |            |           | =c/postgres          +
               |          |          |                 |            |            |            |           | postgres=CTc/postgres
 template1     | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |            |           | =c/postgres          +
               |          |          |                 |            |            |            |           | postgres=CTc/postgres
(4 rows)

# create a database named "demo"
postgres=# create database demo;
CREATE DATABASE

# verify that the "demo" database has been created
postgres=# \l
                                                        List of databases
     Name      |  Owner   | Encoding | Locale Provider |  Collate   |   Ctype    | ICU Locale | ICU Rules |   Access privileges   
---------------+----------+----------+-----------------+------------+------------+------------+-----------+-----------------------
 demo          | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |            |           | 
 kubedb_system | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |            |           | 
 postgres      | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |            |           | 
 template0     | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |            |           | =c/postgres          +
               |          |          |                 |            |            |            |           | postgres=CTc/postgres
 template1     | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |            |           | =c/postgres          +
               |          |          |                 |            |            |            |           | postgres=CTc/postgres
(5 rows)

# connect to the "demo" database
postgres=# \c demo
You are now connected to database "demo" as user "postgres".

# create a sample table
demo=# CREATE TABLE COMPANY( NAME TEXT NOT NULL, EMPLOYEE INT NOT NULL);
CREATE TABLE

# verify that the table has been created
demo=# \d
          List of relations
 Schema |  Name   | Type  |  Owner   
--------+---------+-------+----------
 public | company | table | postgres
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
  deletionPolicy: WipeOut
```

Let's create the BackupStorage we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/backup/kubestash/logical/examples/backupstorage.yaml
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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/backup/kubestash/logical/examples/retentionpolicy.yaml
retentionpolicy.storage.kubestash.com/demo-retention created
```

### Backup

We have to create a `BackupConfiguration` targeting respective `sample-postgres` PostgreSQL database. Then, KubeStash will create a `CronJob` for each session to take periodic backup of that database.

At first, we need to create a secret with a Restic password for backup data encryption.

**Create Secret:**

Let's create a secret called `encrypt-secret` with the Restic password,

```bash
$ echo -n 'changeit' > RESTIC_PASSWORD
$ kubectl create secret generic -n demo encrypt-secret \
    --from-file=./RESTIC_PASSWORD \
secret "encrypt-secret" created
```

Below is the YAML for `BackupConfiguration` CR to backup the `sample-postgres` database that we have deployed earlier,

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: sample-postgres-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Postgres
    namespace: demo
    name: sample-postgres
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
        - name: gcs-postgres-repo
          backend: gcs-backend
          directory: /postgres
          encryptionSecret:
            name: encrypt-secret
            namespace: demo
      addon:
        name: postgres-addon
        tasks:
          - name: logical-backup
```

- `.spec.sessions[*].schedule` specifies that we want to backup the database at `5 minutes` interval.
- `.spec.target` refers to the targeted `sample-postgres` PostgreSQL database that we created earlier.

Let's create the `BackupConfiguration` CR that we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/kubestash/logical/examples/backupconfiguration.yaml
backupconfiguration.core.kubestash.com/sample-postgres-backup created
```

**Verify Backup Setup Successful**

If everything goes well, the phase of the `BackupConfiguration` should be `Ready`. The `Ready` phase indicates that the backup setup is successful. Let's verify the `Phase` of the BackupConfiguration,

```bash
$ kubectl get backupconfiguration -n demo
NAME                     PHASE   PAUSED   AGE
sample-postgres-backup   Ready            2m50s
```

Additionally, we can verify that the `Repository` specified in the `BackupConfiguration` has been created using the following command,

```bash
$ kubectl get repo -n demo
NAME                  INTEGRITY   SNAPSHOT-COUNT   SIZE     PHASE   LAST-SUCCESSFUL-BACKUP   AGE
gcs-postgres-repo                 0                0 B      Ready                            3m
```

KubeStash keeps the backup for `Repository` YAMLs. If we navigate to the GCS bucket, we will see the `Repository` YAML stored in the `demo/postgres` directory.

**Verify CronJob:**

It will also create a `CronJob` with the schedule specified in `spec.sessions[*].scheduler.schedule` field of `BackupConfiguration` CR.

Verify that the `CronJob` has been created using the following command,

```bash
$ kubectl get cronjob -n demo
NAME                                             SCHEDULE      SUSPEND   ACTIVE   LAST SCHEDULE   AGE
trigger-sample-postgres-backup-frequent-backup   */5 * * * *             0        2m45s           3m25s
```

**Verify BackupSession:**

KubeStash triggers an instant backup as soon as the `BackupConfiguration` is ready. After that, backups are scheduled according to the specified schedule.

```bash
$ kubectl get backupsession -n demo -w
NAME                                                INVOKER-TYPE          INVOKER-NAME              PHASE       DURATION   AGE
sample-postgres-backup-frequent-backup-1725449400   BackupConfiguration   sample-postgres-backup    Succeeded              7m22s
```

We can see from the above output that the backup session has succeeded. Now, we are going to verify whether the backed up data has been stored in the backend.

**Verify Backup:**

Once a backup is complete, KubeStash will update the respective `Repository` CR to reflect the backup. Check that the repository `sample-postgres-backup` has been updated by the following command,

```bash
$ kubectl get repository -n demo sample-postgres-backup
NAME                       INTEGRITY   SNAPSHOT-COUNT   SIZE    PHASE   LAST-SUCCESSFUL-BACKUP   AGE
sample-postgres-backup     true        1                806 B   Ready   8m27s                    9m18s
```

At this moment we have one `Snapshot`. Run the following command to check the respective `Snapshot` which represents the state of a backup run for an application.

```bash
$ kubectl get snapshots -n demo -l=kubestash.com/repo-name=gcs-postgres-repo
NAME                                                                  REPOSITORY          SESSION           SNAPSHOT-TIME          DELETION-POLICY   PHASE       AGE
gcs-postgres-repo-sample-postgres-backup-frequent-backup-1725449400   gcs-postgres-repo   frequent-backup   2024-01-23T13:10:54Z   Delete            Succeeded   16h
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
$ kubectl get snapshots -n demo gcs-postgres-repo-sample-postgres-backup-frequent-backup-1725449400 -oyaml
```

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: Snapshot
metadata:
  creationTimestamp: "2024-09-04T11:30:00Z"
  finalizers:
  - kubestash.com/cleanup
  generation: 1
  labels:
    kubestash.com/app-ref-kind: Postgres
    kubestash.com/app-ref-name: sample-postgres
    kubestash.com/app-ref-namespace: demo
    kubestash.com/repo-name: gcs-postgres-repo
  annotations:
    kubedb.com/db-version: "16.1"
  name: gcs-postgres-repo-sample-postgres-backup-frequent-backup-1725449400
  namespace: demo
  ownerReferences:
  - apiVersion: storage.kubestash.com/v1alpha1
    blockOwnerDeletion: true
    controller: true
    kind: Repository
    name: gcs-postgres-repo
    uid: 1009bd4a-b211-49f1-a64c-3c979c699a81
  resourceVersion: "253523"
  uid: c6757c49-e13b-4a36-9f7d-64eae350423f
spec:
  appRef:
    apiGroup: kubedb.com
    kind: Postgres
    name: sample-postgres
    namespace: demo
  backupSession: sample-postgres-backup-frequent-backup-1725449400
  deletionPolicy: Delete
  repository: gcs-postgres-repo
  session: frequent-backup
  snapshotID: 01J6YCRWEWAKACMGZYR2R7YJ5C
  type: FullBackup
  version: v1
status:
  components:
    dump:
      driver: Restic
      duration: 11.526138009s
      integrity: true
      path: repository/v1/frequent-backup/dump
      phase: Succeeded
      resticStats:
      - hostPath: dumpfile.sql
        id: 008eb87193e7db112e9ad8f42c9302c851a1fbacb7165a5cb3aa2d27dd210764
        size: 3.345 KiB
        uploaded: 299 B
      size: 2.202 KiB
  conditions:
  - lastTransitionTime: "2024-09-04T11:30:00Z"
    message: Recent snapshot list updated successfully
    reason: SuccessfullyUpdatedRecentSnapshotList
    status: "True"
    type: RecentSnapshotListUpdated
  - lastTransitionTime: "2024-09-04T11:30:32Z"
    message: Metadata uploaded to backend successfully
    reason: SuccessfullyUploadedSnapshotMetadata
    status: "True"
    type: SnapshotMetadataUploaded
  integrity: true
  phase: Succeeded
  size: 2.201 KiB
  snapshotTime: "2024-09-04T11:30:00Z"
  totalComponents: 1
```

> KubeStash uses a logical backup approach to take backups of target `PostgreSQL` databases. Therefore, the component name for logical backups is set as `dump`. Do the same for auto-backup, application backup and customize backup if necessary.

Now, if we navigate to the GCS bucket, we will see the backed up data stored in the `demo/popstgres/repository/v1/frequent-backup/dump` directory. KubeStash also keeps the backup for `Snapshot` YAMLs, which can be found in the `demo/postgres/snapshots` directory.

> Note: KubeStash stores all dumped data encrypted in the backup directory, meaning it remains unreadable until decrypted.

## Restore

In this section, we are going to restore the database from the backup we have taken in the previous section. We are going to deploy a new database and initialize it from the backup.

Now, we have to deploy the restored database similarly as we have deployed the original `sample-postgres` database. However, this time there will be the following differences:

- We are going to specify `.spec.init.waitForInitialRestore` field that tells KubeDB to wait for first restore to complete before marking this database is ready to use.

Below is the YAML for `PostgreSQL` CR we are going deploy to initialize from backup,

```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: restored-postgres
  namespace: demo
spec:
  init:
    waitForInitialRestore: true
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

Let's create the above database,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/backup/kubestash/logical/examples/restored-postgres.yaml
postgres.kubedb.com/restore-postgres created
```

If you check the database status, you will see it is stuck in **`Provisioning`** state.

```bash
$ kubectl get postgres -n demo restored-postgres
NAME                VERSION   STATUS         AGE
restored-postgres   8.2.0     Provisioning   61s
```

#### Create RestoreSession:

Now, we need to create a `RestoreSession` CR pointing to targeted `PostgreSQL` database.

Below, is the contents of YAML file of the `RestoreSession` object that we are going to create to restore backed up data into the newly created `PostgreSQL` database named `restored-postgres`.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: sample-postgres-restore
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Postgres
    namespace: demo
    name: restored-postgres
  dataSource:
    repository: gcs-postgres-repo
    snapshot: latest
    encryptionSecret:
      name: encrypt-secret
      namespace: demo
  addon:
    name: postgres-addon
    tasks:
      - name: logical-backup-restore
```

Here,

- `.spec.target` refers to the newly created `restored-postgres` PostgreSQL object to where we want to restore backup data.
- `.spec.dataSource.repository` specifies the Repository object that holds the backed up data.
- `.spec.dataSource.snapshot` specifies to restore from latest `Snapshot`.

Let's create the RestoreSession CRD object we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/backup/kubestash/logical/examples/restoresession.yaml
restoresession.core.kubestash.com/sample-postgres-restore created
```

Once, you have created the `RestoreSession` object, KubeStash will create restore Job. Run the following command to watch the phase of the `RestoreSession` object,

```bash
$ watch kubectl get restoresession -n demo
Every 2.0s: kubectl get restores... AppsCode-PC-03: Wed Aug 21 10:44:05 2024
NAME                      REPOSITORY          FAILURE-POLICY   PHASE       DURATION   AGE
sample-postgres-restore   gcs-postgres-repo                    Succeeded   7s         116s
```

The `Succeeded` phase means that the restore process has been completed successfully.

#### Verify Restored Data:

In this section, we are going to verify whether the desired data has been restored successfully. We are going to connect to the database server and check whether the database and the table we created earlier in the original database are restored.

At first, check if the database has gone into **`Ready`** state by the following command,

```bash
$ kubectl get postgres -n demo restored-postgres
NAME                VERSION   STATUS   AGE
restored-postgres   16.1      Ready    6m31s
```

Now, find out the database `Pod` by the following command,

```bash
$ kubectl get pods -n demo --selector="app.kubernetes.io/instance=restored-postgres"
NAME                            READY   STATUS      RESTARTS   AGE
restored-postgres-0             2/2     Running     0          6m7s
restored-postgres-1             2/2     Running     0          6m1s
restored-postgres-2             2/2     Running     0          5m55s
```

Now, lets exec one of the `Pod` and verify restored data.

```bash
$ kubectl exec -it -n demo restored-postgres-0 -- /bin/sh
# login as "postgres" superuser.
/ # psql -U postgres
psql (11.11)
Type "help" for help.

# verify that the "demo" database has been restored
postgres=# \l
                                                        List of databases
     Name      |  Owner   | Encoding | Locale Provider |  Collate   |   Ctype    | ICU Locale | ICU Rules |   Access privileges   
---------------+----------+----------+-----------------+------------+------------+------------+-----------+-----------------------
 demo          | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |            |           | 
 kubedb_system | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |            |           | 
 postgres      | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |            |           | 
 template0     | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |            |           | =c/postgres          +
               |          |          |                 |            |            |            |           | postgres=CTc/postgres
 template1     | postgres | UTF8     | libc            | en_US.utf8 | en_US.utf8 |            |           | =c/postgres          +
               |          |          |                 |            |            |            |           | postgres=CTc/postgres
(5 rows)

# connect to the "demo" database
postgres=# \c demo
You are now connected to database "demo" as user "postgres".

# verify that the sample table has been restored
demo=# \d
          List of relations
 Schema |  Name   | Type  |  Owner   
--------+---------+-------+----------
 public | company | table | postgres
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

So, from the above output, we can see the `demo` database we had created in the original database `sample-postgres` has been restored in the `restored-postgres` database.

## Cleanup

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete backupconfigurations.core.kubestash.com  -n demo sample-postgres-backup
kubectl delete restoresessions.core.kubestash.com -n demo restore-sample-postgres
kubectl delete retentionpolicies.storage.kubestash.com -n demo demo-retention
kubectl delete backupstorage -n demo gcs-storage
kubectl delete secret -n demo gcs-secret
kubectl delete secret -n demo encrypt-secret
kubectl delete postgres -n demo restored-postgres
kubectl delete postgres -n demo sample-postgres
```