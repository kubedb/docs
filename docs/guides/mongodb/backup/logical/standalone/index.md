---
title: Backup & Restore Standalone MongoDB | Stash
description: Backup and restore standalone MongoDB database using Stash
menu:
  docs_{{ .version }}:
    identifier: guides-mongodb-backup-logical-standalone
    name: Standalone MongoDB
    parent: guides-mongodb-backup-logical
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Backup and Restore MongoDB database using Stash

Stash 0.9.0+ supports backup and restoration of MongoDB databases. This guide will show you how you can backup and restore your MongoDB database with Stash.

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

## Backup MongoDB

This section will demonstrate how to backup MongoDB database. Here, we are going to deploy a MongoDB database using KubeDB. Then, we are going to backup this database into a GCS bucket. Finally, we are going to restore the backed up data into another MongoDB database.

### Deploy Sample MongoDB Database

Let's deploy a sample MongoDB database and insert some data into it.

**Create MongoDB CRD:**

Below is the YAML of a sample MongoDB crd that we are going to create for this tutorial:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  name: sample-mongodb
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
  terminationPolicy: WipeOut
```

Create the above `MongoDB` crd,

```console
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mongodb/backup/logical/standalone/examples/mongodb.yaml
mongodb.kubedb.com/sample-mongodb created
```

KubeDB will deploy a MongoDB database according to the above specification. It will also create the necessary secrets and services to access the database.

Let's check if the database is ready to use,

```console
$ kubectl get mg -n demo sample-mongodb
NAME             VERSION       STATUS    AGE
sample-mongodb   4.4.26         Ready   2m9s
```

The database is `Ready`. Verify that KubeDB has created a Secret and a Service for this database using the following commands,

```console
$ kubectl get secret -n demo -l=app.kubernetes.io/instance=sample-mongodb
NAME                  TYPE     DATA   AGE
sample-mongodb-auth   Opaque   2      2m28s

$ kubectl get service -n demo -l=app.kubernetes.io/instance=sample-mongodb
NAME                 TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)     AGE
sample-mongodb       ClusterIP   10.107.58.222   <none>        27017/TCP   2m48s
sample-mongodb-gvr   ClusterIP   None            <none>        27017/TCP   2m48s
```

Here, we have to use service `sample-mongodb` and secret `sample-mongodb-auth` to connect with the database. KubeDB creates an [AppBinding](/docs/guides/mongodb/concepts/appbinding.md) crd that holds the necessary information to connect with the database.

**Verify AppBinding:**

Verify that the `AppBinding` has been created successfully using the following command,

```console
$ kubectl get appbindings -n demo
NAME             AGE
sample-mongodb   20m
```

Let's check the YAML of the above `AppBinding`,

```console
$ kubectl get appbindings -n demo sample-mongodb -o yaml
```

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1alpha2","kind":"MongoDB","metadata":{"annotations":{},"name":"sample-mongodb","namespace":"demo"},"spec":{"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}},"storageClassName":"standard"},"storageType":"Durable","terminationPolicy":"WipeOut","version":"4.4.26"}}
  creationTimestamp: "2022-10-26T05:13:07Z"
  generation: 1
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: sample-mongodb
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: mongodbs.kubedb.com
  name: sample-mongodb
  namespace: demo
  ownerReferences:
    - apiVersion: kubedb.com/v1alpha2
      blockOwnerDeletion: true
      controller: true
      kind: MongoDB
      name: sample-mongodb
      uid: 51676df9-682a-40ab-8f99-c6050b35f2f2
  resourceVersion: "580968"
  uid: ca88e369-a15a-4149-9386-24e876c5aa4b
spec:
  appRef:
    apiGroup: kubedb.com
    kind: MongoDB
    name: sample-mongodb
    namespace: demo
  clientConfig:
    service:
      name: sample-mongodb
      port: 27017
      scheme: mongodb
  parameters:
    apiVersion: config.kubedb.com/v1alpha1
    kind: MongoConfiguration
    stash:
      addon:
        backupTask:
          name: mongodb-backup-4.4.6
        restoreTask:
          name: mongodb-restore-4.4.6
  secret:
    name: sample-mongodb-auth
  type: kubedb.com/mongodb
  version: 4.4.26
```

Stash uses the `AppBinding` crd to connect with the target database. It requires the following two fields to set in AppBinding's `Spec` section.

- `spec.appRef` refers to the underlying application.
- `spec.clientConfig` defines how to communicate with the application.
- `spec.clientConfig.service.name` specifies the name of the service that connects to the database.
- `spec.secret` specifies the name of the secret that holds necessary credentials to access the database.
- `spec.parameters.stash` contains the Stash Addon information that will be used to backup/restore this MongoDB database.
- `spec.type` specifies the types of the app that this AppBinding is pointing to. KubeDB generated AppBinding follows the following format: `<app group>/<app resource type>`.

**Insert Sample Data:**

