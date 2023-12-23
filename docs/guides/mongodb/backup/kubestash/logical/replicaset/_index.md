---
title: Backup & Restore Replicaset MongoDB | KubeStash
description: Backup and restore replicaset MongoDB database using KubeStash
menu:
  docs_{{ .version }}:
    identifier: guides-mongodb-backup-kubestash-logical-replicaset
    name: MongoDB Replicaset Cluster
    parent: guides-mongodb-backup-kubestash-logical
    weight: 21
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Backup and Restore MongoDB database using KubeStash

KubeStash v0.1.0+ supports backup and restoration of MongoDB databases. This guide will show you how you can backup and restore your MongoDB database with KubeStash.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using Minikube.
- Install KubeDB in your cluster following the steps [here](/docs/setup/README.md).
- Install KubeStash Enterprise in your cluster following the steps [here(link needed)]().
- Install KubeStash `kubectl` plugin following the steps [here(link needed)]().
- If you are not familiar with how KubeStash backup and restore MongoDB databases, please check the following guide [here](/docs/guides/mongodb/backup/kubestash/overview/index.md).

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

## Backup MongoDB

This section will demonstrate how to backup MongoDB database. Here, we are going to deploy a MongoDB database using KubeDB. Then, we are going to backup this database into a S3 bucket. Finally, we are going to restore the backed up data into another MongoDB database.

### Deploy Sample MongoDB Database

Let's deploy a sample MongoDB database and insert some data into it.

**Create MongoDB CRD:**

Below is the YAML of a sample MongoDB crd that we are going to create for this tutorial:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  name: sample-mg-rs
  namespace: demo
spec:
  version: "4.2.3"
  replicas: 3
  replicaSet:
    name: rs0
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: WipeOut
```

Create the above `MongoDB` crd,

```console
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mongodb/backup/kubestash/logical/replicaset/examples/mongodb-replicaset.yaml
mongodb.kubedb.com/sample-mg-rs created
```

KubeDB will deploy a MongoDB database according to the above specification. It will also create the necessary secrets and services to access the database.

Let's check if the database is ready to use,

```console
$ kubectl get mongodb -n demo sample-mg-rs
NAME           VERSION   STATUS   AGE
sample-mg-rs   4.2.3     Ready    2m27s
```

The database is `Ready`. Verify that KubeDB has created a Secret and a Service for this database using the following commands,

```console
$ kubectl get secret -n demo -l=app.kubernetes.io/instance=sample-mg-rs
NAME                TYPE                       DATA   AGE
sample-mg-rs-auth   kubernetes.io/basic-auth   2      3m53s
sample-mg-rs-key    Opaque                     1      3m53s

$ kubectl get service -n demo -l=app.kubernetes.io/instance=sample-mg-rs
NAME                TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)     AGE
sample-mg-rs        ClusterIP   10.96.211.27   <none>        27017/TCP   4m38s
sample-mg-rs-pods   ClusterIP   None           <none>        27017/TCP   4m38s
```

Here, we have to use service `sample-mg-rs` and secret `sample-mg-rs-auth` to connect with the database.

**Insert Sample Data:**

> Note: You can insert data into this `MongoDB` database using our [KubeDB CLI](https://kubedb.com/docs/latest/setup/install/kubectl_plugin/).

For simplicity, we are going to exec into the database pod and create some sample data. At first, find out the database pod using the following command,

```console
$ kubectl get pods -n demo --selector="app.kubernetes.io/instance=sample-mg-rs"
NAME             READY   STATUS    RESTARTS   AGE
sample-mg-rs-0   2/2     Running   0          6m15s
sample-mg-rs-1   2/2     Running   0          5m36s
sample-mg-rs-2   2/2     Running   0          5m14s
```

Now, let's exec into the pod and create a table,

```console
$ export USER=$(kubectl get secrets -n demo sample-mg-rs-auth -o jsonpath='{.data.\username}' | base64 -d)

$ export PASSWORD=$(kubectl get secrets -n demo sample-mg-rs-auth -o jsonpath='{.data.\password}' | base64 -d)

$ kubectl exec -it -n demo sample-mg-rs-0 -- mongo admin -u $USER -p $PASSWORD

rs0:PRIMARY> show dbs
admin          0.000GB
config         0.000GB
kubedb-system  0.000GB
local          0.000GB

rs0:PRIMARY> show users
{
	"_id" : "admin.root",
	"userId" : UUID("12716b1d-9186-4710-a669-4a2050a548bc"),
	"user" : "root",
	"db" : "admin",
	"roles" : [
		{
			"role" : "root",
			"db" : "admin"
		}
	],
	"mechanisms" : [
		"SCRAM-SHA-1",
		"SCRAM-SHA-256"
	]
}

