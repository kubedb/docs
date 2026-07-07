---
title: Qdrant StorageClass Migration Guide
menu:
  docs_{{ .version }}:
    identifier: qdrant-migration-storageClass
    name: StorageClass Migration
    parent: qdrant-migration
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---


> New to KubeDB? Please start [here](/docs/README.md).

# Qdrant StorageClass Migration

This guide will show you how to use `KubeDB` Ops Manager to migrate `StorageClass` of Qdrant database.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- You must have at least two `StorageClass` resources in order to perform a migration.

- Install `KubeDB` operator in your cluster following the steps [here](/docs/setup/README.md).

To keep everything isolated, we are going to use a separate namespace called `demo` throughout this tutorial.

```bash
kubectl create ns demo
```
namespace/demo created

## Prepare Qdrant Database

At first verify that your cluster has at least two `StorageClass`. Let's check,

```bash
➤ kubectl get storageclass
NAME                   PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
local-path (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  12d
longhorn               driver.longhorn.io      Delete          Immediate              true                   12d
longhorn-custom        driver.longhorn.io      Delete          WaitForFirstConsumer   true                   2d20h
longhorn-static        driver.longhorn.io      Delete          Immediate              true                   12d
```
From the above output we can see that we have more than two `StorageClass` resources. We will now deploy a `Qdrant` database using `standard` StorageClass and insert some data into it.
After that, we will apply `QdrantOpsRequest` to migrate StorageClass from `standard` to `longhorn-custom`.

> Both the old and new PVCs should be on the same node. Therefore, the new StorageClass `VOLUMEBINDINGMODE` should be `WaitForFirstConsumer` if the old one uses `WaitForFirstConsumer`. If the old one uses `Immediate` any mode is allowed.

KubeDB implements a `Qdrant` CRD to define the specification of a Qdrant database. Below is the `Qdrant` object created in this tutorial.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Qdrant
metadata:
  name: sample-qdrant
  namespace: demo
spec:
  version: "1.17.0"
  replicas: 3
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 2Gi
  deletionPolicy: WipeOut
```

```bash
kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/migration/sample-qdrant.yaml
```
qdrant.kubedb.com/sample-qdrant created
Now, wait until sample-qdrant has status `Ready` and check the `StorageClass`,

```bash
kubectl get qdrant,pvc -n demo
```
NAME                    VERSION   STATUS   AGE
sample-qdrant   1.17.0     Ready    101s

NAME                                             STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   VOLUMEATTRIBUTESCLASS   AGE
persistentvolumeclaim/data-sample-qdrant-0   Bound    pvc-64cca3c6-85aa-426f-abc3-b300ecfe365a   2Gi        RWO            standard       <unset>                 96s
persistentvolumeclaim/data-sample-qdrant-1   Bound    pvc-1de36b06-8e32-4e9a-a01b-3b6d7c618688   2Gi        RWO            standard       <unset>                 90s
persistentvolumeclaim/data-sample-qdrant-2   Bound    pvc-a75bd538-8a71-4f62-8d38-3f4e42ffb225   2Gi        RWO            standard       <unset>                 85s

The database is `Ready` and all the `PersistentVolumeClaim` uses `standard` StorageClass. Let's create a collection and insert some data.

# get the API key from the auth secret
```bash
export API_KEY=$(kubectl get secret -n demo sample-qdrant-auth -o jsonpath='{.data.api-key}' | base64 -d)
```

# port-forward the Qdrant service
```bash
kubectl port-forward -n demo svc/sample-qdrant 6333:6333 &
```
Forwarding from 127.0.0.1:6333 -> 6333

# create a collection
```bash
curl -X PUT 'http://localhost:6333/collections/demo_vectors' \
  -H "api-key: $API_KEY" \
  -H 'Content-Type: application/json' \
  -d '{
```
    "vectors": { "size": 4, "distance": "Cosine" }
  }'
{"result":true,"status":"ok","time":0.123}

# insert points
```bash
curl -X PUT 'http://localhost:6333/collections/demo_vectors/points' \
  -H "api-key: $API_KEY" \
  -H 'Content-Type: application/json' \
  -d '{
