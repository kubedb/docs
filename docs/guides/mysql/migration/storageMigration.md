---
title: MySQL StorageClass Migration Guide
menu:
  docs_{{ .version }}:
    identifier: guides-mysql-migration-storageclass
    name: StorageClass Migration
    parent: guides-mysql-migration
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MySQL StorageClass Migration

This guide will show you how to use `KubeDB` Ops Manager to  migrate `StorageClass` of MySQL database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- You must have at least two `StorageClass` resources in order to perform a migration.

- Install `KubeDB` operator in your cluster following the steps [here](/docs/setup/README.md).

- You should be familiar with the following `KubeDB` concepts:
    - [MySQL](/docs/guides/mysql/concepts/mysqldatabase)
    - [MySQLOpsRequest](/docs/guides/mysql/concepts/opsrequest)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Prepare MySQL Database

At first verify that your cluster has at least two `StorageClass`. Let's check,

```bash
➤ kubectl get storageclass
NAME                   PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
local-path (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  12d
longhorn               driver.longhorn.io      Delete          Immediate              true                   12d
longhorn-custom        driver.longhorn.io      Delete          WaitForFirstConsumer   true                   2d20h
longhorn-static        driver.longhorn.io      Delete          Immediate              true                   12d
```
From the above output we can see that we have more than two `StorageClass` resources. We will now deploy a `MySQL` database using `local-path` StorageClass and insert some data into it. 
After that, we will apply `MySQLOpsRequest` to migrate StorageClass from `local-path` to `longhorn-custom`.

> Note: If the `VOLUMEBINDINGMODE` of previous StorageClass is  set to `WaitForFirstConsumer` then the `VOLUMEBINDINGMODE` of new StorageClass must set to `WaitForFirstConsumer`

KubeDB implements a `MySQL` CRD to define the specification of a MySQL database. Below is the `MySQL` object created in this tutorial. 

```yaml
apiVersion: kubedb.com/v1
kind: MySQL
metadata:
  name: sample-mysql
  namespace: demo
spec:
  version: "9.1.0"
  replicas: 3
  topology:
    mode: GroupReplication
    group:
      mode: Single-Primary
  storageType: Durable
  storage:
    storageClassName: local-path
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mysql/migration/sample-mysql.yaml
mysql.kubedb.com/sample-mysql created
```
Now, wait until sample-mysql has status `Ready` and check the `StorageClass`,

```bash
$ kubectl get mysql,pvc -n demo
NAME                            VERSION   STATUS   AGE
mysql.kubedb.com/sample-mysql   9.1.0     Ready    101s

NAME                                        STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   VOLUMEATTRIBUTESCLASS   AGE
persistentvolumeclaim/data-sample-mysql-0   Bound    pvc-64cca3c6-85aa-426f-abc3-b300ecfe365a   1Gi        RWO            local-path     <unset>                 96s
persistentvolumeclaim/data-sample-mysql-1   Bound    pvc-1de36b06-8e32-4e9a-a01b-3b6d7c618688   1Gi        RWO            local-path     <unset>                 90s
persistentvolumeclaim/data-sample-mysql-2   Bound    pvc-a75bd538-8a71-4f62-8d38-3f4e42ffb225   1Gi        RWO            local-path     <unset>                 85s
```

The database is `Ready` and all the `PersistentVolumeClaim` uses `local-path`  StorageClass, Let's create a table in the primary.

```bash
# find the primary pod
kubectl get pods -n demo --show-labels | grep primary | awk '{ print $1 }'
sample-mysql-0

# exec into the primary pod
$ kubectl exec -it -n demo sample-mysql-0 -- bash
Defaulted container "mysql" out of: mysql, mysql-coordinator, mysql-init (init)
bash-5.1$ mysql -uroot -p$MYSQL_ROOT_PASSWORD
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 1780
Server version: 9.1.0 MySQL Community Server - GPL

Copyright (c) 2000, 2024, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> create database hello;
Query OK, 1 row affected (0.02 sec)

mysql> show databases;
+--------------------+
| Database           |
+--------------------+
| hello              |
| information_schema |
| kubedb_system      |
| mysql              |
| performance_schema |
| sys                |
+--------------------+
6 rows in set (0.00 sec)

mysql> use hello;
Database changed

# Create a table
mysql> CREATE TABLE users (
       id INT AUTO_INCREMENT PRIMARY KEY,
       name VARCHAR(50),
       email VARCHAR(100)
      );

Query OK, 0 rows affected (0.03 sec)

# Insert some data into the table

mysql> INSERT INTO users (name, email) VALUES
      ('David', 'david@example.com'),
      ('Eva', 'eva@example.com'),
      ('Frank', 'frank@example.com'),
      ('Grace', 'grace@example.com'),
      ('Hannah', 'hannah@example.com'),
      ('Ian', 'ian@example.com'),
      ('Jack', 'jack@example.com'),
      ('Karen', 'karen@example.com'),
      ('Liam', 'liam@example.com'),
      ('Mona', 'mona@example.com'),
      ('Nathan', 'nathan@example.com'),
      ('Olivia', 'olivia@example.com'),
      ('Paul', 'paul@example.com'),
      ('Quincy', 'quincy@example.com'),
      ('Rachel', 'rachel@example.com'),
      ('Steve', 'steve@example.com'),
      ('Tina', 'tina@example.com'),
      ('Uma', 'uma@example.com'),
      ('Victor', 'victor@example.com'),
      ('Wendy', 'wendy@example.com');

Query OK, 20 rows affected (0.02 sec)
Records: 20  Duplicates: 0  Warnings: 0
```

