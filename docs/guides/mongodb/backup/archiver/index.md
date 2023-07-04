---
title: MongoDB Backup and Recovery using MongoDBArchiver
menu:
  docs_{{ .version }}:
    identifier: guides-mongodb-backup-mg-archiver
    name: MongoDBArchiver
    parent: guides-mongodb-backup
    weight: 11
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MongoDB Backup & Restore using MongoDBArchiver
This guide will show you how to use `KubeDB` to backup and restore your database using MongoDBArchiver.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` Provisioner, Ops-manager operator in your cluster following the steps [here](/docs/setup/README.md).

- Install `KubeStash` from [here](https://kubestash.com)

- Install `External-Snapshotter` from [here](https://github.com/kubernetes-csi/external-snapshotter#usage)

- Install `longhorn` from [here](https://longhorn.io/docs/1.4.2/deploy/install/install-with-helm/)

- You should be familiar with the following `KubeDB` concepts:
    - [MongoDB](/docs/guides/mongodb/concepts/mongodb.md)
    - [MongoDBArchiver](/docs/guides/mongodb/concepts/mongodbarchiver.md)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

## Backup MongoDB using MongoDBArchiver

First, we are going to need a BackupStorage and RetentionPolicy which are part of the KubeStash project. BackupStorage is used to define our remote storage and RetentionPolicy is used to define how long our backup will be stored.

Let's deploy The BackupStorage using the yaml below.
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
      bucket: mg-test
      endpoint: https://us-southeast-1.linodeobjects.com
      region: us-southeast-1
      prefix: backup
      secret: storage
  usagePolicy:
    allowedNamespaces:
      from: All
  deletionPolicy: WipeOut
```

Here, 
- `spec.storage.provider` specifies the provider as `s3`.
- `spec.storage.s3` specifies different s3 options and `spec.storage.s3.secret` specifies a secret named `storage` which contains the secret keys to connect with our s3 bucket.
- `spec.deletionPolicy` specifies what will happen if the backup storage is deleted. As we have specified WipeOut, all data will be deleted when we delete the backup storage.

As you can see, we need a storage secret to connect to the backup storage. So, we will create that secret first.

```yaml
apiVersion: v1
kind: Secret
type: Opaque
metadata:
  name: storage
  namespace: demo
stringData:
  AWS_ACCESS_KEY_ID: "<AWS_ACCESS_KEY_ID>"
  AWS_SECRET_ACCESS_KEY: "<AWS_SECRET_ACCESS_KEY>"
  AWS_ENDPOINT: "https://us-southeast-1.linodeobjects.com"
```

Let's replace the <AWS_ACCESS_KEY_ID> and <AWS_SECRET_ACCESS_KEY> with their appropriate values and save the yaml as `storage-secret.yaml` and apply it.

```bash
$ kubectl apply -f storage-secret.yaml
secret/storage created
```

Now, Let’s deploy the above `BackupStorage` CRO

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mongob/backup/archiver/examples/backupstorage.yaml
backupstorage.storage.kubestash.com/linode-storage created
```

Now, we will deploy the retention policy yaml:
```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: RetentionPolicy
metadata:
  name: mongodb-retention-policy
  namespace: demo
spec:
  maxRetentionPeriod: "30d"
  successfulSnapshots:
    last: 100
    hourly: 12
    daily: 100
    weekly: 1000
    monthly: 5000
    yearly: 100000
  failedSnapshots:
    last: 5
```

Here,
- `spec.maxRetentionPeriod` specifies how long the data will be stored, we want to store it for 30 days.
- `spec.successfulSnapshots` specifies how many snapshots will be kept for different time period such as hourly, daily, weekly etc.
- `spec.failedSnapshots` specifies how many failed snapshots will be kept.

Now, Let’s create the above `RetentionPolicy` CRO

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mongob/backup/archiver/examples/retention-policy.yaml
retentionpolicy.storage.kubestash.com/mongodb-retention-policy created
```

Now, let's deploy the MongoDB Archiver yaml:
```yaml
apiVersion: archiver.kubedb.com/v1alpha1
kind: MongoDBArchiver
metadata:
  name: mongodbarchiver-sample
  namespace: demo
spec:
  pause: false
  databases:
    namespaces:
      from: "Same"
    selector:
      matchLabels:
        archiver: "true"
  retentionPolicy:
    name: mongodb-retention-policy
    namespace: demo
  fullBackup:
    driver: "CSISnapshotter"
    csiSnapshotter:
      volumeSnapshotClassName: "longhorn-snapshot-vsc"
    scheduler:
      schedule: "*/5 * * * *"
  manifestBackup:
    encryptionSecret:
      name: "encrypt-secret"
      namespace: "demo"
    scheduler:
      schedule: "*/3 * * * *"
  backupStorage:
    ref:
      apiGroup: "storage.kubestash.com"
      kind: "BackupStorage"
      name: "linode-storage"
      namespace: "demo"
```

