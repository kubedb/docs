---
title: Backup & Restore MariaDB | KubeStash
description: Backup ans Restore MariaDB database using KubeStash
menu:
  docs_{{ .version }}:
    identifier: guides-mariadb-logical-backup-stashv2
    name: Logical Backup
    parent: guides-mariadb-backup-stashv2
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Backup and Restore MariaDB database using KubeStash

KubeStash allows you to backup and restore `MariaDB` databases. It supports backups for `MariaDB` instances running in Standalone,  and Galera cluster configurations. KubeStash makes managing your `MariaDB` backups and restorations more straightforward and efficient.

This guide will give you an overview how you can take backup and restore your `MariaDB` databases using `Kubestash`.


## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using `Minikube` or `Kind`.
- Install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md).
- Install `KubeStash` in your cluster following the steps [here](https://kubestash.com/docs/latest/setup/install/kubestash).
- Install KubeStash `kubectl` plugin following the steps [here](https://kubestash.com/docs/latest/setup/install/kubectl-plugin/).
- If you are not familiar with how KubeStash backup and restore MariaDB databases, please check the following guide [here](/docs/guides/mariadb/backup/kubestash/overview/index.md).

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

> **Note:** YAML files used in this tutorial are stored in [docs/guides/mariadb/backup/kubestash/logical/examples](https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/backup/kubestash/logical/examples) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.


## Backup MariaDB

KubeStash supports backups for `MariaDB` instances across different configurations, including Standalone and Galera Cluster setups. In this demonstration, we'll focus on a `MariaDB` database using Galera cluster configuration. The backup and restore process is similar for Standalone configuration.

This section will demonstrate how to backup a `MariaDB` database. Here, we are going to deploy a `MariaDB` database using KubeDB. Then, we are going to backup this database into a `GCS` bucket. Finally, we are going to restore the backup up data into another `MariaDB` database.


### Deploy Sample MariaDB Database

Let's deploy a sample `MariaDB` database and insert some data into it.

**Create MariaDB CR:**

Below is the YAML of a sample `MariaDB` CR that we are going to create for this tutorial:

```yaml
apiVersion: kubedb.com/v1
kind: MariaDB
metadata:
  name: sample-mariadb
  namespace: demo
spec:
  version: 11.1.3
  replicas: 3
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Create the above `MariaDB` CR,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/backup/kubestash/logical/examples/sample-mariadb.yaml
mariadb.kubedb.com/sample-mariadb created
```

KubeDB will deploy a `MariaDB` database according to the above specification. It will also create the necessary `Secrets` and `Services` to access the database.

Let's check if the database is ready to use,

```bash
$ kubectl get md -n demo sample-mariadb
NAME                                VERSION   STATUS   AGE
mariadb.kubedb.com/sample-mariadb   11.1.3    Ready    5m4s
```

The database is `Ready`. Verify that KubeDB has created a `Secret` and a `Service` for this database using the following commands,

```bash
$ kubectl get secret -n demo 
NAME                          TYPE                       DATA   AGE
sample-mariadb-auth           kubernetes.io/basic-auth   2      5m49s

$ kubectl get service -n demo -l=app.kubernetes.io/instance=sample-mariadb
NAME                      TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)                      AGE
sample-mariadb            ClusterIP   10.128.7.155   <none>        3306/TCP                     6m28s
sample-mariadb-pods       ClusterIP   None           <none>        3306/TCP                     6m28s     
```

Here, we have to use service `sample-mariadb` and secret `sample-mariadb-auth` to connect with the database. `KubeDB` creates an [AppBinding](/docs/guides/mariadb/concepts/appbinding/index.md) CR that holds the necessary information to connect with the database.


**Verify AppBinding:**

Verify that the `AppBinding` has been created successfully using the following command,

```bash
$ kubectl get appbindings -n demo
NAME                       TYPE                  VERSION   AGE
sample-mariadb             kubedb.com/mariadb    11.1.3    7m56s
```

Let's check the YAML of the above `AppBinding`,

```bash
$ kubectl get appbindings -n demo sample-mariadb -o yaml
```

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1","kind":"MariaDB","metadata":{"annotations":{},"name":"sample-mariadb","namespace":"demo"},"spec":{"deletionPolicy":"WipeOut","replicas":3,"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}}},"storageType":"Durable","version":"11.1.3"}}
  creationTimestamp: "2024-09-17T10:07:37Z"
  generation: 1
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: sample-mariadb
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: mariadbs.kubedb.com
  name: sample-mariadb
  namespace: demo
  ownerReferences:
    - apiVersion: kubedb.com/v1
      blockOwnerDeletion: true
      controller: true
      kind: MariaDB
      name: sample-mariadb
      uid: c19117ca-582b-4d6c-90e5-ac80d5cf95b9
  resourceVersion: "1561857"
  uid: fd3868ef-f54b-4cc9-87b0-4986d9a8aaf0
