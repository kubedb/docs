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
- To install `External-snapshotter`  in your cluster following the steps [here](https://github.com/kubernetes-csi/external-snapshotter/tree/release-5.0).

To keep things isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
$ kubectl create ns demo
namespace/demo created
```

> **Note:** YAML files used in this tutorial are stored in [docs/examples/qdrant/backup/volume-snapshot](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/qdrant/backup/volume-snapshot) directory of [kubedb/docs](https://github.com/kubedb/docs) repository.

### Ensure VolumeSnapshotClass

```bash
$ kubectl get volumesnapshotclasses
NAME                    DRIVER               DELETIONPOLICY   AGE
longhorn-snapshot-vsc   driver.longhorn.io   Delete           7d22h
```

If not any, create a `VolumeSnapshotClass` using the following YAML,

```yaml
kind: VolumeSnapshotClass
apiVersion: snapshot.storage.k8s.io/v1
metadata:
  name: longhorn-snapshot-vsc
driver: driver.longhorn.io
deletionPolicy: Delete
parameters:
  type: snap
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/backup/volume-snapshot/volume-snapshot-class.yaml
volumesnapshotclass.snapshot.storage.k8s.io/longhorn-snapshot-vsc created
```

> **Note:** Ensure that the `VolumeSnapshotClass` is provisioned with the same storage class driver used for provisioning your Qdrant database. In our case, we are using the `longhorn` storageclass as our database provisioner, with the driver set to `driver.longhorn.io`.

### Prepare Backend

We are going to store our backed up data into a MinIO bucket. We have to create a Secret with necessary credentials and a `BackupStorage` CR to use this backend. If you want to use a different backend, please read the respective backend configuration doc from [here](https://kubestash.com/docs/latest/guides/backends/overview/).

**Create Secret:**

Let's create a secret called `storage-secret` with access credentials to our desired MinIO backend,

```bash
$ kubectl create secret generic -n demo storage-secret \
    --from-literal=AWS_ACCESS_KEY_ID=minioadmin \
    --from-literal=AWS_SECRET_ACCESS_KEY=minioadmin
secret/storage-secret created
```

**Create BackupStorage:**

Now, create a `BackupStorage` using this secret. Below is the YAML of `BackupStorage` CR we are going to create,

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
      secretName: storage-secret
  usagePolicy:
    allowedNamespaces:
      from: All
  default: true
  deletionPolicy: Delete
```

Let's create the BackupStorage we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/backup/volume-snapshot/backup-storage.yaml
backupstorage.storage.kubestash.com/minio-storage created
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
  successfulSnapshots:
    last: 5
  usagePolicy:
    allowedNamespaces:
      from: All
```

Let's create the above `RetentionPolicy`,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/backup/volume-snapshot/retention-policy.yaml
retentionpolicy.storage.kubestash.com/demo-retention created
```

## Deploy Sample Qdrant Database

Let's deploy a sample `Qdrant` database and insert some data into it.

**Create Qdrant CR:**

Below is the YAML of a sample `Qdrant` CRD that we are going to create for this tutorial:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Qdrant
metadata:
  name: qdrant-sample
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
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/backup/volume-snapshot/qdrant.yaml
qdrant.kubedb.com/qdrant-sample created
```

KubeDB will deploy a Qdrant database according to the above specification. It will also create the necessary `Secrets` and `Services` to access the database.

Let's check if the database is ready to use,

```bash
$ kubectl get qdrant -n demo
NAME            VERSION   STATUS   AGE
qdrant-sample   1.17.0    Ready    65s
```

The database is `Ready`. Verify that KubeDB has created a `Secret` and a `Service` for this database using the following commands,

```bash
$ kubectl get secret -n demo -l=app.kubernetes.io/instance=qdrant-sample
NAME                   TYPE     DATA   AGE
qdrant-sample-auth     Opaque   2      65s

$ kubectl get service -n demo -l=app.kubernetes.io/instance=qdrant-sample
NAME                 TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)             AGE
qdrant-sample        ClusterIP   10.43.69.124   <none>        6333/TCP,6334/TCP   65s
qdrant-sample-pods   ClusterIP   None           <none>        6335/TCP            65s
```

KubeDB creates an [AppBinding](/docs/guides/qdrant/concepts/appbinding/index.md) CR that holds the necessary information to connect with the database.

**Verify AppBinding:**

Verify that the `AppBinding` has been created successfully using the following command,

```bash
$ kubectl get appbindings -n demo
NAME            TYPE                VERSION   AGE
qdrant-sample   kubedb.com/qdrant   1.17.0    64s
```

**Insert Sample Data:**

Now, let's get the API key and port-forward to create a collection with sample data:

```bash
# Get the API key from the auth secret
$ export API_KEY=$(kubectl get secret -n demo qdrant-sample-auth -o jsonpath='{.data.api-key}' | base64 -d)

