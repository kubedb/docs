---
title: Redis Auto-Backup | KubeStash
description: Backup Redis using KubeStash Auto-Backup
menu:
  docs_{{ .version }}:
    identifier: guides-rd-auto-backup-stashv2
    name: Auto-Backup
    parent: guides-rd-backup-stashv2
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Backup Redis using KubeStash Auto-Backup

KubeStash can automatically be configured to backup any `Redis` databases in your cluster. KubeStash enables cluster administrators to deploy backup `blueprints` ahead of time so database owners can easily backup any `Redis` database with a few annotations.

In this tutorial, we are going to show how you can configure a backup blueprint for `Redis` databases in your cluster and backup them with a few annotations.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using `Minikube` or `Kind`.
- Install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md).
- Install `KubeStash` in your cluster following the steps [here](https://kubestash.com/docs/latest/setup/install/kubestash).
- Install KubeStash `kubectl` plugin following the steps [here](https://kubestash.com/docs/latest/setup/install/kubectl-plugin/).
- If you are not familiar with how KubeStash backup and restore `Redis` databases, please check the following guide [here](/docs/guides/redis/backup/kubestash/overview/index.md).

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

> **Note:** YAML files used in this tutorial are stored in [docs/guides/redis/backup/kubestash/auto-backup/examples](https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/redis/backup/kubestash/auto-backup/examples) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

### Prepare Backend

We are going to store our backed up data into a `GCS` bucket. We have to create a `Secret` with necessary credentials and a `BackupStorage` CR to use this backend. If you want to use a different backend, please read the respective backend configuration doc from [here](https://kubestash.com/docs/latest/guides/backends/overview/).

**Create Secret:**

Let's create a secret called `gcs-secret` with access credentials to our desired GCS bucket,

```bash
$ echo -n '<your-project-id>' > GOOGLE_PROJECT_ID
$ cat /path/to/downloaded-sa-key.json > GOOGLE_SERVICE_ACCOUNT_JSON_KEY
$ kubectl create secret generic -n demo gcs-secret \
    --from-file=./GOOGLE_PROJECT_ID \
    --from-file=./GOOGLE_SERVICE_ACCOUNT_JSON_KEY
secret/gcs-secret created
```

**Create BackupStorage:**

Now, create a `BackupStorage` using this secret. Below is the YAML of `BackupStorage` CR we are going to create,

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: BackupStorage
metadata:
  name: gcs-storage
  namespace: demo
spec:
  storage:
    provider: gcs
    gcs:
      bucket: neaj-demo
      prefix: blueprint
      secretName: gcs-secret
  usagePolicy:
    allowedNamespaces:
      from: All
  default: true
  deletionPolicy: Delete
```

Let's create the BackupStorage we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/redis/backup/kubestash/auto-backup/examples/backupstorage.yaml
backupstorage.storage.kubestash.com/gcs-storage created
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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/redis/backup/kubestash/auto-backup/examples/retentionpolicy.yaml
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

In this section, we are going to backup a `Redis` database of `demo` namespace. We are going to use the default configurations which will be specified in the `BackupBlueprint` CR.

**Prepare Backup Blueprint**

A `BackupBlueprint` allows you to specify a template for the `Repository`,`Session` or `Variables` of `BackupConfiguration` in a Kubernetes native way.

Now, we have to create a `BackupBlueprint` CR with a blueprint for `BackupConfiguration` object.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupBlueprint
metadata:
  name: redis-default-backup-blueprint
  namespace: demo
spec:
  usagePolicy:
    allowedNamespaces:
      from: All
  backupConfigurationTemplate:
    deletionPolicy: OnDelete
    backends:
      - name: gcs-backend
        storageRef:
          namespace: demo
          name: gcs-storage
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
            backend: gcs-backend
            directory: /default-blueprint
            encryptionSecret:
              name: encrypt-secret
              namespace: demo
        addon:
          name: redis-addon
          tasks:
            - name: logical-backup
```

Here,

- `.spec.backupConfigurationTemplate.backends[*].storageRef` refers our earlier created `gcs-storage` backupStorage.
- `.spec.backupConfigurationTemplate.sessions[*].schedule` specifies that we want to backup the database at `5 minutes` interval.

Let's create the `BackupBlueprint` we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/redis/backup/kubestash/auto-backup/examples/default-backupblueprint.yaml
backupblueprint.core.kubestash.com/redis-default-backup-blueprint created
```

Now, we are ready to backup our `Redis` databases using few annotations.

**Create Database**

Now, we are going to create an `Redis` CR in demo namespace. 

Below is the YAML of the `Redis` object that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: Redis
metadata:
  name: redis-standalone
  namespace: demo
  annotations:
    blueprint.kubestash.com/name: redis-default-backup-blueprint
    blueprint.kubestash.com/namespace: demo
spec:
  version: 7.4.0
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: Delete
```

Here,

- `.spec.annotations.blueprint.kubestash.com/name: redis-default-backup-blueprint` specifies the name of the `BackupBlueprint` that will use in backup.
- `.spec.annotations.blueprint.kubestash.com/namespace: demo` specifies the name of the `namespace` where the `BackupBlueprint` resides.

Let's create the `Redis` we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/redis/backup/kubestash/auto-backup/examples/redis-standalone.yaml
redis.kubedb.com/redis-standalone created
```

**Verify BackupConfiguration**

If everything goes well, KubeStash should create a `BackupConfiguration` for our Redis in demo namespace and the phase of that `BackupConfiguration` should be `Ready`. Verify the `BackupConfiguration` object by the following command,

```bash
$ kubectl get backupconfiguration -n demo
NAME                         PHASE   PAUSED   AGE
appbinding-redis-standalone   Ready            2m50m
```

Now, let’s check the YAML of the `BackupConfiguration`.

```bash
$ kubectl get backupconfiguration -n demo appbinding-redis-standalone  -o yaml
```

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  creationTimestamp: "2024-09-18T12:15:07Z"
  finalizers:
    - kubestash.com/cleanup
  generation: 1
  labels:
    app.kubernetes.io/managed-by: kubestash.com
    kubestash.com/invoker-name: redis-default-backup-blueprint
    kubestash.com/invoker-namespace: demo
  name: appbinding-redis-standalone
  namespace: demo
  resourceVersion: "1176493"
  uid: b7a37776-4a9b-4aaa-9e1d-15e7d6a83d56
spec:
  backends:
    - name: gcs-backend
      retentionPolicy:
        name: demo-retention
        namespace: demo
      storageRef:
        name: gcs-storage
        namespace: demo
  sessions:
    - addon:
        name: redis-addon
        tasks:
          - name: logical-backup
      name: frequent-backup
      repositories:
        - backend: gcs-backend
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
    kind: Redis
    name: redis-standalone
    namespace: demo
status:
  backends:
    - name: gcs-backend
      ready: true
      retentionPolicy:
        found: true
        ref:
          name: demo-retention
          namespace: demo
      storage:
        phase: Ready
        ref:
          name: gcs-storage
          namespace: demo
  conditions:
    - lastTransitionTime: "2024-09-18T12:15:07Z"
      message: Validation has been passed successfully.
      reason: ResourceValidationPassed
      status: "True"
      type: ValidationPassed
  dependencies:
    - found: true
      kind: Addon
      name: redis-addon
  phase: Ready
  repositories:
    - name: default-blueprint
      phase: Ready
  sessions:
    - conditions:
        - lastTransitionTime: "2024-09-18T12:15:37Z"
          message: Scheduler has been ensured successfully.
          reason: SchedulerEnsured
          status: "True"
          type: SchedulerEnsured
        - lastTransitionTime: "2024-09-18T12:15:38Z"
          message: Initial backup has been triggered successfully.
          reason: SuccessfullyTriggeredInitialBackup
          status: "True"
          type: InitialBackupTriggered
      name: frequent-backup
  targetFound: true
```

Notice the `spec.backends`, `spec.sessions` and `spec.target` sections, KubeStash automatically resolved those info from the `BackupBluePrint` and created the above `BackupConfiguration`.

**Verify BackupSession:**

KubeStash triggers an instant backup as soon as the `BackupConfiguration` is ready. After that, backups are scheduled according to the specified schedule.

```bash
$ kubectl get backupsession -n demo -w
NAME                                                     INVOKER-TYPE          INVOKER-NAME                  PHASE       DURATION   AGE
appbinding-redis-standalone-frequent-backup-1726661707   BackupConfiguration   appbinding-redis-standalone   Succeeded   2m26s      9m56s
```

We can see from the above output that the backup session has succeeded. Now, we are going to verify whether the backed up data has been stored in the backend.

**Verify Backup:**

Once a backup is complete, KubeStash will update the respective `Repository` CR to reflect the backup. Check that the repository `redis-standalone-backup` has been updated by the following command,

```bash
$ kubectl get repository -n demo default-blueprint
NAME                INTEGRITY   SNAPSHOT-COUNT   SIZE        PHASE   LAST-SUCCESSFUL-BACKUP   AGE
default-blueprint   true        1                1.111 KiB   Ready   3m7s                     13m
```

At this moment we have one `Snapshot`. Run the following command to check the respective `Snapshot` which represents the state of a backup run for an application.

```bash
$ kubectl get snapshots -n demo -l=kubestash.com/repo-name=default-blueprint
NAME                                                              REPOSITORY          SESSION           SNAPSHOT-TIME          DELETION-POLICY   PHASE       AGE
  default-blueprint-appbinding-redlone-frequent-backup-1726661707   default-blueprint   frequent-backup   2024-09-18T12:15:38Z   Delete            Succeeded   14m
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
$ kubectl get snapshots -n demo default-blueprint-appbinding-redlone-frequent-backup-1726661707 -oyaml
```

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: Snapshot
metadata:
  creationTimestamp: "2024-09-18T12:15:38Z"
  finalizers:
    - kubestash.com/cleanup
  generation: 1
  labels:
    kubedb.com/db-version: 7.4.0
    kubestash.com/app-ref-kind: Redis
    kubestash.com/app-ref-name: redis-standalone
    kubestash.com/app-ref-namespace: demo
    kubestash.com/repo-name: default-blueprint
  name: default-blueprint-appbinding-redlone-frequent-backup-1726661707
  namespace: demo
  ownerReferences:
    - apiVersion: storage.kubestash.com/v1alpha1
      blockOwnerDeletion: true
      controller: true
      kind: Repository
      name: default-blueprint
      uid: 4d4085d1-f51d-48b2-95cb-f0e0503bb456
  resourceVersion: "1176748"
  uid: 8b3e77f2-3771-497e-a92c-e43031f84031
spec:
  appRef:
    apiGroup: kubedb.com
    kind: Redis
    name: redis-standalone
    namespace: demo
  backupSession: appbinding-redis-standalone-frequent-backup-1726661707
  deletionPolicy: Delete
  repository: default-blueprint
  session: frequent-backup
  snapshotID: 01J82GYFH8RTZ2GNJT2R9FQFHA
  type: FullBackup
  version: v1
status:
  components:
    dump:
      driver: Restic
      duration: 29.616729384s
      integrity: true
      path: repository/v1/frequent-backup/dump
      phase: Succeeded
      resticStats:
        - hostPath: dumpfile.resp
          id: afce1d29a21d2b05a2aadfb5bdd08f0d5b7c2b2e70fc1d5d77843ebbbef258c1
          size: 184 B
          uploaded: 483 B
      size: 381 B
  conditions:
    - lastTransitionTime: "2024-09-18T12:15:38Z"
      message: Recent snapshot list updated successfully
      reason: SuccessfullyUpdatedRecentSnapshotList
      status: "True"
      type: RecentSnapshotListUpdated
    - lastTransitionTime: "2024-09-18T12:17:59Z"
      message: Metadata uploaded to backend successfully
      reason: SuccessfullyUploadedSnapshotMetadata
      status: "True"
      type: SnapshotMetadataUploaded
  integrity: true
  phase: Succeeded
  size: 381 B
  snapshotTime: "2024-09-18T12:15:38Z"
  totalComponents: 1
```

> KubeStash uses [redis-dump-go](https://github.com/yannh/redis-dump-go) to perform backups of target `Redis` databases. Therefore, the component name for logical backups is set as `dump`.

Now, if we navigate to the GCS bucket, we will see the backed up data stored in the `blueprint/default-blueprint/repository/v1/frequent-backup/dump` directory. KubeStash also keeps the backup for `Snapshot` YAMLs, which can be found in the `blueprint/default-blueprint/snapshots` directory.

> Note: KubeStash stores all dumped data encrypted in the backup directory, meaning it remains unreadable until decrypted.

## Auto-backup with custom configurations

In this section, we are going to backup a `Redis` database of `demo` namespace. We are going to use the custom configurations which will be specified in the `BackupBlueprint` CR.

**Prepare Backup Blueprint**

A `BackupBlueprint` allows you to specify a template for the `Repository`,`Session` or `Variables` of `BackupConfiguration` in a Kubernetes native way.

Now, we have to create a `BackupBlueprint` CR with a blueprint for `BackupConfiguration` object.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupBlueprint
metadata:
  name: redis-customize-backup-blueprint
  namespace: demo
spec:
  usagePolicy:
    allowedNamespaces:
      from: All
  backupConfigurationTemplate:
    deletionPolicy: OnDelete
    backends:
      - name: gcs-backend
        storageRef:
          namespace: demo
          name: gcs-storage
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
            backend: gcs-backend
            directory: ${namespace}/${targetName}
            encryptionSecret:
              name: encrypt-secret
              namespace: demo
        addon:
          name: redis-addon
          tasks:
            - name: logical-backup
```

Note that we have used some variables (format: `${<variable name>}`) in different fields. KubeStash will substitute these variables with values from the respective target’s annotations. You’re free to use any variables you like.

Here,

- `.spec.backupConfigurationTemplate.backends[*].storageRef` refers our earlier created `gcs-storage` backupStorage.
- `.spec.backupConfigurationTemplate.sessions[*]`:
    - `.schedule` defines `${schedule}` variable, which determines the time interval for the backup.
    - `.repositories[*].name` defines the `${repoName}` variable, which specifies the name of the backup `Repository`.
    - `.repositories[*].directory` defines two variables, `${namespace}` and `${targetName}`, which are used to determine the path where the backup will be stored.

Let's create the `BackupBlueprint` we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/redis/backup/kubestash/auto-backup/examples/customize-backupblueprint.yaml
backupblueprint.core.kubestash.com/redis-customize-backup-blueprint created
```

Now, we are ready to backup our `Redis` databases using few annotations. You can check available auto-backup annotations for a databases from [here](https://kubestash.com/docs/latest/concepts/crds/backupblueprint/).

**Create Database**

Now, we are going to create an `Redis` CR in demo namespace.

Below is the YAML of the `Redis` object that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: Redis
metadata:
  name: redis-standalone-2
  namespace: demo
  annotations:
    blueprint.kubestash.com/name: redis-customize-backup-blueprint
    blueprint.kubestash.com/namespace: demo
    variables.kubestash.com/schedule: "*/10 * * * *"
    variables.kubestash.com/repoName: customize-blueprint
    variables.kubestash.com/namespace: demo
    variables.kubestash.com/targetName: redis-standalone-2
spec:
  version: 7.4.0
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: Delete
```

Notice the `metadata.annotations` field, where we have defined the annotations related to the automatic backup configuration. Specifically, we've set the `BackupBlueprint` name as `redis-customize-backup-blueprint` and the namespace as `demo`. We have also provided values for the blueprint template variables, such as the backup `schedule`, `repositoryName`, `namespace`, and `targetName`. These annotations will be used to create a `BackupConfiguration` for this `Redis` database.

Let's create the `Redis` we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/redis/backup/kubestash/auto-backup/examples/redis-standalone-2.yaml
redis.kubedb.com/redis-standalone-2 created
```

**Verify BackupConfiguration**

If everything goes well, KubeStash should create a `BackupConfiguration` for our `Redis` in `demo` namespace and the phase of that `BackupConfiguration` should be `Ready`. Verify the `BackupConfiguration` object by the following command,

```bash
$ kubectl get backupconfiguration -n demo
NAME                            PHASE   PAUSED   AGE
appbinding-redis-standalone-2   Ready            61s
```

Now, let’s check the YAML of the `BackupConfiguration`.

```bash
$ kubectl get backupconfiguration -n demo appbinding-redis-standalone-2  -o yaml
```

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  creationTimestamp: "2024-09-18T13:04:15Z"
  finalizers:
    - kubestash.com/cleanup
  generation: 1
  labels:
    app.kubernetes.io/managed-by: kubestash.com
    kubestash.com/invoker-name: redis-customize-backup-blueprint
    kubestash.com/invoker-namespace: demo
  name: appbinding-redis-standalone-2
  namespace: demo
  resourceVersion: "1181988"
  uid: 83f7a345-6e1b-41e3-8b6d-520e4a1852e5
spec:
  backends:
    - name: gcs-backend
      retentionPolicy:
        name: demo-retention
        namespace: demo
      storageRef:
        name: gcs-storage
        namespace: demo
  sessions:
    - addon:
        name: redis-addon
        tasks:
          - name: logical-backup
      name: frequent-backup
      repositories:
        - backend: gcs-backend
          directory: demo/redis-standalone-2
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
    kind: Redis
    name: redis-standalone-2
    namespace: demo
status:
  backends:
    - name: gcs-backend
      ready: true
      retentionPolicy:
        found: true
        ref:
          name: demo-retention
          namespace: demo
      storage:
        phase: Ready
        ref:
          name: gcs-storage
          namespace: demo
  conditions:
    - lastTransitionTime: "2024-09-18T13:04:15Z"
      message: Validation has been passed successfully.
      reason: ResourceValidationPassed
      status: "True"
      type: ValidationPassed
  dependencies:
    - found: true
      kind: Addon
      name: redis-addon
  phase: Ready
  repositories:
    - name: customize-blueprint
      phase: Ready
  sessions:
    - conditions:
        - lastTransitionTime: "2024-09-18T13:04:35Z"
          message: Scheduler has been ensured successfully.
          reason: SchedulerEnsured
          status: "True"
          type: SchedulerEnsured
        - lastTransitionTime: "2024-09-18T13:04:35Z"
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
NAME                                                       INVOKER-TYPE          INVOKER-NAME                    PHASE       DURATION   AGE
appbinding-redis-standalone-2-frequent-backup-1726664655   BackupConfiguration   appbinding-redis-standalone-2   Succeeded   2m33s      4m16s
```

We can see from the above output that the backup session has succeeded. Now, we are going to verify whether the backed up data has been stored in the backend.

**Verify Backup:**

Once a backup is complete, KubeStash will update the respective `Repository` CR to reflect the backup. Check that the repository `customize-blueprint` has been updated by the following command,

```bash
$ kubectl get repository -n demo customize-blueprint
NAME                  INTEGRITY   SNAPSHOT-COUNT   SIZE    PHASE   LAST-SUCCESSFUL-BACKUP   AGE
customize-blueprint   true        1                380 B   Ready   5m44s                    6m4s
```

At this moment we have one `Snapshot`. Run the following command to check the respective `Snapshot` which represents the state of a backup run for an application.

```bash
$ kubectl get snapshots -n demo -l=kubestash.com/repo-name=customize-blueprint
NAME                                                              REPOSITORY            SESSION           SNAPSHOT-TIME          DELETION-POLICY   PHASE       AGE
customize-blueprint-appbinding-rne-2-frequent-backup-1726664655   customize-blueprint   frequent-backup   2024-09-18T13:04:35Z   Delete            Succeeded   6m7s
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
$ kubectl get snapshots -n demo customize-blueprint-appbinding-rne-2-frequent-backup-1726664655 -oyaml
```

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: Snapshot
metadata:
  creationTimestamp: "2024-09-18T13:04:35Z"
  finalizers:
    - kubestash.com/cleanup
  generation: 1
  labels:
    kubedb.com/db-version: 7.4.0
    kubestash.com/app-ref-kind: Redis
    kubestash.com/app-ref-name: redis-standalone-2
    kubestash.com/app-ref-namespace: demo
    kubestash.com/repo-name: customize-blueprint
  name: customize-blueprint-appbinding-rne-2-frequent-backup-1726664655
  namespace: demo
  ownerReferences:
    - apiVersion: storage.kubestash.com/v1alpha1
      blockOwnerDeletion: true
      controller: true
      kind: Repository
      name: customize-blueprint
      uid: c107da60-af66-4ad6-83cc-d80053a11de3
  resourceVersion: "1182349"
  uid: 93b20b59-abce-41a5-88da-f2ce6e98713d
spec:
  appRef:
    apiGroup: kubedb.com
    kind: Redis
    name: redis-standalone-2
    namespace: demo
  backupSession: appbinding-redis-standalone-2-frequent-backup-1726664655
  deletionPolicy: Delete
  repository: customize-blueprint
  session: frequent-backup
  snapshotID: 01J82KR4C7ZER9ZM0W52TVBEET
  type: FullBackup
  version: v1
status:
  components:
    dump:
      driver: Restic
      duration: 29.378445351s
      integrity: true
      path: repository/v1/frequent-backup/dump
      phase: Succeeded
      resticStats:
        - hostPath: dumpfile.resp
          id: 73cf596a525bcdb439e87812045e7a25c6bd82574513351ab434793c134fc817
          size: 184 B
          uploaded: 483 B
      size: 380 B
  conditions:
    - lastTransitionTime: "2024-09-18T13:04:35Z"
      message: Recent snapshot list updated successfully
      reason: SuccessfullyUpdatedRecentSnapshotList
      status: "True"
      type: RecentSnapshotListUpdated
    - lastTransitionTime: "2024-09-18T13:07:06Z"
      message: Metadata uploaded to backend successfully
      reason: SuccessfullyUploadedSnapshotMetadata
      status: "True"
      type: SnapshotMetadataUploaded
  integrity: true
  phase: Succeeded
  size: 380 B
  snapshotTime: "2024-09-18T13:04:35Z"
  totalComponents: 1
```

> KubeStash uses [redis-dump-go](https://github.com/yannh/redis-dump-go) to perform backups of target `Redis` databases. Therefore, the component name for logical backups is set as `dump`.

Now, if we navigate to the GCS bucket, we will see the backed up data stored in the `blueprint/demo/redis-standalone-2/repository/v1/frequent-backup/dump` directory. KubeStash also keeps the backup for `Snapshot` YAMLs, which can be found in the `blueprint/demo/redis-standalone-2/snapshots` directory.

> Note: KubeStash stores all dumped data encrypted in the backup directory, meaning it remains unreadable until decrypted.

## Cleanup

To cleanup the resources crated by this tutorial, run the following commands,

```bash
kubectl delete backupblueprints.core.kubestash.com  -n demo redis-default-backup-blueprint
kubectl delete backupblueprints.core.kubestash.com  -n demo redis-customize-backup-blueprint
kubectl delete retentionpolicies.storage.kubestash.com -n demo demo-retention
kubectl delete backupstorage -n demo gcs-storage
kubectl delete secret -n demo gcs-secret
kubectl delete secret -n demo encrypt-secret
kubectl delete redis -n demo redis-standalone
kubectl delete redis -n demo redis-standalone-2
```