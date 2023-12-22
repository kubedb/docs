---
title: MySQL Auto-Backup | Stash
description: Backup MySQL using Stash Auto-Backup
menu:
  docs_{{ .version }}:
    identifier: guides-mysql-backup-auto-backup
    name: Auto-Backup
    parent: guides-mysql-backup
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Backup MySQL using Stash Auto-Backup

Stash can be configured to automatically backup any MySQL database in your cluster. Stash enables cluster administrators to deploy backup blueprints ahead of time so that the database owners can easily backup their database with just a few annotations.

In this tutorial, we are going to show how you can configure a backup blueprint for MySQL databases in your cluster and backup them with few annotations.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.
- Install Stash in your cluster following the steps [here](https://stash.run/docs/latest/setup/install/stash/).
- Install KubeDB in your cluster following the steps [here](/docs/setup/README.md).
- If you are not familiar with how Stash backup and restore MySQL databases, please check the following guide [here](/docs/guides/mysql/backup/overview/index.md).
- If you are not familiar with how auto-backup works in Stash, please check the following guide [here](https://stash.run/docs/latest/guides/auto-backup/overview/).
- If you are not familiar with the available auto-backup options for databases in Stash, please check the following guide [here](https://stash.run/docs/latest/guides/auto-backup/database/).

You should be familiar with the following `Stash` concepts:

- [BackupBlueprint](https://stash.run/docs/latest/concepts/crds/backupblueprint/)
- [BackupConfiguration](https://stash.run/docs/latest/concepts/crds/backupconfiguration/)
- [BackupSession](https://stash.run/docs/latest/concepts/crds/backupsession/)
- [Repository](https://stash.run/docs/latest/concepts/crds/repository/)
- [Function](https://stash.run/docs/latest/concepts/crds/function/)
- [Task](https://stash.run/docs/latest/concepts/crds/task/)

In this tutorial, we are going to show backup of three different MySQL databases on three different namespaces named `demo`, `demo-2`, and `demo-3`. Create the namespaces as below if you haven't done it already.

```bash
❯ kubectl create ns demo
namespace/demo created

❯ kubectl create ns demo-2
namespace/demo-2 created

❯ kubectl create ns demo-3
namespace/demo-3 created
```

When you install the Stash, it automatically installs all the official database addons. Verify that it has installed the MySQL addons using the following command.

```bash
❯ kubectl get tasks.stash.appscode.com | grep mysql
mysql-backup-5.7.25            2d2h
mysql-backup-8.0.14            2d2h
mysql-backup-8.0.21            2d2h
mysql-backup-8.0.3             2d2h
mysql-restore-5.7.25           2d2h
mysql-restore-8.0.14           2d2h
mysql-restore-8.0.21           2d2h
mysql-restore-8.0.3            2d2h
```

## Prepare Backup Blueprint

To backup an MySQL database using Stash, you have to create a `Secret` containing the backend credentials, a `Repository` containing the backend information, and a `BackupConfiguration` containing the schedule and target information. A `BackupBlueprint` allows you to specify a template for the `Repository` and the `BackupConfiguration`.

The `BackupBlueprint` is a non-namespaced CRD. So, once you have created a `BackupBlueprint`, you can use it to backup any MySQL database of any namespace just by creating the storage `Secret` in that namespace and adding few annotations to your MySQL CRO. Then, Stash will automatically create a `Repository` and a `BackupConfiguration` according to the template to backup the database.

Below is the `BackupBlueprint` object that we are going to use in this tutorial,

```yaml
apiVersion: stash.appscode.com/v1beta1
kind: BackupBlueprint
metadata:
  name: mysql-backup-template
spec:
  # ============== Blueprint for Repository ==========================
  backend:
    gcs:
      bucket: stash-testing
      prefix: mysql-backup/${TARGET_NAMESPACE}/${TARGET_APP_RESOURCE}/${TARGET_NAME}
    storageSecretName: gcs-secret
  # ============== Blueprint for BackupConfiguration =================
  schedule: "*/5 * * * *"
  retentionPolicy:
    name: 'keep-last-5'
    keepLast: 5
    prune: true
```

Here, we are using a GCS bucket as our backend. We are providing `gcs-secret` at the `storageSecretName` field. Hence, we have to create a secret named `gcs-secret` with the access credentials of our bucket in every namespace where we want to enable backup through this blueprint.

Notice the `prefix` field of `backend` section. We have used some variables in form of `${VARIABLE_NAME}`. Stash will automatically resolve those variables from the database information to make the backend prefix unique for each database instance.

Let's create the `BackupBlueprint` we have shown above,

```bash
❯ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/backup/auto-backup/examples/backupblueprint.yaml
backupblueprint.stash.appscode.com/mysql-backup-template created
```

Now, we are ready to backup our MySQL databases using few annotations. You can check available auto-backup annotations for a databases from [here](https://stash.run/docs/latest/guides/auto-backup/database/#available-auto-backup-annotations-for-database).

## Auto-backup with default configurations

In this section, we are going to backup an MySQL database of `demo` namespace. We are going to use the default configurations specified in the `BackupBlueprint`.

### Create Storage Secret

At first, let's create the `gcs-secret` in `demo` namespace with the access credentials to our GCS bucket.

```bash
❯ echo -n 'changeit' > RESTIC_PASSWORD
❯ echo -n '<your-project-id>' > GOOGLE_PROJECT_ID
❯ cat downloaded-sa-key.json > GOOGLE_SERVICE_ACCOUNT_JSON_KEY
❯ kubectl create secret generic -n demo gcs-secret \
    --from-file=./RESTIC_PASSWORD \
    --from-file=./GOOGLE_PROJECT_ID \
    --from-file=./GOOGLE_SERVICE_ACCOUNT_JSON_KEY
secret/gcs-secret created
```

### Create Database

Now, we are going to create an MySQL CRO in `demo` namespace. Below is the YAML of the MySQL object that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MySQL
metadata:
  name: sample-mysql
  namespace: demo
  annotations:
    stash.appscode.com/backup-blueprint: mysql-backup-template
spec:
  version: "8.0.35"
  replicas: 1
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 50Mi
  terminationPolicy: WipeOut
```

Notice the `annotations` section. We are pointing to the `BackupBlueprint` that we have created earlier though `stash.appscode.com/backup-blueprint` annotation. Stash will watch this annotation and create a `Repository` and a `BackupConfiguration` according to the `BackupBlueprint`.

Let's create the above MySQL CRO,

```bash
❯ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/backup/auto-backup/examples/sample-mysql.yaml
mysql.kubedb.com/sample-mysql created
```

### Verify Auto-backup configured

In this section, we are going to verify whether Stash has created the respective `Repository` and `BackupConfiguration` for our MySQL database we have just deployed or not.

#### Verify Repository

At first, let's verify whether Stash has created a `Repository` for our MySQL or not.

```bash
❯ kubectl get repository -n demo
NAME                 INTEGRITY   SIZE   SNAPSHOT-COUNT   LAST-SUCCESSFUL-BACKUP   AGE
app-sample-mysql                                                                  10s 
```

Now, let's check the YAML of the `Repository`.

```yaml
❯ kubectl get repository -n demo app-sample-mysql -o yaml
apiVersion: stash.appscode.com/v1alpha1
kind: Repository
metadata:
  creationTimestamp: "2022-06-30T05:45:43Z"
  finalizers:
  - stash
  generation: 1
  name: app-sample-mysql
  namespace: demo
  resourceVersion: "363862"
  uid: 23781855-6cd9-4ef8-84d4-c6360a88bbef
spec:
  backend:
    gcs:
      bucket: stash-testing
      prefix: mysql-backup/demo/mysql/sample-mysql
    storageSecretName: gcs-secret

```

Here, you can see that Stash has resolved the variables in `prefix` field and substituted them with the equivalent information from this database.

#### Verify BackupConfiguration

If everything goes well, Stash should create a `BackupConfiguration` for our MySQL in `demo` namespace and the phase of that `BackupConfiguration` should be `Ready`. Verify the `BackupConfiguration` crd by the following command,

```bash
❯ kubectl get backupconfiguration -n demo
NAMESPACE   NAME               TASK   SCHEDULE      PAUSED   PHASE   AGE
demo        app-sample-mysql          */5 * * * *            Ready   3m56s
```

Now, let's check the YAML of the `BackupConfiguration`.

```yaml
❯ kubectl get backupconfiguration -n demo app-sample-mysql -o yaml
apiVersion: stash.appscode.com/v1beta1
kind: BackupConfiguration
metadata:
  creationTimestamp: "2022-06-30T05:45:43Z"
  finalizers:
  - stash.appscode.com
  generation: 1
  name: app-sample-mysql
  namespace: demo
  ownerReferences:
  - apiVersion: appcatalog.appscode.com/v1alpha1
    blockOwnerDeletion: true
    controller: true
    kind: AppBinding
    name: sample-mysql
    uid: 02bacaf0-9f3c-4b48-84d4-305a7d854eb2
  resourceVersion: "363877"
  uid: d101e8fc-4507-42cc-93f7-782f29d8898d
spec:
  driver: Restic
  repository:
    name: app-sample-mysql
  retentionPolicy:
    keepLast: 5
    name: keep-last-5
    prune: true
  runtimeSettings: {}
  schedule: '*/5 * * * *'
  target:
    ref:
      apiVersion: appcatalog.appscode.com/v1alpha1
      kind: AppBinding
      name: sample-mysql
  task: {}
  tempDir: {}
status:
  conditions:
  - lastTransitionTime: "2022-06-30T05:45:43Z"
    message: Repository demo/app-sample-mysql exist.
    reason: RepositoryAvailable
    status: "True"
    type: RepositoryFound
  - lastTransitionTime: "2022-06-30T05:45:43Z"
    message: Backend Secret demo/gcs-secret exist.
    reason: BackendSecretAvailable
    status: "True"
    type: BackendSecretFound
  - lastTransitionTime: "2022-06-30T05:45:43Z"
    message: Backup target appcatalog.appscode.com/v1alpha1 appbinding/sample-mysql
      found.
    reason: TargetAvailable
    status: "True"
    type: BackupTargetFound
  - lastTransitionTime: "2022-06-30T05:45:43Z"
    message: Successfully created backup triggering CronJob.
    reason: CronJobCreationSucceeded
    status: "True"
    type: CronJobCreated
  observedGeneration: 1
```

Notice the `target` section. Stash has automatically added the MySQL as the target of this `BackupConfiguration`.

#### Verify Backup

Now, let's wait for a backup run to complete. You can watch for `BackupSession` as below,

```bash
❯ kubectl get backupsession -n demo -w
NAME                          INVOKER-TYPE          INVOKER-NAME       PHASE     DURATION   AGE
app-sample-mysql-1643879707   BackupConfiguration   app-sample-mysql   Running              40s
```

Once the backup has been completed successfully, you should see the backed up data has been stored in the bucket at the directory pointed by the `prefix` field of the `Repository`.

<figure align="center">
  <img alt="Backup data in GCS Bucket" src="/docs/guides/mysql/backup/auto-backup/images/sample-mysql.png">
  <figcaption align="center">Fig: Backup data in GCS Bucket</figcaption>
</figure>

## Auto-backup with a custom schedule

In this section, we are going to backup an MySQL database of `demo-2` namespace. This time, we are going to overwrite the default schedule used in the `BackupBlueprint`.

### Create Storage Secret

At first, let's create the `gcs-secret` in `demo-2` namespace with the access credentials to our GCS bucket.

```bash
❯ kubectl create secret generic -n demo-2 gcs-secret \
    --from-file=./RESTIC_PASSWORD \
    --from-file=./GOOGLE_PROJECT_ID \
    --from-file=./GOOGLE_SERVICE_ACCOUNT_JSON_KEY
secret/gcs-secret created
```

### Create Database

Now, we are going to create an MySQL CRO in `demo-2` namespace. Below is the YAML of the MySQL object that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MySQL
metadata:
  name: sample-mysql-2
  namespace: demo-2
  annotations:
    stash.appscode.com/backup-blueprint: mysql-backup-template
    stash.appscode.com/schedule: "*/3 * * * *"
spec:
  version: "8.0.35"
  replicas: 1
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 50Mi
  terminationPolicy: WipeOut
```

Notice the `annotations` section. This time, we have passed a schedule via `stash.appscode.com/schedule` annotation along with the `stash.appscode.com/backup-blueprint` annotation.

Let's create the above MySQL CRO,

```bash
❯ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/backup/auto-backup/examples/sample-mysql-2.yaml
mysql.kubedb.com/sample-mysql-2 created
```

### Verify Auto-backup configured

Now, let's verify whether the auto-backup has been configured properly or not.

#### Verify Repository

At first, let's verify whether Stash has created a `Repository` for our MySQL or not.

```bash
❯ kubectl get repository -n demo-2
NAME                   INTEGRITY   SIZE   SNAPSHOT-COUNT   LAST-SUCCESSFUL-BACKUP   AGE
app-sample-myaql-2                                                                  4s
```

Now, let's check the YAML of the `Repository`.

```yaml
❯ kubectl get repository -n demo-2 app-sample-mysql-2  -o yaml
apiVersion: stash.appscode.com/v1alpha1
kind: Repository
metadata:
  creationTimestamp: "2022-06-30T05:45:43Z"
  finalizers:
  - stash
  generation: 1
  name: app-sample-mysql-2
  namespace: demo-2
  resourceVersion: "365836"
  uid: f37e737c-c5f1-4620-9c22-1d7b14127eab
spec:
  backend:
    gcs:
      bucket: stash-testing
      prefix: mysql-backup/demo/mysql/sample-mysql-2
    storageSecretName: gcs-secret
```

Here, you can see that Stash has resolved the variables in `prefix` field and substituted them with the equivalent information from this new database.

#### Verify BackupConfiguration

If everything goes well, Stash should create a `BackupConfiguration` for our MySQL in `demo-2` namespace and the phase of that `BackupConfiguration` should be `Ready`. Verify the `BackupConfiguration` crd by the following command,

```bash
❯ kubectl get backupconfiguration -n demo-2
NAMESPACE   NAME                 TASK   SCHEDULE      PAUSED   PHASE   AGE
demo-2      app-sample-mysql-2          */3 * * * *            Ready   113s
```

Now, let's check the YAML of the `BackupConfiguration`.

```yaml
❯ kubectl get backupconfiguration -n demo-2 app-sample-mysql-2 -o yaml

apiVersion: stash.appscode.com/v1beta1
kind: BackupConfiguration
metadata:
  creationTimestamp: "2022-06-30T05:45:43Z"
  finalizers:
  - stash.appscode.com
  generation: 1
  name: app-sample-mysql-2
  namespace: demo-2
  ownerReferences:
  - apiVersion: appcatalog.appscode.com/v1alpha1
    blockOwnerDeletion: true
    controller: true
    kind: AppBinding
    name: sample-mysql-2
    uid: 478d802c-585b-408b-9fbe-b2f90d55b26e
  resourceVersion: "366551"
  uid: de8448a1-0e8c-41b5-a1c4-07239ae0fef2
spec:
  driver: Restic
  repository:
    name: app-sample-mysql-2
  retentionPolicy:
    keepLast: 5
    name: keep-last-5
    prune: true
  runtimeSettings: {}
  schedule: '*/3 * * * *'
  target:
    ref:
      apiVersion: appcatalog.appscode.com/v1alpha1
      kind: AppBinding
      name: sample-mysql-2
  task: {}
  tempDir: {}
status:
  conditions:
  - lastTransitionTime: "2022-06-30T05:45:43Z"
    message: Repository demo-2/app-sample-mysql-2 exist.
    reason: RepositoryAvailable
    status: "True"
    type: RepositoryFound
  - lastTransitionTime: "2022-06-30T05:45:43Z"
    message: Backend Secret demo-2/gcs-secret exist.
    reason: BackendSecretAvailable
    status: "True"
    type: BackendSecretFound
  - lastTransitionTime: "2022-06-30T05:45:43Z"
    message: Backup target appcatalog.appscode.com/v1alpha1 appbinding/sample-mysql-2
      found.
    reason: TargetAvailable
    status: "True"
    type: BackupTargetFound
  - lastTransitionTime: "2022-06-30T05:45:43Z"
    message: Successfully created backup triggering CronJob.
    reason: CronJobCreationSucceeded
    status: "True"
    type: CronJobCreated
  observedGeneration: 1
```

Notice the `schedule` section. This time the `BackupConfiguration` has been created with the schedule we have provided via annotation.

Also, notice the `target` section. Stash has automatically added the new MySQL as the target of this `BackupConfiguration`.

#### Verify Backup

Now, let's wait for a backup run to complete. You can watch for `BackupSession` as below,

```bash
❯ kubectl get backupsession -n demo-2 -w
NAMESPACE   NAME                            INVOKER-TYPE          INVOKER-NAME         PHASE       DURATION   AGE
demo-2      app-sample-mysql-2-1643880964   BackupConfiguration   app-sample-mysql-2   Succeeded   35s        108s

```

Once the backup has been completed successfully, you should see that Stash has created a new directory as pointed by the `prefix` field of the new `Repository` and stored the backed up data there.

<figure align="center">
  <img alt="Backup data in GCS Bucket" src="/docs/guides/mysql/backup/auto-backup/images/sample-mysql-2.png">
  <figcaption align="center">Fig: Backup data in GCS Bucket</figcaption>
</figure>

## Auto-backup with custom parameters

In this section, we are going to backup an MySQL database of `demo-3` namespace. This time, we are going to pass some parameters for the Task through the annotations.

### Create Storage Secret

At first, let's create the `gcs-secret` in `demo-3` namespace with the access credentials to our GCS bucket.

```bash
❯ kubectl create secret generic -n demo-3 gcs-secret \
    --from-file=./RESTIC_PASSWORD \
    --from-file=./GOOGLE_PROJECT_ID \
    --from-file=./GOOGLE_SERVICE_ACCOUNT_JSON_KEY
secret/gcs-secret created
```

### Create Database

Now, we are going to create an MySQL CRO in `demo-3` namespace. Below is the YAML of the MySQL object that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MySQL
metadata:
  name: sample-mysql-3
  namespace: demo-3
  annotations:
    stash.appscode.com/backup-blueprint: mysql-backup-template
    params.stash.appscode.com/args: --databases mysql
spec:
  version: "8.0.35"
  replicas: 1
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 50Mi
  terminationPolicy: WipeOut
```

Notice the `annotations` section. This time, we have passed an argument via `params.stash.appscode.com/args` annotation along with the `stash.appscode.com/backup-blueprint` annotation.

Let's create the above MySQL CRO,

```bash
❯ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/backup/auto-backup/examples/sample-mysql-3.yaml
mysql.kubedb.com/sample-mysql-3 created
```

### Verify Auto-backup configured

Now, let's verify whether the auto-backup resources has been created or not.

#### Verify Repository

At first, let's verify whether Stash has created a `Repository` for our MySQL or not.

```bash
❯ kubectl get repository -n demo-3
NAME                 INTEGRITY   SIZE   SNAPSHOT-COUNT   LAST-SUCCESSFUL-BACKUP   AGE
app-sample-mysql-3                                                                5s                                                          8s
```

Now, let's check the YAML of the `Repository`.

```yaml
❯ kubectl get repository -n demo-3 app-sample-mysql-3 -o yaml
apiVersion: stash.appscode.com/v1alpha1
kind: Repository
metadata:
  creationTimestamp: "2022-06-30T05:45:43Z"
  finalizers:
  - stash
  generation: 1
  name: app-sample-mysql-3
  namespace: demo-3
  resourceVersion: "371009"
  uid: 244f30d0-cc1e-4d8a-8b76-fcca702783d6
spec:
  backend:
    gcs:
      bucket: stash-testing
      prefix: mysql-backup/demo-3/mysql/sample-mysql-3
    storageSecretName: gcs-secret
```

Here, you can see that Stash has resolved the variables in `prefix` field and substituted them with the equivalent information from this new database.

#### Verify BackupConfiguration

If everything goes well, Stash should create a `BackupConfiguration` for our MySQL in `demo` namespace and the phase of that `BackupConfiguration` should be `Ready`. Verify the `BackupConfiguration` crd by the following command,

```bash
❯ kubectl get backupconfiguration -n demo-3
NAMESPACE   NAME                 TASK   SCHEDULE      PAUSED   PHASE   AGE
demo-3      app-sample-mysql-3          */5 * * * *            Ready   107s
```

Now, let's check the YAML of the `BackupConfiguration`.

```yaml
❯ kubectl get backupconfiguration -n demo-3 app-sample-mysql-3 -o yaml
apiVersion: stash.appscode.com/v1beta1
kind: BackupConfiguration
metadata:
  creationTimestamp: "2022-06-30T05:45:43Z"
  finalizers:
  - stash.appscode.com
  generation: 1
  name: app-sample-mysql-3
  namespace: demo-3
  ownerReferences:
  - apiVersion: appcatalog.appscode.com/v1alpha1
    blockOwnerDeletion: true
    controller: true
    kind: AppBinding
    name: sample-mysql-3
    uid: 3a0682ef-62a5-4acf-adee-fc48f80a0ef7
  resourceVersion: "371026"
  uid: 444901df-64de-44e8-b592-d4b26dfe00de
spec:
  driver: Restic
  repository:
    name: app-sample-mysql-3
  retentionPolicy:
    keepLast: 5
    name: keep-last-5
    prune: true
  runtimeSettings: {}
  schedule: '*/5 * * * *'
  target:
    ref:
      apiVersion: appcatalog.appscode.com/v1alpha1
      kind: AppBinding
      name: sample-mysql-3
  task:
    params:
    - name: args
      value: --databases mysql
  tempDir: {}
status:
  conditions:
  - lastTransitionTime: "2022-06-30T05:45:43Z"
    message: Repository demo-3/app-sample-mysql-3 exist.
    reason: RepositoryAvailable
    status: "True"
    type: RepositoryFound
  - lastTransitionTime: "2022-06-30T05:45:43Z"
    message: Backend Secret demo-3/gcs-secret exist.
    reason: BackendSecretAvailable
    status: "True"
    type: BackendSecretFound
  - lastTransitionTime: "2022-06-30T05:45:43Z"
    message: Backup target appcatalog.appscode.com/v1alpha1 appbinding/sample-mysql-3
      found.
    reason: TargetAvailable
    status: "True"
    type: BackupTargetFound
  - lastTransitionTime: "2022-06-30T05:45:43Z"
    message: Successfully created backup triggering CronJob.
    reason: CronJobCreationSucceeded
    status: "True"
    type: CronJobCreated
  observedGeneration: 1
```

Notice the `task` section. The `args` parameter that we had passed via annotations has been added to the `params` section.

Also, notice the `target` section. Stash has automatically added the new MySQL as the target of this `BackupConfiguration`.

#### Verify Backup

Now, let's wait for a backup run to complete. You can watch for `BackupSession` as below,

```bash
❯ kubectl get backupsession -n demo-3 -w
NAMESPACE   NAME                            INVOKER-TYPE          INVOKER-NAME         PHASE       DURATION   AGE
demo-3      app-sample-mysql-3-1643883304   BackupConfiguration   app-sample-mysql-3   Succeeded   41s        78s
```

Once the backup has been completed successfully, you should see that Stash has created a new directory as pointed by the `prefix` field of the new `Repository` and stored the backed up data there.

<figure align="center">
  <img alt="Backup data in GCS Bucket" src="/docs/guides/mysql/backup/auto-backup/images/sample-mysql-3.png">
  <figcaption align="center">Fig: Backup data in GCS Bucket</figcaption>
</figure>

## Cleanup

To cleanup the resources crated by this tutorial, run the following commands,

```bash
❯ kubectl delete -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mysql/backup/auto-backup/examples/
backupblueprint.stash.appscode.com "mysql-backup-template" deleted
mysql.kubedb.com "sample-mysql-2" deleted
mysql.kubedb.com "sample-mysql-3" deleted
mysql.kubedb.com "sample-mysql" deleted

❯ kubectl delete repository -n demo --all
repository.stash.appscode.com "app-sample-mysql" deleted
❯ kubectl delete repository -n demo-2 --all
repository.stash.appscode.com "app-sample-mysql-2" deleted
❯ kubectl delete repository -n demo-3 --all
repository.stash.appscode.com "app-sample-mysql-3" deleted
```
