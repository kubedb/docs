---
title: Elasticsearch Backup Customization | KubeStash
description: Customizing Elasticsearch Backup and Restore process with KubeStash
menu:
  docs_{{ .version }}:
    identifier: guides-es-backup-customization-stashv2
    name: Customizing Backup & Restore Process
    parent: guides-es-backup-stashv2
    weight: 50
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Customizing Backup and Restore Process

KubeStash provides rich customization supports for the backup and restore process to meet the requirements of various cluster configurations. This guide will show you some examples of these customizations.

## Customizing Backup Process

In this section, we are going to show you how to customize the backup process. Here, we are going to show some examples of providing arguments to the backup process, running the backup process as a specific user, etc.

### Passing arguments to the backup process

KubeStash Elasticsearch addon uses the `multielasticdump` command by default for backups. You can pass arguments to the `multielasticdump` through `args` param under `task.params` section.

The below example shows how you can pass the `--ignoreType` argument to ignore `template` and `settings` during backup.


```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: es-quickstart-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Elasticsearch
    namespace: demo
    name: es-quickstart
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
        - name: s3-elasticsearch-repo
          backend: s3-backend
          directory: /es
          encryptionSecret:
            name: encrypt-secret
            namespace: demo
      addon:
        name: elasticsearch-addon
        tasks:
          - name: logical-backup
            params:
              args: --ignoreType=template,settings
```


### Ignoring SearchGuard Indexes

If you are using the Search Guard variant for your Elasticsearch, you can pass a regex through the `--match` argument to ignore the Search Guard specific indexes during backup.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: es-quickstart-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Elasticsearch
    namespace: demo
    name: es-quickstart
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
        - name: s3-elasticsearch-repo
          backend: s3-backend
          directory: /es
          encryptionSecret:
            name: encrypt-secret
            namespace: demo
      addon:
        name: elasticsearch-addon
        tasks:
          - name: logical-backup
            params:
              args: --match=^(?![.])(?!searchguard).+ --ignoreType=template
```

> **WARNING**: Make sure that your provided database has been created before taking backup.

### Using multiple backends

You can configure multiple backends within a single `backupConfiguration`. To back up the same data to different backends, such as S3 and GCS, declare each backend in the `.spe.backends` section. Then, reference these backends in the `.spec.sessions[*].repositories` section.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: es-quickstart-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Elasticsearch
    namespace: demo
    name: es-quickstart
  backends:
    - name: s3-backend
      storageRef:
        namespace: demo
        name: s3-storage
      retentionPolicy:
        name: demo-retention
        namespace: demo
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
        - name: s3-elasticsearch-repo
          backend: s3-backend
          directory: /es
          encryptionSecret:
            name: encrypt-secret
            namespace: demo
        - name: gcs-elasticsearch-repo
          backend: gcs-backend
          directory: /es
          encryptionSecret:
            name: encrypt-secret
            namespace: demo
      addon:
        name: elasticsearch-addon
        tasks:
          - name: logical-backup
```

### Running backup job as a specific user

If your cluster requires running the backup job as a specific user, you can provide `securityContext` under `addon.jobTemplate.spec.securityContext` section. The below example shows how you can run the backup job as the `root` user.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: es-quickstart-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Elasticsearch
    namespace: demo
    name: es-quickstart
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
        - name: s3-elasticsearch-repo
          backend: s3-backend
          directory: /es
          encryptionSecret:
            name: encrypt-secret
            namespace: demo
      addon:
        name: elasticsearch-addon
        tasks:
        - name: logical-backup
        jobTemplate:
          spec:
            securityContext:
              runAsUser: 1000
              runAsGroup: 1000
```

### Specifying Memory/CPU limit/request for the backup job

If you want to specify the Memory/CPU limit/request for your backup job, you can specify `resources` field under `addon.jobTemplate.spec` section.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: es-quickstart-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Elasticsearch
    namespace: demo
    name: es-quickstart
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
        - name: s3-elasticsearch-repo
          backend: s3-backend
          directory: /es
          encryptionSecret:
            name: encrypt-secret
            namespace: demo
      addon:
        name: elasticsearch-addon
        tasks:
          - name: logical-backup
        jobTemplate:
          spec:
            resources:
              requests:
                cpu: "200m"
                memory: "1Gi"
              limits:
                cpu: "200m"
                memory: "1Gi"
```