spec:
  appRef:
    apiGroup: kubedb.com
    kind: MariaDB
    name: sample-mariadb
    namespace: demo
  clientConfig:
    service:
      name: sample-mariadb
      port: 3306
      scheme: tcp
    url: tcp(sample-mariadb.demo.svc:3306)/
  parameters:
    address: gcomm://sample-mariadb-0.sample-mariadb-pods.demo,sample-mariadb-1.sample-mariadb-pods.demo,sample-mariadb-2.sample-mariadb-pods.demo
    apiVersion: config.kubedb.com/v1alpha1
    group: sample-mariadb
    kind: GaleraArbitratorConfiguration
    sstMethod: xtrabackup-v2
    stash:
      addon:
        backupTask:
          name: mariadb-backup-10.5.8
        restoreTask:
          name: mariadb-restore-10.5.8
  secret:
    name: sample-mariadb-auth
    kind: Secret
  type: kubedb.com/mariadb
  version: 11.1.3
```

KubeStash uses the `AppBinding` CR to connect with the target database. It requires the following two fields to set in AppBinding's `.spec` section.

Here,

- `.spec.clientConfig.service.name` specifies the name of the Service that connects to the database.
- `.spec.secret` specifies the name of the Secret that holds necessary credentials to access the database.
- `.spec.type` specifies the types of the app that this AppBinding is pointing to. KubeDB generated AppBinding follows the following format: `<app group>/<app resource type>`.


**Insert Sample Data:**

Now, we are going to exec into one of the database pod and create some sample data. At first, find out the database `Pod` using the following command,

```bash
$ kubectl get pods -n demo --selector="app.kubernetes.io/instance=sample-mariadb" 
NAME                READY   STATUS    RESTARTS   AGE
sample-mariadb-0    2/2     Running   0          10m
sample-mariadb-1    2/2     Running   0          10m
sample-mariadb-2    2/2     Running   0          10m
```

Now, let’s exec into the pod and create a table,

```bash
$ kubectl exec -it -n demo  mariadb-0 -- bash

bash-4.4$ mariadb -uroot -p$MYSQL_ROOT_PASSWORD

MariaDB> create database hello;

MariaDB> use hello;

MariaDB [hello]> CREATE TABLE `demo_table`(
    ->     `id` BIGINT(20) NOT NULL,
    ->     `name` VARCHAR(255) DEFAULT NULL,
    ->     PRIMARY KEY (`id`)
    -> );

MariaDB [hello]> INSERT INTO `demo_table` (`id`, `name`)
    -> VALUES
    ->     (1, 'John'),
    ->     (2, 'Jane'),
    ->     (3, 'Bob'),
    ->     (4, 'Alice'),
    ->     (5, 'Charlie'),
    ->     (6, 'Diana'),
    ->     (7, 'Eve'),
    ->     (8, 'Frank'),
    ->     (9, 'Grace'),
    ->     (10, 'Henry');
    
MariaDB [hello]> select count(*) from demo_table;
+----------+
| count(*) |
+----------+
|       10 |
+----------+

```

Now, we are ready to backup the database.

### Prepare Backend

We are going to store our backup data into a `GCS` bucket. We have to create a `Secret` with necessary credentials and a `BackupStorage` CR to use this backend. If you want to use a different backend, please read the respective backend configuration doc from [here](https://kubestash.com/docs/latest/guides/backends/overview/).

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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/backup/kubestash/logical/examples/backupstorage.yaml
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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/backup/kubestash/logical/examples/retentionpolicy.yaml
retentionpolicy.storage.kubestash.com/demo-retention created
```

