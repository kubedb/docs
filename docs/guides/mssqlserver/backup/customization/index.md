---
title: MSSQLServer Backup Customization | KubeStash
description: Customizing MSSQLServer Backup and Restore process with KubeStash
menu:
  docs_{{ .version }}:
    identifier: guides-msssqlserver-backup-customization
    name: Customizing Backup & Restore Process
    parent: guides-mssqlserver-backup
    weight: 50
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Customizing Backup and Restore Process

KubeStash provides rich customization supports for the backup and restore process to meet the requirements of various cluster configurations. This guide will show you some examples of these customizations.

### Passing target databases to the backup process

KubeStash MSSQLServer addon uses the [wal-g](https://wal-g.readthedocs.io/) for backup. Addon has implemented a `databases` params which indicates your targeted backup databases.

The below example shows how you can pass the `--databases` option during backup.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: sample-mssqlserver-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: MSSQLServer
    namespace: demo
    name: sample-mssqlserver
  backends:
    - name: gcs-backend
      storageRef:
        namespace: demo
        name: gcs-storage
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
        - name: gcs-mssqlserver-repo
          backend: gcs-backend
          directory: /mssqlserver
      addon:
        name: mssqlserver-addon
        jobTemplate:
          spec:
            securityContext:
              runAsUser: 0
        tasks:
          - name: logical-backup
            params:
              databases: agdb1,agdb2
```

> **WARNING**: Make sure that your provides databases has been created before taking backup.

Here,
- `addon.tasks[*].databases` options indicates targeted databases. By default `mssqlserver-addon` takes all non-system databases. If you want to backup all databases keep the `databases` params empty.  

### Using multiple backends

You can configure multiple backends within a single `backupConfiguration`. To back up the same data to different backends, such as S3 and GCS, declare each backend in the `.spe.backends` section. Then, reference these backends in the `.spec.sessions[*].repositories` section.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: sample-mssqlserver-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: MSSQLServer
    namespace: demo
    name: sample-mssqlserver
  backends:
    - name: gcs-backend
      storageRef:
        namespace: demo
        name: gcs-storage
      retentionPolicy:
        name: demo-retention
        namespace: demo
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
        - name: gcs-mssqlserver-repo
          backend: gcs-backend
          directory: /mssqlserver
          encryptionSecret:
            name: encrypt-secret
            namespace: demo
        - name: s3-mssqlserver-repo
          backend: s3-backend
          directory: /mssqlserver-copy
          encryptionSecret:
            name: encrypt-secret
            namespace: demo
      addon:
        name: mssqlserver-addon
        jobTemplate:
          spec:
            securityContext:
              runAsUser: 0
        tasks:
          - name: logical-backup
```

### Specifying Memory/CPU limit/request for the backup job

If you want to specify the Memory/CPU limit/request for your backup job, you can specify `resources` field under `addon.jobTemplate.spec` section.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: sample-mssqlserver-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: MSSQLServer
    namespace: demo
    name: sample-mssqlserver
  backends:
    - name: gcs-backend
      storageRef:
        namespace: demo
        name: gcs-storage
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
        - name: gcs-mssqlserver-repo
          backend: gcs-backend
          directory: /mssqlserver
      addon:
        name: mssqlserver-addon
        jobTemplate:
          spec:
            securityContext:
              runAsUser: 0
            resources:
              requests:
                cpu: "200m"
                memory: "1Gi"
              limits:
                cpu: "200m"
                memory: "1Gi"
        tasks:
          - name: logical-backup
```

## Customizing Restore Process

KubeStash also uses `wal-g` during the restore process. In this section, we are going to show how you can pass arguments to the restore process, restore a specific snapshot, run restore job as a specific user, etc.

### Passing target databases to the restore process

KubeStash MSSQLServer addon uses the [wal-g](https://wal-g.readthedocs.io/) for backup. Addon has implemented a `databases` params which indicates your targeted restore databases.

The below example shows how you can pass the `--databases` option during restore.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: sample-mssqlserver-restore
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: MSSQLServer
    namespace: demo
    name: restored-mssqlserver
  dataSource:
    repository: gcs-mssqlserver-repo
    snapshot: latest
  addon:
    name: mssqlserver-addon
    jobTemplate:
      spec:
        securityContext:
          runAsUser: 0
    tasks:
      - name: logical-backup-restore
        params:
          databases: agdb1
```

> **WARNING**: Make sure that your provides databases has been backed up previously. You can check list of backup database in `.status.components[dump].walGStats.databases` section of `Snapshot` CR.

Here,
- `addon.tasks[*].databases` options indicates targeted databases. By default `mssqlserver-addon` restores all databases which are bakeked up previously. If you want to restore all databases keep the `databases` params empty.  


### Restore specific snapshot

You can also restore a specific snapshot. At first, list the available snapshot as bellow,

```bash
$ kubectl get snapshots.storage.kubestash.com -n demo -l=kubestash.com/repo-name=gcs-mssqlserver-repo
NAME                                                              REPOSITORY             SESSION           SNAPSHOT-TIME          DELETION-POLICY   PHASE       AGE
gcs-mssqlserver-repo-sample-mssqckup-frequent-backup-1727355681   gcs-mssqlserver-repo   frequent-backup   2024-09-26T13:01:22Z   Delete            Succeeded   5m8s
gcs-mssqlserver-repo-sample-mssqckup-frequent-backup-1727355730   gcs-mssqlserver-repo   frequent-backup   2024-09-26T13:02:10Z   Delete            Succeeded   4m20s
gcs-mssqlserver-repo-sample-mssqckup-frequent-backup-1727355900   gcs-mssqlserver-repo   frequent-backup   2024-09-26T13:05:00Z   Delete            Succeeded   90s
```

The below example shows how you can pass a specific snapshot name in `.dataSource` section.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: sample-mssqlserver-restore
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: MSSQLServer
    namespace: demo
    name: restored-mssqlserver
  dataSource:
    repository: gcs-mssqlserver-repo
    snapshot: gcs-mssqlserver-repo-sample-mssqckup-frequent-backup-1727355730
  addon:
    name: mssqlserver-addon
    jobTemplate:
      spec:
        securityContext:
          runAsUser: 0
    tasks:
      - name: logical-backup-restore
```

### Specifying Memory/CPU limit/request for the restore job

Similar to the backup process, you can also provide `resources` field under the `addon.jobTemplate.spec.resources` section to limit the Memory/CPU for your restore job.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: sample-mssqlserver-restore
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: MSSQLServer
    namespace: demo
    name: restored-mssqlserver
  dataSource:
    repository: gcs-mssqlserver-repo
    snapshot: latest
  addon:
    name: mssqlserver-addon
    jobTemplate:
      spec:
        securityContext:
          runAsUser: 0
        resources:
          requests:
            cpu: "200m"
            memory: "1Gi"
          limits:
            cpu: "200m"
            memory: "1Gi"
    tasks:
      - name: logical-backup-restore
```

> You can configure additional runtime settings for restore jobs within the `addon.jobTemplate.spec` sections. For further details, please refer to the [reference](https://kubestash.com/docs/latest/concepts/crds/restoresession/#podtemplate-spec).