rs0:PRIMARY> use newdb
switched to db newdb

rs0:PRIMARY> db.movie.insert({"name":"batman"});
WriteResult({ "nInserted" : 1 })

rs0:PRIMARY> db.movie.find().pretty()
{ "_id" : ObjectId("657182194f1f295758f449a5"), "name" : "batman" }

rs0:PRIMARY> exit
bye

```

Now, we are ready to backup this sample database.

### Prepare Backend

We are going to store our backed up data into a S3 bucket. At first, we need to create a secret with S3 credentials then we need to create a `BackupStorage` crd. If you want to use a different backend, please read the respective backend configuration doc from [here](https://stash.run/docs/latest/guides/backends/overview/).

**Create Storage Secret:**

Let's create a secret called `s3-secret` with access credentials to our desired S3 bucket,

```console
$ echo -n '<your-aws-access-key-id-here>' > AWS_ACCESS_KEY_ID
$ echo -n '<your-aws-secret-access-key-here>' > AWS_SECRET_ACCESS_KEY
$ kubectl create secret generic -n demo s3-secret \
    --from-file=./AWS_ACCESS_KEY_ID \
    --from-file=./AWS_SECRET_ACCESS_KEY
secret/s3-secret created
```

**Create BackupStorage:**

Now, crete a `BackupStorage` using this secret. Below is the YAML of BackupStorage crd we are going to create,

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: BackupStorage
metadata:
  name: s3-storage-replicaset
  namespace: demo
spec:
  storage:
    provider: s3
    s3:
      endpoint: us-east-1.linodeobjects.com
      bucket: kubestash-testing
      region: us-east-1
      prefix: demo-replicaset
      secret: s3-secret
  usagePolicy:
    allowedNamespaces:
      from: All
  deletionPolicy: WipeOut
```

Let's create the `BackupStorage` we have shown above,

```console
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mongodb/backup/kubestash/logical/replicaset/examples/backupstorage-replicaset.yaml
backupstorage.storage.kubestash.com/s3-storage-replicaset created
```

Now, we are ready to backup our database to our desired backend.

### Backup

We have to create a `BackupConfiguration` targeting respective MongoDB crd of our desired database. Then KubeStash will create a CronJob to periodically backup the database. Before that we need to create an secret for encrypt data and retention policy.

**Create Encryption Secret:**

EncryptionSecret refers to the Secret containing the encryption key which will be used to encode/decode the backed up data. Let's create a secret called `encry-secret`

```console
$ kubectl create secret generic encry-secret -n demo \
    --from-literal=RESTIC_PASSWORD='123' -n demo
secret/encry-secret created
```

**Create Retention Policy:**

`RetentionPolicy` specifies how the old Snapshots should be cleaned up. This is a namespaced CRD.However, we can refer it from other namespaces as long as it is permitted via `.spec.usagePolicy`. Below is the YAML of the `RetentionPolicy` called `backup-rp`

```console
apiVersion: storage.kubestash.com/v1alpha1
kind: RetentionPolicy
metadata:
  name: backup-rp
  namespace: demo
spec:
  maxRetentionPeriod: 2mo
  successfulSnapshots:
    last: 10
  usagePolicy:
    allowedNamespaces:
      from: All
```

Let's create the RetentionPolicy we have shown above,

```console
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mongodb/backup/kubestash/logical/replicaset/examples/retentionpolicy.yaml
retentionpolicy.storage.kubestash.com/backup-rp created
```

**Create BackupConfiguration:**

As we just create our encryption secret and retention policy, we are now ready to apply `BackupConfiguration` crd to take backup out database.

Below is the YAML for `BackupConfiguration` crd to backup the `sample-mg-rs` database we have deployed earlier.,

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: mg
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: MongoDB
    namespace: demo
    name: sample-mg-rs
  backends:
    - name: s3-backend
      storageRef:
        namespace: demo
        name: s3-storage-replicaset
      retentionPolicy:
        name: backup-rp
        namespace: demo        
  sessions:
    - name: frequent
      scheduler:
        schedule: "*/3 * * * *"
      repositories:
        - name: s3-repo
          backend: s3-backend
          directory: /replicaset
          encryptionSecret:
           name: encry-secret
           namespace: demo
      addon:
        name: mongodb-addon
        tasks:
          - name: LogicalBackup
