---
title: Continuous Archiving and Point-in-time Recovery Using Restic Driver
menu:
  docs_{{ .version }}:
    identifier: restic
    name: Restic
    parent: pitr-mysql
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# KubeDB MySQL - Continuous Archiving and Point-in-time Recovery

Here, we will demonstrate how to use KubeDB to provision a MySQL database for continuous archiving and point-in-time restoration
This process utilizes [Percona XtraBackup](https://www.percona.com/mysql/software/percona-xtrabackup), a robust tool for taking physical backups of MySQL databases, ensuring data integrity and consistency. We use XtraBackup as the base backup for our MySQL archiver.
Let's explore how we use XtraBackup in MySQL database archiving.
## Before You Begin

To get started with archiving MySQL using Percona XtraBackup, you’ll need a Kubernetes cluster with the kubectl command-line tool configured to interact with it. If you don’t already have a cluster, you can easily create one using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Next, install the `KubeDB` operator in your cluster by following the steps outlined [here](/docs/setup/README.md).

To install the `KubeStash` operator in your cluster, follow the steps outlined [here](https://github.com/kubestash/installer/tree/master/charts/kubestash).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```
> Note: The yaml files used in this tutorial are stored in [docs/guides/mysql/pitr/restic/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/mysql/pitr/restic/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Continuous Archiving
Continuous archiving involves making regular copies (or "archives") of the MySQL transaction log files.To ensure continuous archiving to a remote location we need prepare `BackupStorage`,`RetentionPolicy`,`MySQLArchiver` for the KubeDB Managed MySQL Databases.

### BackupStorage
BackupStorage is a CR provided by KubeStash that can manage storage from various providers like GCS, S3, and more.
Here we are using AWS s3 bucket.
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
      bucket: mysql-xtrabackup
      region: us-east-1
      prefix: my-demo
      secretName: s3-secret
  usagePolicy:
    allowedNamespaces:
      from: All
  deletionPolicy: WipeOut
```

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
    driver: "Restic"
    scheduler:
      successfulJobsHistoryLimit: 1
      failedJobsHistoryLimit: 1
      schedule: "0 0 * * *"
    sessionHistoryLimit: 2
  manifestBackup:
    scheduler:
      successfulJobsHistoryLimit: 1
      failedJobsHistoryLimit: 1
      schedule: "0 0 * * *"
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
 $ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/pirt/yamls/mysqlarchiver.yaml
 mysqlarchiver.archiver.kubedb.com/mysqlarchiver-sample created
 $ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/pirt/yamls/encryptionSecret.yaml
```

# Deploy MySQL
So far we are ready with setup for continuously archive MySQL, We deploy a mysql object referring the MySQL archiver object

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
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  archiver:
    ref:
      name: mysqlarchiver-sample
      namespace: demo
  deletionPolicy: WipeOut
```

```bash
$ kubectl get pod -n demo
NAME                                                              READY   STATUS      RESTARTS   AGE
mysql-0                                                           2/2     Running     0          15m
mysql-1                                                           2/2     Running     0          15m
mysql-2                                                           2/2     Running     0          15m
mysql-archiver-full-backup-1733120326-4n8nd                       0/1     Completed   0          10m
mysql-archiver-manifest-backup-1733120326-9gw4f                   0/1     Completed   0          10m
mysql-sidekick                                                    1/1     Running     0          10m
retention-policy-mysql-archiver-full-backup-1733120326-7rx9t      0/1     Completed   0          9m31s
retention-policy-mysql-archiver-manifest-backup-1733120326l79mb   0/1     Completed   0          9m56s    28h
```

`mysql-sidekick` is responsible for uploading binlog files

The pod mysql-archiver-full-backup-1733120326-4n8nd is responsible for creating the base backup of MySQL. 
`mysql-archiver-manifest-backup-1733120326-9gw4f` is the pod of the manifest backup related to MySQL object.

### Validate BackupConfiguration and BackupSession

```bash

$ kubectl get backupconfigurations -n demo

NAME             PHASE   PAUSED   AGE
mysql-archiver   Ready            14m

$ kubectl get backupsession -n demo
NAME                                        INVOKER-TYPE          INVOKER-NAME     PHASE       DURATION   AGE
mysql-archiver-full-backup-1733120326       BackupConfiguration   mysql-archiver   Succeeded   50s        14m
mysql-archiver-manifest-backup-1733120326   BackupConfiguration   mysql-archiver   Succeeded   25s        14m
                                      1Gi           longhorn-snapshot-vsc   snapcontent-735e97ad-1dfa-4b70-b416-33f7270d792c   2m5s           2m5s
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
| 2024-12-02 06:38:42 |
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
| 2024-12-02 06:38:42 |
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
        name: mysql-repository
        namespace: demo
      recoveryTimestamp: "2024-12-02T06:38:42Z"
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
restore-mysql-0                                                   2/2     Running     0          6m40s
restore-mysql-1                                                   2/2     Running     0          5m37s
restore-mysql-2                                                   2/2     Running     0          5m21s
restore-mysql-binlog-restorer-0                                   0/2     Completed   0          5m58s
restore-mysql-manifest-restorer-pzx5z                             0/1     Completed   0          6m54s
```

```bash
$ kubectl get mysql -n demo
NAME                             VERSION   STATUS   AGE
mysql.kubedb.com/mysql           8.2.0     Ready    32m
mysql.kubedb.com/restore-mysql   8.2.0     Ready    2m53s
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

The ReplicationStrategy determines how MySQL restores are managed when using the Restic driver in a group replication setup. We support three strategies: `none`, `sync`, and `fscopy`, with `none` being the default.
 
To configure the desired strategy, set the spec.init.archiver.replicationStrategy field in your configuration. These strategies are applicable only when restoring a MySQL database in group replication mode.

**Strategies Overview:**

***none***

Each MySQL replica independently restores the base backup and binlog files. After completing the restore process, the replicas individually join the replication group.

***sync***

The base backup and binlog files are restored exclusively on pod-0. Other replicas then synchronize their data by leveraging the MySQL clone plugin to replicate from pod-0. 

***fscopy***

The base backup and binlog files are restored on pod-0. The data is then copied from pod-0's data directory to the data directories of other replicas using file system copy. Once the data transfer is complete, the group replication process begins.

Choose the replication strategy that best fits your restoration and replication requirements.
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