Here,
- `spec.pause` specifies if the MongoDB Archiver is currently paused or not. If paused, no backup will be taken.
- `spec.databases` specifies the selectors and namespaces by which the mongodb databases will be selected.
- `spec.retentionPolicy` specifies the retention policy name and namespace that we have just created.
- `spec.fullBackup` specifies the full database backup options.
- `spec.manifestBackup` specifies the manifest such as storage secret, config secret etc. backup options.
- `spec.backupStorage` specifies the backup storage apiGroup, kind, name and namespace that we have just created.

`spec.fullBackup` has the following options:
- `driver` specifies the driver we are using for full backup. We are using the CSISnapshotter driver.
- `csiSnapshotter` specifies the csiSnapshotter driver options such as the volume snapshot class name.
- `scheduler` specifies the scheduler options for the full database backup such as the schedule, job template etc.
- `containerRuntimeSettings` specifies the container runtime settings for the full backup. For more information check [here](https://github.com/kmodules/offshoot-api/blob/master/api/v1/runtime_settings_types.go#L122-L173).
- `jobTemplate` specifies the job template that is used in the backup session created for the full backup. For more information check [here](https://github.com/kmodules/offshoot-api/blob/master/api/v1/types.go#L42-L57).
- `retryConfig` specifies the behavior of the retry.
- `timeout` specifies the timeout for the backup.
- `sessionHistoryLimit` specifies how many backup Jobs and associate resources Stash should keep for debugging purpose.


`spec.manifestBackup` has the following options:
- `encryptionSecret` specifies the secret name and namespace which is used to encrypt the data backed up.
- `scheduler` specifies the scheduler options for the full database backup such as the schedule, job template etc.

Now, we are using longhorn to take full backup of our database. So, we need a volume snapshot class to take the snapshot. We can see from the previous yaml that we have specified the volumeSnapshotClassName in `spec.fullBackup.csiSnapshotter`. So, before deploying our MongoDBArchiver we need to create this. Below is the yaml of the VolumeSnapshotClass that we are going to deploy:
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

Let’s create the above `VolumeSnapshotClass` CRO

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mongob/backup/archiver/examples/volumesnapshotclass.yaml
volumesnapshotclass.snapshot.storage.k8s.io/longhorn-snapshot-vsc created
```

Also, we can see from the archiver yaml that we have specified the encryption secret via `spec.manifestBackup.encryptionSecret`. So, before deploying our MongoDBArchiver we need to create this secret too. Below is the yaml of the secret that we are going to deploy:
```yaml
apiVersion: v1
kind: Secret
type: Opaque
metadata:
  name: encrypt-secret
  namespace: demo 
stringData:
  RESTIC_PASSWORD: "testpass"
```

Let’s create the above secret

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mongob/backup/archiver/examples/encryption-secret.yaml
secret/encrypt-secret created
```

Now, we have deployed all the supporting resources. Finally, we can deploy the mongodb archiver yaml.

Let’s create the above `MongoDBArchiver` CRO

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mongob/backup/archiver/examples/mongodbarchiver.yaml
mongodbarchiver.archiver.kubedb.com/mongodbarchiver-sample created
```

So, we have successfully deployed our mongodb archiver.

Now, we will create a MongoDB database that will be selected by this mongodb archiver. Below is the yaml of the MongoDB:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  name: mg-rs
  namespace: demo
  labels:
    archiver: "true"
spec:
  version: "4.4.6"
  replicaSet:
    name: "rs" 
  replicas: 3
  storage:
    storageClassName: longhorn
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
```


Let’s create the above `MongoDB` CR

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mongob/backup/archiver/examples/mg-rs.yaml
mongodb.kubedb.com/mg-rs created
```

Now, wait until `mg-rs` has status `Ready`. i.e,

```bash
$ kubectl get mg -n demo                                                                                                                                            
NAME    VERSION   STATUS   AGE
mg-rs   4.4.6     Ready    19m
```


**Insert Sample Data:**

Now, we are going to exec into the database pod and create some sample data. At first, find out the database pod using the following command,

```bash
$ kubectl get pods -n demo --selector="app.kubernetes.io/instance=mg-rs"
NAME      READY   STATUS    RESTARTS   AGE
mg-rs-0   2/2     Running   0          2m3s
mg-rs-1   2/2     Running   0          85s
mg-rs-2   2/2     Running   0          43s
```

Now, let's wait for a few minutes so that the first backup sessions for both manifest & full backup is executed. We can check that by watching the backupsession CRD.

```bash
$ watch kubectl get backupsessions -n demo
NAME                        INVOKER-TYPE          INVOKER-NAME   PHASE       DURATION   AGE
mg-rs-full-1687796400       BackupConfiguration   mg-rs          Succeeded              39s
mg-rs-manifest-1687796400   BackupConfiguration   mg-rs          Succeeded              40s
```

We can see from the above output that, both full and manifest backups are Succeeded.

Now, let's exec into the pod, create a db and collection, and insert some data:

```bash
$ kubectl exec -it -n demo mg-rs-0 -- bash

root@mg-rs-0:/# mongo -u $MONGO_INITDB_ROOT_USERNAME -p $MONGO_INITDB_ROOT_PASSWORD

rs:PRIMARY> rs.isMaster().primary
mg-rs-0.mg-rs-pods.demo.svc.cluster.local:27017

rs:PRIMARY> show dbs
admin          0.000GB
config         0.000GB
kubedb-system  0.000GB
local          0.000GB

rs:PRIMARY> use newdb
switched to db newdb

rs:PRIMARY> for(var i = 1; i <= 100; i++ ) { db.movie.insert({ "name": "movie"+i }) }
WriteResult({ "nInserted" : 1 })

rs:PRIMARY> db.movie.find()
{ "_id" : ObjectId("64952882d96667e772f3b6e4"), "name" : "movie1" }
{ "_id" : ObjectId("64952882d96667e772f3b6e5"), "name" : "movie2" }
{ "_id" : ObjectId("64952882d96667e772f3b6e6"), "name" : "movie3" }
{ "_id" : ObjectId("64952882d96667e772f3b6e7"), "name" : "movie4" }
{ "_id" : ObjectId("64952882d96667e772f3b6e8"), "name" : "movie5" }
{ "_id" : ObjectId("64952882d96667e772f3b6e9"), "name" : "movie6" }
{ "_id" : ObjectId("64952882d96667e772f3b6ea"), "name" : "movie7" }
{ "_id" : ObjectId("64952882d96667e772f3b6eb"), "name" : "movie8" }
{ "_id" : ObjectId("64952882d96667e772f3b6ec"), "name" : "movie9" }
{ "_id" : ObjectId("64952882d96667e772f3b6ed"), "name" : "movie10" }
{ "_id" : ObjectId("64952882d96667e772f3b6ee"), "name" : "movie11" }
{ "_id" : ObjectId("64952882d96667e772f3b6ef"), "name" : "movie12" }
{ "_id" : ObjectId("64952882d96667e772f3b6f0"), "name" : "movie13" }
{ "_id" : ObjectId("64952882d96667e772f3b6f1"), "name" : "movie14" }
{ "_id" : ObjectId("64952882d96667e772f3b6f2"), "name" : "movie15" }
{ "_id" : ObjectId("64952882d96667e772f3b6f3"), "name" : "movie16" }
{ "_id" : ObjectId("64952882d96667e772f3b6f4"), "name" : "movie17" }
{ "_id" : ObjectId("64952882d96667e772f3b6f5"), "name" : "movie18" }
{ "_id" : ObjectId("64952882d96667e772f3b6f6"), "name" : "movie19" }
{ "_id" : ObjectId("64952882d96667e772f3b6f7"), "name" : "movie20" }
Type "it" for more

rs:PRIMARY> db.movie.count()
100

rs:PRIMARY> exit
bye

root@mg-rs-0:/# date --rfc-3339=seconds | sed 's/ /T/'
2023-06-26T16:28:49+00:00

root@mg-rs-0:/# exit
exit
```

So, we've inserted `100` documents in a db called `newdb`. Now, we are ready to restore the database using the backups taken by the mongodb archiver.

## Point in time recovery using MongoDBArchiver

Now, we already marked the time after we've inserted our documents. So, we'll use that time to create a new database to point-in-time-recover the data. 

Let's deploy a new mongodb which will have the data till the time we specify, which is the time after inserting the data in the `newdb` database in our previous mongodb.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  name: mg-rs-restored
  namespace: demo
spec:
  version: "4.4.6"
  replicaSet:
    name: "rs"
  replicas: 3
  podTemplate:
    spec:
      resources:
        requests:
          cpu: "500m"
          memory: "500Mi"
  storage:
    storageClassName: longhorn
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  init:
    archiver:
      recoveryTimestamp: "2023-06-26T16:28:58+00:00"
      fullDBRestore:
        repository:
          name: "mg-rs-full"
          namespace: "demo"
      manifestRestore:
        repository:
          name: "mg-rs-manifest"
          namespace: "demo"
        encryptionSecret:
          name: "encrypt-secret"
          namespace: "demo"
  terminationPolicy: "WipeOut"
```

Here, we can see that, the mongodb yaml is almost similar with our previous mongodb yaml, except it has an `init` section. In this section, we have specified that we want to point-in-time-recover the database.
In the `init` section,
- `archiver.recoveryTimestamp` specifies the timestamp in which we want to point-in-time-recover our database.
- `archiver.fullDBRestore` specifies the repository in which the full backup exists.
- `archiver.manifestRestore` specifies the repository in which the manifest backup exists and the encryption secret that was used to encrypt the data.

Let’s create the above `MongoDB` CRO

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mongob/backup/archiver/examples/mg-rs-restored.yaml
mongodb.kubedb.com/mg-rs-restored created
```

Let's wait until `mg-rs-restored` has status `Ready`. i.e,

```bash
$ kubectl get mg -n demo                                                                                                                                            
NAME              VERSION   STATUS   AGE
mg-rs-restored    4.4.6     Ready    10m
```

Now, we'll check the documents in the `newdb` database to see if it has all the data till the time we specified. So, it should have `100` documents in the `movie` collections under the `newdb` database. Now, let's exec into the pod of `mg-rs-restored` and check the data:

```bash
$ kubectl exec -it -n demo mg-rs-restored-0 -- bash

root@mg-rs-restored-0:/# mongo -u $MONGO_INITDB_ROOT_USERNAME -p $MONGO_INITDB_ROOT_PASSWORD 

rs:PRIMARY> show dbs
admin          0.000GB
config         0.000GB
kubedb-system  0.000GB
local          0.001GB
newdb          0.000GB

rs:PRIMARY> use newdb
switched to db newdb

rs:PRIMARY> show collections
movie

rs:PRIMARY> db.movie.find()
{ "_id" : ObjectId("6499bc948886197db9bfd4a2"), "name" : "movie1" }
{ "_id" : ObjectId("6499bc948886197db9bfd4a3"), "name" : "movie2" }
{ "_id" : ObjectId("6499bc948886197db9bfd4a4"), "name" : "movie3" }
{ "_id" : ObjectId("6499bc948886197db9bfd4a5"), "name" : "movie4" }
{ "_id" : ObjectId("6499bc948886197db9bfd4a6"), "name" : "movie5" }
{ "_id" : ObjectId("6499bc948886197db9bfd4a7"), "name" : "movie6" }
{ "_id" : ObjectId("6499bc948886197db9bfd4a8"), "name" : "movie7" }
{ "_id" : ObjectId("6499bc948886197db9bfd4a9"), "name" : "movie8" }
{ "_id" : ObjectId("6499bc948886197db9bfd4aa"), "name" : "movie9" }
{ "_id" : ObjectId("6499bc948886197db9bfd4ab"), "name" : "movie10" }
{ "_id" : ObjectId("6499bc948886197db9bfd4ac"), "name" : "movie11" }
{ "_id" : ObjectId("6499bc948886197db9bfd4ad"), "name" : "movie12" }
{ "_id" : ObjectId("6499bc948886197db9bfd4ae"), "name" : "movie13" }
{ "_id" : ObjectId("6499bc948886197db9bfd4af"), "name" : "movie14" }
{ "_id" : ObjectId("6499bc948886197db9bfd4b0"), "name" : "movie15" }
{ "_id" : ObjectId("6499bc948886197db9bfd4b1"), "name" : "movie16" }
{ "_id" : ObjectId("6499bc948886197db9bfd4b2"), "name" : "movie17" }
{ "_id" : ObjectId("6499bc948886197db9bfd4b3"), "name" : "movie18" }
{ "_id" : ObjectId("6499bc948886197db9bfd4b4"), "name" : "movie19" }
{ "_id" : ObjectId("6499bc948886197db9bfd4b5"), "name" : "movie20" }
Type "it" for more

rs:PRIMARY> db.movie.count()
100

rs:PRIMARY> exit
bye

root@mg-rs-restored-0:/# exit
exit
```
We can see from the above output that we have successfully done the point-in-time-recovery of the database `mg-rs`.

## Next Steps

- Backup a standalone MongoDB databases using Stash following the guide from [here](/docs/guides/mongodb/backup/logical/standalone/index.md).
- Backup a MongoDB Replicaset cluster using Stash following the guide from [here](/docs/guides/mongodb/backup/logical/replicaset/index.md).
- Backup a sharded MongoDB cluster using Stash following the guide from [here](/docs/guides/mongodb/backup/logical/sharding/index.md).
