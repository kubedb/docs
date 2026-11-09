---
title: Branch Snapshot Configuration
menu:
  docs_{{ .version }}:
    identifier: branching-snapshots
    name: Snapshot Configuration
    parent: branching
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: operatormanual
---

> New to KubeDB? Please start [here](/docs/README.md).

# Branch Snapshot Configuration

Branching uses CSI snapshots and cloned volumes. Configure **Ceph RBD or equivalent CoW-capable CSI storage** before creating a Branch. The examples use a StorageClass named `ceph-rbd`.

## Required snapshot support

The cluster needs the CSI snapshot CRDs, snapshot controller, and the driver's snapshot support. The StorageClass provisioner and VolumeSnapshotClass driver must match.

```bash
kubectl get storageclass ceph-rbd -o yaml
kubectl get volumesnapshotclass -o yaml
kubectl get crd volumesnapshots.snapshot.storage.k8s.io \
  volumesnapshotcontents.snapshot.storage.k8s.io \
  volumesnapshotclasses.snapshot.storage.k8s.io
```

For example, a Ceph RBD snapshot class can use:

```yaml
apiVersion: snapshot.storage.k8s.io/v1
kind: VolumeSnapshotClass
metadata:
  name: ceph-rbd-snapclass
  annotations:
    snapshot.storage.kubernetes.io/is-default-class: "true"
driver: rbd.csi.ceph.com
deletionPolicy: Delete
parameters:
  csi.storage.k8s.io/snapshotter-secret-name: rook-csi-rbd-provisioner
  csi.storage.k8s.io/snapshotter-secret-namespace: rook-ceph
```

Use your actual Ceph CSI driver name and snapshotter Secret. A Rook installation can use a namespace-prefixed driver name; copy the provisioner from your StorageClass instead of assuming `rbd.csi.ceph.com`.

## Snapshot-class selection

Courier selects a compatible VolumeSnapshotClass for the source driver, preferring its default class. Set `spec.volumeSnapshotClassName` on the Branch to select an explicit class. For a branch into another namespace or cluster, that class name is also used for the importing snapshots and must exist wherever it is referenced.

When multiple source data volumes are involved, Courier uses a VolumeGroupSnapshot when a compatible group-snapshot class is available; otherwise it creates individual VolumeSnapshots. Group snapshots require the group-snapshot APIs and controller/driver support. Individual snapshots do not provide an atomic snapshot across volumes. Choose the storage and consistency guarantees appropriate for your database workload.

## Branching into another namespace

A PVC can reference a VolumeSnapshot only in its own namespace. Courier imports the source snapshot handle into the target namespace using a VolumeSnapshotContent and a corresponding VolumeSnapshot before creating the target PVCs. Users create the Branch; Courier manages these snapshot resources.

## Cross-cluster backend compatibility

Both clusters must reach **the same Ceph backend over the network**. Source and target nodes and CSI components need backend access and appropriate credentials. The source database's PVCs must already use this shared backend.

Verify compatible settings on both clusters:

```bash
for context in source target; do
  kubectl --context "$context" get storageclass ceph-rbd \
  -o jsonpath='{.provisioner}{" "}{.parameters.clusterID}{" "}{.parameters.pool}{" "}{.parameters.radosNamespace}{"\n"}'
done
```

For Ceph RBD, check the CSI driver identity, backend cluster identity, pool, RADOS namespace, and CSI instance identity. Also check CSI configuration and credentials beyond the StorageClass fields. A snapshot handle created on the source must resolve on the target against the same storage metadata.

StorageClass names alone do not establish compatibility. Using identical class names for two independent Ceph installations does not share snapshots. Courier's cross-cluster control plane delivers snapshot references; the shared backend provides the data and CoW clone mechanism.

See the [PostgreSQL configuration guide](/docs/guides/postgres/branch/configuration.md) for a Branch with an explicit snapshot-class override.
