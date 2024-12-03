---
title: Continuous Archiving and Point-in-time Recovery
menu:
  docs_{{ .version }}:
    identifier: volumesnapshot
    name: VolumeSnapshot
    parent: pitr-mysql
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# KubeDB MySQL - Continuous Archiving and Point-in-time Recovery using VolumeSnapshot

Here, will show you how to use KubeDB to provision a MySQL to Archive continuously and Restore point-in-time.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now,install `KubeDB` operator in your cluster following the steps [here](/docs/setup/README.md).

To install `KubeStash` operator in your cluster following the steps [here](https://github.com/kubestash/installer/tree/master/charts/kubestash).

To install `External-snapshotter`  in your cluster following the steps [here](https://github.com/kubernetes-csi/external-snapshotter/tree/release-5.0).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```
> Note: The yaml files used in this tutorial are stored in [docs/guides/mysql/pitr/volumesnapshot/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/mysql/pitr/volumesnapshot/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Continuous Archiving
Continuous archiving involves making regular copies (or "archives") of the MySQL transaction log files.To ensure continuous archiving to a remote location we need prepare `BackupStorage`,`RetentionPolicy`,`MySQLArchiver` for the KubeDB Managed MySQL Databases.

### BackupStorage
BackupStorage is a CR provided by KubeStash that can manage storage from various providers like GCS, S3, and more.

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: BackupStorage
metadata:
  name: storage
  namespace: demo
spec:
  storage:
    provider: s3
    s3:
      endpoint: s3.amazonaws.com
      bucket: mysql-archiver
      region: us-east-1
      prefix: my-demo
      secretName: s3-secret
  usagePolicy:
    allowedNamespaces:
      from: All
  deletionPolicy: WipeOut
```

Note: Before applying this yaml, verify that a bucket named `mysql-archiver` is already created on your bucket provider.

```bash
   $ kubectl apply -f backupstorage.yaml
   backupstorage.storage.kubestash.com/storage created
```

### secrets for backup-storage
```yaml
apiVersion: v1
kind: Secret
type: Opaque
metadata:
  name: s3-secret
  namespace: demo
stringData:
  AWS_ACCESS_KEY_ID: "*************26CX"
  AWS_SECRET_ACCESS_KEY: "************jj3lp"
  AWS_ENDPOINT: s3.amazonaws.com
```

```bash
  $ kubectl apply -f storage-secret.yaml 
  secret/s3-secret created
```

### Retention policy
RetentionPolicy is a CR provided by KubeStash that allows you to set how long you'd like to retain the backup data.

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: RetentionPolicy
metadata:
  name: mysql-retention-policy
  namespace: demo
spec:
  maxRetentionPeriod: "30d"
  successfulSnapshots:
    last: 10
  failedSnapshots:
    last: 2
```
```bash
$ kubectl apply -f  https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/pitr/yamls/retention-policy.yaml 
retentionpolicy.storage.kubestash.com/mysql-retention-policy created
```

### MySQLArchiver
MySQLArchiver is a CR provided by KubeDB for managing the archiving of MySQL binlog files and performing volume-level backups

```yaml
apiVersion: archiver.kubedb.com/v1alpha1
kind: MySQLArchiver
metadata:
  name: mysqlarchiver-sample
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
    name: mysql-retention-policy
    namespace: demo
  encryptionSecret:
    name: "encrypt-secret"
    namespace: "demo"
  fullBackup:
    driver: "VolumeSnapshotter"
    task:
      params:
        volumeSnapshotClassName: "longhorn-snapshot-vsc"
    scheduler:
      successfulJobsHistoryLimit: 1
      failedJobsHistoryLimit: 1
      schedule: "*/30 * * * *"
    sessionHistoryLimit: 2
  manifestBackup:
    scheduler:
      successfulJobsHistoryLimit: 1
      failedJobsHistoryLimit: 1
      schedule: "*/30 * * * *"
    sessionHistoryLimit: 2
  backupStorage:
    ref:
      name: "storage"
      namespace: "demo"

```

### EncryptionSecret

```yaml
apiVersion: v1
kind: Secret
type: Opaque
metadata:
  name: encrypt-secret
  namespace: demo
stringData:
  RESTIC_PASSWORD: "changeit"
```

```bash 
 $ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/pitr/volumesnapshot/yamls/mysqlarchiver.yaml
 mysqlarchiver.archiver.kubedb.com/mysqlarchiver-sample created
 $ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/pitr/volumesnapshot/yamls/encryptionSecret.yaml
 secret/encrypt-secret created
```

## Ensure VolumeSnapshotClass

```bash
$ kubectl get volumesnapshotclasses
NAME                    DRIVER               DELETIONPOLICY   AGE
longhorn-snapshot-vsc   driver.longhorn.io   Delete           7d22h

```
If not any, try using `longhorn` or any other [volumeSnapshotClass](https://kubernetes.io/docs/concepts/storage/volume-snapshot-classes/).
```yaml
kind: VolumeSnapshotClass
apiVersion: snapshot.storage.k8s.io/v1
metadata:
  name: longhorn-snapshot-vsc
driver: driver.longhorn.io
deletionPolicy: Delete
parameters:
  type: snap

```

```bash
$ helm install longhorn longhorn/longhorn --namespace longhorn-system --create-namespace

$ kubectl apply -f volumesnapshotclass.yaml
  volumesnapshotclass.snapshot.storage.k8s.io/longhorn-snapshot-vsc unchanged
```

Note: Ensure that the VolumeSnapshotClass is provisioned with the same storage class driver used for provisioning your MySQL database. In our case, we are using the `longhorn` storageclass as our database provisioner, with the driver set to `driver.longhorn.io`.

# Deploy MySQL
We are now ready with the setup for continuous MySQL archiving. We will deploy a MySQL object that references the MySQL archiver object.

```yaml
apiVersion: kubedb.com/v1
kind: MySQL
metadata:
  name: mysql
  namespace: demo
  labels:
    archiver: "true"
spec:
  version: "8.2.0"
  replicas: 3
  topology:
    mode: GroupReplication
  storageType: Durable
  storage:
    storageClassName: "longhorn"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 10Gi
  archiver:
    ref:
      name: mysqlarchiver-sample
      namespace: demo
  deletionPolicy: WipeOut
```


```bash
$ kubectl get pod -n demo
NAME                                                              READY   STATUS      RESTARTS        AGE
mysql-0                                                           2/2     Running     0               28h
mysql-1                                                           2/2     Running     0               28h
mysql-2                                                           2/2     Running     0               28h
mysql-archiver-full-backup-1733206003-hq4pb                       0/1     Completed   0               28h
mysql-archiver-manifest-backup-1733206003-q78jj                   0/1     Completed   0               28h
mysql-sidekick                                                    1/1     Running     0               28h
retention-policy-mysql-archiver-full-backup-1733206003-b2b42      0/1     Completed   0               28h
retention-policy-mysql-archiver-manifest-backup-1733206003skwqc   0/1     Completed   0               28h

```

`mysql-sidekick` is responsible for uploading binlog files

`mysql-backup-config-full-backup-1703680982-vqf7c` are the pod of volumes levels backups for MySQL.

`mysql-backup-config-manifest-1703680982-62x97` are the pod of the manifest backup related to MySQL object

`retention-policy-mysql-archiver-full-backup-1733206003-b2b42` will automatically clean up previous full-backup of volumesnapshots according to the rules defined in the `mysql-retention-policy` custom resource (CR).

`retention-policy-mysql-archiver-manifest-backup-1733206003skwqc` will automatically clean up previous manifest-backup snapshots according to the rules specified in the `mysql-retention-policy` custom resource (CR).



### Validate BackupConfiguration and VolumeSnapshots

```bash

$ kubectl get backupconfigurations -n demo

NAME                    PHASE   PAUSED   AGE
mysql-archiver          Ready            2m43s

$ kubectl get backupsession -n demo
NAME                                           INVOKER-TYPE          INVOKER-NAME          PHASE       DURATION   AGE
mysql-archiver-full-backup-1733206003          BackupConfiguration   mysql-backup-config   Succeeded              74s
mysql-archiver-manifest-backup-1733206003      BackupConfiguration   mysql-backup-config   Succeeded              74s

kubectl get volumesnapshots -n demo
NAME                           READYTOUSE   SOURCEPVC                  SOURCESNAPSHOTCONTENT   RESTORESIZE   SNAPSHOTCLASS           SNAPSHOTCONTENT                                    CREATIONTIME   AGE
mysql-1702388096               true         data-mysql-1                                       1Gi           longhorn-snapshot-vsc   snapcontent-735e97ad-1dfa-4b70-b416-33f7270d792c   2m5s           2m5s

$ kubectl get repository.storage.kubestash.com -n demo 
NAME             INTEGRITY   SNAPSHOT-COUNT   SIZE        PHASE   LAST-SUCCESSFUL-BACKUP   AGE
mysql-full       true        1                2.073 KiB   Ready   2m43s                    2m43s
mysql-manifest   true        1                2.073 KiB   Ready   2m43s                    2m43s
```

## Data Insert and Switch Binlog File
After each and every binlog switch the binlog files will be uploaded to backup storage

```bash
$ kubectl exec -it -n demo  mysql-0 -- bash

bash-4.4$ mysql -uroot -p$MYSQL_ROOT_PASSWORD

mysql> create database hello;

mysql> use hello;

mysql> CREATE TABLE `demo_table`(
    ->     `id` BIGINT(20) NOT NULL,
    ->     `name` VARCHAR(255) DEFAULT NULL,
    ->     PRIMARY KEY (`id`)
    -> );

mysql> INSERT INTO `demo_table` (`id`, `name`)
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

mysql> select now();
+---------------------+
| now()               |
+---------------------+
| 2024-12-03 06:09:34 |
+---------------------+
+---------------------+

mysql> select count(*) from demo_table;
+----------+
| count(*) |
+----------+
|       10 |
+----------+

```

> At this point We have 10 rows in our newly created table `demo_table` on database `hello`

## Point-in-time Recovery
Point-In-Time Recovery allows you to restore a MySQL database to a specific point in time using the archived transaction logs. This is particularly useful in scenarios where you need to recover to a state just before a specific error or data corruption occurred.
Let's say accidentally our db drops the table `demo_table` and we want to restore that.

```bash
$ kubectl exec -it -n demo  mysql-0 -- bash

mysql> drop table demo_table;

mysql> flush logs;

```
We can't restore from a full backup since at this point no full backup was perform. so we can choose a specific time in which time we want to restore.We can get the specfice time from the wal that archived in the backup storage . Go to the binlog file and find where to store. You can parse binlog-files using  `mysqlbinlog`.


For the demo I will use the previous time we get form `select now()`

```bash 
mysql> select now();
+---------------------+
| now()               |
+---------------------+
| 2024-12-03 06:09:34 |
+---------------------+
```
### Restore MySQL

```yaml
apiVersion: kubedb.com/v1
kind: MySQL
metadata:
  name: restore-mysql
  namespace: demo
spec:
  init:
    archiver:
      replicationStrategy: sync
      encryptionSecret:
        name: encrypt-secret
        namespace: demo
      fullDBRepository:
        name: mysql-full
        namespace: demo
      recoveryTimestamp: "2024-12-03T06:09:34Z"
  version: "8.2.0"
  replicas: 3
  topology:
    mode: GroupReplication
  storageType: Durable
  storage:
    storageClassName: "longhorn"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut

```

```bash
$ kubectl apply -f restore.yaml
mysql.kubedb.com/restore-mysql created
```

**check for Restored MySQL**

```bash
$ kubectl get pod -n demo
restore-mysql-0                                          1/1     Running     0             44s
restore-mysql-1                                          1/1     Running     0             42s
restore-mysql-2                                          1/1     Running     0             41s
restore-mysql-restorer-z4brz                             0/2     Completed   0             113s
restore-mysql-restoresession-lk6jq                       0/1     Completed   0             2m6s

```

```bash
$ kubectl get mysql -n demo
NAME            VERSION   STATUS   AGE
mysql           8.2.0     Ready    28h
restore-mysql   8.2.0     Ready    5m37s
```

**Validating data on Restored MySQL**

```bash
$ kubectl exec -it -n demo restore-mysql-0 -- bash
bash-4.4$ mysql -uroot -p$MYSQL_ROOT_PASSWORD

mysql> use hello

mysql> select count(*) from demo_table;
+----------+
| count(*) |
+----------+
|       10 |
+----------+
1 row in set (0.00 sec)

```

**so we are able to successfully recover from a disaster**

**ReplicationStrategy**

The ReplicationStrategy determines how MySQL restores are managed when using the VolumeSnapshot. We support three strategies: `none`, `sync`, and `fscopy`, with `none` being the default.

To configure the desired strategy, set the `spec.init.archiver.replicationStrategy` field in your configuration. These strategies are applicable only when restoring a MySQL database in group replication mode.

**Strategies Overview:**

***none***

Each MySQL replica independently restores the base backup volumesnapshot and binlog files. After completing the restore process, the replicas individually join the replication group.

***sync***

The base backup volumesnapshot and binlog files are restored exclusively on pod-0. Other replicas then synchronize their data by leveraging the MySQL clone plugin to replicate from pod-0.

***fscopy***

The base backup and binlog files are restored on pod-0. The data is then copied from pod-0's data directory to the data directories of other replicas using file system copy. Once the data transfer is complete, the group replication process begins.  Please note that `fscopy` does not support cross-zone operations.

***clone***

If you have a different type of base backup(ex: VolumeSnapshot, Restic), the clone process will ensure that the VolumeSnapshot is restored as the base backup. Each MySQL replica independently restores the base backup volumesnapshot and binlog files. After completing the restore process, the replicas individually join the replication group. 


Choose the replication strategy that best fits your restoration and replication requirements. On this demonstration, we have used the sync replication strategy.


## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete -n demo mysql/mysql
$ kubectl delete -n demo mysql/restore-mysql
$ kubectl delete -n demo backupstorage/storage
$ kubectl delete -n demo mysqlarchiver/mysqlarchiver-sample
$ kubectl delete ns demo
```

## Next Steps

- Learn about [backup and restore](/docs/guides/mysql/backup/stash/overview/index.md) MySQL database using Stash.
- Learn about initializing [MySQL with Script](/docs/guides/mysql/initialization/script_source.md).
- Learn about [custom MySQLVersions](/docs/guides/mysql/custom-versions/setup.md).
- Want to setup MySQL cluster? Check how to [configure Highly Available MySQL Cluster](/docs/guides/mysql/clustering/ha_cluster.md)
- Monitor your MySQL database with KubeDB using [built-in Prometheus](/docs/guides/mysql/monitoring/using-builtin-prometheus.md).
- Monitor your MySQL database with KubeDB using [Prometheus operator](/docs/guides/mysql/monitoring/using-prometheus-operator.md).
- Detail concepts of [MySQL object](/docs/guides/mysql/concepts/mysql.md).
- Use [private Docker registry](/docs/guides/mysql/private-registry/using-private-registry.md) to deploy MySQL with KubeDB.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).