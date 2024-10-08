---
title: Backup & Restore Sharded MongoDB | KubeStash
description: Backup and restore sharded MongoDB database using KubeStash
menu:
  docs_{{ .version }}:
    identifier: guides-mongodb-backup-kubestash-logical-sharded
    name: MongoDB Sharded Cluster
    parent: guides-mongodb-backup-kubestash-logical
    weight: 22
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Backup and Restore MongoDB database using KubeStash

KubeStash v0.1.0+ supports backup and restoration of MongoDB databases. This guide will show you how you can backup and restore your MongoDB database with KubeStash.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using Minikube.
- Install KubeDB in your cluster following the steps [here](/docs/setup/README.md).
- Install KubeStash Enterprise in your cluster following the steps [here](https://kubestash.com/docs/latest/setup/install/kubestash/).
- Install KubeStash `kubectl` plugin following the steps [here](https://kubestash.com/docs/latest/setup/install/kubectl-plugin/).
- If you are not familiar with how KubeStash backup and restore MongoDB databases, please check the following guide [here](/docs/guides/mongodb/backup/kubestash/overview/index.md).

You have to be familiar with following custom resources:

- [AppBinding](/docs/guides/mongodb/concepts/appbinding.md)
- [Function](https://kubestash.com/docs/latest/concepts/crds/function/)
- [Addon](https://kubestash.com/docs/latest/concepts/crds/addon/)
- [BackupConfiguration](https://kubestash.com/docs/latest/concepts/crds/backupconfiguration/)
- [RestoreSession](https://kubestash.com/docs/latest/concepts/crds/restoresession/)

To keep things isolated, we are going to use a separate namespace called `demo` throughout this tutorial. Create `demo` namespace if you haven't created yet.

```bash
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
  name: sample-mg-sh
  namespace: demo
spec:
  version: 4.2.24
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

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mongodb/backup/kubestash/logical/sharding/examples/mongodb-sharding.yaml
mongodb.kubedb.com/sample-mg-sh created
```

KubeDB will deploy a MongoDB database according to the above specification. It will also create the necessary secrets and services to access the database.

Let's check if the database is ready to use,

```bash
$ kubectl get mongodb -n demo sample-mg-sh
NAME           VERSION   STATUS   AGE
sample-mg-sh   4.2.24     Ready    5m39s
```

The database is `Ready`. Verify that KubeDB has created a Secret and a Service for this database using the following commands,

```bash
$ kubectl get secret -n demo -l=app.kubernetes.io/instance=sample-mg-sh
NAME                TYPE                       DATA   AGE
sample-mg-sh-auth   kubernetes.io/basic-auth   2      21m
sample-mg-sh-key    Opaque                     1      21m

$ kubectl get service -n demo -l=app.kubernetes.io/instance=sample-mg-sh
NAME                          TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)     AGE
sample-mg-sh                  ClusterIP   10.96.80.43   <none>        27017/TCP   21m
sample-mg-sh-configsvr-pods   ClusterIP   None          <none>        27017/TCP   21m
sample-mg-sh-mongos-pods      ClusterIP   None          <none>        27017/TCP   21m
sample-mg-sh-shard0-pods      ClusterIP   None          <none>        27017/TCP   21m
sample-mg-sh-shard1-pods      ClusterIP   None          <none>        27017/TCP   21m
sample-mg-sh-shard2-pods      ClusterIP   None          <none>        27017/TCP   21m
```

Here, we have to use service `sample-mg-sh` and secret `sample-mg-sh-auth` to connect with the database.

**Insert Sample Data:**

> Note: You can insert data into this `MongoDB` database using our [KubeDB CLI](https://kubedb.com/docs/latest/setup/install/kubectl_plugin/).

For simplicity, we are going to exec into the database pod and create some sample data. At first, find out the database mongos pod using the following command,

```bash
$ kubectl get pods -n demo --selector="mongodb.kubedb.com/node.mongos=sample-mg-sh-mongos"
NAME                    READY   STATUS    RESTARTS   AGE
sample-mg-sh-mongos-0   1/1     Running   0          21m
sample-mg-sh-mongos-1   1/1     Running   0          21m
```

Now, let's exec into the pod and create a table,

```bash
$ export USER=$(kubectl get secrets -n demo sample-mg-sh-auth -o jsonpath='{.data.\username}' | base64 -d)

$ export PASSWORD=$(kubectl get secrets -n demo sample-mg-sh-auth -o jsonpath='{.data.\password}' | base64 -d)

$ kubectl exec -it -n demo sample-mg-sh-mongos-0 -- mongo admin -u $USER -p $PASSWORD

mongos> show dbs
admin          0.000GB
config         0.002GB
kubedb-system  0.000GB

mongos> show users
{
	"_id" : "admin.root",
	"userId" : UUID("61a02236-1e25-4f58-9b92-9a41a80726bc"),
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

mongos> use newdb
switched to db newdb

mongos> db.movie.insert({"name":"batman"});
WriteResult({ "nInserted" : 1 })

mongos> exit
bye

```

Now, we are ready to backup this sample database.

### Prepare Backend

We are going to store our backed up data into a S3 bucket. At first, we need to create a secret with S3 credentials then we need to create a `BackupStorage` crd. If you want to use a different backend, please read the respective backend configuration doc from [here](https://kubestash.com/docs/latest/guides/backends/overview/).

**Create Storage Secret:**

Let's create a secret called `s3-secret` with access credentials to our desired S3 bucket,

```bash
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
  name: s3-storage-sharding
  namespace: demo
spec:
  storage:
    provider: s3
    s3:
      endpoint: us-east-1.linodeobjects.com
      bucket: kubestash-testing
      region: us-east-1
      prefix: demo-sharding
      secretName: s3-secret
  usagePolicy:
    allowedNamespaces:
      from: All
  deletionPolicy: WipeOut
```

Let's create the `BackupStorage` we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mongodb/backup/kubestash/logical/sharding/examples/backupstorage-sharding.yaml
backupstorage.storage.kubestash.com/s3-storage-sharding created
```

Now, we are ready to backup our database to our desired backend.

### Backup

We have to create a `BackupConfiguration` targeting respective MongoDB crd of our desired database. Then KubeStash will create a CronJob to periodically backup the database. Before that we need to create an secret for encrypt data and retention policy.

**Create Encryption Secret:**

EncryptionSecret refers to the Secret containing the encryption key which will be used to encode/decode the backed up data. Let's create a secret called `encry-secret`

```bash
$ kubectl create secret generic encry-secret -n demo \
    --from-literal=RESTIC_PASSWORD='123' -n demo
secret/encry-secret created
```

**Create Retention Policy:**

`RetentionPolicy` specifies how the old Snapshots should be cleaned up. This is a namespaced CRD.However, we can refer it from other namespaces as long as it is permitted via `.spec.usagePolicy`. Below is the YAML of the `RetentionPolicy` called `backup-rp`

```bash
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

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mongodb/backup/kubestash/logical/sharding/examples/retentionpolicy.yaml
retentionpolicy.storage.kubestash.com/backup-rp created
```

**Create BackupConfiguration:**

As we just create our encryption secret and retention policy, we are now ready to apply `BackupConfiguration` crd to take backup out database.

Below is the YAML for `BackupConfiguration` crd to backup the `sample-mg-sh` database we have deployed earlier.,

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
    name: sample-mg-sh
  backends:
    - name: s3-backend
      storageRef:
        namespace: demo
        name: s3-storage-sharding
      retentionPolicy:
        name: backup-rp
        namespace: demo
  sessions:
    - name: frequent
      scheduler:
        jobTemplate:
          backoffLimit: 1
        schedule: "*/3 * * * *"
      repositories:
        - name: s3-repo
          backend: s3-backend
          directory: /sharding
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

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mongodb/backup/kubestash/logical/sharding/examples/backupconfiguration-sharding.yaml
backupconfiguration.core.kubestash.com/mg created
```

**Verify Backup Setup Successful:**

If everything goes well, the phase of the `BackupConfiguration` should be `Ready`. The `Ready` phase indicates that the backup setup is successful. Let's verify the `Phase` of the BackupConfiguration,

```bash
$ kubectl get backupconfiguration -n demo
NAME   PHASE   PAUSED   AGE
mg     Ready            85s
```

**Verify CronJob:**

KubeStash will create a CronJob with the schedule specified in `spec.sessions.schedule` field of `BackupConfiguration` crd.

Verify that the CronJob has been created using the following command,

```bash
$ kubectl get cronjob -n demo
NAME                  SCHEDULE      SUSPEND   ACTIVE   LAST SCHEDULE   AGE
trigger-mg-frequent   */3 * * * *   False     0        <none>          101s
```

**Wait for BackupSession:**

The `trigger-mg-frequent` CronJob will trigger a backup on each schedule by creating a `BackpSession` crd.

Wait for the next schedule. Run the following command to watch `BackupSession` crd,

```bash
$ kubectl get backupsession -n demo
NAME                     INVOKER-TYPE          INVOKER-NAME   PHASE       DURATION   AGE
mg-frequent-1701950402   BackupConfiguration   mg             Succeeded              3m5s
mg-frequent-1701950582   BackupConfiguration   mg             Running                5s
```

We can see above that the backup session has succeeded. Now, we are going to verify that the backed up data has been stored in the backend.

**Verify Backup:**

Once a backup is complete, KubeStash will update the respective `Snapshot` crd to reflect the backup. It will be created when a backup is triggered. Check that the `Snapshot` Phase to verify backup.

```bash
$ kubectl get snapshot -n demo
NAME                             REPOSITORY   SESSION    SNAPSHOT-TIME          DELETION-POLICY   PHASE       VERIFICATION-STATUS   AGE
s3-repo-mg-frequent-1701950402   s3-repo      frequent   2023-12-07T12:00:11Z   Delete            Succeeded                         3m37s
s3-repo-mg-frequent-1701950582   s3-repo      frequent   2023-12-07T12:03:08Z   Delete            Succeeded                         37s
```

KubeStash will also update the respective `Repository` crd to reflect the backup. Check that the repository `s3-repo` has been updated by the following command,

```bash
$ kubectl get repository -n demo s3-repo
NAME      INTEGRITY   SNAPSHOT-COUNT   SIZE         PHASE   LAST-SUCCESSFUL-BACKUP   AGE
s3-repo   true        2                95.660 KiB   Ready   41s                      4m3s
```

Now, if we navigate to the S3 bucket, we are going to see backed up data has been stored in `demo/sharding/` directory as specified by `spec.sessions.repositories.directory` field of `BackupConfiguration` crd.

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

```bash
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
  name: sample-mg-sh-restore
  namespace: demo
spec:
  version: 4.2.24
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

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mongodb/backup/kubestash/logical/sharding/examples/mongodb-sharding-restore.yaml
mongodb.kubedb.com/sample-mg-sh-restore created
```

Let's check if the database is ready to use,

```bash
$ kubectl get mg -n demo sample-mg-sh-restore
NAME                   VERSION   STATUS   AGE
sample-mg-sh-restore   4.2.24     Ready    7m47s
```

Let's verify all the databases of this `sample-mg-sh-restore` by exec into its mongos pod

```bash
$ export USER=$(kubectl get secrets -n demo sample-mg-sh-restore-auth -o jsonpath='{.data.\username}' | base64 -d)

$ export PASSWORD=$(kubectl get secrets -n demo sample-mg-sh-restore-auth -o jsonpath='{.data.\password}' | base64 -d)

$ kubectl exec -it -n demo sample-mg-sh-restore-mongos-0 -- mongo admin -u $USER -p $PASSWORD

mongos> show dbs
admin          0.000GB
config         0.002GB
kubedb-system  0.000GB

mongos> show users
{
	"_id" : "admin.root",
	"userId" : UUID("4400d0cc-bba7-4626-bf5b-7521fca30ff9"),
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

mongos> exit
bye
```

As we can see no database named `newdb` exist in this new `sample-mg-sh-restore` database.

#### Create RestoreSession:

Now, we need to create a `RestoreSession` crd pointing to the `sample-mg-sh-restore` database.
Below is the YAML for the `RestoreSession` crd that we are going to create to restore the backed up data.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: mg-sh-restore
  namespace: demo
spec:
  target:
    name: sample-mg-sh-restore
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
- `spec.target` refers to the `MongoDB` crd for the `sample-mg-sh-restore` database.
- `spec.dataSource.snapshot` specifies that we are restoring from the latest backup snapshot of the `spec.dataSource.repository`.

Let's create the `RestoreSession` crd we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mongodb/backup/kubestash/logical/sharding/examples/restoresession-sharding.yaml
restoresession.core.kubestash.com/mg-sh-restore created
```

Once, you have created the `RestoreSession` crd, KubeStash will create a job to restore. We can watch the `RestoreSession` phase to check if the restore process is succeeded or not.

Run the following command to watch `RestoreSession` phase,

```bash
$ kubectl get restoresession -n demo mg-sh-restore -w
NAME            REPOSITORY   FAILURE-POLICY   PHASE       DURATION   AGE
mg-sh-restore   s3-repo                       Succeeded   15s        48s
```

So, we can see from the output of the above command that the restore process succeeded.

#### Verify Restored Data:

In this section, we are going to verify that the desired data has been restored successfully. We are going to connect to the `sample-mg-sh-restore` database and check whether the table we had created earlier is restored or not.

Lets, exec into the database's mongos pod and list available tables,

```bash
$ kubectl exec -it -n demo sample-mg-sh-restore-mongos-0 -- mongo admin -u $USER -p $PASSWORD

mongos> show dbs
admin          0.000GB
config         0.002GB
kubedb-system  0.000GB
newdb          0.000GB

mongos> show users
{
	"_id" : "admin.root",
	"userId" : UUID("4400d0cc-bba7-4626-bf5b-7521fca30ff9"),
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

mongos> use newdb
switched to db newdb

mongos> db.movie.find().pretty()
{ "_id" : ObjectId("6571b0465bde128e0c4d7caa"), "name" : "batman" }

mongos> exit
bye
```

So, from the above output, we can see the database `newdb` that we had created earlier is restored into another new `MongoDB` database.

## Cleanup

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete -n demo restoresession mg-sh-restore
kubectl delete -n demo backupconfiguration mg
kubectl delete -n demo mg sample-mg-sh
kubectl delete -n demo mg sample-mg-sh-restore
kubectl delete -n demo backupstorage s3-storage-sharding
```
