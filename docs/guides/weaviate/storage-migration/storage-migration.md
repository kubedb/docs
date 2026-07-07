---
title: Weaviate Storage Migration
menu:
  docs_{{ .version }}:
    identifier: weaviate-storage-migration-cluster
    name: Storage Migration
    parent: weaviate-storage-migration
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Weaviate StorageClass Migration

This guide will show you how to use the `KubeDB` Ops Manager to migrate a Weaviate cluster from one `StorageClass` to another.

## Before You Begin

- At first, you need to have a Kubernetes cluster, and the `kubectl` command-line tool must be configured to communicate with your cluster.

- Install `KubeDB` in your cluster following the steps [here](/docs/setup/README.md).

- You need at least two `StorageClass`es in your cluster — the one the database currently runs on, and the one you want to migrate to. Verify with:

  ```bash
  kubectl get storageclass
  ```
  NAME                   PROVISIONER             RECLAIMPOLICY   VOLUMEBINDINGMODE      ALLOWVOLUMEEXPANSION   AGE
  local-path (default)   rancher.io/local-path   Delete          WaitForFirstConsumer   false                  38h
  longhorn               driver.longhorn.io      Delete          Immediate              true                   30m

- You should be familiar with the following `KubeDB` concepts:
  - [Weaviate](/docs/guides/weaviate/concepts/weaviate.md)
  - [Storage Migration Overview](/docs/guides/weaviate/storage-migration/overview.md)

To keep things isolated, this tutorial uses a separate namespace called `demo` throughout this tutorial.

```bash
kubectl create ns demo
```
namespace/demo created

> **Note:** YAML files used in this tutorial are stored in [docs/examples/weaviate/storage-migration](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/weaviate/storage-migration) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Deploy Weaviate

In this section, we are going to deploy a Weaviate cluster on the `longhorn` StorageClass. We will migrate it to `local-path` in the next step.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Weaviate
metadata:
  name: weaviate-sample
  namespace: demo
spec:
  version: 1.33.1
  replicas: 2
  storageType: Durable
  storage:
    storageClassName: longhorn
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 3Gi
  deletionPolicy: WipeOut
```

Let's create the `Weaviate` CR and wait for it to become `Ready`. Then check the current StorageClass of the PVCs:

```bash
kubectl get pvc -n demo -o custom-columns=NAME:.metadata.name,SC:.spec.storageClassName,SIZE:.status.capacity.storage
```
NAME                     SC         SIZE
data-weaviate-sample-0   longhorn   3Gi
data-weaviate-sample-1   longhorn   3Gi

The cluster is currently running on the `longhorn` StorageClass.

## Apply StorageMigration OpsRequest

Now, we are going to migrate the cluster from `longhorn` to `local-path`.

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: WeaviateOpsRequest
metadata:
  name: storage-migration
  namespace: demo
spec:
  type: StorageMigration
  databaseRef:
    name: weaviate-sample
  timeout: 10m
  migration:
    storageClassName: local-path
    oldPVReclaimPolicy: Delete
```

- `spec.type` specifies that this is a `StorageMigration` operation.
- `spec.migration.storageClassName` specifies the target `StorageClass`. It must be different from the current one.
- `spec.migration.oldPVReclaimPolicy` specifies the reclaim policy applied to the old PersistentVolumes after migration (here, `Delete`).

Let's create the `WeaviateOpsRequest` CR:

```bash
kubectl create -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/weaviate/storage-migration/ops-request.yaml
```
weaviateopsrequest.ops.kubedb.com/storage-migration created

For each node, the Ops Manager provisions a new PVC on the target StorageClass, runs a migrator job to copy the data, deletes the old PVC, and re-points the node at the new volume.

```bash
kubectl get weaviateopsrequest -n demo storage-migration
```
NAME                TYPE               STATUS       AGE
storage-migration   StorageMigration   Successful   6m

Let's look at the (abbreviated) `status.conditions` of the `WeaviateOpsRequest`:

```bash
kubectl get weaviateopsrequest -n demo storage-migration -o yaml
```
...
status:
  conditions:
  - message: StorageClass migration is in progress
    reason: Running
    status: "True"
    type: Running
  - message: get storage class; ConditionStatus:True
    status: "True"
    type: GetStorageClass
  - message: 'p v c created; ConditionStatus:True; PodName:data-migrate-weaviate-sample-0'
    status: "True"
    type: PVCCreated--data-migrate-weaviate-sample-0
  - message: 'job created; ConditionStatus:True; PodName:migrator-weaviate-sample-0'
    status: "True"
    type: JobCreated--migrator-weaviate-sample-0
  - message: PVC Migration Completed for weaviate-sample-0
    reason: PodMigrationCompleted-weaviate-sample-0
    status: "True"
    type: PodMigrationCompleted-weaviate-sample-0
  - message: PVC Migration Completed for weaviate-sample-1
    reason: PodMigrationCompleted-weaviate-sample-1
    status: "True"
    type: PodMigrationCompleted-weaviate-sample-1
  - message: Successfully migrated StorageClass for Weaviate
    reason: StorageMigration
    status: "True"
    type: StorageMigration
  - message: Successfully Migrated Weaviate StorageClass for demo/weaviate-sample
    reason: Successful
    status: "True"
    type: Successful
  observedGeneration: 1
  phase: Successful

## Verify the StorageClass Migrated Successfully

Verify that the PVCs are now on the `local-path` StorageClass and the database is back to `Ready`:

```bash
kubectl get pvc -n demo -o custom-columns=NAME:.metadata.name,SC:.spec.storageClassName,SIZE:.status.capacity.storage
```
NAME                     SC           SIZE
data-weaviate-sample-0   local-path   3Gi
data-weaviate-sample-1   local-path   3Gi

```bash
kubectl get weaviate -n demo weaviate-sample -o jsonpath='{.spec.storage.storageClassName}{"  "}{.status.phase}'
```
local-path  Ready

The StorageClass has been migrated from `longhorn` to `local-path` successfully.

## Next Steps

- Detail concepts of [Weaviate object](/docs/guides/weaviate/concepts/weaviate.md).
- Expand the volume of a cluster: [Volume Expansion](/docs/guides/weaviate/volume-expansion/volume-expansion.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

## Cleaning Up

To clean up the Kubernetes resources created by this tutorial, run:

```bash
kubectl delete weaviateopsrequest -n demo storage-migration
```

```bash
kubectl delete weaviate -n demo weaviate-sample
```

```bash
kubectl delete ns demo
```
