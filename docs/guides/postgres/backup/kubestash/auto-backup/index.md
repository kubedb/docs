---
title: PostgreSQL Auto-Backup | KubeStash
description: Backup PostgreSQL using KubeStash Auto-Backup
menu:
  docs_{{ .version }}:
    identifier: guides-pg-auto-backup-stashv2
    name: Auto-Backup
    parent: guides-pg-backup-stashv2
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Backup PostgreSQL using KubeStash Auto-Backup

KubeStash can automatically be configured to backup any `PostgreSQL` databases in your cluster. KubeStash enables cluster administrators to deploy backup `blueprints` ahead of time so database owners can easily backup any `PostgreSQL` database with a few annotations.

In this tutorial, we are going to show how you can configure a backup blueprint for `PostgreSQL` databases in your cluster and backup them with a few annotations.


## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using `Minikube` or `Kind`.
- Install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md).
- Install `KubeStash` in your cluster following the steps [here](https://kubestash.com/docs/latest/setup/install/kubestash).
- Install KubeStash `kubectl` plugin following the steps [here](https://kubestash.com/docs/latest/setup/install/kubectl-plugin/).
- If you are not familiar with how KubeStash backup and restore `PostgreSQL` databases, please check the following guide [here](/docs/guides/postgres/backup/kubestash/overview/index.md).

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

> **Note:** YAML files used in this tutorial are stored in [docs/guides/postgres/backup/kubestash/auto-backup/examples](https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/backup/kubestash/auto-backup/examples) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.


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
      bucket: kubestash-qa
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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/backup/kubestash/auto-backup/examples/backupstorage.yaml
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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/backup/kubestash/auto-backup/examples/retentionpolicy.yaml
retentionpolicy.storage.kubestash.com/demo-retention created
```

**Create Secret:**

We also need to create a secret with a `Restic` password for backup data encryption.

Let's create a secret called `encrypt-secret` with the Restic password,

```bash
$ echo -n 'changeit' > RESTIC_PASSWORD
$ kubectl create secret generic -n demo encrypt-secret \
    --from-file=./RESTIC_PASSWORD 
secret "encrypt-secret" created
```

## Auto-backup with default configurations

In this section, we are going to backup a `PostgreSQL` database of `demo` namespace. We are going to use the default configurations which will be specified in the `BackupBlueprint` CR.

**Prepare Backup Blueprint**

A `BackupBlueprint` allows you to specify a template for the `Repository`,`Session` or `Variables` of `BackupConfiguration` in a Kubernetes native way.

Now, we have to create a `BackupBlueprint` CR with a blueprint for `BackupConfiguration` object.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupBlueprint
metadata:
  name: postgres-default-backup-blueprint
  namespace: demo
spec:
  usagePolicy:
    allowedNamespaces:
      from: All
  backupConfigurationTemplate:
    deletionPolicy: OnDelete
    # ============== Blueprint for Backends of BackupConfiguration  =================
    backends:
      - name: gcs-backend
        storageRef:
          namespace: demo
          name: gcs-storage
        retentionPolicy:
          name: demo-retention
          namespace: demo
    # ============== Blueprint for Sessions of BackupConfiguration  =================
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
          name: postgres-addon
          tasks:
            - name: logical-backup
```

Here,

- `.spec.backupConfigurationTemplate.backends[*].storageRef` refers our earlier created `gcs-storage` backupStorage.
- `.spec.backupConfigurationTemplate.sessions[*].schedule` specifies that we want to backup the database at `5 minutes` interval.

Let's create the `BackupBlueprint` we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/backup/kubestash/auto-backup/examples/default-backupblueprint.yaml
backupblueprint.core.kubestash.com/postgres-default-backup-blueprint created
```

Now, we are ready to backup our `PostgreSQL` databases using few annotations.

**Create Database**

Now, we are going to create an `PostgreSQL` CR in demo namespace. 

Below is the YAML of the `PostgreSQL` object that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: sample-postgres
  namespace: demo
  annotations:
    blueprint.kubestash.com/name: postgres-default-backup-blueprint
    blueprint.kubestash.com/namespace: demo
spec:
  version: "16.1"
  replicas: 3
  standbyMode: Hot
  streamingMode: Synchronous
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Here,

- `.spec.annotations.blueprint.kubestash.com/name: postgres-default-backup-blueprint` specifies the name of the `BackupBlueprint` that will use in backup.
- `.spec.annotations.blueprint.kubestash.com/namespace: demo` specifies the name of the `namespace` where the `BackupBlueprint` resides.

Let's create the `PostgreSQL` we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/backup/kubestash/auto-backup/examples/sample-postgres.yaml
postgres.kubedb.com/sample-postgres created
```

**Verify BackupConfiguration**

If everything goes well, KubeStash should create a `BackupConfiguration` for our PostgreSQL in demo namespace and the phase of that `BackupConfiguration` should be `Ready`. Verify the `BackupConfiguration` object by the following command,

```bash
$ kubectl get backupconfiguration -n demo
NAME                         PHASE   PAUSED   AGE
appbinding-sample-postgres   Ready            2m50m
```

Now, let’s check the YAML of the `BackupConfiguration`.

```bash
$ kubectl get backupconfiguration -n demo appbinding-sample-postgres  -o yaml
```

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  creationTimestamp: "2024-09-05T10:53:48Z"
  finalizers:
    - kubestash.com/cleanup
  generation: 1
  labels:
    app.kubernetes.io/managed-by: kubestash.com
    kubestash.com/invoker-name: postgres-default-backup-blueprint
    kubestash.com/invoker-namespace: demo
  name: appbinding-sample-postgres
  namespace: demo
  resourceVersion: "298502"
  uid: b6537c60-051f-4348-9ca4-c28f3880dbc1
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
        name: postgres-addon
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
    kind: Postgres
    name: sample-postgres
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
    - lastTransitionTime: "2024-09-05T10:53:48Z"
      message: Validation has been passed successfully.
      reason: ResourceValidationPassed
      status: "True"
      type: ValidationPassed
  dependencies:
    - found: true
      kind: Addon
      name: postgres-addon
  phase: Ready
  repositories:
    - name: default-blueprint
      phase: Ready
  sessions:
    - conditions:
        - lastTransitionTime: "2024-09-05T10:53:59Z"
          message: Scheduler has been ensured successfully.
          reason: SchedulerEnsured
          status: "True"
          type: SchedulerEnsured
        - lastTransitionTime: "2024-09-05T10:53:59Z"
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
NAME                                                    INVOKER-TYPE          INVOKER-NAME                 PHASE       DURATION   AGE
appbinding-sample-postgres-frequent-backup-1725533628   BackupConfiguration   appbinding-sample-postgres   Succeeded   23s        6m40s
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
default-blueprint-appbinding-samgres-frequent-backup-1725533628   default-blueprint   frequent-backup   2024-09-05T10:53:59Z   Delete            Succeeded   7m48s
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
$ kubectl get snapshots -n demo default-blueprint-appbinding-samgres-frequent-backup-1725533628 -oyaml
```

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: Snapshot
metadata:
  creationTimestamp: "2024-09-05T10:53:59Z"
  finalizers:
    - kubestash.com/cleanup
  generation: 1
  labels:
    kubestash.com/app-ref-kind: Postgres
    kubestash.com/app-ref-name: sample-postgres
    kubestash.com/app-ref-namespace: demo
    kubestash.com/repo-name: default-blueprint
  annotations:
    kubedb.com/db-version: "16.1"
  name: default-blueprint-appbinding-samgres-frequent-backup-1725533628
  namespace: demo
  ownerReferences:
    - apiVersion: storage.kubestash.com/v1alpha1
      blockOwnerDeletion: true
      controller: true
      kind: Repository
      name: default-blueprint
      uid: 1125a82f-2bd8-4029-aae6-078ff5413383
  resourceVersion: "298559"
  uid: c179b758-6ba4-4a32-81f1-fa41ba3dc527
spec:
  appRef:
    apiGroup: kubedb.com
    kind: Postgres
    name: sample-postgres
    namespace: demo
  backupSession: appbinding-sample-postgres-frequent-backup-1725533628
  deletionPolicy: Delete
  repository: default-blueprint
  session: frequent-backup
  snapshotID: 01J70X3MGSYT4TJK77R8YXEV3T
  type: FullBackup
  version: v1
status:
  components:
    dump:
      driver: Restic
      duration: 5.952466363s
      integrity: true
      path: repository/v1/frequent-backup/dump
      phase: Succeeded
      resticStats:
        - hostPath: dumpfile.sql
          id: a30f8ec138e24cbdbcce088a73e5b9d73a58750c38793ef05ff7d570148ddd2c
          size: 3.345 KiB
          uploaded: 3.637 KiB
      size: 1.132 KiB
  conditions:
    - lastTransitionTime: "2024-09-05T10:53:59Z"
      message: Recent snapshot list updated successfully
      reason: SuccessfullyUpdatedRecentSnapshotList
      status: "True"
      type: RecentSnapshotListUpdated
    - lastTransitionTime: "2024-09-05T10:54:20Z"
      message: Metadata uploaded to backend successfully
      reason: SuccessfullyUploadedSnapshotMetadata
      status: "True"
      type: SnapshotMetadataUploaded
  integrity: true
  phase: Succeeded
  size: 1.132 KiB
  snapshotTime: "2024-09-05T10:53:59Z"
  totalComponents: 1
```

> KubeStash uses `pg_dump` or `pg_dumpall` to perform backups of target `PostgreSQL` databases. Therefore, the component name for logical backups is set as `dump`.

Now, if we navigate to the GCS bucket, we will see the backed up data stored in the `blueprint/default-blueprint/repository/v1/frequent-backup/dump` directory. KubeStash also keeps the backup for `Snapshot` YAMLs, which can be found in the `blueprint/default-blueprint/snapshots` directory.

> Note: KubeStash stores all dumped data encrypted in the backup directory, meaning it remains unreadable until decrypted.


## Auto-backup with custom configurations

In this section, we are going to backup a `PostgreSQL` database of `demo` namespace. We are going to use the custom configurations which will be specified in the `BackupBlueprint` CR.

**Prepare Backup Blueprint**

A `BackupBlueprint` allows you to specify a template for the `Repository`,`Session` or `Variables` of `BackupConfiguration` in a Kubernetes native way.

Now, we have to create a `BackupBlueprint` CR with a blueprint for `BackupConfiguration` object.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupBlueprint
metadata:
  name: postgres-customize-backup-blueprint
  namespace: demo
spec:
  usagePolicy:
    allowedNamespaces:
      from: All
  backupConfigurationTemplate:
    deletionPolicy: OnDelete
    # ============== Blueprint for Backends of BackupConfiguration  =================
    backends:
      - name: gcs-backend
        storageRef:
          namespace: demo
          name: gcs-storage
        retentionPolicy:
          name: demo-retention
          namespace: demo
    # ============== Blueprint for Sessions of BackupConfiguration  =================
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
          name: postgres-addon
          tasks:
            - name: logical-backup
              params:
                backupCmd: pg_dump
                args: ${targetedDatabase}
```

Note that we have used some variables (format: `${<variable name>}`) in different fields. KubeStash will substitute these variables with values from the respective target’s annotations. You’re free to use any variables you like.

Here,

- `.spec.backupConfigurationTemplate.backends[*].storageRef` refers our earlier created `gcs-storage` backupStorage.
- `.spec.backupConfigurationTemplate.sessions[*]`:
    - `.schedule` defines `${schedule}` variable, which determines the time interval for the backup.
    - `.repositories[*].name` defines the `${repoName}` variable, which specifies the name of the backup `Repository`.
    - `.repositories[*].directory` defines two variables, `${namespace}` and `${targetName}`, which are used to determine the path where the backup will be stored.
    - `.addon.tasks[*].params.args` defines `${targetedDatabase}` variable, which identifies a single database to backup.

Let's create the `BackupBlueprint` we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/backup/kubestash/auto-backup/examples/customize-backupblueprint.yaml
backupblueprint.core.kubestash.com/postgres-customize-backup-blueprint created
```

Now, we are ready to backup our `PostgreSQL` databases using few annotations. You can check available auto-backup annotations for a databases from [here](https://kubestash.com/docs/latest/concepts/crds/backupblueprint/).


**Create Database**

Now, we are going to create an `PostgreSQL` CR in demo namespace.

Below is the YAML of the `PostgreSQL` object that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: Postgres
metadata:
  name: sample-postgres-2
  namespace: demo
  annotations:
    blueprint.kubestash.com/name: postgres-customize-backup-blueprint
    blueprint.kubestash.com/namespace: demo
    variables.kubestash.com/schedule: "*/10 * * * *"
    variables.kubestash.com/repoName: customize-blueprint
    variables.kubestash.com/namespace: demo
    variables.kubestash.com/targetName: sample-postgres-2
    variables.kubestash.com/targetedDatabase: postgres
spec:
  version: "16.1"
  replicas: 3
  standbyMode: Hot
  streamingMode: Synchronous
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Notice the `metadata.annotations` field, where we have defined the annotations related to the automatic backup configuration. Specifically, we've set the `BackupBlueprint` name as `postgres-customize-backup-blueprint` and the namespace as `demo`. We have also provided values for the blueprint template variables, such as the backup `schedule`, `repositoryName`, `namespace`, `targetName`, and `targetedDatabase`. These annotations will be used to create a `BackupConfiguration` for this `postgreSQL` database.

Let's create the `PostgreSQL` we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/postgres/backup/kubestash/auto-backup/examples/sample-postgres-2.yaml
postgres.kubedb.com/sample-postgres-2 created
```

**Verify BackupConfiguration**

If everything goes well, KubeStash should create a `BackupConfiguration` for our PostgreSQL in demo namespace and the phase of that `BackupConfiguration` should be `Ready`. Verify the `BackupConfiguration` object by the following command,

```bash
$ kubectl get backupconfiguration -n demo
NAME                           PHASE   PAUSED      AGE
appbinding-sample-postgres-2   Ready               2m50m
```

Now, let’s check the YAML of the `BackupConfiguration`.

```bash
$ kubectl get backupconfiguration -n demo appbinding-sample-postgres-2  -o yaml
```

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  creationTimestamp: "2024-09-05T12:39:37Z"
  finalizers:
    - kubestash.com/cleanup
  generation: 1
  labels:
    app.kubernetes.io/managed-by: kubestash.com
    kubestash.com/invoker-name: postgres-customize-backup-blueprint
    kubestash.com/invoker-namespace: demo
  name: appbinding-sample-postgres-2
  namespace: demo
  resourceVersion: "309511"
  uid: b4091166-2813-4183-acda-e2c80eaedbb5
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
        name: postgres-addon
        tasks:
          - name: logical-backup
            params:
              args: postgres
              backupCmd: pg_dump
      name: frequent-backup
      repositories:
        - backend: gcs-backend
          directory: demo/sample-postgres-2
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
    kind: Postgres
    name: sample-postgres-2
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
    - lastTransitionTime: "2024-09-05T12:39:37Z"
      message: Validation has been passed successfully.
      reason: ResourceValidationPassed
      status: "True"
      type: ValidationPassed
  dependencies:
    - found: true
      kind: Addon
      name: postgres-addon
  phase: Ready
  repositories:
    - name: customize-blueprint
      phase: Ready
  sessions:
    - conditions:
        - lastTransitionTime: "2024-09-05T12:39:37Z"
          message: Scheduler has been ensured successfully.
          reason: SchedulerEnsured
          status: "True"
          type: SchedulerEnsured
        - lastTransitionTime: "2024-09-05T12:39:37Z"
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
NAME                                                      INVOKER-TYPE          INVOKER-NAME                   PHASE       DURATION   AGE
appbinding-sample-postgres-frequent-backup-1725597000     BackupConfiguration   appbinding-sample-postgres     Succeeded   58s        112s
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
customize-blueprint-appbinding-ses-2-frequent-backup-1725597000   customize-blueprint   frequent-backup   2024-09-06T04:30:00Z   Delete            Succeeded   6m19s
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
  creationTimestamp: "2024-09-06T04:30:00Z"
  finalizers:
    - kubestash.com/cleanup
  generation: 1
  labels:
    kubedb.com/db-version: "16.1"
    kubestash.com/app-ref-kind: Postgres
    kubestash.com/app-ref-name: sample-postgres-2
    kubestash.com/app-ref-namespace: demo
    kubestash.com/repo-name: customize-blueprint
  name: customize-blueprint-appbinding-ses-2-frequent-backup-1725597000
  namespace: demo
  ownerReferences:
    - apiVersion: storage.kubestash.com/v1alpha1
      blockOwnerDeletion: true
      controller: true
      kind: Repository
      name: customize-blueprint
      uid: 5d4618c5-c28a-456a-9854-f6447161d3d1
  resourceVersion: "315624"
  uid: 7e02a18c-c8a7-40be-bd22-e7312678d6f7
spec:
  appRef:
    apiGroup: kubedb.com
    kind: Postgres
    name: sample-postgres-2
    namespace: demo
  backupSession: appbinding-sample-postgres-2-frequent-backup-1725597000
  deletionPolicy: Delete
  repository: customize-blueprint
  session: frequent-backup
  snapshotID: 01J72SH8XPEHB6SYNXFE00V5PB
  type: FullBackup
  version: v1
status:
  components:
    dump:
      driver: Restic
      duration: 7.060169632s
      integrity: true
      path: repository/v1/frequent-backup/dump
      phase: Succeeded
      resticStats:
        - hostPath: dumpfile.sql
          id: 74d82943e0d676321e989edb503f5e2d6fe5cf4f4be72d386e492ec533358c26
          size: 1.220 KiB
          uploaded: 296 B
      size: 1.873 KiB
  conditions:
    - lastTransitionTime: "2024-09-06T04:30:00Z"
      message: Recent snapshot list updated successfully
      reason: SuccessfullyUpdatedRecentSnapshotList
      status: "True"
      type: RecentSnapshotListUpdated
    - lastTransitionTime: "2024-09-06T04:30:38Z"
      message: Metadata uploaded to backend successfully
      reason: SuccessfullyUploadedSnapshotMetadata
      status: "True"
      type: SnapshotMetadataUploaded
  integrity: true
  phase: Succeeded
  size: 1.872 KiB
  snapshotTime: "2024-09-06T04:30:00Z"
  totalComponents: 1
```

> KubeStash uses `pg_dump` or `pg_dumpall` to perform backups of target `PostgreSQL` databases. Therefore, the component name for logical backups is set as `dump`.

Now, if we navigate to the GCS bucket, we will see the backed up data stored in the `blueprint/demo/sample-postgres-2/repository/v1/frequent-backup/dump` directory. KubeStash also keeps the backup for `Snapshot` YAMLs, which can be found in the `blueprint/demo/sample-postgres-2/snapshots` directory.

> Note: KubeStash stores all dumped data encrypted in the backup directory, meaning it remains unreadable until decrypted.

## Cleanup

To cleanup the resources crated by this tutorial, run the following commands,

```bash
kubectl delete backupblueprints.core.kubestash.com  -n demo postgres-default-backup-blueprint
kubectl delete backupblueprints.core.kubestash.com  -n demo postgres-customize-backup-blueprint
kubectl delete retentionpolicies.storage.kubestash.com -n demo demo-retention
kubectl delete backupstorage -n demo gcs-storage
kubectl delete secret -n demo gcs-secret
kubectl delete secret -n demo encrypt-secret
kubectl delete postgres -n demo sample-postgres
kubectl delete postgres -n demo sample-postgres-2
```