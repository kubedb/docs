---
title: Backup & Restore Sharded MongoDB Cluster| Stash
description: Backup and restore sharded MongoDB cluster using Stash
menu:
  docs_{{ .version }}:
    identifier: guides-mongodb-backup-logical-sharded-cluster
    name: MongoDB Sharded Cluster
    parent: guides-mongodb-backup-logical
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Backup and Restore MongoDB Sharded Clusters using Stash

Stash 0.9.0+ supports taking [backup](https://docs.mongodb.com/manual/tutorial/backup-sharded-cluster-with-database-dumps/) and [restores](https://docs.mongodb.com/manual/tutorial/restore-sharded-cluster/) MongoDB Sharded clusters in ["idiomatic" way](https://docs.mongodb.com/manual/administration/backup-sharded-clusters/). This guide will show you how you can backup and restore your MongoDB Sharded clusters with Stash.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using Minikube.
- Install KubeDB in your cluster following the steps [here](/docs/setup/README.md).
- Install Stash Enterprise in your cluster following the steps [here](https://stash.run/docs/latest/setup/install/enterprise/).
- Install Stash `kubectl` plugin following the steps [here](https://stash.run/docs/latest/setup/install/kubectl-plugin/).
- If you are not familiar with how Stash backup and restore MongoDB databases, please check the following guide [here](/docs/guides/mongodb/backup/overview/index.md).

You have to be familiar with following custom resources:

- [AppBinding](/docs/guides/mongodb/concepts/appbinding.md)
- [Function](https://stash.run/docs/latest/concepts/crds/function/)
- [Task](https://stash.run/docs/latest/concepts/crds/task/)
- [BackupConfiguration](https://stash.run/docs/latest/concepts/crds/backupconfiguration/)
- [RestoreSession](https://stash.run/docs/latest/concepts/crds/restoresession/)

To keep things isolated, we are going to use a separate namespace called `demo` throughout this tutorial. Create `demo` namespace if you haven't created yet.

```console
$ kubectl create ns demo
namespace/demo created
```

## Backup Sharded MongoDB Cluster

This section will demonstrate how to backup MongoDB cluster. We are going to use [KubeDB](https://kubedb.com) to deploy a sample database. Then, we are going to backup this database into a GCS bucket. Finally, we are going to restore the backed up data into another MongoDB cluster.

### Deploy Sample MongoDB Sharding

Let's deploy a sample MongoDB Sharding database and insert some data into it.

**Create MongoDB CRD:**

Below is the YAML of a sample MongoDB crd that we are going to create for this tutorial:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  name: sample-mgo-sh
  namespace: demo
spec:
  version: 4.4.26
  shardTopology:
    configServer:
      replicas: 3
      storage:
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
    mongos:
      replicas: 2
    shard:
      replicas: 3
      shards: 3
      storage:
        resources:
          requests:
            storage: 1Gi
        storageClassName: standard
  terminationPolicy: WipeOut
```

Create the above `MongoDB` crd,

```console
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mongodb/backup/logical/sharding/examples/mongodb-sharding.yaml
mongodb.kubedb.com/sample-mgo-sh created
```

KubeDB will deploy a MongoDB database according to the above specification. It will also create the necessary secrets and services to access the database.

Let's check if the database is ready to use,

```console
$ kubectl get mg -n demo sample-mgo-sh
NAME            VERSION       STATUS  AGE
sample-mgo-sh   4.4.26         Ready   35m
```

The database is `Ready`. Verify that KubeDB has created a Secret and a Service for this database using the following commands,

```console
$ kubectl get secret -n demo -l=app.kubernetes.io/instance=sample-mgo-sh
NAME                 TYPE     DATA   AGE
sample-mgo-sh-auth   Opaque   2      36m
sample-mgo-sh-cert   Opaque   4      36m

$ kubectl get service -n demo -l=app.kubernetes.io/instance=sample-mgo-sh
NAME                          TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)     AGE
sample-mgo-sh                 ClusterIP   10.107.11.117   <none>        27017/TCP   36m
sample-mgo-sh-configsvr-gvr   ClusterIP   None            <none>        27017/TCP   36m
sample-mgo-sh-shard0-gvr      ClusterIP   None            <none>        27017/TCP   36m
sample-mgo-sh-shard1-gvr      ClusterIP   None            <none>        27017/TCP   36m
sample-mgo-sh-shard2-gvr      ClusterIP   None            <none>        27017/TCP   36m
```

KubeDB creates an [AppBinding](/docs/guides/mongodb/concepts/appbinding.md) crd that holds the necessary information to connect with the database.

**Verify AppBinding:**

Verify that the `AppBinding` has been created successfully using the following command,

```console
$ kubectl get appbindings -n demo
NAME            AGE
sample-mgo-sh   30m
```

Let's check the YAML of the above `AppBinding`,

```console
$ kubectl get appbindings -n demo sample-mgo-sh -o yaml
```

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1alpha2","kind":"MongoDB","metadata":{"annotations":{},"name":"sample-mgo-sh","namespace":"demo"},"spec":{"shardTopology":{"configServer":{"replicas":3,"storage":{"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"}},"mongos":{"replicas":2},"shard":{"replicas":3,"shards":3,"storage":{"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"}}},"terminationPolicy":"WipeOut","version":"4.4.26"}}
  creationTimestamp: "2022-10-26T05:11:20Z"
  generation: 1
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: sample-mgo-sh
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: mongodbs.kubedb.com
  name: sample-mgo-sh
  namespace: demo
  ownerReferences:
    - apiVersion: kubedb.com/v1alpha2
      blockOwnerDeletion: true
      controller: true
      kind: MongoDB
      name: sample-mgo-sh
      uid: 22f704c3-1a4d-468c-9404-7efa739ad0da
  resourceVersion: "580483"
  uid: 69092658-2f4a-45f2-a899-14884bf74a8b
spec:
  appRef:
    apiGroup: kubedb.com
    kind: MongoDB
    name: sample-mgo-sh
    namespace: demo
  clientConfig:
    service:
      name: sample-mgo-sh
      port: 27017
      scheme: mongodb
  parameters:
    apiVersion: config.kubedb.com/v1alpha1
    configServer: cnfRepSet/sample-mgo-sh-configsvr-0.sample-mgo-sh-configsvr-pods.demo.svc:27017,sample-mgo-sh-configsvr-1.sample-mgo-sh-configsvr-pods.demo.svc:27017,sample-mgo-sh-configsvr-2.sample-mgo-sh-configsvr-pods.demo.svc:27017
    kind: MongoConfiguration
    replicaSets:
      host-0: shard0/sample-mgo-sh-shard0-0.sample-mgo-sh-shard0-pods.demo.svc:27017,sample-mgo-sh-shard0-1.sample-mgo-sh-shard0-pods.demo.svc:27017,sample-mgo-sh-shard0-2.sample-mgo-sh-shard0-pods.demo.svc:27017
      host-1: shard1/sample-mgo-sh-shard1-0.sample-mgo-sh-shard1-pods.demo.svc:27017,sample-mgo-sh-shard1-1.sample-mgo-sh-shard1-pods.demo.svc:27017,sample-mgo-sh-shard1-2.sample-mgo-sh-shard1-pods.demo.svc:27017
      host-2: shard2/sample-mgo-sh-shard2-0.sample-mgo-sh-shard2-pods.demo.svc:27017,sample-mgo-sh-shard2-1.sample-mgo-sh-shard2-pods.demo.svc:27017,sample-mgo-sh-shard2-2.sample-mgo-sh-shard2-pods.demo.svc:27017
    stash:
      addon:
        backupTask:
          name: mongodb-backup-4.4.6
        restoreTask:
          name: mongodb-restore-4.4.6
  secret:
    name: sample-mgo-sh-auth
  type: kubedb.com/mongodb
  version: 4.4.26
```

Stash uses the `AppBinding` crd to connect with the target database. It requires the following two fields to set in AppBinding's `Spec` section.

- `spec.appRef` refers to the underlying application.
- `spec.clientConfig` defines how to communicate with the application.
- `spec.clientConfig.service.name` specifies the name of the service that connects to the database.
- `spec.secret` specifies the name of the secret that holds necessary credentials to access the database.
- `spec.parameters.configServer` specifies the dsn of config server of mongodb sharding. The dsn includes the port no too.
- `spec.parameters.replicaSets` contains the dsn of each replicaset of sharding. The DSNs are in key-value pair, where the keys are host-0, host-1 etc, and the values are DSN of each replicaset. If there is no sharding but only one replicaset, then ReplicaSets field contains only one key-value pair where the key is host-0 and the value is dsn of that replicaset.
- `spec.parameters.stash` contains the Stash addon information that will be used to backup and restore this MongoDB.
- `spec.type` specifies the types of the app that this AppBinding is pointing to. KubeDB generated AppBinding follows the following format: `<app group>/<app resource type>`.

**Insert Sample Data:**

Now, we are going to exec into the database pod and create some sample data. At first, find out the database pod using the following command,

```console
$ kubectl get pods -n demo --selector="mongodb.kubedb.com/node.mongos=sample-mgo-sh-mongos"
NAME                                   READY   STATUS    RESTARTS   AGE
sample-mgo-sh-mongos-9459cfc44-4jthd   1/1     Running   0          60m
sample-mgo-sh-mongos-9459cfc44-6d2st   1/1     Running   0          60m
```

Now, let's exec into the pod and create a table,

```console
$ export USER=$(kubectl get secrets -n demo sample-mgo-sh-auth -o jsonpath='{.data.\username}' | base64 -d)

$ export PASSWORD=$(kubectl get secrets -n demo sample-mgo-sh-auth -o jsonpath='{.data.\password}' | base64 -d)

$ kubectl exec -it -n demo sample-mgo-sh-mongos-9459cfc44-4jthd -- mongo admin -u $USER -p $PASSWORD

mongos> show dbs
admin   0.000GB
config  0.001GB


mongos> show users
{
	"_id" : "admin.root",
	"userId" : UUID("b9a1551b-83cf-4ebb-852b-dd23c890f301"),
	"user" : "root",
	"db" : "admin",
	"roles" : [
		{
			"role" : "root",
			"db" : "admin"
		}
	]
}

mongos> use newdb
switched to db newdb

mongos> db.movie.insert({"name":"batman"});
WriteResult({ "nInserted" : 1 })

mongos> db.movie.find().pretty()
{ "_id" : ObjectId("5d3064bf144a1b8fda04cd4f"), "name" : "batman" }

mongos> exit
bye
```

Now, we are ready to backup this sample database.

### Prepare Backend

We are going to store our backed up data into a GCS bucket. At first, we need to create a secret with GCS credentials then we need to create a `Repository` crd. If you want to use a different backend, please read the respective backend configuration doc from [here](https://stash.run/docs/latest/guides/backends/overview/).

**Create Storage Secret:**

Let's create a secret called `gcs-secret` with access credentials to our desired GCS bucket,

```console
$ echo -n 'changeit' > RESTIC_PASSWORD
$ echo -n '<your-project-id>' > GOOGLE_PROJECT_ID
$ cat downloaded-sa-key.json > GOOGLE_SERVICE_ACCOUNT_JSON_KEY
$ kubectl create secret generic -n demo gcs-secret \
    --from-file=./RESTIC_PASSWORD \
    --from-file=./GOOGLE_PROJECT_ID \
    --from-file=./GOOGLE_SERVICE_ACCOUNT_JSON_KEY
secret/gcs-secret created
```

**Create Repository:**

Now, crete a `Repository` using this secret. Below is the YAML of Repository crd we are going to create,

```yaml
apiVersion: stash.appscode.com/v1alpha1
kind: Repository
metadata:
  name: gcs-repo-sharding
  namespace: demo
spec:
  backend:
    gcs:
      bucket: appscode-qa
      prefix: demo/mongodb/sample-mgo-sh
    storageSecretName: gcs-secret
```

Let's create the `Repository` we have shown above,

```console
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mongodb/backup/logical/sharding/examples/repository-sharding.yaml
repository.stash.appscode.com/gcs-repo-sharding created
```

Now, we are ready to backup our database to our desired backend.

### Backup MongoDB Sharding

We have to create a `BackupConfiguration` targeting respective AppBinding crd of our desired database. Then Stash will create a CronJob to periodically backup the database.

**Create BackupConfiguration:**

Below is the YAML for `BackupConfiguration` crd to backup the `sample-mgo-sh` database we have deployed earlier.,

```yaml
apiVersion: stash.appscode.com/v1beta1
kind: BackupConfiguration
metadata:
  name: sample-mgo-sh-backup
  namespace: demo
spec:
  schedule: "*/5 * * * *"
  repository:
    name: gcs-repo-sharding
  target:
    ref:
      apiVersion: appcatalog.appscode.com/v1alpha1
      kind: AppBinding
      name: sample-mgo-sh
  retentionPolicy:
    name: keep-last-5
    keepLast: 5
    prune: true
```

Here,

- `spec.schedule` specifies that we want to backup the database at 5 minutes interval.
- `spec.target.ref` refers to the `AppBinding` crd that was created for `sample-mgo-sh` database.

Let's create the `BackupConfiguration` crd we have shown above,

```console
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mongodb/backup/logical/sharding/examples/backupconfiguration-sharding.yaml
backupconfiguration.stash.appscode.com/sample-mgo-sh-backup created
```

**Verify Backup Setup Successful:**

If everything goes well, the phase of the `BackupConfiguration` should be `Ready`. The `Ready` phase indicates that the backup setup is successful. Let's verify the `Phase` of the BackupConfiguration,

```console
$ kubectl get backupconfiguration -n demo
NAME                    TASK                    SCHEDULE      PAUSED   PHASE      AGE
sample-mgo-sh-backup    mongodb-backup-4.4.6    */5 * * * *            Ready      11s
```

**Verify CronJob:**

Stash will create a CronJob with the schedule specified in `spec.schedule` field of `BackupConfiguration` crd.

Verify that the CronJob has been created using the following command,

```console
$ kubectl get cronjob -n demo
NAME                   SCHEDULE      SUSPEND   ACTIVE   LAST SCHEDULE   AGE
sample-mgo-sh-backup   */5 * * * *   False     0        <none>          13s
```

**Wait for BackupSession:**

The `sample-mgo-sh-backup` CronJob will trigger a backup on each schedule by creating a `BackupSession` crd.

Wait for the next schedule. Run the following command to watch `BackupSession` crd,

```console
$ kubectl get backupsession -n demo -w
NAME                              INVOKER-TYPE          INVOKER-NAME           PHASE       AGE
sample-mgo-sh-backup-1563512707   BackupConfiguration   sample-mgo-sh-backup   Running     5m19s
sample-mgo-sh-backup-1563512707   BackupConfiguration   sample-mgo-sh-backup   Succeeded   5m45s
```

We can see above that the backup session has succeeded. Now, we are going to verify that the backed up data has been stored in the backend.

**Verify Backup:**

Once a backup is complete, Stash will update the respective `Repository` crd to reflect the backup. Check that the repository `gcs-repo-sharding` has been updated by the following command,

```console
$ kubectl get repository -n demo gcs-repo-sharding
NAME                INTEGRITY   SIZE         SNAPSHOT-COUNT   LAST-SUCCESSFUL-BACKUP   AGE
gcs-repo-sharding   true        66.453 KiB   12               1m                       20m
```

Now, if we navigate to the GCS bucket, we are going to see backed up data has been stored in `demo/mongodb/sample-mgo-sh` directory as specified by `spec.backend.gcs.prefix` field of Repository crd.

> Note: Stash keeps all the backed up data encrypted. So, data in the backend will not make any sense until they are decrypted.

## Restore MongoDB Sharding
You can restore your data into the same database you have backed up from or into a different database in the same cluster or a different cluster. In this section, we are going to show you how to restore in the same database which may be necessary when you have accidentally deleted any data from the running database.


#### Stop Taking Backup of the Old Database:

At first, let's stop taking any further backup of the old database so that no backup is taken during restore process. We are going to pause the `BackupConfiguration` crd that we had created to backup the `sample-mgo-sh` database. Then, Stash will stop taking any further backup for this database.

Let's pause the `sample-mgo-sh-backup` BackupConfiguration,
```bash
$ kubectl patch backupconfiguration -n demo sample-mgo-sh-backup --type="merge" --patch='{"spec": {"paused": true}}'
backupconfiguration.stash.appscode.com/sample-mgo-sh-backup patched
```

Or you can use the Stash `kubectl` plugin to pause the `BackupConfiguration`,
```bash
$ kubectl stash pause backup -n demo --backupconfig=sample-mgo-sh-backup
BackupConfiguration demo/sample-mgo-sh-backup has been paused successfully.
```

Now, wait for a moment. Stash will pause the BackupConfiguration. Verify that the BackupConfiguration  has been paused,

```console
$ kubectl get backupconfiguration -n demo sample-mgo-sh-backup
NAME                  TASK                         SCHEDULE      PAUSED   PHASE   AGE
sample-mgo-sh-backup  mongodb-restore-4.4.6        */5 * * * *   true     Ready   26m
```

Notice the `PAUSED` column. Value `true` for this field means that the BackupConfiguration has been paused.

#### Simulate Disaster:

Now, let’s simulate an accidental deletion scenario. Here, we are going to exec into the database pod and delete the `newdb` database we had created earlier.
```console
$ kubectl exec -it -n demo sample-mgo-sh-mongos-9459cfc44-4jthd -- mongo admin -u $USER -p $PASSWORD

mongos> use newdb
switched to db newdb

mongos> db.dropDatabase()
{ "dropped" : "newdb", "ok" : 1 }

mongos> show dbs
admin   0.000GB
config  0.000GB
local   0.000GB

mongos> exit
bye
```

#### Create RestoreSession:

Now, we need to create a `RestoreSession` crd pointing to the AppBinding of `sample-mgo-sh` database.

Below is the YAML for the `RestoreSession` crd that we are going to create to restore the backed up data.

```yaml
apiVersion: stash.appscode.com/v1beta1
kind: RestoreSession
metadata:
  name: sample-mgo-sh-restore
  namespace: demo
spec:
  repository:
    name: gcs-repo-sharding
  target:
    ref:
      apiVersion: appcatalog.appscode.com/v1alpha1
      kind: AppBinding
      name: sample-mgo-sh
  rules:
  - snapshots: [latest]
```

Here,

- `spec.repository.name` specifies the `Repository` crd that holds the backend information where our backed up data has been stored.
- `spec.target.ref` refers to the AppBinding crd for the `restored-mgo-sh` database.
- `spec.rules` specifies that we are restoring from the latest backup snapshot of the database.

Let's create the `RestoreSession` crd we have shown above,

```console
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mongodb/backup/logical/sharding/examples/restoresession-sharding.yaml
restoresession.stash.appscode.com/sample-mgo-sh-restore created
```

Once, you have created the `RestoreSession` crd, Stash will create a job to restore. We can watch the `RestoreSession` phase to check if the restore process is succeeded or not.

Run the following command to watch `RestoreSession` phase,

```console
$ kubectl get restoresession -n demo sample-mgo-sh-restore -w
NAME                    REPOSITORY-NAME      PHASE       AGE
sample-mgo-sh-restore   gcs-repo-sharding    Running     5s
sample-mgo-sh-restore   gcs-repo-sharding    Succeeded   43s
```

So, we can see from the output of the above command that the restore process succeeded.

#### Verify Restored Data:

In this section, we are going to verify that the desired data has been restored successfully. We are going to connect to `mongos` and check whether the table we had created earlier is restored or not.

Lets, exec into the database pod and list available tables,

```console

$ kubectl exec -it -n demo sample-mgo-sh-mongos-9459cfc44-4jthd -- mongo admin -u $USER -p $PASSWORD

mongos> show dbs
admin   0.000GB
config  0.001GB
newdb   0.000GB


mongos> show users
{
	"_id" : "admin.root",
	"userId" : UUID("a57cb466-ec66-453b-b795-654169a0f035"),
	"user" : "root",
	"db" : "admin",
	"roles" : [
		{
			"role" : "root",
			"db" : "admin"
		}
	]
}

mongos> use newdb
switched to db newdb

mongos> db.movie.find().pretty()
{ "_id" : ObjectId("5d3064bf144a1b8fda04cd4f"), "name" : "batman" }

mongos> exit
bye
```

So, from the above output, we can see the database `newdb` that we had created earlier is restored.

## Backup MongoDB Sharded Cluster and Restore into a Standalone database

It is possible to take backup of a MongoDB Sharded Cluster and restore it into a standalone database, but user need to create the appbinding for this process.

### Backup a sharded cluster

Keep all the fields of appbinding that is explained earlier in this guide, except `spec.parameter`. Do not set `spec.parameter.configServer` and `spec.parameter.replicaSet`. By doing this, the job will use `spec.clientConfig.service.name` as host, which is `mongos` router DSN. So, the backup will treat this cluster as a standalone and will skip the [`idiomatic way` of taking backups of a sharded cluster](https://docs.mongodb.com/manual/tutorial/backup-sharded-cluster-with-database-dumps/). Then follow the rest of the procedure as described above.

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: sample-mgo-sh-custom
  namespace: demo
spec:
  clientConfig:
    service:
      name: sample-mgo-sh
      port: 27017
      scheme: mongodb
  secret:
    name: sample-mgo-sh-auth
  type: kubedb.com/mongodb

---
apiVersion: stash.appscode.com/v1alpha1
kind: Repository
metadata:
  name: gcs-repo-custom
  namespace: demo
spec:
  backend:
    gcs:
      bucket: appscode-qa
      prefix: demo/mongodb/sample-mgo-sh/standalone
    storageSecretName: gcs-secret

---
apiVersion: stash.appscode.com/v1beta1
kind: BackupConfiguration
metadata:
  name: sample-mgo-sh-backup2
  namespace: demo
spec:
  schedule: "*/5 * * * *"
  task:
    name: mongodb-backup-4.4.6
  repository:
    name: gcs-repo-custom
  target:
    ref:
      apiVersion: appcatalog.appscode.com/v1alpha1
      kind: AppBinding
      name: sample-mgo-sh-custom
  retentionPolicy:
    name: keep-last-5
    keepLast: 5
    prune: true
```

This time, we have to provide Stash addon info in `spec.task` section of `BackupConfiguration` object as the `AppBinding` we are creating manually does not have those info.

```console
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mongodb/backup/logical/sharding/examples/standalone-backup.yaml
appbinding.appcatalog.appscode.com/sample-mgo-sh-custom created
repository.stash.appscode.com/gcs-repo-custom created
backupconfiguration.stash.appscode.com/sample-mgo-sh-backup2 created


$ kubectl get backupsession -n demo
NAME                              BACKUPCONFIGURATION    PHASE       AGE
sample-mgo-sh-backup-1563528902   sample-mgo-sh-backup   Succeeded   35s


$ kubectl get repository -n demo gcs-repo-custom
NAME              INTEGRITY   SIZE         SNAPSHOT-COUNT   LAST-SUCCESSFUL-BACKUP   AGE
gcs-repo-custom   true        22.160 KiB   4                1m                       2m
```

### Restore to a standalone database

No additional configuration is needed to restore the sharded cluster to a standalone database. Follow the normal procedure of restoring a MongoDB Database.

Standalone MongoDB,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  name: restored-mongodb
  namespace: demo
spec:
  version: "4.4.26"
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  init:
    waitForInitialRestore: true
  terminationPolicy: WipeOut
```

This time, we have to provide `spec.task` section in `RestoreSession` object,

```yaml
apiVersion: stash.appscode.com/v1beta1
kind: RestoreSession
metadata:
  name: sample-mongodb-restore
  namespace: demo
spec:
  task:
    name: mongodb-restore-4.4.6
  repository:
    name: gcs-repo-custom
  target:
    ref:
      apiVersion: appcatalog.appscode.com/v1alpha1
      kind: AppBinding
      name: restored-mongodb
  rules:
  - snapshots: [latest]
```

```console
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mongodb/backup/logical/sharding/examples/restored-standalone.yaml
mongodb.kubedb.com/restored-mongodb created

$ kubectl get mg -n demo restored-mongodb
NAME               VERSION       STATUS         AGE
restored-mongodb   4.4.26         Provisioning   56s

$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mongodb/backup/logical/sharding/examples/restoresession-standalone.yaml
restoresession.stash.appscode.com/sample-mongodb-restore created

$ kubectl get mg -n demo restored-mongodb
NAME               VERSION       STATUS  AGE
restored-mongodb   4.4.26         Ready   56s
```

Now, exec into the database pod and list available tables,

```console
$ export USER=$(kubectl get secrets -n demo restored-mongodb-auth -o jsonpath='{.data.\username}' | base64 -d)

$ export PASSWORD=$(kubectl get secrets -n demo restored-mongodb-auth -o jsonpath='{.data.\password}' | base64 -d)

$ kubectl exec -it -n demo restored-mongodb-0 -- mongo admin -u $USER -p $PASSWORD

> show dbs
admin   0.000GB
config  0.000GB
local   0.000GB
newdb   0.000GB

> show users
{
	"_id" : "admin.root",
	"userId" : UUID("98fa7511-2ae0-4466-bb2a-f9a7e17631ad"),
	"user" : "root",
	"db" : "admin",
	"roles" : [
		{
			"role" : "root",
			"db" : "admin"
		}
	]
}

> use newdb
switched to db newdb

> db.movie.find().pretty()
{ "_id" : ObjectId("5d3064bf144a1b8fda04cd4f"), "name" : "batman" }

> exit
bye
```

So, from the above output, we can see the database `newdb` that we had created in the original database `sample-mgo-sh` is restored in the restored database `restored-mongodb`.

## Cleanup

To cleanup the Kubernetes resources created by this tutorial, run:

```console
kubectl delete -n demo restoresession sample-mgo-sh-restore sample-mongodb-restore
kubectl delete -n demo backupconfiguration sample-mgo-sh-backup sample-mgo-sh-backup2
kubectl delete -n demo mg sample-mgo-sh restored-mongodb
kubectl delete -n demo repository gcs-repo-sharding gcs-repo-custom
```
