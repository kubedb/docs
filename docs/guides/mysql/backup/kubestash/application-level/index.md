---
title: Application Level Backup & Restore MySQL | KubeStash
description: Application Level Backup and Restore using KubeStash
menu:
  docs_{{ .version }}:
    identifier: guides-application-level-backup-stashv2
    name: Application Level Backup
    parent: guides-mysql-backup-stashv2
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Application Level Backup and Restore MySQL database using KubeStash

KubeStash offers application-level backup and restore functionality for `MySQL` databases. It captures both manifest and logical data backups of any `MySQL` database in a single snapshot. During the restore process, KubeStash first applies the `MySQL` manifest to the cluster and then restores the data into it.

This guide will give you how you can take application-level backup and restore your `MySQL` databases using `Kubestash`.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using `Minikube` or `Kind`.
- Install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md).
- Install `KubeStash` in your cluster following the steps [here](https://kubestash.com/docs/latest/setup/install/kubestash).
- Install KubeStash `kubectl` plugin following the steps [here](https://kubestash.com/docs/latest/setup/install/kubectl-plugin/).
- If you are not familiar with how KubeStash backup and restore MySQL databases, please check the following guide [here](/docs/guides/mysql/backup/kubestash/overview/index.md).

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

> **Note:** YAML files used in this tutorial are stored in [docs/guides/mysql/backup/kubestash/application-level/examples](docs/guides/mysql/backup/kubestash/application-level/examples) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Backup MySQL

KubeStash supports backups for `MySQL` instances across different configurations, including Standalone, Group Replication, and InnoDB Cluster setups. In this demonstration, we'll focus on a `MySQL` database using Group Replication. The backup and restore process is similar for Standalone and InnoDB Cluster configurations as well.

This section will demonstrate how to take application-level backup of a `MySQL` database. Here, we are going to deploy a `MySQL` database using KubeDB. Then, we are going to back up the database at the application level to a `GCS` bucket. Finally, we will restore the entire `MySQL` database.

### Deploy Sample MySQL Database

Let's deploy a sample `MySQL` database and insert some data into it.

**Create MySQL CR:**

Below is the YAML of a sample `MySQL` CR that we are going to create for this tutorial:

```yaml
apiVersion: kubedb.com/v1
kind: MySQL
metadata:
  name: sample-mysql
  namespace: demo
spec:
  version: "8.2.0"
  replicas: 3
  topology:
    mode: GroupReplication  
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 50Mi
  deletionPolicy: WipeOut
```

Here,
- `.spec.topology` specifies about the clustering configuration of MySQL.
- `.Spec.topology.mode` specifies the mode of MySQL Cluster. During the demonstration we consider to use `GroupReplication`.

Create the above `MySQL` CR,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/backup/kubestash/application-level/examples/sample-mysql.yaml
mysql.kubedb.com/sample-mysql created
```

KubeDB will deploy a MySQL database according to the above specification. It will also create the necessary Secrets and Services to access the database.

Let's check if the database is ready to use,

```bash
$ kubectl get mysqls.kubedb.com -n demo
NAME           VERSION   STATUS    AGE
sample-mysql   8.2.0     Ready     4m22s
```

The database is `Ready`. Verify that KubeDB has created a `Secret` and a `Service` for this database using the following commands,

```bash
$ kubectl get secret -n demo 
NAME                TYPE     DATA   AGE
sample-mysql-auth   Opaque   2      4m58s

$ kubectl get service -n demo -l=app.kubernetes.io/instance=sample-mysql
NAME                   TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)    AGE
sample-mysql           ClusterIP   10.96.55.61     <none>        3306/TCP   97s
sample-mysql-pods      ClusterIP   None            <none>        3306/TCP   97s
sample-mysql-standby   ClusterIP   10.96.211.186   <none>        3306/TCP   97
```

Here, we have to use service `sample-mysql` and secret `sample-mysql-auth` to connect with the database. `KubeDB` creates an [AppBinding](/docs/guides/mysql/concepts/appbinding/index.md) CR that holds the necessary information to connect with the database.

**Verify AppBinding:**

Verify that the `AppBinding` has been created successfully using the following command,

```bash
$ kubectl get appbindings -n demo
NAME           AGE
sample-mysql   9m24s
```

Let's check the YAML of the above `AppBinding`,

```bash
$ kubectl get appbindings -n demo sample-mysql -o yaml
```

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: sample-mysql
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: mysqls.kubedb.com
  name: sample-mysql
  namespace: demo
  ownerReferences:
  - apiVersion: kubedb.com/v1
    blockOwnerDeletion: true
    controller: true
    kind: MySQL
    name: sample-mysql
    uid: edde3e8b-7775-4f91-85a9-4ba4b96315f7
  resourceVersion: "5126"
  uid: 86c9a149-f8ab-44c4-947f-5f9b402aad6c
spec:
  appRef:
    apiGroup: kubedb.com
    kind: MySQL
    name: sample-mysql
    namespace: demo
  clientConfig:
    service:
      name: sample-mysql
      path: /
      port: 3306
      scheme: tcp
    url: tcp(sample-mysql.demo.svc:3306)/
    ...
    ...
  secret:
    name: sample-mysql-auth
  type: kubedb.com/mysql
  version: 8.2.0
```

KubeStash uses the `AppBinding` CR to connect with the target database. It requires the following two fields to set in AppBinding's `.spec` section.

- `.spec.clientConfig.service.name` specifies the name of the Service that connects to the database.
- `.spec.secret` specifies the name of the Secret that holds necessary credentials to access the database.
- `spec.type` specifies the types of the app that this AppBinding is pointing to. KubeDB generated AppBinding follows the following format: `<app group>/<app resource type>`.

**Insert Sample Data:**

Now, we are going to exec into the database pod and create some sample data. At first, find out the database Pod using the following command,

```bash
$ kubectl get pods -n demo --selector="app.kubernetes.io/instance=sample-mysql"
NAME             READY   STATUS    RESTARTS   AGE
sample-mysql-0   2/2     Running   0          33m
sample-mysql-1   2/2     Running   0          33m
sample-mysql-2   2/2     Running   0          33m
```

And copy the username and password of the `root` user to access into `mysql` shell.

```bash
$ kubectl get secret -n demo  sample-mysql-auth -o jsonpath='{.data.username}'| base64 -d
root⏎

$ kubectl get secret -n demo  sample-mysql-auth -o jsonpath='{.data.password}'| base64 -d
DZfmUZd14fNEEOU4⏎
```

Now, Lets exec into the Pod to enter into `mysql` shell and create a database and a table,

```bash
$ kubectl exec -it -n demo sample-mysql-0 -- mysql --user=root --password=DZfmUZd14fNEEOU4
Defaulted container "mysql" out of: mysql, mysql-init (init)
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 977
Server version: 8.2.0 MySQL Community Server - GPL

Copyright (c) 2000, 2023, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> CREATE DATABASE playground;
Query OK, 1 row affected (0.01 sec)

mysql> SHOW DATABASES;
+--------------------+
| Database           |
+--------------------+
| information_schema |
| mysql              |
| performance_schema |
| playground         |
| sys                |
+--------------------+
5 rows in set (0.00 sec)

mysql> CREATE TABLE playground.equipment ( id INT NOT NULL AUTO_INCREMENT, type VARCHAR(50), quant INT, color VARCHAR(25), PRIMARY KEY(id));
Query OK, 0 rows affected (0.01 sec)

mysql> SHOW TABLES IN playground;
+----------------------+
| Tables_in_playground |
+----------------------+
| equipment            |
+----------------------+
1 row in set (0.01 sec)

mysql> INSERT INTO playground.equipment (type, quant, color) VALUES ("slide", 2, "blue");
Query OK, 1 row affected (0.01 sec)

mysql> SELECT * FROM playground.equipment;
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  1 | slide |     2 | blue  |
+----+-------+-------+-------+
1 row in set (0.00 sec)

mysql> exit
Bye
```
Now, we are ready to backup the database.

### Prepare Backend

We are going to store our backed up data into a GCS bucket. We have to create a Secret with necessary credentials and a `BackupStorage` CR to use this backend. If you want to use a different backend, please read the respective backend configuration doc from [here](https://kubestash.com/docs/latest/guides/backends/overview/).

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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/backup/kubestash/application-level/examples/backupstorage.yaml
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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/backup/kubestash/application-level/examples/retentionpolicy.yaml
retentionpolicy.storage.kubestash.com/demo-retention created
```

### Backup

We have to create a `BackupConfiguration` targeting respective `sample-mysql` MySQL database. Then, KubeStash will create a `CronJob` for each session to take periodic backup of that database.

At first, we need to create a secret with a Restic password for backup data encryption.

**Create Secret:**

Let's create a secret called `encrypt-secret` with the Restic password,

```bash
$ echo -n 'changeit' > RESTIC_PASSWORD
$ kubectl create secret generic -n demo encrypt-secret \
    --from-file=./RESTIC_PASSWORD
secret "encrypt-secret" created
```

**Create BackupConfiguration:**

Below is the YAML for `BackupConfiguration` CR to take application-level backup of the `sample-mysql` database that we have deployed earlier,

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: sample-mysql-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: MySQL
    namespace: demo
    name: sample-mysql
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
        - name: gcs-mysql-repo
          backend: gcs-backend
          directory: /mysql
          encryptionSecret:
            name: encrypt-secret
            namespace: demo
      addon:
        name: mysql-addon
        tasks:
          - name: manifest-backup
          - name: logical-backup
```

- `.spec.sessions[*].schedule` specifies that we want to backup at `5 minutes` interval.
- `.spec.target` refers to the targeted `sample-mysql` MySQL database that we created earlier.
- `.spec.sessions[*].addon.tasks[*].name[*]` specifies that both the `manifest-backup` and `logical-backup` tasks will be executed.

Let's create the `BackupConfiguration` CR that we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/backup/kubestash/application-level/examples/backupconfiguration.yaml
backupconfiguration.core.kubestash.com/sample-mysql-backup created
```

**Verify Backup Setup Successful**

If everything goes well, the phase of the `BackupConfiguration` should be `Ready`. The `Ready` phase indicates that the backup setup is successful. Let's verify the `Phase` of the BackupConfiguration,

```bash
$ kubectl get backupconfiguration -n demo
NAME                  PHASE   PAUSED   AGE
sample-mysql-backup   Ready            2m50s
```

Additionally, we can verify that the `Repository` specified in the `BackupConfiguration` has been created using the following command,

```bash
$ kubectl get repo -n demo
NAME               INTEGRITY   SNAPSHOT-COUNT   SIZE     PHASE   LAST-SUCCESSFUL-BACKUP   AGE
gcs-mysql-repo                 0                0 B      Ready                            3m
```

KubeStash keeps the backup for `Repository` YAMLs. If we navigate to the GCS bucket, we will see the `Repository` YAML stored in the `demo/mysql` directory.

**Verify CronJob:**

It will also create a `CronJob` with the schedule specified in `spec.sessions[*].scheduler.schedule` field of `BackupConfiguration` CR.

Verify that the `CronJob` has been created using the following command,

```bash
$ kubectl get cronjob -n demo
NAME                                          SCHEDULE      SUSPEND   ACTIVE   LAST SCHEDULE   AGE
trigger-sample-mysql-backup-frequent-backup   */5 * * * *             0        2m45s           3m25s
```

**Verify BackupSession:**

KubeStash triggers an instant backup as soon as the `BackupConfiguration` is ready. After that, backups are scheduled according to the specified schedule.

Run the following command to watch `BackupSession` CR,

```bash
$ kubectl get backupsession -n demo -w

NAME                                             INVOKER-TYPE          INVOKER-NAME           PHASE       DURATION   AGE
sample-mysql-backup-frequent-backup-1724065200   BackupConfiguration   sample-mysql-backup    Succeeded              7m22s
```

We can see from the above output that the backup session has succeeded. Now, we are going to verify whether the backed up data has been stored in the backend.

**Verify Backup:**

Once a backup is complete, KubeStash will update the respective `Repository` CR to reflect the backup. Check that the repository `sample-mysql-backup` has been updated by the following command,

```bash
$ kubectl get repository -n demo gcs-mysql-repo
NAME                    INTEGRITY   SNAPSHOT-COUNT   SIZE    PHASE   LAST-SUCCESSFUL-BACKUP   AGE
gcs-mysql-repo          true        1                806 B   Ready   8m27s                    9m18s
```

At this moment we have one `Snapshot`. Run the following command to check the respective `Snapshot` which represents the state of a backup run for an application.

```bash
$ kubectl get snapshots -n demo -l=kubestash.com/repo-name=gcs-mysql-repo
NAME                                                            REPOSITORY            SESSION           SNAPSHOT-TIME          DELETION-POLICY   PHASE       AGE
gcs-mysql-repo-sample-mysql-backup-frequent-backup-1725359100   gcs-mysql-repo        frequent-backup   2024-01-23T13:10:54Z   Delete            Succeeded   16h
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
$ kubectl get snapshots -n demo gcs-mysql-repo-sample-mysql-backup-frequent-backup-1725359100 -oyaml
```

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: Snapshot
metadata:
  creationTimestamp: "2024-09-03T10:25:00Z"
  finalizers:
  - kubestash.com/cleanup
  generation: 1
  labels:
    kubestash.com/app-ref-kind: MySQL
    kubestash.com/app-ref-name: sample-mysql
    kubestash.com/app-ref-namespace: demo
    kubestash.com/repo-name: gcs-mysql-repo
  annotations:
    kubedb.com/db-version: 8.2.0
  name: gcs-mysql-repo-sample-mysql-backup-frequent-backup-1725359100
  namespace: demo
  ownerReferences:
  - apiVersion: storage.kubestash.com/v1alpha1
    blockOwnerDeletion: true
    controller: true
    kind: Repository
    name: gcs-mysql-repo
    uid: 1f5ba355-7f99-4b99-8bbf-9f9d4f31c52a
  resourceVersion: "213010"
  uid: 18cabb10-e594-4655-8763-3daa0872508e
spec:
  appRef:
    apiGroup: kubedb.com
    kind: MySQL
    name: sample-mysql
    namespace: demo
  backupSession: sample-mysql-backup-frequent-backup-1725359100
  deletionPolicy: Delete
  repository: gcs-mysql-repo
  session: frequent-backup
  snapshotID: 01J6VPN4TPHDFT1M9Q9YVGMTKF
  type: FullBackup
  version: v1
status:
  components:
    dump:
      driver: Restic
      duration: 7.393324414s
      integrity: true
      path: repository/v1/frequent-backup/dump
      phase: Succeeded
      resticStats:
      - hostPath: dumpfile.sql
        id: f2ffd1bdb98563e15c46d8927d7239873ce7094132d959e12134688e06984736
        size: 3.657 MiB
        uploaded: 706.081 KiB
      size: 893.009 KiB
    manifest:
      driver: Restic
      duration: 12.672292995s
      integrity: true
      path: repository/v1/frequent-backup/manifest
      phase: Succeeded
      resticStats:
      - hostPath: /kubestash-tmp/manifest
        id: ff99eb7ea769a365f7cdc83a252df610c262fc934ec0a3475499bbbb35ca6931
        size: 2.883 KiB
        uploaded: 1.440 KiB
      size: 3.788 KiB
  conditions:
  - lastTransitionTime: "2024-09-03T10:25:00Z"
    message: Recent snapshot list updated successfully
    reason: SuccessfullyUpdatedRecentSnapshotList
    status: "True"
    type: RecentSnapshotListUpdated
  - lastTransitionTime: "2024-09-03T10:25:49Z"
    message: Metadata uploaded to backend successfully
    reason: SuccessfullyUploadedSnapshotMetadata
    status: "True"
    type: SnapshotMetadataUploaded
  integrity: true
  phase: Succeeded
  size: 896.796 KiB
  snapshotTime: "2024-09-03T10:25:00Z"
  totalComponents: 2
```

> KubeStash uses the `mysqldump` command to take backups of target MySQL databases. Therefore, the component name for logical backups is set as `dump`.
> KubeStash set component name as `manifest` for the `manifest backup` of MySQL databases.

Now, if we navigate to the GCS bucket, we will see the backed up data stored in the `demo/mysql/repository/v1/frequent-backup/dump` directory. KubeStash also keeps the backup for `Snapshot` YAMLs, which can be found in the `demo/dep/snapshots` directory.

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
  name: restore-sample-mysql
  namespace: demo
spec:
  manifestOptions:
    mySQL:
      db: true
      restoreNamespace: dev
  dataSource:
    repository: gcs-mysql-repo
    snapshot: latest
    encryptionSecret:
      name: encrypt-secret
      namespace: demo
  addon:
    name: mysql-addon
    tasks:
      - name: logical-backup-restore
      - name: manifest-restore
```

Here,

- `.spec.manifestOptions.mySQL.db` specifies whether to restore the DB manifest or not.
- `.spec.dataSource.repository` specifies the Repository object that holds the backed up data.
- `.spec.dataSource.snapshot` specifies to restore from latest `Snapshot`.
- `.spec.addon.tasks[*]` specifies that both the `manifest-restore` and `logical-backup-restore` tasks.

Let's create the RestoreSession CRD object we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/backup/kubestash/application-level/examples/restoresession.yaml
restoresession.core.kubestash.com/sample-mysql-restore created
```

Once, you have created the `RestoreSession` object, KubeStash will create restore Job. Run the following command to watch the phase of the `RestoreSession` object,

```bash
$ watch kubectl get restoresession -n demo
Every 2.0s: kubectl get restores... AppsCode-PC-03: Wed Aug 21 10:44:05 2024

NAME             REPOSITORY        FAILURE-POLICY   PHASE       DURATION   AGE
sample-restore   gcs-demo-repo                      Succeeded   3s         53s
```
The `Succeeded` phase means that the restore process has been completed successfully.

#### Verify Restored MySQL Manifest:

In this section, we will verify whether the desired `MySQL` database manifest has been successfully applied to the cluster.

```bash
$ kubectl get mysqls.kubedb.com -n dev
NAME           VERSION   STATUS   AGE
sample-mysql   8.2.0     Ready    39m
```

The output confirms that the `MySQL` database has been successfully created with the same configuration as it had at the time of backup.

#### Verify Restored Data:

In this section, we are going to verify whether the desired data has been restored successfully. We are going to connect to the database server and check whether the database and the table we created earlier in the original database are restored.

At first, check if the database has gone into `Ready` state by the following command,

```bash
$ kubectl get my -n dev sample-mysql
NAME             VERSION   STATUS  AGE
sample-mysql     8.2.0     Ready   4m
```

Now, find out the database `Pod` by the following command,

```bash
$ kubectl get pods -n dev --selector="app.kubernetes.io/instance=sample-mysql"
NAME             READY   STATUS    RESTARTS   AGE
sample-mysql-0   2/2     Running   0          2m
sample-mysql-1   2/2     Running   0          2m
sample-mysql-2   2/2     Running   0          2m
```

And then copy the username and password of the `root` user to access into `mysql` shell.

```bash
$ kubectl get secret -n dev  sample-mysql-auth -o jsonpath='{.data.username}'| base64 -d
root

$ kubectl get secret -n dev  sample-mysql-auth -o jsonpath='{.data.password}'| base64 -d
QMm1hi0T*7QFz_yh
```

```bash
$ kubectl exec -it -n dev sample-mysql-0 -- mysql --user=root --password='QMm1hi0T*7QFz_yh'
Defaulted container "mysql" out of: mysql, mysql-coordinator, mysql-init (init)
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 243
Server version: 8.2.0 MySQL Community Server - GPL

Copyright (c) 2000, 2023, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> SHOW DATABASES;
+--------------------+
| Database           |
+--------------------+
| information_schema |
| mysql              |
| performance_schema |
| playground         |
| sys                |
+--------------------+
5 rows in set (0.00 sec)

mysql> SHOW TABLES IN playground;
+----------------------+
| Tables_in_playground |
+----------------------+
| equipment            |
+----------------------+
1 row in set (0.00 sec)

mysql> SELECT * FROM playground.equipment;
+----+-------+-------+-------+
| id | type  | quant | color |
+----+-------+-------+-------+
|  1 | slide |     2 | blue  |
+----+-------+-------+-------+
1 row in set (0.00 sec)

mysql> exit
Bye
```

So, from the above output, we can see that the `playground` database and the `equipment` table we have created earlier in the original database and now, they are restored successfully.

## Cleanup

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete backupconfigurations.core.kubestash.com  -n demo sample-mysql-backup
kubectl delete backupstorage -n demo gcs-storage
kubectl delete secret -n demo gcs-secret
kubectl delete secret -n demo encrypt-secret
kubectl delete retentionpolicies.storage.kubestash.com -n demo demo-retention
kubectl delete restoresessions.core.kubestash.com -n demo restore-sample-mysql
kubectl delete my -n demo sample-mysql
kubectl delete my -n dev sample-mysql
```