Now, we are going to exec into the database pod and create some sample data. At first, find out the database pod using the following command,

```console
$ kubectl get pods -n demo --selector="app.kubernetes.io/instance=sample-mongodb"
NAME               READY   STATUS    RESTARTS   AGE
sample-mongodb-0   1/1     Running   0          12m
```

Now, let's exec into the pod and create a table,

```console
$ export USER=$(kubectl get secrets -n demo sample-mongodb-auth -o jsonpath='{.data.\username}' | base64 -d)

$ export PASSWORD=$(kubectl get secrets -n demo sample-mongodb-auth -o jsonpath='{.data.\password}' | base64 -d)

$ kubectl exec -it -n demo sample-mongodb-0 -- mongo admin -u $USER -p $PASSWORD

> show dbs
admin  0.000GB
local  0.000GB
mydb   0.000GB

> show users
{
    "_id" : "admin.root",
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

> db.movie.insert({"name":"batman"});
WriteResult({ "nInserted" : 1 })

> db.movie.find().pretty()
{ "_id" : ObjectId("5d19d1cdc93d828f44e37735"), "name" : "batman" }

> exit
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
  name: gcs-repo
  namespace: demo
spec:
  backend:
    gcs:
      bucket: stash-testing
      prefix: demo/mongodb/sample-mongodb
    storageSecretName: gcs-secret
```

Let's create the `Repository` we have shown above,

```console
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mongodb/backup/logical/standalone/examples/repository.yaml
repository.stash.appscode.com/gcs-repo created
```

Now, we are ready to backup our database to our desired backend.

### Backup

We have to create a `BackupConfiguration` targeting respective AppBinding crd of our desired database. Then Stash will create a CronJob to periodically backup the database.

**Create BackupConfiguration:**

Below is the YAML for `BackupConfiguration` crd to backup the `sample-mongodb` database we have deployed earlier.,

```yaml
apiVersion: stash.appscode.com/v1beta1
kind: BackupConfiguration
metadata:
  name: sample-mongodb-backup
  namespace: demo
spec:
  schedule: "*/5 * * * *"
  repository:
    name: gcs-repo
  target:
    ref:
      apiVersion: appcatalog.appscode.com/v1alpha1
      kind: AppBinding
      name: sample-mongodb
  retentionPolicy:
    name: keep-last-5
    keepLast: 5
    prune: true
```

Here,

- `spec.schedule` specifies that we want to backup the database at 5 minutes interval.
- `spec.target.ref` refers to the `AppBinding` crd that was created for `sample-mongodb` database.

Let's create the `BackupConfiguration` crd we have shown above,

```console
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mongodb/backup/logical/standalone/examples/backupconfiguration.yaml
backupconfiguration.stash.appscode.com/sample-mongodb-backup created
```

**Verify Backup Setup Successful:**

If everything goes well, the phase of the `BackupConfiguration` should be `Ready`. The `Ready` phase indicates that the backup setup is successful. Let's verify the `Phase` of the BackupConfiguration,

```console
$ kubectl get backupconfiguration -n demo
NAME                    TASK                    SCHEDULE      PAUSED   PHASE      AGE
sample-mongodb-backup   mongodb-backup-4.4.6    */5 * * * *            Ready      11s
```

**Verify CronJob:**

Stash will create a CronJob with the schedule specified in `spec.schedule` field of `BackupConfiguration` crd.

Verify that the CronJob has been created using the following command,

```console
$ kubectl get cronjob -n demo
NAME                    SCHEDULE      SUSPEND   ACTIVE   LAST SCHEDULE   AGE
sample-mongodb-backup   */5 * * * *   False     0        <none>          61s
```

**Wait for BackupSession:**

The `sample-mongodb-backup` CronJob will trigger a backup on each schedule by creating a `BackpSession` crd.

Wait for the next schedule. Run the following command to watch `BackupSession` crd,

```console
$ kubectl get backupsession -n demo -w
NAME                               INVOKER-TYPE          INVOKER-NAME            PHASE       AGE
sample-mongodb-backup-1561974001   BackupConfiguration   sample-mongodb-backup   Running     5m19s
sample-mongodb-backup-1561974001   BackupConfiguration   sample-mongodb-backup   Succeeded   5m45s
```

We can see above that the backup session has succeeded. Now, we are going to verify that the backed up data has been stored in the backend.

**Verify Backup:**

Once a backup is complete, Stash will update the respective `Repository` crd to reflect the backup. Check that the repository `gcs-repo` has been updated by the following command,

```console
$ kubectl get repository -n demo gcs-repo
NAME       INTEGRITY   SIZE        SNAPSHOT-COUNT   LAST-SUCCESSFUL-BACKUP   AGE
gcs-repo   true        1.611 KiB   1                33s                      33m
```

Now, if we navigate to the GCS bucket, we are going to see backed up data has been stored in `demo/mongodb/sample-mongodb` directory as specified by `spec.backend.gcs.prefix` field of Repository crd.

