---
title: PostgreSQL StorageClass Migration Guide
menu:
  docs_{{ .version }}:
    identifier: pg-migration-storageClass
    name: StorageClass Migration
    parent: pg-migration
    weight: 10
menu_name: docs_{{ .version }}
---


> New to KubeDB? Please start [here](/docs/README.md).

# PostgreSQL StorageClass Migration

This guide will show you how to use `KubeDB` Ops Manager to  migrate `StorageClass` of PostgreSQL database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- You must have at least two `StorageClass` resources in order to perform a migration.

- Install `KubeDB` operator in your cluster following the steps [here](/docs/setup/README.md).

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Prepare PostgreSQL Database

At first verify that your cluster has at least two `StorageClass`. Let's check,

```bash
➤ kubectl get storageclass
NAME                   PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
local-path (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  12d
longhorn               driver.longhorn.io      Delete          Immediate              true                   12d
longhorn-custom        driver.longhorn.io      Delete          WaitForFirstConsumer   true                   2d20h
longhorn-static        driver.longhorn.io      Delete          Immediate              true                   12d
```
From the above output we can see that we have more than two `StorageClass` resources. We will now deploy a `PostgreSQL` database using `local-path` StorageClass and insert some data into it.
After that, we will apply `PostgresOpsRequest` to migrate StorageClass from `local-path` to `longhorn-custom`.

> Note: If the `VOLUMEBINDINGMODE` of previous StorageClass is  set to `WaitForFirstConsumer` then the `VOLUMEBINDINGMODE` of new StorageClass must set to `WaitForFirstConsumer`

KubeDB implements a `Postgres` CRD to define the specification of a PostgreSQL database. Below is the `Postgres` object created in this tutorial.

```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: sample-postgres
  namespace: migration
spec:
  version: "13.13"
  replicas: 3
  standbyMode: Hot
  storageType: Durable
  storage:
    storageClassName: "local-path"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 3Gi
  deletionPolicy: WipeOut
```

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/postgres/migration/sample-postgres.yaml
postgres.kubedb.com/sample-postgres created
```
Now, wait until sample-postgres has status `Ready`. i.e,

```bash
$ kubectl get postgres -n demo
NAME              VERSION   STATUS   AGE
sample-postgres   13.13     Ready    41s
```

Lets create a table in the primary.

```bash
# find the primary pod
$ kubectl get pods -n demo --show-labels | grep primary | awk '{ print $1 }'
sample-postgres-0

# exec into the primary and generate some data
$ kubectl exec -it -n demo sample-postgres-0 -- bash
Defaulted container "postgres" out of: postgres, pg-coordinator, postgres-init-container (init)
sample-postgres-0:/$ psql
psql (13.13)
Type "help" for help.

postgres=# create table hello(id int);
CREATE TABLE
postgres=# insert into hello(id) values(generate_series(1,111111));
INSERT 0 111111
postgres=# select count(*) from hello;
 count  
--------
 111111
(1 row)

```

## Apply StorageMigration Ops-Request
To migrate `StorageClass` we have to create a `PostgresOpsRequest` CR with our desired `StorageClass`. Below is the YAML of the `PostgresOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: PostgresOpsRequest
metadata:
  name: storage-migration
  namespace: demo
spec:
  type: StorageMigration
  databaseRef:
    name: sample-postgres
  migration:
    storageClassName: longhorn-custom
    oldPVReclaimPolicy: Delete
```

Here,

- `spec.type` specifies that we are performing `StorageMigration` operation.
- `spec.databaseRef.name` specifies that we are performing StorageMigration operation on `sample-postgres` database.
- `spec.migration.storageClassName` specifies our desired StorageClass
- `spec.migration.oldPVReclaimPolicy` specifies the reclaim policy of previous persistent volume. 

> Note: To retain the old PersistentVolume, set `spec.migration.oldPVReclaimPolicy` to `Retain`.

Let's create the `PostgresOpsRequest` CR we have shown above,

``` bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/postgres/migration/storage-migration.yaml
postgresopsrequest.ops.kubedb.com/storage-migration created
```
## Verify the StorageClass Migrated Successfully

If everything goes well, `KubeDb` operator will migrate the `StorageClass` along with the data.

Let’s wait for `PostgresOpsRequest` to be `Successful`. Run the following command to watch PostgresOpsRequest CR,

``` bash
$ watch kubectl get postgresopsrequest -n demo

Every 2.0s: kubectl get postgresopsrequest -n demo  

NAME                TYPE               STATUS       AGE
storage-migration   StorageMigration   Successful   13m
```

We can see from the above output that the `PostgresOpsRequest` has succeded.

Now, we will verify that the data remains intact after the `StorageMigration` operation. Let's exec into one of the `Postgres` pod and perform read query.

```bash
$ kubectl exec -it -n demo sample-postgres-0 -- bash
Defaulted container "postgres" out of: postgres, pg-coordinator, postgres-init-container (init)
sample-postgres-0:/$ psql
psql (13.13)
Type "help" for help.

postgres=# select count(*) from hello;
 count  
--------
 111111
(1 row)
```

From the above output we can verify that data remains intact after the `StorageMigration` operation.

## CleanUp

To clean up the Kubernetes resources created by this tutorial, run:

```bash
$ kubectl delete postgresopsrequest -n demo storage-migration
$ kubectl delete postgres -n demo sample-postgres
$ kubectl delete ns demo
```