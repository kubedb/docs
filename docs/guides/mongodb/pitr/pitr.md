---
title: Continuous Archiving and Point-in-time Recovery
menu:
  docs_{{ .version }}:
    identifier: pitr-mongo
    name: Overview
    parent: mg-archiver-pitr
    weight: 42
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# KubeDB MongoDB - Continuous Archiving and Point-in-time Recovery

Here, this doc will show you how to use KubeDB to provision a MongoDB to Archive continuously and Restore point-in-time.

## Before You Begin

At first, you need to have a Kubernetes cluster, and the kubectl command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using [kind](https://kind.sigs.k8s.io/docs/user/quick-start/).

Now,
- Install `KubeDB` operator in your cluster following the steps [here](/docs/setup/README.md).
- Install `KubeStash` operator in your cluster following the steps [here](https://github.com/kubestash/installer/tree/master/charts/kubestash).
- Install `SideKick`  in your cluster following the steps [here](https://github.com/kubeops/installer/tree/master/charts/sidekick).
- Install `External-snapshotter`  in your cluster following the steps [here](https://github.com/kubernetes-csi/external-snapshotter/tree/release-5.0), if you don't already have a csi-driver available in the cluster.

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```
> Note: The yaml files used in this tutorial are stored in [mg-archiver-demo](https://github.com/kubedb/mg-archiver-demo)
## Continuous archiving
Continuous archiving involves making regular copies (or "archives") of the MongoDB transaction log files. To ensure continuous archiving to a remote location we need to prepare `BackupStorage`,`RetentionPolicy`,`MongoDBArchiver` for the KubeDB Managed MongoDB Databases.


### BackupStorage
BackupStorage is a CR provided by KubeStash that can manage storage from various providers like GCS, S3, and more.

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: BackupStorage
metadata:
  name: gcs-storage
  namespace: demo
spec:
  storage:
    provider: gcs
    gcs:
      bucket: kubestash-qa
      prefix: mg
      secretName: gcs-secret
  usagePolicy:
    allowedNamespaces:
      from: All
  deletionPolicy: WipeOut # One of: WipeOut, Delete
```

For s3 compatible buckets, the `.spec.storage` will be like this : 
```yaml 
provider: s3
s3:
  endpoint: us-east-1.linodeobjects.com
  bucket: arnob
  region: us-east-1
  prefix: ya
  secret: linode-secret 
```

```bash
   $ kubectl apply -f https://raw.githubusercontent.com/kubedb/mg-archiver-demo/master/gke/backupstorage.yaml
   backupstorage.storage.kubestash.com/gcs-storage created
```

### Secret for BackupStorage

You need to create a credentials which will hold the information about cloud bucket. Here are examples.

For GCS :
```bash
kubectl create secret generic -n demo gcs-secret \
  --from-literal=GOOGLE_PROJECT_ID=<your-project-id> \
  --from-file=./GOOGLE_SERVICE_ACCOUNT_JSON_KEY
```

For S3 :
```bash 
kubectl create secret generic -n demo s3-secret \
    --from-file=./AWS_ACCESS_KEY_ID \
    --from-file=./AWS_SECRET_ACCESS_KEY
```

```bash
  $ kubectl apply -f https://raw.githubusercontent.com/kubedb/mg-archiver-demo/master/gke/storage-secret.yaml
  secret/gcs-secret created
```

### Retention policy
RetentionPolicy is a CR provided by KubeStash that allows you to set how long you'd like to retain the backup data.
```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: RetentionPolicy
metadata:
  name: mongodb-retention-policy
  namespace: demo
spec:
  maxRetentionPeriod: "30d"
  successfulSnapshots:
    last: 5
  failedSnapshots:
    last: 2
```
```bash
$ kubectl apply -https://raw.githubusercontent.com/kubedb/mg-archiver-demo/master/common/retention-policy.yaml
retentionpolicy.storage.kubestash.com/mongodb-retention-policy created
```


## Ensure volumeSnapshotClass

```bash
kubectl get volumesnapshotclasses
NAME                    DRIVER               DELETIONPOLICY   AGE
longhorn-snapshot-vsc   driver.longhorn.io   Delete           7d22h

```
If not any, try using `longhorn` or any other [volumeSnapshotClass](https://kubernetes.io/docs/concepts/storage/volume-snapshot-classes/).

```bash
$ helm install longhorn longhorn/longhorn --namespace longhorn-system --create-namespace
  ...
  ...
  kubectl get pod -n longhorn-system
````


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

If you already have a csi driver installed in your cluster, You need to refer that in the `.driver` section.  Here is an example for GCS :

```yaml
apiVersion: snapshot.storage.k8s.io/v1
kind: VolumeSnapshotClass
metadata:
  name: gke-vsc
driver: pd.csi.storage.gke.io
deletionPolicy: Delete
```


```bash
$ kubectl apply -f https://raw.githubusercontent.com/kubedb/mg-archiver-demo/master/gke/volume-snapshot-class.yaml
  volumesnapshotclass.snapshot.storage.k8s.io/gke-vsc unchanged
```


### MongoDBArchiver
MongoDBArchiver is a CR provided by KubeDB for managing the archiving of MongoDB oplog files and performing volume-level backups

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
  encryptionSecret:
    name: encrypt-secret
    namespace: demo
  fullBackup:
    driver: VolumeSnapshotter
    task:
      params:
        volumeSnapshotClassName: gke-vsc  # change it accordingly
    scheduler:
      successfulJobsHistoryLimit: 1
      failedJobsHistoryLimit: 1
      schedule: "*/50 * * * *"
    sessionHistoryLimit: 2
  manifestBackup:
    scheduler:
      successfulJobsHistoryLimit: 1
      failedJobsHistoryLimit: 1
      schedule: "*/5 * * * *"
    sessionHistoryLimit: 2
  backupStorage:
    ref:
      name: gcs-storage
      namespace: demo

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
 $ kubectl create -f https://raw.githubusercontent.com/kubedb/mg-archiver-demo/master/common/encrypt-secret.yaml
 $ kubectl create -f https://raw.githubusercontent.com/kubedb/mg-archiver-demo/master/common/archiver.yaml
```


# Deploy MongoDB
So far we are ready with setup for continuously archive MongoDB, We deploy a MongoDB referring the MongoDB archiver object

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  name: mg-rs
  namespace: demo
  labels:
    archiver: "true"
spec:
  version: "4.4.26"
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
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi

```

The `archiver: "true"` label is important here. Because that's how we are specifying that continous archiving will be done in this db.


```bash
$ kubectl get pod -n demo
NAME                                                  READY   STATUS      RESTARTS   AGE
mg-rs-0                                               2/2     Running     0          8m30s
mg-rs-1                                               2/2     Running     0          7m32s
mg-rs-2                                               2/2     Running     0          6m34s
mg-rs-backup-full-backup-1702457252-lvcbn             0/1     Completed   0          65s
mg-rs-backup-manifest-backup-1702457110-fjpw5         0/1     Completed   0          3m28s
mg-rs-backup-manifest-backup-1702457253-f4chq         0/1     Completed   0          65s
mg-rs-sidekick                                        1/1     Running     0          5m29s
trigger-mg-rs-backup-manifest-backup-28374285-rdcfq   0/1     Completed   0          3m38s

```
`mg-rs-sidekick` is responsible for uploading oplog-files
`mg-rs-full-backup-*****` are the volumes levels backups for MongoDB.
`mg-rs-manifest-backup-*****` are the backups of the manifest relate to MongoDB object

### Validate BackupConfiguration and VolumeSnapshot

```bash
$ kubectl get backupstorage,backupconfigurations,backupsession,volumesnapshots -A

NAMESPACE   NAME                                              PROVIDER   DEFAULT   DELETION-POLICY   TOTAL-SIZE   PHASE   AGE
demo        backupstorage.storage.kubestash.com/gcs-storage   gcs                  WipeOut           3.292 KiB    Ready   11m

NAMESPACE   NAME                                                  PHASE   PAUSED   AGE
demo        backupconfiguration.core.kubestash.com/mg-rs-backup   Ready            6m45s

NAMESPACE   NAME                                                                       INVOKER-TYPE          INVOKER-NAME   PHASE       DURATION   AGE
demo        backupsession.core.kubestash.com/mg-rs-backup-full-backup-1702457252       BackupConfiguration   mg-rs-backup   Succeeded              2m20s
demo        backupsession.core.kubestash.com/mg-rs-backup-manifest-backup-1702457110   BackupConfiguration   mg-rs-backup   Succeeded              4m43s
demo        backupsession.core.kubestash.com/mg-rs-backup-manifest-backup-1702457253   BackupConfiguration   mg-rs-backup   Succeeded              2m20s

NAMESPACE   NAME                                                      READYTOUSE   SOURCEPVC         SOURCESNAPSHOTCONTENT   RESTORESIZE   SNAPSHOTCLASS   SNAPSHOTCONTENT                                    CREATIONTIME   AGE
demo        volumesnapshot.snapshot.storage.k8s.io/mg-rs-1702457262   true         datadir-mg-rs-1                           1Gi           gke-vsc         snapcontent-87f1013f-cd7e-4153-b245-da9552d2e44f   2m7s           2m11s

```

## data insert and switch oplog
After each and every oplog switch the oplog files will be uploaded to backup storage
```bash
$ kubectl exec -it -n demo mg-rs-0 bash
kubectl exec [POD] [COMMAND] is DEPRECATED and will be removed in a future version. Use kubectl exec [POD] -- [COMMAND] instead.
Defaulted container "mongodb" out of: mongodb, replication-mode-detector, copy-config (init)
mongodb@mg-rs-0:/$ 
mongodb@mg-rs-0:/$ mongo -u root -p $MONGO_INITDB_ROOT_PASSWORD 
MongoDB shell version v4.4.26
connecting to: mongodb://127.0.0.1:27017/?compressors=disabled&gssapiServiceName=mongodb
Implicit session: session { "id" : UUID("4a51b9fc-a26c-487b-848d-341cf5512c86") }
MongoDB server version: 4.4.26
Welcome to the MongoDB shell.
For interactive help, type "help".
For more comprehensive documentation, see
	https://docs.mongodb.com/
Questions? Try the MongoDB Developer Community Forums
	https://community.mongodb.com
---
The server generated these startup warnings when booting: 
        2023-12-13T08:40:40.423+00:00: Using the XFS filesystem is strongly recommended with the WiredTiger storage engine. See http://dochub.mongodb.org/core/prodnotes-filesystem
---
rs:PRIMARY> show dbs
admin          0.000GB
config         0.000GB
kubedb-system  0.000GB
local          0.000GB
rs:PRIMARY> use pink_floyd
switched to db pink_floyd
rs:PRIMARY> db.songs.insert({"name":"shine on you crazy diamond"})
WriteResult({ "nInserted" : 1 })
rs:PRIMARY> show collections
songs
rs:PRIMARY> db.songs.find()
{ "_id" : ObjectId("657970c1f965be0513c7f4d7"), "name" : "shine on you crazy diamond" }
rs:PRIMARY> 
```
> At this point We have a document in our newly created collection `songs` on database `pink_floyd`
## Point-in-time Recovery
Point-In-Time Recovery allows you to restore a MongoDB database to a specific point in time using the archived transaction logs. This is particularly useful in scenarios where you need to recover to a state just before a specific error or data corruption occurred.
Let's say accidentally our dba drops the the table tab_1 and we want to restore.
```bash
```bash

rs:PRIMARY> use pink_floyd
switched to db pink_floyd

rs:PRIMARY> db.dropDatabase()
{
	"dropped" : "pink_floyd",
	"ok" : 1,
	"$clusterTime" : {
		"clusterTime" : Timestamp(1702457742, 2),
		"signature" : {
			"hash" : BinData(0,"QFpwWOtec/NdQ0iKKyFCx9Jz8/A="),
			"keyId" : NumberLong("7311996497896144901")
		}
	},
	"operationTime" : Timestamp(1702457742, 2)
}


```

Time time `1702457742` is unix timestamp. This is `Wed Dec 13 2023 08:55:42 GMT+0000` in human readable format.
We can't restore from a full backup since at this point no full backup was perform. so we can choose a specific time in (just before this timestamp, for example `08:55:30`) which time we want to restore.

### Restore MongoDB
```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  name: mg-rs-restored
  namespace: demo
spec:
  version: "4.4.26"
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
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  init:
    archiver:
      recoveryTimestamp: "2023-12-13T08:55:30Z"
      encryptionSecret:
        name: encrypt-secret
        namespace: demo
      fullDBRepository:
        name: mg-rs-full
        namespace: demo
      manifestRepository:
        name: mg-rs-manifest
        namespace: demo
  deletionPolicy: WipeOut

```
```bash
kubectl apply -f restore.yaml
mongo.kubedb.com/restore-mg created
```
**check for Restored MongoDB**
```bash
 kubectl get pods -n demo | grep restore
mg-rs-restored-0                                      2/2     Running     0          4m43s
mg-rs-restored-1                                      2/2     Running     0          3m52s
mg-rs-restored-2                                      2/2     Running     0          2m59s
mg-rs-restored-manifest-restorer-2qb46                0/1     Completed   0          4m58s
mg-rs-restored-wal-restorer-nkxfl                     0/1     Completed   0          41s

```
```bash
$ kubectl get mg -n demo
NAME             VERSION   STATUS   AGE
mg-rs-restored   4.4.26    Ready    5m47s

```
**Validating data on Restored MongoDB**
```bash
$ kubectl exec -it -n demo mg-rs-restored-0 bash
mongodb@mg-rs-restored-0:/$ mongo -u root -p $MONGO_INITDB_ROOT_PASSWORD 
MongoDB shell version v4.4.26
connecting to: mongodb://127.0.0.1:27017/?compressors=disabled&gssapiServiceName=mongodb
Implicit session: session { "id" : UUID("50d3fc74-bffc-4c97-a1e6-a2ea63cb88e1") }
MongoDB server version: 4.4.26
Welcome to the MongoDB shell.
For interactive help, type "help".
For more comprehensive documentation, see
	https://docs.mongodb.com/
Questions? Try the MongoDB Developer Community Forums
	https://community.mongodb.com
---
The server generated these startup warnings when booting: 
        2023-12-13T09:05:42.205+00:00: Using the XFS filesystem is strongly recommended with the WiredTiger storage engine. See http://dochub.mongodb.org/core/prodnotes-filesystem
---
rs:PRIMARY> show dbs
admin          0.000GB
config         0.000GB
kubedb-system  0.000GB
local          0.000GB
pink_floyd     0.000GB
rs:PRIMARY> use pink_floyd
switched to db pink_floyd
rs:PRIMARY> show collections
songs
rs:PRIMARY> db.songs.find()
{ "_id" : ObjectId("657970c1f965be0513c7f4d7"), "name" : "shine on you crazy diamond" }

```
**so we are able to successfully recover from a disaster**
## Cleaning up
To cleanup the Kubernetes resources created by this tutorial, run:
```bash
kubectl delete -n demo mg/mg-rs
kubectl delete -n demo mg/mg-rs-restored
kubectl delete -n demo backupstorage/gcs-storage
kubectl delete ns demo
```
## Next Steps
- Learn about [backup and restore](/docs/guides/mongodb/backup/overview/index.md) MongoDB database using Stash.
- Learn about initializing [MongoDB with Script](/docs/guides/mongodb/initialization/using-script.md).
- Learn about [custom mongoVersions](/docs/guides/mongodb/concepts/catalog.md).
- Want to setup MongoDB cluster? Check how to [configure Highly Available MongoDB Cluster](/docs/guides/mongodb/clustering/replicaset.md)
- Monitor your MongoDB database with KubeDB using [built-in Prometheus](/docs/guides/mongodb/monitoring/using-builtin-prometheus.md).
- Monitor your MongoDB database with KubeDB using [Prometheus operator](/docs/guides/mongodb/monitoring/using-prometheus-operator.md).
- Detail concepts of [mongo object](/docs/guides/mongodb/concepts/mongodb.md).
- Use [private Docker registry](/docs/guides/mongodb/private-registry/using-private-registry.md) to deploy MongoDB with KubeDB.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).