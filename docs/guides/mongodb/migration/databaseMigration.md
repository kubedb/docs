---
title: MongoDB Database Migration Guide
menu:
  docs_{{ .version }}:
    identifier: guides-mongodb-migration-database
    name: MongoDB Database Migration
    parent: guides-mongodb-migration
    weight: 11
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MongoDB Database Migration

This guide will show you how to use `KubeDB` Migrator to migrate an existing `MongoDB` database — such as one running on MongoDB Atlas or any external instance — entirely into a KubeDB-managed `MongoDB` with minimal downtime.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` operator with the Migrator operator enabled in your cluster following the steps [here](/docs/operatormanual/migration/).

- The source `MongoDB` instance must be network-reachable from within your Kubernetes cluster.

- The source `MongoDB` instance must be part of a replica set with the oplog enabled. The database user provided for migration must have appropriate read privileges on all databases.

- You should be familiar with the following `KubeDB` concepts:
    - [AppBinding](/docs/guides/mongodb/concepts/appbinding/)
    - [MongoDB](/docs/guides/mongodb/concepts/mongodb)
    - [Migration](/docs/operatormanual/migration/)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Prepare Source Connection Information

First, create an authentication secret to communicate with the source MongoDB database:

```bash
$ kubectl create secret generic source-mongodb-auth -n demo \
                --type=kubernetes.io/basic-auth \
                --from-literal=username=<username> \
                --from-literal=password=<password>
```

Now create an `AppBinding` with the necessary information. The Migrator operator reads the source MongoDB connection information from this AppBinding CR. Use the following YAML to create your AppBinding:

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: mgo-source
  namespace: demo
spec:
  type: mongodb
  version: "5.0.3"
  clientConfig:
    url: "mongodb://host:port"
  secret:
    name: source-mongodb-auth
```

Here,

- `spec.clientConfig.url` is the connection URL of the source MongoDB instance.
- `spec.secret.name` is the reference to the secret we created earlier, containing the MongoDB authentication information.

> For a `KubeDB`-managed database, an `AppBinding` is created by default. So there is no need to create one for the target database.

## Create Target MongoDB Database

KubeDB implements a `MongoDB` CRD to define the specification of a MongoDB database. Use the following `MongoDB` object to create the target database.

```yaml
apiVersion: kubedb.com/v1
kind: MongoDB
metadata:
  name: mgo-destination
  namespace: demo
spec:
  version: "4.4.26"
  replicaSet:
    name: "rs1"
  replicas: 2
  storageType: Durable
  storage:
    storageClassName: "local-path"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 15Gi
  serviceTemplates:
  - alias: primary
    metadata:
      annotations:
        passMe: ToService
    spec:
      type: NodePort
      ports:
      - name:  http
        port:  27017
```

```bash
$ kubectl apply -f target-mongodb.yaml
mongodb.kubedb.com/mgo-destination created
```

> Note: Adjust the `resources.requests.storage` based on the source database size.

Wait until `mgo-destination` has status `Ready`.

## Apply Migrator CR

To migrate the database we have to create a `Migrator` CR. KubeDB uses `mongoshake` to perform the migration. Below is the YAML of the `Migrator` CR that we are going to create:

```yaml
apiVersion: migrator.kubedb.com/v1alpha1
kind: Migrator
metadata:
  name: mongodb-migrate
  namespace: demo
spec:
  source:
    mongodb:
      connectionInfo:
        appBinding:
          name: mgo-source
          namespace: demo
      mongoshake:
        syncMode: all
        extraConfiguration:
          full_sync.executor.insert_on_dup_update: "true"
  target:
    mongodb:
      connectionInfo:
        appBinding:
          name: mgo-destination
          namespace: demo
```

Here,

**`spec.source.mongodb` / `spec.target.mongodb` — connectionInfo:**
- `appBinding.name` / `appBinding.namespace` — references the `AppBinding` for the source or target MongoDB instance.

**`spec.source.mongodb.mongoshake` — migration configuration:**
- `syncMode: all` — performs a full data sync (snapshot + incremental oplog replay).
- `extraConfiguration` — additional `mongoshake` configuration parameters. For example:
  - `full_sync.executor.insert_on_dup_update: "true"` — uses upsert instead of insert during full sync to handle duplicate key errors gracefully.

## Watch Migration Progress

Let's wait for the `LAG` to reach near zero. Run the following command to watch `Migrator` CR:

```bash
Every 2.0s: kubectl get migrator -n demo

NAME              PHASE     DBTYPE    STAGE   LAG   PROGRESS   AGE
mongodb-migrate   Running   mongodb   incr    0                17h
```

## Cutover

Once the `LAG` drops to near zero, stop all writes to the source database. Wait until the `LAG` reaches exactly zero — at that point both databases are fully in sync.

Now delete the `Migrator` CR to stop the migration process:

```bash
$ kubectl delete migrator -n demo mongodb-migrate
migrator.migrator.kubedb.com "mongodb-migrate" deleted
```

Finally, update your application's connection string to point to the target KubeDB-managed `MongoDB` database. The migration is complete.
