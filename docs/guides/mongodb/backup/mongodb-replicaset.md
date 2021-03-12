---
title: Backup & Restore MongoDB ReplicaSet Cluster | Stash
description: Backup and restore MongoDB ReplicaSet cluster using Stash
menu:
  docs_{{ .version }}:
    identifier: mongodb-replicaset-{{ .subproject_version }}
    name: MongoDB ReplicaSet Cluster
    parent: stash-mongodb-guides-{{ .subproject_version }}
    weight: 20
product_name: stash
menu_name: docs_{{ .version }}
section_menu_id: stash-addons
---

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [Stash Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Backup and Restore MongoDB ReplicaSet Clusters using Stash

Stash supports taking [backup and restores MongoDB ReplicaSet clusters in "idiomatic" way](https://docs.mongodb.com/manual/tutorial/restore-replica-set-from-backup/). This guide will show you how you can backup and restore your MongoDB ReplicaSet clusters with Stash.

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

## Backup MongoDB ReplicaSet using Stash

This section will demonstrate how to backup MongoDB ReplicaSet cluster. Here, we are going to deploy a MongoDB ReplicaSet using KubeDB. Then, we are going to backup this database into a GCS bucket. Finally, we are going to restore the backed up data into another MongoDB ReplicaSet.

### Deploy Sample MongoDB ReplicaSet

Let's deploy a sample MongoDB ReplicaSet database and insert some data into it.

**Create MongoDB CRD:**

Below is the YAML of a sample MongoDB crd that we are going to create for this tutorial:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  name: sample-mgo-rs
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
$ kubectl apply -f https://github.com/stashed/mongodb/raw/{{< param "info.subproject_version" >}}/docs/examples/backup/replicaset/mongodb-replicaset.yaml
mongodb.kubedb.com/sample-mgo-rs created
```

KubeDB will deploy a MongoDB database according to the above specification. It will also create the necessary secrets and services to access the database.

Let's check if the database is ready to use,

```console
$ kubectl get mg -n demo sample-mgo-rs
NAME            VERSION      STATUS    AGE
sample-mgo-rs   4.2.3       Running   1m
```

The database is `Running`. Verify that KubeDB has created a Secret and a Service for this database using the following commands,

```console
$ kubectl get secret -n demo -l=kubedb.com/name=sample-mgo-rs
NAME                 TYPE     DATA   AGE
sample-mgo-rs-auth   Opaque   2      117s
sample-mgo-rs-cert   Opaque   4      116s

$ kubectl get service -n demo -l=kubedb.com/name=sample-mgo-rs
NAME                TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)     AGE
sample-mgo-rs       ClusterIP   10.107.13.16   <none>        27017/TCP   2m14s
sample-mgo-rs-gvr   ClusterIP   None           <none>        27017/TCP   2m14s
```

KubeDB creates an [AppBinding](/docs/concepts/crds/appbinding.md) crd that holds the necessary information to connect with the database.

**Verify AppBinding:**

Verify that the `AppBinding` has been created successfully using the following command,

```console
$ kubectl get appbindings -n demo
NAME            AGE
sample-mgo-rs   58s
```

Let's check the YAML of the above `AppBinding`,

```console
$ kubectl get appbindings -n demo sample-mgo-rs -o yaml
```

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: sample-mgo-rs
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: mongodbs.kubedb.com
    kubedb.com/name: sample-mgo-rs
  name: sample-mgo-rs
  namespace: demo
spec:
  clientConfig:
    service:
      name: sample-mgo-rs
      port: 27017
      scheme: mongodb
  parameters:
    apiVersion: config.kubedb.com/v1alpha1
    kind: MongoConfiguration
    replicaSets:
      host-0: rs0/sample-mgo-rs-0.sample-mgo-rs-gvr.demo.svc,sample-mgo-rs-1.sample-mgo-rs-gvr.demo.svc,sample-mgo-rs-2.sample-mgo-rs-gvr.demo.svc
    stash:
      addon:
        backupTask:
          name: mongodb-backup-{{< param "info.subproject_version" >}}
        restoreTask:
          name: mongodb-restore-{{< param "info.subproject_version" >}}
  secret:
    name: sample-mgo-rs-auth
  type: kubedb.com/mongodb
  version: "4.2.3"
```

Stash uses the `AppBinding` crd to connect with the target database. It requires the following two fields to set in AppBinding's `Spec` section.

- `spec.clientConfig.service.name` specifies the name of the service that connects to the database.
- `spec.secret` specifies the name of the secret that holds necessary credentials to access the database.
- `spec.parameters.replicaSets` contains the dsn of replicaset. The DSNs are in key-value pair. If there is only one replicaset (replicaset can be multiple, because of sharding), then ReplicaSets field contains only one key-value pair where the key is host-0 and the value is dsn of that replicaset.
- `spec.parameters.stash` section specifies the Stash Addons that will be used to backup and restore this MongoDB.
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
  name: sample-mgo-rs-ssl
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
  clusterAuthMode: x509
  sslMode: requireSSL
```

After the deploy is done, kubedb will create a appbinding that will look like:

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: sample-mgo-rs-ssl
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: mongodbs.kubedb.com
    kubedb.com/name: sample-mgo-rs-ssl
  name: sample-mgo-rs-ssl
  namespace: demo
spec:
  clientConfig:
    caBundle: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUM0RENDQWNpZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFoTVJJd0VBWURWUVFLRXdscmRXSmwKWkdJNlkyRXhDekFKQmdOVkJBTVRBbU5oTUI0WERURTVNRGt6TURFME1UYzBORm9YRFRJNU1Ea3lOekUwTVRjMApORm93SVRFU01CQUdBMVVFQ2hNSmEzVmlaV1JpT21OaE1Rc3dDUVlEVlFRREV3SmpZVENDQVNJd0RRWUpLb1pJCmh2Y05BUUVCQlFBRGdnRVBBRENDQVFvQ2dnRUJBS2tMMmh6Q3ZzTXdUdUJ2cjJyTnlCcGZzcHduNGlLQS9ncloKZVM4RHFkTmJicVFZUy91WEhJbUFwaitnQ3lPTmJ2QU01b1c1VEgyMUtXWm1SSmxZUkxpWXhrVVdJS3kxa2lxMApvRTVCR0c4TWZaZEl6RXROWEpJUnpaSFoyVW5hMTZzWG5CNUFxWGQvSkJiY0tzWVM0TzczSFZLbUg1VWVTdVBjCnhRVk0xRVNlZHBmOVJLcVlBbnVxN0ZrTWE2RXJUaFk5VDM5ZjJDdWR1SVFzcFFUb2VmdjZPeEcrVWtBZkJNN2sKeThweEFKUk5hSjhXRmN2Z2s2NlZBK1V4SHUwRW1uK09SYWZlWG1VVFBOT09uaE9tam9PV2ZtbitWZmxNeWk3RgpkK2dYZWYwdHRSc0JROVFjb25ZR1N4TzBGR0dMR3pLdTlaelROSkFCcFhwcXd5cTVYeUVDQXdFQUFhTWpNQ0V3CkRnWURWUjBQQVFIL0JBUURBZ0trTUE4R0ExVWRFd0VCL3dRRk1BTUJBZjh3RFFZSktvWklodmNOQVFFTEJRQUQKZ2dFQkFGSXh1T1VKU0xWakVZTVQ1R2xWZ0o3QU1Remczd2dpV1lHR0J6SEo3UzBaMnJSb3FXUFpvNlVwZVlrWApTRDFUM1ZuTDJBQVhCYWpuVUtpb1ZZNlNFSlBMU3Urelg1VEx4TGs2SzRWMTZjOFJDT3lBYUFqM0NXTjk4bVNCCmhtTEliNjZYR1R1M3JBNlVpMFYxVU9JeW1GZS9jd09FMFExT3lzamwyZHQ2c3pITWtjNENFVG1nOFZsei9SVngKeVBDdGhmaGpCSm9oc25VcTkvU3FiamMyZC9yOW1lK1hNZjd3ZGUyMEZFU2g1cW1scUNybnp2UENHMEpkRlFVUgpDLzJQODJZcjJiUUgzckg2RkxiN0l4cFRLTWdDbjZVbUcxL3VacnpRTEpBTTAxUkxSQVZVRy9OTzBoYjJoWE1tCnZkaFNqTFM2N05Kakl5RVRRczc5QzFhZkt4RT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
    service:
      name: sample-mgo-rs-ssl
      port: 27017
      scheme: mongodb
  parameters:
    apiVersion: config.kubedb.com/v1alpha1
    kind: MongoConfiguration
    replicaSets:
      host-0: rs0/sample-mgo-rs-ssl-0.sample-mgo-rs-ssl-gvr.demo.svc,sample-mgo-rs-ssl-1.sample-mgo-rs-ssl-gvr.demo.svc,sample-mgo-rs-ssl-2.sample-mgo-rs-ssl-gvr.demo.svc
    stash:
      addon:
        backupTask:
          name: mongodb-backup-{{< param "info.subproject_version" >}}
        restoreTask:
          name: mongodb-restore-{{< param "info.subproject_version" >}}
  secret:
    name: sample-mgo-rs-ssl-cert
  type: kubedb.com/mongodb
  version: 4.2.3
```

Here, `mongo-sh-rs-cert` contains few required certificates, and one of them is `client.pem` which is required to backup/restore ssl enabled mongodb server using stash-mongodb.

**Creating AppBinding Manually:**

If you deploy MongoDB database without KubeDB, you have to create the AppBinding crd manually in the same namespace as the service and secret of the database.

**Insert Sample Data:**

Now, we are going to exec into the database pod and create some sample data. At first, find out the database pod using the following command,

```console
$ kubectl get pods -n demo --selector="kubedb.com/name=sample-mgo-rs"
NAME              READY   STATUS    RESTARTS   AGE
sample-mgo-rs-0   1/1     Running   0          16m
sample-mgo-rs-1   1/1     Running   0          15m
sample-mgo-rs-2   1/1     Running   0          15m
```

Now, let's exec into the pod and create a table,

```console
$ kubectl get secrets -n demo sample-mgo-rs-auth -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo sample-mgo-rs-auth -o jsonpath='{.data.\password}' | base64 -d
CRz6EuxvKdFjopfP

$ kubectl exec -it -n demo sample-mgo-rs-0 bash

mongodb@sample-mgo-rs-0:/$ mongo admin -u root -p CRz6EuxvKdFjopfP

rs0:PRIMARY> rs.isMaster().primary
sample-mgo-rs-0.sample-mgo-rs-gvr.demo.svc.cluster.local:27017

rs0:PRIMARY> show dbs
admin   0.000GB
config  0.000GB
local   0.000GB

rs0:PRIMARY> show users
{
	"_id" : "admin.root",
	"userId" : UUID("0e9345cc-27ea-4175-acc4-295c987ac06b"),
	"user" : "root",
	"db" : "admin",
	"roles" : [
		{
			"role" : "root",
			"db" : "admin"
		}
	]
}

rs0:PRIMARY> use newdb
switched to db newdb

rs0:PRIMARY> db.movie.insert({"name":"batman"});
WriteResult({ "nInserted" : 1 })

rs0:PRIMARY> db.movie.find().pretty()
{ "_id" : ObjectId("5d31b9d44db670db130d7a5c"), "name" : "batman" }

rs0:PRIMARY> exit
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
  name: gcs-repo-replicaset
  namespace: demo
spec:
  backend:
    gcs:
      bucket: appscode-qa
      prefix: demo/mongodb/sample-mgo-rs
    storageSecretName: gcs-secret
```

Let's create the `Repository` we have shown above,

```console
$ kubectl apply -f https://github.com/stashed/mongodb/raw/{{< param "info.subproject_version" >}}/docs/examples/backup/replicaset/repository-replicaset.yaml
repository.stash.appscode.com/gcs-repo-replicaset created
```

Now, we are ready to backup our database to our desired backend.

### Backup MongoDB ReplicaSet

We have to create a `BackupConfiguration` targeting respective AppBinding crd of our desired database. Then Stash will create a CronJob to periodically backup the database.

**Create BackupConfiguration:**

Below is the YAML for `BackupConfiguration` crd to backup the `sample-mgo-rs` database we have deployed earlier.,

```yaml
apiVersion: stash.appscode.com/v1beta1
kind: BackupConfiguration
metadata:
  name: sample-mgo-rs-backup
  namespace: demo
spec:
  schedule: "*/5 * * * *"
#  task: # Uncomment if you are not using KubeDB to deploy your database
#    name: mongodb-backup-{{< param "info.subproject_version" >}}
  repository:
    name: gcs-repo-replicaset
  target:
    ref:
      apiVersion: appcatalog.appscode.com/v1alpha1
      kind: AppBinding
      name: sample-mgo-rs
  retentionPolicy:
    name: keep-last-5
    keepLast: 5
    prune: true
```

Here,

- `spec.schedule` specifies that we want to backup the database at 5 minutes interval.
- `spec.task.name` specifies the name of the task crd that specifies the necessary Function and their execution order to backup a MongoDB database.
- `spec.target.ref` refers to the `AppBinding` crd that was created for `sample-mgo-rs` database.

Let's create the `BackupConfiguration` crd we have shown above,

```console
$ kubectl apply -f https://github.com/stashed/mongodb/raw/{{< param "info.subproject_version" >}}/docs/examples/backup/replicaset/backupconfiguration-replicaset.yaml
backupconfiguration.stash.appscode.com/sample-mgo-rs-backup created
```

**Verify CronJob:**

If everything goes well, Stash will create a CronJob with the schedule specified in `spec.schedule` field of `BackupConfiguration` crd.

Verify that the CronJob has been created using the following command,

```console
$ kubectl get cronjob -n demo
NAME                   SCHEDULE      SUSPEND   ACTIVE   LAST SCHEDULE   AGE
sample-mgo-rs-backup   */5 * * * *   False     0        <none>          62s
```

**Wait for BackupSession:**

The `sample-mgo-rs-backup` CronJob will trigger a backup on each schedule by creating a `BackupSession` crd.

Wait for the next schedule. Run the following command to watch `BackupSession` crd,

```console
$ kubectl get backupsession -n demo -w
NAME                              INVOKER-TYPE          INVOKER-NAME           PHASE       AGE
sample-mgo-rs-backup-1563540308   BackupConfiguration   sample-mgo-rs-backup   Running     5m19s
sample-mgo-rs-backup-1563540308   BackupConfiguration   sample-mgo-rs-backup   Succeeded   5m45s
```

We can see above that the backup session has succeeded. Now, we are going to verify that the backed up data has been stored in the backend.

**Verify Backup:**

Once a backup is complete, Stash will update the respective `Repository` crd to reflect the backup. Check that the repository `gcs-repo-replicaset` has been updated by the following command,

```console
$ kubectl get repository -n demo gcs-repo-replicaset
NAME                  INTEGRITY   SIZE        SNAPSHOT-COUNT   LAST-SUCCESSFUL-BACKUP   AGE
gcs-repo-replicaset   true        3.844 KiB   2                14s                      10m
```

Now, if we navigate to the GCS bucket, we are going to see backed up data has been stored in `demo/mongodb/sample-mgo-rs` directory as specified by `spec.backend.gcs.prefix` field of Repository crd.

> Note: Stash keeps all the backed up data encrypted. So, data in the backend will not make any sense until they are decrypted.

## Restore MongoDB ReplicaSet

In this section, we are going to restore the database from the backup we have taken in the previous section. We are going to deploy a new replicaset database and initialize it from the backup.

**Stop Taking Backup of the Old Database:**

At first, let's stop taking any further backup of the old database so that no backup is taken during restore process. We are going to pause the `BackupConfiguration` crd that we had created to backup the `sample-mgo-rs` database. Then, Stash will stop taking any further backup for this database.

Let's pause the `sample-mgo-rs-backup` BackupConfiguration,

```console
$ kubectl patch backupconfiguration -n demo sample-mgo-rs-backup --type="merge" --patch='{"spec": {"paused": true}}'
backupconfiguration.stash.appscode.com/sample-mgo-rs-backup patched
```

Now, wait for a moment. Stash will pause the BackupConfiguration. Verify that the BackupConfiguration  has been paused,

```console
$ kubectl get backupconfiguration -n demo sample-mgo-rs-backup
NAME                  TASK                       SCHEDULE      PAUSED   AGE
sample-mgo-rs-backup  mongodb-backup-{{< param "info.subproject_version" >}}      */5 * * * *   true     26m
```

Notice the `PAUSED` column. Value `true` for this field means that the BackupConfiguration has been paused.

**Deploy Restored Database:**

Now, we have to deploy the restored database similarly as we have deployed the original `sample-mgo-rs` database. However, this time there will be the following differences:

- We have to specify `spec.init.waitForInitialRestore` section that tells KubeDB to wait until the first restore completes before marking this database ready to use.

Below is the YAML for `MongoDB` crd we are going deploy to initialize from backup,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  name: restored-mgo-rs
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
  init:
    waitForInitialRestore: true
```

Let's create the above database,

```console
$ kubectl apply -f https://github.com/stashed/mongodb/raw/{{< param "info.subproject_version" >}}/docs/examples/restore/replicaset/restored-mongodb-replicaset.yaml
mongodb.kubedb.com/restored-mgo-rs created
```

If you check the database status, you will see it is stuck in `Provisioning` state.

```console
$ kubectl get mg -n demo restored-mgo-rs
NAME              VERSION        STATUS         AGE
restored-mgo-rs   4.2.3         Provisioning   2m
```

**Create RestoreSession:**

Now, we need to create a `RestoreSession` crd pointing to the AppBinding for this restored database.

Check AppBinding has been created for the `restored-mgo-rs` database using the following command,

```console
$ kubectl get appbindings -n demo restored-mgo-rs
NAME               AGE
restored-mgo-rs    29s
```

NB. The appbinding `restored-mgo-rs` also contains `spec.parametrs` field. the number of hosts in `spec.parameters.replicaSets` needs to be similar to the old appbinding. Otherwise, the replicaset recover may not be accurate.

> If you are not using KubeDB to deploy database, create the AppBinding manually.

Below is the YAML for the `RestoreSession` crd that we are going to create to restore backed up data into `restored-mgo-rs` database.

```yaml
apiVersion: stash.appscode.com/v1beta1
kind: RestoreSession
metadata:
  name: sample-mgo-rs-restore
  namespace: demo
spec:
#  task: # Uncomment if you are not using KubeDB to deploy your database
#    name: mongodb-restore-{{< param "info.subproject_version" >}}
  repository:
    name: gcs-repo-replicaset
  target:
    ref:
      apiVersion: appcatalog.appscode.com/v1alpha1
      kind: AppBinding
      name: restored-mgo-rs
  rules:
  - snapshots: [latest]
```

Here,

- `spec.task.name` specifies the name of the `Task` crd that specifies the Functions and their execution order to restore a MongoDB database.
- `spec.repository.name` specifies the `Repository` crd that holds the backend information where our backed up data has been stored.
- `spec.target.ref` refers to the AppBinding crd for the `restored-mgo-rs` database.
- `spec.rules` specifies that we are restoring from the latest backup snapshot of the database.

Let's create the `RestoreSession` crd we have shown above,

```console
$ kubectl apply -f https://github.com/stashed/mongodb/raw/{{< param "info.subproject_version" >}}/docs/examples/restore/replicaset/restoresession-replicaset.yaml
restoresession.stash.appscode.com/sample-mgo-rs-restore created
```

Once, you have created the `RestoreSession` crd, Stash will create a job to restore. We can watch the `RestoreSession` phase to check if the restore process is succeeded or not.

Run the following command to watch `RestoreSession` phase,

```console
$ kubectl get restoresession -n demo sample-mgo-rs-restore -w
NAME                    REPOSITORY-NAME        PHASE       AGE
sample-mgo-rs-restore   gcs-repo-replicaset    Running     5s
sample-mgo-rs-restore   gcs-repo-replicaset    Succeeded   43s
```

So, we can see from the output of the above command that the restore process succeeded.

**Verify Restored Data:**

In this section, we are going to verify that the desired data has been restored successfully. We are going to connect to `mongos` and check whether the table we had created in the original database is restored or not.

At first, check if the database has gone into `Running` state by the following command,

```console
$ kubectl get mg -n demo restored-mgo-rs
NAME              VERSION        STATUS    AGE
restored-mgo-rs   4.2.3         Running   3m
```

Now, exec into the database pod and list available tables,

```console
$ kubectl get secrets -n demo sample-mgo-rs-auth -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo sample-mgo-rs-auth -o jsonpath='{.data.\password}' | base64 -d
CRz6EuxvKdFjopfP

$ kubectl exec -it -n demo restored-mgo-rs-0 bash

mongodb@restored-mgo-rs-0:/$ mongo admin -u root -p CRz6EuxvKdFjopfP

rs0:PRIMARY> rs.isMaster().primary
restored-mgo-rs-0.restored-mgo-rs-gvr.demo.svc.cluster.local:27017

rs0:PRIMARY> show dbs
admin   0.000GB
config  0.000GB
local   0.000GB
newdb   0.000GB

rs0:PRIMARY> show users
{
	"_id" : "admin.root",
	"userId" : UUID("00f521b5-2b43-4712-ba80-efaa6b382813"),
	"user" : "root",
	"db" : "admin",
	"roles" : [
		{
			"role" : "root",
			"db" : "admin"
		}
	]
}

rs0:PRIMARY> use newdb
switched to db newdb

rs0:PRIMARY> db.movie.find().pretty()
{ "_id" : ObjectId("5d31b9d44db670db130d7a5c"), "name" : "batman" }

rs0:PRIMARY> exit
bye
```

So, from the above output, we can see the database `newdb` that we had created in the original database `sample-mgo-rs` is restored in the restored database `restored-mgo-rs`.

## Backup MongoDB ReplicaSet Cluster and Restore into a Standalone database

It is possible to take backup of a MongoDB ReplicaSet Cluster and restore it into a standalone database, but user need to create the appbinding for this process.

### Backup a replicaset cluster

Keep all the fields of appbinding that is explained earlier in this guide, except `spec.parameter`. Do not set `spec.parameter.configServer` and `spec.parameter.replicaSet`. By doing this, the job will use `spec.clientConfig.service.name` as host, which is replicaset DSN. So, the backup will treat this cluster as a standalone and will skip the [`idiomatic way` of taking backups of a replicaset cluster](https://docs.mongodb.com/manual/tutorial/restore-replica-set-from-backup/). Then follow the rest of the procedure as described above.

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  name: sample-mgo-rs-custom
  namespace: demo
spec:
  clientConfig:
    service:
      name: sample-mgo-rs
      port: 27017
      scheme: mongodb
  secret:
    name: sample-mgo-rs-auth
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
      prefix: demo/mongodb/sample-mgo-rs/standalone
    storageSecretName: gcs-secret

---
apiVersion: stash.appscode.com/v1beta1
kind: BackupConfiguration
metadata:
  name: sample-mgo-rs-backup2
  namespace: demo
spec:
  schedule: "*/5 * * * *"
  task:
    name: mongodb-backup-{{< param "info.subproject_version" >}}
  repository:
    name: gcs-repo-custom
  target:
    ref:
      apiVersion: appcatalog.appscode.com/v1alpha1
      kind: AppBinding
      name: sample-mgo-rs-custom
  retentionPolicy:
    name: keep-last-5
    keepLast: 5
    prune: true
```

This time, we have to provide the Stash Addon information in `spec.task` section of `BackupConfiguration` object as it does not present in the `AppBinding` object that we are creating manually.

```console
$ kubectl create -f https://github.com/stashed/mongodb/raw/{{< param "info.subproject_version" >}}/docs/examples/backup/replicaset/standalone-backup.yaml
appbinding.appcatalog.appscode.com/sample-mgo-rs-custom created
repository.stash.appscode.com/gcs-repo-custom created
backupconfiguration.stash.appscode.com/sample-mgo-rs-backup2 created


$ kubectl get backupsession -n demo
NAME                               BACKUPCONFIGURATION    PHASE       AGE
sample-mgo-rs-backup2-1563541509   sample-mgo-rs-backup   Succeeded   35s


$ kubectl get repository -n demo gcs-repo-custom
NAME              INTEGRITY   SIZE        SNAPSHOT-COUNT   LAST-SUCCESSFUL-BACKUP   AGE
gcs-repo-custom   true        1.640 KiB   1                1m                       5m
```

### Restore to a standalone database

No additional configuration is needed to restore the replicaset cluster to a standalone database. Follow the normal procedure of restoring a MongoDB Database.

Standalone MongoDB,

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

You have to provide the respective Stash restore `Task` info in `spec.task` section of `RestoreSession` object,

```yaml
apiVersion: stash.appscode.com/v1beta1
kind: RestoreSession
metadata:
  name: sample-mongodb-restore
  namespace: demo
spec:
  task:
    name: mongodb-restore-{{< param "info.subproject_version" >}}
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
$ kubectl create -f https://github.com/stashed/mongodb/raw/{{< param "info.subproject_version" >}}/docs/examples/restore/replicaset/restored-standalone.yaml
mongodb.kubedb.com/restored-mongodb created

$ kubectl get mg -n demo restored-mongodb
NAME               VERSION        STATUS         AGE
restored-mongodb   4.2.3         Provisioning   56s

$ kubectl create -f https://github.com/stashed/mongodb/raw/{{< param "info.subproject_version" >}}/docs/examples/restore/replicaset/restoresession-standalone.yaml
restoresession.stash.appscode.com/sample-mongodb-restore created

$ kubectl get mg -n demo restored-mongodb
NAME               VERSION        STATUS    AGE
restored-mongodb   4.2.3         Running   2m
```

Now, exec into the database pod and list available tables,

```console
$ kubectl get secrets -n demo sample-mgo-rs-auth -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo sample-mgo-rs-auth -o jsonpath='{.data.\password}' | base64 -d
CRz6EuxvKdFjopfP

$ kubectl exec -it -n demo restored-mongodb-0 bash

mongodb@restored-mongodb-0:/$ mongo admin -u root -p CRz6EuxvKdFjopfP

> show dbs
admin   0.000GB
config  0.000GB
local   0.000GB
newdb   0.000GB

> show users
{
	"_id" : "admin.root",
	"userId" : UUID("11e00a38-7b08-4864-b452-ae356350e50f"),
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
{ "_id" : ObjectId("5d31b9d44db670db130d7a5c"), "name" : "batman" }

> exit
bye
```

So, from the above output, we can see the database `newdb` that we had created in the original database `sample-mgo-rs` is restored in the restored database `restored-mongodb`.

## Cleanup

To cleanup the Kubernetes resources created by this tutorial, run:

```console
kubectl delete -n demo restoresession sample-mgo-rs-restore sample-mongodb-restore
kubectl delete -n demo backupconfiguration sample-mgo-rs-backup sample-mgo-rs-backup2
kubectl delete -n demo mg sample-mgo-rs sample-mgo-rs-ssl restored-mgo-rs restored-mgo-rs restored-mongodb
kubectl delete -n demo repository gcs-repo-replicaset gcs-repo-custom
kubectl delete -n demo appbinding sample-mgo-rs-custom
```