```

Here,

- `spec.target` specifies our targeted `MongoDB` database.
- `spec.backends` specifies `BackupStorage` information for storing data.
- `spec.sessions` specifies common session configurations for this backup
- `spec.sessions.schedule` specifies that we want to backup the database at 5 minutes interval.
- `spec.sessions.addon` refers to the `Addon` crd for backup task

Let's create the `BackupConfiguration` crd we have shown above,

```console
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mongodb/backup/kubestash/logical/replicaset/examples/backupconfiguration-replicaset.yaml
backupconfiguration.core.kubestash.com/mg created
```

**Verify Backup Setup Successful:**

If everything goes well, the phase of the `BackupConfiguration` should be `Ready`. The `Ready` phase indicates that the backup setup is successful. Let's verify the `Phase` of the BackupConfiguration,

```console
$ kubectl get backupconfiguration -n demo
NAME   PHASE   PAUSED   AGE
mg     Ready            85s
```

**Verify CronJob:**

KubeStash will create a CronJob with the schedule specified in `spec.sessions.schedule` field of `BackupConfiguration` crd.

Verify that the CronJob has been created using the following command,

```console
$ kubectl get cronjob -n demo
NAME                  SCHEDULE      SUSPEND   ACTIVE   LAST SCHEDULE   AGE
trigger-mg-frequent   */3 * * * *   False     0        <none>          101s
```

**Wait for BackupSession:**

The `trigger-mg-frequent` CronJob will trigger a backup on each schedule by creating a `BackpSession` crd.

Wait for the next schedule. Run the following command to watch `BackupSession` crd,

```console
$ kubectl get backupsession -n demo
NAME                     INVOKER-TYPE          INVOKER-NAME   PHASE       DURATION   AGE
mg-frequent-1701940862   BackupConfiguration   mg             Succeeded              3m16s
mg-frequent-1701941042   BackupConfiguration   mg             Running                16s
```

We can see above that the backup session has succeeded. Now, we are going to verify that the backed up data has been stored in the backend.

**Verify Backup:**

Once a backup is complete, KubeStash will update the respective `Snapshot` crd to reflect the backup. It will be created when a backup is triggered. Check that the `Snapshot` Phase to verify backup.

```console
$ kubectl get snapshot -n demo
NAME                             REPOSITORY   SESSION    SNAPSHOT-TIME          DELETION-POLICY   PHASE       VERIFICATION-STATUS   AGE
s3-repo-mg-frequent-1701940862   s3-repo      frequent   2023-12-07T09:21:07Z   Delete            Succeeded                         3m53s
s3-repo-mg-frequent-1701941042   s3-repo      frequent   2023-12-07T09:24:08Z   Delete            Succeeded                         53s
```

KubeStash will also update the respective `Repository` crd to reflect the backup. Check that the repository `s3-repo` has been updated by the following command,

```console
$ kubectl get repository -n demo s3-repo
NAME      INTEGRITY   SNAPSHOT-COUNT   SIZE        PHASE   LAST-SUCCESSFUL-BACKUP   AGE
s3-repo   true        2                2.883 KiB   Ready   55s                      8m5s
```

Now, if we navigate to the S3 bucket, we are going to see backed up data has been stored in `demo/replicaset/` directory as specified by `spec.sessions.repositories.directory` field of `BackupConfiguration` crd.

> Note: KubeStash keeps all the backed up data encrypted. So, data in the backend will not make any sense until they are decrypted.

## Restore MongoDB
You can restore your data into the same database you have backed up from or into a different database in the same cluster or a different cluster. In this section, we are going to show you how to restore in the same database which may be necessary when you have accidentally deleted any data from the running database.

#### Stop Taking Backup of the Old Database:

It's important to stop taking any further backup of the old database so that no backup is stored in our repository during restore process. KubeStash operator will automatically pause the `BackupConfiguration` when a `RestoreSession` is running. However if we want to pause the `BackupConfiguration` manually, we can do that by patching or using KubeStash CLI.

Let's pause the `mg` BackupConfiguration by patching,
```bash
$ kubectl patch backupconfiguration -n demo mg --type="merge" --patch='{"spec": {"paused": true}}'
backupconfiguration.core.kubestash.com/mg patched
```

Now, wait for a moment. KubeStash will pause the BackupConfiguration. Verify that the BackupConfiguration  has been paused,

```console
$ kubectl get backupconfiguration -n demo mg
NAME   PHASE   PAUSED   AGE
mg     Ready   true     11m
```

Notice the `PAUSED` column. Value `true` for this field means that the BackupConfiguration has been paused.

#### Deploy New MongoDB Database For Restoring:

We are going to deploy a new mongodb replicaset database for restoring backed up data.

Below is the YAML of a sample `MongoDB` crd that we are going to create

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  name: sample-mg-rs-restore
  namespace: demo
spec:
  version: "4.2.3"
  replicas: 3
  replicaSet:
    name: rs0
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: WipeOut
```

Create the above `MongoDB` crd,

```console
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mongodb/backup/kubestash/logical/replicaset/examples/mongodb-replicaset-restore.yaml
mongodb.kubedb.com/sample-mg-rs-restore created
```

Let's check if the database is ready to use,

