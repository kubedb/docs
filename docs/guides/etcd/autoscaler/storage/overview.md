---
title: Etcd Storage Autoscaling Overview
menu:
  docs_{{ .version }}:
    identifier: etcd-autoscaling-storage-overview
    name: Overview
    parent: etcd-autoscaling-storage
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

<!-- Drafted from source code and CRD schemas; not yet verified against a live cluster. -->

> New to KubeDB? Please start [here](/docs/README.md).

# Etcd Storage Autoscaling

This guide gives an overview of how the KubeDB Autoscaler operator autoscales the storage of an
etcd cluster using the `EtcdAutoscaler` CRD.

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Etcd](/docs/guides/etcd/concepts/etcd.md)
  - [EtcdAutoscaler](/docs/guides/etcd/concepts/etcdautoscaler.md)
  - [EtcdOpsRequest](/docs/guides/etcd/concepts/etcdopsrequest.md)

## How Storage Autoscaling Works

The autoscaling process consists of the following steps:

1. A user creates an `Etcd` Custom Resource (CR).

2. The `KubeDB` Provisioner operator watches the `Etcd` CR.

3. When the operator finds an `Etcd` CR, it creates the `PetSet` and the related resources.
   Each ordinal of the PetSet gets a PersistentVolume from the volume claim template, which is
   where the etcd member's data directory (the bbolt backend and the WAL) lives.

4. To set up storage autoscaling for that cluster, the user creates an `EtcdAutoscaler` CRO with
   the desired configuration.

5. The `KubeDB` Autoscaler operator watches the `EtcdAutoscaler` CRO.

6. The `KubeDB` Autoscaler operator continuously reads the PVC usage of the database's pods from
   the `custom.metrics.k8s.io` API backed by the KubeDB storage-metrics apiserver and compares it
   against `spec.storage.etcd.usageThreshold`.
   - If the usage exceeds the threshold, the operator computes a new size and creates an
     `EtcdOpsRequest` of type `VolumeExpansion` to expand the storage.

7. The `KubeDB` Ops-manager operator watches that `EtcdOpsRequest` CRO.

8. The `KubeDB` Ops-manager operator then expands the volumes of the etcd members as specified in
   the `EtcdOpsRequest` CRO.

## Why storage autoscaling matters more for etcd

etcd is unusually sensitive to a full disk. Two properties are worth keeping in mind:

- The backend database file never shrinks on its own. Deleting keys, and even compacting the
  keyspace history, only marks pages as reusable inside the file — the file keeps its size until
  it is defragmented. See [Maintenance Overview](/docs/guides/etcd/maintenance/overview.md).
- When the backend grows past `--quota-backend-bytes` (`spec.configuration.tuning.quotaBackendBytes`),
  etcd raises a cluster-wide `NOSPACE` alarm and the cluster goes read-only until the alarm is
  cleared.

Storage autoscaling gives the volume headroom; it does not by itself reclaim space inside the
backend file. The two are complementary — use autoscaling for the volume, and
`Compact` + `Defragment` for the backend file.

## The generated EtcdOpsRequest

As with compute autoscaling, the autoscaler never edits the `Etcd` object directly. When the
usage threshold is crossed it creates an `EtcdOpsRequest`:

- Named with the `etcdops-` prefix plus a random suffix, for example
  `etcdops-etcd-autoscale-w7q2rk`.
- Owned by the `EtcdAutoscaler` through an owner reference.
- With `spec.type: VolumeExpansion`, `spec.volumeExpansion.etcd` set to the newly computed size,
  and `spec.volumeExpansion.mode` copied from `spec.storage.etcd.expansionMode`.
- With `spec.timeout`, `spec.apply` and `spec.maxRetries` taken from
  `EtcdAutoscaler.spec.opsRequestOptions`, when that block is present.

The request is only created when the newly computed size is actually **larger** than the current
`spec.storage.resources.requests.storage` of the `Etcd` object; storage is never scaled down.

In the next doc, we show a step-by-step guide for autoscaling the storage of an etcd cluster
using the `EtcdAutoscaler` CRD.

## Next Steps

- [Autoscale the storage of an Etcd cluster](/docs/guides/etcd/autoscaler/storage/storage-autoscale.md).
- [Compute Autoscaling Overview](/docs/guides/etcd/autoscaler/compute/overview.md).
