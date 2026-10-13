---
title: Backup & Restore Milvus | KubeStash
description: Back up and restore Milvus with KubeStash - logical and physical backups, MilvusArchiver and point-in-time recovery
menu:
  docs_{{ .version }}:
    identifier: milvus-backup-guide
    name: Guide
    parent: milvus-backup
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

<!-- Drafted from source code, CRD schemas and the operator's sample manifests; not yet verified against a live cluster. -->

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="Please install [KubeStash](https://kubestash.com/docs/latest/setup/install/kubestash/) to try this feature. Database backup with KubeStash is already included in the KubeDB license. So, you don't need a separate license for KubeStash." >}}

# Backup & Restore Milvus using KubeStash

A Milvus deployment keeps its durable state in two places: the **metadata** in its etcd, and the **segments, indexes and write-ahead log** in object storage. A backup is only useful if it captures both at one consistent moment. This guide explains the three ways KubeDB backs Milvus up, which one to pick, and how to restore.

> **Maturity.** The Milvus backup plugin images that the addon references are tagged `ghcr.io/kubedb/milvus-backup-plugin:v0.1.0-rc.0`, a release candidate. Everything below is derived from the addon, the operator source and the CRD schemas. It has **not** been verified on a live cluster, and the sections [What is not confirmed](#what-is-not-confirmed) lists the points that are still open.

## Which backup should I use?

| | Logical backup | Physical backup | Archiver (physical backups + point-in-time recovery) |
|---|---|---|---|
| KubeStash task | `logical-backup` | `physical-backup` | `fullBackup` + `logBackup` + `manifestBackup` of a `MilvusArchiver` |
| How it reads Milvus | Through the Milvus API | Copies the etcd metadata and the object-storage objects directly | Same as physical, repeated on a schedule, plus a continuous change log |
| Scope | Chosen collections (default: all) | The whole deployment | The whole deployment |
| Restore result | New collections, named with a suffix (`_restored` by default) | A whole Milvus, with the same data | A whole Milvus, at a chosen point in time |
| You write the `BackupConfiguration` | Yes | Yes | No. The operator generates it |
| Use it for | Moving collections between clusters; migrating a legacy RocksMQ Standalone | A single consistent full copy | Production: scheduled base backups with point-in-time recovery |

The addon all three use is `milvus-addon`, shipped in the KubeDB installer under `catalog/kubestash/raw/milvus`. It defines these tasks:

| Direction | Task | Driver | What it is for |
|---|---|---|---|
| Backup | `logical-backup` | `Milvus` | Collection-level backup through the Milvus API |
| Backup | `physical-backup` | `Restic` | Consistent copy of etcd metadata and object-storage data |
| Backup | `volume-snapshot` | `VolumeSnapshotter` | CSI snapshot of the **Standalone** data volume, taken inside the same write fence as a physical backup |
| Backup | `manifest-backup` | `Restic` | The Kubernetes objects of the database (for example the auth Secret) |
| Restore | `logical-backup-restore` | `Milvus` | Restores collections |
| Restore | `physical-backup-restore` | `Restic` | Restores a whole Milvus |
| Restore | `manifest-restore` | `Restic` | Restores the Kubernetes objects |

## Before You Begin

- A Kubernetes cluster with `kubectl` configured, KubeDB installed with the `Milvus` and `Etcd` feature gates on (see [Prepare Dependencies](/docs/guides/milvus/quickstart/prerequisites.md)), and KubeStash installed.
- A `demo` namespace:

  ```bash
  $ kubectl create namespace demo
  namespace/demo created
  ```

- A running Milvus. The examples below use a Standalone Milvus named `milvus-pitr` and the storage secret `milvus-storage` from [Prepare Dependencies](/docs/guides/milvus/quickstart/prerequisites.md).
- A KubeStash `BackupStorage`, a `RetentionPolicy` and an encryption `Secret` holding `RESTIC_PASSWORD`. Create them as described in [Prepare the Backend](/docs/guides/etcd/backup/kubestash/snapshot/index.md#prepare-the-backend) of the Etcd backup guide. The examples here call them `s3-storage`, `milvus-retention` and `encrypt-secret`. **Use a bucket and credentials that are separate from the Milvus data bucket**, so a delete on the Milvus side cannot reach the backups.

## Physical backups with a consistent recovery point

A physical backup does not read through the Milvus API. It briefly **fences writes**, captures the etcd metadata and the object-storage objects inside that fence, and releases the fence. The moment the metadata was captured is stored on the `Snapshot` as its **fence time**. That is the recovery point the snapshot restores to, and it is the same point for the etcd metadata and for the object storage, which is what makes the backup consistent across both stores.

This is a crash-consistent capture taken under a bounded write freeze. It is not an online copy: writes are denied for the duration of the fence (up to `fenceTimeout`, 60 seconds by default) and clients should expect failed writes during it.

### Quiesce modes and their parameters

The `physical-backup` task (and `volume-snapshot`) take these parameters:

| Parameter | Default | Meaning |
|---|---|---|
| `quiesce` | `DenyWrites` | How writes are stopped: `DenyWrites`, `Flush` or `None` |
| `fenceSettle` | `10s` | Time to let Milvus pick up the write fence. Milvus polls its etcd configuration every 5 seconds |
| `fenceTimeout` | `60s` | Maximum time writes may stay denied |
| `flushTimeout` | `10m` | Timeout of the `FlushAll` call (`physical-backup` only) |
| `allowQuiesceDowngrade` | `false` | Fall back to `Flush` when `quotaAndLimits.enabled` is false |
| `parallelism` | `16` | Parallel object copies (`physical-backup` only) |

- `DenyWrites` fences writes by setting Milvus's dynamic `forceDeny` setting. That setting only takes effect when Milvus's QuotaCenter is running, which means `quotaAndLimits.enabled` must be `true`. KubeDB renders a complete `milvus.yaml` with `quotaAndLimits.enabled: true`, so this holds by default. **If you override that value to `false` through `spec.configuration`, `DenyWrites` is a no-op.** The `allowQuiesceDowngrade` parameter exists for that case: when it is `"true"` the backup falls back to `Flush`. It is `false` by default, so the backup does not downgrade by itself.
- `Flush` flushes in-memory data (`FlushAll`, bounded by `flushTimeout`) without denying writes, so it does not give the write fence that `DenyWrites` does.
- `None` takes the capture without any quiesce step. Use it only if you accept a copy that is not write-fenced.

### A manual physical backup

Create a `BackupConfiguration` that targets the Milvus and uses the `physical-backup` task:

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: milvus-physical-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Milvus
    namespace: demo
    name: milvus-pitr
  backends:
    - name: s3-backend
      storageRef:
        namespace: demo
        name: s3-storage
      retentionPolicy:
        name: milvus-retention
        namespace: demo
  sessions:
    - name: physical
      scheduler:
        schedule: "0 */6 * * *"
        jobTemplate:
          backoffLimit: 1
      repositories:
        - name: milvus-physical-repo
          backend: s3-backend
          directory: /milvus
          encryptionSecret:
            name: encrypt-secret
            namespace: demo
      addon:
        name: milvus-addon
        tasks:
          - name: physical-backup
            params:
              quiesce: DenyWrites
              fenceSettle: 10s
              fenceTimeout: 60s
```

## Continuous archiving and point-in-time recovery: `MilvusArchiver`

For production, let the operator generate the backup configuration. A `MilvusArchiver` describes the policy once and applies it to every Milvus that opts in. It creates:

- a scheduled **full backup** (`fullBackup`), taken as a physical backup,
- a **change-log archiver** (`logBackup`) that continuously records every etcd change and every new object version, so the database can be restored to any point inside the recorded window,
- a scheduled **manifest backup** (`manifestBackup`) of the database's Kubernetes objects.

```yaml
apiVersion: archiver.kubedb.com/v1alpha1
kind: MilvusArchiver
metadata:
  name: milvus-archiver
  namespace: demo
spec:
  pause: false
  databases:
    namespaces:
      from: Selector
      selector: {matchLabels: {kubernetes.io/metadata.name: demo}}
    selector:
      matchLabels: {archiver: restic}
  retentionPolicy: {name: milvus-retention, namespace: demo}
  encryptionSecret: {name: encrypt-secret, namespace: demo}
  backupStorage:
    ref: {name: s3-storage, namespace: demo}
    subDir: milvus
  deletionPolicy: WipeOut
  fullBackup:
    driver: Restic
    task:
      params:
        quiesce: DenyWrites
        fenceSettle: 10s
        fenceTimeout: 60s
    scheduler:
      schedule: "0 */6 * * *"
      successfulJobsHistoryLimit: 2
      failedJobsHistoryLimit: 2
    sessionHistoryLimit: 3
  logBackup:
    retentionPeriod: 2d
    retentionSchedule: "0 */6 * * *"
    successfulLogHistoryLimit: 5
    failedLogHistoryLimit: 5
  manifestBackup:
    scheduler:
      schedule: "0 */6 * * *"
      successfulJobsHistoryLimit: 1
      failedJobsHistoryLimit: 1
    sessionHistoryLimit: 2
```

This follows the operator's `config/samples/milvus_archiver.yaml`, with production-style schedules in place of that sample's 30-minute test schedules. A few notes on the fields:

- `spec.fullBackup.driver` is `Restic` or `VolumeSnapshotter`. Metadata and objects are always archived by the backup plugin. `VolumeSnapshotter` also takes a CSI snapshot of the data volume inside the same fence, so a restored **Standalone** Milvus starts with a warm local cache. It is **Standalone-only**: with a Distributed Milvus the archiver reports the `ArchiverUnsupported` condition and you must use `Restic`.
- `spec.fullBackup.task.params` takes the same quiesce parameters as the manual `physical-backup` task above.
- `spec.logBackup.retentionPeriod` is the retention of the archived change log, written as a number and a unit in `d`, `w`, `m` or `y` (for example `2d`). `retentionSchedule` is a cron expression for the pruning job. Without a `logBackup` section only full backups are taken, and there is no point-in-time recovery.
- `spec.logBackup.configSecret` can tune the change-log archiver. It is a `GenericSecretReference` where both `name` and `envToSecretKey` are required. The sample sets these keys: `META_SEGMENT_INTERVAL`, `META_SEGMENT_MAX_EVENTS`, `TIME_INDEX_INTERVAL`, `OBJECT_SYNC_MODE`, `OBJECT_SYNC_INTERVAL`, `OBJECT_LAG_ALERT` and `SKEW_GRACE`. Leave it out to use the defaults. The sample sets `OBJECT_SYNC_MODE: notify` for MinIO; check the right mode for your object store before copying that value.
- `spec.encryptionSecret` holds `RESTIC_PASSWORD`. It encrypts the restic repositories and the archived change log.
- `spec.pause: true` deletes the change-log archiver and pauses the backup configuration.

### Opt a Milvus in

An archiver only serves a Milvus that it selects (`spec.databases`) **and** that references it. Label the Milvus so the selector matches, and reference the archiver from `spec.archiver.ref`:

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Milvus
metadata:
  name: milvus-pitr
  namespace: demo
  labels:
    archiver: restic
spec:
  version: "2.6.11"
  topology:
    mode: Standalone
  objectStorage:
    configSecret:
      name: milvus-storage
  archiver:
    ref:
      name: milvus-archiver
      namespace: demo
  storageType: Durable
  storage:
    accessModes: [ReadWriteOnce]
    resources:
      requests:
        storage: 5Gi
  deletionPolicy: WipeOut
```

The operator's repository sample `config/samples/milvus_woodpecker_archived.yaml` opts in by label alone. The API field comment describes `spec.archiver` as the reference that completes the double opt-in, so set both. The operator starts the archiver once the Milvus is `Ready`.

### Requirements and restrictions of the archiver

The admission webhook enforces these, and the archiver rejects a Milvus that breaks them:

- **Woodpecker write-ahead log only.** Every new Milvus uses the Woodpecker log stored in the object-storage bucket; there is no field to change it. A Standalone Milvus that already existed when this operator first reconciled it keeps its write-ahead log in RocksMQ on the data volume. It is marked with the annotation `kubedb.com/milvus-wal: rocksmq`, the webhook rejects `spec.archiver` and `spec.init.archiver` on it, and the condition `ArchiverUnsupported` is set. Migrate it with a **logical backup** restored into a new Milvus.
- **A KubeDB-managed etcd, dedicated to this Milvus.** `spec.metaStorage.externallyManaged: true` is rejected for a Milvus that uses the archiver, because Woodpecker keeps its log metadata at the etcd root and the etcd cannot be shared. The managed etcd that KubeDB creates for a Milvus is annotated `kubedb.com/archiver-skip: "true"`, so no separate `EtcdArchiver` attaches to it. Its contents are captured inside the Milvus fence instead.
- **Milvus 2.x only.** The webhook rejects archiver use on Milvus versions starting with `3.`, because the archiver is not validated for Milvus 3.x yet.
- **`VolumeSnapshotter` is Standalone-only**, as above.

## Restore

### Restore a whole Milvus, at a point in time

Create a **new** `Milvus` with `spec.init.archiver`. The operator withholds the PetSets until the metadata and the objects are restored, then verifies them and starts Milvus.

```yaml
apiVersion: kubedb.com/v1alpha2
kind: Milvus
metadata:
  name: milvus-restored
  namespace: demo
spec:
  version: "2.6.11"
  topology:
    mode: Standalone
  objectStorage:
    configSecret:
      name: milvus-restore-storage
  init:
    archiver:
      recoveryTimestamp: "2026-09-30T16:22:29Z"
      encryptionSecret: {name: encrypt-secret, namespace: demo}
      fullDBRepository: {name: milvus-pitr-full, namespace: demo}
  storageType: Durable
  storage:
    accessModes: [ReadWriteOnce]
    resources:
      requests:
        storage: 5Gi
  deletionPolicy: WipeOut
```

This is the operator's `config/samples/milvus_restore_from_archive.yaml`. Notes:

- `recoveryTimestamp`, `fullDBRepository` and `encryptionSecret` are all **required** by the webhook.
- `fullDBRepository` names the KubeStash `Repository` that holds the full backups of the source Milvus. The operator names it `<milvus-name>-full` (the manifest repository is `<milvus-name>-manifest`), so the repository for the `milvus-pitr` example is `milvus-pitr-full`.
- The operator picks the newest successful full snapshot whose **fence time** is at or before `recoveryTimestamp`, then replays the archived change log forward to `recoveryTimestamp`. If no snapshot qualifies, the restore fails instead of picking a later one.
- The sample restores into an **empty bucket with the same `rootPath` as the source**, which is why `milvus-restore-storage` is a different secret from the source's.
- `spec.init.archiver` cannot be changed once the database is initialized.
- The restore needs the same restrictions as the archiver: Woodpecker log, a managed etcd, Milvus 2.x.
- A physical restore verifies the result according to the `verify` parameter of the `physical-backup-restore` task: `sample` (the default), `full` or `none`. `parallelism` (default `16`) sets the number of parallel object copies.
- A Standalone restore taken with the `VolumeSnapshotter` driver uses the CSI snapshot as the data volume's data source for a warm cache and sets the condition `WarmCacheRestored`. If the snapshot is unusable, the Milvus starts with a cold cache.

### Restore collections from a logical backup

A logical backup is taken with a `BackupConfiguration` using the `logical-backup` task. `collections` is a comma-separated list; leave it empty to back up all collections.

```yaml
      addon:
        name: milvus-addon
        tasks:
          - name: logical-backup
            params:
              collections: "demo_collection"
```

Restore it into a running Milvus with a `RestoreSession`:

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: restore-milvus-collections
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Milvus
    namespace: demo
    name: milvus-target
  dataSource:
    repository: milvus-logical-repo
    snapshot: latest
    encryptionSecret:
      name: encrypt-secret
      namespace: demo
  addon:
    name: milvus-addon
    tasks:
      - name: logical-backup-restore
        params:
          restoreSuffix: "_restored"
```

`restoreSuffix` (default `_restored`) is appended to each restored collection name, so a restore into a cluster that still has the original collections does not collide with them.

### Rebuild the etcd

The Milvus-managed etcd is not backed up separately, because its contents are inside the Milvus backup. If you run a standalone KubeDB `Etcd` (for example as an external metadata store), back it up with an `EtcdArchiver` and restore it with a bootstrap-time `spec.init.archiver` or an in-place `EtcdOpsRequest` of type `Restore`. See [Backup & Restore Etcd](/docs/guides/etcd/backup/kubestash/overview/index.md) and [spec.restore](/docs/guides/etcd/concepts/etcdopsrequest.md#specrestore). etcd backup is snapshot-only, with no point-in-time recovery.

## What is not confirmed

These points are open and the guide does not claim them:

- **Database-level restore.** The logical restore function takes a list of collections (`--collections` on backup, collection names restored with `--restore-suffix`). There is no database parameter, so restoring a whole Milvus database as a unit is not offered by the addon today.
- **Restore into a different cluster with re-bound etcd and object-storage targets.** A restore creates a new `Milvus`, so a new namespace is straightforward. Pointing the restored Milvus at a different etcd or bucket without hand edits has not been tested, and no documentation or test for it exists yet.
- **The consistency guarantee under load.** The fence-time recovery point is how the operator describes the guarantee. It has not been measured on a live cluster, including the write impact during the fence.
- **The plugin release.** `milvus-backup-plugin` is a release candidate (`v0.1.0-rc.0`). Pin the image tag you test and check for a newer one before production use.

## Related

- [Prepare Milvus Dependencies](/docs/guides/milvus/quickstart/prerequisites.md)
- [Milvus CRD](/docs/guides/milvus/concepts/milvus.md)
- [Backup & Restore Etcd | KubeStash](/docs/guides/etcd/backup/kubestash/snapshot/index.md)
