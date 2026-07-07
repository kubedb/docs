---
title: Storage Migration HanaDB
menu:
  docs_{{ .version }}:
    identifier: guides-hanadb-storage-migration-storage-migration
    name: Storage Migration
    parent: guides-hanadb-storage-migration
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Storage Migration of HanaDB

This guide shows how to move a HanaDB's data volumes to a different `StorageClass` using a
`HanaDBOpsRequest` of type `StorageMigration`. KubeDB provisions new PVCs on the target StorageClass,
copies the data with a migrator Job, swaps the volumes in, and recreates each pod (primary last for a
cluster).

> Note: The YAML files used in this tutorial are stored in [docs/examples/hanadb/storage-migration](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/hanadb/storage-migration) folder in the GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Before You Begin

- Install the KubeDB Provisioner and Ops-manager operators following the steps [here](/docs/setup/README.md).
- You need a **source** and a **target** `StorageClass`. This guide migrates between two
  [Longhorn](https://longhorn.io/) StorageClasses, `longhorn-single` and `longhorn-single-migrated`:

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: longhorn-single-migrated
provisioner: driver.longhorn.io
allowVolumeExpansion: true
reclaimPolicy: Delete
volumeBindingMode: Immediate
parameters:
  numberOfReplicas: "1"
  staleReplicaTimeout: "30"
  fsType: ext4
  dataLocality: disabled
  dataEngine: v1
```

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/storage-migration/longhorn-single.yaml
```

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/storage-migration/longhorn-single-migrated.yaml
```

## Deploy a HanaDB on the Source StorageClass

The base manifest places the data on `longhorn-single` and adds an init container that fixes the volume
permissions (HANA runs as `12000:79`):

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/storage-migration/storage-migration-base.yaml
```
hanadb.kubedb.com/hanadb-cluster created

Wait until `hanadb-cluster` is `Ready`, then note the source StorageClass of the PVCs:

```bash
kubectl get pvc -n demo -l app.kubernetes.io/instance=hanadb-cluster \
  -o custom-columns=NAME:.metadata.name,SC:.spec.storageClassName,SIZE:.status.capacity.storage
```
NAME                            SC                SIZE
data-hanadb-cluster-0           longhorn-single   64Gi
data-hanadb-cluster-1           longhorn-single   64Gi
data-hanadb-cluster-arbiter-0   longhorn-single   2Gi

## Create a StorageMigration HanaDBOpsRequest

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: HanaDBOpsRequest
metadata:
  name: hdbops-storage-migration
  namespace: demo
spec:
  type: StorageMigration
  databaseRef:
    name: hanadb-cluster
  migration:
    storageClassName: longhorn-single-migrated
  timeout: 30m
  apply: IfReady
```

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/hanadb/storage-migration/storage-migration.yaml
```
hanadbopsrequest.ops.kubedb.com/hdbops-storage-migration created

Here,

- `spec.migration.storageClassName` is the **target** StorageClass.
- `spec.timeout` is **required** for `StorageMigration`; migrations copy all data and can take a while.

## Verify the Migration

```bash
kubectl get hdbops -n demo hdbops-storage-migration
```
NAME                       TYPE               STATUS       AGE
hdbops-storage-migration   StorageMigration   Successful   14m

```bash
kubectl describe hdbops -n demo hdbops-storage-migration
```
...
Status:
  Conditions:
    Message:  Successfully migrated StorageClass for HanaDB
    Reason:   StorageMigration
    Status:   True
    Type:     StorageMigration
    Message:  Successfully Migrated HanaDB StorageClass
    Reason:   Successful
    Status:   True
    Type:     Successful
  Phase:      Successful

KubeDB migrates the data PVCs one node at a time (the primary last); the small arbiter volume is left on
its original StorageClass. Confirm the data PVCs are now bound to the target StorageClass and the database
is `Ready`:

```bash
kubectl get pvc -n demo -l app.kubernetes.io/instance=hanadb-cluster \
  -o custom-columns=NAME:.metadata.name,SC:.spec.storageClassName,SIZE:.status.capacity.storage
```
NAME                            SC                         SIZE
data-hanadb-cluster-0           longhorn-single-migrated   64Gi
data-hanadb-cluster-1           longhorn-single-migrated   64Gi
data-hanadb-cluster-arbiter-0   longhorn-single            2Gi

```bash
kubectl get hanadb.kubedb.com -n demo hanadb-cluster
```
NAME             VERSION   STATUS   AGE
hanadb-cluster   2.0.82    Ready    26m

## Cleaning Up

```bash
kubectl delete hdbops -n demo hdbops-storage-migration
```

```bash
kubectl delete hanadb.kubedb.com -n demo hanadb-cluster
```

```bash
kubectl delete ns demo
```

## Next Steps

- [Expand the volume](/docs/guides/hanadb/volume-expansion/volume-expansion.md) of a HanaDB.
- Review the [HanaDBOpsRequest CRD](/docs/guides/hanadb/concepts/opsrequest.md).
