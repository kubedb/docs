---
title: Volume Snapshot Backup & Restore Qdrant
description: Backup Qdrant database using Volume Snapshot
menu:
  docs_{{ .version }}:
    identifier: guides-qdrant-volume-snapshot
    name: Volume Snapshot
    parent: qdrant-backup
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Volume Snapshot Backup and Restore Qdrant Database

KubeStash allows you to take volume snapshot backups of Qdrant databases. Volume snapshots provide a fast and efficient way to backup and restore the entire storage volume of your Qdrant cluster. This guide will show you how to configure volume snapshot backup and restore for Qdrant databases.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster. If you do not already have a cluster, you can create one by using `Minikube` or `Kind`.
- Install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md).
- Install `KubeStash` in your cluster following the steps [here](https://kubestash.com/docs/latest/setup/install/kubestash).
- Install KubeStash `kubectl` plugin following the steps [here](https://kubestash.com/docs/latest/setup/install/kubectl-plugin/).
- Ensure your storage provider supports VolumeSnapshots (e.g., Longhorn, AWS EBS, GCE PD).
- If you are not familiar with how KubeStash backup and restore Qdrant databases, please check the following guide [here](/docs/guides/qdrant/backup/overview/index.md).

You should be familiar with the following `KubeStash` concepts:

- [BackupStorage](https://kubestash.com/docs/latest/concepts/crds/backupstorage/)
- [BackupConfiguration](https://kubestash.com/docs/latest/concepts/crds/backupconfiguration/)
- [BackupSession](https://kubestash.com/docs/latest/concepts/crds/backupsession/)
- [RestoreSession](https://kubestash.com/docs/latest/concepts/crds/restoresession/)
- [Addon](https://kubestash.com/docs/latest/concepts/crds/addon/)
- [Function](https://kubestash.com/docs/latest/concepts/crds/function/)

To keep things isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/guides/qdrant/backup/volumesnapshot/examples](docs/guides/qdrant/backup/volumesnapshot/examples) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

## Prepare Backup Infrastructure

We are going to store our backed up data using VolumeSnapshots. We have to create a `VolumeSnapshotClass`, `Secret`, `BackupStorage`, and `RetentionPolicy` CR to use this backend. If you want to use a different backend, please read the respective backend configuration doc from [here](https://kubestash.com/docs/latest/guides/backends/overview/).

### Ensure VolumeSnapshotClass

First, ensure that the `VolumeSnapshotClass` for your storage provider is available. For Longhorn:

```bash
$ kubectl get volumesnapshotclasses
NAME                    DRIVER               DELETIONPOLICY   AGE
longhorn-snapshot-vsc   driver.longhorn.io   Delete           7d22h
```

If not available, create one:

```yaml
apiVersion: snapshot.storage.k8s.io/v1
kind: VolumeSnapshotClass
metadata:
  name: longhorn-snapshot-vsc
driver: driver.longhorn.io
deletionPolicy: Delete
parameters:
  type: snap
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/qdrant/backup/volumesnapshot/examples/volumesnapshotclass.yaml
volumesnapshotclass.snapshot.storage.k8s.io/longhorn-snapshot-vsc created
```

Note: Ensure that the VolumeSnapshotClass is provisioned with the same storage class driver used for provisioning your Qdrant database.

### Create BackupStorage

Create a `BackupStorage` CR to configure the backup storage:

```yaml
apiVersion: storage.kubestash.com/v1alpha1
kind: BackupStorage
metadata:
  name: minio-storage
  namespace: demo
spec:
  storage:
    provider: s3
    s3:
      bucket: qdrant-backups
      endpoint: http://minio.demo.svc:9000
      insecureTLS: true
      prefix: backup/demo
      region: us-east-1
      secretName: aws-secret
  usagePolicy:
    allowedNamespaces:
      from: All
  default: true
  deletionPolicy: Delete
```

Apply the BackupStorage:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/qdrant/backup/volumesnapshot/examples/backupstorage.yaml
backupstorage.storage.kubestash.com/minio-storage created
```

### Create Storage Secret

Create a secret with credentials to access the storage:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/qdrant/backup/volumesnapshot/examples/aws-secret.yaml
secret/aws-secret created
```

### Create Encryption Secret

Create a secret for encrypting the backup data:

```bash
$ echo -n 'changeit' > RESTIC_PASSWORD
$ kubectl create secret generic -n demo encrypt-secret \
    --from-file=./RESTIC_PASSWORD
secret "encrypt-secret" created
```

### Create RetentionPolicy

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
  maxNumberOfSnapshots: 5
  usagePolicy:
    allowedNamespaces:
      from: All
```

Let's create the above `RetentionPolicy`,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/qdrant/backup/volumesnapshot/examples/retentionpolicy.yaml
retentionpolicy.storage.kubestash.com/demo-retention created
```

## Deploy Sample Qdrant Database

Let's deploy a sample `Qdrant` database and insert some data into it.

**Create Qdrant CR:**

Below is the YAML of a sample `Qdrant` CRD that we are going to create for this tutorial:

```yaml
apiVersion: kubedb.com/v1
kind: Qdrant
metadata:
  name: sample-qdrant
  namespace: demo
spec:
  version: "1.17.0"
  mode: Distributed
  replicas: 3
  storage:
    storageClassName: longhorn
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 200Mi
  deletionPolicy: WipeOut
```

Create the above `Qdrant` CR,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/qdrant/backup/volumesnapshot/examples/sample-qdrant.yaml
qdrant.kubedb.com/sample-qdrant created
```

KubeDB will deploy a Qdrant database according to the above specification.

Let's check if the database is ready to use,

```bash
$ kubectl get qdrant -n demo
NAME            VERSION   STATUS    AGE
sample-qdrant   1.17.0    Ready     4m22s
```

**Insert Sample Data:**

Now, we are going to exec into the database pod and create some sample data. At first, find out the database `Pod` using the following command,

```bash
$ kubectl get pods -n demo --selector="app.kubernetes.io/instance=sample-qdrant"
NAME                      READY   STATUS    RESTARTS   AGE
sample-qdrant-0          1/1     Running   0          2m41s
sample-qdrant-1          1/1     Running   0          2m35s
sample-qdrant-2          1/1     Running   0          2m29s
```

Now, let's exec into the Pod to insert some sample data into Qdrant:

```bash
$ kubectl exec -it -n demo sample-qdrant-0 -- sh
# Upload some sample points to a collection
$ wget -qO- --header 'Content-Type: application/json' \
  --post-data '{
    "vectors": [
      {"id": 1, "vector": [0.1, 0.2, 0.3, 0.4]},
      {"id": 2, "vector": [0.5, 0.6, 0.7, 0.8]}
    ]
  }' \
  http://localhost:6333/collections/my_collection/points
# Exit the pod
$ exit
```

Now, we are ready to backup the database.

## Backup

We have to create a `BackupConfiguration` targeting respective `sample-qdrant` Qdrant database. Then, KubeStash will create a `CronJob` for each session to take periodic backup of that database using volume snapshots.

**Create BackupConfiguration:**

Below is the YAML for `BackupConfiguration` CR to backup the `sample-qdrant` database that we have deployed earlier,

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: sample-qdrant-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Qdrant
    namespace: demo
    name: sample-qdrant
  backends:
    - name: minio-backend
      storageRef:
        namespace: demo
        name: minio-storage
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
        - name: minio-qdrant-repo
          backend: minio-backend
          directory: /qdrant
          encryptionSecret:
            name: encrypt-secret
            namespace: demo
      addon:
        name: qdrant-addon
        tasks:
          - name: volume-snapshot
            params:
              volumeSnapshotClassName: "longhorn-snapshot-vsc"
```

Let's create the `BackupConfiguration` CR that we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/qdrant/backup/volumesnapshot/examples/backupconfiguration.yaml
backupconfiguration.core.kubestash.com/sample-qdrant-backup created
```

**Verify Backup Setup Successful:**

If everything goes well, the phase of the `BackupConfiguration` should be `Ready`. The `Ready` phase indicates that the backup setup is successful. Let's verify the `Phase` of the BackupConfiguration,

```bash
$ kubectl get backupconfiguration -n demo
NAME                    PHASE   PAUSED   AGE
sample-qdrant-backup    Ready            2m50s
```

Additionally, we can verify that the `Repository` specified in the `BackupConfiguration` has been created using the following command,

```bash
$ kubectl get repo -n demo
NAME                INTEGRITY   SNAPSHOT-COUNT   SIZE     PHASE   LAST-SUCCESSFUL-BACKUP   AGE
minio-qdrant-repo               0                0 B      Ready                            3m
```

**Verify VolumeSnapshot:**

It will create a `VolumeSnapshot` for each PVC of the Qdrant database.

Verify that the `VolumeSnapshot` has been created using the following command,

```bash
$ kubectl get volumesnapshot -n demo
NAME                             READYTOUSE   SOURCEPVC                      SOURCESNAPSHOTCONTENT   RESTORESIZE   SNAPSHOTCLASS              SNAPSHOTCONTENT                                    CREATIONTIME   AGE
minio-qdrant-repo-xyz            true         data-sample-qdrant-0                        1Gi          longhorn-snapshot-vsc      snapcontent-xyz                                     2m            2m
```

**Verify BackupSession:**

KubeStash triggers an instant backup as soon as the `BackupConfiguration` is ready. After that, backups are scheduled according to the specified schedule.

```bash
$ kubectl get backupsession -n demo -w

NAME                                             INVOKER-TYPE          INVOKER-NAME           PHASE       DURATION   AGE
sample-qdrant-backup-frequent-backup-xyz        BackupConfiguration   sample-qdrant-backup    Succeeded              7m22s
```

We can see from the above output that the backup session has succeeded.

## Restore

In this section, we are going to restore the database from the volume snapshot backup we have taken in the previous section. We are going to deploy a new database and initialize it from the backup.

#### Deploy Restored Database:

Now, we have to deploy the restored database similarly as we have deployed the original `sample-qdrant` database. However, this time there will be the following differences:

- We are going to specify `.spec.init.waitForInitialRestore` field that tells KubeDB to wait for first restore to complete before marking this database is ready to use.

Below is the YAML for `Qdrant` CRD we are going deploy to initialize from backup,

```yaml
apiVersion: kubedb.com/v1
kind: Qdrant
metadata:
  name: restored-qdrant
  namespace: demo
spec:
  init:
    waitForInitialRestore: true
  version: "1.17.0"
  mode: Distributed
  replicas: 3
  storage:
    storageClassName: longhorn
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 200Mi
  deletionPolicy: WipeOut
```

Let's create the above database,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/qdrant/backup/volumesnapshot/examples/restored-qdrant.yaml
qdrant.kubedb.com/restored-qdrant created
```

If you check the database status, you will see it is stuck in `Provisioning` state.

```bash
$ kubectl get qdrant -n demo restored-qdrant
NAME               VERSION   STATUS         AGE
restored-qdrant    1.17.0    Provisioning   61s
```

#### Create RestoreSession:

Now, we need to create a RestoreSession CRD pointing to targeted `Qdrant` database.

Below, is the contents of YAML file of the `RestoreSession` object that we are going to create to restore backed up data into the newly created database provisioned by Qdrant object named `restored-qdrant`.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: restore-sample-qdrant
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Qdrant
    namespace: demo
    name: restored-qdrant
  dataSource:
    repository: minio-qdrant-repo
    snapshot: latest
    encryptionSecret:
      name: encrypt-secret
      namespace: demo
  addon:
    name: qdrant-addon
    tasks:
      - name: volume-snapshot-restore
```

Let's create the RestoreSession CRD object we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/qdrant/backup/volumesnapshot/examples/restoresession.yaml
restoresession.core.kubestash.com/restore-sample-qdrant created
```

Once, you have created the `RestoreSession` object, KubeStash will create restore Job. Run the following command to watch the phase of the `RestoreSession` object,

```bash
$ watch kubectl get restoresession -n demo
Every 2.0s: kubectl get restores... AppsCode-PC-03: Wed Aug 21 10:44:05 2024

NAME                    REPOSITORY           FAILURE-POLICY   PHASE       DURATION   AGE
restore-sample-qdrant   minio-qdrant-repo                     Succeeded   3s         53s
```

The `Succeeded` phase means that the restore process has been completed successfully.

#### Verify Restored Data:

In this section, we are going to verify whether the desired data has been restored successfully. We are going to connect to the database server and check whether the collection we created earlier in the original database are restored.

At first, check if the database has gone into `Ready` state by the following command,

```bash
$ kubectl get qdrant -n demo restored-qdrant
NAME               VERSION   STATUS  AGE
restored-qdrant    1.17.0    Ready   34m
```

Now, find out the database `Pod` by the following command,

```bash
$ kubectl get pods -n demo --selector="app.kubernetes.io/instance=restored-qdrant"
NAME                      READY   STATUS    RESTARTS   AGE
restored-qdrant-0         1/1     Running   0          39m
```

Now, let's exec into the Pod to enter into Qdrant and verify restored data,

```bash
$ kubectl exec -it -n demo restored-qdrant-0 -- sh
# Check if the collection exists and has data
$ wget -qO- http://localhost:6333/collections/my_collection
# Exit the pod
$ exit
```

So, from the above output, we can see that the `my_collection` collection we created earlier in the original database and now, it is restored successfully.

## Cleanup

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete backupconfigurations.core.kubestash.com  -n demo sample-qdrant-backup
kubectl delete restoresessions.core.kubestash.com -n demo restore-sample-qdrant
kubectl delete retentionpolicies.storage.kubestash.com -n demo demo-retention
kubectl delete backupstorage -n demo minio-storage
kubectl delete secret -n demo aws-secret
kubectl delete secret -n demo encrypt-secret
kubectl delete qdrant -n demo restored-qdrant
kubectl delete qdrant -n demo sample-qdrant
```
