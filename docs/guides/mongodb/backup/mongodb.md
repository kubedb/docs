---
title: Backup & Restore Standalone MongoDB | Stash
description: Backup and restore standalone MongoDB database using Stash
menu:
  docs_{{ .version }}:
    identifier: standalone-mongodb-{{ .subproject_version }}
    name: Standalone MongoDB
    parent: stash-mongodb-guides-{{ .subproject_version }}
    weight: 10
product_name: stash
menu_name: docs_{{ .version }}
section_menu_id: stash-addons
---

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [Stash Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Backup and Restore MongoDB database using Stash

Stash 0.9.0+ supports backup and restoration of MongoDB databases. This guide will show you how you can backup and restore your MongoDB database with Stash.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using Minikube.
- Install Stash in your cluster following the steps [here](/docs/setup/README.md).
- Install MongoDB addon for Stash following the steps [here](/docs/addons/mongodb/setup/install.md).
- Install [KubeDB](https://kubedb.com) in your cluster following the steps [here](https://kubedb.com/docs/latest/setup/install/). This step is optional. You can deploy your database using any method you want. We are using KubeDB because KubeDB simplifies many of the difficult or tedious management tasks of running a production grade databases on private and public clouds.
- If you are not familiar with how Stash backup and restore MongoDB databases, please check the following guide [here](/docs/addons/mongodb/overview.md).

You have to be familiar with following custom resources:

- [AppBinding](/docs/concepts/crds/appbinding.md)
- [Function](/docs/concepts/crds/function.md)
- [Task](/docs/concepts/crds/task.md)
- [BackupConfiguration](/docs/concepts/crds/backupconfiguration.md)
- [RestoreSession](/docs/concepts/crds/restoresession.md)

To keep things isolated, we are going to use a separate namespace called `demo` throughout this tutorial. Create `demo` namespace if you haven't created yet.

```console
$ kubectl create ns demo
namespace/demo created
```

> Note: YAML files used in this tutorial are stored [here](https://github.com/stashed/mongodb/tree/{{< param "info.subproject_version" >}}/docs/examples).

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
  version: "4.2.3"
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
$ kubectl apply -f https://github.com/stashed/mongodb/raw/{{< param "info.subproject_version" >}}/docs/examples/backup/standalone/mongodb.yaml
mongodb.kubedb.com/sample-mongodb created
```

KubeDB will deploy a MongoDB database according to the above specification. It will also create the necessary secrets and services to access the database.

Let's check if the database is ready to use,

```console
$ kubectl get mg -n demo sample-mongodb
NAME             VERSION        STATUS    AGE
sample-mongodb   4.2.3         Running   2m9s
```

The database is `Ready`. Verify that KubeDB has created a Secret and a Service for this database using the following commands,

```console
$ kubectl get secret -n demo -l=kubedb.com/name=sample-mongodb
NAME                  TYPE     DATA   AGE
sample-mongodb-auth   Opaque   2      2m28s

$ kubectl get service -n demo -l=kubedb.com/name=sample-mongodb
NAME                 TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)     AGE
sample-mongodb       ClusterIP   10.107.58.222   <none>        27017/TCP   2m48s
sample-mongodb-gvr   ClusterIP   None            <none>        27017/TCP   2m48s
```

Here, we have to use service `sample-mongodb` and secret `sample-mongodb-auth` to connect with the database. KubeDB creates an [AppBinding](/docs/concepts/crds/appbinding.md) crd that holds the necessary information to connect with the database.

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
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: sample-mongodb
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: mongodbs.kubedb.com
    kubedb.com/name: sample-mongodb
  name: sample-mongodb
  namespace: demo
spec:
  clientConfig:
    service:
      name: sample-mongodb
      port: 27017
      scheme: mongodb
  secret:
    name: sample-mongodb-auth
  parameters:
    apiVersion: config.kubedb.com/v1alpha1
    kind: MongoConfiguration
    stash:
      addon:
        backupTask:
          name: mongodb-backup-{{< param "info.subproject_version" >}}
        restoreTask:
          name: mongodb-restore-{{< param "info.subproject_version" >}}
  type: kubedb.com/mongodb
  version: "4.2.3"
```

Stash uses the `AppBinding` crd to connect with the target database. It requires the following two fields to set in AppBinding's `Spec` section.

- `spec.clientConfig.service.name` specifies the name of the service that connects to the database.
- `spec.secret` specifies the name of the secret that holds necessary credentials to access the database.
- `spec.parameters.stash` contains the Stash Addon information that will be used to backup/restore this MongoDB database.
- `spec.type` specifies the types of the app that this AppBinding is pointing to. KubeDB generated AppBinding follows the following format: `<app group>/<app resource type>`.

### AppBinding for SSL

If `SSLMode` of the MongoDB server is either of `requireSSL` or `preferSSL`, you can provide ssl connection information through AppBinding Specs.

User need to provide the following fields in case of SSL is enabled,

- `spec.clientConfig.caBundle` specifies the CA certificate that is used in [`--sslCAFile`](https://docs.mongodb.com/manual/reference/program/mongod/index.html#cmdoption-mongod-sslcafile) flag of `mongod`.
- `spec.secret` specifies the name of the secret that holds `client.pem` file. Follow the [mongodb official doc](https://docs.mongodb.com/manual/tutorial/configure-x509-client-authentication/) to learn how to create `client.pem` and add the subject of `client.pem` as user (with appropriate roles) to mongodb server.

**KubeDB does these automatically**. It has added the subject of `client.pem` in the mongodb server with `root` role. So, user can just use the appbinding that is created by KubeDB without doing any hurdle! See the [MongoDB with TLS/SSL (Transport Encryption)](https://github.com/kubedb/docs/blob/master/docs/guides/mongodb/tls-ssl-encryption/tls-ssl-encryption.md) guide to learn about the ssl options in mongodb in details.

So, in KubeDB, the following `CRD` deploys a mongodb replicaset where ssl is enabled (`requireSSL` sslmode),

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  name: sample-mongodb-ssl
  namespace: demo
spec:
  version: "4.2.3"
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: WipeOut
  sslMode: requireSSL
```

After the deploy is done, kubedb will create a appbinding that will look like:

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: sample-mongodb-ssl
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: mongodb
    app.kubernetes.io/version: 4.2.3
    app.kubernetes.io/name: mongodbs.kubedb.com
    kubedb.com/name: sample-mongodb-ssl
  name: sample-mongodb-ssl
  namespace: demo
spec:
  clientConfig:
    caBundle: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUM0RENDQWNpZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFoTVJJd0VBWURWUVFLRXdscmRXSmwKWkdJNlkyRXhDekFKQmdOVkJBTVRBbU5oTUI0WERURTVNRGt6TURFek1EYzFPRm9YRFRJNU1Ea3lOekV6TURjMQpPRm93SVRFU01CQUdBMVVFQ2hNSmEzVmlaV1JpT21OaE1Rc3dDUVlEVlFRREV3SmpZVENDQVNJd0RRWUpLb1pJCmh2Y05BUUVCQlFBRGdnRVBBRENDQVFvQ2dnRUJBTXRRYmNiaitiL2JES3UvcERDMGI4SlYvREFjTTY4TDFTbDIKVzJMSHh4YzZZMDNnNEZKV290ZjJaRk5IczhMUmNQVmt5Qms1ZkRnVnJWS0FNY1N2Q2UrTHI2ek9LTXFXVEtwZgpqWGZ5dXpsUnpna3FsZEdLRXowNndtQTJHZE5od1VKL2RWVEFrbVU5dlp5ekZoaHhUcFRoZFBLRGlVRlNxdGlKCk5rWHUzZmJYK0Y4RkRVYVQ5Y3FKR3c5N0xRQW9NaGF5ZVJabDkrM2NTZ1NvdFhBVFlnTTZIU200UnFyaGdqMEEKU2MxSkV3TERkMDFBK25TeGtlVmkzV3M3SGo2Wlp6TUtlNVN1b2Z0NFlUMmlzcjBpRXpHQXhCRllwVFRUN1VuQgp3SlNCcEFRZTNFNWpuMGpqR21qVFV1TUJiR3VtcUhZck80akJ5TXRrUXZxVlFqdndWU0VDQXdFQUFhTWpNQ0V3CkRnWURWUjBQQVFIL0JBUURBZ0trTUE4R0ExVWRFd0VCL3dRRk1BTUJBZjh3RFFZSktvWklodmNOQVFFTEJRQUQKZ2dFQkFMTjVwSElGd0lBUFpaaWU0THRwY1ltanZ5eHBWS3MwdlY5TXZPZnVRVGtydktNQnZxbkFlU0NJUDEycQp3OThNQnhYV29BNFNtUDVPZHA5SklSYWdCQmJOV2tVUFJsY3dkWUdpZGtnMWhjZ3ZMTTZUaXlCVnNEMDB2c1N5CjgwTzlpQnVJaGdqdW9QYzdCdUFMOSsraDdzR0ZXWXpVVXBrdHRRMkgrSGtlOGpHUEQvTytzT3Q1OUFvaCtPOFUKTDBZSno4YkZ0UnFEdUZYcUlRMXpzaEFpMFkzSmloUTBZWGFPQU8yeUttZENkY2ZNdXlUSUhNWnpTOUMzZEEwRwpLb3dNYjh4d0hjRld6WW5WdnR5K2g1Qmd6SEx4UU9pd2Foc280RW9vR01xbTlzVVBOSWJZTGRNUndVZWRMZDcwCnRlMWw2ak5DVys0ZVVJb2czc3BvcW9kL3ZZMD0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
    service:
      name: sample-mongodb-ssl
      port: 27017
      scheme: mongodb
  secret:
    name: sample-mongodb-ssl-cert
  parameters:
    apiVersion: config.kubedb.com/v1alpha1
    kind: MongoConfiguration
    stash:
      addon:
        backupTask:
          name: mongodb-backup-{{< param "info.subproject_version" >}}
        restoreTask:
          name: mongodb-restore-{{< param "info.subproject_version" >}}
  type: kubedb.com/mongodb
  version: "4.2.3"
```

Here, `sample-mongodb-cert` contains few required certificates, and one of them is `client.pem` which is required to backup/restore ssl enabled mongodb server using stash-mongodb.

**Creating AppBinding Manually:**

If you deploy MongoDB database without KubeDB, you have to create the AppBinding crd manually in the same namespace as the service and secret of the database.

The following YAML shows a minimal AppBinding specification that you have to create if you deploy MongoDB database without KubeDB.

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: my-custom-appbinding
  namespace: my-database-namespace
spec:
  clientConfig:
    service:
      name: my-database-service
      port: 27017
      scheme: mongodb
  secret:
    name: my-database-credentials-secret
  # type field is optional. you can keep it empty.
  # if you keep it empty then the value of TARGET_APP_RESOURCE variable
  # will be set to "appbinding" during auto-backup.
  type: mongodb
```

**Insert Sample Data:**

Now, we are going to exec into the database pod and create some sample data. At first, find out the database pod using the following command,

```console
$ kubectl get pods -n demo --selector="kubedb.com/name=sample-mongodb"
NAME               READY   STATUS    RESTARTS   AGE
sample-mongodb-0   1/1     Running   0          12m
```

Now, let's exec into the pod and create a table,

```console
$ kubectl get secrets -n demo sample-mongodb-auth -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo sample-mongodb-auth -o jsonpath='{.data.\password}' | base64 -d
Tv1pSiLjGqZ9W4jE

$ kubectl exec -it -n demo sample-mongodb-0 bash

mongodb@sample-mongodb-0:/$ mongo admin -u root -p Tv1pSiLjGqZ9W4jE

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

We are going to store our backed up data into a GCS bucket. At first, we need to create a secret with GCS credentials then we need to create a `Repository` crd. If you want to use a different backend, please read the respective backend configuration doc from [here](/docs/guides/latest/backends/overview.md).

**Create Storage Secret:**

Let's create a secret called `gcs-secret` with access credentials to our desired GCS bucket,

```console
$ echo -n 'changeit' > RESTIC_PASSWORD
$ echo -n '<your-project-id>' > GOOGLE_PROJECT_ID
$ cat downloaded-sa-json.key > GOOGLE_SERVICE_ACCOUNT_JSON_KEY
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
$ kubectl apply -f https://github.com/stashed/mongodb/raw/{{< param "info.subproject_version" >}}/docs/examples/backup/standalone/repository.yaml
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
#  task: # Uncomment if you are not using KubeDB to deploy your database
#    name: mongodb-backup-{{< param "info.subproject_version" >}}
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
- `spec.task.name` specifies the name of the task crd that specifies the necessary Function and their execution order to backup a MongoDB database.
- `spec.target.ref` refers to the `AppBinding` crd that was created for `sample-mongodb` database.

Let's create the `BackupConfiguration` crd we have shown above,

```console
$ kubectl apply -f https://github.com/stashed/mongodb/raw/{{< param "info.subproject_version" >}}/docs/examples/backup/standalone/backupconfiguration.yaml
backupconfiguration.stash.appscode.com/sample-mongodb-backup created
```

**Verify CronJob:**

If everything goes well, Stash will create a CronJob with the schedule specified in `spec.schedule` field of `BackupConfiguration` crd.

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

In this section, we are going to restore the database from the backup we have taken in the previous section. We are going to deploy a new database and initialize it from the backup.

**Stop Taking Backup of the Old Database:**

At first, let's stop taking any further backup of the old database so that no backup is taken during restore process. We are going to pause the `BackupConfiguration` crd that we had created to backup the `sample-mongodb` database. Then, Stash will stop taking any further backup for this database.

Let's pause the `sample-mongodb-backup` BackupConfiguration,

```console
$ kubectl patch backupconfiguration -n demo sample-mongodb-backup --type="merge" --patch='{"spec": {"paused": true}}'
backupconfiguration.stash.appscode.com/sample-mongodb-backup patched
```

Now, wait for a moment. Stash will pause the BackupConfiguration. Verify that the BackupConfiguration  has been paused,

```console
$ kubectl get backupconfiguration -n demo sample-mongodb-backup
NAME                   TASK                         SCHEDULE      PAUSED   AGE
sample-mongodb-backup  mongodb-backup-{{< param "info.subproject_version" >}}        */5 * * * *   true     26m
```

Notice the `PAUSED` column. Value `true` for this field means that the BackupConfiguration has been paused.

**Deploy Restored Database:**

Now, we have to deploy the restored database similarly as we have deployed the original `sample-psotgres` database. However, this time there will be the following differences:

- We are going to specify `spec.init.waitForInitialRestore` field that will tell KubeDB to wait for first restore to complete before marking this database ready to use.

Below is the YAML for `MongoDB` crd we are going deploy to initialize from backup,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  name: restored-mongodb
  namespace: demo
spec:
  version: "4.2.3"
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

Let's create the above database,

```console
$ kubectl apply -f https://github.com/stashed/mongodb/raw/{{< param "info.subproject_version" >}}/docs/examples/restore/standalone/restored-mongodb.yaml
mongodb.kubedb.com/restored-mongodb created
```

If you check the database status, you will see it is stuck in `Provisioning` state.

```console
$ kubectl get mg -n demo restored-mongodb
NAME               VERSION        STATUS         AGE
restored-mongodb   4.2.3         Provisioning   17s
```

**Create RestoreSession:**

Now, we need to create a `RestoreSession` crd pointing to the AppBinding for this restored database.

Check AppBinding has been created for the `restored-mongodb` database using the following command,

```console
$ kubectl get appbindings -n demo restored-mongodb
NAME               AGE
restored-mongodb   29s
```

> If you are not using KubeDB to deploy database, create the AppBinding manually.

Below is the YAML for the `RestoreSession` crd that we are going to create to restore backed up data into `restored-mongodb` database.

```yaml
apiVersion: stash.appscode.com/v1beta1
kind: RestoreSession
metadata:
  name: sample-mongodb-restore
  namespace: demo
spec:
#  task: # Uncomment if you are not using KubeDB to deploy your database
#    name: mongodb-restore-{{< param "info.subproject_version" >}}
  repository:
    name: gcs-repo
  target:
    ref:
      apiVersion: appcatalog.appscode.com/v1alpha1
      kind: AppBinding
      name: restored-mongodb
  rules:
  - snapshots: [latest]
```

Here,

- `spec.task.name` specifies the name of the `Task` crd that specifies the Functions and their execution order to restore a MongoDB database.
- `spec.repository.name` specifies the `Repository` crd that holds the backend information where our backed up data has been stored.
- `spec.target.ref` refers to the AppBinding crd for the `restored-mongodb` database.
- `spec.rules` specifies that we are restoring from the latest backup snapshot of the database.

Let's create the `RestoreSession` crd we have shown above,

```console
$ kubectl apply -f https://github.com/stashed/mongodb/raw/{{< param "info.subproject_version" >}}/docs/examples/restore/standalone/restoresession.yaml
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

**Verify Restored Data:**

In this section, we are going to verify that the desired data has been restored successfully. We are going to connect to the database and check whether the table we had created in the original database is restored or not.

At first, check if the database has gone into `Running` state by the following command,

```console
$ kubectl get mg -n demo restored-mongodb
NAME               VERSION        STATUS    AGE
restored-mongodb   4.2.3         Running   105m
```

Now, find out the database pod by the following command,

```console
$ kubectl get pods -n demo --selector="kubedb.com/name=restored-mongodb"
NAME                 READY   STATUS    RESTARTS   AGE
restored-mongodb-0   1/1     Running   0          106m
```

Now, exec into the database pod and list available tables,

```console
$ kubectl get secrets -n demo sample-mongodb-auth -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo sample-mongodb-auth -o jsonpath='{.data.\password}' | base64 -d
Tv1pSiLjGqZ9W4jE

$ kubectl exec -it -n demo restored-mongodb-0 bash

mongodb@restored-mongodb-0:/$ mongo admin -u root -p Tv1pSiLjGqZ9W4jE

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

So, from the above output, we can see the database `newdb` that we had created in the original database `sample-mongodb` is restored in the restored database `restored-mongodb`.

## Cleanup

To cleanup the Kubernetes resources created by this tutorial, run:

```console
kubectl delete -n demo restoresession sample-mongodb-restore sample-mongo
kubectl delete -n demo backupconfiguration sample-mongodb-backup
kubectl delete -n demo mg sample-mongodb sample-mongodb-ssl restored-mongodb
kubectl delete -n demo repository gcs-repo
```
