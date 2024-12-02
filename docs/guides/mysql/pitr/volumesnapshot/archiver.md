---
title: Continuous Archiving and Point-in-time Recovery
menu:
  docs_{{ .version }}:
    identifier: pitr-mysql-archiver
    name: VolumeSnapShot
    parent: pitr-mysql
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# KubeDB MySQL - Continuous Archiving and Point-in-time Recovery

Here, will show you how to use KubeDB to provision a MySQL to Archive continuously and Restore point-in-time.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now,install `KubeDB` operator in your cluster following the steps [here](/docs/setup/README.md).

To install `KubeStash` operator in your cluster following the steps [here](https://github.com/kubestash/installer/tree/master/charts/kubestash).

To install `SideKick`  in your cluster following the steps [here](https://github.com/kubeops/installer/tree/master/charts/sidekick).

To install `External-snapshotter`  in your cluster following the steps [here](https://github.com/kubernetes-csi/external-snapshotter/tree/release-5.0).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```
> Note: The yaml files used in this tutorial are stored in [docs/guides/mysql/remote-replica/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/mysql/remote-replica/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## continuous archiving
Continuous archiving involves making regular copies (or "archives") of the MySQL transaction log files.To ensure continuous archiving to a remote location we need prepare `BackupStorage`,`RetentionPolicy`,`MySQLArchiver` for the KubeDB Managed MySQL Databases.

### BackupStorage
BackupStorage is a CR provided by KubeStash that can manage storage from various providers like GCS, S3, and more.

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: BackupStorage
metadata:
  name: linode-storage
  namespace: demo
spec:
  storage:
    provider: s3
    s3:
      bucket: mehedi-mysql-wal-g
      endpoint: https://ap-south-1.linodeobjects.com
      region: ap-south-1
      prefix: backup
      secretName: storage
  usagePolicy:
    allowedNamespaces:
      from: All
  default: true
  deletionPolicy: WipeOut
```

```bash
   $ kubectl apply -f backupstorage.yaml
   backupstorage.storage.kubestash.com/linode-storage created
```

### secrets for backup-storage
```yaml
apiVersion: v1
kind: Secret
type: Opaque
metadata:
  name: storage
  namespace: demo
stringData:
  AWS_ACCESS_KEY_ID: "*************26CX"
  AWS_SECRET_ACCESS_KEY: "************jj3lp"
  AWS_ENDPOINT: https://ap-south-1.linodeobjects.com
```

```bash
  $ kubectl apply -f storage-secret.yaml 
  secret/storage created
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
    last: 100
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
      name: "linode-storage"
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
 $ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/pirt/yamls/mysqlarchiver.yaml
 mysqlarchiver.archiver.kubedb.com/mysqlarchiver-sample created
 $ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/pirt/yamls/encryptionSecret.yaml
```

## Ensure volumeSnapshotClass

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

# Deploy MySQL
So far we are ready with setup for continuously archive MySQL, We deploy a mysqlql referring the MySQL archiver object

```yaml
apiVersion: kubedb.com/v1
kind: MySQL
metadata:
  name: mysql
  namespace: demo
  labels:
    archiver: "true"
spec:
  authSecret:
    name: my-auth
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
NAME                                                     READY   STATUS      RESTARTS        AGE
mysql-0                                                  2/2     Running     0               28h
mysql-1                                                  2/2     Running     0               28h
mysql-2                                                  2/2     Running     0               28h
mysql-backup-config-full-backup-1703680982-vqf7c         0/1     Completed   0               28h
mysql-backup-config-manifest-1703680982-62x97            0/1     Completed   0               28h
mysql-sidekick                                           1/1     Running     0               28h
```

`mysql-sidekick` is responsible for uploading binlog files

`mysql-backup-config-full-backup-1703680982-vqf7c` are the pod of volumes levels backups for MySQL.

`mysql-backup-config-manifest-1703680982-62x97` are the pod of the manifest backup related to MySQL object

### validate BackupConfiguration and VolumeSnapshots

```bash

$ kubectl get backupconfigurations -n demo

NAME                    PHASE   PAUSED   AGE
mysql-backup-config   Ready            2m43s

$ kubectl get backupsession -n demo
NAME                                           INVOKER-TYPE          INVOKER-NAME            PHASE       DURATION   AGE
mysql-backup-config-full-backup-1702388088   BackupConfiguration   mysql-backup-config   Succeeded              74s
mysql-backup-config-manifest-1702388088      BackupConfiguration   mysql-backup-config   Succeeded              74s

kubectl get volumesnapshots -n demo
NAME                           READYTOUSE   SOURCEPVC                  SOURCESNAPSHOTCONTENT   RESTORESIZE   SNAPSHOTCLASS           SNAPSHOTCONTENT                                    CREATIONTIME   AGE
mysql-1702388096               true         data-mysql-1                                       1Gi           longhorn-snapshot-vsc   snapcontent-735e97ad-1dfa-4b70-b416-33f7270d792c   2m5s           2m5s
```

## data insert and switch wal
After each and every wal switch the wal files will be uploaded to backup storage

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
| 2023-12-28 17:10:54 |mysql> select now();
+---------------------+
| now()               |
+---------------------+
| 2023-12-28 17:10:54 |
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
Let's say accidentally our dba drops the the table tab_1 and we want to restore.

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
| 2023-12-28 17:10:54 |
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
      encryptionSecret:
        name: encrypt-secret
        namespace: demo
      fullDBRepository:
        name: mysql-repository
        namespace: demo
      recoveryTimestamp: "2023-12-28T17:10:54Z"
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
mysql           8.2.0    Ready    28h
restore-mysql   8.2.0    Ready    5m37s
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

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete -n demo mysql/mysql
$ kubectl delete -n demo mysql/restore-mysql
$ kubectl delete -n demo backupstorage
$ kubectl delete -n demo mysqlarchiver
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