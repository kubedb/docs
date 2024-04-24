---
title: Continuous Archiving and Point-in-time Recovery
menu:
  docs_{{ .version }}:
    identifier: pitr-postgres-archiver
    name: Overview
    parent: pitr-postgres
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# KubeDB PostgreSQL - Continuous Archiving and Point-in-time Recovery

Here, will show you how to use KubeDB to provision a PostgreSQL to Archive continuously and Restore point-in-time.

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
> Note: The yaml files used in this tutorial are stored in [docs/guides/postgres/remote-replica/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/postgres/remote-replica/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## continuous archiving
Continuous archiving involves making regular copies (or "archives") of the PostgreSQL transaction log files.To ensure continuous archiving to a remote location we need prepare `BackupStorage`,`RetentionPolicy`,`PostgresArchiver` for the KubeDB Managed PostgreSQL Databases.

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
      bucket: mehedi-pg-wal-g
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
  name: postgres-retention-policy
  namespace: demo
spec:
  maxRetentionPeriod: "30d"
  successfulSnapshots:
    last: 100
  failedSnapshots:
    last: 2
```
```bash
$ kubectl apply -f  https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/pitr/yamls/retention-policy.yaml 
retentionpolicy.storage.kubestash.com/postgres-retention-policy created
```

### PostgreSQLArchiver
PostgreSQLArchiver is a CR provided by KubeDB for managing the archiving of MongoDB oplog files and performing volume-level backups

```yaml
apiVersion: archiver.kubedb.com/v1alpha1
kind: PostgresArchiver
metadata:
  name: postgresarchiver-sample
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
    name: postgres-retention-policy
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
      schedule: "/30 * * * *"
    sessionHistoryLimit: 2
  manifestBackup:
    scheduler:
      successfulJobsHistoryLimit: 1
      failedJobsHistoryLimit: 1 
      schedule: "/30 * * * *"
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
 $ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/pirt/yamls/postgresarchiver.yaml
 postgresarchiver.archiver.kubedb.com/postgresarchiver-sample created
 $ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/pirt/yamls/encryptionSecret.yaml
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

# Deploy PostgreSQL
So far we are ready with setup for continuously archive PostgreSQL, We deploy a postgresql referring the PostgreSQL archiver object

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Postgres
metadata:
  name: demo-pg
  namespace: demo
  labels:
    archiver: "true"
spec:
  version: "13.13"
  replicas: 3
  standbyMode: Hot
  storageType: Durable
  storage:
    storageClassName: "longhorn"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  archiver:
    ref:
      name: postgresarchiver-sample
      namespace: demo
  terminationPolicy: WipeOut

```


```bash
$ kubectl get pod -n demo
NAME                                                 READY   STATUS      RESTARTS   AGE
demo-pg-0                                            2/2     Running     0          8m52s
demo-pg-1                                            2/2     Running     0          8m22s
demo-pg-2                                            2/2     Running     0          7m57s
demo-pg-backup-config-full-backup-1702388088-z4qbz   0/1     Completed   0          37s
demo-pg-backup-config-manifest-1702388088-hpx6m      0/1     Completed   0          37s
demo-pg-sidekick                                     1/1     Running     0          7m31s
```

`demo-pg-sidekick` is responsible for uploading wal-files

`demo-pg-backup-config-full-backup-1702388088-z4qbz ` are the pod of volumes levels backups for postgreSQL.

`demo-pg-backup-config-manifest-1702388088-hpx6m ` are the pod of the manifest backup related to PostgreSQL object

### validate BackupConfiguration and VolumeSnapshots

```bash

$ kubectl get backupconfigurations -n demo

NAME                    PHASE   PAUSED   AGE
demo-pg-backup-config   Ready            2m43s

$ kubectl get backupsession -n demo
NAME                                           INVOKER-TYPE          INVOKER-NAME            PHASE       DURATION   AGE
demo-pg-backup-config-full-backup-1702388088   BackupConfiguration   demo-pg-backup-config   Succeeded              74s
demo-pg-backup-config-manifest-1702388088      BackupConfiguration   demo-pg-backup-config   Succeeded              74s

kubectl get volumesnapshots -n demo
NAME                           READYTOUSE   SOURCEPVC                  SOURCESNAPSHOTCONTENT   RESTORESIZE   SNAPSHOTCLASS           SNAPSHOTCONTENT                                    CREATIONTIME   AGE
demo-pg-1702388096             true         data-demo-pg-1                                     1Gi           longhorn-snapshot-vsc   snapcontent-735e97ad-1dfa-4b70-b416-33f7270d792c   2m5s           2m5s
```

## data insert and switch wal
After each and every wal switch the wal files will be uploaded to backup storage

```bash
$ kubectl exec -it -n demo  demo-pg-0 -- bash

bash-5.1$ psql

postgres=# create database hi;
CREATE DATABASE
postgres=# \c hi
hi=# create table tab_1 (a int);
CREATE TABLE
hi=# insert into tab_1 values(generate_series(1,100));
INSERT 0 100
hi=# select pg_switch_wal();
 0/504A0D8
(1 row)

hi=# insert into tab_1 values(generate_series(1,100));
INSERT 0 100

hi=# select now(); 
 2023-12-12 13:43:41.300216+00
 
hi=# select pg_switch_wal();
 0/6013240

hi=# select count(*) from tab_1 ;
   200
```

> At this point We have 200 rows in our newly created table `tab_1` on database `hi`

## Point-in-time Recovery
Point-In-Time Recovery allows you to restore a PostgreSQL database to a specific point in time using the archived transaction logs. This is particularly useful in scenarios where you need to recover to a state just before a specific error or data corruption occurred.
Let's say accidentally our dba drops the the table tab_1 and we want to restore.

```bash
$ kubectl exec -it -n demo  demo-pg-0 -- bash
bash-5.1$ psql
postgres=# \c hi

hi=# drop table tab_1;
DROP TABLE
hi=# select count(*) from tab_1 ;
ERROR:  relation "tab_1" does not exist
LINE 1: select count(*) from tab_1 ;
```
We can't restore from a full backup since at this point no full backup was perform. so we can choose a specific time in which time we want to restore.We can get the specfice time from the wal that archived in the backup storage . Go to the binlog file and find where to store. You can parse wal-files using  `pg-waldump`.


For the demo I will use the previous time we get form `select now()`

```bash 
hi=# select now(); 
 2023-12-12 13:43:41.300216+00
```
### Restore PostgreSQL

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Postgres
metadata:
  name: restore-pg
  namespace: demo
spec:
  init:
    archiver:
      encryptionSecret:
        name: encrypt-secret
        namespace: demo
      fullDBRepository:
        name: demo-pg-repository
        namespace: demo
      manifestRepository:
        name: demo-pg-manifest
        namespace: demo
      recoveryTimestamp: "2023-12-12T13:43:41.300216Z"
  version: "13.13"
  replicas: 3
  standbyMode: Hot
  storageType: Durable
  storage:
    storageClassName: "longhorn"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: WipeOut
```

```bash
$ kubectl apply -f restore.yaml
postgres.kubedb.com/restore-pg created
```

**check for Restored PostgreSQL**

```bash
$ kubectl get pod -n demo
NAME                                                 READY   STATUS      RESTARTS   AGE
restore-pg-0                                         2/2     Running     0          46s
restore-pg-1                                         2/2     Running     0          41s
restore-pg-2                                         2/2     Running     0          22s
restore-pg-restorer-4d4dg                            0/1     Completed   0          104s
restore-pg-restoresession-2tsbv                      0/1     Completed   0          115s
```

```bash
$ kubectl get pg -n demo
NAME         VERSION   STATUS   AGE
demo-pg      13.6      Ready    44m
restore-pg   13.6      Ready    2m36s
```

**Validating data on Restored PostgreSQL**

```bash
$ kubectl exec -it -n demo  restore-pg-0 -- bash
bash-5.1$ psql

postgres=# \c hi

hi=# select count(*) from tab_1 ;
   200
```

**so we are able to successfully recover from a disaster**

## Cleaning up

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete -n demo pg/demo-pg
$ kubectl delete -n demo pg/restore-pg
$ kubectl delete -n demo backupstorage
$ kubectl delete -n demo postgresqlarchiver
$ kubectl delete ns demo
```

## Next Steps

- Learn about [backup and restore](/docs/guides/postgres/backup/overview/index.md) PostgreSQL database using Stash.
- Learn about initializing [PostgreSQL with Script](/docs/guides/postgres/initialization/script_source.md).
- Learn about [custom PostgresVersions](/docs/guides/postgres/custom-versions/setup.md).
- Want to setup PostgreSQL cluster? Check how to [configure Highly Available PostgreSQL Cluster](/docs/guides/postgres/clustering/ha_cluster.md)
- Monitor your PostgreSQL database with KubeDB using [built-in Prometheus](/docs/guides/postgres/monitoring/using-builtin-prometheus.md).
- Monitor your PostgreSQL database with KubeDB using [Prometheus operator](/docs/guides/postgres/monitoring/using-prometheus-operator.md).
- Detail concepts of [Postgres object](/docs/guides/postgres/concepts/postgres.md).
- Use [private Docker registry](/docs/guides/postgres/private-registry/using-private-registry.md) to deploy PostgreSQL with KubeDB.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).