# Port-forward to the Qdrant service
$ kubectl port-forward -n demo svc/qdrant-sample 6333:6333 &
# Create a collection
$ curl -X PUT 'http://localhost:6333/collections/demo_collection' \
  -H "api-key: $API_KEY" \
  -H 'Content-Type: application/json' \
  -d '{"vectors": {"size": 4, "distance": "Cosine"}}'
# Insert points
$ curl -X PUT 'http://localhost:6333/collections/demo_collection/points' \
  -H "api-key: $API_KEY" \
  -H 'Content-Type: application/json' \
  -d '{
    "points": [
      { "id": 1, "vector": [0.1, 0.2, 0.3, 0.4], "payload": {"label": "a"} },
      { "id": 2, "vector": [0.5, 0.6, 0.7, 0.8], "payload": {"label": "b"} }
    ]
  }'
```

Now, we are ready to backup the database.

## Backup

We have to create a `BackupConfiguration` targeting respective `qdrant-sample` Qdrant database. Then, KubeStash will create a `CronJob` for each session to take periodic backup of that database using volume snapshots.

At first, we need to create a secret with a Restic password for backup data encryption.

**Create Secret:**

Let's create a secret called `encrypt-secret` with the Restic password,

```bash
$ echo -n 'changeit' > RESTIC_PASSWORD
$ kubectl create secret generic -n demo encrypt-secret \
    --from-file=./RESTIC_PASSWORD
secret "encrypt-secret" created
```

**Create BackupConfiguration:**

Below is the YAML for `BackupConfiguration` CR to backup the `qdrant-sample` database that we have deployed earlier,

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: qdrant-sample-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Qdrant
    namespace: demo
    name: qdrant-sample
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

Here,
- `.spec.sessions[*].schedule` specifies that we want to backup the database at `5 minutes` interval.
- `.spec.target` refers to the targeted `qdrant-sample` Qdrant database that we created earlier.
- `.spec.sessions[*].addon.tasks[*].params[*].volumeSnapshotClassName` specifies the `VolumeSnapshotClass` to use for creating volume snapshots.

Let's create the `BackupConfiguration` CR that we have shown above,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/backup/volume-snapshot/backup-configuration.yaml
backupconfiguration.core.kubestash.com/qdrant-sample-backup created
```

**Verify Backup Setup Successful:**

If everything goes well, the phase of the `BackupConfiguration` should be `Ready`. The `Ready` phase indicates that the backup setup is successful. Let's verify the `Phase` of the BackupConfiguration,

```bash
$ kubectl get backupconfiguration -n demo
NAME                    PHASE   PAUSED   AGE
qdrant-sample-backup    Ready            36s
```

Additionally, we can verify that the `Repository` specified in the `BackupConfiguration` has been created using the following command,

```bash
$ kubectl get repo -n demo
NAME                INTEGRITY   SNAPSHOT-COUNT   SIZE         PHASE   LAST-SUCCESSFUL-BACKUP   AGE
minio-qdrant-repo   true        5                19.914 KiB   Ready   91s                      101s
```

**Verify VolumeSnapshot:**

It will create a `VolumeSnapshot` for each PVC of the Qdrant database.

Verify that the `VolumeSnapshot` has been created using the following command,

