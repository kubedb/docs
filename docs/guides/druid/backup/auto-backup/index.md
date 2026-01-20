---
title: Druid Auto-Backup | KubeStash
description: Backup Druid database using KubeStash
menu:
  docs_{{ .version }}:
    identifier: guides-druid-backup-auto-backup
    name: Auto Backup
    parent: guides-druid-backup
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Backup Druid using KubeStash Auto-Backup

[KubeStash](https://kubestash.com) can automatically be configured to backup any `Druid` databases in your cluster. KubeStash enables cluster administrators to deploy backup `blueprints` ahead of time so database owners can easily backup any `Druid` database with a few annotations.

In this tutorial, we are going to show how you can configure a backup blueprint for `Druid` databases in your cluster and backup them with a few annotations.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using `Minikube` or `Kind`.
- Now, install KubeDB cli on your workstation and KubeDB operator in your cluster following the steps [here](/docs/setup/README.md) and make sure to include the flags `--set global.featureGates.Druid=true` to ensure **Druid CRD** and `--set global.featureGates.ZooKeeper=true` to ensure **ZooKeeper CRD** as Druid depends on ZooKeeper for external dependency with helm command.
- Install `KubeStash` in your cluster following the steps [here](https://kubestash.com/docs/latest/setup/install/kubestash).
- Install KubeStash `kubectl` plugin following the steps [here](https://kubestash.com/docs/latest/setup/install/kubectl-plugin/).
- If you are not familiar with how KubeStash backup and restore Druid databases, please check the following guide [here](/docs/guides/druid/backup/overview/index.md).

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

### Prepare Backend

We are going to store our backed up data into a GCS bucket. We have to create a Secret with necessary credentials and a `BackupStorage` CR to use this backend. If you want to use a different backend, please read the respective backend configuration doc from [here](https://kubestash.com/docs/latest/guides/backends/overview/).

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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/druid/backup/auto-backup/examples/backupstorage.yaml
backupstorage.storage.kubestash.com/gcs-storage created
```

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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/druid/backup/auto-backup/examples/retentionpolicy.yaml
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

In this section, we are going to backup a `Druid` database of `demo` namespace. We are going to use the default configurations which will be specified in the `Backup Blueprint` CR.

**Prepare Backup Blueprint**

A `BackupBlueprint` allows you to specify a template for the `Repository`,`Session` or `Variables` of `BackupConfiguration` in a Kubernetes native way.

Now, we have to create a `BackupBlueprint` CR with a blueprint for `BackupConfiguration` object.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupBlueprint
metadata:
  name: druid-default-backup-blueprint
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
          name: druid-addon
          tasks:
            - name: mysql-metadata-storage-backup
```

Here,

- `.spec.backupConfigurationTemplate.backends[*].storageRef` refers to our earlier created `gcs-storage` backupStorage.
- `.spec.backupConfigurationTemplate.sessions[*].schedule` specifies that we want to backup the database at `5 minutes` interval.

Let's create the `BackupBlueprint` we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/druid/backup/auto-backup/examples/default-backupblueprint.yaml
backupblueprint.core.kubestash.com/druid-default-backup-blueprint created
```

Now, we are ready to backup our `Druid` databases using a few annotations.

## Deploy Sample Druid Database


**Create External Dependency (Deep Storage):**

One of the external dependency of Druid is deep storage where the segments are stored. It is a storage mechanism that Apache Druid does not provide. **Amazon S3**, **Google Cloud Storage**, or **Azure Blob Storage**, **S3-compatible storage** (like **Minio**), or **HDFS** are generally convenient options for deep storage.

In this tutorial, we will run a `minio-server` as deep storage in our local `kind` cluster using `minio-operator` and create a bucket named `druid` in it, which the deployed druid database will use.

```bash

$ helm repo add minio https://operator.min.io/
$ helm repo update minio
$ helm upgrade --install --namespace "minio-operator" --create-namespace "minio-operator" minio/operator --set operator.replicaCount=1

$ helm upgrade --install --namespace "demo" --create-namespace druid-minio minio/tenant \
--set tenant.pools[0].servers=1 \
--set tenant.pools[0].volumesPerServer=1 \
--set tenant.pools[0].size=1Gi \
--set tenant.certificate.requestAutoCert=false \
--set tenant.buckets[0].name="druid" \
--set tenant.pools[0].name="default"

```

Now we need to create a `Secret` named `deep-storage-config`. It contains the necessary connection information using which the druid database will connect to the deep storage.

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: deep-storage-config
  namespace: demo
stringData:
  druid.storage.type: "s3"
  druid.storage.bucket: "druid"
  druid.storage.baseKey: "druid/segments"
  druid.s3.accessKey: "minio"
  druid.s3.secretKey: "minio123"
  druid.s3.protocol: "http"
  druid.s3.enablePathStyleAccess: "true"
  druid.s3.endpoint.signingRegion: "us-east-1"
  druid.s3.endpoint.url: "http://myminio-hl.demo.svc.cluster.local:9000/"
```

Let’s create the `deep-storage-config` Secret shown above:

```bash
$ kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/druid/backup/auto-backup/examples/deep-storage-config.yaml
secret/deep-storage-config created
```

Let's deploy a sample `Druid` database and insert some data into it.

**Create Druid CR:**

Below is the YAML of a sample `Druid` CRD that we are going to create for this tutorial:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Druid
metadata:
  name: sample-druid
  namespace: demo
  annotations:
    blueprint.kubestash.com/name: druid-default-backup-blueprint
    blueprint.kubestash.com/namespace: demo
spec:
  version: 30.0.1
  deepStorage:
    type: s3
    configuration:
      secretName: deep-storage-config
  topology:
    routers:
      replicas: 1
  deletionPolicy: WipeOut
```
Here,

- `.spec.annotations.blueprint.kubestash.com/name: druid-default-backup-blueprint` specifies the name of the `BackupBlueprint` that will use in backup.
- `.spec.annotations.blueprint.kubestash.com/namespace: demo` specifies the name of the `namespace` where the `BackupBlueprint` resides.

Create the above `Druid` CR,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/druid/backup/auto-backup/examples/sample-druid.yaml
druid.kubedb.com/sample-druid created
```

**Verify BackupConfiguration**

If everything goes well, KubeStash should create a `BackupConfiguration` for our Druid in demo namespace and the phase of that `BackupConfiguration` should be `Ready`. Verify the `BackupConfiguration` object by the following command,

```bash
$ kubectl get backupconfiguration -n demo
NAME                      PHASE   PAUSED   AGE
appbinding-sample-druid   Ready            8m48s
```

Now, let’s check the YAML of the `BackupConfiguration`.

```bash
$ kubectl get backupconfiguration -n demo appbinding-sample-druid  -o yaml
```

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  creationTimestamp: "2024-09-19T10:30:46Z"
  finalizers:
    - kubestash.com/cleanup
  generation: 1
  labels:
    app.kubernetes.io/managed-by: kubestash.com
    kubestash.com/invoker-name: druid-default-backup-blueprint
    kubestash.com/invoker-namespace: demo
  name: appbinding-sample-druid
  namespace: demo
  resourceVersion: "1594861"
  uid: 8c5a21cd-780b-4b67-b95a-d6338d038dd4
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
        name: druid-addon
        tasks:
          - name: mysql-metadata-storage-backup
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
    kind: Druid
    name: sample-druid
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
```

Notice the `spec.backends`, `spec.sessions` and `spec.target` sections, KubeStash automatically resolved those info from the `BackupBluePrint` and created above `BackupConfiguration`. 

**Verify BackupSession:**

KubeStash triggers an instant backup as soon as the `BackupConfiguration` is ready. After that, backups are scheduled according to the specified schedule.

```bash
$ kubectl get backupsession -n demo -w

NAME                                                 INVOKER-TYPE          INVOKER-NAME              PHASE       DURATION   AGE
appbinding-sample-druid-frequent-backup-1726741846   BackupConfiguration   appbinding-sample-druid   Succeeded   28s        10m
appbinding-sample-druid-frequent-backup-1726742101   BackupConfiguration   appbinding-sample-druid   Succeeded   35s        6m37s
appbinding-sample-druid-frequent-backup-1726742400   BackupConfiguration   appbinding-sample-druid   Succeeded   29s        98s
```

We can see from the above output that the backup session has succeeded. Now, we are going to verify whether the backed up data has been stored in the backend.

**Verify Backup:**

Once a backup is complete, KubeStash will update the respective `Repository` CR to reflect the backup. Check that the repository `default-blueprint` has been updated by the following command,

```bash
$ kubectl get repository -n demo default-blueprint
NAME                INTEGRITY   SNAPSHOT-COUNT   SIZE        PHASE   LAST-SUCCESSFUL-BACKUP   AGE
default-blueprint   true        3                1.757 MiB   Ready   2m23s                    11m
```

At this moment we have one `Snapshot`. Run the following command to check the respective `Snapshot` which represents the state of a backup run for an application.

```bash
$ kubectl get snapshots -n demo -l=kubestash.com/repo-name=default-blueprint
NAME                                                              REPOSITORY          SESSION           SNAPSHOT-TIME          DELETION-POLICY   PHASE       AGE
default-blueprint-appbinding-samruid-frequent-backup-1726741846   default-blueprint   frequent-backup   2024-09-19T10:30:56Z   Delete            Succeeded   11m
default-blueprint-appbinding-samruid-frequent-backup-1726742101   default-blueprint   frequent-backup   2024-09-19T10:35:01Z   Delete            Succeeded   7m49s
default-blueprint-appbinding-samruid-frequent-backup-1726742400   default-blueprint   frequent-backup   2024-09-19T10:40:00Z   Delete            Succeeded   2m50s
```

> **Note**: KubeStash creates a `Snapshot` with the following labels:
> - `kubestash.com/app-ref-kind: <target-kind>`
> - `kubestash.com/app-ref-name: <target-name>`
> - `kubestash.com/app-ref-namespace: <target-namespace>`
> - `kubestash.com/repo-name: <repository-name>`
>
> These labels can be used to watch only the `Snapshot`s related to our target Database or `Repository`.

If we check the YAML of the `Snapshot`, we can find the information about the backed up components of the Database.

```bash
$ kubectl get snapshots -n demo default-blueprint-appbinding-samruid-frequent-backup-1726741846 -oyaml
```

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: Snapshot
metadata:
  creationTimestamp: "2024-09-19T10:30:56Z"
  finalizers:
    - kubestash.com/cleanup
  generation: 1
  labels:
    kubestash.com/app-ref-kind: Druid
    kubestash.com/app-ref-name: sample-druid
    kubestash.com/app-ref-namespace: demo
    kubestash.com/repo-name: default-blueprint
  annotations:
    kubedb.com/db-version: 30.0.1
  name: default-blueprint-appbinding-samruid-frequent-backup-1726741846
  namespace: demo
  ownerReferences:
    - apiVersion: storage.kubestash.com/v1alpha1
      blockOwnerDeletion: true
      controller: true
      kind: Repository
      name: default-blueprint
      uid: 7ced6866-349b-48c0-821d-d1ecfee1c80e
  resourceVersion: "1594964"
  uid: 8ec9bb0c-590c-47b8-944b-22af92d62470
spec:
  appRef:
    apiGroup: kubedb.com
    kind: Druid
    name: sample-druid
    namespace: demo
  backupSession: appbinding-sample-druid-frequent-backup-1726741846
  deletionPolicy: Delete
  repository: default-blueprint
  session: frequent-backup
  snapshotID: 01J84XBGGY0JKG7JKTRCGV3HYM
  type: FullBackup
  version: v1
status:
  components:
    dump:
      driver: Restic
      duration: 9.614587405s
      integrity: true
      path: repository/v1/frequent-backup/dump
      phase: Succeeded
      resticStats:
        - hostPath: dumpfile.sql
          id: 8f2b5f5d8a7a18304917e2d4c5a3636f8927085b15c652c35d5fca4a9988515d
          size: 3.750 MiB
          uploaded: 3.751 MiB
      size: 674.017 KiB
```

> KubeStash uses the `mysqldump`/`postgresdump` command to take backups of metadata storage of target Druid databases. Therefore, the component name for `logical backups` is set as `dump`.

Now, if we navigate to the GCS bucket, we will see the backed up data stored in the `/blueprint/default-blueprint/repository/v1/frequent-backup/dump` directory. KubeStash also keeps the backup for `Snapshot` YAMLs, which can be found in the `blueprint/default-blueprintrepository/snapshots` directory.

> **Note**: KubeStash stores all dumped data encrypted in the backup directory, meaning it remains unreadable until decrypted.

## Auto-backup with custom configurations

In this section, we are going to backup a `Druid` database of `demo` namespace. We are going to use the custom configurations which will be specified in the `BackupBlueprint` CR.

**Prepare Backup Blueprint**

A `BackupBlueprint` allows you to specify a template for the `Repository`,`Session` or `Variables` of `BackupConfiguration` in a Kubernetes native way.

Now, we have to create a `BackupBlueprint` CR with a blueprint for `BackupConfiguration` object.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupBlueprint
metadata:
  name: druid-customize-backup-blueprint
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
          name: druid-addon
          tasks:
            - name: mysql-metadata-storage-backup
              params:
                databases: ${targetedDatabases}
```

Note that we have used some variables (format: `${<variable name>}`) in different fields. KubeStash will substitute these variables with values from the respective target’s annotations. You’re free to use any variables you like.

Here,

- `.spec.backupConfigurationTemplate.backends[*].storageRef` refers our earlier created `gcs-storage` backupStorage.
- `.spec.backupConfigurationTemplate.sessions[*]`:
  - `.schedule` defines `${schedule}` variable, which determines the time interval for the backup.
  - `.repositories[*].name` defines the `${repoName}` variable, which specifies the name of the backup `Repository`.
  - `.repositories[*].directory` defines two variables, `${namespace}` and `${targetName}`, which are used to determine the path where the backup will be stored. 
  - `.addon.tasks[*]databases` defines `${targetedDatabases}` variable, which identifies list of databases to backup.

> **Note**: To create `BackupBlueprint` for druid with `PostgreSQL` as metadata storage just update `spec.sessions[*].addon.tasks.name` to `postgres-metadata-storage-restore`

Let's create the `BackupBlueprint` we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/druid/backup/auto-backup/examples/customize-backupblueprint.yaml
backupblueprint.core.kubestash.com/druid-customize-backup-blueprint created
```

Now, we are ready to backup our `Druid` databases using few annotations. You can check available auto-backup annotations for a databases from [here](https://kubestash.com/docs/latest/concepts/crds/backupblueprint/).

**Create Database**

Before proceeding to creating a new `Druid` database, let us clean up the resources of the previous step:
```bash
kubectl delete backupblueprints.core.kubestash.com  -n demo druid-default-backup-blueprint
kubectl delete druid -n demo sample-druid
```

Now, we are going to create a new `Druid` CR in demo namespace. Below is the YAML of the Druid object that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Druid
metadata:
  name: sample-druid-2
  namespace: demo
  annotations:
    blueprint.kubestash.com/name: druid-customize-backup-blueprint
    blueprint.kubestash.com/namespace: demo
    variables.kubestash.com/schedule: "*/10 * * * *"
    variables.kubestash.com/repoName: customize-blueprint
    variables.kubestash.com/namespace: demo
    variables.kubestash.com/targetName: sample-druid-2
    variables.kubestash.com/targetedDatabases: druid
spec:
  version: 30.0.1
  deepStorage:
    type: s3
    configuration:
      secretName: deep-storage-config
  topology:
    routers:
      replicas: 1
  deletionPolicy: WipeOut
```

Notice the `metadata.annotations` field, where we have defined the annotations related to the automatic backup configuration. Specifically, we've set the `BackupBlueprint` name as `druid-customize-backup-blueprint` and the namespace as `demo`. We have also provided values for the blueprint template variables, such as the backup `schedule`, `repositoryName`, `namespace`, `targetName`, and `targetedDatabases`. These annotations will be used to create a `BackupConfiguration` for this `Druid` database.

Let's create the `Druid` we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/druid/backup/auto-backup/examples/sample-druid-2.yaml
druid.kubedb.com/sample-druid-2 created
```

**Verify BackupConfiguration**

If everything goes well, KubeStash should create a `BackupConfiguration` for our Druid in demo namespace and the phase of that `BackupConfiguration` should be `Ready`. Verify the `BackupConfiguration` object by the following command,

```bash
$ kubectl get backupconfiguration -n demo
NAME                        PHASE   PAUSED   AGE
appbinding-sample-druid-2   Ready            2m50m
```

Now, let’s check the YAML of the `BackupConfiguration`.

```bash
$ kubectl get backupconfiguration -n demo appbinding-sample-druid-2  -o yaml
```

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  creationTimestamp: "2024-09-19T11:00:56Z"
  finalizers:
    - kubestash.com/cleanup
  generation: 1
  labels:
    app.kubernetes.io/managed-by: kubestash.com
    kubestash.com/invoker-name: druid-customize-backup-blueprint
    kubestash.com/invoker-namespace: demo
  name: appbinding-sample-druid-2
  namespace: demo
  resourceVersion: "1599083"
  uid: 1c979902-33cd-4212-ae6d-ea4e4198bcaf
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
        name: druid-addon
        tasks:
          - name: mysql-metadata-storage-backup
            params:
              databases: druid
      name: frequent-backup
      repositories:
        - backend: gcs-backend
          directory: demo/sample-druid-2
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
    kind: Druid
    name: sample-druid-2
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
```

Notice the `spec.backends`, `spec.sessions` and `spec.target` sections, KubeStash automatically resolved those info from the `BackupBluePrint` and created above `BackupConfiguration`. 

**Verify BackupSession:**

KubeStash triggers an instant backup as soon as the `BackupConfiguration` is ready. After that, backups are scheduled according to the specified schedule.

```bash
$ kubectl get backupsession -n demo -w

NAME                                                   INVOKER-TYPE          INVOKER-NAME                PHASE       DURATION   AGE
appbinding-sample-druid-2-frequent-backup-1726743656   BackupConfiguration   appbinding-sample-druid-2   Succeeded   30s        2m32s
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
customize-blueprint-appbinding-sid-2-frequent-backup-1726743656   customize-blueprint   frequent-backup   2024-09-19T11:01:06Z   Delete            Succeeded   2m56s
```

> **Note**: KubeStash creates a `Snapshot` with the following labels:
> - `kubestash.com/app-ref-kind: <target-kind>`
> - `kubestash.com/app-ref-name: <target-name>`
> - `kubestash.com/app-ref-namespace: <target-namespace>`
> - `kubestash.com/repo-name: <repository-name>`
>
> These labels can be used to watch only the `Snapshot`s related to our target Database or `Repository`.

If we check the YAML of the `Snapshot`, we can find the information about the backed up components of the Database.

```bash
$ kubectl get snapshots -n demo customize-blueprint-appbinding-sid-2-frequent-backup-1726743656 -oyaml
```

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: Snapshot
metadata:
  creationTimestamp: "2024-09-19T11:01:06Z"
  finalizers:
    - kubestash.com/cleanup
  generation: 1
  labels:
    kubestash.com/app-ref-kind: Druid
    kubestash.com/app-ref-name: sample-druid-2
    kubestash.com/app-ref-namespace: demo
    kubestash.com/repo-name: customize-blueprint
  annotations:
    kubedb.com/db-version: 30.0.1
  name: customize-blueprint-appbinding-sid-2-frequent-backup-1726743656
  namespace: demo
  ownerReferences:
    - apiVersion: storage.kubestash.com/v1alpha1
      blockOwnerDeletion: true
      controller: true
      kind: Repository
      name: customize-blueprint
      uid: 5eaccae6-046c-4c6a-9b76-087d040f001a
  resourceVersion: "1599190"
  uid: 014c050d-0e91-43eb-b60a-36eefbd4b048
spec:
  appRef:
    apiGroup: kubedb.com
    kind: Druid
    name: sample-druid-2
    namespace: demo
  backupSession: appbinding-sample-druid-2-frequent-backup-1726743656
  deletionPolicy: Delete
  repository: customize-blueprint
  session: frequent-backup
  snapshotID: 01J84Z2R6R64FH8E7QYNNZGC1S
  type: FullBackup
  version: v1
status:
  components:
    dump:
      driver: Restic
      duration: 9.132887467s
      integrity: true
      path: repository/v1/frequent-backup/dump
      phase: Succeeded
      resticStats:
        - hostPath: dumpfile.sql
          id: a1061e74f1ad398a9fe85bcbae34f540f2437a97061fd26c5b3e6bde3b5b7642
          size: 10.859 KiB
          uploaded: 11.152 KiB
      size: 2.127 KiB
```

> KubeStash uses the `mysqldump`/`postgresdump` command to take backups of the metadata storage of the target Druid databases. Therefore, the component name for `logical backups` is set as `dump`.

Now, if we navigate to the GCS bucket, we will see the backed up data stored in the `/blueprint/custom-blueprint/repository/v1/frequent-backup/dump` directory. KubeStash also keeps the backup for `Snapshot` YAMLs, which can be found in the `blueprint/custom-blueprint/snapshots` directory.

> **Note**: KubeStash stores all dumped data encrypted in the backup directory, meaning it remains unreadable until decrypted.

## Cleanup

To cleanup the resources crated by this tutorial, run the following commands,

```bash
kubectl delete backupblueprints.core.kubestash.com  -n demo druid-default-backup-blueprint
kubectl delete backupblueprints.core.kubestash.com  -n demo druid-customize-backup-blueprint
kubectl delete backupstorage -n demo gcs-storage
kubectl delete secret -n demo gcs-secret
kubectl delete secret -n demo encrypt-secret
kubectl delete retentionpolicies.storage.kubestash.com -n demo demo-retention
kubectl delete druid -n demo sample-druid
kubectl delete druid -n demo sample-druid-2
```
