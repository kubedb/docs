---
title: MongoDB Auto-Backup | Stash
description: Backup MongoDB using Stash Auto-Backup
menu:
  docs_{{ .version }}:
    identifier: guides-mongodb-backup-auto-backup
    name: Auto-Backup
    parent: guides-mongodb-backup
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Backup MongoDB using Stash Auto-Backup

Stash can be configured to automatically backup any MongoDB database in your cluster. Stash enables cluster administrators to deploy backup blueprints ahead of time so that the database owners can easily backup their database with just a few annotations.

In this tutorial, we are going to show how you can configure a backup blueprint for MongoDB databases in your cluster and backup them with few annotations.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.
- Install Stash in your cluster following the steps [here](https://stash.run/docs/latest/setup/install/stash/).
- Install KubeDB in your cluster following the steps [here](/docs/setup/README.md).
- If you are not familiar with how Stash backup and restore MongoDB databases, please check the following guide [here](/docs/guides/mongodb/backup/overview/index.md).
- If you are not familiar with how auto-backup works in Stash, please check the following guide [here](https://stash.run/docs/latest/guides/auto-backup/overview/).
- If you are not familiar with the available auto-backup options for databases in Stash, please check the following guide [here](https://stash.run/docs/latest/guides/auto-backup/database/).

You should be familiar with the following `Stash` concepts:

- [BackupBlueprint](https://stash.run/docs/latest/concepts/crds/backupblueprint/)
- [BackupConfiguration](https://stash.run/docs/latest/concepts/crds/backupconfiguration/)
- [BackupSession](https://stash.run/docs/latest/concepts/crds/backupsession/)
- [Repository](https://stash.run/docs/latest/concepts/crds/repository/)
- [Function](https://stash.run/docs/latest/concepts/crds/function/)
- [Task](https://stash.run/docs/latest/concepts/crds/task/)

In this tutorial, we are going to show backup of three different MongoDB databases on three different namespaces named `demo`, `demo-2`, and `demo-3`. Create the namespaces as below if you haven't done it already.

```bash
❯ kubectl create ns demo
namespace/demo created

❯ kubectl create ns demo-2
namespace/demo-2 created

❯ kubectl create ns demo-3
namespace/demo-3 created
```

When you install the Stash, it automatically installs all the official database addons. Verify that it has installed the MongoDB addons using the following command.

```bash
❯ kubectl get tasks.stash.appscode.com | grep mongodb
mongodb-backup-3.4.17          23h
mongodb-backup-3.4.22          23h
mongodb-backup-3.6.13          23h
mongodb-backup-3.6.8           23h
mongodb-backup-4.0.11          23h
mongodb-backup-4.0.3           23h
mongodb-backup-4.0.5           23h
mongodb-backup-4.1.13          23h
mongodb-backup-4.1.4           23h
mongodb-backup-4.1.7           23h
mongodb-backup-4.4.6           23h
mongodb-backup-4.4.6           23h
mongodb-backup-5.0.3           23h
mongodb-restore-3.4.17         23h
mongodb-restore-3.4.22         23h
mongodb-restore-3.6.13         23h
mongodb-restore-3.6.8          23h
mongodb-restore-4.0.11         23h
mongodb-restore-4.0.3          23h
mongodb-restore-4.0.5          23h
mongodb-restore-4.1.13         23h
mongodb-restore-4.1.4          23h
mongodb-restore-4.1.7          23h
mongodb-restore-4.4.6          23h
mongodb-restore-4.4.6          23h
mongodb-restore-5.0.3          23h

```

## Prepare Backup Blueprint

To backup an MongoDB database using Stash, you have to create a `Secret` containing the backend credentials, a `Repository` containing the backend information, and a `BackupConfiguration` containing the schedule and target information. A `BackupBlueprint` allows you to specify a template for the `Repository` and the `BackupConfiguration`.

The `BackupBlueprint` is a non-namespaced CRD. So, once you have created a `BackupBlueprint`, you can use it to backup any MongoDB database of any namespace just by creating the storage `Secret` in that namespace and adding few annotations to your MongoDB CRO. Then, Stash will automatically create a `Repository` and a `BackupConfiguration` according to the template to backup the database.

Below is the `BackupBlueprint` object that we are going to use in this tutorial,

```yaml
apiVersion: stash.appscode.com/v1beta1
kind: BackupBlueprint
metadata:
  name: mongodb-backup-template
spec:
  # ============== Blueprint for Repository ==========================
  backend:
    gcs:
      bucket: stash-testing
      prefix: mongodb-backup/${TARGET_NAMESPACE}/${TARGET_APP_RESOURCE}/${TARGET_NAME}
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
❯ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mongodb/backup/auto-backup/examples/backupblueprint.yaml
backupblueprint.stash.appscode.com/mongodb-backup-template created
```

Now, we are ready to backup our MongoDB databases using few annotations. You can check available auto-backup annotations for a databases from [here](https://stash.run/docs/latest/guides/auto-backup/database/#available-auto-backup-annotations-for-database).

## Auto-backup with default configurations

In this section, we are going to backup a MongoDB database from `demo` namespace and we are going to use the default configurations specified in the `BackupBlueprint`.

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

Now, we are going to create an MongoDB CRO in `demo` namespace. Below is the YAML of the MongoDB object that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: MongoDB
metadata:
  name: sample-mongodb
  namespace: demo
  annotations:
    stash.appscode.com/backup-blueprint: mongodb-backup-template
spec:
  version: "4.4.26"
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Notice the `annotations` section. We are pointing to the `BackupBlueprint` that we have created earlier though `stash.appscode.com/backup-blueprint` annotation. Stash will watch this annotation and create a `Repository` and a `BackupConfiguration` according to the `BackupBlueprint`.

Let's create the above MongoDB CRO,

```bash
❯ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mongob/backup/auto-backup/examples/sample-mongodb.yaml
mongodb.kubedb.com/sample-mongodb created
```

### Verify Auto-backup configured

In this section, we are going to verify whether Stash has created the respective `Repository` and `BackupConfiguration` for our MongoDB database we have just deployed or not.

#### Verify Repository

At first, let's verify whether Stash has created a `Repository` for our MongoDB or not.

```bash
❯ kubectl get repository -n demo
NAME                 INTEGRITY   SIZE   SNAPSHOT-COUNT   LAST-SUCCESSFUL-BACKUP   AGE
app-sample-mongodb                                                                10s 
```

Now, let's check the YAML of the `Repository`.

```yaml
❯ kubectl get repository -n demo app-sample-mongodb -o yaml
apiVersion: stash.appscode.com/v1alpha1
kind: Repository
metadata:
  creationTimestamp: "2022-02-02T05:49:00Z"
  finalizers:
  - stash
  generation: 1
  name: app-sample-mongodb
  namespace: demo
  resourceVersion: "283554"
  uid: d025358c-2f60-4d35-8efb-27c42439d28e
spec:
  backend:
    gcs:
      bucket: stash-testing
      prefix: mongodb-backup/demo/mongodb/sample-mongodb
    storageSecretName: gcs-secret
```

Here, you can see that Stash has resolved the variables in `prefix` field and substituted them with the equivalent information from this database.

#### Verify BackupConfiguration

If everything goes well, Stash should create a `BackupConfiguration` for our MongoDB in `demo` namespace and the phase of that `BackupConfiguration` should be `Ready`. Verify the `BackupConfiguration` crd by the following command,

```bash
❯ kubectl get backupconfiguration -n demo
NAMESPACE   NAME                 TASK   SCHEDULE      PAUSED   PHASE   AGE
demo        app-sample-mongodb          */5 * * * *            Ready   4m11s
```

Now, let's check the YAML of the `BackupConfiguration`.

```yaml
❯ kubectl get backupconfiguration -n demo app-sample-mongodb -o yaml
apiVersion: stash.appscode.com/v1beta1
kind: BackupConfiguration
metadata:
  creationTimestamp: "2022-02-02T05:49:00Z"
  finalizers:
  - stash.appscode.com
  generation: 1
  name: app-sample-mongodb
  namespace: demo
  ownerReferences:
  - apiVersion: appcatalog.appscode.com/v1alpha1
    blockOwnerDeletion: true
    controller: true
    kind: AppBinding
    name: sample-mongodb
    uid: 481ea54c-5a77-43a9-8230-f906f9d240bf
  resourceVersion: "283559"
  uid: aa2a1195-8ed7-4238-b807-66fb5b09505f
spec:
  driver: Restic
  repository:
    name: app-sample-mongodb
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
      name: sample-mongodb
  task: {}
  tempDir: {}
status:
  conditions:
  - lastTransitionTime: "2022-02-02T05:49:00Z"
    message: Repository demo/app-sample-mongodb exist.
    reason: RepositoryAvailable
    status: "True"
    type: RepositoryFound
  - lastTransitionTime: "2022-02-02T05:49:00Z"
    message: Backend Secret demo/ does not exist.
    reason: BackendSecretNotAvailable
    status: "False"
    type: BackendSecretFound
  observedGeneration: 1
```

Notice the `target` section. Stash has automatically added the MongoDB as the target of this `BackupConfiguration`.

#### Verify Backup

Now, let's wait for a backup run to complete. You can watch for `BackupSession` as below,

```bash
❯ kubectl get backupsession -n demo -w
NAME                            INVOKER-TYPE          INVOKER-NAME         PHASE      DURATION   AGE
app-sample-mongodb-1643781603   BackupConfiguration   app-sample-mongodb   Running               30s
app-sample-mongodb-1643781603   BackupConfiguration   app-sample-mongodb   Succeeded  31s        30s

```

Once the backup has been completed successfully, you should see the backed up data has been stored in the bucket at the directory pointed by the `prefix` field of the `Repository`.

<figure align="center">
  <img alt="Backup data in GCS Bucket" src="/docs/guides/mongodb/backup/auto-backup/images/sample-mongodb.png">
  <figcaption align="center">Fig: Backup data in GCS Bucket</figcaption>
</figure>

## Auto-backup with a custom schedule

In this section, we are going to backup an MongoDB database from `demo-2` namespace. This time, we are going to overwrite the default schedule used in the `BackupBlueprint`.

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

Now, we are going to create an MongoDB CRO in `demo-2` namespace. Below is the YAML of the MongoDB object that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: MongoDB
metadata:
  name: sample-mongodb-2
  namespace: demo-2
  annotations:
    stash.appscode.com/backup-blueprint: mongodb-backup-template
    stash.appscode.com/schedule: "*/3 * * * *"
spec:
  version: "4.4.26"
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Notice the `annotations` section. This time, we have passed a schedule via `stash.appscode.com/schedule` annotation along with the `stash.appscode.com/backup-blueprint` annotation.

Let's create the above MongoDB CRO,

```bash
❯ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mongodb/backup/auto-backup/examples/sample-mongodb-2.yaml
mongodb.kubedb.com/sample-mongodb-2 created
```

### Verify Auto-backup configured

Now, let's verify whether the auto-backup has been configured properly or not.

#### Verify Repository

At first, let's verify whether Stash has created a `Repository` for our MongoDB or not.

```bash
❯ kubectl get repository -n demo-2
NAME                   INTEGRITY   SIZE   SNAPSHOT-COUNT   LAST-SUCCESSFUL-BACKUP   AGE
app-sample-mongodb-2                                                                4s
```

Now, let's check the YAML of the `Repository`.

```yaml
❯ kubectl get repository -n demo-2 app-sample-mongodb-2  -o yaml
apiVersion: stash.appscode.com/v1alpha1
kind: Repository
metadata:
  creationTimestamp: "2022-02-02T06:19:21Z"
  finalizers:
  - stash
  generation: 1
  name: app-sample-mongodb-2
  namespace: demo-2
  resourceVersion: "286925"
  uid: e1948d2d-2a15-41ea-99f9-5b59394c10c1
spec:
  backend:
    gcs:
      bucket: stash-testing
      prefix: mongodb-backup/demo-2/mongodb/sample-mongodb-2
    storageSecretName: gcs-secret
```

Here, you can see that Stash has resolved the variables in `prefix` field and substituted them with the equivalent information from this new database.

#### Verify BackupConfiguration

If everything goes well, Stash should create a `BackupConfiguration` for our MongoDB in `demo-2` namespace and the phase of that `BackupConfiguration` should be `Ready`. Verify the `BackupConfiguration` crd by the following command,

```bash
❯ kubectl get backupconfiguration -n demo-2
NAME                   TASK                    SCHEDULE      PAUSED   PHASE   AGE
app-sample-mongodb-2   mongodb-backup-10.5.23   */3 * * * *            Ready   3m24s
```

Now, let's check the YAML of the `BackupConfiguration`.

```yaml
❯ kubectl get backupconfiguration -n demo-2 app-sample-mongodb-2 -o yaml
apiVersion: stash.appscode.com/v1beta1
kind: BackupConfiguration
metadata:
  creationTimestamp: "2022-02-02T06:19:21Z"
  finalizers:
  - stash.appscode.com
  generation: 1
  name: app-sample-mongodb-2
  namespace: demo-2
  ownerReferences:
  - apiVersion: appcatalog.appscode.com/v1alpha1
    blockOwnerDeletion: true
    controller: true
    kind: AppBinding
    name: sample-mongodb-2
    uid: 7c18485f-ed8e-4c01-b160-3bbc4e5049db
  resourceVersion: "286938"
  uid: 279c0471-0618-4b73-85d0-edd70ec2e132
spec:
  driver: Restic
  repository:
    name: app-sample-mongodb-2
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
      name: sample-mongodb-2
  task: {}
  tempDir: {}
status:
  conditions:
  - lastTransitionTime: "2022-02-02T06:19:21Z"
    message: Repository demo-2/app-sample-mongodb-2 exist.
    reason: RepositoryAvailable
    status: "True"
    type: RepositoryFound
  - lastTransitionTime: "2022-02-02T06:19:21Z"
    message: Backend Secret demo-2/gcs-secret exist.
    reason: BackendSecretAvailable
    status: "True"
    type: BackendSecretFound
  - lastTransitionTime: "2022-02-02T06:19:21Z"
    message: Backup target appcatalog.appscode.com/v1alpha1 appbinding/sample-mongodb-2
      found.
    reason: TargetAvailable
    status: "True"
    type: BackupTargetFound
  - lastTransitionTime: "2022-02-02T06:19:21Z"
    message: Successfully created backup triggering CronJob.
    reason: CronJobCreationSucceeded
    status: "True"
    type: CronJobCreated
  observedGeneration: 1
```

Notice the `schedule` section. This time the `BackupConfiguration` has been created with the schedule we have provided via annotation.

Also, notice the `target` section. Stash has automatically added the new MongoDB as the target of this `BackupConfiguration`.

#### Verify Backup

Now, let's wait for a backup run to complete. You can watch for `BackupSession` as below,

```bash
❯ kubectl get backupsession -n demo-2 -w
NAME                              INVOKER-TYPE          INVOKER-NAME           PHASE       DURATION   AGE
app-sample-mongodb-2-1643782861   BackupConfiguration   app-sample-mongodb-2   Succeeded   31s        2m17s
```

Once the backup has been completed successfully, you should see that Stash has created a new directory as pointed by the `prefix` field of the new `Repository` and stored the backed up data there.

<figure align="center">
  <img alt="Backup data in GCS Bucket" src="/docs/guides/mongodb/backup/auto-backup/images/sample-mongodb-2.png">
  <figcaption align="center">Fig: Backup data in GCS Bucket</figcaption>
</figure>

## Auto-backup with custom parameters

In this section, we are going to backup an MongoDB database of `demo-3` namespace. This time, we are going to pass some parameters for the Task through the annotations.

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

Now, we are going to create an MongoDB CRO in `demo-3` namespace. Below is the YAML of the MongoDB object that we are going to create,

```yaml
apiVersion: kubedb.com/v1
kind: MongoDB
metadata:
  name: sample-mongodb-3
  namespace: demo-3
  annotations:
    stash.appscode.com/backup-blueprint: mongodb-backup-template
    params.stash.appscode.com/args: "--db=testdb"
spec:
  version: "4.4.26"
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  deletionPolicy: WipeOut
```

Notice the `annotations` section. This time, we have passed an argument via `params.stash.appscode.com/args` annotation along with the `stash.appscode.com/backup-blueprint` annotation.

Let's create the above MongoDB CRO,

```bash
❯ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mongodb/backup/auto-backup/examples/sample-mongob-3.yaml
mongob.kubedb.com/sample-mongodb-3 created
```

### Verify Auto-backup configured

Now, let's verify whether the auto-backup resources has been created or not.

#### Verify Repository

At first, let's verify whether Stash has created a `Repository` for our MongoDB or not.

```bash
❯ kubectl get repository -n demo-3
NAME                   INTEGRITY   SIZE   SNAPSHOT-COUNT   LAST-SUCCESSFUL-BACKUP   AGE
app-sample-mongodb-3                                                                8s
```

Now, let's check the YAML of the `Repository`.

```yaml
❯ kubectl get repository -n demo-3 app-sample-mongodb-3 -o yaml
apiVersion: stash.appscode.com/v1alpha1
kind: Repository
metadata:
  creationTimestamp: "2022-02-02T06:45:56Z"
  finalizers:
  - stash
  generation: 1
  name: app-sample-mongodb-3
  namespace: demo-3
  resourceVersion: "302950"
  uid: 00b74653-fd08-42ba-a699-1b012e1e7da8
spec:
  backend:
    gcs:
      bucket: stash-testing
      prefix: mongodb-backup/demo-3/mongodb/sample-mongodb-3
    storageSecretName: gcs-secret
```

Here, you can see that Stash has resolved the variables in `prefix` field and substituted them with the equivalent information from this new database.

#### Verify BackupConfiguration

If everything goes well, Stash should create a `BackupConfiguration` for our MongoDB in `demo-3` namespace and the phase of that `BackupConfiguration` should be `Ready`. Verify the `BackupConfiguration` crd by the following command,

```bash
❯ kubectl get backupconfiguration -n demo-3
NAME                   TASK                    SCHEDULE      PAUSED   PHASE   AGE
app-sample-mongodb-3   mongodb-backup-10.5.23   */5 * * * *            Ready   106s
```

Now, let's check the YAML of the `BackupConfiguration`.

```yaml
❯ kubectl get backupconfiguration -n demo-3 app-sample-mongodb-3 -o yaml
apiVersion: stash.appscode.com/v1beta1
kind: BackupConfiguration
metadata:
  creationTimestamp: "2022-02-02T08:29:43Z"
  finalizers:
  - stash.appscode.com
  generation: 1
  name: app-sample-mongodb-3
  namespace: demo-3
  ownerReferences:
  - apiVersion: appcatalog.appscode.com/v1alpha1
    blockOwnerDeletion: true
    controller: true
    kind: AppBinding
    name: sample-mongodb-3
    uid: 54deac95-790b-4fc1-93ec-fd3758cac71e
  resourceVersion: "301618"
  uid: 6ecb511e-1c6c-4d0b-b241-277c0b0d1059
spec:
  driver: Restic
  repository:
    name: app-sample-mongodb-3
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
      name: sample-mongodb-3
  task:
    params:
    - name: args
      value: --db=testdb
  tempDir: {}
status:
  conditions:
  - lastTransitionTime: "2022-02-02T08:29:43Z"
    message: Repository demo-3/app-sample-mongodb-3 exist.
    reason: RepositoryAvailable
    status: "True"
    type: RepositoryFound
  - lastTransitionTime: "2022-02-02T08:29:43Z"
    message: Backend Secret demo-3/gcs-secret exist.
    reason: BackendSecretAvailable
    status: "True"
    type: BackendSecretFound
  - lastTransitionTime: "2022-02-02T08:29:43Z"
    message: Backup target appcatalog.appscode.com/v1alpha1 appbinding/sample-mongodb-3
      found.
    reason: TargetAvailable
    status: "True"
    type: BackupTargetFound
  - lastTransitionTime: "2022-02-02T08:29:43Z"
    message: Successfully created backup triggering CronJob.
    reason: CronJobCreationSucceeded
    status: "True"
    type: CronJobCreated
  observedGeneration: 1
```

Notice the `task` section. The `args` parameter that we had passed via annotations has been added to the `params` section.

Also, notice the `target` section. Stash has automatically added the new MongoDB as the target of this `BackupConfiguration`.

#### Verify Backup

Now, let's wait for a backup run to complete. You can watch for `BackupSession` as below,

```bash
❯ kubectl get backupsession -n demo-3 -w
NAME                              INVOKER-TYPE          INVOKER-NAME           PHASE       DURATION   AGE
app-sample-mongodb-3-1643792101   BackupConfiguration   app-sample-mongodb-3   Succeeded   39s        118s

```

Once the backup has been completed successfully, you should see that Stash has created a new directory as pointed by the `prefix` field of the new `Repository` and stored the backed up data there.

<figure align="center">
  <img alt="Backup data in GCS Bucket" src="/docs/guides/mongodb/backup/auto-backup/images/sample-mongodb-3.png">
  <figcaption align="center">Fig: Backup data in GCS Bucket</figcaption>
</figure>

## Cleanup

To cleanup the resources crated by this tutorial, run the following commands,

```bash
❯ kubectl delete -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mongodb/backup/auto-backup/examples/
backupblueprint.stash.appscode.com "mongodb-backup-template" deleted
mongodb.kubedb.com "sample-mongodb-2" deleted
mongodb.kubedb.com "sample-mongodb-3" deleted
mongodb.kubedb.com "sample-mongodb" deleted

❯ kubectl delete repository -n demo --all
repository.stash.appscode.com "app-sample-mongodb" deleted
❯ kubectl delete repository -n demo-2 --all
repository.stash.appscode.com "app-sample-mongodb-2" deleted
❯ kubectl delete repository -n demo-3 --all
repository.stash.appscode.com "app-sample-mongodb-3" deleted
```