### Backup

We have to create a `BackupConfiguration` targeting respective `sample-mariadb` MariaDB database. Then, KubeStash will create a `CronJob` for each session to take periodic backup of that database.

At first, we need to create a secret with a Restic password for backup data encryption.

**Create Secret:**

Let's create a secret called `encrypt-secret` with the Restic password,

```bash
$ echo -n 'changeit' > RESTIC_PASSWORD
$ kubectl create secret generic -n demo encrypt-secret \
    --from-file=./RESTIC_PASSWORD \
secret "encrypt-secret" created
```

Below is the YAML for `BackupConfiguration` CR to backup the `sample-mariadb` database that we have deployed earlier,

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: sample-mariadb-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: MariaDB
    namespace: demo
    name: sample-mariadb
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
        - name: gcs-mariadb-repo
          backend: gcs-backend
          directory: /mariadb
          encryptionSecret:
            name: encrypt-secret
            namespace: demo
      addon:
        name: mariadb-addon
        tasks:
          - name: logical-backup
```

- `.spec.sessions[*].schedule` specifies that we want to backup the database at `5 minutes` interval.
- `.spec.target` refers to the targeted `sample-mariadb` MariaDB database that we created earlier.

Let's create the `BackupConfiguration` CR that we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/kubestash/logical/examples/backupconfiguration.yaml
backupconfiguration.core.kubestash.com/sample-mariadb-backup created
```

**Verify Backup Setup Successful**

If everything goes well, the phase of the `BackupConfiguration` should be `Ready`. The `Ready` phase indicates that the backup setup is successful. Let's verify the `Phase` of the BackupConfiguration,

```bash
$ kubectl get backupconfiguration -n demo
NAME                     PHASE   PAUSED   AGE
sample-mariadb-backup    Ready            2m50s
```

Additionally, we can verify that the `Repository` specified in the `BackupConfiguration` has been created using the following command,

```bash
$ kubectl get repo -n demo
NAME                  INTEGRITY   SNAPSHOT-COUNT   SIZE        PHASE   LAST-SUCCESSFUL-BACKUP   AGE
gcs-mariadb-repo      true        1                1.096 KiB   Ready   3m3s                     3m13s

```

KubeStash keeps the backup for `Repository` YAMLs. If we navigate to the GCS bucket, we will see the `Repository` YAML stored in the `demo/mariadb` directory.

**Verify CronJob:**

It will also create a `CronJob` with the schedule specified in `spec.sessions[*].scheduler.schedule` field of `BackupConfiguration` CR.

Verify that the `CronJob` has been created using the following command,

```bash
$ kubectl get cronjob -n demo
NAME                                             SCHEDULE     TIMEZONE   SUSPEND   ACTIVE   LAST SCHEDULE   AGE
trigger-sample-mariadb-backup-frequent-backup   */5 * * * *   <none>     False     0        <none>          4m23s
```

**Verify BackupSession:**

KubeStash triggers an instant backup as soon as the `BackupConfiguration` is ready. After that, backups are scheduled according to the specified schedule.

```bash
$ kubectl get backupsession -n demo -w
NAME                                                INVOKER-TYPE          INVOKER-NAME              PHASE       DURATION   AGE
sample-mariadb-backup-frequent-backup-1725449400    BackupConfiguration   sample-mariadb-backup     Succeeded              7m22s
```

We can see from the above output that the backup session has succeeded. Now, we are going to verify whether the backup data has been stored in the backend.

**Verify Backup:**

Once a backup is complete, KubeStash will update the respective `Repository` CR to reflect the backup. Check that the repository `sample-mariadb-backup` has been updated by the following command,

```bash
$ kubectl get repository -n demo gcs-mariadb-repo
NAME                       INTEGRITY   SNAPSHOT-COUNT   SIZE    PHASE   LAST-SUCCESSFUL-BACKUP   AGE
gcs-mariadb-repo           true        1                806 B   Ready   8m27s                    9m18s
```

At this moment we have one `Snapshot`. Run the following command to check the respective `Snapshot` which represents the state of a backup run for an application.

