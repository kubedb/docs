---
title: ZooKeeper Auto-Backup | KubeStash
description: Backup ZooKeeper using KubeStash Auto-Backup
menu:
  docs_{{ .version }}:
    identifier: guides-zk-auto-backup-stashv2
    name: Auto-Backup
    parent: guides-zk-backup-stashv2
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Backup ZooKeeper using KubeStash Auto-Backup

KubeStash can automatically be configured to backup any `ZooKeeper` in your cluster. KubeStash enables cluster administrators to deploy backup `blueprints` ahead of time so database owners can easily backup any `ZooKeeper` database with a few annotations.

In this tutorial, we are going to show how you can configure a backup blueprint for `ZooKeeper` in your cluster and backup them with a few annotations.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using `Minikube` or `Kind`.
- Install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md).
- Install `KubeStash` in your cluster following the steps [here](https://kubestash.com/docs/latest/setup/install/kubestash).
- Install KubeStash `kubectl` plugin following the steps [here](https://kubestash.com/docs/latest/setup/install/kubectl-plugin/).
- If you are not familiar with how KubeStash backup and restore `ZooKeeper`, please check the following guide [here](/docs/guides/zookeeper/backup/kubestash/overview/index.md).

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

> **Note:** YAML files used in this tutorial are stored in [docs/guides/zookeeper/backup/kubestash/auto-backup/examples](https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/zookeeper/backup/kubestash/auto-backup/examples) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

### Prepare Backend

We are going to store our backed up data into a `s3` bucket. We have to create a `Secret` with necessary credentials and a `BackupStorage` CR to use this backend. If you want to use a different backend, please read the respective backend configuration doc from [here](https://kubestash.com/docs/latest/guides/backends/overview/).

**Create Secret:**

Let's create a secret called `s3-secret` with access credentials to our desired GCS bucket,

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
      bucket: rudro
      region: ap-south-1
      prefix: blueprint
      secretName: s3-secret
  usagePolicy:
    allowedNamespaces:
      from: All
  deletionPolicy: WipeOut
```

Let's create the BackupStorage we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/zookeeper/backup/kubestash/auto-backup/examples/backupstorage.yaml
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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/zookeeper/backup/kubestash/auto-backup/examples/retentionpolicy.yaml
retentionpolicy.storage.kubestash.com/demo-retention created
```

**Create Secret:**

We also need to create a secret with a `Restic` password for backup data encryption.

Let's create a secret called `encrypt-secret` with the Restic password,

```bash
$ echo -n 'changeit' > RESTIC_PASSWORD
$ kubectl create secret generic -n demo encrypt-secret \
    --from-file=./RESTIC_PASSWORD \
secret "encrypt-secret" created
```

## Auto-backup with default configurations

In this section, we are going to backup a `ZooKeeper` of `demo` namespace. We are going to use the default configurations which will be specified in the `BackupBlueprint` CR.

**Prepare Backup Blueprint**

A `BackupBlueprint` allows you to specify a template for the `Repository`,`Session` or `Variables` of `BackupConfiguration` in a Kubernetes native way.

Now, we have to create a `BackupBlueprint` CR with a blueprint for `BackupConfiguration` object.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupBlueprint
metadata:
  name: zookeeper-default-backup-blueprint
  namespace: demo
spec:
  usagePolicy:
    allowedNamespaces:
      from: All
  backupConfigurationTemplate:
    deletionPolicy: OnDelete
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
        sessionHistoryLimit: 3
        scheduler:
          schedule: "*/5 * * * *"
          jobTemplate:
            backoffLimit: 1
        repositories:
          - name: default-blueprint
            backend: s3-backend
            directory: /default-blueprint
            encryptionSecret:
              name: encrypt-secret
              namespace: demo
        addon:
          name: zookeeper-addon
          tasks:
            - name: logical-backup
```

Here,

- `.spec.backupConfigurationTemplate.backends[*].storageRef` refers our earlier created `s3-storage` backupStorage.
- `.spec.backupConfigurationTemplate.sessions[*].schedule` specifies that we want to backup the database at `5 minutes` interval.

Let's create the `BackupBlueprint` we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/zookeeper/backup/kubestash/auto-backup/examples/default-backupblueprint.yaml
backupblueprint.core.kubestash.com/zookeeper-default-backup-blueprint created
```

Now, we are ready to backup our `ZooKeeper` using few annotations.

**Create Database**

Now, we are going to create an `ZooKeeper` CR in demo namespace.

Below is the YAML of the `ZooKeeper` object that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ZooKeeper
metadata:
  name: sample-zookeeper
  namespace: demo
  annotations:
    blueprint.kubestash.com/name: zookeeper-default-backup-blueprint
    blueprint.kubestash.com/namespace: demo
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

Here,

- `.spec.annotations.blueprint.kubestash.com/name: zookeeper-default-backup-blueprint` specifies the name of the `BackupBlueprint` that will use in backup.
- `.spec.annotations.blueprint.kubestash.com/namespace: demo` specifies the name of the `namespace` where the `BackupBlueprint` resides.

Let's create the `ZooKeeper` we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/zookeeper/backup/kubestash/auto-backup/examples/sample-zookeeper.yaml
zookeeper.kubedb.com/sample-zookeeper created
```

**Verify BackupConfiguration**

If everything goes well, KubeStash should create a `BackupConfiguration` for our ZooKeeper in demo namespace and the phase of that `BackupConfiguration` should be `Ready`. Verify the `BackupConfiguration` object by the following command,

```bash
$ kubectl get backupconfiguration -n demo
NAME                          PHASE   PAUSED   AGE
appbinding-sample-zookeeper   Ready            2m50m
```

Now, let’s check the YAML of the `BackupConfiguration`.

```bash
$ kubectl get backupconfiguration -n demo appbinding-sample-zookeeper  -o yaml
```

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  creationTimestamp: "2024-09-19T08:50:44Z"
  finalizers:
    - kubestash.com/cleanup
  generation: 1
  labels:
    app.kubernetes.io/managed-by: kubestash.com
    kubestash.com/invoker-name: zookeeper-default-backup-blueprint
    kubestash.com/invoker-namespace: demo
  name: appbinding-sample-zookeeper
  namespace: demo
  resourceVersion: "1509329"
  uid: 1e99efc6-7ede-4c32-9dd0-da9dec0eb28c
spec:
  backends:
    - name: s3-backend
      retentionPolicy:
        name: demo-retention
        namespace: demo
      storageRef:
        name: s3-storage
        namespace: demo
  sessions:
    - addon:
        name: zookeeper-addon
        tasks:
          - name: logical-backup
      name: frequent-backup
      repositories:
        - backend: s3-backend
          directory: /default-blueprint
          encryptionSecret:
            name: encrypt-secret
            namespace: demo
          name: default-blueprint
      scheduler:
        jobTemplate:
          backoffLimit: 1
          template:
            controller: {}
            metadata: {}
            spec:
              resources: {}
        schedule: '*/5 * * * *'
      sessionHistoryLimit: 3
  target:
    apiGroup: kubedb.com
    kind: ZooKeeper
    name: sample-zookeeper
    namespace: demo
status:
  backends:
    - name: s3-backend
      ready: true
      retentionPolicy:
        found: true
        ref:
          name: demo-retention
          namespace: demo
      storage:
        phase: Ready
        ref:
          name: s3-storage
          namespace: demo
  conditions:
    - lastTransitionTime: "2024-09-19T08:50:44Z"
      message: Validation has been passed successfully.
      reason: ResourceValidationPassed
      status: "True"
      type: ValidationPassed
  dependencies:
    - found: true
      kind: Addon
      name: zookeeper-addon
  phase: Ready
  repositories:
    - name: default-blueprint
      phase: Ready
  sessions:
    - conditions:
        - lastTransitionTime: "2024-09-19T08:50:54Z"
          message: Scheduler has been ensured successfully.
          reason: SchedulerEnsured
          status: "True"
          type: SchedulerEnsured
        - lastTransitionTime: "2024-09-19T08:50:54Z"
          message: Initial backup has been triggered successfully.
          reason: SuccessfullyTriggeredInitialBackup
          status: "True"
          type: InitialBackupTriggered
      name: frequent-backup
  targetFound: true

```

Notice the `spec.backends`, `spec.sessions` and `spec.target` sections, KubeStash automatically resolved those info from the `BackupBluePrint` and created above `BackupConfiguration`.

**Verify BackupSession:**

KubeStash triggers an instant backup as soon as the `BackupConfiguration` is ready. After that, backups are scheduled according to the specified schedule.

```bash
$ kubectl get backupsession -n demo -w
NAME                                                     INVOKER-TYPE          INVOKER-NAME                  PHASE       DURATION   AGE
appbinding-sample-zookeeper-frequent-backup-1726735844   BackupConfiguration   appbinding-sample-zookeeper   Succeeded   23s        6m40s
```

We can see from the above output that the backup session has succeeded. Now, we are going to verify whether the backed up data has been stored in the backend.

**Verify Backup:**

Once a backup is complete, KubeStash will update the respective `Repository` CR to reflect the backup. Check that the repository `default-blueprint` has been updated by the following command,

```bash
$ kubectl get repository -n demo default-blueprint
NAME                INTEGRITY   SNAPSHOT-COUNT   SIZE        PHASE   LAST-SUCCESSFUL-BACKUP   AGE
default-blueprint   true        1                1.559 KiB   Ready   80s                      7m32s
```

At this moment we have one `Snapshot`. Run the following command to check the respective `Snapshot` which represents the state of a backup run for an application.

```bash
$ kubectl get snapshots -n demo -l=kubestash.com/repo-name=default-blueprint
NAME                                                              REPOSITORY          SESSION           SNAPSHOT-TIME          DELETION-POLICY   PHASE       AGE
default-blueprint-appbinding-samgres-frequent-backup-1726736101   default-blueprint   frequent-backup   2024-09-19T08:55:01Z   Delete            Succeeded   7m48s
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
$ kubectl get snapshots -n demo default-blueprint-appbinding-sameper-frequent-backup-1726736101 -oyaml
```

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: Snapshot
metadata:
  creationTimestamp: "2024-09-19T08:55:01Z"
  finalizers:
    - kubestash.com/cleanup
  generation: 1
  labels:
    kubedb.com/db-version: 3.8.3
    kubestash.com/app-ref-kind: ZooKeeper
    kubestash.com/app-ref-name: sample-zookeeper
    kubestash.com/app-ref-namespace: demo
    kubestash.com/repo-name: default-blueprint
  name: default-blueprint-appbinding-sameper-frequent-backup-1726736101
  namespace: demo
  ownerReferences:
    - apiVersion: storage.kubestash.com/v1alpha1
      blockOwnerDeletion: true
      controller: true
      kind: Repository
      name: default-blueprint
      uid: 7cfd944f-1daa-4306-95c8-2ff7f41fd766
  resourceVersion: "1509911"
  uid: f0dd0e7f-e4ec-42a4-8432-e4b44b2d0ada
spec:
  appRef:
    apiGroup: kubedb.com
    kind: ZooKeeper
    name: sample-zookeeper
    namespace: demo
  backupSession: appbinding-sample-zookeeper-frequent-backup-1726736101
  deletionPolicy: Delete
  repository: default-blueprint
  session: frequent-backup
  snapshotID: 01J84QVVR9Y1ZCRKXJANQB21WE
  type: FullBackup
  version: v1
status:
  components:
    dump:
      driver: Restic
      duration: 2.069889135s
      integrity: true
      path: repository/v1/frequent-backup/dump
      phase: Succeeded
      resticStats:
        - hostPath: /kubestash-interim/data
          id: 280cbe5908773859259f0921d89e677f4c0ab40c3e4bedee6b95ab2c9ef474e9
          size: 718 B
          uploaded: 1.075 KiB
      size: 1.835 KiB
  conditions:
    - lastTransitionTime: "2024-09-19T08:55:01Z"
      message: Recent snapshot list updated successfully
      reason: SuccessfullyUpdatedRecentSnapshotList
      status: "True"
      type: RecentSnapshotListUpdated
    - lastTransitionTime: "2024-09-19T08:55:08Z"
      message: Metadata uploaded to backend successfully
      reason: SuccessfullyUploadedSnapshotMetadata
      status: "True"
      type: SnapshotMetadataUploaded
  integrity: true
  phase: Succeeded
  size: 1.835 KiB
  snapshotTime: "2024-09-19T08:55:01Z"
  totalComponents: 1

```

> KubeStash uses `zk-dump-go` to perform backups of target `ZooKeeper`. Therefore, the component name for logical backups is set as `dump`.

Now, if we navigate to the s3 bucket, we will see the backed up data stored in the `blueprint/default-blueprint/repository/v1/frequent-backup/dump` directory. KubeStash also keeps the backup for `Snapshot` YAMLs, which can be found in the `blueprint/default-blueprint/snapshots` directory.

> Note: KubeStash stores all dumped data encrypted in the backup directory, meaning it remains unreadable until decrypted.

## Auto-backup with custom configurations

In this section, we are going to backup a `ZooKeeper` of `demo` namespace. We are going to use the custom configurations which will be specified in the `BackupBlueprint` CR.

**Prepare Backup Blueprint**

A `BackupBlueprint` allows you to specify a template for the `Repository`,`Session` or `Variables` of `BackupConfiguration` in a Kubernetes native way.

Now, we have to create a `BackupBlueprint` CR with a blueprint for `BackupConfiguration` object.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupBlueprint
metadata:
  name: zookeeper-customize-backup-blueprint
  namespace: demo
spec:
  usagePolicy:
    allowedNamespaces:
      from: Same
  backupConfigurationTemplate:
    deletionPolicy: OnDelete
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
        sessionHistoryLimit: 3
        scheduler:
          schedule: ${schedule}
          jobTemplate:
            backoffLimit: 1
        repositories:
          - name: ${repoName}
            backend: s3-backend
            directory: ${namespace}/${targetName}
            encryptionSecret:
              name: encrypt-secret
              namespace: demo
        addon:
          name: zookeeper-addon
          tasks:
            - name: logical-backup
```

Note that we have used some variables (format: `${<variable name>}`) in different fields. KubeStash will substitute these variables with values from the respective target’s annotations. You’re free to use any variables you like.

Here,

- `.spec.backupConfigurationTemplate.backends[*].storageRef` refers our earlier created `s3-storage` backupStorage.
- `.spec.backupConfigurationTemplate.sessions[*]`:
    - `.schedule` defines `${schedule}` variable, which determines the time interval for the backup.
    - `.repositories[*].name` defines the `${repoName}` variable, which specifies the name of the backup `Repository`.
    - `.repositories[*].directory` defines two variables, `${namespace}` and `${targetName}`, which are used to determine the path where the backup will be stored.
    - `.addon.tasks[*].params.args` defines `${targetedDatabase}` variable, which identifies a single database to backup.

Let's create the `BackupBlueprint` we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/zookeeper/backup/kubestash/auto-backup/examples/customize-backupblueprint.yaml
backupblueprint.core.kubestash.com/zookeeper-customize-backup-blueprint created
```

Now, we are ready to backup our `ZooKeeper` using few annotations. You can check available auto-backup annotations for a databases from [here](https://kubestash.com/docs/latest/concepts/crds/backupblueprint/).

**Create Database**

Now, we are going to create an `ZooKeeper` CR in demo namespace.

Below is the YAML of the `ZooKeeper` object that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: ZooKeeper
metadata:
  name: sample-zookeeper-2
  namespace: demo
  annotations:
    blueprint.kubestash.com/name: zookeeper-customize-backup-blueprint
    blueprint.kubestash.com/namespace: demo
    variables.kubestash.com/schedule: "*/10 * * * *"
    variables.kubestash.com/repoName: customize-blueprint
    variables.kubestash.com/namespace: demo
    variables.kubestash.com/targetName: sample-zookeeper-2
    variables.kubestash.com/targetedDatabase: zookeeper
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

Notice the `metadata.annotations` field, where we have defined the annotations related to the automatic backup configuration. Specifically, we've set the `BackupBlueprint` name as `zookeeper-customize-backup-blueprint` and the namespace as `demo`. We have also provided values for the blueprint template variables, such as the backup `schedule`, `repositoryName`, `namespace`, `targetName`, and `targetedDatabase`. These annotations will be used to create a `BackupConfiguration` for this `ZooKeeper` database.

Let's create the `ZooKeeper` we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/zookeeper/backup/kubestash/auto-backup/examples/sample-zookeeper-2.yaml
zookeeper.kubedb.com/sample-zookeeper-2 created
```

**Verify BackupConfiguration**

If everything goes well, KubeStash should create a `BackupConfiguration` for our ZooKeeper in demo namespace and the phase of that `BackupConfiguration` should be `Ready`. Verify the `BackupConfiguration` object by the following command,

```bash
$ kubectl get backupconfiguration -n demo
NAME                           PHASE   PAUSED      AGE
appbinding-sample-zookeeper-2   Ready               2m50m
```

Now, let’s check the YAML of the `BackupConfiguration`.

```bash
$ kubectl get backupconfiguration -n demo appbinding-sample-zookeeper-2  -o yaml
```

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  creationTimestamp: "2024-09-19T10:37:06Z"
  finalizers:
    - kubestash.com/cleanup
  generation: 1
  labels:
    app.kubernetes.io/managed-by: kubestash.com
    kubestash.com/invoker-name: zookeeper-customize-backup-blueprint
    kubestash.com/invoker-namespace: demo
  name: appbinding-sample-zookeeper-2
  namespace: demo
  resourceVersion: "1521726"
  uid: 2c80f9a0-b79e-4733-9e68-715ca6b55b93
spec:
  backends:
    - name: s3-backend
      retentionPolicy:
        name: demo-retention
        namespace: demo
      storageRef:
        name: s3-storage
        namespace: demo
  sessions:
    - addon:
        name: zookeeper-addon
        tasks:
          - name: logical-backup
      name: frequent-backup
      repositories:
        - backend: s3-backend
          directory: demo/sample-zookeeper-2
          encryptionSecret:
            name: encrypt-secret
            namespace: demo
          name: customize-blueprint
      scheduler:
        jobTemplate:
          backoffLimit: 1
          template:
            controller: {}
            metadata: {}
            spec:
              resources: {}
        schedule: '*/10 * * * *'
      sessionHistoryLimit: 3
  target:
    apiGroup: kubedb.com
    kind: ZooKeeper
    name: sample-zookeeper-2
    namespace: demo
status:
  backends:
    - name: s3-backend
      ready: true
      retentionPolicy:
        found: true
        ref:
          name: demo-retention
          namespace: demo
      storage:
        phase: Ready
        ref:
          name: s3-storage
          namespace: demo
  conditions:
    - lastTransitionTime: "2024-09-19T10:37:06Z"
      message: Validation has been passed successfully.
      reason: ResourceValidationPassed
      status: "True"
      type: ValidationPassed
  dependencies:
    - found: true
      kind: Addon
      name: zookeeper-addon
  phase: Ready
  repositories:
    - name: customize-blueprint
      phase: Ready
  sessions:
    - conditions:
        - lastTransitionTime: "2024-09-19T10:37:16Z"
          message: Scheduler has been ensured successfully.
          reason: SchedulerEnsured
          status: "True"
          type: SchedulerEnsured
        - lastTransitionTime: "2024-09-19T10:37:16Z"
          message: Initial backup has been triggered successfully.
          reason: SuccessfullyTriggeredInitialBackup
          status: "True"
          type: InitialBackupTriggered
      name: frequent-backup
  targetFound: true
```

Notice the `spec.backends`, `spec.sessions` and `spec.target` sections, KubeStash automatically resolved those info from the `BackupBluePrint` and created above `BackupConfiguration`.

**Verify BackupSession:**

KubeStash triggers an instant backup as soon as the `BackupConfiguration` is ready. After that, backups are scheduled according to the specified schedule.

```bash
$ kubectl get backupsession -n demo -w
NAME                                                        INVOKER-TYPE          INVOKER-NAME                   PHASE       DURATION   AGE
appbinding-sample-zookeeper-2-frequent-backup-1726742400    BackupConfiguration   appbinding-sample-zookeeper     Succeeded   58s        112s
```

We can see from the above output that the backup session has succeeded. Now, we are going to verify whether the backed up data has been stored in the backend.

**Verify Backup:**

Once a backup is complete, KubeStash will update the respective `Repository` CR to reflect the backup. Check that the repository `customize-blueprint` has been updated by the following command,

```bash
$ kubectl get repository -n demo customize-blueprint
NAME                         INTEGRITY   SNAPSHOT-COUNT   SIZE    PHASE   LAST-SUCCESSFUL-BACKUP   AGE
customize-blueprint          true        1                806 B   Ready   8m27s                    9m18s
```

At this moment we have one `Snapshot`. Run the following command to check the respective `Snapshot` which represents the state of a backup run for an application.

```bash
$ kubectl get snapshots -n demo -l=kubestash.com/repo-name=customize-blueprint
NAME                                                              REPOSITORY            SESSION           SNAPSHOT-TIME          DELETION-POLICY   PHASE       AGE
customize-blueprint-appbinding-ser-2-frequent-backup-1726742400   customize-blueprint   frequent-backup   2024-09-19T10:40:01Z   Delete            Succeeded   6m19s
```

> Note: KubeStash creates a `Snapshot` with the following labels:
> - `kubedb.com/db-version: <db-version>`
> - `kubestash.com/app-ref-kind: <target-kind>`
> - `kubestash.com/app-ref-name: <target-name>`
> - `kubestash.com/app-ref-namespace: <target-namespace>`
> - `kubestash.com/repo-name: <repository-name>`
>
> These labels can be used to watch only the `Snapshot`s related to our target Database or `Repository`.

If we check the YAML of the `Snapshot`, we can find the information about the backed up components of the Database.

```bash
$ kubectl get snapshots -n demo customize-blueprint-appbinding-sql-2-frequent-backup-1725597000 -oyaml
```

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: Snapshot
metadata:
  creationTimestamp: "2024-09-19T10:40:01Z"
  finalizers:
    - kubestash.com/cleanup
  generation: 1
  labels:
    kubedb.com/db-version: 3.8.3
    kubestash.com/app-ref-kind: ZooKeeper
    kubestash.com/app-ref-name: sample-zookeeper-2
    kubestash.com/app-ref-namespace: demo
    kubestash.com/repo-name: customize-blueprint
  name: customize-blueprint-appbinding-ser-2-frequent-backup-1726742400
  namespace: demo
  ownerReferences:
    - apiVersion: storage.kubestash.com/v1alpha1
      blockOwnerDeletion: true
      controller: true
      kind: Repository
      name: customize-blueprint
      uid: c23b7c97-f97a-4a46-81b1-74a3724d08da
  resourceVersion: "1522158"
  uid: 193748f3-e417-4b62-a092-be38ee1fa6b7
spec:
  appRef:
    apiGroup: kubedb.com
    kind: ZooKeeper
    name: sample-zookeeper-2
    namespace: demo
  backupSession: appbinding-sample-zookeeper-2-frequent-backup-1726742400
  deletionPolicy: Delete
  repository: customize-blueprint
  session: frequent-backup
  snapshotID: 01J84XW407F24E7EFKJFBVS36E
  type: FullBackup
  version: v1
status:
  components:
    dump:
      driver: Restic
      duration: 2.426610729s
      integrity: true
      path: repository/v1/frequent-backup/dump
      phase: Succeeded
      resticStats:
        - hostPath: /kubestash-interim/data
          id: 5356f92f84ad9b1ce170a99796eee91983e6df62e224d592d85fb0811a2fbb38
          size: 719 B
          uploaded: 1.075 KiB
      size: 1.839 KiB
  conditions:
    - lastTransitionTime: "2024-09-19T10:40:01Z"
      message: Recent snapshot list updated successfully
      reason: SuccessfullyUpdatedRecentSnapshotList
      status: "True"
      type: RecentSnapshotListUpdated
    - lastTransitionTime: "2024-09-19T10:40:09Z"
      message: Metadata uploaded to backend successfully
      reason: SuccessfullyUploadedSnapshotMetadata
      status: "True"
      type: SnapshotMetadataUploaded
  integrity: true
  phase: Succeeded
  size: 1.839 KiB
  snapshotTime: "2024-09-19T10:40:01Z"
  totalComponents: 1

```

> KubeStash uses `zk-dump-go` to perform backups of target `ZooKeeper`. Therefore, the component name for logical backups is set as `dump`.

Now, if we navigate to the s3 bucket, we will see the backed up data stored in the `blueprint/demo/sample-zookeeper-2/repository/v1/frequent-backup/dump` directory. KubeStash also keeps the backup for `Snapshot` YAMLs, which can be found in the `blueprint/demo/sample-zookeeper-2/snapshots` directory.

> Note: KubeStash stores all dumped data encrypted in the backup directory, meaning it remains unreadable until decrypted.

## Cleanup

To cleanup the resources crated by this tutorial, run the following commands,

```bash
kubectl delete backupblueprints.core.kubestash.com  -n demo zookeeper-default-backup-blueprint
kubectl delete backupblueprints.core.kubestash.com  -n demo zookeeper-customize-backup-blueprint
kubectl delete retentionpolicies.storage.kubestash.com -n demo demo-retention
kubectl delete backupstorage -n demo s3-storage
kubectl delete secret -n demo s3-secret
kubectl delete secret -n demo encrypt-secret
kubectl delete zookeeper -n demo sample-zookeeper
kubectl delete zookeeper -n demo sample-zookeeper-2
```