---
title: Adopt a Helm-managed Milvus into KubeDB
description: A runbook for moving a Helm-deployed Milvus under KubeDB management against the same object-storage data, including moving its etcd metadata
menu:
  docs_{{ .version }}:
    identifier: milvus-adoption-guide
    name: Guide
    parent: milvus-adoption
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

<!-- Drafted from source code and CRD schemas; no step in this runbook has been verified against a live cluster. -->

> New to KubeDB? Please start [here](/docs/README.md).

# Adopt a Helm-managed Milvus into KubeDB

This runbook describes how to move a Milvus that was installed with the upstream Helm chart under KubeDB management, keeping its data in the same object-storage bucket and moving its etcd metadata into a KubeDB-managed `Etcd`.

> **Status: a plan, not a verified procedure.** Nothing here has been run on a live cluster. Two capabilities the procedure depends on do **not exist in KubeDB yet** (see [What KubeDB cannot do yet](#what-kubedb-cannot-do-yet)), and the guide marks each step that is unverified. Rehearse the whole procedure on a copy of your deployment first. Do not run it for the first time against production.

## What changes and what stays

| | Stays where it is | Is replaced |
|---|---|---|
| Segments, indexes and write-ahead log | In your object-storage bucket, under the same root path. KubeDB does not convert or copy them | Nothing |
| Milvus pods | | Every role's pods are replaced once by KubeDB-managed PetSets named `<milvus-name>-<role>`. Expect a short interruption per role |
| Metadata | In the Helm deployment's etcd until step 2, which moves it | The etcd that holds it, after step 2 |

KubeDB does not take ownership of the Helm release's pods. It creates its own, so the Helm release and the KubeDB `Milvus` exist side by side until you remove the Helm release.

## What KubeDB cannot do yet

Be clear about these before you plan a maintenance window:

1. **Client TLS and credentials for an external etcd are not rendered.** With `spec.metaStorage.externallyManaged: true`, KubeDB writes only the etcd endpoints into `milvus.yaml`. The `etcd.auth` and `etcd.ssl` blocks are generated only for a KubeDB-managed etcd. A Helm etcd that requires authentication or mutual TLS therefore cannot be used as an external store today. A change that adds this is being worked on separately; it is not available in this release.
2. **The etcd root path cannot be set.** KubeDB always renders `etcd.rootPath` as `<milvus-name>-etcd`, whatever the Milvus manifest says. The upstream Helm chart uses `by-dev`. Metadata written under `by-dev` is not found by a KubeDB Milvus that reads `<milvus-name>-etcd`. A typed field for the root path is being added separately; it is not available in this release.

Until both land, the procedure below cannot reuse the Helm deployment's metadata in place. You can still adopt the **data** into a fresh KubeDB Milvus with a [logical backup and restore](/docs/guides/milvus/backup/guide.md), which copies collections through the Milvus API instead of reusing metadata.

## Before You Begin

- KubeDB installed with the `Milvus` and `Etcd` feature gates on, and KubeStash installed. See [Prepare Dependencies](/docs/guides/milvus/quickstart/prerequisites.md).
- A **full copy** of your object-storage data and an `etcdctl snapshot` of the Helm etcd, taken before you start. The rollback below depends on them.
- The Helm deployment's S3 settings: address, port, bucket, root path (`minio.rootPath`), TLS and credentials, and which write-ahead log it uses (Woodpecker or RocksMQ).
- The Helm etcd's endpoints, root path, version, and whether it uses authentication or TLS.
- Check that the Helm etcd's version is one KubeDB's `EtcdVersion` catalog carries (see `kubectl get etcdversion`). The target `Etcd` must run a version your Milvus version supports.

## Step 1: Create a KubeDB Milvus against the existing data

### 1a. Point at the existing object storage

`spec.objectStorage.configSecret` names a Secret whose keys are copied verbatim into the `minio:` section of `milvus.yaml`. Use the Helm deployment's own values, including the **same `rootPath`** so KubeDB finds the existing objects:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: milvus-adopt-storage
  namespace: demo
type: Opaque
stringData:
  address: "s3.example.internal"
  port: "443"
  useSSL: "true"
  accessKeyID: "<access key>"
  secretAccessKey: "<secret key>"
  bucketName: "milvus"
  rootPath: "<the Helm deployment's minio.rootPath>"
```

Keys are written to `milvus.yaml` by name, and `port` is converted to a number. The key names must be the `milvus.yaml` `minio:` keys (`address`, `port`, `useSSL`, `accessKeyID`, `secretAccessKey`, `bucketName`, `rootPath`, and so on). There is no typed field for them, and no field for KMS settings, which would go through `spec.configuration`.

### 1b. Point at the existing etcd, or let KubeDB create one

There are two ways to start, and they trade risk differently.

**Option A: start on a new, empty KubeDB-managed etcd.** Omit `spec.metaStorage`. KubeDB creates an `Etcd` named `<milvus-name>-etcd` (3 members by default) and renders its TLS and credentials into `milvus.yaml`. The new Milvus has no metadata of the Helm deployment's collections; you populate it in step 2.

**Option B: start on the Helm etcd.** Set `spec.metaStorage.externallyManaged: true` with the Helm etcd's endpoints:

```yaml
spec:
  metaStorage:
    externallyManaged: true
    endpoints:
      - http://helm-milvus-etcd-0.demo.svc.cluster.local:2379
      - http://helm-milvus-etcd-1.demo.svc.cluster.local:2379
      - http://helm-milvus-etcd-2.demo.svc.cluster.local:2379
```

Option B is subject to both limits above: it only works if the Helm etcd needs no authentication or TLS, and KubeDB reads the metadata from `<milvus-name>-etcd`, not `by-dev`. Do not use it for production metadata until both are fixed. Two Milvus control planes (the Helm one and the KubeDB one) writing to one etcd root is not something this guide has verified is safe.

A KubeDB Milvus that uses the archiver (point-in-time recovery) must use a KubeDB-managed, dedicated etcd, so externally managed etcd is a stepping stone, not the destination.

### 1c. Create the Milvus

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Milvus
metadata:
  name: milvus-adopted
  namespace: demo
spec:
  version: "2.6.11"      # use the catalog version that matches your Helm Milvus
  topology:
    mode: Distributed
    distributed:
      mixcoord:
        replicas: 2
      datanode:
        replicas: 2
      querynode:
        replicas: 2
      streamingnode:
        replicas: 3
        storageType: Durable
        storage:
          accessModes: [ReadWriteOnce]
          resources:
            requests:
              storage: 10Gi
      proxy:
        replicas: 2
  objectStorage:
    configSecret:
      name: milvus-adopt-storage
  deletionPolicy: Halt
```

Use a `deletionPolicy` that does not delete data (`Halt` or `DoNotTerminate`) while you rehearse. The Milvus catalog version must match the Helm deployment's Milvus version; check `kubectl get milvusversion`.

**Not verified:** moving roles over one at a time, and running the KubeDB-managed and Helm-managed roles of one cluster against the same etcd and bucket at once. The KubeDB Milvus creates all of its roles together, so a staged, role-by-role move would need you to scale the Helm roles down as the KubeDB roles come up. Whether Milvus tolerates the two control planes overlapping is untested. The safest approach to validate is to stop the Helm release's writers before the KubeDB roles start serving.

## Step 2: Move the metadata into the KubeDB-managed etcd

This step is only needed if you started on the Helm etcd (Option B) or want the Helm deployment's metadata inside the KubeDB `Etcd`. Its stages are:

1. **Freeze writes.** Stop clients writing to Milvus and stop the Helm release's coordinators and data roles so nothing writes metadata while you snapshot. The write freeze is the outage window of the etcd move.
2. **Snapshot the Helm etcd** and put the snapshot into a KubeStash `Repository`.
3. **Restore the snapshot into the KubeDB `Etcd`** with an `EtcdOpsRequest` of type `Restore`.
4. **Repoint Milvus** at the KubeDB `Etcd` and start it.

### 2a. Snapshot the Helm etcd into a KubeStash repository

The KubeStash etcd addon (`etcd-addon`, task `etcd-backup`) streams a snapshot off an etcd endpoint and stores it, encrypted, in a KubeStash `Repository` under the file name `snapshot.db`. Its parameters include `endpoint` (an etcd client endpoint to snapshot from, which defaults to the target `AppBinding`'s URL), `user`, `dialTimeout` and `insecureSkipTLSVerify`.

The intended route is to create an `AppBinding` for the Helm etcd and a `BackupConfiguration` whose target is that `AppBinding`, using the `etcd-backup` task with `endpoint` set to the Helm etcd.

**Not verified:** that the plugin snapshots a non-KubeDB etcd this way, including how it picks up client TLS and credentials for it. There is no documented procedure for importing a snapshot that was taken outside KubeStash. A snapshot taken with `etcdctl snapshot save` is a useful **rollback artifact**, but this guide does not claim you can load it into a `Repository` by hand. Settle this on a copy first. The restore side finds the snapshot by looking for a file named `snapshot.db` under the restored directory, which is the shape the KubeStash backup produces.

### 2b. Restore into the KubeDB Etcd

The `Etcd` that KubeDB created for the Milvus is named `<milvus-name>-etcd`. Restore the snapshot into it in place:

```yaml
apiVersion: ops.kubedb.com/v1alpha1
kind: EtcdOpsRequest
metadata:
  name: milvus-adopted-etcd-restore
  namespace: demo
spec:
  type: Restore
  databaseRef:
    name: milvus-adopted-etcd
  restore:
    fullDBRepository:
      name: helm-etcd-full
      namespace: demo
    encryptionSecret:
      name: encrypt-secret
      namespace: demo
  timeout: 30m
  apply: Always
```

This is the in-place restore described in [Restore a Snapshot into an Existing Etcd Cluster](/docs/guides/etcd/restore/restore.md). The fields are:

- `spec.restore.fullDBRepository` is **required**. It names the KubeStash `Repository` that holds the snapshot.
- `spec.restore.recoveryTimestamp` is optional. Leave it out to use the latest snapshot.
- `spec.restore.encryptionSecret` is the Secret with the restic password the snapshot was written with.
- `spec.timeout` is **required** for `Restore`, and `spec.apply: Always` is the safe choice because the cluster is usually degraded during a restore.

**It replaces the entire keyspace of the `Etcd`. There is no undo.** The operation takes the cluster apart down to a single empty volume, writes the snapshot into it, starts a single-member cluster, and lets the ordinary membership reconciliation grow it back to the configured number of members. `spec.storageType` must be `Durable`.

**Not verified:** how the Milvus controller behaves while this runs, that is, whether it keeps reconciling the `Etcd` it owns. Pause the Milvus (`spec.halted: true`) for the duration, and check the result.

### 2c. The root path

After the restore, the keys are under the root path the Helm deployment used (`by-dev` by default), but a KubeDB Milvus reads `<milvus-name>-etcd`. This is the second limit above. Until a typed root-path field exists, you would have to rewrite the key prefix yourself between the snapshot and the restore. This guide does not provide or verify such a procedure.

### 2d. Start Milvus and check

Un-halt the Milvus and check collection and row counts against the numbers you recorded before the freeze. A move succeeds only if every collection and its row count are unchanged and queries return the same results. No reindex should be needed, because the indexes stay in the bucket.

## Rollback

The rollback boundary is **the point where you delete the Helm release and its etcd**. Before that point, rollback is straightforward:

1. Halt or delete the KubeDB `Milvus` (with a `deletionPolicy` that keeps the bucket data, which is why the example uses `Halt`).
2. Start the Helm release's writers again, pointed at its original etcd. This procedure never writes to that etcd: the snapshot only reads from it, and the restore writes into the separate KubeDB `Etcd`. If you keep the Helm etcd read-only (or its writers stopped) from the freeze until the cutover is accepted, it is exactly the state you froze.

After you delete the Helm release and its etcd, rollback is a restore from the backups you took in [Before You Begin](#before-you-begin). Keep the Helm release (scaled to zero) and its etcd for at least one full backup cycle of the new deployment before you remove them.

## After the move

- Enable continuous backup and point-in-time recovery with a `MilvusArchiver`. See the [backup guide](/docs/guides/milvus/backup/guide.md). A Standalone Milvus created by an older operator keeps its write-ahead log in RocksMQ and cannot use the archiver; check which write-ahead log your Helm deployment uses before you plan on point-in-time recovery.
- Configure per-role pod settings, GPU and SR-IOV networking as described in the [GPU and SR-IOV guide](/docs/guides/milvus/gpu-sriov/guide.md).

## Related

- [Prepare Milvus Dependencies](/docs/guides/milvus/quickstart/prerequisites.md)
- [Backup & Restore Milvus](/docs/guides/milvus/backup/guide.md)
- [Restore a Snapshot into an Existing Etcd Cluster](/docs/guides/etcd/restore/restore.md)
- [EtcdOpsRequest](/docs/guides/etcd/concepts/etcdopsrequest.md)
