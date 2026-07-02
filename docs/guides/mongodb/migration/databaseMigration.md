---
title: MongoDB Database Migration Guide
menu:
  docs_{{ .version }}:
    identifier: guides-mongodb-migration-database
    name: MongoDB Database Migration
    parent: guides-mongodb-migration
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MongoDB Database Migration

This guide will show you how to use `KubeDB` Migrator to migrate an existing `MongoDB` database — such as one running on DigitalOcean Managed MongoDB or any external instance — entirely into a KubeDB-managed `MongoDB` with minimal downtime.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` operator with the Migrator operator enabled in your cluster following the steps [here](/docs/operatormanual/migration/).

- The source `MongoDB` instance must be network-reachable from within your Kubernetes cluster.

- The source `MongoDB` instance must be part of a replica set with the oplog enabled. The database user provided for migration must have appropriate read privileges on all databases.

- You should be familiar with the following `KubeDB` concepts:
    - [AppBinding](/docs/guides/mongodb/concepts/appbinding.md)
    - [MongoDB](/docs/guides/mongodb/concepts/mongodb.md)
    - [Migrator](/docs/guides/mongodb/concepts/migrator.md)
    - [Migration](/docs/operatormanual/migration/)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

## Prepare Source Database

We will use a **DigitalOcean Managed MongoDB** cluster as the source. Connect to it to verify the prerequisites, set up the migration user, and insert test data.

<details>
<summary><b>Configuring your source instance.</b></summary>

<br> **Self-hosted MongoDB** <br>

A standalone `mongod` must first be [converted to a single-node replica set](https://www.mongodb.com/docs/manual/tutorial/convert-standalone-to-replica-set/). Once running as a replica set, the oplog is enabled automatically. Then create the migration user:
```bash
use admin
db.createUser({
  user: "migrator",
  pwd: "yourStrongPassword",
  roles: [
    { role: "readAnyDatabase", db: "admin" },
    { role: "clusterMonitor", db: "admin" }
  ]
})
```

**MongoDB Atlas** <br>
Atlas clusters run as replica sets by default — no extra configuration needed. Create a database user with **Read Any Database** and **Cluster Monitor** built-in roles in **Database Access** settings.

<br> <br> **DigitalOcean Managed MongoDB** <br>
DigitalOcean managed MongoDB clusters run as replica sets by default. Create a database user with `readAnyDatabase` and `clusterMonitor` roles under the **Users & Databases** section of your cluster dashboard.

See the official [MongoDB Replica Set](https://www.mongodb.com/docs/manual/replication/) docs for more details.

</details>

### Verify prerequisites

Connect to the source instance and verify that the oplog is available:

```bash
$ mongosh "mongodb+srv://<digitalocean-host>.mongo.ondigitalocean.com" -u admin -p
```

```bash
use local
switched to db local

show collections
clustermanager
oplog.rs
replset.election
replset.initialSyncId
replset.minvalid
replset.oplogTruncateAfterPoint
startup_log
system.replset
system.rollback.id
system.tenantMigration.oplogView
system.views

db.oplog.rs.findOne()
{
  op: 'n',
  ns: '',
  o: { msg: 'initiating set' },
  ts: Timestamp({ t: 1782795685, i: 1 }),
  v: Long('2'),
  wall: ISODate('2026-06-30T05:01:25.530Z')
}
```

The `oplog.rs` collection must exist and contain entries — this confirms the source is running as a replica set with the oplog enabled.

### Create a dedicated migration user

Create a dedicated user with the minimum required privileges:

```bash
use admin
db.createUser({
  user: "migrator",
  pwd: "<password>",
  roles: [
    { role: "readAnyDatabase", db: "admin" },
    { role: "clusterMonitor", db: "admin" }
  ]
})
```

The `migrator` user is referenced in the Kubernetes secret and AppBinding for the rest of this guide.

### Create collection and seed data

```bash
use shop