> Note: Stash keeps all the backed up data encrypted. So, data in the backend will not make any sense until they are decrypted.

## Restore MongoDB
You can restore your data into the same database you have backed up from or into a different database in the same cluster or a different cluster. In this section, we are going to show you how to restore in the same database which may be necessary when you have accidentally deleted any data from the running database.

#### Stop Taking Backup of the Old Database:

At first, let's stop taking any further backup of the old database so that no backup is taken during restore process. We are going to pause the `BackupConfiguration` crd that we had created to backup the `sample-mongodb` database. Then, Stash will stop taking any further backup for this database.

Let's pause the `sample-mongodb-backup` BackupConfiguration,
```bash
$ kubectl patch backupconfiguration -n demo sample-mongodb-backup --type="merge" --patch='{"spec": {"paused": true}}'
backupconfiguration.stash.appscode.com/sample-mongodb-backup patched
```

Or you can use the Stash `kubectl` plugin to pause the `BackupConfiguration`,
```bash
$ kubectl stash pause backup -n demo --backupconfig=sample-mongodb-backup
BackupConfiguration demo/sample-mongodb-backup has been paused successfully.
```

Now, wait for a moment. Stash will pause the BackupConfiguration. Verify that the BackupConfiguration  has been paused,

```console
$ kubectl get backupconfiguration -n demo sample-mongodb-backup
NAME                   TASK                        SCHEDULE      PAUSED   PHASE   AGE
sample-mongodb-backup  mongodb-backup-4.4.6        */5 * * * *   true     Ready   26m
```

Notice the `PAUSED` column. Value `true` for this field means that the BackupConfiguration has been paused.

#### Simulate Disaster:

Now, let’s simulate an accidental deletion scenario. Here, we are going to exec into the database pod and delete the `newdb` database we had created earlier.
```console
$ kubectl exec -it -n demo sample-mongodb-0 -- mongo admin -u $USER -p $PASSWORD
> use newdb
switched to db newdb

> db.dropDatabase()
{ "dropped" : "newdb", "ok" : 1 }

> show dbs
admin   0.000GB
config  0.000GB
local   0.000GB

> exit
bye
```
#### Create RestoreSession:

Now, we need to create a `RestoreSession` crd pointing to the AppBinding of `sample-mongodb` database.
Below is the YAML for the `RestoreSession` crd that we are going to create to restore the backed up data.

```yaml
apiVersion: stash.appscode.com/v1beta1
kind: RestoreSession
metadata:
  name: sample-mongodb-restore
  namespace: demo
spec:
  repository:
    name: gcs-repo
  target:
    ref:
      apiVersion: appcatalog.appscode.com/v1alpha1
      kind: AppBinding
      name: sample-mongodb
  rules:
  - snapshots: [latest]
```

Here,

- `spec.repository.name` specifies the `Repository` crd that holds the backend information where our backed up data has been stored.
- `spec.target.ref` refers to the AppBinding crd for the `restored-mongodb` database.
- `spec.rules` specifies that we are restoring from the latest backup snapshot of the database.

Let's create the `RestoreSession` crd we have shown above,

```console
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mongodb/backup/logical/standalone/examples/restoresession.yaml
restoresession.stash.appscode.com/sample-mongodb-restore created
```

Once, you have created the `RestoreSession` crd, Stash will create a job to restore. We can watch the `RestoreSession` phase to check if the restore process is succeeded or not.

Run the following command to watch `RestoreSession` phase,

```console
$ kubectl get restoresession -n demo sample-mongodb-restore -w
NAME                     REPOSITORY-NAME   PHASE       AGE
sample-mongodb-restore   gcs-repo          Running     5s
sample-mongodb-restore   gcs-repo          Succeeded   43s
```

So, we can see from the output of the above command that the restore process succeeded.

#### Verify Restored Data:

In this section, we are going to verify that the desired data has been restored successfully. We are going to connect to the database and check whether the table we had created earlier is restored or not.

Lets, exec into the database pod and list available tables,

```console
$ kubectl exec -it -n demo sample-mongodb-0 -- mongo admin -u $USER -p $PASSWORD

> show dbs
admin   0.000GB
config  0.000GB
local   0.000GB
newdb   0.000GB

> show users
{
    "_id" : "admin.root",
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
{ "_id" : ObjectId("5d19d1cdc93d828f44e37735"), "name" : "batman" }

> exit
bye
```

So, from the above output, we can see the database `newdb` that we had created earlier is restored.

## Cleanup

To cleanup the Kubernetes resources created by this tutorial, run:

```console
kubectl delete -n demo restoresession sample-mongodb-restore sample-mongo
kubectl delete -n demo backupconfiguration sample-mongodb-backup
kubectl delete -n demo mg sample-mongodb
kubectl delete -n demo repository gcs-repo
```