```bash
$ kubectl get snapshot.storage.kubestash.com -n demo -l=kubestash.com/repo-name=gcs-mariadb-repo
NAME                                                                  REPOSITORY          SESSION           SNAPSHOT-TIME          DELETION-POLICY   PHASE       AGE
gcs-mariadb-repo-sample-mariadb-ckup-frequent-backup-1726569774       gcs-mariadb-repo    frequent-backup   2024-09-17T10:43:04Z   Delete            Succeeded   41m
```

> Note: KubeStash creates a `Snapshot` with the following labels:
> - `kubestash.com/app-ref-kind: <target-kind>`
> - `kubestash.com/app-ref-name: <target-name>`
> - `kubestash.com/app-ref-namespace: <target-namespace>`
> - `kubestash.com/repo-name: <repository-name>`
>
> These labels can be used to watch only the `Snapshot`s related to our target Database or `Repository`.

If we check the YAML of the `Snapshot`, we can find the information about the backup components of the Database.

```bash
$ kubectl get snapshot.storage.kubestash.com -n demo gcs-mariadb-repo-sample-mariadb-ckup-frequent-backup-1726569774 -oyaml
```

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: Snapshot
metadata:
  creationTimestamp: "2024-09-17T10:43:04Z"
  finalizers:
    - kubestash.com/cleanup
  generation: 1
  labels:
    kubestash.com/app-ref-kind: MariaDB
    kubestash.com/app-ref-name: sample-mariadb
    kubestash.com/app-ref-namespace: demo
    kubestash.com/repo-name: gcs-mariadb-repo
  annotations:
    kubedb.com/db-version: 11.1.3
  name: gcs-mariadb-repo-sample-mariadb-ckup-frequent-backup-1726569774
  namespace: demo
  ownerReferences:
    - apiVersion: storage.kubestash.com/v1alpha1
      blockOwnerDeletion: true
      controller: true
      kind: Repository
      name: gcs-mariadb-repo
      uid: 7abe82a2-5ecc-4904-b848-910becab54bc
  resourceVersion: "1566893"
  uid: a18e95f2-d7ec-4334-9ba4-f689280a51ab
spec:
  appRef:
    apiGroup: kubedb.com
    kind: MariaDB
    name: sample-mariadb
    namespace: demo
  backupSession: sample-mariadb-backup-frequent-backup-1726569774
  deletionPolicy: Delete
  repository: gcs-mariadb-repo
  session: frequent-backup
  snapshotID: 01J7ZS88VNPD5B7HFVXVP576P6
  type: FullBackup
  version: v1
status:
  components:
    dump:
      driver: Restic
      duration: 3.279948033s
      integrity: true
      path: repository/v1/frequent-backup/dump
      phase: Succeeded
      resticStats:
        - hostPath: dumpfile.sql
          id: 0a6dfb754cb32bdaf17581fa42b20e8915aabd0b48f37c854b72812f53b7e5b6
          size: 2.206 KiB
          uploaded: 2.498 KiB
      size: 1.096 KiB
  conditions:
    - lastTransitionTime: "2024-09-17T10:43:04Z"
      message: Recent snapshot list updated successfully
      reason: SuccessfullyUpdatedRecentSnapshotList
      status: "True"
      type: RecentSnapshotListUpdated
    - lastTransitionTime: "2024-09-17T10:43:24Z"
      message: Metadata uploaded to backend successfully
      reason: SuccessfullyUploadedSnapshotMetadata
      status: "True"
      type: SnapshotMetadataUploaded
  integrity: true
  phase: Succeeded
  size: 1.096 KiB
  snapshotTime: "2024-09-17T10:43:04Z"
  totalComponents: 1
```

> KubeStash uses `mariadb-dump` to perform backups of target `MariaDB` databases. Therefore, the component name for logical backups is set as `dump`.

Now, if we navigate to the GCS bucket, we will see the backup data stored in the `demo/mariadb/repository/v1/frequent-backup/dump` directory. KubeStash also keeps the backup for `Snapshot` YAMLs, which can be found in the `demo/mariadb/snapshots` directory.

> Note: KubeStash stores all dumped data encrypted in the backup directory, meaning it remains unreadable until decrypted.

## Restore

In this section, we are going to restore the database from the backup we have taken in the previous section. We are going to deploy a new database and initialize it from the backup.

Now, we have to deploy the restored database similarly as we have deployed the original `sample-mariadb` database. However, this time there will be the following differences:

- We are going to specify `.spec.init.waitForInitialRestore` field that tells KubeDB to wait for first restore to complete before marking this database is ready to use.

Below is the YAML for `MariaDB` CR we are going deploy to initialize from backup,

```yaml
apiVersion: kubedb.com/v1
kind: MariaDB
metadata:
  name: restored-mariadb
  namespace: demo
