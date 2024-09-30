---
title: Microsoft SQL Server Auto-Backup | KubeStash
description: Backup Microsoft SQL Server using KubeStash Auto-Backup
menu:
  docs_{{ .version }}:
    identifier: guides-mssqlserver-auto-backup
    name: Auto-Backup
    parent: guides-mssqlserver-backup
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Backup Microsoft SQL Server using KubeStash Auto-Backup

KubeStash can automatically be configured to backup any `Microsoft SQL Server` databases in your cluster. KubeStash enables cluster administrators to deploy backup `blueprints` ahead of time so database owners can easily backup any `Microsoft SQL Server` database with a few annotations.

In this tutorial, we are going to show how you can configure a backup blueprint for `Microsoft SQL Server` databases in your cluster and backup them with a few annotations.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using `Minikube` or `Kind`.
- Install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md).
- Install `KubeStash` in your cluster following the steps [here](https://kubestash.com/docs/latest/setup/install/kubestash).
- Install KubeStash `kubectl` plugin following the steps [here](https://kubestash.com/docs/latest/setup/install/kubectl-plugin/).
- If you are not familiar with how KubeStash backup and restore `Microsoft SQL Server` databases, please check the following guide [here](/docs/guides/mssqlserver/backup/overview/index.md).

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

> **Note:** YAML files used in this tutorial are stored in [docs/guides/mssqlserver/backup/auto-backup/examples](https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mssqlserver/backup/auto-backup/examples) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.


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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mssqlserver/backup/auto-backup/examples/backupstorage.yaml
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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mssqlserver/backup/auto-backup/examples/retentionpolicy.yaml
retentionpolicy.storage.kubestash.com/demo-retention created
```

### Prepare Issuer/ClusterIssuer

By default, a KubeDB-managed `Microsoft SQL Server` instance run with TLS disabled. However, the `.spec.tls` field is mandatory and will be used during backup and restore operations.

**Create Issuer/ClusterIssuer:**

Now, we are going to create an example `Issuer` CR that will be used throughout the duration of this tutorial. Alternatively, you can follow this [cert-manager](https://cert-manager.io/docs/configuration/ca/) tutorial to create your own `Issuer` CR.

By following the below steps, we are going to create our desired issuer,

- Start off by generating our ca-certificates using openssl,

```bash
$ openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout ./ca.key -out ./ca.crt -subj "/CN=mssqlserver/O=kubedb"
```

- create a secret using the certificate files we have just generated,

```bash
$ kubectl create secret tls mssqlserver-ca --cert=ca.crt  --key=ca.key --namespace=demo 
secret/mssqlserver-ca created
```

Now, we are going to create an `Issuer` using the `mssqlserver-ca` secret that contains the ca-certificate we have just created. Below is the YAML of the `Issuer` cr that we are going to create,

```yaml
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
 name: mssqlserver-ca-issuer
 namespace: demo
spec:
 ca:
   secretName: mssqlserver-ca
```

Let’s create the `Issuer` CR we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mssqlserver/backup/auto-backup/examples/mssqlserver-ca-issuer.yaml
issuer.cert-manager.io/mssqlserver-ca-issuer.yaml created
```

## Auto-backup with default configurations

In this section, we are going to backup a `Microsoft SQL Server` database of `demo` namespace. We are going to use the default configurations which will be specified in the `BackupBlueprint` CR.

**Prepare Backup Blueprint**

A `BackupBlueprint` allows you to specify a template for the `Repository`,`Session` or `Variables` of `BackupConfiguration` in a Kubernetes native way.

Now, we have to create a `BackupBlueprint` CR with a blueprint for `BackupConfiguration` object.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupBlueprint
metadata:
  name: mssqlserver-default-backup-blueprint
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
        addon:
          name: mssqlserver-addon
          jobTemplate:
            spec:
              securityContext:
                runAsUser: 0
          tasks:
            - name: logical-backup
```

Here,

- `.spec.backupConfigurationTemplate.backends[*].storageRef` refers our earlier created `gcs-storage` backupStorage.
- `.spec.backupConfigurationTemplate.sessions[*].schedule` specifies that we want to backup the database at `5 minutes` interval.

Let's create the `BackupBlueprint` we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mssqlserver/backup/auto-backup/examples/default-backupblueprint.yaml
backupblueprint.core.kubestash.com/mssqlserver-default-backup-blueprint created
```

Now, we are ready to backup our `Microsoft SQL Server` databases using few annotations.

**Create Database**

Now, we are going to create an `MSSQLServer` CR in demo namespace. Below is the YAML of the `MSSQLServer` object that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MSSQLServer
metadata:
  name: sample-mssqlserver
  namespace: demo
  annotations:
    blueprint.kubestash.com/name: mssqlserver-default-backup-blueprint
    blueprint.kubestash.com/namespace: demo
spec:
  version: "2022-cu12"
  replicas: 1
  storageType: Durable
  tls:
    issuerRef:
      name: mssqlserver-ca-issuer
      kind: Issuer
      apiGroup: "cert-manager.io"
    clientTLS: false
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Here,

- `.spec.annotations.blueprint.kubestash.com/name: mssqlserver-default-backup-blueprint` specifies the name of the `BackupBlueprint` that will use in backup.
- `.spec.annotations.blueprint.kubestash.com/namespace: demo` specifies the name of the `namespace` where the `BackupBlueprint` resides.

Let's create the `MSSQLServer` we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mssqlserver/backup/auto-backup/examples/sample-mssqlserver.yaml
mssqlserver.kubedb.com/sample-mssqlserver created
```

**Verify BackupConfiguration**

If everything goes well, KubeStash should create a `BackupConfiguration` for our MSSQLServer in demo namespace and the phase of that `BackupConfiguration` should be `Ready`. Verify the `BackupConfiguration` object by the following command,

```bash
$ kubectl get backupconfiguration -n demo
NAME                            PHASE   PAUSED   AGE
appbinding-sample-mssqlserver   Ready            2m50m
```

Now, let’s check the YAML of the `BackupConfiguration`.

```bash
$ kubectl get backupconfiguration -n demo appbinding-sample-mssqlserver  -o yaml
```

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  creationTimestamp: "2024-09-26T05:50:37Z"
  finalizers:
  - kubestash.com/cleanup
  generation: 1
  labels:
    app.kubernetes.io/managed-by: kubestash.com
    kubestash.com/invoker-name: mssqlserver-default-backup-blueprint
    kubestash.com/invoker-namespace: demo
  name: appbinding-sample-mssqlserver
  namespace: demo
  resourceVersion: "502597"
  uid: 4989c9eb-9a91-4540-af2d-da325c5c9bc6
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
      jobTemplate:
        controller: {}
        metadata: {}
        spec:
          resources: {}
          securityContext:
            runAsUser: 0
      name: mssqlserver-addon
      tasks:
      - name: logical-backup
    name: frequent-backup
    repositories:
    - backend: gcs-backend
      directory: /default-blueprint
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
    kind: MSSQLServer
    name: sample-mssqlserver
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
  - lastTransitionTime: "2024-09-26T05:50:37Z"
    message: Validation has been passed successfully.
    reason: ResourceValidationPassed
    status: "True"
    type: ValidationPassed
  dependencies:
  - found: true
    kind: Addon
    name: mssqlserver-addon
  phase: Ready
  repositories:
  - name: default-blueprint
    phase: Ready
  sessions:
  - conditions:
    - lastTransitionTime: "2024-09-26T05:50:47Z"
      message: Scheduler has been ensured successfully.
      reason: SchedulerEnsured
      status: "True"
      type: SchedulerEnsured
    - lastTransitionTime: "2024-09-26T05:50:47Z"
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
appbinding-sample-mssqlserver-frequent-backup-1727329837   BackupConfiguration   appbinding-sample-mssqlserver   Succeeded   23s        6m40s
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
default-blueprint-appbinding-samrver-frequent-backup-1727329837   default-blueprint   frequent-backup   2024-09-05T10:53:59Z   Delete            Succeeded   7m48s
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
$ kubectl get snapshots -n demo default-blueprint-appbinding-samrver-frequent-backup-1727329837  -oyaml
```

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: Snapshot
metadata:
  annotations:
    kubedb.com/db-version: "2022"
  creationTimestamp: "2024-09-26T05:50:47Z"
  finalizers:
    - kubestash.com/cleanup
  generation: 1
  labels:
    kubestash.com/app-ref-kind: MSSQLServer
    kubestash.com/app-ref-name: sample-mssqlserver
    kubestash.com/app-ref-namespace: demo
    kubestash.com/repo-name: default-blueprint
  name: default-blueprint-appbinding-samrver-frequent-backup-1727329837
  namespace: demo
  ownerReferences:
    - apiVersion: storage.kubestash.com/v1alpha1
      blockOwnerDeletion: true
      controller: true
      kind: Repository
      name: default-blueprint
      uid: 8099935a-4c9d-4910-b784-12148f67e1d6
  resourceVersion: "502744"
  uid: 9484ee27-cf0e-4a56-87e0-14daa65019e0
spec:
  appRef:
    apiGroup: kubedb.com
    kind: MSSQLServer
    name: sample-mssqlserver
    namespace: demo
  backupSession: appbinding-sample-mssqlserver-frequent-backup-1727329837
  deletionPolicy: Delete
  repository: default-blueprint
  session: frequent-backup
  snapshotID: 01J8PE3HZS4WSJ15AZHCKN7TVX
  type: FullBackup
  version: v1
status:
  components:
    dump:
      driver: WalG
      duration: 16.633003s
      path: repository/v1/frequent-backup/dump
      phase: Succeeded
      walGStats:
        databases:
          - kubedb_system
        id: base_20240926T055114Z
        startTime: "2024-09-26T05:51:14Z"
        stopTime: "2024-09-26T05:51:31Z"
  conditions:
    - lastTransitionTime: "2024-09-26T05:50:47Z"
      message: Recent snapshot list updated successfully
      reason: SuccessfullyUpdatedRecentSnapshotList
      status: "True"
      type: RecentSnapshotListUpdated
    - lastTransitionTime: "2024-09-26T05:51:33Z"
      message: Metadata uploaded to backend successfully
      reason: SuccessfullyUploadedSnapshotMetadata
      status: "True"
      type: SnapshotMetadataUploaded
  phase: Succeeded
  snapshotTime: "2024-09-26T05:50:47Z"
  totalComponents: 1
```

Now, if we navigate to the GCS bucket, we will see the backed up data stored in the `blueprint/default-blueprint/repository/v1/frequent-backup/dump` directory. KubeStash also keeps the backup for `Snapshot` YAMLs, which can be found in the `blueprint/default-blueprint/snapshots` directory.

## Auto-backup with custom configurations

In this section, we are going to backup a `Microsoft SQL Server` database of `demo` namespace. We are going to use the custom configurations which will be specified in the `BackupBlueprint` CR.

**Prepare Backup Blueprint**

A `BackupBlueprint` allows you to specify a template for the `Repository`,`Session` or `Variables` of `BackupConfiguration` in a Kubernetes native way.

Now, we have to create a `BackupBlueprint` CR with a blueprint for `BackupConfiguration` object.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupBlueprint
metadata:
  name: mssqlserver-customize-backup-blueprint
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
          name: mssqlserver-addon
          jobTemplate:
            spec:
              securityContext:
                runAsUser: 0
          tasks:
            - name: logical-backup
              params:
                databases: ${targetedDatabase}
```

Note that we have used some variables (format: `${<variable name>}`) in different fields. KubeStash will substitute these variables with values from the respective target’s annotations. You’re free to use any variables you like.

Here,

- `.spec.backupConfigurationTemplate.backends[*].storageRef` refers our earlier created `gcs-storage` backupStorage.
- `.spec.backupConfigurationTemplate.sessions[*]`:
    - `.schedule` defines `${schedule}` variable, which determines the time interval for the backup.
    - `.repositories[*].name` defines the `${repoName}` variable, which specifies the name of the backup `Repository`.
    - `.repositories[*].directory` defines two variables, `${namespace}` and `${targetName}`, which are used to determine the path where the backup will be stored.
    - `.addon.tasks[*].params.args` defines `${targetedDatabase}` variable, which identifies list of databases to backup.

Let's create the `BackupBlueprint` we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mssqlserver/backup/auto-backup/examples/customize-backupblueprint.yaml
backupblueprint.core.kubestash.com/mssqlserver-customize-backup-blueprint created
```

Now, we are ready to backup our `Microsoft SQL Server` databases using few annotations. You can check available auto-backup annotations for a databases from [here](https://kubestash.com/docs/latest/concepts/crds/backupblueprint/).

**Create Database**

We will now deploy an SQL Server Availability Group cluster by creating an `MSSQLServer` CR in the demo namespace. Below is the YAML configuration for the `MSSQLServer` object we are about to create:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MSSQLServer
metadata:
  name: sample-mssqlserver-2
  namespace: demo
  annotations:
    blueprint.kubestash.com/name: mssqlserver-customize-backup-blueprint
    blueprint.kubestash.com/namespace: demo
    variables.kubestash.com/schedule: "*/10 * * * *"
    variables.kubestash.com/repoName: customize-blueprint
    variables.kubestash.com/namespace: demo
    variables.kubestash.com/targetName: sample-mssqlserver-2
    variables.kubestash.com/targetedDatabase: agdb1
spec:
  version: "2022-cu12"
  replicas: 3
  topology:
    mode: AvailabilityGroup
    availabilityGroup:
      databases:
        - agdb1
        - agdb2
  internalAuth:
    endpointCert:
      issuerRef:
        apiGroup: cert-manager.io
        name: mssqlserver-ca-issuer
        kind: Issuer
  tls:
    issuerRef:
      name: mssqlserver-ca-issuer
      kind: Issuer
      apiGroup: cert-manager.io
    clientTLS: false
  storageType: Durable
  storage:
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Notice the `metadata.annotations` field, where we have defined the annotations related to the automatic backup configuration. Specifically, we've set the `BackupBlueprint` name as `mssqlserver-customize-backup-blueprint` and the namespace as `demo`. We have also provided values for the blueprint template variables, such as the backup `schedule`, `repositoryName`, `namespace`, `targetName`, and `targetedDatabase`. These annotations will be used to create a `BackupConfiguration` for this `MSSQLServer` database.

Let's create the `MSSQLServer` object  we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mssqlserver/backup/auto-backup/examples/sample-mssqlserver-2.yaml
mssqlserver.kubedb.com/sample-mssqlserver-2 created
```

**Verify BackupConfiguration**

If everything goes well, KubeStash should create a `BackupConfiguration` for our MSSQLServer in demo namespace and the phase of that `BackupConfiguration` should be `Ready`. Verify the `BackupConfiguration` object by the following command,

```bash
$ kubectl get backupconfiguration -n demo
NAME                              PHASE   PAUSED      AGE
appbinding-sample-mssqlserver-2   Ready               2m50m
```

Now, let’s check the YAML of the `BackupConfiguration`.

```bash
$ kubectl get backupconfiguration -n demo appbinding-sample-mssqlserver-2  -o yaml
```

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  creationTimestamp: "2024-09-26T06:54:16Z"
  finalizers:
  - kubestash.com/cleanup
  generation: 2
  labels:
    app.kubernetes.io/managed-by: kubestash.com
    kubestash.com/invoker-name: mssqlserver-customize-backup-blueprint
    kubestash.com/invoker-namespace: demo
  name: appbinding-sample-mssqlserver-2
  namespace: demo
  resourceVersion: "511522"
  uid: 3db263e2-6445-453c-bb16-7bae07876864
spec:
  backends:
  - name: gcs-backend
    retentionPolicy:
      name: demo-retention
      namespace: demo
    storageRef:
      name: gcs-storage
      namespace: demo
  paused: true
  sessions:
  - addon:
      jobTemplate:
        controller: {}
        metadata: {}
        spec:
          resources: {}
          securityContext:
            runAsUser: 0
      name: mssqlserver-addon
      tasks:
      - name: logical-backup
        params:
          databases: agdb1
    name: frequent-backup
    repositories:
    - backend: gcs-backend
      directory: demo/sample-mssqlserver-2
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
    kind: MSSQLServer
    name: sample-mssqlserver-2
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
  - lastTransitionTime: "2024-09-26T06:54:16Z"
    message: Validation has been passed successfully.
    reason: ResourceValidationPassed
    status: "True"
    type: ValidationPassed
  dependencies:
  - found: true
    kind: Addon
    name: mssqlserver-addon
  phase: Ready
  repositories:
  - name: customize-blueprint
    phase: Ready
  sessions:
  - conditions:
    - lastTransitionTime: "2024-09-26T06:54:26Z"
      message: Scheduler has been ensured successfully.
      reason: SchedulerEnsured
      status: "True"
      type: SchedulerEnsured
    - lastTransitionTime: "2024-09-26T06:54:26Z"
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
NAME                                                         INVOKER-TYPE          INVOKER-NAME                      PHASE       DURATION   AGE
appbinding-sample-mssqlserver-2-frequent-backup-1727333656   BackupConfiguration   appbinding-sample-mssqlserver-2   Succeeded   1m18s      2m48s
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
customize-blueprint-appbinding-ser-2-frequent-backup-1727333656   customize-blueprint   frequent-backup   2024-09-26T06:54:26Z   Delete            Succeeded   4m52s
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
$ kubectl get snapshots -n demo customize-blueprint-appbinding-ser-2-frequent-backup-1727333656  -oyaml
```

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: Snapshot
metadata:
  annotations:
    kubedb.com/db-version: "2022"
  creationTimestamp: "2024-09-26T06:54:26Z"
  finalizers:
  - kubestash.com/cleanup
  generation: 1
  labels:
    kubestash.com/app-ref-kind: MSSQLServer
    kubestash.com/app-ref-name: sample-mssqlserver-2
    kubestash.com/app-ref-namespace: demo
    kubestash.com/repo-name: customize-blueprint
  name: customize-blueprint-appbinding-ser-2-frequent-backup-1727333656
  namespace: demo
  ownerReferences:
  - apiVersion: storage.kubestash.com/v1alpha1
    blockOwnerDeletion: true
    controller: true
    kind: Repository
    name: customize-blueprint
    uid: 8181c375-96e3-4969-a137-ea3dd52abf36
  resourceVersion: "511423"
  uid: 7bef59c5-3844-481d-9221-584f2f02ce5f
spec:
  appRef:
    apiGroup: kubedb.com
    kind: MSSQLServer
    name: sample-mssqlserver-2
    namespace: demo
  backupSession: appbinding-sample-mssqlserver-2-frequent-backup-1727333656
  deletionPolicy: Delete
  repository: customize-blueprint
  session: frequent-backup
  snapshotID: 01J8PHR3T0YAFD8RVRC43YB3XT
  type: FullBackup
  version: v1
status:
  components:
    dump:
      driver: WalG
      duration: 23.565835s
      path: repository/v1/frequent-backup/dump
      phase: Succeeded
      walGStats:
        databases:
        - agdb1
        id: base_20240926T065513Z
        startTime: "2024-09-26T06:55:13Z"
        stopTime: "2024-09-26T06:55:37Z"
  conditions:
  - lastTransitionTime: "2024-09-26T06:54:26Z"
    message: Recent snapshot list updated successfully
    reason: SuccessfullyUpdatedRecentSnapshotList
    status: "True"
    type: RecentSnapshotListUpdated
  - lastTransitionTime: "2024-09-26T06:55:41Z"
    message: Metadata uploaded to backend successfully
    reason: SuccessfullyUploadedSnapshotMetadata
    status: "True"
    type: SnapshotMetadataUploaded
  phase: Succeeded
  snapshotTime: "2024-09-26T06:54:26Z"
  totalComponents: 1
```

Now, if we navigate to the GCS bucket, we will see the backed up data stored in the `blueprint/demo/sample-mssqlserver-2/repository/v1/frequent-backup/dump` directory. KubeStash also keeps the backup for `Snapshot` YAMLs, which can be found in the `blueprint/demo/sample-mssqlserver-2/snapshots` directory.

## Cleanup

To cleanup the resources crated by this tutorial, run the following commands,

```bash
kubectl delete backupblueprints.core.kubestash.com  -n demo mssqlserver-default-backup-blueprint
kubectl delete backupblueprints.core.kubestash.com  -n demo mssqlserver-customize-backup-blueprint
kubectl delete retentionpolicies.storage.kubestash.com -n demo demo-retention
kubectl delete backupstorage -n demo gcs-storage
kubectl delete secret -n demo gcs-secret
kubectl delete secrets -n demo mssqlserver-ca
kubectl delete issuer -n demo mssqlserver-ca-issuer
kubectl delete mssqlserver -n demo sample-mssqlserver
kubectl delete mssqlserver -n demo sample-mssqlserver-2
```