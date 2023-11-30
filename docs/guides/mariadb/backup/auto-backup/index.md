---
title: MariaDB Auto-Backup | Stash
description: Backup MariaDB using Stash Auto-Backup
menu:
  docs_{{ .version }}:
    identifier: guides-mariadb-backup-auto-backup
    name: Auto-Backup
    parent: guides-mariadb-backup
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Backup MariaDB using Stash Auto-Backup

Stash can be configured to automatically backup any MariaDB database in your cluster. Stash enables cluster administrators to deploy backup blueprints ahead of time so that the database owners can easily backup their database with just a few annotations.

In this tutorial, we are going to show how you can configure a backup blueprint for MariaDB databases in your cluster and backup them with few annotations.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.
- Install Stash Enterprise in your cluster following the steps [here](https://stash.run/docs/latest/setup/install/enterprise/).
- Install KubeDB in your cluster following the steps [here](/docs/setup/README.md).
- If you are not familiar with how Stash backup and restore MariaDB databases, please check the following guide [here](/docs/guides/mariadb/backup/overview/index.md).
- If you are not familiar with how auto-backup works in Stash, please check the following guide [here](https://stash.run/docs/latest/guides/auto-backup/overview/).
- If you are not familiar with the available auto-backup options for databases in Stash, please check the following guide [here](https://stash.run/docs/latest/guides/auto-backup/database/).

You should be familiar with the following `Stash` concepts:

- [BackupBlueprint](https://stash.run/docs/latest/concepts/crds/backupblueprint/)
- [BackupConfiguration](https://stash.run/docs/latest/concepts/crds/backupconfiguration/)
- [BackupSession](https://stash.run/docs/latest/concepts/crds/backupsession/)
- [Repository](https://stash.run/docs/latest/concepts/crds/repository/)
- [Function](https://stash.run/docs/latest/concepts/crds/function/)
- [Task](https://stash.run/docs/latest/concepts/crds/task/)

In this tutorial, we are going to show backup of three different MariaDB databases on three different namespaces named `demo`, `demo-2`, and `demo-3`. Create the namespaces as below if you haven't done it already.

```bash
❯ kubectl create ns demo
namespace/demo created

❯ kubectl create ns demo-2
namespace/demo-2 created

❯ kubectl create ns demo-3
namespace/demo-3 created
```

When you install the Stash Enterprise edition, it automatically installs all the official database addons. Verify that it has installed the MariaDB addons using the following command.

```bash
❯ kubectl get tasks.stash.appscode.com | grep mariadb
mariadb-backup-10.5.23    62m
mariadb-restore-10.5.23   62m
```

## Prepare Backup Blueprint

To backup an MariaDB database using Stash, you have to create a `Secret` containing the backend credentials, a `Repository` containing the backend information, and a `BackupConfiguration` containing the schedule and target information. A `BackupBlueprint` allows you to specify a template for the `Repository` and the `BackupConfiguration`.

The `BackupBlueprint` is a non-namespaced CRD. So, once you have created a `BackupBlueprint`, you can use it to backup any MariaDB database of any namespace just by creating the storage `Secret` in that namespace and adding few annotations to your MariaDB CRO. Then, Stash will automatically create a `Repository` and a `BackupConfiguration` according to the template to backup the database.

Below is the `BackupBlueprint` object that we are going to use in this tutorial,

```yaml
apiVersion: stash.appscode.com/v1beta1
kind: BackupBlueprint
metadata:
  name: mariadb-backup-template
spec:
  # ============== Blueprint for Repository ==========================
  backend:
    gcs:
      bucket: stash-testing
      prefix: mariadb-backup/${TARGET_NAMESPACE}/${TARGET_APP_RESOURCE}/${TARGET_NAME}
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
❯ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/backup/auto-backup/examples/backupblueprint.yaml
backupblueprint.stash.appscode.com/mariadb-backup-template created
```

Now, we are ready to backup our MariaDB databases using few annotations. You can check available auto-backup annotations for a databases from [here](https://stash.run/docs/latest/guides/auto-backup/database/#available-auto-backup-annotations-for-database).

## Auto-backup with default configurations

In this section, we are going to backup an MariaDB database of `demo` namespace. We are going to use the default configurations specified in the `BackupBlueprint`.

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

Now, we are going to create an MariaDB CRO in `demo` namespace. Below is the YAML of the MariaDB object that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MariaDB
metadata:
  name: sample-mariadb
  namespace: demo
  annotations:
    stash.appscode.com/backup-blueprint: mariadb-backup-template
spec:
  version: "10.5.23"
  replicas: 1
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: WipeOut
```

Notice the `annotations` section. We are pointing to the `BackupBlueprint` that we have created earlier though `stash.appscode.com/backup-blueprint` annotation. Stash will watch this annotation and create a `Repository` and a `BackupConfiguration` according to the `BackupBlueprint`.

Let's create the above MariaDB CRO,

```bash
❯ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/backup/auto-backup/examples/sample-mariadb.yaml
mariadb.kubedb.com/sample-mariadb created
```

### Verify Auto-backup configured

In this section, we are going to verify whether Stash has created the respective `Repository` and `BackupConfiguration` for our MariaDB database we have just deployed or not.

#### Verify Repository

At first, let's verify whether Stash has created a `Repository` for our MariaDB or not.

```bash
❯ kubectl get repository -n demo
NAME                 INTEGRITY   SIZE   SNAPSHOT-COUNT   LAST-SUCCESSFUL-BACKUP   AGE
app-sample-mariadb                                                                10s 
```

Now, let's check the YAML of the `Repository`.

```yaml
❯ kubectl get repository -n demo app-sample-mariadb -o yaml
apiVersion: stash.appscode.com/v1alpha1
kind: Repository
metadata:
...  
  name: app-sample-mariadb
  namespace: demo
spec:
  backend:
    gcs:
      bucket: stash-testing
      prefix: mariadb-backup/demo/mariadb/sample-mariadb
    storageSecretName: gcs-secret
```

Here, you can see that Stash has resolved the variables in `prefix` field and substituted them with the equivalent information from this database.

#### Verify BackupConfiguration

If everything goes well, Stash should create a `BackupConfiguration` for our MariaDB in `demo` namespace and the phase of that `BackupConfiguration` should be `Ready`. Verify the `BackupConfiguration` crd by the following command,

```bash
❯ kubectl get backupconfiguration -n demo
NAME                 TASK                    SCHEDULE      PAUSED   PHASE   AGE
app-sample-mariadb   mariadb-backup-10.5.23   */5 * * * *            Ready   7m28s
```

Now, let's check the YAML of the `BackupConfiguration`.

```yaml
❯ kubectl get backupconfiguration -n demo app-sample-mariadb -o yaml
apiVersion: stash.appscode.com/v1beta1
kind: BackupConfiguration
metadata:
  name: app-sample-mariadb
  namespace: demo
  ...
  spec:
  driver: Restic
  repository:
    name: app-sample-mariadb
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
      name: sample-mariadb
  tempDir: {}
status:
  conditions:
  - lastTransitionTime: "2021-02-25T05:14:51Z"
    message: Repository demo/app-sample-mariadb exist.
    reason: RepositoryAvailable
    status: "True"
    type: RepositoryFound
  - lastTransitionTime: "2021-02-25T05:14:51Z"
    message: Backend Secret demo/gcs-secret exist.
    reason: BackendSecretAvailable
    status: "True"
    type: BackendSecretFound
  - lastTransitionTime: "2021-02-25T05:14:51Z"
    message: Backup target appcatalog.appscode.com/v1alpha1 appbinding/sample-mariadb
      found.
    reason: TargetAvailable
    status: "True"
    type: BackupTargetFound
  - lastTransitionTime: "2021-02-25T05:14:51Z"
    message: Successfully created backup triggering CronJob.
    reason: CronJobCreationSucceeded
    status: "True"
    type: CronJobCreated
  observedGeneration: 1

```

Notice the `target` section. Stash has automatically added the MariaDB as the target of this `BackupConfiguration`.

#### Verify Backup

Now, let's wait for a backup run to complete. You can watch for `BackupSession` as below,

```bash
❯ kubectl get backupsession -n demo -w
NAME                            INVOKER-TYPE          INVOKER-NAME         PHASE       AGE
app-sample-mariadb-1614230401   BackupConfiguration   app-sample-mariadb   Succeeded   5m40s
app-sample-mariadb-1614230701   BackupConfiguration   app-sample-mariadb   Running     39s
```

Once the backup has been completed successfully, you should see the backed up data has been stored in the bucket at the directory pointed by the `prefix` field of the `Repository`.

<figure align="center">
  <img alt="Backup data in GCS Bucket" src="/docs/guides/mariadb/backup/auto-backup/images/sample-mariadb.png">
  <figcaption align="center">Fig: Backup data in GCS Bucket</figcaption>
</figure>

## Auto-backup with a custom schedule

In this section, we are going to backup an MariaDB database of `demo-2` namespace. This time, we are going to overwrite the default schedule used in the `BackupBlueprint`.

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

Now, we are going to create an MariaDB CRO in `demo-2` namespace. Below is the YAML of the MariaDB object that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MariaDB
metadata:
  name: sample-mariadb-2
  namespace: demo-2
  annotations:
    stash.appscode.com/backup-blueprint: mariadb-backup-template
    stash.appscode.com/schedule: "*/3 * * * *"
spec:
  version: "10.5.23"
  replicas: 1
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: WipeOut
```

Notice the `annotations` section. This time, we have passed a schedule via `stash.appscode.com/schedule` annotation along with the `stash.appscode.com/backup-blueprint` annotation.

Let's create the above MariaDB CRO,

```bash
❯ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/backup/auto-backup/examples/sample-mariadb-2.yaml
mariadb.kubedb.com/sample-mariadb-2 created
```

### Verify Auto-backup configured

Now, let's verify whether the auto-backup has been configured properly or not.

#### Verify Repository

At first, let's verify whether Stash has created a `Repository` for our MariaDB or not.

```bash
❯ kubectl get repository -n demo-2
NAME                   INTEGRITY   SIZE   SNAPSHOT-COUNT   LAST-SUCCESSFUL-BACKUP   AGE
app-sample-mariadb-2                                                                4s
```

Now, let's check the YAML of the `Repository`.

```yaml
❯ kubectl get repository -n demo-2 app-sample-mariadb-2  -o yaml
apiVersion: stash.appscode.com/v1alpha1
kind: Repository
metadata:
  name: app-sample-mariadb-2
  namespace: demo-2
  ...
spec:
  backend:
    gcs:
      bucket: stash-testing
      prefix: mariadb-backup/demo-2/mariadb/sample-mariadb-2
    storageSecretName: gcs-secret
```

Here, you can see that Stash has resolved the variables in `prefix` field and substituted them with the equivalent information from this new database.

#### Verify BackupConfiguration

If everything goes well, Stash should create a `BackupConfiguration` for our MariaDB in `demo-2` namespace and the phase of that `BackupConfiguration` should be `Ready`. Verify the `BackupConfiguration` crd by the following command,

```bash
❯ kubectl get backupconfiguration -n demo-2
NAME                   TASK                    SCHEDULE      PAUSED   PHASE   AGE
app-sample-mariadb-2   mariadb-backup-10.5.23   */3 * * * *            Ready   3m24s
```

Now, let's check the YAML of the `BackupConfiguration`.

```yaml
❯ kubectl get backupconfiguration -n demo-2 app-sample-mariadb-2 -o yaml

apiVersion: stash.appscode.com/v1beta1
kind: BackupConfiguration
metadata:
  name: app-sample-mariadb-2
  namespace: demo-2
  ...
  ownerReferences:
  - apiVersion: appcatalog.appscode.com/v1alpha1
    blockOwnerDeletion: true
    controller: true
    kind: AppBinding
    name: sample-mariadb-2
    uid: 7cbdf140-5fd1-487a-b04f-1847def418e8
  resourceVersion: "56888"
  selfLink: /apis/stash.appscode.com/v1beta1/namespaces/demo-2/backupconfigurations/app-sample-mariadb-2
  uid: e85dd3db-fa41-48b8-b253-5731ee8cc956
spec:
  driver: Restic
  repository:
    name: app-sample-mariadb-2
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
      name: sample-mariadb-2
  tempDir: {}
status:
  conditions:
  - lastTransitionTime: "2021-02-25T06:10:14Z"
    message: Repository demo-2/app-sample-mariadb-2 exist.
    reason: RepositoryAvailable
    status: "True"
    type: RepositoryFound
  - lastTransitionTime: "2021-02-25T06:10:14Z"
    message: Backend Secret demo-2/gcs-secret exist.
    reason: BackendSecretAvailable
    status: "True"
    type: BackendSecretFound
  - lastTransitionTime: "2021-02-25T06:10:14Z"
    message: Backup target appcatalog.appscode.com/v1alpha1 appbinding/sample-mariadb-2
      found.
    reason: TargetAvailable
    status: "True"
    type: BackupTargetFound
  - lastTransitionTime: "2021-02-25T06:10:14Z"
    message: Successfully created backup triggering CronJob.
    reason: CronJobCreationSucceeded
    status: "True"
    type: CronJobCreated
  observedGeneration: 1
```

Notice the `schedule` section. This time the `BackupConfiguration` has been created with the schedule we have provided via annotation.

Also, notice the `target` section. Stash has automatically added the new MariaDB as the target of this `BackupConfiguration`.

#### Verify Backup

Now, let's wait for a backup run to complete. You can watch for `BackupSession` as below,

```bash
❯ kubectl get backupsession -n demo-2 -w
NAME                              INVOKER-TYPE          INVOKER-NAME           PHASE       AGE
app-sample-mariadb-2-1614233715   BackupConfiguration   app-sample-mariadb-2   Succeeded   3m2s
app-sample-mariadb-2-1614233880   BackupConfiguration   app-sample-mariadb-2   Running     17s
```

Once the backup has been completed successfully, you should see that Stash has created a new directory as pointed by the `prefix` field of the new `Repository` and stored the backed up data there.

<figure align="center">
  <img alt="Backup data in GCS Bucket" src="/docs/guides/mariadb/backup/auto-backup/images/sample-mariadb-2.png">
  <figcaption align="center">Fig: Backup data in GCS Bucket</figcaption>
</figure>

## Auto-backup with custom parameters

In this section, we are going to backup an MariaDB database of `demo-3` namespace. This time, we are going to pass some parameters for the Task through the annotations.

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

Now, we are going to create an MariaDB CRO in `demo-3` namespace. Below is the YAML of the MariaDB object that we are going to create,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: MariaDB
metadata:
  name: sample-mariadb-3
  namespace: demo-3
  annotations:
    stash.appscode.com/backup-blueprint: mariadb-backup-template
    params.stash.appscode.com/args: --databases mysql
spec:
  version: "10.5.23"
  replicas: 1
  storageType: Durable
  storage:
    storageClassName: "standard"
    accessModes:
    - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  terminationPolicy: WipeOut

```

Notice the `annotations` section. This time, we have passed an argument via `params.stash.appscode.com/args` annotation along with the `stash.appscode.com/backup-blueprint` annotation.

Let's create the above MariaDB CRO,

```bash
❯ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/backup/auto-backup/examples/sample-mariadb-3.yaml
mariadb.kubedb.com/sample-mariadb-3 created
```

### Verify Auto-backup configured

Now, let's verify whether the auto-backup resources has been created or not.

#### Verify Repository

At first, let's verify whether Stash has created a `Repository` for our MariaDB or not.

```bash
❯ kubectl get repository -n demo-3
NAME                   INTEGRITY   SIZE   SNAPSHOT-COUNT   LAST-SUCCESSFUL-BACKUP   AGE
app-sample-mariadb-3                                                                8s
```

Now, let's check the YAML of the `Repository`.

```yaml
❯ kubectl get repository -n demo-3 app-sample-mariadb-3 -o yaml
apiVersion: stash.appscode.com/v1alpha1
kind: Repository
metadata:
  name: app-sample-mariadb-3
  namespace: demo-3
  ...
spec:
  backend:
    gcs:
      bucket: stash-testing
      prefix: mariadb-backup/demo-3/mariadb/sample-mariadb-3
    storageSecretName: gcs-secret
```

Here, you can see that Stash has resolved the variables in `prefix` field and substituted them with the equivalent information from this new database.

#### Verify BackupConfiguration

If everything goes well, Stash should create a `BackupConfiguration` for our MariaDB in `demo-3` namespace and the phase of that `BackupConfiguration` should be `Ready`. Verify the `BackupConfiguration` crd by the following command,

```bash
❯ kubectl get backupconfiguration -n demo-3
NAME                   TASK                    SCHEDULE      PAUSED   PHASE   AGE
app-sample-mariadb-3   mariadb-backup-10.5.23   */5 * * * *            Ready   106s
```

Now, let's check the YAML of the `BackupConfiguration`.

```yaml
❯ kubectl get backupconfiguration -n demo-3 app-sample-mariadb-3 -o yaml
apiVersion: stash.appscode.com/v1beta1
kind: BackupConfiguration
metadata:
  name: app-sample-mariadb-3
  namespace: demo-3
  ...
spec:
  driver: Restic
  repository:
    name: app-sample-mariadb-3
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
      name: sample-mariadb-3
  task:
    params:
    - name: args
      value: --databases mysql
  tempDir: {}
status:
  conditions:
  - lastTransitionTime: "2021-02-25T11:58:12Z"
    message: Repository demo-3/app-sample-mariadb-3 exist.
    reason: RepositoryAvailable
    status: "True"
    type: RepositoryFound
  - lastTransitionTime: "2021-02-25T11:58:12Z"
    message: Backend Secret demo-3/gcs-secret exist.
    reason: BackendSecretAvailable
    status: "True"
    type: BackendSecretFound
  - lastTransitionTime: "2021-02-25T11:58:12Z"
    message: Backup target appcatalog.appscode.com/v1alpha1 appbinding/sample-mariadb-3
      found.
    reason: TargetAvailable
    status: "True"
    type: BackupTargetFound
  - lastTransitionTime: "2021-02-25T11:58:12Z"
    message: Successfully created backup triggering CronJob.
    reason: CronJobCreationSucceeded
    status: "True"
    type: CronJobCreated
  observedGeneration: 1
```

Notice the `task` section. The `args` parameter that we had passed via annotations has been added to the `params` section.

Also, notice the `target` section. Stash has automatically added the new MariaDB as the target of this `BackupConfiguration`.

#### Verify Backup

Now, let's wait for a backup run to complete. You can watch for `BackupSession` as below,

```bash
❯ kubectl get backupsession -n demo-3 -w
NAME                              INVOKER-TYPE          INVOKER-NAME           PHASE       AGE
app-sample-mariadb-3-1614254408   BackupConfiguration   app-sample-mariadb-3   Succeeded   5m23s
app-sample-mariadb-3-1614254708   BackupConfiguration   app-sample-mariadb-3   Running     23s
```

Once the backup has been completed successfully, you should see that Stash has created a new directory as pointed by the `prefix` field of the new `Repository` and stored the backed up data there.

<figure align="center">
  <img alt="Backup data in GCS Bucket" src="/docs/guides/mariadb/backup/auto-backup/images/sample-mariadb-3.png">
  <figcaption align="center">Fig: Backup data in GCS Bucket</figcaption>
</figure>

## Cleanup

To cleanup the resources crated by this tutorial, run the following commands,

```bash
❯ kubectl delete -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/mariadb/backup/auto-backup/examples/
backupblueprint.stash.appscode.com "mariadb-backup-template" deleted
mariadb.kubedb.com "sample-mariadb-2" deleted
mariadb.kubedb.com "sample-mariadb-3" deleted
mariadb.kubedb.com "sample-mariadb" deleted

❯ kubectl delete repository -n demo --all
repository.stash.appscode.com "app-sample-mariadb" deleted
❯ kubectl delete repository -n demo-2 --all
repository.stash.appscode.com "app-sample-mariadb-2" deleted
❯ kubectl delete repository -n demo-3 --all
repository.stash.appscode.com "app-sample-mariadb-3" deleted
```