## Apply StorageMigration Ops-Request
To migrate `StorageClass` we have to create a `MySQLOpsRequest` CR with our desired `StorageClass`. Below is the YAML of the `MySQLOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MySQLOpsRequest
metadata:
  name: storage-migration
  namespace: demo
spec:
  type: StorageMigration
  databaseRef:
    name: sample-mysql
  migration:
    storageClassName: longhorn-custom
    oldPVReclaimPolicy: Delete
```

Here,

- `spec.type` specifies that we are performing `StorageMigration` operation.
- `spec.databaseRef.name` specifies that we are performing StorageMigration operation on `sample-mysql` database.
- `spec.migration.storageClassName` specifies our desired StorageClass
- `spec.migration.oldPVReclaimPolicy` specifies the reclaim policy of previous persistent volume.

> Note: To retain the old PersistentVolume, set `spec.migration.oldPVReclaimPolicy` to `Retain`.

Let's create the `MySQLOpsRequest` CR we have shown above,

``` bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mysql/migration/storage-migration.yaml
mysqlopsrequest.ops.kubedb.com/storage-migration created
```
## Verify the StorageClass Migrated Successfully

If everything goes well, `KubeDb` operator will migrate the `StorageClass` along with the data.

Let’s wait for `MySQLOpsRequest` to be `Successful`. Run the following command to watch MySQLOpsRequest CR,

``` bash
$ watch kubectl get mysqlopsrequest -n demo

Every 2.0s: kubectl get mysqlopsrequest -n demo

NAME                TYPE               STATUS       AGE
storage-migration   StorageMigration   Successful   12m
```
We can see from the above output that the `MySQLOpsRequest` has succeeded. Let's verify the StorageClass.

``` bash
$ kubectl get pvc -n demo
NAME                  STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS        VOLUMEATTRIBUTESCLASS   AGE
data-sample-mysql-0   Bound    pvc-64cca3c6-85aa-426f-abc3-b300ecfe365a   1Gi        RWO            longhorn-custom     <unset>                 21m
data-sample-mysql-1   Bound    pvc-1de36b06-8e32-4e9a-a01b-3b6d7c618688   1Gi        RWO            longhorn-custom     <unset>                 21m
data-sample-mysql-2   Bound    pvc-a75bd538-8a71-4f62-8d38-3f4e42ffb225   1Gi        RWO            longhorn-custom     <unset>                 21m
```

The `PersistentVolumeClaim` StorageClass has changed to `longhorn-custom`.  Now, we will verify that the data remains intact after the `StorageMigration` operation. Let's exec into one of the `MySQL` pod and perform read query.

```bash
$ kubectl exec -it -n demo sample-mysql-0 -- bash
Defaulted container "mysql" out of: mysql, mysql-coordinator, mysql-init (init)
bash-5.1$ mysql -uroot -p$MYSQL_ROOT_PASSWORD
mysql: [Warning] Using a password on the command line interface can be insecure.
Welcome to the MySQL monitor.  Commands end with ; or \g.
Your MySQL connection id is 225
Server version: 9.1.0 MySQL Community Server - GPL

Copyright (c) 2000, 2024, Oracle and/or its affiliates.

Oracle is a registered trademark of Oracle Corporation and/or its
affiliates. Other names may be trademarks of their respective
owners.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

mysql> select * from hello.users;
+----+--------+--------------------+
| id | name   | email              |
+----+--------+--------------------+
|  1 | David  | david@example.com  |
|  2 | Eva    | eva@example.com    |
|  3 | Frank  | frank@example.com  |
|  4 | Grace  | grace@example.com  |
|  5 | Hannah | hannah@example.com |
|  6 | Ian    | ian@example.com    |
|  7 | Jack   | jack@example.com   |
|  8 | Karen  | karen@example.com  |
|  9 | Liam   | liam@example.com   |
| 10 | Mona   | mona@example.com   |
| 11 | Nathan | nathan@example.com |
| 12 | Olivia | olivia@example.com |
| 13 | Paul   | paul@example.com   |
| 14 | Quincy | quincy@example.com |
| 15 | Rachel | rachel@example.com |
| 16 | Steve  | steve@example.com  |
| 17 | Tina   | tina@example.com   |
| 18 | Uma    | uma@example.com    |
| 19 | Victor | victor@example.com |
| 20 | Wendy  | wendy@example.com  |
+----+--------+--------------------+
20 rows in set (0.00 sec)

```

From the above output we can verify that data remains intact after the `StorageMigration` operation.

## CleanUp

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete mysqlopsrequest -n demo storage-migration
$ kubectl delete mysql -n demo sample-mysql
$ kubectl delete ns demo
```
