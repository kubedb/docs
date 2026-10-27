---
title: YugabyteDB Storage Migration
menu:
  docs_{{ .version }}:
    identifier: yugabytedb-storage-migration-guide
    name: Storage Migration
    parent: yugabytedb-storage-migration
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Migrate YugabyteDB Volumes to Another StorageClass

> **Not verified on a cluster.** This feature has not been run on a real cluster. What was checked: the admission webhook rules of the API (`apimachinery`, branch `yugabytedb`) have unit tests, and they pass. Everything below describes what the API types and those rules say, not an observed run, so expect to find differences when you try it.

A PVC cannot change its StorageClass in place. A `YugabyteDBOpsRequest` with `spec.type: StorageMigration` moves the data volumes of the masters, the tservers and the read replicas to another StorageClass, **one pod at a time**. It does not use a second PetSet, because the PetSet name is part of every pod address.

## What it does

The step names are from the API constants.

For every tserver (and read replica) pod:

1. `MigrateBlacklistPod`: the tserver is blacklisted so that its tablets move away.
2. `MigrateDrainPod`: waits until the tablets and leaders left the tserver.
3. `MigrateDeletePod`: deletes the pod and its PVCs.
4. `MigrateRecreatePod`: waits for the PetSet to recreate the pod on the new StorageClass.
5. `MigrateUnblacklistPod`: removes the tserver from the blacklist.
6. `MigrateWaitBalancer`: waits until the load balancer is idle before the next pod.

For every master pod, `MigrateMasterReplace` deletes the pod and its PVC and waits until the rebuilt master is a VOTER. The PetSet volumeClaimTemplates are updated first (`MigrateUpdatePetSets`).

**Cost:** every tserver moves all of its data once. The other tservers need enough free capacity to hold the data of the pod that is being migrated.

## The request

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: YugabyteDBOpsRequest
metadata:
  name: yb-storage-migration
  namespace: demo
spec:
  type: StorageMigration
  databaseRef:
    name: yb-demo
  migration:
    storageClassName: fast-ssd
    oldPVReclaimPolicy: Delete
  # every pod moves its data, size the timeout for the whole database
  timeout: 2h
```

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/yugabytedb/storage-migration/storage-migration.yaml
```

Fields of `spec.migration`:

- `storageClassName` (required): the target StorageClass.
- `oldPVReclaimPolicy`: `Retain` or `Delete`; the policy applied to the previous PersistentVolume. By default the policy the PV already had is kept. `Retain` keeps the old volume after the migration.

## Rules checked by the webhook

- `spec.timeout` is **required**: the operation moves the data of every pod, so set it from the size of the database.
- The StorageClass must exist, and at least one role must use a different class than the target.
- If the current class uses `WaitForFirstConsumer`, so must the new one.
- Roles with `storageType: Ephemeral` have no volume to migrate and are rejected.
- A **warning** (not an error) is returned when the database has fewer than `replicationFactor + 1` tservers: the data of a tserver then cannot be moved off before its volume is replaced, and the migration may not complete.

## Cleaning up

```bash
$ kubectl delete ybops -n demo yb-storage-migration
```
