---
title: Elasticsearch Auto-Backup | KubeStash
description: Backup Elasticsearch using KubeStash Auto-Backup
menu:
  docs_{{ .version }}:
    identifier: guides-es-auto-backup-stashv2
    name: Auto-Backup
    parent: guides-es-backup-stashv2
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Backup Elasticsearch using KubeStash Auto-Backup

KubeStash can automatically be configured to backup any `Elasticsearch` databases in your cluster. KubeStash enables cluster administrators to deploy backup `blueprints` ahead of time so database owners can easily backup any `Elasticsearch` database with a few annotations.

In this tutorial, we are going to show how you can configure a backup blueprint for `Elasticsearch` databases in your cluster and backup them with a few annotations.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using `Minikube` or `Kind`.
- Install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md).
- Install `KubeStash` in your cluster following the steps [here](https://kubestash.com/docs/latest/setup/install/kubestash).
- Install KubeStash `kubectl` plugin following the steps [here](https://kubestash.com/docs/latest/setup/install/kubectl-plugin/).
- If you are not familiar with how KubeStash backup and restore `Elasticsearch` databases, please check the following guide [here](/docs/guides/elasticsearch/backup/kubestash/overview/index.md).

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

> **Note:** YAML files used in this tutorial are stored in [docs/guides/elasticsearch/backup/kubestash/auto-backup/examples](https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/elasticsearch/backup/kubestash/auto-backup/examples) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

### Prepare Backend

We are going to store our backed up data into a `S3` bucket. We have to create a `Secret` with necessary credentials and a `BackupStorage` CR to use this backend. If you want to use a different backend, please read the respective backend configuration doc from [here](https://kubestash.com/docs/latest/guides/backends/overview/).

**Create Secret:**

Let's create a secret called `gcs-secret` with access credentials to our desired GCS bucket,

```bash
$ echo -n '<your-access-key>' > AWS_ACCESS_KEY_ID
$ echo -n '<your-secret-key>' > AWS_SECRET_ACCESS_KEY
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
      endpoint: us-east-1.linodeobjects.com
      bucket: esbackup
      region: us-east-1
      prefix: elastic
      secretName: s3-secret
  usagePolicy:
    allowedNamespaces:
      from: All
  default: true
  deletionPolicy: Delete
```

Let's create the BackupStorage we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/elasticsearch/backup/kubestash/auto-backup/examples/backupstorage.yaml
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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/elasticsearch/backup/kubestash/auto-backup/examples/retentionpolicy.yaml
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

In this section, we are going to backup a `Elasticsearch` database of `demo` namespace. We are going to use the default configurations which will be specified in the `BackupBlueprint` CR.

**Prepare Backup Blueprint**

A `BackupBlueprint` allows you to specify a template for the `Repository`,`Session` or `Variables` of `BackupConfiguration` in a Kubernetes native way.

Now, we have to create a `BackupBlueprint` CR with a blueprint for `BackupConfiguration` object.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupBlueprint
metadata:
  name: es-quickstart-backup-blueprint
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
        scheduler:
          schedule: "*/5 * * * *"
          jobTemplate:
            backoffLimit: 1
        repositories:
          - name: s3-elasticsearch-repo
            backend: s3-backend
            directory: /es
            encryptionSecret:
              name: encrypt-secret
              namespace: demo
        addon:
          name: elasticsearch-addon
          tasks:
            - name: logical-backup
```

Here,

- `.spec.backupConfigurationTemplate.backends[*].storageRef` refers our earlier created `s3-storage` backupStorage.
- `.spec.backupConfigurationTemplate.sessions[*].schedule` specifies that we want to backup the database at `5 minutes` interval.

Let's create the `BackupBlueprint` we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/elasticsearch/backup/kubestash/auto-backup/examples/default-backupblueprint.yaml
backupblueprint.core.kubestash.com/es-quickstart-backup-blueprint created
```

Now, we are ready to backup our `Elasticsearch` databases using few annotations.

**Create Database**

Now, we are going to create an `Elasticsearch` CR in demo namespace.

Below is the YAML of the `Elasticsearch` object that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: Elasticsearch
metadata:
  name: es-quickstart
  namespace: demo
  annotations:
    blueprint.kubestash.com/name: es-quickstart-backup-blueprint
    blueprint.kubestash.com/namespace: demo
spec:
  version: xpack-8.15.0
  enableSSL: true
  replicas: 2
  storageType: Durable
  storage:
    storageClassName: standard
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: Delete
```

Here,

- `.spec.annotations.blueprint.kubestash.com/name: es-quickstart-backup-blueprint` specifies the name of the `BackupBlueprint` that will use in backup.
- `.spec.annotations.blueprint.kubestash.com/namespace: demo` specifies the name of the `namespace` where the `BackupBlueprint` resides.

Let's create the `Elasticsearch` we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/elasticsearch/backup/kubestash/auto-backup/examples/sample-es.yaml
elasticsearch.kubedb.com/es-quickstart created
```

**Verify BackupConfiguration**

If everything goes well, KubeStash should create a `BackupConfiguration` for our Elasticsearch in demo namespace and the phase of that `BackupConfiguration` should be `Ready`. Verify the `BackupConfiguration` object by the following command,

```bash
$ kubectl get backupconfiguration -n demo
NAME                         PHASE   PAUSED   AGE
appbinding-es-quickstart     Ready            2m50m
```

Now, let’s check the YAML of the `BackupConfiguration`.

```bash
$ kubectl get backupconfiguration -n demo appbinding-es-quickstart -oyaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  creationTimestamp: "2024-09-19T04:49:38Z"
  finalizers:
  - kubestash.com/cleanup
  generation: 1
  labels:
    app.kubernetes.io/managed-by: kubestash.com
    kubestash.com/invoker-name: es-quickstart-backup-blueprint
    kubestash.com/invoker-namespace: demo
  name: appbinding-es-quickstart
  namespace: demo
  resourceVersion: "80802"
  uid: 1cb6aaf2-b949-4b27-8a29-0d711b88b7e4
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
      name: elasticsearch-addon
      tasks:
      - name: logical-backup
    name: frequent-backup
    repositories:
    - backend: s3-backend
      directory: /es
      encryptionSecret:
        name: encrypt-secret
        namespace: demo
      name: s3-elasticsearch-repo
    scheduler:
      jobTemplate:
        backoffLimit: 1
        template:
          controller: {}
          metadata: {}
          spec:
            resources: {}
      schedule: '*/5 * * * *'
    sessionHistoryLimit: 1
  target:
    apiGroup: kubedb.com
    kind: Elasticsearch
    name: es-quickstart
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
  - lastTransitionTime: "2024-09-19T04:49:38Z"
    message: Validation has been passed successfully.
    reason: ResourceValidationPassed
    status: "True"
    type: ValidationPassed
  dependencies:
  - found: true
    kind: Addon
    name: elasticsearch-addon
  phase: Ready
  repositories:
  - name: s3-elasticsearch-repo
    phase: Ready
  sessions:
  - conditions:
    - lastTransitionTime: "2024-09-19T04:49:48Z"
      message: Scheduler has been ensured successfully.
      reason: SchedulerEnsured
      status: "True"
      type: SchedulerEnsured
    - lastTransitionTime: "2024-09-19T04:49:48Z"
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
$ kubectl get backupsession -n demo
NAME                                                  INVOKER-TYPE          INVOKER-NAME               PHASE       DURATION   AGE
appbinding-es-quickstart-frequent-backup-1726722240   BackupConfiguration   appbinding-es-quickstart   Running                12s
```

We can see from the above output that the backup session has succeeded. Now, we are going to verify whether the backed up data has been stored in the backend.

**Verify Backup:**

Once a backup is complete, KubeStash will update the respective `Repository` CR to reflect the backup. Check that the repository `s3-elasticsearch-repo` has been updated by the following command,

```bash
$ kubectl get repo -n demo
NAME                    INTEGRITY   SNAPSHOT-COUNT   SIZE        PHASE   LAST-SUCCESSFUL-BACKUP   AGE
s3-elasticsearch-repo   true        8                6.836 KiB   Ready   64s                      15m
```

At this moment we have one `Snapshot`. Run the following command to check the respective `Snapshot` which represents the state of a backup run for an application.

```bash
kubectl get snapshots -n demo -l=kubestash.com/repo-name=s3-elasticsearch-repo
NAME                                                              REPOSITORY              SESSION           SNAPSHOT-TIME          DELETION-POLICY   PHASE       AGE
s3-elasticsearch-repo-appbindingtart-frequent-backup-1726722361   s3-elasticsearch-repo   frequent-backup   2024-09-19T05:06:01Z   Delete            Succeeded   45s

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
$ kubectl get snapshots -n demo s3-elasticsearch-repo-appbindingtart-frequent-backup-1726722361 -oyaml
apiVersion: storage.kubestash.com/v1alpha1
kind: Snapshot
metadata:
  creationTimestamp: "2024-09-19T05:06:01Z"
  finalizers:
  - kubestash.com/cleanup
  generation: 1
  labels:
    kubedb.com/db-version: 8.15.0
    kubestash.com/app-ref-kind: Elasticsearch
    kubestash.com/app-ref-name: es-quickstart
    kubestash.com/app-ref-namespace: demo
    kubestash.com/repo-name: s3-elasticsearch-repo
  name: s3-elasticsearch-repo-appbindingtart-frequent-backup-1726722361
  namespace: demo
  ownerReferences:
  - apiVersion: storage.kubestash.com/v1alpha1
    blockOwnerDeletion: true
    controller: true
    kind: Repository
    name: s3-elasticsearch-repo
    uid: c6859c35-2c70-45b7-a8ed-e9969b009b69
  resourceVersion: "82707"
  uid: a38da3f9-d1fd-4e07-bb05-cc7f4bc19bf6
spec:
  appRef:
    apiGroup: kubedb.com
    kind: Elasticsearch
    name: es-quickstart
    namespace: demo
  backupSession: appbinding-es-quickstart-frequent-backup-1726722361
  deletionPolicy: Delete
  repository: s3-elasticsearch-repo
  session: frequent-backup
  snapshotID: 01J84ARHWBR8PFEVENBRKEA9PD
  type: FullBackup
  version: v1
status:
  components:
    dump:
      driver: Restic
      duration: 1.933950652s
      integrity: true
      path: repository/v1/frequent-backup/dump
      phase: Succeeded
      resticStats:
      - hostPath: /kubestash-interim/data
        id: 147fa51e71e523631e74ba3195499995696c6ac69560e1c7f4ab1b4222a97a73
        size: 509 B
        uploaded: 2.141 KiB
      size: 7.791 KiB
  conditions:
  - lastTransitionTime: "2024-09-19T05:06:01Z"
    message: Recent snapshot list updated successfully
    reason: SuccessfullyUpdatedRecentSnapshotList
    status: "True"
    type: RecentSnapshotListUpdated
  - lastTransitionTime: "2024-09-19T05:06:12Z"
    message: Metadata uploaded to backend successfully
    reason: SuccessfullyUploadedSnapshotMetadata
    status: "True"
    type: SnapshotMetadataUploaded
  integrity: true
  phase: Succeeded
  size: 7.790 KiB
  snapshotTime: "2024-09-19T05:06:01Z"
  totalComponents: 1
```

> KubeStash uses `multielasticdump` to perform backups of target `Elasticsearch` databases. Therefore, the component name for logical backups is set as `dump`.

Now, if we navigate to the S3 bucket, we will see the backed up data stored in the `elastic/es/default/repository/v1/frequent-backup/dump` directory. KubeStash also keeps the backup for `Snapshot` YAMLs, which can be found in the `elastic/es/defaultsnapshots` directory.

> Note: KubeStash stores all dumped data encrypted in the backup directory, meaning it remains unreadable until decrypted.

## Auto-backup with custom configurations

In this section, we are going to backup a `Elasticsearch` database of `demo` namespace. We are going to use the custom configurations which will be specified in the `BackupBlueprint` CR.

**Prepare Backup Blueprint**

A `BackupBlueprint` allows you to specify a template for the `Repository`,`Session` or `Variables` of `BackupConfiguration` in a Kubernetes native way.

Now, we have to create a `BackupBlueprint` CR with a blueprint for `BackupConfiguration` object.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupBlueprint
metadata:
  name: es-quickstart-custom-backup-blueprint
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
          namespace: ${namespace}
          name: s3-store
        retentionPolicy:
          name: demo-retention
          namespace: ${namespace}
    sessions:
      - name: frequent-backup
        scheduler:
          schedule: ${schedule}
          jobTemplate:
            backoffLimit: 1
        repositories:
          - name: ${repoName}
            backend: s3-backend
            directory: /ess
            encryptionSecret:
              name: encrypt-secret
              namespace: demo
        addon:
          name: elasticsearch-addon
          tasks:
            - name: logical-backup
              params:
                args: ${args}
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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/elasticsearch/backup/kubestash/auto-backup/examples/custom-backup-blueprint.yaml
backupblueprint.core.kubestash.com/es-quickstart-custom-backup-blueprint created
```

Now, we are ready to backup our `Elasticsearch` databases using few annotations. You can check available auto-backup annotations for a databases from [here](https://kubestash.com/docs/latest/concepts/crds/backupblueprint/).

**Create Database**

Now, we are going to create an `Elasticsearch` CR in demo namespace.

Below is the YAML of the `Elasticsearch` object that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: Elasticsearch
metadata:
  name: es-quickstart-2
  namespace: demo
  annotations:
    blueprint.kubestash.com/name: es-quickstart-custom-backup-blueprint
    blueprint.kubestash.com/namespace: demo
    variables.kubestash.com/schedule: "*/5 * * * *"
    variables.kubestash.com/repoName: s3-elasticsearch-repo
    variables.kubestash.com/namespace: demo
    variables.kubestash.com/args: --ignoreType=template,settings
spec:
  version: xpack-8.15.0
  enableSSL: true
  replicas: 2
  storageType: Durable
  storage:
    storageClassName: standard
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: Delete
```

Notice the `metadata.annotations` field, where we have defined the annotations related to the automatic backup configuration. Specifically, we've set the `BackupBlueprint` name as `es-quickstart-custom-backup-blueprint` and the namespace as `demo`. We have also provided values for the blueprint template variables, such as the backup `schedule`, `repositoryName`, `namespace`, `targetName`, and `targetedDatabase`. These annotations will be used to create a `BackupConfiguration` for this `Elasticsearch` database.

Let's create the `Elasticsearch` we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/elasticsearch/backup/kubestash/auto-backup/examples/sample-es-2.yaml
elasticsearch.kubedb.com/es-quickstart-2 created
```

**Verify BackupConfiguration**

If everything goes well, KubeStash should create a `BackupConfiguration` for our Elasticsearch in demo namespace and the phase of that `BackupConfiguration` should be `Ready`. Verify the `BackupConfiguration` object by the following command,

```bash
kubectl get backupconfiguration -n demo
NAME                       PHASE   PAUSED   AGE
appbinding-es-quickstart-2   Ready            8s
```

Now, let’s check the YAML of the `BackupConfiguration`.

```bash
$ kubectl get bacupconfiguration -n demo appbinding-es-quickstart-2 -oyaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  creationTimestamp: "2024-09-19T06:15:53Z"
  finalizers:
  - kubestash.com/cleanup
  generation: 1
  labels:
    app.kubernetes.io/managed-by: kubestash.com
    kubestash.com/invoker-name: es-quickstart-custom-backup-blueprint
    kubestash.com/invoker-namespace: demo
  name: appbinding-es-quickstart
  namespace: demo
  resourceVersion: "87411"
  uid: 23e39d2e-03ab-42cb-9380-d3a371ae4d84
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
      name: elasticsearch-addon
      tasks:
      - name: logical-backup
        params:
          args: --ignoreType=template,settings
    name: frequent-backup
    repositories:
    - backend: s3-backend
      directory: /es
      encryptionSecret:
        name: encrypt-secret
        namespace: demo
      name: s3-elasticsearch-repo
    scheduler:
      jobTemplate:
        backoffLimit: 1
        template:
          controller: {}
          metadata: {}
          spec:
            resources: {}
      schedule: '*/5 * * * *'
    sessionHistoryLimit: 1
  target:
    apiGroup: kubedb.com
    kind: Elasticsearch
    name: es-quickstart-2
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
  - lastTransitionTime: "2024-09-19T06:15:53Z"
    message: Validation has been passed successfully.
    reason: ResourceValidationPassed
    status: "True"
    type: ValidationPassed
  dependencies:
  - found: true
    kind: Addon
    name: elasticsearch-addon
  phase: Ready
  repositories:
  - name: s3-elasticsearch-repo
    phase: Ready
  sessions:
  - conditions:
    - lastTransitionTime: "2024-09-19T06:15:53Z"
      message: Scheduler has been ensured successfully.
      reason: SchedulerEnsured
      status: "True"
      type: SchedulerEnsured
    - lastTransitionTime: "2024-09-19T06:15:54Z"
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
$ kubectl get backupsession -n demo
NAME                                                  INVOKER-TYPE          INVOKER-NAME               PHASE       DURATION   AGE
appbinding-es-quickstart-2-frequent-backup-1726726553   BackupConfiguration   appbinding-es-quickstart-2   Succeeded   19s        2m51s

We can see from the above output that the backup session has succeeded. Now, we are going to verify whether the backed up data has been stored in the backend.

**Verify Backup:**

Once a backup is complete, KubeStash will update the respective `Repository` CR to reflect the backup. Check that the repository `s3-elasticsearch-repo` has been updated by the following command,

```bash
$ kubectl get repo -n demo s3-elasticsearch-repo
NAME                    INTEGRITY   SNAPSHOT-COUNT   SIZE         PHASE   LAST-SUCCESSFUL-BACKUP   AGE
s3-elasticsearch-repo   true        10               15.974 KiB   Ready   17s                      100m
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
$ kubectl get snapshot -n demo s3-elasticsearch-repo-appbindingtart-frequent-backup-1726727401 -oyaml
apiVersion: storage.kubestash.com/v1alpha1
kind: Snapshot
metadata:
  creationTimestamp: "2024-09-19T06:30:01Z"
  finalizers:
  - kubestash.com/cleanup
  generation: 1
  labels:
    kubedb.com/db-version: 8.15.0
    kubestash.com/app-ref-kind: Elasticsearch
    kubestash.com/app-ref-name: es-quickstart-2
    kubestash.com/app-ref-namespace: demo
    kubestash.com/repo-name: s3-elasticsearch-repo
  name: s3-elasticsearch-repo-appbindingtart-frequent-backup-1726727401
  namespace: demo
  ownerReferences:
  - apiVersion: storage.kubestash.com/v1alpha1
    blockOwnerDeletion: true
    controller: true
    kind: Repository
    name: s3-elasticsearch-repo
    uid: c6859c35-2c70-45b7-a8ed-e9969b009b69
  resourceVersion: "88432"
  uid: 5ba5eb63-5076-43a3-91bc-2938c1a35391
spec:
  appRef:
    apiGroup: kubedb.com
    kind: Elasticsearch
    name: es-quickstart-2
    namespace: demo
  backupSession: appbinding-es-quickstart-2-frequent-backup-1726727401
  deletionPolicy: Delete
  repository: s3-elasticsearch-repo
  session: frequent-backup
  snapshotID: 01J84FJBJZD2GJYHGTRCSQG17F
  type: FullBackup
  version: v1
status:
  components:
    dump:
      driver: Restic
      duration: 1.93010181s
      integrity: true
      path: repository/v1/frequent-backup/dump
      phase: Succeeded
      resticStats:
      - hostPath: /kubestash-interim/data
        id: 4e15656770c55e4e08ed6dbfe6a190eb96db979259ca9c3900a5918cac116330
        size: 11.717 KiB
        uploaded: 3.835 KiB
      size: 15.974 KiB
  conditions:
  - lastTransitionTime: "2024-09-19T06:30:01Z"
    message: Recent snapshot list updated successfully
    reason: SuccessfullyUpdatedRecentSnapshotList
    status: "True"
    type: RecentSnapshotListUpdated
  - lastTransitionTime: "2024-09-19T06:30:13Z"
    message: Metadata uploaded to backend successfully
    reason: SuccessfullyUploadedSnapshotMetadata
    status: "True"
    type: SnapshotMetadataUploaded
  integrity: true
  phase: Succeeded
  size: 15.974 KiB
  snapshotTime: "2024-09-19T06:30:01Z"
  totalComponents: 1
```

> KubeStash uses `multielasticdump` to perform backups of target `Elasticsearch` databases. Therefore, the component name for logical backups is set as `dump`.

Now, if we navigate to the S3 bucket, we will see the backed up data stored in the `elastic/es/custom/repository/v1/frequent-backup/dump` directory. KubeStash also keeps the backup for `Snapshot` YAMLs, which can be found in the `elastic/es/custom/snapshots` directory.

> Note: KubeStash stores all dumped data encrypted in the backup directory, meaning it remains unreadable until decrypted.

## Cleanup

To cleanup the resources crated by this tutorial, run the following commands,

```bash
kubectl delete backupblueprints.core.kubestash.com  -n demo es-quickstart-backup-blueprint
kubectl delete backupblueprints.core.kubestash.com  -n demo es-quickstart-custom-backup-blueprint
kubectl delete retentionpolicies.storage.kubestash.com -n demo demo-retention
kubectl delete backupstorage -n demo s3-storage
kubectl delete secret -n demo s3-secret
kubectl delete secret -n demo encrypt-secret
kubectl delete es -n demo es-quickstart
kubectl delete es -n demo es-quickstart-2
```