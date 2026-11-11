---
title: Logical Backup and Restore of Milvus | KubeStash
menu:
  docs_{{ .version }}:
    identifier: milvus-backup-logical-guide
    name: Guide
    parent: milvus-backup-logical
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Logical Backup and Restore of Milvus with KubeStash

A logical backup exports collections through Milvus itself with `milvus-backup`. It works for Standalone and Distributed Milvus, with a KubeDB-managed or an externally managed meta etcd, because it never reads etcd directly. It can back up and restore a subset of collections, and restore either alongside the existing collections or in place over them.

## Before You Begin

- Install [KubeStash](https://kubestash.com/docs/latest/setup/) and the KubeDB KubeStash catalog, which provides the `milvus-addon`.
- Create a `BackupStorage` and a `RetentionPolicy` in the `demo` namespace, as in the [KubeStash quickstart](https://kubestash.com/docs/latest/concepts/crds/backupstorage/).
- Have a Ready Milvus named `milvus` in `demo`.

> Note: The yaml files used in this tutorial are stored in [docs/guides/milvus/backup/logical/yamls](https://github.com/kubedb/docs/tree/{{< param "info.version" >}}/docs/guides/milvus/backup/logical/yamls) folder in GitHub repository [kubedb/docs](https://github.com/kubedb/docs).

## Parameters

Set these under `addon.tasks[].params`. All are strings.

| Task | Parameter | Default | Meaning |
|---|---|---|---|
| `logical-backup` | `collections` | `""` | Comma-separated `<db>.<collection>`. Empty backs up every collection |
| `logical-backup` | `backupTimeout` | `0s` | Hard limit on the backup, as a Go duration. `0s` means no limit |
| `logical-backup-restore` | `restoreSuffix` | `_restored` | Appended to every restored collection name |
| `logical-backup-restore` | `collections` | `""` | Comma-separated `<db>.<collection>` to restore. Empty restores every collection in the snapshot |
| `logical-backup-restore` | `inPlaceRestore` | `false` | Restore over the existing collections, replacing them. Takes precedence over the default `restoreSuffix` |
| `logical-backup-restore` | `dropExistCollection` | `false` | Drop a colliding collection in the target instead of failing |
| `logical-backup-restore` | `restoreTimeout` | `0s` | Hard limit on the restore, as a Go duration |

Users and roles are captured in every backup but are not restored. A restored user would carry the source password hash, which could lock the operator out of the target.

## Back up

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/backup/logical/yamls/backupconfiguration.yaml
$ kubectl get backupsession -n demo -w
$ kubectl get snapshot -n demo -l kubestash.com/repo-name=milvus-logical-repo
```

For a large database set `backupTimeout` well above the time a backup takes, rather than leaving it unlimited, so a hung backup fails and the next schedule runs.

## Restore

A restore writes into the Milvus named in `spec.target`. Choose one of three forms.

### Alongside, into another Milvus (recommended for verification)

Restore the latest snapshot into a separate Milvus and compare it with the source before you trust the backup. Each collection is created as `<name>_restored`:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/backup/logical/yamls/restore-suffix.yaml
```

### Only some collections

Restore two collections next to the live ones in the source database:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/backup/logical/yamls/restore-collections.yaml
```

The granularity is the collection. A partition cannot be restored on its own.

### In place

`inPlaceRestore: "true"` drops each restored collection in the target and recreates it from the backup. Before dropping anything, the plugin checks that the backup holds a replacement for every collection it is about to drop. Always combine it with `collections`, so only what you mean to replace is touched:

```bash
$ kubectl apply -f https://github.com/kubedb/docs/raw/{{< param "info.version" >}}/docs/guides/milvus/backup/logical/yamls/restore-in-place.yaml
```

Without `inPlaceRestore`, a restore always uses a suffix. Running the same restore twice fails on the second run because `<name>_restored` already exists; nothing is dropped unless you set `dropExistCollection` or `inPlaceRestore`.

## Check a restore

```bash
$ kubectl get restoresession -n demo
$ kubectl describe restoresession -n demo milvus-restore-check
```

Then compare row counts and run a few known queries against the restored collections.

## What a logical backup does not cover

- **Point-in-time recovery and physical backups** come from a `MilvusArchiver`, which needs a KubeDB-managed meta etcd dedicated to the Milvus. A Milvus with `spec.metaStorage.externallyManaged: true` can use logical backups only.
- **Partition-level restore** is not available.
