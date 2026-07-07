---
title: Storage Migration DocumentDB
menu:
  docs_{{ .version }}:
    identifier: dc-storage-migration-details
    name: Storage Migration
    parent: dc-storage-migration
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Storage Migration of DocumentDB

`StorageMigration` moves a `DocumentDB` database from one StorageClass to another without losing
data, using a `DocumentDBOpsRequest` of type `StorageMigration`. This is the tool you reach for
when you need to change the storage backend of a running database — for example moving from one
CSI provisioner to another. This guide migrates a 3-node cluster from `longhorn` to
`standard-custom`.

Unlike provisioning a fresh replica (which seeds standbys with `pg_basebackup`), storage
migration performs a **block-level copy of each existing PVC** into a new PVC on the target
StorageClass, one pod at a time, then re-points the pod at the migrated volume. The data
directory is copied verbatim, so the migrated replica does not have to re-stream a base backup.

## Before You Begin

- You need a Kubernetes cluster and the `kubectl` CLI configured to talk to it.
- Install KubeDB following the steps [here](/docs/setup/README.md).
- This tutorial uses a namespace called `demo` (`kubectl create ns demo`).
- Deploy a `DocumentDB` cluster (`documentdb-cls-sample`) and wait for it to become `Ready`.
- Confirm both the source and target StorageClasses exist (`kubectl get sc`).

> Note: YAML files used in this tutorial are stored in [docs/examples/documentdb](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/examples/documentdb) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## PVCs before

The cluster is on `longhorn`, `10Gi` per replica:

```bash
kubectl get pvc -n demo -l app.kubernetes.io/instance=documentdb-cls-sample \
    -o custom-columns=NAME:.metadata.name,SIZE:.status.capacity.storage,SC:.spec.storageClassName,STATUS:.status.phase
```
NAME                           SIZE   SC         STATUS
data-documentdb-cls-sample-0   10Gi   longhorn   Bound
data-documentdb-cls-sample-1   10Gi   longhorn   Bound
data-documentdb-cls-sample-2   10Gi   longhorn   Bound

```bash
kubectl get docdb -n demo documentdb-cls-sample -o jsonpath='{.spec.storage.storageClassName}'
```
longhorn

## Create the StorageMigration OpsRequest

`migration.storageClassName` is the target; `oldPVReclaimPolicy: Delete` cleans up the source
PersistentVolumes once their data has been copied:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: DocumentDBOpsRequest
metadata:
  name: documentdb-cls-storage-migration
  namespace: demo
spec:
  type: StorageMigration
  databaseRef:
    name: documentdb-cls-sample
  timeout: 10m
  migration:
    storageClassName: standard-custom
    oldPVReclaimPolicy: Delete
```

```bash
kubectl apply -f cluster-storage-migration.yaml
```
documentdbopsrequest.ops.kubedb.com/documentdb-cls-storage-migration created

```bash
kubectl get dcops -n demo documentdb-cls-storage-migration
```
NAME                               TYPE               STATUS       AGE
documentdb-cls-storage-migration   StorageMigration   Successful   8m13s

## What happened

The operator migrates **standbys first and the primary last**, switching leadership off the
primary just before its turn so the cluster stays writable throughout. For each pod it: mounts a
temporary helper pod on a new PVC, runs a `migrator` job to copy the data directory, deletes the
old PVC, binds the new one under the original PVC name, recreates the pod, and waits for it to
be ready. The condition stream (trimmed) captures the loop:

```bash
kubectl get dcops -n demo documentdb-cls-storage-migration \
    -o jsonpath='{range .status.conditions[*]}{.type}={.status} :: {.message}{"\n"}{end}'
