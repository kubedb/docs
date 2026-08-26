---
title: Etcd Storage Migration Overview
menu:
  docs_{{ .version }}:
    identifier: etcd-storage-migration-overview
    name: Overview
    parent: etcd-storage-migration
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

<!-- Drafted from source code and CRD schemas; not yet verified against a live cluster. -->

> New to KubeDB? Please start [here](/docs/README.md).

# Etcd Storage Migration

This guide gives an overview of how the KubeDB Ops-manager operator moves the data volumes of an
`Etcd` cluster from one `StorageClass` to another.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Etcd](/docs/guides/etcd/concepts/etcd.md)
  - [EtcdOpsRequest](/docs/guides/etcd/concepts/etcdopsrequest.md)

## Why It Is Not Just an Edit

`spec.storageClassName` on a PVC is immutable, and so is `spec.volumeClaimTemplates` on a `PetSet`.
Moving an etcd member onto a different `StorageClass` therefore means provisioning a **new** volume,
copying the data across, and giving the new volume the old claim's name so the member picks it up
again. `StorageMigration` is the `EtcdOpsRequest` type that does all of that for you.

## How the Storage Migration Process Works

The mechanics are the shared KubeDB migration machinery, identical to every other KubeDB database.
The only etcd-specific part is the **ordering**: members are migrated one at a time, **followers
first and the Raft leader last**, so the cluster keeps a quorum throughout.

1. The user creates an `EtcdOpsRequest` of type `StorageMigration` with
   `spec.migration.storageClassName` (the target `StorageClass`) and a `spec.timeout`.

2. The Ops-manager operator pauses the `Etcd` object, so the Provisioner operator does not fight the
   migration.

3. It deletes the `PetSet` **with an orphan propagation policy**. The `PetSet` has to go first —
   otherwise it would immediately recreate every pod the migration deletes — but the pods are
   orphaned, not deleted, so every member keeps serving until its own turn comes.

4. It resolves the current Raft leader and splits the members into followers and the leader, so the
   leader's volume is migrated last.

5. For each member, in that order:
   - create a new PVC on the target `StorageClass`, sized from the old one;
   - if the target `StorageClass` uses `WaitForFirstConsumer`, run a short-lived mounter pod pinned to
     the member's node, so the new volume is provisioned in the right topology;
   - stash the member pod's manifest as an annotation on the ops request and delete the pod;
   - run a migrator `Job` that `rsync`s the data directory from the old volume to the new one;
   - swap the names: delete the original PVC (with its PV temporarily set to `Retain`, so no data is
     lost) and re-create the claim under the original name bound to the new volume;
   - recreate the member pod from the stashed manifest and wait for it to become `Ready`;
   - **wait for a healthy quorum before moving to the next member.**

6. Once every member has been migrated, the operator patches
   `spec.storage.storageClassName` on the `Etcd` object, resumes it, and marks the request
   `Successful`.

## Preconditions

- `spec.migration.storageClassName` is required, and the target `StorageClass` must exist.
- **`spec.timeout` is required.** The webhook rejects a `StorageMigration` without it. Every step of
  the migration is bounded by that timeout, and the data-copy `Job` has no other deadline at all — so
  size it against how much data you have, not against how long a pod takes to restart.
- `spec.storageType` must be `Durable`; there is nothing to migrate on an `Ephemeral` etcd.
- If the **source** `StorageClass` uses `volumeBindingMode: WaitForFirstConsumer`, the target must use
  it too. The webhook rejects the mismatch.

> **Keeping the source volumes.** While the claims are being swapped, each PV's reclaim policy is
> temporarily forced to `Retain` so nothing is lost mid-flight, and the original policy is restored
> once the member is back — which means a source PV provisioned by a `Delete`-reclaiming
> `StorageClass` is eventually reclaimed. Set `spec.migration.oldPVReclaimPolicy: Retain` if you want
> the previous volumes to survive the migration.

In the [next](/docs/guides/etcd/storage-migration/storage-migration.md) doc, we are going to show a
step-by-step guide on migrating the storage of an Etcd cluster using the `EtcdOpsRequest` CRD.