spec:
  init:
    waitForInitialRestore: true
  version: 11.1.3
  replicas: 3
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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/backup/kubestash/logical/examples/restored-mariadb.yaml
mariadb.kubedb.com/restore-mariadb created
```

If you check the database status, you will see it is stuck in **`Provisioning`** state.

```bash
$ kubectl get mariadb -n demo restored-mariadb
NAME                VERSION   STATUS         AGE
restored-mariadb    11.1.3    Provisioning   110s
```

#### Create RestoreSession:

Now, we need to create a `RestoreSession` CR pointing to targeted `MariaDB` database.

Below, is the contents of YAML file of the `RestoreSession` object that we are going to create to restore backup data into the newly created `MariaDB` database named `restored-mariadb`.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: sample-mariadb-restore
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: MariaDB
    namespace: demo
    name: restored-mariadb
  dataSource:
    repository: gcs-mariadb-repo
    snapshot: latest
    encryptionSecret:
      name: encrypt-secret
      namespace: demo
  addon:
    name: mariadb-addon
    tasks:
      - name: logical-backup-restore
```

Here,

- `.spec.target` refers to the newly created `restored-mariadb` MariaDB object to where we want to restore backup data.
- `.spec.dataSource.repository` specifies the Repository object that holds the backup data.
- `.spec.dataSource.snapshot` specifies to restore from latest `Snapshot`.

Let's create the RestoreSession CRD object we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/backup/kubestash/logical/examples/restoresession.yaml
restoresession.core.kubestash.com/sample-mariadb-restore created
```

Once, you have created the `RestoreSession` object, KubeStash will create restore Job. Run the following command to watch the phase of the `RestoreSession` object,

```bash
$ watch kubectl get restoresession -n demo
NAME                      REPOSITORY          FAILURE-POLICY   PHASE       DURATION   AGE
sample-mariadb-restore   gcs-mariadb-repo                    Succeeded   7s         116s
```

The `Succeeded` phase means that the restore process has been completed successfully.

#### Verify Restored Data:

In this section, we are going to verify whether the desired data has been restored successfully. We are going to connect to the database server and check whether the database and the table we created earlier in the original database are restored.

At first, check if the database has gone into **`Ready`** state by the following command,

```bash
$ kubectl get mariadb -n demo restored-mariadb
NAME                VERSION   STATUS   AGE
restored-mariadb    11.1.3    Ready    6m
```

Now, find out the database `Pod` by the following command,

```bash
$ kubectl get pods -n demo --selector="app.kubernetes.io/instance=restored-mariadb"
NAME                            READY   STATUS      RESTARTS   AGE
restored-mariadb-0              2/2     Running     0          7m
restored-mariadb-1              2/2     Running     0          7m
restored-mariadb-2              2/2     Running     0          7m
```

Now, lets exec one of the `Pod` and verify restored data.

```bash
$ kubectl exec -it -n demo restored-mariadb-0 -- bash
mysql@restored-mariadb-0:/$ mariadb -uroot -p$MYSQL_ROOT_PASSWORD

MariaDB> use hello;

MariaDB [hello]> select count(*) from demo_table;
+----------+
| count(*) |
+----------+
|       10 |
+----------+

```

So, from the above output, we can see the `hello` database we had created in the original database `sample-mariadb` has been restored in the `restored-mariadb` database.

## Cleanup

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete backupconfigurations.core.kubestash.com  -n demo sample-mariadb-backup
kubectl delete restoresessions.core.kubestash.com -n demo restore-sample-mariadb
kubectl delete retentionpolicies.storage.kubestash.com -n demo demo-retention
kubectl delete backupstorage -n demo gcs-storage
kubectl delete secret -n demo gcs-secret
kubectl delete secret -n demo encrypt-secret
kubectl delete mariadb -n demo restored-mariadb
kubectl delete mariadb -n demo sample-mariadb
```