```bash
$ kubectl get volumesnapshot -n demo
NAME                         READYTOUSE   SOURCEPVC              RESTORESIZE   SNAPSHOTCLASS           CREATIONTIME   AGE
qdrant-sample-0-1779334719   true         data-qdrant-sample-0   200Mi         longhorn-snapshot-vsc   2m38s          2m38s
qdrant-sample-1-1779334729   true         data-qdrant-sample-1   200Mi         longhorn-snapshot-vsc   2m28s          2m28s
qdrant-sample-2-1779334744   true         data-qdrant-sample-2   200Mi         longhorn-snapshot-vsc   2m13s          2m13s
```

**Verify BackupSession:**

KubeStash triggers an instant backup as soon as the `BackupConfiguration` is ready. After that, backups are scheduled according to the specified schedule.

```bash
$ kubectl get backupsession -n demo -w

NAME                                              INVOKER-TYPE          INVOKER-NAME           PHASE       DURATION   AGE
qdrant-sample-backup-frequent-backup-1779334706   BackupConfiguration   qdrant-sample-backup   Succeeded   41s        91s
```

We can see from the above output that the backup session has succeeded.

## Restore

In this section, we are going to restore the database from the volume snapshot backup we have taken in the previous section. First, delete the original database, then deploy a new one initialized from the backup using the `init.archiver` field.

```bash
$ kubectl delete qdrant -n demo qdrant-sample
qdrant.kubedb.com "qdrant-sample" deleted
```

#### Deploy Restored Database:

Below is the YAML for `Qdrant` CRD we are going deploy to initialize from backup,

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Qdrant
metadata:
  name: qdrant-sample
  namespace: demo
spec:
  version: "1.17.0"
  mode: Distributed
  replicas: 3
  init:
    archiver:
      recoveryTimestamp: "2026-12-12T00:00:00Z"
      fullDBRepository:
        name: minio-qdrant-repo
        namespace: demo
  storage:
    storageClassName: longhorn
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 200Mi
  deletionPolicy: WipeOut
```

Here,
- `.spec.init.archiver.recoveryTimestamp` specifies the timestamp to recover to. KubeDB will restore the database to the state at this timestamp using the volume snapshots.
- `.spec.init.archiver.fullDBRepository` specifies the Repository object that holds the backed up data.

Let's create the above database,

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/backup/volume-snapshot/qdrant-restore.yaml
qdrant.kubedb.com/qdrant-sample created
```

KubeDB will automatically restore the database from the volume snapshot backup. Wait for the database to become `Ready`,

```bash
$ kubectl get qdrant -n demo qdrant-sample
NAME            VERSION   STATUS   AGE
qdrant-sample   1.17.0    Ready    2m1s
```

#### Verify Restored Data:

In this section, we are going to verify whether the desired data has been restored successfully.

Now, let's get the API key and port-forward to verify the restored data,

```bash
# Get the API key from the auth secret
$ export API_KEY=$(kubectl get secret -n demo qdrant-sample-auth -o jsonpath='{.data.api-key}' | base64 -d)

$ kubectl port-forward -n demo svc/qdrant-sample 6333:6333 &
# Scroll points to verify the restored data
$ curl -X POST 'http://localhost:6333/collections/demo_collection/points/scroll' \
  -H "api-key: $API_KEY" \
  -H 'Content-Type: application/json' \
  -d '{"limit": 10, "with_payload": true, "with_vector": true}'
```

```json
{
  "result": {
    "points": [
      {
        "id": 1,
        "payload": {"label": "a"},
        "vector": [0.18257418, 0.36514837, 0.5477226, 0.73029673]
      },
      {
        "id": 2,
        "payload": {"label": "b"},
        "vector": [0.37904903, 0.45485884, 0.5306686, 0.60647845]
      }
    ],
    "next_page_offset": null
  },
  "status": "ok",
  "time": 0.006733993
}
```

So, from the above output, we can see that the `demo_collection` we created earlier in the original database is now restored successfully.

## Cleanup

To cleanup the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete backupconfiguration.core.kubestash.com -n demo qdrant-sample-backup
kubectl delete retentionpolicy.storage.kubestash.com -n demo demo-retention
kubectl delete backupstorage -n demo minio-storage
kubectl delete secret -n demo storage-secret
kubectl delete secret -n demo encrypt-secret
kubectl delete qdrant -n demo qdrant-sample
```
