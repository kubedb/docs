---
title: Qdrant Backup Overview
menu:
  docs_{{ .version }}:
    identifier: qdrant-backup-overview
    name: Overview
    parent: qdrant-backup
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Qdrant Backup

This guide summarizes backup approaches for Qdrant and follows the same step-by-step flow as the operations guides.

## Before You Begin

- Install KubeDB, KubeStash, and a CSI snapshot-capable storage plugin if you plan to use volume snapshots.
- Deploy a Qdrant database in namespace `demo`.
- A single example file is not included under `docs/examples/qdrant/backup` because the exact `BackupConfiguration`, storage backend, and snapshot resources depend on your environment.

```bash
kubectl create ns demo
```

## Deploy Qdrant

```bash
kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/examples/qdrant/quickstart/distributed.yaml
kubectl get qdrant -n demo qdrant-sample -w
```

## Apply Backup Resources

## Logical Backup (Restic Plugin)

Use KubeStash with restic backend for object-level backup workflows.

Typical resources:

- `BackupStorage`
- `RetentionPolicy`
- `BackupConfiguration`
- `RestoreSession`

Create backend-specific backup resources that point to your Qdrant database and storage backend.

## Volume Snapshot Backup

Use CSI snapshot compatible storage with snapshot classes for PVC level backup and restore.

Typical resources:

- `VolumeSnapshotClass`
- `BackupConfiguration`
- `RestoreSession`

Use a snapshot-capable `StorageClass` and matching `VolumeSnapshotClass` for the PVCs created by your Qdrant deployment.

## Verify

```bash
kubectl get backupconfiguration -n demo
kubectl get restoresession -n demo
```

## Cleaning up

```bash
kubectl delete backupconfiguration -n demo --all
kubectl delete restoresession -n demo --all
kubectl delete qdrant -n demo qdrant-sample
kubectl delete ns demo
```
