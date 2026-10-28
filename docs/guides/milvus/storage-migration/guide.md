---
title: Milvus Storage Migration
menu:
  docs_{{ .version }}:
    identifier: milvus-storage-migration-guide
    name: Guide
    parent: milvus-storage-migration
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Milvus Storage Migration

This guide will show you how to use the `KubeDB` Ops-manager operator to migrate the persistent volumes of a Milvus database from one `StorageClass` to another.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Milvus](/docs/guides/milvus/concepts/milvus.md)
  - [MilvusOpsRequest](/docs/guides/milvus/concepts/milvusopsrequest.md)

- Complete the dependency setup from [Prepare Dependencies](/docs/guides/milvus/quickstart/prerequisites.md). It installs MinIO and creates the object-storage secrets (`milvus-storage-config` for `milvus-standalone`, `milvus-cluster-storage-config` for `milvus-cluster`).

- You need **at least two** `StorageClass`es — the current one and the target one. This guide migrates from `local-path` to `longhorn-custom`:

  ```bash
  $ kubectl get sc
  NAME                   PROVISIONER             ALLOWVOLUMEEXPANSION   AGE
  local-path (default)   rancher.io/local-path   false                  11h
  longhorn-custom        driver.longhorn.io      true                   11h
  ```

> Note: The yaml files used in this tutorial are stored in [docs/guides/milvus/storage-migration/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/milvus/storage-migration/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Storage Migration — Standalone Milvus

Deploy the [standalone quickstart](/docs/guides/milvus/quickstart/standalone.md) manifest, which uses `local-path`, and wait until it is `Ready`:

> If you deleted a Milvus of the same name in an earlier guide, delete its object-storage prefix first, as shown in [Prepare Dependencies](/docs/guides/milvus/quickstart/prerequisites.md#delete-a-milvus-instances-object-storage-data). Otherwise the new database can hang in `Provisioning`.

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/quickstart/yamls/standalone.yaml
milvus.kubedb.com/milvus-standalone created

$ kubectl get milvuses.kubedb.com -n demo milvus-standalone -w

$ kubectl get pvc -n demo -l app.kubernetes.io/instance=milvus-standalone
NAME                       STATUS   VOLUME      CAPACITY   ACCESS MODES   STORAGECLASS   AGE
data-milvus-standalone-0   Bound    pvc-...     1Gi        RWO            local-path     14m
```

### Apply the StorageMigration OpsRequest

`storage-migration-standalone.yaml`

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MilvusOpsRequest
metadata:
  name: storage-migration
  namespace: demo
spec:
  type: StorageMigration
  databaseRef:
    name: milvus-standalone
  migration:
    storageClassName: longhorn-custom
    oldPVReclaimPolicy: Delete
  timeout: 10m
```

Here,

- `spec.migration.storageClassName` is the target `StorageClass`.
- `spec.migration.oldPVReclaimPolicy` controls what happens to the old PersistentVolume (`Delete` or `Retain`).

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/storage-migration/yamls/storage-migration-standalone.yaml
milvusopsrequest.ops.kubedb.com/storage-migration created
```

### Watch Progress

```bash
$ kubectl get milvusopsrequest storage-migration -n demo
NAME                TYPE               STATUS       AGE
storage-migration   StorageMigration   Successful   87s
```

```bash
$ kubectl describe milvusopsrequest storage-migration -n demo
...
Status:
  Conditions:
    Message:  StorageClass migration is in progress
    Reason:   Running
    Type:     Running
    Message:  pet set deleted; ConditionStatus:True; PodName:milvus-standalone
    Type:     PetSetDeleted--milvus-standalone
    Message:  get storage class; ConditionStatus:True
    Type:     GetStorageClass
    ...
  Phase:      Successful
```

During migration, the operator runs a migrator job to copy the data, recreates the PVC on the new `StorageClass`, and recreates the pod. A standalone Milvus has a single pod, so the database is unavailable while it is migrated (about a minute and a half in our test).

### Verify the Migration

The PVC is now backed by `longhorn-custom`, and the database spec reflects the new class:

```bash
$ kubectl get pvc -n demo -l app.kubernetes.io/instance=milvus-standalone
NAME                       STATUS   VOLUME      CAPACITY   ACCESS MODES   STORAGECLASS      AGE
data-milvus-standalone-0   Bound    pvc-...     1Gi        RWO            longhorn-custom   33s

$ kubectl get milvuses.kubedb.com milvus-standalone -n demo -o jsonpath='{.spec.storage.storageClassName}'
longhorn-custom
```

## Storage Migration — Distributed Milvus

For a distributed Milvus, migration targets the workloads that carry persistent storage — i.e. `streamingnode`. Point `spec.databaseRef.name` at `milvus-cluster`.

This example uses two `streamingnode` replicas on `local-path`. `storage-migration/yamls/milvus-distributed.yaml` is the [distributed quickstart](/docs/guides/milvus/quickstart/distributed.md) manifest with `streamingnode.replicas: 2`. Deploy it and wait until it is `Ready`:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/storage-migration/yamls/milvus-distributed.yaml
milvus.kubedb.com/milvus-cluster created

$ kubectl get milvuses.kubedb.com -n demo milvus-cluster -w
```

The distributed OpsRequest reuses the name of the standalone one, so delete the standalone OpsRequest first:

```bash
$ kubectl delete milvusopsrequest -n demo storage-migration
```

`storage-migration-distributed.yaml`

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: MilvusOpsRequest
metadata:
  name: storage-migration
  namespace: demo
spec:
  type: StorageMigration
  databaseRef:
    name: milvus-cluster
  migration:
    storageClassName: longhorn-custom
    oldPVReclaimPolicy: Delete
  timeout: 10m
```

The operator migrates the PVC of every `streamingnode` replica. Starting from `local-path`:

```bash
$ kubectl get pvc -n demo -l app.kubernetes.io/instance=milvus-cluster -o custom-columns=NAME:.metadata.name,SIZE:.status.capacity.storage,SC:.spec.storageClassName
NAME                                  SIZE   SC
data-milvus-cluster-streamingnode-0   1Gi    local-path
data-milvus-cluster-streamingnode-1   1Gi    local-path
```

Apply the OpsRequest. The `streamingnode` replicas are migrated one at a time, so one of them keeps running throughout; the other roles are not restarted:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/storage-migration/yamls/storage-migration-distributed.yaml
milvusopsrequest.ops.kubedb.com/storage-migration created
```

After the migration completes:

```bash
$ kubectl get milvusopsrequest storage-migration -n demo
NAME                TYPE               STATUS       AGE
storage-migration   StorageMigration   Successful   3m40s

$ kubectl get pvc -n demo -l app.kubernetes.io/instance=milvus-cluster -o custom-columns=NAME:.metadata.name,SIZE:.status.capacity.storage,SC:.spec.storageClassName
NAME                                  SIZE   SC
data-milvus-cluster-streamingnode-0   1Gi    longhorn-custom
data-milvus-cluster-streamingnode-1   1Gi    longhorn-custom

$ kubectl get milvuses.kubedb.com milvus-cluster -n demo -o jsonpath='{.spec.topology.distributed.streamingnode.storage.storageClassName}'
longhorn-custom
```

Both `streamingnode` PVCs are migrated.

## Cleaning up

```bash
$ kubectl delete milvusopsrequest -n demo storage-migration
$ kubectl patch -n demo milvus.kubedb.com milvus-standalone -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
$ kubectl patch -n demo milvus.kubedb.com milvus-cluster -p '{"spec":{"deletionPolicy":"WipeOut"}}' --type="merge"
$ kubectl delete milvus.kubedb.com -n demo milvus-standalone milvus-cluster
$ kubectl delete ns demo
```

## Next Steps

- Learn about [volume expansion](/docs/guides/milvus/volume-expansion/guide.md) of a Milvus database.
- Detail concepts of [Milvus object](/docs/guides/milvus/concepts/milvus.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