```
    "points": [
      { "id": 1, "vector": [0.1, 0.2, 0.3, 0.4], "payload": { "label": "a" } },
      { "id": 2, "vector": [0.2, 0.3, 0.4, 0.5], "payload": { "label": "b" } }
    ]
  }'
{"result":null,"status":"ok","time":0.045}

# verify points count
```bash
curl 'http://localhost:6333/collections/demo_vectors' \
  -H "api-key: $API_KEY"
```
{"result":{"status":"green","vectors_count":2,"segments_count":4,...},"status":"ok","time":0.001}

## Apply StorageMigration Ops-Request

To migrate `StorageClass` we have to create a `QdrantOpsRequest` CR with our desired `StorageClass`. Below is the YAML of the `QdrantOpsRequest` CR that we are going to create,

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: QdrantOpsRequest
metadata:
  name: storage-migration
  namespace: demo
spec:
  type: StorageMigration
  databaseRef:
    name: sample-qdrant
  migration:
    storageClassName: longhorn-custom
    oldPVReclaimPolicy: Delete
```

Here,

- `spec.type` specifies that we are performing `StorageMigration` operation.
- `spec.databaseRef.name` specifies that we are performing StorageMigration operation on `sample-qdrant` database.
- `spec.migration.storageClassName` specifies our desired StorageClass
- `spec.migration.oldPVReclaimPolicy` specifies the reclaim policy of previous persistent volume.

> Note: To retain the old PersistentVolume, set `spec.migration.oldPVReclaimPolicy` to `Retain`.

Let's create the `QdrantOpsRequest` CR we have shown above,

```bash
kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/migration/storage-migration.yaml
```
qdrantopsrequest.ops.kubedb.com/storage-migration created
## Verify the StorageClass Migrated Successfully

If everything goes well, `KubeDB` operator will migrate the `StorageClass` along with the data.

Let's wait for `QdrantOpsRequest` to be `Successful`. Run the following command to watch QdrantOpsRequest CR,

```bash
watch kubectl get qdrantopsrequest -n demo
```
Every 2.0s: kubectl get qdrantopsrequest -n demo  

NAME                TYPE               STATUS       AGE
storage-migration   StorageMigration   Successful   13m

We can see from the above output that the `QdrantOpsRequest` has succeeded. Let's verify the StorageClass.

```bash
kubectl get pvc -n demo
```
NAME                       STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS        VOLUMEATTRIBUTESCLASS   AGE
data-sample-qdrant-0   Bound    pvc-64cca3c6-85aa-426f-abc3-b300ecfe365a   2Gi        RWO            longhorn-custom     <unset>                 21m
data-sample-qdrant-1   Bound    pvc-1de36b06-8e32-4e9a-a01b-3b6d7c618688   2Gi        RWO            longhorn-custom     <unset>                 21m
data-sample-qdrant-2   Bound    pvc-a75bd538-8a71-4f62-8d38-3f4e42ffb225   2Gi        RWO            longhorn-custom     <unset>                 21m

The `PersistentVolumeClaim` StorageClass has changed to `longhorn-custom`. Now, we will verify that the data remains intact after the `StorageMigration` operation.

# get the API key from the auth secret
```bash
export API_KEY=$(kubectl get secret -n demo sample-qdrant-auth -o jsonpath='{.data.api-key}' | base64 -d)
```

# port-forward the Qdrant service
```bash
kubectl port-forward -n demo svc/sample-qdrant 6333:6333 &
```
Forwarding from 127.0.0.1:6333 -> 6333

# check the collection exists and data is intact
```bash
curl 'http://localhost:6333/collections/demo_vectors' \
  -H "api-key: $API_KEY"
```
{"result":{"status":"green","vectors_count":2,"segments_count":4,...},"status":"ok","time":0.001}

From the above output we can verify that data remains intact after the `StorageMigration` operation.

## CleanUp

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete qdrantopsrequest -n demo storage-migration
```

```bash
kubectl delete qdrant -n demo sample-qdrant
```

```bash
kubectl delete ns demo
```
