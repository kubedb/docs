---
title: PostgreSQL Backup Customization | KubeStash
description: Customizing PostgreSQL Backup and Restore process with KubeStash
menu:
  docs_{{ .version }}:
    identifier: guides-pg-backup-customization-stashv2
    name: Customizing Backup & Restore Process
    parent: guides-pg-backup-stashv2
    weight: 50
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Customizing Backup and Restore Process

KubeStash provides rich customization supports for the backup and restore process to meet the requirements of various cluster configurations. This guide will show you some examples of these customizations.

## Customizing Backup Process

In this section, we are going to show you how to customize the backup process. Here, we are going to show some examples of providing arguments to the backup process, running the backup process as a specific user, etc.

### Passing arguments to the backup process

KubeStash PostgreSQL addon uses the [pg_dumpall](https://www.postgresql.org/docs/current/app-pg-dumpall.html) command by default for backups. However, you can change the dump command to [pg_dump](https://www.postgresql.org/docs/current/app-pgdump.html) by setting the `backupCmd` parameter under the `addon.tasks[*].params` section. You can pass supported options for either `pg_dumpall` or `pg_dump` through the `args` parameter in the same section.

The below example shows how you can pass the `--clean` to include SQL commands to clean (drop) databases before recreating them.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: sample-postgres-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Postgres
    namespace: demo
    name: sample-postgres
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
        - name: gcs-postgres-repo
          backend: gcs-backend
          directory: /postgres
          encryptionSecret:
            name: encrypt-secret
            namespace: demo
      addon:
        name: postgres-addon
        tasks:
          - name: logical-backup
            params:
              args: --clean
```


### Passing a target database to the backup process

KubeStash PostgreSQL addon uses the [pg_dumpall](https://www.postgresql.org/docs/current/app-pg-dumpall.html) command by default for backups. If you want to back up a single database, you’ll need to switch the command to [pg_dump](https://www.postgresql.org/docs/current/app-pgdump.html). You can do this by setting `backupCmd` to `pg_dump` under the `addon.tasks[*].params` section and specifying the database name using the `args` parameter in the same section.

The below example shows how you can set `pg_dump` and pass target database name during backup. 

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: sample-postgres-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Postgres
    namespace: demo
    name: sample-postgres
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
        - name: gcs-postgres-repo
          backend: gcs-backend
          directory: /postgres
          encryptionSecret:
            name: encrypt-secret
            namespace: demo
      addon:
        name: postgres-addon
        tasks:
          - name: logical-backup
            params:
              backupCmd: pg_dump
              args: testdb
```

> **WARNING**: Make sure that your provided database has been created before taking backup.

### Using multiple backends

You can configure multiple backends within a single `backupConfiguration`. To back up the same data to different backends, such as S3 and GCS, declare each backend in the `.spe.backends` section. Then, reference these backends in the `.spec.sessions[*].repositories` section.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: sample-postgres-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Postgres
    namespace: demo
    name: sample-postgres
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
        - name: gcs-postgres-repo
          backend: gcs-backend
          directory: /postgres
          encryptionSecret:
            name: encrypt-secret
            namespace: demo
        - name: s3-postgres-repo
          backend: s3-backend
          directory: /postgres-copy
          encryptionSecret:
            name: encrypt-secret
            namespace: demo
      addon:
        name: postgres-addon
        tasks:
          - name: logical-backup
```

### Running backup job as a specific user

If your cluster requires running the backup job as a specific user, you can provide `securityContext` under `addon.jobTemplate.spec.securityContext` section. The below example shows how you can run the backup job as the `root` user.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: sample-postgres-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Postgres
    namespace: demo
    name: sample-postgres
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
        - name: gcs-postgres-repo
          backend: gcs-backend
          directory: /postgres
          encryptionSecret:
            name: encrypt-secret
            namespace: demo
      addon:
        name: postgres-addon
        jobTemplate:
          spec:
            securityContext:
              runAsUser: 0
              runAsGroup: 0
        tasks:
          - name: logical-backup
```

### Specifying Memory/CPU limit/request for the backup job

If you want to specify the Memory/CPU limit/request for your backup job, you can specify `resources` field under `addon.jobTemplate.spec` section.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: sample-postgres-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Postgres
    namespace: demo
    name: sample-postgres
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
        - name: gcs-postgres-repo
          backend: gcs-backend
          directory: /postgres
          encryptionSecret:
            name: encrypt-secret
            namespace: demo
      addon:
        name: postgres-addon
        jobTemplate:
          spec:
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

> You can configure additional runtime settings for backup jobs within the `addon.jobTemplate.spec` sections. For further details, please refer to the [reference](https://kubestash.com/docs/latest/concepts/crds/backupconfiguration/#podtemplate-spec).

## Customizing Restore Process

`KubeStash` uses [psql](https://www.postgresql.org/docs/current/app-psql.html) during the restore process. In this section, we are going to show how you can pass arguments to the restore process, restore a specific snapshot, run restore job as a specific user, etc.

### Passing arguments to the restore process

You can pass any supported `psql` arguments to the restore process using the `args` field within the `addon.tasks[*].params` section. This example demonstrates how to specify a database `testdb` to connect to during the restore process.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: sample-postgres-restore
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Postgres
    namespace: demo
    name: restored-postgres
  dataSource:
    repository: gcs-postgres-repo
    snapshot: latest
    encryptionSecret:
      name: encrypt-secret
      namespace: demo
  addon:
    name: postgres-addon
    tasks:
      - name: logical-backup-restore
        params:
          args: --dbname=testdb
```

### Restore specific snapshot

You can also restore a specific snapshot. At first, list the available snapshot as bellow,

```bash
$ kubectl get snapshots.storage.kubestash.com -n demo -l=kubestash.com/repo-name=gcs-postgres-repo
NAME                                                                    REPOSITORY          SESSION           SNAPSHOT-TIME          DELETION-POLICY   PHASE       AGE
gcs-postgres-repo-sample-postgres-backup-frequent-backup-1725257849   gcs-postgres-repo   frequent-backup   2024-09-02T06:18:01Z      Delete            Succeeded   15m
gcs-postgres-repo-sample-postgres-backup-frequent-backup-1725258000   gcs-postgres-repo   frequent-backup   2024-09-02T06:20:00Z      Delete            Succeeded   13m
gcs-postgres-repo-sample-postgres-backup-frequent-backup-1725258300   gcs-postgres-repo   frequent-backup   2024-09-02T06:25:00Z      Delete            Succeeded   8m34s
gcs-postgres-repo-sample-postgres-backup-frequent-backup-1725258600   gcs-postgres-repo   frequent-backup   2024-09-02T06:30:00Z      Delete            Succeeded   3m34s
```

The below example shows how you can pass a specific snapshot name in `.spec.dataSource` section.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: sample-postgres-restore
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Postgres
    namespace: demo
    name: restored-postgres
  dataSource:
    repository: gcs-postgres-repo
    snapshot: gcs-postgres-repo-sample-postgres-backup-frequent-backup-1725258000
    encryptionSecret:
      name: encrypt-secret
      namespace: demo
  addon:
    name: postgres-addon
    tasks:
      - name: logical-backup-restore
```

### Running restore job as a specific user

Similar to the backup process under the `addon.jobTemplate.spec.` you can provide `securityContext` to run the restore job as a specific user.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: sample-postgres-restore
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Postgres
    namespace: demo
    name: restored-postgres
  dataSource:
    repository: gcs-postgres-repo
    snapshot: latest
    encryptionSecret:
      name: encrypt-secret
      namespace: demo
  addon:
    name: postgres-addon
    jobTemplate:
      spec:
        securityContext:
          runAsUser: 0
          runAsGroup: 0
    tasks:
      - name: logical-backup-restore
```

### Specifying Memory/CPU limit/request for the restore job

Similar to the backup process, you can also provide `resources` field under the `addon.jobTemplate.spec.resources` section to limit the Memory/CPU for your restore job.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: sample-postgres-restore
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Postgres
    namespace: demo
    name: restored-postgres
  dataSource:
    repository: gcs-postgres-repo
    snapshot: latest
    encryptionSecret:
      name: encrypt-secret
      namespace: demo
  addon:
    name: postgres-addon
    jobTemplate:
      spec:
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