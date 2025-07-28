---
title: Backup & Restore Cassandra | KubeStash
description: Backup Cassandra database using KubeStash
menu:
  docs_{{ .version }}:
    identifier: guides-cas-backup-logical-backup-stashv2
    name: Logical Backup
    parent: guides-cas-backup-stashv2
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Backup and Restore Cassandra database using KubeStash

KubeStash allows you to backup and restore `Cassandra` databases. It supports backups for `Cassandra` instances running in Standalone, and cluster configurations. KubeStash makes managing your `Cassandra` backups and restorations more straightforward and efficient.

This guide will give you how you can take backup and restore your `Cassandra` databases using `Kubestash`.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using `Minikube` or `Kind`.
- Install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md).
- Install `KubeStash` in your cluster following the steps [here](https://kubestash.com/docs/latest/setup/install/kubestash).
- Install KubeStash `kubectl` plugin following the steps [here](https://kubestash.com/docs/latest/setup/install/kubectl-plugin/).
- If you are not familiar with how KubeStash backup and restore Cassandra databases, please check the following guide [here](/docs/guides/cassandra/backup/kubestash/overview/index.md).

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

> **Note:** YAML files used in this tutorial are stored in [docs/guides/cassandra/backup/kubestash/logical/examples](/docs/guides/cassandra/backup/kubestash/logical/examples) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Backup Cassandra

KubeStash supports backups for `Cassandra` instances across different configurations, including Standalone, and Cluster setups. In this demonstration, we'll focus on a `Cassandra` database using Clustering mode. The backup and restore process is similar for Standalone and Cluster configurations as well.

This section will demonstrate how to backup a `Cassandra` database. Here, we are going to deploy a `Cassandra` database using KubeDB. Then, we are going to backup this database into a `S3` bucket. Finally, we are going to restore the backup up data into another `Cassandra` database.

### Create Cassandra License Secret

We need Cassandra License to create Cassandra Database. So, Ensure that you have acquired a license and then simply pass the license by secret.


### Deploy Sample Cassandra Database

Let's deploy a sample `Cassandra` database and insert some data into it.

**Create Cassandra CR:**

Below is the YAML of a sample `Cassandra` CRD that we are going to create for this tutorial:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Cassandra
metadata:
  name: cas-sample
  namespace: demo
spec:
  version: "8.7.10"
  topology:
    aggregator:
      replicas: 2
      podTemplate:
        spec:
          containers:
          - name: cassandra
            resources:
              limits:
                memory: "2Gi"
                cpu: "600m"
              requests:
                memory: "2Gi"
                cpu: "600m"
      storage:
        accessModes:
        - ReadWriteOnce
        resources:
          requests:
            storage: 1Gi
    leaf:
      replicas: 3
      podTemplate:
        spec:
          containers:
            - name: cassandra
              resources:
                limits:
                  memory: "2Gi"
                  cpu: "600m"
                requests:
                  memory: "2Gi"
                  cpu: "600m"                      
      storage:
        accessModes:
          - ReadWriteOnce
        resources:
          requests:
            storage: 10Gi
  licenseSecret:
    name: license-secret
  storageType: Durable
  deletionPolicy: WipeOut
```

Here,

- `spec.version` is the name of the CassandraVersion CRD where the docker images are specified. In this tutorial, a Cassandra `8.7.10` database is going to be created.
- `spec.topology` specifies that it will be used as cluster mode. If this field is nil it will be work as standalone mode.
- `spec.topology.aggregator.replicas` or `spec.topology.leaf.replicas` specifies that the number replicas that will be used for aggregator or leaf.
- `spec.storageType` specifies the type of storage that will be used for Cassandra database. It can be `Durable` or `Ephemeral`. Default value of this field is `Durable`. If `Ephemeral` is used then KubeDB will create Cassandra database using `EmptyDir` volume. In this case, you don't have to specify `spec.storage` field. This is useful for testing purposes.
- `spec.topology.aggregator.storage` or `spec.topology.leaf.storage` specifies the StorageClass of PVC dynamically allocated to store data for this database. This storage spec will be passed to the PetSet created by KubeDB operator to run database pods. You can specify any StorageClass available in your cluster with appropriate resource requests.
- `spec.deletionPolicy` gives flexibility whether to `nullify`(reject) the delete operation of `Cassandra` crd or which resources KubeDB should keep or delete when you delete `Cassandra` crd. If admission webhook is enabled, It prevents users from deleting the database as long as the `spec.deletionPolicy` is set to `DoNotTerminate`. Learn details of all `DeletionPolicy` [here](/docs/guides/mysql/concepts/database/index.md#specdeletionpolicy)

> Note: `spec.storage` section is used to create PVC for database pod. It will create PVC with storage size specified in `storage.resources.requests` field. Don't specify limits here. PVC does not get resized automatically.

Create the above `Cassandra` CR,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/cassandra/backup/kubestash/logical/examples/cas-sample.yaml
cassandra.kubedb.com/cas-sample created
```

KubeDB will deploy a Cassandra database according to the above specification. It will also create the necessary `Secrets` and `Services` to access the database.

Let's check if the database is ready to use,

```bash
$ kubectl get cassandras.kubedb.com -n demo 
NAME         TYPE                  VERSION   STATUS   AGE
cas-sample   kubedb.com/v1alpha2   5.0.3     Ready    3m6s
```

The database is `Ready`. Verify that KubeDB has created a `Secret` and a `Service` for this database using the following commands,

```bash
$  kubectl get secret -n demo -l=app.kubernetes.io/instance=cas-sample
NAME                TYPE                       DATA   AGE
cas-sample-auth     kubernetes.io/basic-auth   2      3m33s
cas-sample-config   Opaque                     1      3m33s

$  kubectl get service -n demo -l=app.kubernetes.io/instance=cas-sample
NAME                      TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)                               AGE
cas-sample                ClusterIP   10.96.77.149   <none>        9042/TCP,7000/TCP,7199/TCP,7001/TCP   3m57s
cas-sample-rack-r0-pods   ClusterIP   None           <none>        9042/TCP,7000/TCP,7199/TCP,7001/TCP   3m57s
```

Here, we have to use service `cas-sample` and secret `cas-sample-auth` to connect with the database. `KubeDB` creates an [AppBinding](/docs/guides/mysql/concepts/appbinding/index.md) CR that holds the necessary information to connect with the database.

**Verify AppBinding:**

Verify that the `AppBinding` has been created successfully using the following command,

```bash
$  kubectl get appbindings -n demo
NAME         TYPE                   VERSION   AGE
cas-sample   kubedb.com/cassandra   5.0.3     4m23s
```

Let's check the YAML of the above `AppBinding`,

```bash
$ kubectl get appbindings -n demo cas-sample -o yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1alpha2","kind":"Cassandra","metadata":{"annotations":{},"name":"cas-sample","namespace":"demo"},"spec":{"configuration":null,"deletionPolicy":"WipeOut","topology":{"rack":[{"name":"r0","podTemplate":{"spec":{"containers":[{"name":"cassandra","resources":{"limits":{"cpu":2,"memory":"2Gi"},"requests":{"cpu":1,"memory":"1Gi"}}}]}},"replicas":2,"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}}},"storageType":"Durable"}]},"version":"5.0.3"}}
  creationTimestamp: "2025-07-28T05:04:35Z"
  generation: 1
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: cas-sample
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: cassandras.kubedb.com
  name: cas-sample
  namespace: demo
  ownerReferences:
  - apiVersion: kubedb.com/v1alpha2
    blockOwnerDeletion: true
    controller: true
    kind: Cassandra
    name: cas-sample
    uid: de9c3313-c9f2-4235-8f84-3d9a92d22503
  resourceVersion: "1844"
  uid: f04d76e2-1f90-4475-8ee5-e6fdfe80079e
spec:
  appRef:
    apiGroup: kubedb.com
    kind: Cassandra
    name: cas-sample
    namespace: demo
  clientConfig:
    service:
      name: cas-sample
      port: 9042
      scheme: http
  secret:
    name: cas-sample-auth
  type: kubedb.com/cassandra
  version: 5.0.3
```

KubeStash uses the `AppBinding` CR to connect with the target database. It requires the following two fields to set in AppBinding's `.spec` section.

- `.spec.parameters.masterAggregator` specifies the dns of master aggregator node that we have to mention in mysqldump command when taken backup or restore.
- `.spec.clientConfig.service.name` specifies the name of the Service that connects to the database.
- `.spec.secret` specifies the name of the Secret that holds necessary credentials to access the database.
- `spec.type` specifies the types of the app that this AppBinding is pointing to. KubeDB generated AppBinding follows the following format: `<app group>/<app resource type>`.

**Insert Sample Data:**

Now, we are going to exec into the any  pod and create some sample data. At first, find out the database `Pod` using the following command,

```bash
$  kubectl get pods -n demo --selector="app.kubernetes.io/instance=cas-sample"
NAME                   READY   STATUS    RESTARTS   AGE
cas-sample-rack-r0-0   1/1     Running   0          5m28s
cas-sample-rack-r0-1   1/1     Running   0          4m28s
```

And copy the username and password of the database to access into `cqlsh` shell.

```bash
$  kubectl get secret -n demo  cas-sample-auth -o jsonpath='{.data.username}'| base64 -d
admin⏎                                         
 kubectl get secret -n demo  cas-sample-auth -o jsonpath='{.data.password}'| base64 -d
gkebeP3HJbxubvCM⏎                     
```

Now, Lets exec into the any aggregator `Pod` to enter into `cqlsh` shell and create a database and a table,

```bash
$ kubectl exec -it -n demo cas-sample-rack-r0-0 -- cqlsh -u admin -p gkebeP3HJbxubvCM
Defaulted container "cassandra" out of: cassandra, cassandra-init (init), medusa-init (init)

Warning: Using a password on the command line interface can be insecure.
Recommendation: use the credentials file to securely provide the password.

Connected to Test Cluster at 127.0.0.1:9042
[cqlsh 6.2.0 | Cassandra 5.0.3 | CQL spec 3.4.7 | Native protocol v5]
Use HELP for help.
admin@cqlsh> 


admin@cqlsh> CREATE KEYSPACE kubedb  WITH REPLICATION = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };
admin@cqlsh> USE kubedb;
admin@cqlsh:kubedb> CREATE TABLE users (
          ... id UUID PRIMARY KEY,
          ... name TEXT,
          ... email TEXT
          ... );
admin@cqlsh:kubedb> INSERT INTO kubedb.users (id, name, email) VALUES (uuid(), 'demo_name1', 'kubedb@demo1.com');
admin@cqlsh:kubedb> INSERT INTO kubedb.users (id, name, email) VALUES (uuid(), 'demo_name2', 'kubedb@demo2.com');
admin@cqlsh:kubedb> SELECT * FROM kubedb.users;

 id                                   | email            | name
--------------------------------------+------------------+------------
 e778de6b-5a71-447b-b015-4c9e0b62bfd6 | kubedb@demo1.com | demo_name1
 17dd25bd-749f-476b-a29e-f9ae97820224 | kubedb@demo2.com | demo_name2

(2 rows)
admin@cqlsh:kubedb> exit
⏎  
```

Now, we are ready to backup the database.

### Prepare Backend

We are going to store our backed up data into a S3 bucket. We have to create a Secret with necessary credentials and a `BackupStorage` CR to use this backend. If you want to use a different backend, please read the respective backend configuration doc from [here](https://kubestash.com/docs/latest/guides/backends/overview/).

**Create Secret:**

Let's create a secret called `medusa-cred` with access credentials to our desired S3 bucket,

```bash
$  kubectl create secret generic -n demo medusa-cred \
     --from-file=./AWS_ACCESS_KEY_ID \
     --from-file=./AWS_SECRET_ACCESS_KEY
secret/medusa-cred created


```

**Create BackupStorage:**

Now, create a `BackupStorage` using this secret. Below is the YAML of `BackupStorage` CR we are going to create,

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
      bucket: anisur
      prefix: medusa-jul
      secretName: medusa-cred
      region: us-east-1
  usagePolicy:
    allowedNamespaces:
      from: All
  default: true
  deletionPolicy: Delete
```

Let's create the BackupStorage we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/cassandra/backup/kubestash/logical/examples/backupstorage.yaml
backupstorage.storage.kubestash.com/s3-storage created
```

Now, we are ready to backup our database to our desired backend.

**Create RetentionPolicy:**

Now, let's create a `RetentionPolicy` to specify how the old Snapshots should be cleaned up.

Below is the YAML of the `RetentionPolicy` object that we are going to create,

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: RetentionPolicy
metadata:
  name: demo-retention
  namespace: demo
spec:
  default: true
  failedSnapshots:
    last: 2
  maxRetentionPeriod: 2mo
  successfulSnapshots:
    last: 5
  usagePolicy:
    allowedNamespaces:
      from: All
```

Let’s create the above `RetentionPolicy`,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/cassandra/backup/kubestash/logical/examples/retentionpolicy.yaml
retentionpolicy.storage.kubestash.com/demo-retention created
```

### Backup

We have to create a `BackupConfiguration` targeting respective `cas-sample` Cassandra database. Then, KubeStash will create a `CronJob` for each session to take periodic backup of that database.


**Create BackupConfiguration:**

Below is the YAML for `BackupConfiguration` CR to backup the `cas-sample` database that we have deployed earlier,

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: sample-cas-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Cassandra
    namespace: demo
    name: cas-sample
  backends:
    - name: s3-backend
      storageRef:
        namespace: demo
        name: s3-storage
      retentionPolicy:
        name: demo-retention
        namespace: demo
  sessions:
    - name: frequent-backup
      scheduler:
        schedule: "*/5 * * * *"
        jobTemplate:
          backoffLimit: 1
      repositories:
        - name: s3-cassandra-repo
          backend: s3-backend
          directory: /cas
      addon:
        name: cassandra-addon
        tasks:
          - name: logical-backup
```

- `.spec.sessions[*].schedule` specifies that we want to backup the database at `5 minutes` interval.
- `.spec.target` refers to the targeted `cas-sample` SigleStore database that we created earlier.

Let's create the `BackupConfiguration` CR that we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/cassandra/backup/kubestash/logical/examples/backupconfiguration.yaml
backupconfiguration.core.kubestash.com/sample-cas-backup created
```

**Verify Backup Setup Successful**

If everything goes well, the phase of the `BackupConfiguration` should be `Ready`. The `Ready` phase indicates that the backup setup is successful. Let's verify the `Phase` of the BackupConfiguration,

```bash
$ kubectl get backupconfiguration -n demo
NAME                PHASE   PAUSED   AGE
sample-cas-backup   Ready            107s
```

Additionally, we can verify that the `Repository` specified in the `BackupConfiguration` has been created using the following command,

```bash
$ kubectl get repo -n demo
NAME                INTEGRITY   SNAPSHOT-COUNT   SIZE   PHASE   LAST-SUCCESSFUL-BACKUP   AGE
s3-cassandra-repo               1                0 B    Ready   2m15s                    2m48s
```

KubeStash keeps the backup for `Repository` YAMLs. If we navigate to the S3 bucket, we will see the `Repository` YAML stored in the `demo/cassandra` directory.

**Verify CronJob:**

It will also create a `CronJob` with the schedule specified in `spec.sessions[*].scheduler.schedule` field of `BackupConfiguration` CR.

Verify that the `CronJob` has been created using the following command,

```bash
$  kubectl get cronjob -n demo
NAME                                        SCHEDULE      TIMEZONE   SUSPEND   ACTIVE   LAST SCHEDULE   AGE
trigger-sample-cas-backup-frequent-backup   */5 * * * *   <none>     False     0        47s             2m39s
```

**Verify BackupSession:**

KubeStash triggers an instant backup as soon as the `BackupConfiguration` is ready. After that, backups are scheduled according to the specified schedule.

```bash
$ kubectl get backupsession -n demo -w
NAME                                           INVOKER-TYPE          INVOKER-NAME        PHASE       DURATION   AGE
sample-cas-backup-frequent-backup-1753682588   BackupConfiguration   sample-cas-backup   Succeeded   2m2s       2m59s
```

We can see from the above output that the backup session has succeeded. Now, we are going to verify whether the backed up data has been stored in the backend.

**Verify Backup:**

Once a backup is complete, KubeStash will update the respective `Repository` CR to reflect the backup. Check that the repository `sample-cas-backup` has been updated by the following command,

```bash
$  kubectl get repository -n demo s3-cassandra-repo
NAME                INTEGRITY   SNAPSHOT-COUNT   SIZE   PHASE   LAST-SUCCESSFUL-BACKUP   AGE
s3-cassandra-repo               1                0 B    Ready   3m46s                    4m19s
```

At this moment we have one `Snapshot`. Run the following command to check the respective `Snapshot` which represents the state of a backup run for an application.

```bash
$ kubectl get snapshots -n demo -l=kubestash.com/repo-name=s3-cassandra-repo
NAME                                                             REPOSITORY          SESSION           SNAPSHOT-TIME          DELETION-POLICY   PHASE       AGE
s3-cassandra-repo-sample-cas-backup-frequent-backup-1753682588   s3-cassandra-repo   frequent-backup   2025-07-28T06:03:08Z   Delete            Succeeded   4m12s
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
$ kubectl get snapshots -n demo s3-cassandra-repo-sample-cas-backup-frequent-backup-1753682588 -oyaml
```

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: Snapshot
metadata:
  annotations:
    kubedb.com/db-version: 5.0.3
  creationTimestamp: "2025-07-28T06:03:08Z"
  finalizers:
    - kubestash.com/cleanup
  generation: 1
  labels:
    kubestash.com/app-ref-kind: Cassandra
    kubestash.com/app-ref-name: cas-sample
    kubestash.com/app-ref-namespace: demo
    kubestash.com/repo-name: s3-cassandra-repo
  name: s3-cassandra-repo-sample-cas-backup-frequent-backup-1753682588
  namespace: demo
  ownerReferences:
    - apiVersion: storage.kubestash.com/v1alpha1
      blockOwnerDeletion: true
      controller: true
      kind: Repository
      name: s3-cassandra-repo
      uid: 2e408d1b-081a-4f8b-8c94-c7856f267411
  resourceVersion: "8506"
  uid: 222c3109-62bf-42a5-a8e3-ecab3180f4c7
spec:
  appRef:
    apiGroup: kubedb.com
    kind: Cassandra
    name: cas-sample
    namespace: demo
  backupSession: sample-cas-backup-frequent-backup-1753682588
  deletionPolicy: Delete
  repository: s3-cassandra-repo
  session: frequent-backup
  snapshotID: 01K17T1CVZMMQBRKKJWRPPBAPS
  type: FullBackup
  version: v1
status:
  components:
    dump:
      driver: Medusa
      duration: 0s
      medusaStats:
        backupName: s3-cassandra-repo-sample-cas-backup-frequent-backup-1753682588
      path: repository/v1/frequent-backup/dump
      phase: Succeeded
  conditions:
    - lastTransitionTime: "2025-07-28T06:03:08Z"
      message: Recent snapshot list updated successfully
      reason: SuccessfullyUpdatedRecentSnapshotList
      status: "True"
      type: RecentSnapshotListUpdated
    - lastTransitionTime: "2025-07-28T06:05:07Z"
      message: Metadata uploaded to backend successfully
      reason: SuccessfullyUploadedSnapshotMetadata
      status: "True"
      type: SnapshotMetadataUploaded
  phase: Succeeded
  snapshotTime: "2025-07-28T06:03:08Z"
  totalComponents: 1
  verificationStatus: NotVerified
```


Now, if we navigate to the S3 bucket, we will see the backed up data stored in the `demo/cassandra/repository/v1/frequent-backup/dump` directory. KubeStash also keeps the backup for `Snapshot` YAMLs, which can be found in the `demo/dep/snapshots` directory.

> Note: KubeStash stores all dumped data encrypted in the backup directory, meaning it remains unreadable until decrypted.

## Restore

In this section, we are going to restore the database from the backup we have taken in the previous section. We are going to deploy a new database and initialize it from the backup.

#### Deploy Restored Database:

Now, we have to deploy the restored database similarly as we have deployed the original `cas-sample` database. 
#### Create RestoreSession:

Now, we need to create a RestoreSession CRD pointing to targeted `Cassandra` database.

Below, is the contents of YAML file of the `RestoreSession` object that we are going to create to restore backed up data into the newly created database provisioned by Cassandra object named `restored-cassandra`.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: restore-sample-cassandra
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Cassandra
    namespace: demo
    name: cas-sample
  dataSource:
    repository: s3-cassandra-repo
    snapshot: latest
  addon:
    name: cassandra-addon
    tasks:
      - name: logical-backup-restore
```

Here,

- `.spec.target` refers to the newly created `restored-cassandra` Cassandra object to where we want to restore backup data.
- `.spec.dataSource.repository` specifies the Repository object that holds the backed up data.
- `.spec.dataSource.snapshot` specifies to restore from latest `Snapshot`.

Let's create the RestoreSession CRD object we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/cassandra/backup/kubestash/logical/examples/restoresession.yaml
restoresession.core.kubestash.com/sample-cassandra-restore created
```

Once, you have created the `RestoreSession` object, KubeStash will create restore Job. Run the following command to watch the phase of the `RestoreSession` object,

```bash
$  kubectl get restoresession -n demo
NAME                       REPOSITORY          PHASE     DURATION   AGE
restore-sample-cassandra   s3-cassandra-repo   Running              100s
```

The `Succeeded` phase means that the restore process has been completed successfully.


#### Verify Restored Data:

In this section, we are going to verify whether the desired data has been restored successfully. We are going to connect to the database server and check whether the database and the table we created earlier in the original database are restored.

At first, check if the database has gone into `Ready` state by the following command,

```bash
$ kubectl get cassandra -n demo cas-sample
NAME         TYPE                  VERSION   STATUS   AGE
cas-sample   kubedb.com/v1alpha2   5.0.3     Ready    136m
```

Now, find out the database `Pod` by the following command,

```bash
$ kubectl get pods -n demo --selector="app.kubernetes.io/instance=cas-
sample"
NAME                             READY   STATUS    RESTARTS   AGE
cas-sample-rack-r0-0             1/1     Running   0          137m
cas-sample-rack-r0-1             1/1     Running   0          136m
```

And then copy the user name and password of the `root` user to access into `cqlsh` shell.

```bash
$  kubectl get secret -n demo  cas-sample-auth -o jsonpath='{.data.username}'| base64 -d
admin⏎                                         
 kubectl get secret -n demo  cas-sample-auth -o jsonpath='{.data.password}'| base64 -d
gkebeP3HJbxubvCM⏎    
```

Now, Lets exec into the any aggregator `Pod` to enter into `mysql` shell and create a database and a table,

```bash
$  kubectl exec -it -n demo cas-sample-rack-r0-0 -- cqlsh -u admin -p gkebeP3HJbxubvCM
Defaulted container "cassandra" out of: cassandra, cassandra-init (init), medusa-init (init)

Warning: Using a password on the command line interface can be insecure.
Recommendation: use the credentials file to securely provide the password.

Connected to Test Cluster at 127.0.0.1:9042
[cqlsh 6.2.0 | Cassandra 5.0.3 | CQL spec 3.4.7 | Native protocol v5]
Use HELP for help.
admin@cqlsh> SELECT * FROM kubedb.users;

 id                                   | email            | name
--------------------------------------+------------------+------------
 e778de6b-5a71-447b-b015-4c9e0b62bfd6 | kubedb@demo1.com | demo_name1
 17dd25bd-749f-476b-a29e-f9ae97820224 | kubedb@demo2.com | demo_name2

(2 rows)

```

So, from the above output, we can see that the `users` table we have created earlier in the original database and now, they are restored successfully.

## Cleanup

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete backupconfigurations.core.kubestash.com  -n demo sample-cas-backup
kubectl delete restoresessions.core.kubestash.com -n demo restore-sample-cassandra
kubectl delete retentionpolicies.storage.kubestash.com -n demo demo-retention
kubectl delete backupstorage -n demo s3-storage
kubectl delete secret -n demo medusa-cred
kubectl delete my -n demo cas-sample
```