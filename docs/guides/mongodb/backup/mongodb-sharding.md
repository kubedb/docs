---
title: Backup & Restore Sharded MongoDB Cluster| Stash
description: Backup and restore sharded MongoDB cluster using Stash
menu:
  docs_{{ .version }}:
    identifier: sharded-mongodb-{{ .subproject_version }}
    name: MongoDB Sharded Cluster
    parent: stash-mongodb-guides-{{ .subproject_version }}
    weight: 30
product_name: stash
menu_name: docs_{{ .version }}
section_menu_id: stash-addons
---

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [Stash Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Backup and Restore MongoDB Sharded Clusters using Stash

Stash 0.9.0+ supports taking [backup](https://docs.mongodb.com/manual/tutorial/backup-sharded-cluster-with-database-dumps/) and [restores](https://docs.mongodb.com/manual/tutorial/restore-sharded-cluster/) MongoDB Sharded clusters in ["idiomatic" way](https://docs.mongodb.com/manual/administration/backup-sharded-clusters/). This guide will show you how you can backup and restore your MongoDB Sharded clusters with Stash.

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
  version: 4.2.3
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
$ kubectl apply -f https://github.com/stashed/mongodb/raw/{{< param "info.subproject_version" >}}/docs/examples/backup/sharding/mongodb-sharding.yaml
mongodb.kubedb.com/sample-mgo-sh created
```

KubeDB will deploy a MongoDB database according to the above specification. It will also create the necessary secrets and services to access the database.

Let's check if the database is ready to use,

```console
$ kubectl get mg -n demo sample-mgo-sh
NAME            VERSION        STATUS    AGE
sample-mgo-sh   4.2.3         Running   35m
```

The database is `Running`. Verify that KubeDB has created a Secret and a Service for this database using the following commands,

```console
$ kubectl get secret -n demo -l=kubedb.com/name=sample-mgo-sh
NAME                 TYPE     DATA   AGE
sample-mgo-sh-auth   Opaque   2      36m
sample-mgo-sh-cert   Opaque   4      36m

$ kubectl get service -n demo -l=kubedb.com/name=sample-mgo-sh
NAME                          TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)     AGE
sample-mgo-sh                 ClusterIP   10.107.11.117   <none>        27017/TCP   36m
sample-mgo-sh-configsvr-gvr   ClusterIP   None            <none>        27017/TCP   36m
sample-mgo-sh-shard0-gvr      ClusterIP   None            <none>        27017/TCP   36m
sample-mgo-sh-shard1-gvr      ClusterIP   None            <none>        27017/TCP   36m
sample-mgo-sh-shard2-gvr      ClusterIP   None            <none>        27017/TCP   36m
```

KubeDB creates an [AppBinding](/docs/concepts/crds/appbinding.md) crd that holds the necessary information to connect with the database.

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
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: sample-mgo-sh
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: mongodbs.kubedb.com
    kubedb.com/name: sample-mgo-sh
  name: sample-mgo-sh
  namespace: demo
spec:
  clientConfig:
    service:
      name: sample-mgo-sh
      port: 27017
      scheme: mongodb
  parameters:
    apiVersion: config.kubedb.com/v1alpha1
    kind: MongoConfiguration
    configServer: cnfRepSet/sample-mgo-sh-configsvr-0.sample-mgo-sh-configsvr-gvr.demo.svc:27017,sample-mgo-sh-configsvr-1.sample-mgo-sh-configsvr-gvr.demo.svc:27017,sample-mgo-sh-configsvr-2.sample-mgo-sh-configsvr-gvr.demo.svc:27017
    replicaSets:
      host-0: shard0/sample-mgo-sh-shard0-0.sample-mgo-sh-shard0-gvr.demo.svc:27017,sample-mgo-sh-shard0-1.sample-mgo-sh-shard0-gvr.demo.svc:27017,sample-mgo-sh-shard0-2.sample-mgo-sh-shard0-gvr.demo.svc:27017
      host-1: shard1/sample-mgo-sh-shard1-0.sample-mgo-sh-shard1-gvr.demo.svc:27017,sample-mgo-sh-shard1-1.sample-mgo-sh-shard1-gvr.demo.svc:27017,sample-mgo-sh-shard1-2.sample-mgo-sh-shard1-gvr.demo.svc:27017
      host-2: shard2/sample-mgo-sh-shard2-0.sample-mgo-sh-shard2-gvr.demo.svc:27017,sample-mgo-sh-shard2-1.sample-mgo-sh-shard2-gvr.demo.svc:27017,sample-mgo-sh-shard2-2.sample-mgo-sh-shard2-gvr.demo.svc:27017
    stash:
      addon:
        backupTask:
          name: mongodb-backup-{{< param "info.subproject_version" >}}
        restoreTask:
          name: mongodb-restore-{{< param "info.subproject_version" >}}
  secret:
    name: sample-mgo-sh-auth
  type: kubedb.com/mongodb
  version: 4.2.3
```

Stash uses the `AppBinding` crd to connect with the target database. It requires the following two fields to set in AppBinding's `Spec` section.

- `spec.clientConfig.service.name` specifies the name of the service that connects to the database.
- `spec.secret` specifies the name of the secret that holds necessary credentials to access the database.
- `spec.parameters.configServer` specifies the dsn of config server of mongodb sharding. The dsn includes the port no too.
- `spec.parameters.replicaSets` contains the dsn of each replicaset of sharding. The DSNs are in key-value pair, where the keys are host-0, host-1 etc, and the values are DSN of each replicaset. If there is no sharding but only one replicaset, then ReplicaSets field contains only one key-value pair where the key is host-0 and the value is dsn of that replicaset.
- `spec.parameters.stash` contains the Stash addon information that will be used to backup and restore this MongoDB.
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
  name: sample-mgo-sh-ssl
  namespace: demo
spec:
  version: 4.2.3
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
    app.kubernetes.io/instance: sample-mgo-sh-ssl
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: mongodbs.kubedb.com
    kubedb.com/name: sample-mgo-sh-ssl
  name: sample-mgo-sh-ssl
  namespace: demo
spec:
  clientConfig:
    caBundle: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUM0RENDQWNpZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREFoTVJJd0VBWURWUVFLRXdscmRXSmwKWkdJNlkyRXhDekFKQmdOVkJBTVRBbU5oTUI0WERURTVNVEF3TWpBMU5UTXlPRm9YRFRJNU1Ea3lPVEExTlRNeQpPRm93SVRFU01CQUdBMVVFQ2hNSmEzVmlaV1JpT21OaE1Rc3dDUVlEVlFRREV3SmpZVENDQVNJd0RRWUpLb1pJCmh2Y05BUUVCQlFBRGdnRVBBRENDQVFvQ2dnRUJBTi9wY2JsQVdGUkNvc1JuWExNK1ZZUDJ5Z1hSajVuQTZsc2sKZU1RRnZaeFdTeE9RUnY4ODBWMG1UOGV6SWE1TmdRUm9XaVZxNm9sMVdwR3ZMVzBia1FyUEZ2M1lTTG5IeDRFMgoxdlR3VzMvM2kvY1M5MGZzcTc1TVJabG5ZMjhZNlhZcU14N05iYnVSUWM2Z2pkYm50Y1dtWmZ1TUNXWHRlWnAvCnBRMThoVVJodHRKNHR5RHh2djlWSlNzZ3JPQTlMVWc2WU5xamJBM0p2OXBLTjRPVzlaNG11dTFxeUpsZ3RNOHMKNUNUaDhtZlZvc2NjbSt5eFpXZTByY1EyVWwwL21RNVhKcTYvbHdyVy9wVGF5S3BQSkprc2tDZzl1cDc5eWJJcgp5OEpVQlNaQXZPZC9JbEkrUVk5OXZ1cUdCNzZSeGRDZHlMUURBMUxaR1BIUm12ZUlybWtDQXdFQUFhTWpNQ0V3CkRnWURWUjBQQVFIL0JBUURBZ0trTUE4R0ExVWRFd0VCL3dRRk1BTUJBZjh3RFFZSktvWklodmNOQVFFTEJRQUQKZ2dFQkFBdzUwdkZhSHVFVHZiay9vVW1udEVPaGhMTmI5WllrREZDdVZzRWZRdnRhdkVUT3dJNFVlaS9GUnFsTwpab3JNZEF6c0V0dDhwSVc5aXJzK0ZSakxUTjk3SnZFL29LbzlNNXlmLy9kZHRRWW1ZNjFTZzVIdjVQWWJ6ZzI5Cm9POHdZTkR2STQzT1Y2aUtEMXFJSE1meEcyZ0l5aXNod1JJeXJLMUp6UVRMcVEzSGJSU0tMNldKdFppeFIwVUwKcVR5Wk5jWFFKVVY2Yk9FMjVSSHdvWGVJUFNQanh6T1o0L3g1bnZOMU5rVkZJMFZ1ZGVBdU85Q0ZaaW9UNy8zbQpNV3VSQytDRDcyMUd1RzlhZmZmdU5CNGtKNnlvUmgxT093THZrT3hid0tveEVCR1B1UlFHem1KV3YrbEhOWVpHClg2dExwRkFaRHA3R3ZiZ1I3RnR3ampJb0N5TT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
    service:
      name: sample-mgo-sh-ssl
      port: 27017
      scheme: mongodb
  parameters:
    apiVersion: config.kubedb.com/v1alpha1
    kind: MongoConfiguration
    configServer: cnfRepSet/sample-mgo-sh-ssl-configsvr-0.sample-mgo-sh-ssl-configsvr-gvr.demo.svc:27017,sample-mgo-sh-ssl-configsvr-1.sample-mgo-sh-ssl-configsvr-gvr.demo.svc:27017,sample-mgo-sh-ssl-configsvr-2.sample-mgo-sh-ssl-configsvr-gvr.demo.svc:27017
    replicaSets:
      host-0: shard0/sample-mgo-sh-ssl-shard0-0.sample-mgo-sh-ssl-shard0-gvr.demo.svc:27017,sample-mgo-sh-ssl-shard0-1.sample-mgo-sh-ssl-shard0-gvr.demo.svc:27017,sample-mgo-sh-ssl-shard0-2.sample-mgo-sh-ssl-shard0-gvr.demo.svc:27017
      host-1: shard1/sample-mgo-sh-ssl-shard1-0.sample-mgo-sh-ssl-shard1-gvr.demo.svc:27017,sample-mgo-sh-ssl-shard1-1.sample-mgo-sh-ssl-shard1-gvr.demo.svc:27017,sample-mgo-sh-ssl-shard1-2.sample-mgo-sh-ssl-shard1-gvr.demo.svc:27017
      host-2: shard2/sample-mgo-sh-ssl-shard2-0.sample-mgo-sh-ssl-shard2-gvr.demo.svc:27017,sample-mgo-sh-ssl-shard2-1.sample-mgo-sh-ssl-shard2-gvr.demo.svc:27017,sample-mgo-sh-ssl-shard2-2.sample-mgo-sh-ssl-shard2-gvr.demo.svc:27017
    stash:
      addon:
        backupTask:
          name: mongodb-backup-{{< param "info.subproject_version" >}}
        restoreTask:
          name: mongodb-restore-{{< param "info.subproject_version" >}}
  secret:
    name: sample-mgo-sh-ssl-cert
  type: kubedb.com/mongodb
  version: 4.2.3
```

Here, `sample-mgo-sh-cert` contains few required certificates, and one of them is `client.pem` which is required to backup/restore ssl enabled mongodb server using stash-mongodb.

**Creating AppBinding Manually:**

If you deploy MongoDB database without KubeDB, you have to create the AppBinding crd manually in the same namespace as the service and secret of the database.

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
$ kubectl get secrets -n demo sample-mgo-sh-auth -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo sample-mgo-sh-auth -o jsonpath='{.data.\password}' | base64 -d
JJPcMxNKJev0SzgX

$ kubectl exec -it -n demo sample-mgo-sh-mongos-9459cfc44-4jthd bash

mongodb@sample-mgo-sh-0:/$ mongo admin -u root -p JJPcMxNKJev0SzgX

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
$ kubectl apply -f https://github.com/stashed/mongodb/raw/{{< param "info.subproject_version" >}}/docs/examples/backup/sharding/repository-sharding.yaml
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
#  task: # Uncomment if you are not using KubeDB to deploy your database
#    name: mongodb-backup-{{< param "info.subproject_version" >}}
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
- `spec.task.name` specifies the name of the task crd that specifies the necessary Function and their execution order to backup a MongoDB database.
- `spec.target.ref` refers to the `AppBinding` crd that was created for `sample-mgo-sh` database.

Let's create the `BackupConfiguration` crd we have shown above,

```console
$ kubectl apply -f https://github.com/stashed/mongodb/raw/{{< param "info.subproject_version" >}}/docs/examples/backup/sharding/backupconfiguration-sharding.yaml
backupconfiguration.stash.appscode.com/sample-mgo-sh-backup created
```

**Verify CronJob:**

If everything goes well, Stash will create a CronJob with the schedule specified in `spec.schedule` field of `BackupConfiguration` crd.

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

In this section, we are going to restore the database from the backup we have taken in the previous section. We are going to deploy a new sharded database and initialize it from the backup.

**Stop Taking Backup of the Old Database:**

At first, let's stop taking any further backup of the old database so that no backup is taken during restore process. We are going to pause the `BackupConfiguration` crd that we had created to backup the `sample-mgo-sh` database. Then, Stash will stop taking any further backup for this database.

Let's pause the `sample-mgo-sh-backup` BackupConfiguration,

```console
$ kubectl patch backupconfiguration -n demo sample-mgo-sh-backup --type="merge" --patch='{"spec": {"paused": true}}'
backupconfiguration.stash.appscode.com/sample-mgo-sh-backup patched
```

Now, wait for a moment. Stash will pause the BackupConfiguration. Verify that the BackupConfiguration  has been paused,

```console
$ kubectl get backupconfiguration -n demo sample-mgo-sh-backup
NAME                  TASK                         SCHEDULE      PAUSED   AGE
sample-mgo-sh-backup  mongodb-backup-{{< param "info.subproject_version" >}}        */5 * * * *   true     26m
```

Notice the `PAUSED` column. Value `true` for this field means that the BackupConfiguration has been paused.

**Deploy Restored Database:**

Now, we have to deploy the restored database similarly as we have deployed the original `sample-mgo-sh` database. However, this time there will be the following differences:

- We have to specify `spec.init.waitForInitialRestore` field that tells KubeDB to wait until first restore completes before marking this MongoDB ready to use.

Below is the YAML for `MongoDB` crd we are going deploy to initialize from backup,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MongoDB
metadata:
  name: restored-mgo-sh
  namespace: demo
spec:
  version: 4.2.3
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
  init:
    waitForInitialRestore: true
  terminationPolicy: WipeOut
```

Let's create the above database,

```console
$ kubectl apply -f https://github.com/stashed/mongodb/raw/{{< param "info.subproject_version" >}}/docs/examples/restore/sharding/restored-mongodb-sharding.yaml
mongodb.kubedb.com/restored-mgo-sh created
```

If you check the database status, you will see it is stuck in `Provisioning` state.

```console
$ kubectl get mg -n demo restored-mgo-sh
NAME              VERSION        STATUS         AGE
restored-mgo-sh   4.2.3         Provisioning   48m
```

**Create RestoreSession:**

Now, we need to create a `RestoreSession` crd pointing to the AppBinding for this restored database.

Check AppBinding has been created for the `restored-mgo-sh` database using the following command,

```console
$ kubectl get appbindings -n demo restored-mgo-sh
NAME               AGE
restored-mgo-sh    29s
```

NB. The appbinding `restored-mgo-sh` also contains `spec.parametrs` field. the number of hosts in `spec.parameters.replicaSets` needs to be similar to the old appbinding. Otherwise, the sharding recover may not be accurate.

> If you are not using KubeDB to deploy database, create the AppBinding manually.

Below is the YAML for the `RestoreSession` crd that we are going to create to restore backed up data into `restored-mgo-sh` database.

```yaml
apiVersion: stash.appscode.com/v1beta1
kind: RestoreSession
metadata:
  name: sample-mgo-sh-restore
  namespace: demo
spec:
#  task: # Uncomment if you are not using KubeDB to deploy your database
#    name: mongodb-restore-{{< param "info.subproject_version" >}}
  repository:
    name: gcs-repo-sharding
  target:
    ref:
      apiVersion: appcatalog.appscode.com/v1alpha1
      kind: AppBinding
      name: restored-mgo-sh
  rules:
  - snapshots: [latest]
```

Here,

- `spec.task.name` specifies the name of the `Task` crd that specifies the Functions and their execution order to restore a MongoDB database.
- `spec.repository.name` specifies the `Repository` crd that holds the backend information where our backed up data has been stored.
- `spec.target.ref` refers to the AppBinding crd for the `restored-mgo-sh` database.
- `spec.rules` specifies that we are restoring from the latest backup snapshot of the database.

Let's create the `RestoreSession` crd we have shown above,

```console
$ kubectl apply -f https://github.com/stashed/mongodb/raw/{{< param "info.subproject_version" >}}/docs/examples/restore/sharding/restoresession-sharding.yaml
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

**Verify Restored Data:**

In this section, we are going to verify that the desired data has been restored successfully. We are going to connect to `mongos` and check whether the table we had created in the original database is restored or not.

At first, check if the database has gone into `Running` state by the following command,

```console
$ kubectl get mg -n demo restored-mgo-sh
NAME              VERSION        STATUS    AGE
restored-mgo-sh   4.2.3         Running   2h
```

Now, find out the `mongos` pod,

```console
$ kubectl get pods -n demo --selector="mongodb.kubedb.com/node.mongos=restored-mgo-sh-mongos"
NAME                                      READY   STATUS    RESTARTS   AGE
restored-mgo-sh-mongos-7bccd5d684-2z5xs   1/1     Running   0          169m
restored-mgo-sh-mongos-7bccd5d684-vvdxb   1/1     Running   0          169m
```

Now, exec into the database pod and list available tables,

```console
$ kubectl get secrets -n demo sample-mgo-sh-auth -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo sample-mgo-sh-auth -o jsonpath='{.data.\password}' | base64 -d
JJPcMxNKJev0SzgX

$ kubectl exec -it -n demo restored-mgo-sh-mongos-7bccd5d684-2z5xs bash

mongodb@restored-mgo-sh-0:/$ mongo admin -u root -p JJPcMxNKJev0SzgX

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

So, from the above output, we can see the database `newdb` that we had created in the original database `sample-mgo-sh` is restored in the restored database `restored-mgo-sh`.

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
    name: mongodb-backup-{{< param "info.subproject_version" >}}
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
$ kubectl create -f https://github.com/stashed/mongodb/raw/{{< param "info.subproject_version" >}}/docs/examples/backup/sharding/standalone-backup.yaml
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

This time, we have to provide `spec.task` section in `RestoreSession` object,

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
$ kubectl create -f https://github.com/stashed/mongodb/raw/{{< param "info.subproject_version" >}}/docs/examples/restore/sharding/restored-standalone.yaml
mongodb.kubedb.com/restored-mongodb created

$ kubectl get mg -n demo restored-mongodb
NAME               VERSION        STATUS         AGE
restored-mongodb   4.2.3         Provisioning   56s

$ kubectl create -f https://github.com/stashed/mongodb/raw/{{< param "info.subproject_version" >}}/docs/examples/restore/sharding/restoresession-standalone.yaml
restoresession.stash.appscode.com/sample-mongodb-restore created

$ kubectl get mg -n demo restored-mongodb
NAME               VERSION        STATUS         AGE
restored-mongodb   4.2.3         Running   56s
```

Now, exec into the database pod and list available tables,

```console
$ kubectl get secrets -n demo sample-mgo-sh-auth -o jsonpath='{.data.\username}' | base64 -d
root

$ kubectl get secrets -n demo sample-mgo-sh-auth -o jsonpath='{.data.\password}' | base64 -d
JJPcMxNKJev0SzgX

$ kubectl exec -it -n demo restored-mongodb-0 bash

mongodb@restored-mongodb-0:/$ mongo admin -u root -p JJPcMxNKJev0SzgX

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
kubectl delete -n demo mg sample-mgo-sh sample-mgo-sh-ssl restored-mgo-sh restored-mgo-sh restored-mongodb
kubectl delete -n demo repository gcs-repo-sharding gcs-repo-custom
```
