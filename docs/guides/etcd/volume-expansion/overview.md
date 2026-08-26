---
title: Etcd Volume Expansion Overview
menu:
  docs_{{ .version }}:
    identifier: etcd-volume-expansion-overview
    name: Overview
    parent: etcd-volume-expansion
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

<!-- Drafted from source code and CRD schemas; not yet verified against a live cluster. -->

> New to KubeDB? Please start [here](/docs/README.md).

# Etcd Volume Expansion

This guide gives an overview of how the KubeDB Ops-manager operator expands the data volumes of an
`Etcd` cluster.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Etcd](/docs/guides/etcd/concepts/etcd.md)
  - [EtcdOpsRequest](/docs/guides/etcd/concepts/etcdopsrequest.md)

## How Volume Expansion Process Works

Each etcd member owns one PVC, provisioned from the `PetSet`'s volume claim template and named
`data-<db-name>-<ordinal>` (for example `data-etcd-cluster-0`). It holds etcd's data directory
(`/var/lib/etcd`) — the WAL and the backend database file.

The volume expansion process consists of the following steps:

1. At first, a user creates an `Etcd` Custom Resource (CR).

2. The `KubeDB` Provisioner operator watches the `Etcd` CR and creates the `PetSet`, whose volume
   claim template provisions one PVC per member.

3. In order to expand those volumes, the user creates an `EtcdOpsRequest` CR of type
   `VolumeExpansion` with the desired size and a mode.

4. The `KubeDB` Ops-manager operator watches the `EtcdOpsRequest` CR.

5. When it finds one, it **pauses** the referenced `Etcd` object, so the Provisioner operator does not
   reconcile it during the expansion, and records the current member count on the ops request (it has
   to be restored later — see step 8).

6. It then expands the volumes in the requested mode (`Online` or `Offline`, see below): each member's
   PVC is patched with the new storage request, and the operator waits until the PVC's
   `status.capacity` actually reflects the new size — the CSI driver, not KubeDB, does the growing.

7. Once every PVC has grown, the operator deletes the `PetSet` **with an orphan propagation policy**.
   This step is unavoidable: `spec.volumeClaimTemplates` is immutable, so the `PetSet` has to be
   recreated to carry the new size. Orphaning means the member pods are *not* deleted along with it —
   which is what keeps an `Online` expansion non-disruptive.

8. The operator persists the new size onto `spec.storage.resources.requests.storage` of the `Etcd`
   object and re-renders the `PetSet` from it. A freshly created `PetSet` is seeded with a single
   replica (the bootstrap seed), so the operator immediately restores the member count it recorded in
   step 5, while the database is still paused — this keeps the `PetSet` controller from deleting the
   higher-ordinal pods in between.

9. Finally, the operator resumes the `Etcd` object and marks the `EtcdOpsRequest` `Successful`.

## Volume Expansion Modes

`spec.volumeExpansion.mode` is **required** and selects how the members are treated while their
volumes grow:

- **`Online`** — the member pods keep running throughout. The operator patches every PVC and lets the
  CSI driver grow the filesystem underneath the live pods. Because the `PetSet` is deleted with
  orphaned pods and recreated afterwards, etcd never stops serving and quorum is never affected. This
  requires a `StorageClass` with `allowVolumeExpansion: true` **and** a CSI driver that supports
  online (mounted) filesystem expansion.

- **`Offline`** — the operator first scales the `PetSet` to `0` and waits for every member pod to
  terminate, then patches the PVCs, then recreates everything. Use this when the CSI driver requires
  the volume to be unmounted before it can be resized. The cluster is **fully unavailable** for the
  duration. This is safe for etcd — the Raft log and the snapshot live on the very volumes being
  expanded, so a full cluster stop loses nothing, it just loses availability.

Note that in both modes the PVCs are patched for **all members together**, not one member at a time.
`Online` is therefore not a rolling operation with a per-member quorum gate; it is a single
non-disruptive resize of every volume.

## Preconditions

- `spec.storageType` must be `Durable`. There is nothing to expand on an `Ephemeral` etcd, and the
  ops request fails outright.
- The `StorageClass` backing the volumes must have `allowVolumeExpansion: true`.
- Volumes can only grow. The validating webhook rejects a `spec.volumeExpansion.etcd` value that is
  smaller than the current request.

> **Sizing tip.** etcd's disk usage is bounded by its backend quota
> (`spec.configuration.tuning.quotaBackendBytes`), not by how much data you write over time — but the
> on-disk file only shrinks after a `Compact` followed by a `Defragment`. If you are expanding
> because the backend file grew, consider whether compaction/defragmentation is the actual fix before
> buying more disk.

In the [next](/docs/guides/etcd/volume-expansion/volume-expansion.md) doc, we are going to show a
step-by-step guide on expanding the volumes of an Etcd cluster using the `EtcdOpsRequest` CRD.