db.orders.insertMany([
  { customer_name: 'Alice', product: 'Laptop',     quantity: 1, status: 'shipped',   created_at: new Date('2026-06-29T08:00:00Z') },
  { customer_name: 'Bob',   product: 'Headphones', quantity: 2, status: 'pending',   created_at: new Date('2026-06-29T08:00:01Z') },
  { customer_name: 'Carol', product: 'Keyboard',   quantity: 3, status: 'delivered', created_at: new Date('2026-06-29T08:00:02Z') }
])

db.orders.find().pretty()
[
  {
    _id: ObjectId("..."),
    customer_name: 'Alice',
    product: 'Laptop',
    quantity: 1,
    status: 'shipped',
    created_at: ISODate('2026-06-29T08:00:00.000Z')
  },
  {
    _id: ObjectId("..."),
    customer_name: 'Bob',
    product: 'Headphones',
    quantity: 2,
    status: 'pending',
    created_at: ISODate('2026-06-29T08:00:01.000Z')
  },
  {
    _id: ObjectId("..."),
    customer_name: 'Carol',
    product: 'Keyboard',
    quantity: 3,
    status: 'delivered',
    created_at: ISODate('2026-06-29T08:00:02.000Z')
  }
]
```

## Prepare Source Connection Information

First, create an authentication secret using the `migrator` user credentials:

```bash
$ kubectl create secret generic source-mongodb-auth -n demo \
                --type=kubernetes.io/basic-auth \
                --from-literal=username=migrator \
                --from-literal=password=<password>
```

If your database has TLS enabled, create a secret with the CA certificate:

```bash
kubectl create secret generic ca-secret \
  --from-file=ca.crt=$CERT_PATH/ca.crt \
  --namespace=demo
```

> **Note:** For mTLS, also include the client certificate and key by appending <br> `--from-file=tls.crt=$CERT_PATH/tls.crt` <br> `--from-file=tls.key=$CERT_PATH/tls.key` <br> to the command above.

Now create an `AppBinding` with the necessary information. The Migrator operator reads the source MongoDB connection information from this AppBinding CR. Use the following YAML to create your AppBinding:

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: mgo-source
  namespace: demo
spec:
  type: mongodb
  version: "4.4.26"
  clientConfig:
    url: "mongodb+srv://<digitalocean-host>.mongo.ondigitalocean.com"
  secret:
    name: source-mongodb-auth
  tlsSecret: # omit if TLS is disabled
    name: ca-secret
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
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/migration/mgo-destination.yaml
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

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/mongodb/migration/mongodb-migrate.yaml
migrator.migrator.kubedb.com/mongodb-migrate created
```

Here,

**`spec.source.mongodb` / `spec.target.mongodb` — connectionInfo:**
- `appBinding.name` / `appBinding.namespace` — references the `AppBinding` for the source or target MongoDB instance.

**`spec.source.mongodb.mongoshake` — migration configuration:**
- `syncMode: all` — performs a full data sync (snapshot + incremental oplog replay).
- `extraConfiguration` — additional `mongoshake` configuration parameters. For example:
  - `full_sync.executor.insert_on_dup_update: "true"` — uses upsert instead of insert during full sync to handle duplicate key errors gracefully.

For a full description of every field, see the [Migrator CRD reference](/docs/guides/mongodb/concepts/migrator.md).

## Watch Migration Progress

Let's wait for the Migration to finish the full sync and enter the incremental sync. Run the following command to watch `Migrator` CR:

```bash
Every 2.0s: kubectl get migrator -n demo
```

During the **full** stage, you'll see progress advancing to 100%:

```bash
NAME              PHASE     DBTYPE    STAGE   LAG   PROGRESS   AGE
mongodb-migrate   Running   mongodb   full          100.00%    22s
```

When the `LAG` drops to near zero, both databases are fully in sync:

```bash
NAME              PHASE     DBTYPE    STAGE   LAG   PROGRESS   AGE
mongodb-migrate   Running   mongodb   incr    0                17h
```

### View detailed progress via pod logs

You can also see collection-wise progress, detailed checkpoints, and sync metrics by checking the migrator pod logs:

```bash
$ kubectl logs -n demo migrator-<migrator-pod-name>
```

