---
title: EtcdArchiver CRD
menu:
  docs_{{ .version }}:
    identifier: etcd-archiver-concepts
    name: EtcdArchiver
    parent: etcd-concepts-etcd
    weight: 35
menu_name: docs_{{ .version }}
section_menu_id: guides
---

<!-- Drafted from source code and CRD schemas; not yet verified against a live cluster. -->

> New to KubeDB? Please start [here](/docs/README.md).

# EtcdArchiver

## What is EtcdArchiver

`EtcdArchiver` is a Kubernetes `Custom Resource Definitions` (CRD). It is a **policy object**: it describes, once, how the etcd clusters that opt into it should be backed up — where the snapshots go, on what schedule, with what retention and encryption — so that individual [Etcd](/docs/guides/etcd/concepts/etcd.md) objects only have to point at it.

When an `Etcd` object references an `EtcdArchiver` through `spec.archiver.ref` (and the archiver's own `spec.databases` selector allows that namespace), the KubeDB provisioner builds a [KubeStash](https://kubestash.com/) `BackupConfiguration` for it.

## Snapshot-only: no PITR

This is the most important thing to know about `EtcdArchiver`, and it is a deliberate scope decision rather than a missing feature.

Databases like Postgres have a continuous, shippable write-ahead log, which is what lets KubeDB offer point-in-time recovery: take a base backup, then replay the log up to any chosen instant. **etcd has no equivalent primitive that can be streamed out of the cluster.** Its Raft log is an internal replication mechanism, not an archivable, restorable log stream.

Consequently:

- `EtcdArchiver.spec` has **no `logBackup` section** — unlike `PostgresArchiver` or `MSSQLServerArchiver`.
- There is **no point-in-time recovery** between two snapshots. Recovery lands you exactly on the snapshot you restore, and everything written after it is lost.
- Your recovery point objective is therefore your snapshot interval. Choose `fullBackup.scheduler.schedule` accordingly.

Also note that a snapshot is always replayed into an **empty** data directory: etcd requires a seed member's data directory to be either genuinely empty or a fully restored, valid snapshot, so it can never be replayed into a running member. There are two ways to get there — restore at bootstrap through [spec.init](/docs/guides/etcd/concepts/etcd.md#specinit), or an `EtcdOpsRequest` of type `Restore`, which tears an existing cluster down to one empty volume first. See [In-place Restore](/docs/guides/etcd/restore/overview.md).

## EtcdArchiver CRD Specifications

Like any official Kubernetes resource, an `EtcdArchiver` has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

```yaml
apiVersion: archiver.kubedb.com/v1alpha1
kind: EtcdArchiver
metadata:
  name: etcd-archiver
  namespace: demo
spec:
  pause: false
  databases:
    namespaces:
      from: Selector
      selector:
        matchLabels:
          kubernetes.io/metadata.name: demo
    selector:
      matchLabels:
        archiver: etcd
  fullBackup:
    driver: Restic
    scheduler:
      successfulJobsHistoryLimit: 1
      failedJobsHistoryLimit: 1
      schedule: "0 */2 * * *"
    sessionHistoryLimit: 3
  manifestBackup:
    scheduler:
      successfulJobsHistoryLimit: 1
      failedJobsHistoryLimit: 1
      schedule: "0 */2 * * *"
    sessionHistoryLimit: 3
  backupStorage:
    ref:
      name: s3-storage
      namespace: demo
  retentionPolicy:
    name: keep-1mo
    namespace: demo
  encryptionSecret:
    name: encrypt-secret
    namespace: demo
  deletionPolicy: WipeOut
```

Here, we are going to describe the various sections of an `EtcdArchiver` crd.

### spec.databases

`spec.databases` is a required field that defines which `Etcd` objects are allowed to consume this archiver. It is a double opt-in: the `Etcd` object has to reference the archiver in `spec.archiver.ref`, **and** the archiver has to allow the database's namespace here.

- `namespaces.from` can be `Same` (only this namespace), `All`, or `Selector`.
- `namespaces.selector` is the label selector applied to namespaces when `from` is `Selector`.
- `selector` further restricts which `Etcd` objects inside those namespaces may use this archiver.

### spec.pause

`spec.pause` is an optional boolean. When set to `true`, the backup driven by this archiver is stopped for every database that consumes it. To pause the backup of just one database, set `spec.archiver.pause` on that `Etcd` object instead.

### spec.fullBackup

`spec.fullBackup` defines the session configuration of the full snapshot backup. Its options end up on the generated KubeStash `BackupConfiguration`'s `full` session.

- `driver` is the KubeStash driver. For etcd this is the `Restic` driver: the backup task streams a live etcd snapshot out of the cluster and uploads it to the backend, rather than taking a volume snapshot.
- `scheduler.schedule` is the cron expression for the snapshot. Because there is no PITR, this interval **is** your recovery point objective.
- `scheduler.concurrencyPolicy`, `scheduler.jobTemplate`, `containerRuntimeSettings`, `jobTemplate`, `retryConfig`, `timeout` and `sessionHistoryLimit` follow the usual KubeStash semantics.

The full backup runs the `etcd-backup` addon task. It takes a snapshot through etcd's client API — the same operation as `etcdctl snapshot save` — so it is served by a live member and does not require access to any member's data directory.

### spec.manifestBackup

`spec.manifestBackup` defines the session configuration of the manifest backup — the Kubernetes-object-level backup of the `Etcd` CR and its companion objects (Secrets, etc.). It runs the same cross-database `manifest-backup` task every KubeDB database uses, unchanged.

### spec.backupStorage

`spec.backupStorage` is the backend where the snapshots are written.

- `ref` is the name/namespace reference of the KubeStash `BackupStorage` object.
- `subDir` optionally scopes this archiver's repositories to a sub-directory of that storage.

### spec.retentionPolicy

`spec.retentionPolicy` is a name/namespace reference to a KubeStash `RetentionPolicy` object, which decides how many snapshots are kept in the repository.

### spec.encryptionSecret

`spec.encryptionSecret` is a name/namespace reference to the Secret holding the Restic encryption key used for the repository. Keep this Secret safe and backed up separately — without it the snapshots cannot be restored.

### spec.deletionPolicy

`spec.deletionPolicy` is the `deletionPolicy` applied to the repositories created by this archiver, and follows KubeStash's `BackupConfigDeletionPolicy` semantics (for example `Delete` or `WipeOut`).

## EtcdArchiver `Status`

### status.databaseRefs

`status.databaseRefs` lists the `Etcd` objects currently managed by this archiver, so you can see at a glance which databases a policy is actually covering.

## Restoring

Restore is driven from the `Etcd` object being created, not from the archiver:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Etcd
metadata:
  name: etcd-restored
  namespace: demo
spec:
  version: 3.6.4
  replicas: 3
  storage:
    storageClassName: "standard"
    accessModes:
      - ReadWriteOnce
    resources:
      requests:
        storage: 1Gi
  init:
    archiver:
      encryptionSecret:
        name: encrypt-secret
        namespace: demo
      fullDBRepository:
        name: etcd-full-repo
        namespace: demo
      recoveryTimestamp: "2026-01-14T09:41:52Z"
  deletionPolicy: WipeOut
```

When `spec.init.archiver.fullDBRepository` is set, the provisioner pre-creates the ordinal-0 PVC, runs a KubeStash `RestoreSession` against it *before* the PetSet or the first pod is ever created, and only then bootstraps the cluster from the restored data directory. The remaining members join the restored seed as ordinary new members. A separate `manifest-restore` session for Kubernetes-object-level restore runs only if you also set `spec.init.archiver.manifestRepository`; setting `fullDBRepository` alone does not trigger one.

To restore a snapshot into an `Etcd` object that **already exists**, use an `EtcdOpsRequest` of type `Restore` instead — it replaces the whole keyspace of the live database. See [In-place Restore](/docs/guides/etcd/restore/overview.md).

## Next Steps

- Learn about the [Etcd](/docs/guides/etcd/concepts/etcd.md) crd.
- Follow the [etcd backup & restore guide](/docs/guides/etcd/backup/kubestash/overview/index.md) for an end-to-end walkthrough.
- Learn more about [KubeStash](https://kubestash.com/).