> You can configure additional runtime settings for backup jobs within the `addon.jobTemplate.spec` sections. For further details, please refer to the [reference](https://kubestash.com/docs/latest/concepts/crds/backupconfiguration/#podtemplate-spec).

## Customizing Restore Process

`KubeStash` uses `multielastic` during the restore process. In this section, we are going to show how you can pass arguments to the restore process, restore a specific snapshot, run restore job as a specific user, etc.

### Passing arguments to the restore process

You can pass any supported `psql` arguments to the restore process using the `args` field within the `addon.tasks[*].params` section. 

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: es-quickstart-restore
  namespace: demo
spec:
  target:
    name: es-quickstart
    namespace: demo
    apiGroup: kubedb.com
    kind: Elasticsearch
  dataSource:
    snapshot: latest
    repository: s3-elasticsearch-repo
    encryptionSecret:
      name: encrypt-secret
      namespace: demo
  addon:
    name: elasticsearch-addon
    tasks:
    - name: logical-backup-restore
      params:
        args: --ignoreType=template,settings
```

### Restore specific snapshot

You can also restore a specific snapshot. At first, list the available snapshot as bellow,

```bash
$ kubectl get snapshots.storage.kubestash.com -n demo -l=kubestash.com/repo-name=s3-elasticsearch-repo
NAME                                                              REPOSITORY              SESSION           SNAPSHOT-TIME          DELETION-POLICY   PHASE       AGE
s3-elasticsearch-repo-es-quickstckup-frequent-backup-1726655113   s3-elasticsearch-repo   frequent-backup   2024-09-18T10:25:23Z   Delete            Succeeded   147m
```

The below example shows how you can pass a specific snapshot name in `.spec.dataSource` section.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: es-quickstart-restore
  namespace: demo
spec:
  target:
    name: es-quickstart
    namespace: demo
    apiGroup: kubedb.com
    kind: Elasticsearch
  dataSource:
    snapshot: s3-elasticsearch-repo-es-quickstckup-frequent-backup-1726655113
    repository: s3-elasticsearch-repo
    encryptionSecret:
      name: encrypt-secret
      namespace: demo
  addon:
    name: elasticsearch-addon
    tasks:
      - name: logical-backup-restore
```

### Running restore job as a specific user

Similar to the backup process under the `addon.jobTemplate.spec.` you can provide `securityContext` to run the restore job as a specific user.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: es-quickstart-restore
  namespace: demo
spec:
  target:
    name: es-quickstart
    namespace: demo
    apiGroup: kubedb.com
    kind: Elasticsearch
  dataSource:
    snapshot: latest
    repository: s3-elasticsearch-repo
    encryptionSecret:
      name: encrypt-secret
      namespace: demo
  addon:
    name: elasticsearch-addon
    tasks:
      - name: logical-backup-restore
    jobTemplate:
      spec:
        securityContext:
          runAsUser: 1000
          runAsGroup: 1000
```

### Specifying Memory/CPU limit/request for the restore job

Similar to the backup process, you can also provide `resources` field under the `addon.jobTemplate.spec.resources` section to limit the Memory/CPU for your restore job.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: es-quickstart-restore
  namespace: demo
spec:
  target:
    name: es-quickstart
    namespace: dev
    apiGroup: kubedb.com
    kind: Elasticsearch
  dataSource:
    snapshot: latest
    repository: s3-elasticsearch-repo
    encryptionSecret:
      name: encrypt-secret
      namespace: demo
  addon:
    name: elasticsearch-addon
    tasks:
      - name: logical-backup-restore
    jobTemplate:
      spec:
        resources:
          requests:
            cpu: "200m"
            memory: "1Gi"
          limits:
            cpu: "200m"
            memory: "1Gi"
```

> You can configure additional runtime settings for restore jobs within the `addon.jobTemplate.spec` sections. For further details, please refer to the [reference](https://kubestash.com/docs/latest/concepts/crds/restoresession/#podtemplate-spec).