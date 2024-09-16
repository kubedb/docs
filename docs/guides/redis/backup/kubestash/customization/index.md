---
title: Redis Backup Customization | KubeStash
description: Customizing Redis Backup and Restore process with KubeStash
menu:
  docs_{{ .version }}:
    identifier: guides-rd-backup-customization-stashv2
    name: Customizing Backup & Restore Process
    parent: guides-rd-backup-stashv2
    weight: 50
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Customizing Backup and Restore Process

KubeStash provides rich customization supports for the backup and restore process to meet the requirements of various cluster configurations. This guide will show you some examples of these customizations.

## Customizing Backup Process

In this section, we are going to show you how to customize the backup process. Here, we are going to show some examples of providing arguments to the backup process, running the backup process as a specific user, etc.

### Passing arguments to the backup process

KubeStash Redis addon uses the [rd_dumpall](https://www.redisql.org/docs/current/app-rd-dumpall.html) command by default for backups. However, you can change the dump command to [rd_dump](https://www.redisql.org/docs/current/app-rddump.html) by setting the `backupCmd` parameter under the `addon.tasks[*].params` section. You can pass supported options for either `rd_dumpall` or `rd_dump` through the `args` parameter in the same section.

The below example shows how you can pass the `--clean` to include SQL commands to clean (drop) databases before recreating them.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: sample-redis-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Redis
    namespace: demo
    name: sample-redis
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
        - name: gcs-redis-repo
          backend: gcs-backend
          directory: /redis
          encryptionSecret:
            name: encrypt-secret
            namespace: demo
      addon:
        name: redis-addon
        tasks:
          - name: logical-backup
            params:
              args: --clean
```


### Passing a target database to the backup process

KubeStash Redis addon uses the [rd_dumpall](https://www.redisql.org/docs/current/app-rd-dumpall.html) command by default for backups. For a single database backup, you need to rewrite the dump command. You can do this by setting `backupCmd` to [rd_dump](https://www.redisql.org/docs/current/app-rddump.html) under the `addon.tasks[*].params` section and specifying the database name using the `args` parameter in the same section.

The below example shows how you can set `rd_dump` and pass target database name during backup. 

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: sample-redis-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Redis
    namespace: demo
    name: sample-redis
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
        - name: gcs-redis-repo
          backend: gcs-backend
          directory: /redis
          encryptionSecret:
            name: encrypt-secret
            namespace: demo
      addon:
        name: redis-addon
        tasks:
          - name: logical-backup
            params:
              backupCmd: rd_dump
              args: testdb
```

> **WARNING**: Make sure that your provided database has been created before taking backup.

### Using multiple repositories

You can configure multiple repositories for the same backend. For example, if you want to back up the `/redis` directory using the `gcs-redis-repo` repository, you can also back up another directory, such as `/redis-copy`, by using a different repository, like `gcs-redis-repo-2`.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: sample-redis-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Redis
    namespace: demo
    name: sample-redis
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
        - name: gcs-redis-repo
          backend: gcs-backend
          directory: /redis
          encryptionSecret:
            name: encrypt-secret
            namespace: demo
        - name: gcs-redis-repo-2
          backend: gcs-backend
          directory: /redis-copy
          encryptionSecret:
            name: encrypt-secret
            namespace: demo
      addon:
        name: redis-addon
        tasks:
          - name: logical-backup
```

### Running backup job as a specific user

If your cluster requires running the backup job as a specific user, you can provide `securityContext` under `addon.jobTemplate.spec.securityContext` section. The below example shows how you can run the backup job as the `root` user.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: sample-redis-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Redis
    namespace: demo
    name: sample-redis
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
        - name: gcs-redis-repo
          backend: gcs-backend
          directory: /redis
          encryptionSecret:
            name: encrypt-secret
            namespace: demo
      addon:
        name: redis-addon
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
  name: sample-redis-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Redis
    namespace: demo
    name: sample-redis
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
        - name: gcs-redis-repo
          backend: gcs-backend
          directory: /redis
          encryptionSecret:
            name: encrypt-secret
            namespace: demo
      addon:
        name: redis-addon
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

`KubeStash` uses [psql](https://www.redisql.org/docs/current/app-psql.html) during the restore process. In this section, we are going to show how you can pass arguments to the restore process, restore a specific snapshot, run restore job as a specific user, etc.

### Passing arguments to the restore process

You can pass any supported `psql` arguments to the restore process using the `args` field within the `addon.tasks[*].params` section. This example demonstrates how to specify a database `testdb` to connect to during the restore process.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: sample-redis-restore
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Redis
    namespace: demo
    name: restored-redis
  dataSource:
    repository: gcs-redis-repo
    snapshot: latest
    encryptionSecret:
      name: encrypt-secret
      namespace: demo
  addon:
    name: redis-addon
    tasks:
      - name: logical-backup-restore
        params:
          args: --dbname=testdb
```

### Restore specific snapshot

You can also restore a specific snapshot. At first, list the available snapshot as bellow,

```bash
$ kubectl get snapshots.storage.kubestash.com -n demo -l=kubestash.com/repo-name=gcs-redis-repo
NAME                                                                    REPOSITORY          SESSION           SNAPSHOT-TIME          DELETION-POLICY   PHASE       AGE
gcs-redis-repo-sample-redis-backup-frequent-backup-1725257849   gcs-redis-repo   frequent-backup   2024-09-02T06:18:01Z      Delete            Succeeded   15m
gcs-redis-repo-sample-redis-backup-frequent-backup-1725258000   gcs-redis-repo   frequent-backup   2024-09-02T06:20:00Z      Delete            Succeeded   13m
gcs-redis-repo-sample-redis-backup-frequent-backup-1725258300   gcs-redis-repo   frequent-backup   2024-09-02T06:25:00Z      Delete            Succeeded   8m34s
gcs-redis-repo-sample-redis-backup-frequent-backup-1725258600   gcs-redis-repo   frequent-backup   2024-09-02T06:30:00Z      Delete            Succeeded   3m34s
```

The below example shows how you can pass a specific snapshot name in `.spec.dataSource` section.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: sample-redis-restore
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Redis
    namespace: demo
    name: restored-redis
  dataSource:
    repository: gcs-redis-repo
    snapshot: gcs-redis-repo-sample-redis-backup-frequent-backup-1725258000
    encryptionSecret:
      name: encrypt-secret
      namespace: demo
  addon:
    name: redis-addon
    tasks:
      - name: logical-backup-restore
```

### Running restore job as a specific user

Similar to the backup process under the `addon.jobTemplate.spec.` you can provide `securityContext` to run the restore job as a specific user.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: sample-redis-restore
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Redis
    namespace: demo
    name: restored-redis
  dataSource:
    repository: gcs-redis-repo
    snapshot: latest
    encryptionSecret:
      name: encrypt-secret
      namespace: demo
  addon:
    name: redis-addon
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
  name: sample-redis-restore
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Redis
    namespace: demo
    name: restored-redis
  dataSource:
    repository: gcs-redis-repo
    snapshot: latest
    encryptionSecret:
      name: encrypt-secret
      namespace: demo
  addon:
    name: redis-addon
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