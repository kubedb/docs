---
title: Application Level Backup & Restore MongoDB | KubeStash
description: Application Level Backup and Restore using KubeStash
menu:
  docs_{{ .version }}:
    identifier: guides-mongodb-backup-kubestash-application-level
    name: Application Level Backup
    parent: guides-mongodb-backup-stashv2
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Application Level Backup and Restore MongoDB database using KubeStash

KubeStash offers application-level backup and restore functionality for `MongoDB` databases. It captures both manifest and data backups of any `MongoDB` database in a single snapshot. During the restore process, KubeStash first applies the `MongoDB` manifest to the cluster and then restores the data into it.

This guide will give you an overview how you can take application-level backup and restore your `MongoDB` databases using `Kubestash`.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using `Minikube` or `Kind`.
- Install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md).
- Install `KubeStash` in your cluster following the steps [here](https://kubestash.com/docs/latest/setup/install/kubestash).
- Install KubeStash `kubectl` plugin following the steps [here](https://kubestash.com/docs/latest/setup/install/kubectl-plugin/).
- If you are not familiar with how KubeStash backup and restore MongoDB databases, please check the following guide [here](/docs/guides/mongodb/backup/kubestash/overview/index.md).

You should be familiar with the following `KubeStash` concepts:

- [BackupStorage](https://kubestash.com/docs/latest/concepts/crds/backupstorage/)
- [BackupConfiguration](https://kubestash.com/docs/latest/concepts/crds/backupconfiguration/)
- [BackupSession](https://kubestash.com/docs/latest/concepts/crds/backupsession/)
- [RestoreSession](https://kubestash.com/docs/latest/concepts/crds/restoresession/)
- [Addon](https://kubestash.com/docs/latest/concepts/crds/addon/)
- [Function](https://kubestash.com/docs/latest/concepts/crds/function/)
- [Task](https://kubestash.com/docs/latest/concepts/crds/addon/#task-specification)

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/guides/mongodb/backup/kubestash/application-level/examples](https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mongodb/backup/kubestash/application-level/examples) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Backup MongoDB

KubeStash supports backups for `MongoDB` instances across different configurations, including Replica Set and Shard setups. In this demonstration, we'll focus on a `MongoDB` database using Replica Set configuration. The backup and restore process is similar for Standalone and Shard configuration.

This section will demonstrate how to take application-level backup of a `MongoDB` database. Here, we are going to deploy a `MongoDB` database using KubeDB. Then, we are going to back up the database at the application level to a `S3` bucket. Finally, we will restore the entire `MongoDB` database.

### Deploy Sample MongoDB Database

Let's deploy a sample `MongoDB` database and insert some data into it.

**Create MongoDB CR:**

Below is the YAML of a sample `MongoDB` CR that we are going to create for this tutorial:

```yaml
apiVersion: kubedb.com/v1
kind: MongoDB
metadata:
  name: sample-mongodb
  namespace: demo
spec:
  version: "4.4.26"
  replicaSet:
    name: "replicaset"
  replicas: 3
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
```

Create the above `MongoDB` CR,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mongodb/backup/kubestash/application-level/examples/sample-mongodb.yaml
mongodb.kubedb.com/sample-mongodb created
```

KubeDB will deploy a `MongoDB` database according to the above specification. It will also create the necessary `Secrets` and `Services` to access the database.

Let's check if the database is ready to use,

```bash
$ kubectl get mg -n demo sample-mongodb
NAME             VERSION   STATUS   AGE
sample-mongodb   4.4.26    Ready    3m53s
```

The database is `Ready`. Verify that KubeDB has created a `Secret` and a `Service` for this database using the following commands,

```bash
$ kubectl get secret -n demo 
NAME                          TYPE                       DATA   AGE
sample-mongodb-auth           kubernetes.io/basic-auth   2      5m20s

$ kubectl get service -n demo -l=app.kubernetes.io/instance=sample-mongodb
NAME                  TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)     AGE
sample-mongodb        ClusterIP   10.128.34.128   <none>        27017/TCP   4m47s
sample-mongodb-pods   ClusterIP   None            <none>        27017/TCP   4m47s
```

Here, we have to use service `sample-mongodb` and secret `sample-mongodb-auth` to connect with the database. `KubeDB` creates an [AppBinding](/docs/guides/mongodb/concepts/appbinding.md) CR that holds the necessary information to connect with the database.


**Verify AppBinding:**

Verify that the `AppBinding` has been created successfully using the following command,

```bash
$ kubectl get appbindings -n demo
NAME             TYPE                 VERSION   AGE
sample-mongodb   mongodb              4.4.26    24h
```

Let's check the YAML of the above `AppBinding`,

```bash
$ kubectl get appbindings -n demo sample-mongodb -o yaml
```

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1alpha2","kind":"MongoDB","metadata":{"annotations":{},"name":"sample-mongodb","namespace":"demo"},"spec":{"replicaSet":{"name":"replicaset"},"replicas":3,"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}}},"storageType":"Durable","terminationPolicy":"WipeOut","version":"4.4.26"}}
  creationTimestamp: "2024-09-17T06:02:46Z"
  generation: 1
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: sample-mongodb
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: mongodbs.kubedb.com
  name: sample-mongodb
  namespace: demo
  ownerReferences:
    - apiVersion: kubedb.com/v1
      blockOwnerDeletion: true
      controller: true
      kind: MongoDB
      name: sample-mongodb
      uid: 8c509564-fe74-4cbf-82a6-799ce0cc7bbd
  resourceVersion: "2973448"
  uid: c90c9fd6-ecde-430f-b97a-fb5b29a8839e
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
    replicaSets:
      host-0: replicaset/sample-mongodb-0.sample-mongodb-pods.demo.svc:27017,sample-mongodb-1.sample-mongodb-pods.demo.svc:27017,sample-mongodb-2.sample-mongodb-pods.demo.svc:27017
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

KubeStash uses the `AppBinding` CR to connect with the target database. It requires the following two fields to set in AppBinding's `.spec` section.

Here,

- `.spec.clientConfig.service.name` specifies the name of the Service that connects to the database.
- `.spec.secret` specifies the name of the Secret that holds necessary credentials to access the database.
- `.spec.type` specifies the types of the app that this AppBinding is pointing to. KubeDB generated AppBinding follows the following format: `<app group>/<app resource type>`.

**Insert Sample Data:**

Now, we are going to exec into one of the database pod and create some sample data. At first, find out the database `Pod` using the following command,

```bash
$ kubectl get pods -n demo --selector="app.kubernetes.io/instance=sample-mongodb" 
NAME                READY   STATUS    RESTARTS   AGE
sample-mongodb-0   2/2     Running   0          16m
sample-mongodb-1   2/2     Running   0          13m
sample-mongodb-2   2/2     Running   0          13m
```

Now, let’s exec into the pod and create a table,

```bash
$ export USER=$(kubectl get secrets -n demo sample-mongodb-auth -o jsonpath='{.data.\username}' | base64 -d)

$ export PASSWORD=$(kubectl get secrets -n demo sample-mongodb-auth -o jsonpath='{.data.\password}' | base64 -d)

$ kubectl exec -it -n demo sample-mongodb-0 -- mongo admin -u $USER -p $PASSWORD

replicaset:PRIMARY> show dbs
admin          0.000GB
config         0.000GB
kubedb-system  0.000GB
local          0.000GB

replicaset:PRIMARY> use newdb
switched to db newdb

replicaset:PRIMARY> db.movie.insert({"name":"batman"});
WriteResult({ "nInserted" : 1 })

replicaset:PRIMARY> db.movie.find().pretty()
{ "_id" : ObjectId("66e91dcdbade8984b312e0b0"), "name" : "batman" }

rs0:PRIMARY> exit
bye
```

Now, we are ready to backup the database.

### Prepare Backend

We are going to store our backed up data into a `S3` bucket. At first, we need to create a secret with S3 credentials then we need to create a `BackupStorage` crd. If you want to use a different backend, please read the respective backend configuration doc from [here](https://kubestash.com/docs/latest/guides/backends/overview/).

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
  name: s3-storage
  namespace: demo
spec:
  storage:
    provider: s3
    s3:
      endpoint: us-east-1.linodeobjects.com
      bucket: kubestash-testing
      region: us-east-1
      prefix: demo-application-level
      secretName: s3-secret
  usagePolicy:
    allowedNamespaces:
      from: All
  deletionPolicy: WipeOut
```

Let's create the BackupStorage we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mongodb/backup/kubestash/application-level/examples/backupstorage.yaml
backupstorage.storage.kubestash.com/s3-storage created
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

```yaml
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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mongodb/backup/kubestash/application-level/examples/retentionpolicy.yaml
retentionpolicy.storage.kubestash.com/backup-rp created
```

**Create BackupConfiguration:**

Below is the YAML for `BackupConfiguration` CR to take application-level backup of the `sample-mongodb` database that we have deployed earlier,

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: sample-mongodb-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: MongoDB
    namespace: demo
    name: sample-mongodb
  backends:
    - name: s3-backend
      storageRef:
        namespace: demo
        name: s3-storage
      retentionPolicy:
        name: backup-rp
        namespace: demo
  sessions:
    - name: frequent-backup
      scheduler:
        schedule: "*/5 * * * *"
        jobTemplate:
          backoffLimit: 1
      repositories:
        - name: s3-mongodb-repo
          backend: s3-backend
          directory: /mongodb
          encryptionSecret:
            name: encry-secret
            namespace: demo
      addon:
        name: mongodb-addon
        tasks:
          - name: manifest-backup
          - name: logical-backup
```

- `.spec.sessions[*].schedule` specifies that we want to backup at `5 minutes` interval.
- `.spec.target` refers to the targeted `sample-mongodb` MongoDB database that we created earlier.
- `.spec.sessions[*].addon.tasks[*].name[*]` specifies that both the `manifest-backup` and `logical-backup` tasks will be executed.

Let's create the `BackupConfiguration` CR that we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mongodb/kubestash/application-level/examples/backupconfiguration.yaml
backupconfiguration.core.kubestash.com/sample-mongodb-backup created
```

**Verify Backup Setup Successful**

If everything goes well, the phase of the `BackupConfiguration` should be `Ready`. The `Ready` phase indicates that the backup setup is successful. Let's verify the `Phase` of the BackupConfiguration,

```bash
$ kubectl get backupconfiguration -n demo
NAME                     PHASE   PAUSED   AGE
sample-mongodb-backup    Ready            2m50s
```

Additionally, we can verify that the `Repository` specified in the `BackupConfiguration` has been created using the following command,

```bash
$ kubectl get repo -n demo
NAME                  INTEGRITY   SNAPSHOT-COUNT   SIZE     PHASE   LAST-SUCCESSFUL-BACKUP   AGE
s3-mongodb-repo       0                            0 B      Ready                            3m
```

KubeStash keeps the backup for `Repository` YAMLs. If we navigate to the S3 bucket, we will see the `Repository` YAML stored in the `demo-application-level/mongodb` directory.

**Verify CronJob:**

It will also create a `CronJob` with the schedule specified in `spec.sessions[*].scheduler.schedule` field of `BackupConfiguration` CR.

Verify that the `CronJob` has been created using the following command,

```bash
$ kubectl get cronjob -n demo
NAME                                             SCHEDULE      SUSPEND   ACTIVE   LAST SCHEDULE   AGE
trigger-sample-mongodb-backup-frequent-backup    */5 * * * *             0        2m45s           3m25s
```

**Verify BackupSession:**

KubeStash triggers an instant backup as soon as the `BackupConfiguration` is ready. After that, backups are scheduled according to the specified schedule.

```bash
$ kubectl get backupsession -n demo -w
NAME                                                INVOKER-TYPE          INVOKER-NAME              PHASE       DURATION   AGE
sample-mongodb-backup-frequent-backup-1725449400    BackupConfiguration   sample-mongodb-backup     Succeeded              7m22s
```

We can see from the above output that the backup session has succeeded. Now, we are going to verify whether the backed up data has been stored in the backend.

**Verify Backup:**

Once a backup is complete, KubeStash will update the respective `Repository` CR to reflect the backup. Check that the repository `s3-mongodb-repo` has been updated by the following command,

```bash
$ kubectl get repository -n demo s3-mongodb-repo
NAME                       INTEGRITY   SNAPSHOT-COUNT   SIZE    PHASE   LAST-SUCCESSFUL-BACKUP   AGE
s3-mongodb-repo            true        1                806 B   Ready   8m27s                    9m18s
```

At this moment we have one `Snapshot`. Run the following command to check the respective `Snapshot` which represents the state of a backup run for an application.

```bash
$ kubectl get snapshots -n demo -l=kubestash.com/repo-name=s3-mongodb-repo
NAME                                                                  REPOSITORY          SESSION           SNAPSHOT-TIME          DELETION-POLICY   PHASE       AGE
s3-mongodb-repo-sample-mongodb-backup-frequent-backup-1725449400      s3-mongodb-repo     frequent-backup   2024-09-17T06:53:42Z   Delete            Succeeded   16h
```

> Note: KubeStash creates a `Snapshot` with the following labels:
> - `kubestash.com/app-ref-kind: <target-kind>`
> - `kubestash.com/app-ref-name: <target-name>`
> - `kubestash.com/app-ref-namespace: <target-namespace>`
> - `kubestash.com/repo-name: <repository-name>`
>
> These labels can be used to watch only the `Snapshot`s related to our target Database or `Repository`.

If we check the YAML of the `Snapshot`, we can find the information about the backed up components of the Database.

```bash
$ kubectl get snapshots -n demo s3-mongodb-repo-sample-mongodb-backup-frequent-backup-1725449400 -oyaml
```

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: Snapshot
metadata:
  creationTimestamp: "2024-09-17T06:53:42Z"
  finalizers:
    - kubestash.com/cleanup
  generation: 1
  labels:
    kubestash.com/app-ref-kind: MongoDB
    kubestash.com/app-ref-name: sample-mongodb
    kubestash.com/app-ref-namespace: demo
    kubestash.com/repo-name: test2
  annotations:
    kubedb.com/db-version: "4.4.26"
  name: s3-mongodb-repo-sample-mongodb-backup-frequent-backup-1725449400
  namespace: demo
  ownerReferences:
    - apiVersion: storage.kubestash.com/v1alpha1
      blockOwnerDeletion: true
      controller: true
      kind: Repository
      name: s3-mongodb-repo
      uid: 39f5baef-d374-4931-ae54-2b9923bd0a4b
  resourceVersion: "2982041"
  uid: e139815d-0f34-42ff-8b0c-a5b0945d6e74
spec:
  appRef:
    apiGroup: kubedb.com
    kind: MongoDB
    name: sample-mongodb
    namespace: demo
  backupSession: sample-mongodb-backup-frequent-backup-1725449400
  deletionPolicy: Delete
  repository: s3-mongodb-repo
  session: frequent-backup
  snapshotID: 01J7ZC4A7GM2K8GEEPQDPQ49T8
  type: FullBackup
  version: v1
status:
  components:
    dump:
      driver: Restic
      duration: 1.434544863s
      integrity: true
      path: repository/v1/frequent-backup/dump
      phase: Succeeded
      resticStats:
        - hostPath: dump
          id: 67a2f5bd65ef78cd5cece9906dadd0d62523decae71db05d9f895140aabe9ec0
          size: 3.340 KiB
          uploaded: 3.624 KiB
      size: 1.524 KiB
    manifest:
      driver: Restic
      duration: 1.242758731s
      integrity: true
      path: repository/v1/frequent-backup/manifest
      phase: Succeeded
      resticStats:
        - hostPath: /kubestash-tmp/manifest
          id: b5660a5a38532f6817769ff693c0c317730148f02b24e7acfc6ac7d8464a9518
          size: 3.866 KiB
          uploaded: 5.299 KiB
      size: 2.345 KiB
  conditions:
    - lastTransitionTime: "2024-09-17T06:53:43Z"
      message: Recent snapshot list updated successfully
      reason: SuccessfullyUpdatedRecentSnapshotList
      status: "True"
      type: RecentSnapshotListUpdated
    - lastTransitionTime: "2024-09-17T06:54:02Z"
      message: Metadata uploaded to backend successfully
      reason: SuccessfullyUploadedSnapshotMetadata
      status: "True"
      type: SnapshotMetadataUploaded
  integrity: true
  phase: Succeeded
  size: 3.868 KiB
  snapshotTime: "2024-09-17T06:53:42Z"
  totalComponents: 2
```

> KubeStash uses `mongodump` to perform backups of target `MongoDB` databases. Therefore, the component name for logical backups is set as `dump`.

> KubeStash set component name as `manifest` for the `manifest backup` of MongoDB databases.

Now, if we navigate to the S3 bucket, we will see the backed up data stored in the `demo-application-level/mongodb/repository/v1/frequent-backup/dump` directory. KubeStash also keeps the backup for `Snapshot` YAMLs, which can be found in the `demo-application-level/mongodb/snapshots` directory.

> Note: KubeStash stores all dumped data encrypted in the backup directory, meaning it remains unreadable until decrypted.

## Restore

In this section, we are going to restore the entire database from the backup that we have taken in the previous section.

For this tutorial, we will restore the database in a separate namespace called `dev`.

First, create the namespace by running the following command:

```bash
$ kubectl create ns dev
namespace/dev created
```

#### Create RestoreSession:

We need to create a RestoreSession CR.

Below, is the contents of YAML file of the `RestoreSession` CR that we are going to create to restore the entire database.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: restore-sample-mongodb
  namespace: demo
spec:
  manifestOptions:
    mongoDB:
      db: true
      restoreNamespace: dev
  dataSource:
    repository: s3-mongodb-repo
    snapshot: latest
    encryptionSecret:
      name: encry-secret
      namespace: demo
  addon:
    name: mongodb-addon
    tasks:
      - name: logical-backup-restore
      - name: manifest-restore
```

Here,

- `.spec.manifestOptions.mongodb.db` specifies whether to restore the DB manifest or not.
- `.spec.dataSource.repository` specifies the Repository object that holds the backed up data.
- `.spec.dataSource.snapshot` specifies to restore from latest `Snapshot`.
- `.spec.addon.tasks[*]` specifies that both the `manifest-restore` and `logical-backup-restore` tasks.

Let's create the RestoreSession CR object we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mongodb/backup/kubestash/application-level/examples/restoresession.yaml
restoresession.core.kubestash.com/restore-sample-mongodb created
```

Once, you have created the `RestoreSession` object, KubeStash will create restore Job. Run the following command to watch the phase of the `RestoreSession` object,

```bash
$ watch kubectl get restoresession -n demo
Every 2.0s: kubectl get restores... AppsCode-PC-03: Wed Aug 21 10:44:05 2024
NAME                      REPOSITORY            FAILURE-POLICY   PHASE       DURATION   AGE
restore-sample-mongodb    s3-mongodb-repo                        Succeeded   3s         53s
```

The `Succeeded` phase means that the restore process has been completed successfully.


#### Verify Restored MongoDB Manifest:

In this section, we will verify whether the desired `MongoDB` database manifest has been successfully applied to the cluster.

```bash
$ kubectl get mongodb -n dev 
NAME              VERSION   STATUS   AGE
sample-mongodb    4.4.26    Ready    9m46s
```

The output confirms that the `MongoDB` database has been successfully created with the same configuration as it had at the time of backup.


#### Verify Restored Data:

In this section, we are going to verify whether the desired data has been restored successfully. We are going to connect to the database server and check whether the database and the table we created earlier in the original database are restored.

At first, check if the database has gone into **`Ready`** state by the following command,

```bash
$ kubectl get mongodb -n dev sample-mongodb
NAME              VERSION   STATUS   AGE
sample-mongodb    4.4.26    Ready    9m46s
```

Now, find out the database `Pod` by the following command,

```bash
$ kubectl get pods -n dev --selector="app.kubernetes.io/instance=sample-mongodb"
NAME                READY   STATUS    RESTARTS   AGE
sample-mongodb-0    2/2     Running   0          12m
sample-mongodb-1    2/2     Running   0          12m
sample-mongodb-2    2/2     Running   0          12m
```


Now, lets exec one of the Pod and verify restored data.

```bash
$ export USER=$(kubectl get secrets -n dev sample-mongodb-auth -o jsonpath='{.data.\username}' | base64 -d)

$ export PASSWORD=$(kubectl get secrets -n dev sample-mongodb-auth -o jsonpath='{.data.\password}' | base64 -d)

$ kubectl exec -it -n demo sample-mongodb-0 -- mongo admin -u $USER -p $PASSWORD

---
replicaset:PRIMARY> show dbs
admin          0.000GB
config         0.000GB
kubedb-system  0.000GB
local          0.000GB
newdb          0.000GB

replicaset:PRIMARY> use newdb
switched to db newdb

replicaset:PRIMARY> show collections
movie

replicaset:PRIMARY> db.movie.find()
{ "_id" : ObjectId("66e91dcdbade8984b312e0b0"), "name" : "batman" }

replicaset:PRIMARY> exit
bye

```

So, from the above output, we can see that in `dev` namespace the original database `sample-mongodb` has been restored successfully.

## Cleanup

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete backupconfigurations.core.kubestash.com  -n demo sample-mongodb-backup
kubectl delete retentionpolicies.storage.kubestash.com -n demo backup-rp
kubectl delete restoresessions.core.kubestash.com -n demo restore-sample-mongodb
kubectl delete backupstorage -n demo s3-storage
kubectl delete secret -n demo s3-secret
kubectl delete secret -n demo encry-secret
kubectl delete mongodb -n demo sample-mongodb
kubectl delete mongodb -n dev sample-mongodb
```