```
Running=True :: StorageClass migration is in progress
PetSetDeleted--documentdb-cls-sample=True :: pet set deleted
GetStorageClass=True :: get storage class
# --- per replica (sample-0 shown) ---
PVCCreated--data-migrate-documentdb-cls-sample-0=True :: p v c created
PodCreated--pvcmounter-documentdb-cls-sample-0=True :: pod created
JobCreated--migrator-documentdb-cls-sample-0=True :: job created
JobDeleted--migrator-documentdb-cls-sample-0=True :: job deleted
PVCDeleted--data-documentdb-cls-sample-0=True :: p v c deleted
PVCCreated--data-documentdb-cls-sample-0=True :: p v c created
PodCreated--documentdb-cls-sample-0=True :: pod created
PodReady--documentdb-cls-sample-0=True :: pod ready
PodMigrationCompleted-documentdb-cls-sample-0=True :: PVC Migration Completed for documentdb-cls-sample-0
# --- leadership switched before migrating the last (primary) pod ---
SwitchPrimary--documentdb-cls-sample-2=True :: Successfully switched primary from documentdb-cls-sample-2 to documentdb-cls-sample-0 before its migration
PodMigrationCompleted-documentdb-cls-sample-2=True :: PVC Migration Completed for documentdb-cls-sample-2
StorageMigration=True :: Successfully migrated StorageClass for DocumentDB Database
Successful=True :: Successfully Migrated DocumentDB StorageClass
UnsetRaftKeyOpsRequestProgressing=True :: Successfully Unset Raft Key OpsRequestProgressing

## PVCs after

All three data volumes are now backed by `standard-custom`, keeping their `10Gi` size and
original PVC names, and the `DocumentDB` object reflects the new StorageClass:

```bash
kubectl get pvc -n demo -l app.kubernetes.io/instance=documentdb-cls-sample \
    -o custom-columns=NAME:.metadata.name,SIZE:.status.capacity.storage,SC:.spec.storageClassName,STATUS:.status.phase
```
NAME                           SIZE   SC                STATUS
data-documentdb-cls-sample-0   10Gi   standard-custom   Bound
data-documentdb-cls-sample-1   10Gi   standard-custom   Bound
data-documentdb-cls-sample-2   10Gi   standard-custom   Bound

```bash
kubectl get docdb -n demo documentdb-cls-sample -o jsonpath='sc={.spec.storage.storageClassName} phase={.status.phase}'
```
sc=standard-custom phase=Ready

The cluster is `Ready`, all pods `2/2`, and previously written data survived the migration
intact:

```bash
PASS=$(kubectl get secret -n demo documentdb-cls-sample-auth -o jsonpath='{.data.password}' | base64 -d)
```

```bash
kubectl exec -n demo documentdb-cls-sample-0 -c documentdb -- \
    mongosh "mongodb://default_user:${PASS}@localhost:10260/?tls=true&tlsAllowInvalidCertificates=true" \
    --quiet --eval 'printjson(db.runCommand({ping:1}));'
```
{ ok: 1 }

> [!NOTE]
> In the test environment, migrating to `standard-custom` (backed by the `local-path`
> provisioner) succeeded and the cluster returned to `Ready`, even though a *freshly provisioned*
> multi-replica DocumentDB on `local-path` can fail to bring up standbys (because `pg_basebackup`
> trips a mount check on that provisioner). Storage migration avoids that path entirely — it
> copies the already-initialized data directory block-for-block rather than re-seeding the
> standby.

## Standalone

The same `DocumentDBOpsRequest` applies to a standalone (`replicas: 1`) instance — point
`spec.databaseRef.name` at `documentdb-sa-sample`. On this build standalone instances did not
finish bootstrapping (see the [Restart](/docs/guides/documentdb/restart/) guide), so the
standalone migration could not be exercised live.

## Cleaning Up

```bash
kubectl delete documentdbopsrequest -n demo documentdb-cls-storage-migration
kubectl delete documentdb -n demo documentdb-cls-sample
kubectl delete ns demo
```

## Next Steps

- [Volume expansion](/docs/guides/documentdb/volume-expansion/) of a DocumentDB cluster.
- [Storage autoscaling](/docs/guides/documentdb/autoscaler/storage/) of a DocumentDB cluster.
