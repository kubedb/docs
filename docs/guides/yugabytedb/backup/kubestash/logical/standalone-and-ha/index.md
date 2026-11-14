---
title: Backup & Restore YugabyteDB | KubeStash
description: Backup and Restore YugabyteDB using KubeStash
menu:
  docs_{{ .version }}:
    identifier: guides-yugabytedb-standalone-ha-logical-backup-stashv2
    name: YugabyteDB Cluster
    parent: guides-yugabytedb-logical-backup-stashv2
    weight: 20
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Backup and Restore YugabyteDB using KubeStash

> **Not verified.** The KubeStash addon `yugabytedb-addon` exists in the KubeDB installer charts (`kubedb-kubestash-catalog`), but the plugin image its functions point to (`ghcr.io/kubedb/yugabytedb-restic-plugin:v0.1.0_${DB_VERSION}`) does not exist yet, so a backup or restore job cannot run. Nothing on this page was applied. There is no example file under `docs/examples` for it on purpose. The manifests below only show the intended shape, following the other KubeDB databases.

## What the addon provides

The addon `yugabytedb-addon` has these tasks:

| Task | Purpose |
|------|---------|
| `logical-backup` | a logical backup of the YSQL databases with `ysql_dump` (parameters: `args`, `databases`, `excludeDatabases`, `includeRoles`, `jobs`) |
| `manifest-backup` | backup of the Kubernetes manifests of the database |
| `logical-backup-restore` | restore with `pg_restore` (parameters: `args`, `databases`, `includeRoles`, `jobs`) |
| `manifest-restore` | restore of the manifests |

The backup is logical (per database dump). A YCQL backup is not provided.

## BackupConfiguration (not verified)

Create the `BackupStorage`, `RetentionPolicy` and the storage secret as in any other KubeDB database guide, then:

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: yb-demo-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: YugabyteDB
    namespace: demo
    name: yb-demo
  backends:
    - name: s3-backend
      storageRef:
        namespace: demo
        name: s3-storage
      retentionPolicy:
        name: demo-retention
        namespace: demo
  sessions:
    - name: frequent-backup
      scheduler:
        schedule: "*/5 * * * *"
        jobTemplate:
          backoffLimit: 1
      repositories:
        - name: s3-yb-repo
          backend: s3-backend
          directory: /backup
      addon:
        name: yugabytedb-addon
        tasks:
          - name: logical-backup
```

## RestoreSession (not verified)

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: yb-demo-restore
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: YugabyteDB
    namespace: demo
    name: restored-yb
  dataSource:
    repository: s3-yb-repo
    snapshot: latest
  addon:
    name: yugabytedb-addon
    tasks:
      - name: logical-backup-restore
```

## Next Steps

- [YugabyteDB CRD](/docs/guides/yugabytedb/concepts/yugabytedb.md)
