---
title: Backup & Restore ZooKeeper | KubeStash
description: Backup ans Restore ZooKeeper using KubeStash
menu:
  docs_{{ .version }}:
    identifier: guides-zk-logical-backup-stashv2
    name: Logical Backup
    parent: guides-zk-backup-stashv2
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Backup and Restore ZooKeeper using KubeStash

KubeStash allows you to backup and restore `ZooKeeper`. KubeStash makes managing your `ZooKeeper` backups and restorations more straightforward and efficient.

This guide will give you an overview how you can take backup and restore your `ZooKeeper` using `Kubestash`.


## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using `Minikube` or `Kind`.
- Install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md).
- Install `KubeStash` in your cluster following the steps [here](https://kubestash.com/docs/latest/setup/install/kubestash).
- Install KubeStash `kubectl` plugin following the steps [here](https://kubestash.com/docs/latest/setup/install/kubectl-plugin/).
- If you are not familiar with how KubeStash backup and restore ZooKeeper, please check the following guide [here](/docs/guides/zookeeper/backup/kubestash/overview/index.md).

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

> **Note:** YAML files used in this tutorial are stored in [docs/guides/zookeeper/backup/kubestash/logical/examples](https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/zookeeper/backup/kubestash/logical/examples) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.


## Backup ZooKeeper

KubeStash supports backups for `ZooKeeper` instances across different configurations, including Standalone and ZooKeeper Ensemble setups. In this demonstration, we'll focus on a `ZooKeeper` using ZooKeeper Ensemble configuration. The backup and restore process is similar for Standalone configuration.

This section will demonstrate how to backup a `ZooKeeper`. Here, we are going to deploy a `ZooKeeper` using KubeDB. Then, we are going to backup this into a `s3` bucket. Finally, we are going to restore the backup up data into another `ZooKeeper`.


### Deploy Sample ZooKeeper 

Let's deploy a sample `ZooKeeper`  and insert some data into it.

**Create ZooKeeper CR:**

Below is the YAML of a sample `ZooKeeper` CR that we are going to create for this tutorial:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ZooKeeper
metadata:
  name: sample-zookeeper
  namespace: demo
spec:
  version: "3.8.3"
  adminServerPort: 8080
  replicas: 3
  storage:
    resources:
      requests:
        storage: "1Gi"
    accessModes:
      - ReadWriteOnce
  deletionPolicy: "WipeOut"
```

Create the above `ZooKeeper` CR,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/zookeeper/backup/kubestash/logical/examples/sample-zookeeper.yaml
zookeeper.kubedb.com/sample-zookeeper created
```

KubeDB will deploy a `ZooKeeper` according to the above specification. It will also create the necessary `Secrets` and `Services` to access.

Let's check if the zookeeper is ready to use,

```bash
$ kubectl get zk -n demo sample-zookeeper
NAME               VERSION   STATUS   AGE
sample-zookeeper   8.3.3     Ready    5m1s
```

The zookeeper is `Ready`. Verify that KubeDB has created a `Secret` and a `Service` for this zookeeper using the following commands,

```bash
$ kubectl get secret -n demo 
NAME                           TYPE                       DATA   AGE
sample-zookeeper-auth          kubernetes.io/basic-auth   2      5m20s

$ kubectl get service -n demo -l=app.kubernetes.io/instance=sample-zookeeper
NAME                           TYPE         CLUSTER-IP        EXTERNAL-IP     PORT(S)                      AGE
sample-zookeeper               ClusterIP    10.128.65.175     <none>          2181/TCP                     5m55s
sample-zookeeper-pods          ClusterIP    None              <none>          2181/TCP,2888/TCP,3888/TCP   5m55s
sample-zookeeper-admin-server  ClusterIP    10.128.163.169    <none>          8080/TCP                     5m55s
```

Here, we have to use service `sample-zookeeper` and secret `sample-zookeeper-auth` to connect with the zookeeper. `KubeDB` creates an [AppBinding](/docs/guides/zookeeper/concepts/appbinding.md) CR that holds the necessary information to connect with the zookeeper.


**Verify AppBinding:**

Verify that the `AppBinding` has been created successfully using the following command,

```bash
$ kubectl get appbindings -n demo
NAME                       TYPE                   VERSION    AGE
sample-zookeeper           kubedb.com/zookeeper   3.8.3      9m30s
```

Let's check the YAML of the above `AppBinding`,

```bash
$ kubectl get appbindings -n demo sample-zookeeper -o yaml
```

```yaml
apiVersion: appcatalog.appscode.com/v1alpha1
kind: AppBinding
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"kubedb.com/v1alpha2","kind":"ZooKeeper","metadata":{"annotations":{},"name":"sample-zookeeper","namespace":"demo"},"spec":{"adminServerPort":8080,"deletionPolicy":"WipeOut","replicas":5,"storage":{"accessModes":["ReadWriteOnce"],"resources":{"requests":{"storage":"1Gi"}}},"version":"3.8.3"}}
  creationTimestamp: "2024-09-12T10:40:48Z"
  generation: 2
  labels:
    app.kubernetes.io/component: database
    app.kubernetes.io/instance: sample-zookeeper
    app.kubernetes.io/managed-by: kubedb.com
    app.kubernetes.io/name: zookeepers.kubedb.com
  name: sample-zookeeper
  namespace: demo
  ownerReferences:
    - apiVersion: kubedb.com/v1alpha2
      blockOwnerDeletion: true
      controller: true
      kind: ZooKeeper
      name: sample-zookeeper
      uid: 6d41f283-1a60-45a2-a529-076a09f21ec2
  resourceVersion: "481401"
  uid: db007231-78f1-4ce8-8d2f-adff7d446095
spec:
  appRef:
    apiGroup: kubedb.com
    kind: ZooKeeper
    name: sample-zookeeper
    namespace: demo
  clientConfig:
    service:
      name: sample-zookeeper
      port: 2181
      scheme: http
  secret:
    name: sample-zookeeper-auth
  type: kubedb.com/zookeeper
  version: 3.8.3
```

KubeStash uses the `AppBinding` CR to connect with the target database. It requires the following two fields to set in AppBinding's `.spec` section.

Here,

- `.spec.clientConfig.service.name` specifies the name of the Service that connects to the database.
- `.spec.secret` specifies the name of the Secret that holds necessary credentials to access the database.
- `.spec.type` specifies the types of the app that this AppBinding is pointing to. KubeDB generated AppBinding follows the following format: `<app group>/<app resource type>`.


**Insert Sample Data:**

Now, we are going to exec into one of the database pod and create some sample data. At first, find out the database `Pod` using the following command,

```bash
$ kubectl get pods -n demo --selector="app.kubernetes.io/instance=sample-zookeeper" 
NAME                 READY   STATUS    RESTARTS   AGE
sample-zookeeper-0   2/2     Running   0          16m
sample-zookeeper-1   2/2     Running   0          13m
sample-zookeeper-2   2/2     Running   0          13m
```

Now, let’s exec into the pod and create a directory,

```bash
$ kubectl exec -it -n demo sample-zookeeper-0 -- sh

Type "help" for help.

# Check if Zookeeper server is running and healthy
$ echo ruok | nc localhost 2181
imok

# Create a znode named /hello-dir with the data "hello-message"
$ zkCli.sh create /hello-dir hello-messege
Connecting to localhost:2181
...
Connection Log Messeges
...
Created /hello-dir

# exit from the pod
/ $ exit
```

Now, we are ready to backup the data.

### Prepare Backend

We are going to store our backed up data into a `S3` bucket. We have to create a `Secret` with necessary credentials and a `BackupStorage` CR to use this backend. If you want to use a different backend, please read the respective backend configuration doc from [here](https://kubestash.com/docs/latest/guides/backends/overview/).

**Create Secret:**

Let's create a secret called `s3-secret` with access credentials to our desired s3 bucket,

```bash
$ echo -n '<your-aws-access-key-id-here>' > AWS_ACCESS_KEY_ID
$ echo -n '<your-aws-secret-access-key-here>' > AWS_SECRET_ACCESS_KEY
$ kubectl create secret generic -n demo s3-secret \
    --from-file=./AWS_ACCESS_KEY_ID \
    --from-file=./AWS_SECRET_ACCESS_KEY
secret/s3-secret created
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
      endpoint: ap-south-1.linodeobjects.com
      bucket: kubestash-zk
      region: ap-south-1
      prefix: sep4
      secretName: s3-secret
  usagePolicy:
    allowedNamespaces:
      from: All
  deletionPolicy: WipeOut
```

Let's create the BackupStorage we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/zookeeper/backup/kubestash/logical/examples/backupstorage.yaml
backupstorage.storage.kubestash.com/s3-storage created
```

Now, we are ready to backup our data to our desired backend.

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
      from: Same
```

Let’s create the above `RetentionPolicy`,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/zookeeper/backup/kubestash/logical/examples/retentionpolicy.yaml
retentionpolicy.storage.kubestash.com/demo-retention created
```

### Backup

We have to create a `BackupConfiguration` targeting respective `sample-zookeeper` ZooKeeper. Then, KubeStash will create a `CronJob` for each session to take periodic backup of that database.

At first, we need to create a secret with a Restic password for backup data encryption.

**Create Secret:**

Let's create a secret called `encrypt-secret` with the Restic password,

```bash
$ echo -n 'changeit' > RESTIC_PASSWORD
$ kubectl create secret generic -n demo encrypt-secret \
    --from-file=./RESTIC_PASSWORD
secret "encrypt-secret" created
```

Below is the YAML for `BackupConfiguration` CR to backup the `sample-zookeeper` that we have deployed earlier,

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: sample-zookeeper-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: ZooKeeper
    name: sample-zookeeper
    namespace: demo
  backends:
    - name: s3-backend
      storageRef:
        name: s3-storage
        namespace: demo
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
        - name: s3-zookeeper-repo
          backend: s3-backend
          directory: /zookeeper
          encryptionSecret:
            name: encrypt-secret
            namespace: demo
      addon:
        name: zookeeper-addon
        tasks:
          - name: logical-backup
```

- `.spec.sessions[*].schedule` specifies that we want to backup the database at `5 minutes` interval.
- `.spec.target` refers to the targeted `sample-zookeeper` ZooKeeper that we created earlier.

Let's create the `BackupConfiguration` CR that we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/zookeeper/kubestash/logical/examples/backupconfiguration.yaml
backupconfiguration.core.kubestash.com/sample-zookeeper-backup created
```

**Verify Backup Setup Successful**

If everything goes well, the phase of the `BackupConfiguration` should be `Ready`. The `Ready` phase indicates that the backup setup is successful. Let's verify the `Phase` of the BackupConfiguration,

```bash
$ kubectl get backupconfiguration -n demo
NAME                      PHASE   PAUSED   AGE
sample-zookeeper-backup   Ready            2m50s
```

Additionally, we can verify that the `Repository` specified in the `BackupConfiguration` has been created using the following command,

```bash
$ kubectl get repo -n demo
NAME                  INTEGRITY   SNAPSHOT-COUNT   SIZE     PHASE   LAST-SUCCESSFUL-BACKUP   AGE
s3-zookeeper-repo                 0                0 B      Ready                            3m
```

KubeStash keeps the backup for `Repository` YAMLs. If we navigate to the s3 bucket, we will see the `Repository` YAML stored in the `demo/zookeeper` directory.

**Verify CronJob:**

It will also create a `CronJob` with the schedule specified in `spec.sessions[*].scheduler.schedule` field of `BackupConfiguration` CR.

Verify that the `CronJob` has been created using the following command,

```bash
$ kubectl get cronjob -n demo
NAME                                             SCHEDULE      SUSPEND   ACTIVE   LAST SCHEDULE   AGE
trigger-sample-zookeeper-backup-frequent-backup   */5 * * * *             0        2m45s           3m25s
```

**Verify BackupSession:**

KubeStash triggers an instant backup as soon as the `BackupConfiguration` is ready. After that, backups are scheduled according to the specified schedule.

```bash
$ kubectl get backupsession -n demo -w
NAME                                                 INVOKER-TYPE          INVOKER-NAME               PHASE       DURATION   AGE
sample-zookeeper-backup-frequent-backup-1726572962   BackupConfiguration   sample-zookeeper-backup    Succeeded              7m22s
```

We can see from the above output that the backup session has succeeded. Now, we are going to verify whether the backed up data has been stored in the backend.

**Verify Backup:**

Once a backup is complete, KubeStash will update the respective `Repository` CR to reflect the backup. Check that the repository `sample-zookeeper-backup` has been updated by the following command,

```bash
$ kubectl get repository -n demo s3-zookeeper-repo
NAME                  INTEGRITY   SNAPSHOT-COUNT   SIZE    PHASE   LAST-SUCCESSFUL-BACKUP   AGE
s3-zookeeper-repo     true        1                806 B   Ready   8m27s                    9m18s
```

At this moment we have one `Snapshot`. Run the following command to check the respective `Snapshot` which represents the state of a backup run for an application.

```bash
$ kubectl get snapshots -n demo -l=kubestash.com/repo-name=s3-zookeeper-repo
NAME                                                                   REPOSITORY          SESSION           SNAPSHOT-TIME          DELETION-POLICY   PHASE       AGE
s3-zookeeper-repo-sample-zookeeper-backup-frequent-backup-1726572962   s3-zookeeper-repo   frequent-backup   2024-01-23T13:10:54Z   Delete            Succeeded   16h
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
$ kubectl get snapshots -n demo s3-zookeeper-repo-sample-zookeeper-backup-frequent-backup-1726572962 -oyaml
```

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: Snapshot
metadata:
  creationTimestamp: "2024-09-04T11:30:00Z"
  finalizers:
  - kubestash.com/cleanup
  generation: 1
  labels:
    kubestash.com/app-ref-kind: ZooKeeper
    kubestash.com/app-ref-name: sample-zookeeper
    kubestash.com/app-ref-namespace: demo
    kubestash.com/repo-name: s3-zookeeper-repo
  annotations:
    kubedb.com/db-version: "3.8.3"
  name: s3-zookeeper-repo-sample-zookeeper-backup-frequent-backup-1726572962
  namespace: demo
  ownerReferences:
  - apiVersion: storage.kubestash.com/v1alpha1
    blockOwnerDeletion: true
    controller: true
    kind: Repository
    name: s3-zookeeper-repo
    uid: dd7e2387-227d-4b89-9489-d6255535e322
  resourceVersion: "1226490"
  uid: dd7e2387-227d-4b89-9489-d6255535e322
spec:
  appRef:
    apiGroup: kubedb.com
    kind: ZooKeeper
    name: sample-zookeeper
    namespace: demo
  backupSession: sample-zookeeper-backup-frequent-backup-1726572962
  deletionPolicy: Delete
  repository: s3-zookeeper-repo
  session: frequent-backup
  snapshotID: 01J7ZW9ANMT1GAG6NP68N6Q0MJ
  type: FullBackup
  version: v1
status:
  components:
    dump:
      driver: Restic
      duration: 11.526138009s
      integrity: true
      path: repository/v1/frequent-backup/dump
      phase: Succeeded
      resticStats:
      - hostPath: /kubestash-interim/data
        id: cd20fca1a2bf6a97e669cb9eacdc74a312a08266da92b9d687aad88841e1205d
        size: 3.345 KiB
        uploaded: 299 B
      size: 2.202 KiB
  conditions:
  - lastTransitionTime: "2024-09-04T11:30:00Z"
    message: Recent snapshot list updated successfully
    reason: SuccessfullyUpdatedRecentSnapshotList
    status: "True"
    type: RecentSnapshotListUpdated
  - lastTransitionTime: "2024-09-04T11:30:32Z"
    message: Metadata uploaded to backend successfully
    reason: SuccessfullyUploadedSnapshotMetadata
    status: "True"
    type: SnapshotMetadataUploaded
  integrity: true
  phase: Succeeded
  size: 2.201 KiB
  snapshotTime: "2024-09-04T11:30:00Z"
  totalComponents: 1
```

> KubeStash uses `zk-dump-go` to perform backups of target `ZooKeeper`. Therefore, the component name for logical backups is set as `dump`.

Now, if we navigate to the s3 bucket, we will see the backed up data stored in the `demo/zookeeper/repository/v1/frequent-backup/dump` directory. KubeStash also keeps the backup for `Snapshot` YAMLs, which can be found in the `demo/zookeeper/snapshots` directory.

> Note: KubeStash stores all dumped data encrypted in the backup directory, meaning it remains unreadable until decrypted.

## Restore

In this section, we are going to restore the database from the backup we have taken in the previous section. We are going to deploy a new database and initialize it from the backup.

Now, we have to deploy the restored database similarly as we have deployed the original `sample-zookeeper`. However, this time there will be the following differences:

- We are going to specify `.spec.init.waitForInitialRestore` field that tells KubeDB to wait for first restore to complete before marking this database is ready to use.

Below is the YAML for `ZooKeeper` CR we are going deploy to initialize from backup,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ZooKeeper
metadata:
  name: restored-zookeeper
  namespace: demo
spec:
  version: "3.8.3"
  adminServerPort: 8080
  replicas: 3
  storage:
    resources:
      requests:
        storage: "1Gi"
    accessModes:
      - ReadWriteOnce
  deletionPolicy: "WipeOut"
```

Let's create the above database,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/zookeeper/backup/kubestash/logical/examples/restored-zookeeper.yaml
zookeeper.kubedb.com/restored-zookeeper created
```

If you check the database status, you will see it is stuck in **`Provisioning`** state.

```bash
$ kubectl get zookeeper -n demo restored-zookeeper
NAME                 VERSION   STATUS         AGE
restored-zookeeper   3.8.3     Provisioning   61s
```

#### Create RestoreSession:

Now, we need to create a `RestoreSession` CR pointing to targeted `ZooKeeper`.

Below, is the contents of YAML file of the `RestoreSession` object that we are going to create to restore backed up data into the newly created `ZooKeeper` named `restored-zookeeper`.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: sample-zookeeper-restore
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: ZooKeeper
    namespace: demo
    name: restored-zookeeper
  dataSource:
    repository: s3-zookeeper-repo
    snapshot: latest
    encryptionSecret:
      name: encrypt-secret
      namespace: demo
  addon:
    name: zookeeper-addon
    tasks:
      - name: logical-backup-restore
```

Here,

- `.spec.target` refers to the newly created `restored-zookeeper` ZooKeeper object to where we want to restore backup data.
- `.spec.dataSource.repository` specifies the Repository object that holds the backed up data.
- `.spec.dataSource.snapshot` specifies to restore from latest `Snapshot`.

Let's create the RestoreSession CRD object we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/zookeeper/backup/kubestash/logical/examples/restoresession.yaml
restoresession.core.kubestash.com/sample-zookeeper-restore created
```

Once, you have created the `RestoreSession` object, KubeStash will create restore Job. Run the following command to watch the phase of the `RestoreSession` object,

```bash
$ watch kubectl get restoresession -n demo
Every 2.0s: kubectl get restores... AppsCode-PC-03: Wed Aug 21 10:44:05 2024
NAME                      REPOSITORY          FAILURE-POLICY   PHASE       DURATION   AGE
sample-zookeeper-restore   gcs-zookeeper-repo                    Succeeded   7s         116s
```

The `Succeeded` phase means that the restore process has been completed successfully.

#### Verify Restored Data:

In this section, we are going to verify whether the desired data has been restored successfully. We are going to connect to the database server and check whether the database and the table we created earlier in the original database are restored.

At first, check if the database has gone into **`Ready`** state by the following command,

```bash
$ kubectl get zookeeper -n demo restored-zookeeper
NAME                 VERSION   STATUS   AGE
restored-zookeeper   8.3.1      Ready    6m31s
```

Now, find out the database `Pod` by the following command,

```bash
$ kubectl get pods -n demo --selector="app.kubernetes.io/instance=restored-zookeeper"
NAME                             READY   STATUS      RESTARTS   AGE
restored-zookeeper-0             2/2     Running     0          6m7s
restored-zookeeper-1             2/2     Running     0          6m1s
restored-zookeeper-2             2/2     Running     0          5m55s
```

Now, lets exec one of the `Pod` and verify restored data.

```bash
$ kubectl exec -it -n demo restored-zookeeper-0 -- sh

Type "help" for help.

# Check if Zookeeper server is running and healthy
$ echo ruok | nc localhost 2181
imok

# List all znodes from the root directory
$ zkCli.sh ls /
Connecting to localhost:2181
...
Connection Log Messeges
...
[hello-dir]

# Verify the data stored in the /hello-dir znode
$ zkCli.sh get /hello-dir
Connecting to localhost:2181
...
Connection Log Messeges
...
hello-messege

# exit from the pod
/ $ exit
```

So, from the above output, we can see the `demo` database we had created in the original database `sample-zookeeper` has been restored in the `restored-zookeeper`.

## Cleanup

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete backupconfigurations.core.kubestash.com  -n demo sample-zookeeper-backup
kubectl delete restoresessions.core.kubestash.com -n demo restore-sample-zookeeper
kubectl delete retentionpolicies.storage.kubestash.com -n demo demo-retention
kubectl delete backupstorage -n demo s3-storage
kubectl delete secret -n demo s3-secret
kubectl delete secret -n demo encrypt-secret
kubectl delete zookeeper -n demo restored-zookeeper
kubectl delete zookeeper -n demo sample-zookeeper
```