Example output during the full sync stage — showing per-collection progress, total/finished/processing/waiting collections:

```log
2026-06-30T12:14:30.632Z	INFO	mongodb	server/utils.go:51	Transferring initial data (Full Sync)	{"Stage": "full", "Progress": "100.00%", "TotalCollections": 1, "FinishedCollections": 1, "ProcessingCollections": 0, "WaitingCollections": 0, "CollectionMetric": {"shop.orders":"100.00% (3/3)"}}
```

Example output during incremental sync — showing LAG, checkpoint timestamps, and LSN details:

```log
2026-06-30T12:58:30.631Z	INFO	mongodb	server/utils.go:51	Incremental replication running	{"Stage": "incr", "Lag": 187, "LSNTime": "2026-06-30 12:58:25", "LSNAckTime": "2026-06-30 12:58:25", "LSNCheckpoint": "2026-06-30 12:55:18"}
```

### Verify initial snapshot on target

Once the migrator reaches the `incr` stage (continuous oplog tailing), exec into the KubeDB target pod and confirm all seed documents were copied over:

```bash
$ kubectl exec -it -n demo mgo-destination-0 -- mongosh -u root -p<root-password>
```

```bash
use shop
db.orders.find().pretty()
[
  {
    _id: ObjectId("..."),
    customer_name: 'Alice',
    product: 'Laptop',
    quantity: 1,
    status: 'shipped',
    created_at: ISODate('2026-06-29T08:00:00.000Z')
  },
  {
    _id: ObjectId("..."),
    customer_name: 'Bob',
    product: 'Headphones',
    quantity: 2,
    status: 'pending',
    created_at: ISODate('2026-06-29T08:00:01.000Z')
  },
  {
    _id: ObjectId("..."),
    customer_name: 'Carol',
    product: 'Keyboard',
    quantity: 3,
    status: 'delivered',
    created_at: ISODate('2026-06-29T08:00:02.000Z')
  }
]
```

### Test live CDC streaming

With the migrator still running, connect to the **source DigitalOcean** instance and run some DML:

```bash
$ mongosh "mongodb+srv://<digitalocean-host>.mongo.ondigitalocean.com" -u migrator -p
```

```bash
use shop

// Insert a new order
db.orders.insertOne({
  customer_name: 'Dave', product: 'Mouse', quantity: 1, status: 'pending', created_at: new Date()
})

// Mark Bob's headphones as delivered
db.orders.updateOne(
  { customer_name: 'Bob' },
  { $set: { status: 'delivered' } }
)

// Remove the already-shipped laptop order
db.orders.deleteOne({ customer_name: 'Alice' })
```

Wait a few seconds for the oplog events to propagate, then re-query the **target**:

```bash
use shop
db.orders.find().pretty()
[
  {
    _id: ObjectId("..."),
    customer_name: 'Bob',
    product: 'Headphones',
    quantity: 2,
    status: 'delivered',
    created_at: ISODate('2026-06-29T08:00:01.000Z')
  },
  {
    _id: ObjectId("..."),
    customer_name: 'Carol',
    product: 'Keyboard',
    quantity: 3,
    status: 'delivered',
    created_at: ISODate('2026-06-29T08:00:02.000Z')
  },
  {
    _id: ObjectId("..."),
    customer_name: 'Dave',
    product: 'Mouse',
    quantity: 1,
    status: 'pending',
    created_at: ISODate('2026-06-29T08:10:00.000Z')
  }
]
```

The INSERT, UPDATE, and DELETE are all reflected on the target — CDC streaming is working correctly.

## Cutover

Once the `LAG` drops to near zero, stop all writes to the source database. Wait until the `LAG` reaches exactly zero — at that point both databases are fully in sync.

Now delete the `Migrator` CR to stop the migration process:

```bash
$ kubectl delete migrator -n demo mongodb-migrate
migrator.migrator.kubedb.com "mongodb-migrate" deleted
```

Finally, update your application's connection string to point to the target KubeDB-managed `MongoDB` database. The migration is complete.