```console
$ kubectl get mg -n demo sample-mg-rs-restore
NAME                   VERSION   STATUS   AGE
sample-mg-rs-restore   4.2.3     Ready    2m45s
```

Let's verify all the databases of this `sample-mg-rs-restore` by exec into its pod

```console
$ export USER=$(kubectl get secrets -n demo sample-mg-rs-restore-auth -o jsonpath='{.data.\username}' | base64 -d)

$ export PASSWORD=$(kubectl get secrets -n demo sample-mg-rs-restore-auth -o jsonpath='{.data.\password}' | base64 -d)

$ kubectl exec -it -n demo sample-mg-rs-restore-0 -- mongo admin -u $USER -p $PASSWORD

rs0:PRIMARY> show dbs
admin          0.000GB
config         0.000GB
kubedb-system  0.000GB
local          0.000GB

rs0:PRIMARY> show users
{
	"_id" : "admin.root",
	"userId" : UUID("b8a10e3f-05a1-421c-87ca-7bdf16f99563"),
	"user" : "root",
	"db" : "admin",
	"roles" : [
		{
			"role" : "root",
			"db" : "admin"
		}
	],
	"mechanisms" : [
		"SCRAM-SHA-1",
		"SCRAM-SHA-256"
	]
}

rs0:PRIMARY> exit
bye
```

As we can see no database named `newdb` exist in this new `sample-mg-rs-restore` database.

#### Create RestoreSession:

Now, we need to create a `RestoreSession` crd pointing to the `sample-mg-rs-restore` database.
Below is the YAML for the `RestoreSession` crd that we are going to create to restore the backed up data.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: mg-rs-restore
  namespace: demo
spec:
  target:
    name: sample-mg-rs-restore
    namespace: demo
    apiGroup: kubedb.com
    kind: MongoDB
  dataSource:
    snapshot: latest
    repository: s3-repo
    encryptionSecret:
      name: encry-secret 
      namespace: demo
  addon:
    name: mongodb-addon
    tasks:
      - name: LogicalBackupRestore
```

Here,

- `spec.dataSource.repository` specifies the `Repository` crd that holds the backend information where our backed up data has been stored.
- `spec.target` refers to the `MongoDB` crd for the `sample-mg-rs-restore` database.
- `spec.dataSource.snapshot` specifies that we are restoring from the latest backup snapshot of the `spec.dataSource.repository`.

Let's create the `RestoreSession` crd we have shown above,

```console
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mongodb/backup/kubestash/logical/replicaset/examples/restoresession-replicaset.yaml
restoresession.core.kubestash.com/mg-rs-restore created
```

Once, you have created the `RestoreSession` crd, KubeStash will create a job to restore. We can watch the `RestoreSession` phase to check if the restore process is succeeded or not.

Run the following command to watch `RestoreSession` phase,

```console
$ kubectl get restoresession -n demo mg-rs-restore -w
NAME            REPOSITORY   FAILURE-POLICY   PHASE       DURATION   AGE
mg-rs-restore   s3-repo                       Succeeded   9s         34s
```

So, we can see from the output of the above command that the restore process succeeded.

#### Verify Restored Data:

In this section, we are going to verify that the desired data has been restored successfully. We are going to connect to the `sample-mg-rs-restore` database and check whether the table we had created earlier is restored or not.

Lets, exec into the database pod and list available tables,

```console
$ kubectl exec -it -n demo sample-mg-rs-restore-0 -- mongo admin -u $USER -p $PASSWORD

rs0:PRIMARY> show dbs
admin          0.000GB
config         0.000GB
kubedb-system  0.000GB
local          0.000GB
newdb          0.000GB

rs0:PRIMARY> show users
{
	"_id" : "admin.root",
	"userId" : UUID("b8a10e3f-05a1-421c-87ca-7bdf16f99563"),
	"user" : "root",
	"db" : "admin",
	"roles" : [
		{
			"role" : "root",
			"db" : "admin"
		}
	],
	"mechanisms" : [
		"SCRAM-SHA-1",
		"SCRAM-SHA-256"
	]
}

rs0:PRIMARY> use newdb
switched to db newdb

rs0:PRIMARY> db.movie.find().pretty()
{ "_id" : ObjectId("657182194f1f295758f449a5"), "name" : "batman" }

rs0:PRIMARY> exit
bye
```

So, from the above output, we can see the database `newdb` that we had created earlier is restored into another new `MongoDB` database.

## Cleanup

To cleanup the Kubernetes resources created by this tutorial, run:

```console
kubectl delete -n demo restoresession mg-rs-restore
kubectl delete -n demo backupconfiguration mg
kubectl delete -n demo mg sample-mg-rs
kubectl delete -n demo mg sample-mg-rs-restore
kubectl delete -n demo backupstorage s3-storage-replicaset
```
