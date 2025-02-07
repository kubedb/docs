---
title: Continuous Archiving and Point-in-time Recovery
menu:
  docs_{{ .version }}:
    identifier: pitr-mariadb-archiver
    name: Overview
    parent: pitr-mariadb
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# KubeDB MariaDB - Continuous Archiving and Point-in-time Recovery

Here, will show you how to use KubeDB to provision a MariaDB to Archive continuously and Restore point-in-time.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. For this demonstration, I'm using linode cluster.

Now,install `KubeDB` operator in your cluster following the steps [here](/docs/setup/README.md).

To install `KubeStash` operator in your cluster following the steps [here](https://github.com/kubestash/installer/tree/master/charts/kubestash).

To install `External-snapshotter`  in your cluster following the steps [here](https://github.com/kubernetes-csi/external-snapshotter/tree/release-5.0).

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```
> Note: The yaml files used in this tutorial are stored in [docs/guides/mariadb/pitr/overview/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/mariadb/remote-replica/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## continuous archiving
Continuous archiving involves making regular copies (or "archives") of the MariaDB transaction log files.To ensure continuous archiving to a remote location we need prepare `BackupStorage`,`RetentionPolicy`,`MariaDBArchiver` for the KubeDB Managed MariaDB Databases.

### BackupStorage
BackupStorage is a CR provided by KubeStash that can manage storage from various providers like GCS, S3, and more.
We are going to store our backup data into a `S3` bucket. We have to create a `Secret` with necessary credentials and a `BackupStorage` CR to use this backend. If you want to use a different backend, please read the respective backend configuration doc from [here](https://kubestash.com/docs/latest/guides/backends/overview/).

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
      bucket: test-archiver
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
RetentionPolicy is a custom resource(CR) provided by KubeStash that allows you to set how long you'd like to retain the backup data.

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: RetentionPolicy
metadata:
  name: mariadb-retention-policy
  namespace: demo
spec:
  maxRetentionPeriod: "30d"
  successfulSnapshots:
    last: 100
  failedSnapshots:
    last: 2
```
```bash
$ kubectl apply -f  https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/pitr/overview/yamls/retention-policy.yaml 
retentionpolicy.storage.kubestash.com/mariadb-retention-policy created
```

### MariaDBArchiver
MariaDBArchiver is a custom resource(CR) provided by KubeDB for managing the archiving of MariaDB binlog files and performing volume-level backups

```yaml
apiVersion: archiver.kubedb.com/v1alpha1
kind: MariaDBArchiver
metadata:
  name: mariadbarchiver-sample
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
    name: mariadb-retention-policy
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
 $ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/pitr/overview/yamls/mariadbarchiver.yaml
 mariadbarchiver.archiver.kubedb.com/mariadbarchiver-sample created
 $ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/pitr/overview/yamls/encryptionSecret.yaml
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

# Deploy MariaDB
So far we are ready with setup for continuously archive MariaDB, We deploy a mariadb referring the MariaDB archiver object.To properly configure MariaDB for archiving, you need to pass specific arguments to the MariaDB container in the `spec.podTemplate.containers["mariadb"].args` field. Below is an example of a YAML configuration for a MariaDB instance managed by KubeDB, with archiving enabled.

```yaml
apiVersion: kubedb.com/v1
kind: MariaDB
metadata:
  name: mariadb
  namespace: demo
  labels:
    archiver: "true"
spec:
  version: "11.1.3"
  replicas: 3
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
      name: mariadbarchiver-sample
      namespace: demo
  deletionPolicy: WipeOut
  podTemplate:
    spec:
      containers:
        - name: mariadb
          args:
            - "--log-bin"
            - "--log-slave-updates"
            - "--wsrep-gtid-mode=ON"
```


```bash
$ kubectl get pod -n demo
NAME                                                              READY   STATUS      RESTARTS        AGE
mariadb-0                                                         2/2     Running     0               4m12s
mariadb-1                                                         2/2     Running     0               4m12s
mariadb-2                                                         2/2     Running     0               3m12s
mariadb-backup-full-backup-1726549703-bjk9w                       0/1     Completed   0               3m22s
mariadb-backup-manifest-backup-1726549703-fx9kx                   0/1     Completed   0               3m22s
mariadb-sidekick                                                  1/1     Running
retention-policy-mariadb-backup-full-backup-1726549703-wg7wt      0/1     Completed   0               3m42s
retention-policy-mariadb-backup-manifest-backup-17265497038pvjd   0/1     Completed   0               3m55s
```

`mariadb-sidekick` is responsible for uploading binlog files

`mariadb-backup-full-backup-1726549703-bjk9w ` are the pod of volumes levels backups for MariaDB.

`mariadb-backup-manifest-backup-1726549703-fx9kx` are the pod of the manifest backup related to MariaDB object

### validate BackupConfiguration and VolumeSnapshots

```bash

$ kubectl get backupconfigurations -n demo

NAME                    PHASE   PAUSED   AGE
mariadb-backup          Ready            2m43s

$ kubectl get backupsession -n demo
NAME                                           INVOKER-TYPE          INVOKER-NAME            PHASE       DURATION   AGE
mariadb-backup-full-backup-1726549703          BackupConfiguration   mariadb-backup          Succeeded   33s        11m
mariadb-backup-manifest-backup-1726549703      BackupConfiguration   mariadb-backup          Succeeded   20s        11m

kubectl get volumesnapshots -n demo
NAME                    READYTOUSE   SOURCEPVC        SOURCESNAPSHOTCONTENT   RESTORESIZE   SNAPSHOTCLASS           SNAPSHOTCONTENT                                    CREATIONTIME   AGE
mariadb-1726549985      true         data-mariadb-0                           10Gi          longhorn-snapshot-vsc   snapcontent-317aaac9-ae4f-438b-9763-4eb81ff828af    11m            11m
```

## Data Insert and Switch Binlog File
After each and every binlog switch the binlog files will be uploaded to backup storage

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

MariaDB [hello]> select now();
+---------------------+
| now()               |
+---------------------+
| 2024-09-17 05:28:26 |
+---------------------+
+---------------------+

MariaDB [hello]> select count(*) from demo_table;
+----------+
| count(*) |
+----------+
|       10 |
+----------+

```

> At this point We have 10 rows in our newly created table `demo_table` on database `hello`

## Point-in-time Recovery
Point-In-Time Recovery allows you to restore a MariaDB database to a specific point in time using the archived transaction logs. This is particularly useful in scenarios where you need to recover to a state just before a specific error or data corruption occurred.
Let's say accidentally our dba drops the table demo_table and we want to restore.

```bash
$ kubectl exec -it -n demo  mariadb-0 -- bash

MariaDB [hello]> drop table demo_table;

MariaDB [hello]> flush logs;

```
We can't restore from a full backup since at this point no full backup was perform. so we can choose a specific time in which time we want to restore.We can get the specfice time from the binlog that archived in the backup storage . Go to the binlog file and find where to store. You can parse binlog-files using  `mariadbbinlog`.


For the demo I will use the previous time we get form `select now()`

```bash 
MariaDB [hello]> select now();
+---------------------+
| now()               |
+---------------------+
| 2024-09-17 05:28:26 |
+---------------------+
```
### Restore MariaDB

```yaml
apiVersion: kubedb.com/v1
kind: MariaDB
metadata:
  name: restore-mariadb
  namespace: demo
spec:
  init:
    archiver:
      encryptionSecret:
        name: encrypt-secret
        namespace: demo
      fullDBRepository:
        name: mariadb-full
        namespace: demo
      recoveryTimestamp: "2024-09-17T05:28:26Z"
  version: "11.1.3"
  replicas: 3
  storageType: Durable
  storage:
    storageClassName: "longhorn"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 10Gi
  deletionPolicy: WipeOut
  podTemplate:
    spec:
      containers:
        - name: mariadb
          args:
            - "--log-bin"
            - "--log-slave-updates"
            - "--wsrep-gtid-mode=ON"
```

```bash
$ kubectl apply -f mariadbrestore.yaml
mariadb.kubedb.com/restore-mariadb created
```

**check for Restored MariaDB**

```bash
$ kubectl get pod -n demo
restore-mariadb-0                                          1/1     Running     0             44s
restore-mariadb-1                                          1/1     Running     0             42s
restore-mariadb-2                                          1/1     Running     0             41s
restore-mariadb-restorer-z4brz                             0/2     Completed   0             113s
restore-mariadb-restoresession-lk6jq                       0/1     Completed   0             2m6s

```

```bash
$ kubectl get mariadb -n demo
NAME              VERSION   STATUS   AGE
mariadb           11.1.3    Ready    14m
restore-mariadb   11.1.3    Ready    5m37s
```

**Validating data on Restored MariaDB**

```bash
$ kubectl exec -it -n demo restore-mariadb-0 -- bash
bash-4.4$ mariadb -uroot -p$MYSQL_ROOT_PASSWORD

mariadb> use hello

MariaDB [hello]> select count(*) from demo_table;
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
$ kubectl delete -n demo mariadb/mariadb
$ kubectl delete -n demo mariadb/restore-mariadb
$ kubectl delete -n demo backupstorage
$ kubectl delete -n demo mariadbarchiver
$ kubectl delete ns demo
```

## Next Steps

- Learn about [backup and restore](/docs/guides/mariadb/backup/kubestash/overview/index.md) MariaDB database using KubeStash.
- Learn about initializing [MariaDB with Script](/docs/guides/mariadb/initialization/using-script/index.md).
- Want to setup MariaDB cluster? Check how to [configure Highly Available MariaDB Cluster](/docs/guides/mariadb/clustering/galera-cluster/index.md)
- Monitor your MariaDB database with KubeDB using [built-in Prometheus](/docs/guides/mariadb/monitoring/builtin-prometheus/index.md).
- Monitor your MariaDB database with KubeDB using [Prometheus operator](/docs/guides/mariadb/monitoring/prometheus-operator/index.md).
- Detail concepts of [MariaDB object](/docs/guides/mariadb/concepts/mariadb/index.md).
- Use [private Docker registry](/docs/guides/mariadb/private-registry/quickstart/index.md) to deploy MariaDB with KubeDB.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).