---
title: ZooKeeper Backup Customization | KubeStash
description: Customizing ZooKeeper Backup and Restore process with KubeStash
menu:
  docs_{{ .version }}:
    identifier: guides-zk-backup-customization-stashv2
    name: Customizing Backup & Restore Process
    parent: guides-zk-backup-stashv2
    weight: 50
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Customizing Backup and Restore Process

KubeStash provides rich customization supports for the backup and restore process to meet the requirements of various cluster configurations. This guide will show you some examples of these customizations.

## Customizing Backup Process

In this section, we are going to show you how to customize the backup process. Here, we are going to show some examples of providing arguments to the backup process, running the backup process as a specific user, etc.

### Using multiple backends

You can configure multiple backends within a single `backupConfiguration`. To back up the same data to different backends, such as S3 and GCS, declare each backend in the `.spec.backends` section. Then, reference these backends in the `.spec.sessions[*].repositories` section.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: sample-zookeeper-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: ZooKeeper
    namespace: demo
    name: sample-zookeeper
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
        - name: gcs-zookeeper-repo
          backend: gcs-backend
          directory: /zookeeper
          encryptionSecret:
            name: encrypt-secret
            namespace: demo
        - name: s3-zookeeper-repo
          backend: s3-backend
          directory: /zookeeper-copy
          encryptionSecret:
            name: encrypt-secret
            namespace: demo
      addon:
        name: zookeeper-addon
        tasks:
          - name: logical-backup
```

### Running backup job as a specific user

If your cluster requires running the backup job as a specific user, you can provide `securityContext` under `addon.jobTemplate.spec.securityContext` section. The below example shows how you can run the backup job as the `root` user.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: sample-zookeeper-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: ZooKeeper
    namespace: demo
    name: sample-zookeeper
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
        - name: s3-zookeeper-repo
          backend: s3-backend
          directory: /zookeeper
          encryptionSecret:
            name: encrypt-secret
            namespace: demo
      addon:
        name: zookeeper-addon
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
  name: sample-zookeeper-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: ZooKeeper
    namespace: demo
    name: sample-zookeeper
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
        - name: s3-zookeeper-repo
          backend: s3-backend
          directory: /zookeeper
          encryptionSecret:
            name: encrypt-secret
            namespace: demo
      addon:
        name: zookeeper-addon
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

### Running restore job as a specific user

Similar to the backup process under the `addon.jobTemplate.spec.` you can provide `securityContext` to run the restore job as a specific user.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: sample-zookeeper-restore
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: ZooKeeper
    namespace: demo
    name: restored-zookeeper
  dataSource:
    repository: s3-zookeeper-repo
    snapshot: latest
    encryptionSecret:
      name: encrypt-secret
      namespace: demo
  addon:
    name: zookeeper-addon
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
  name: sample-zookeeper-restore
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: ZooKeeper
    namespace: demo
    name: restored-zookeeper
  dataSource:
    repository: s3-zookeeper-repo
    snapshot: latest
    encryptionSecret:
      name: encrypt-secret
      namespace: demo
  addon:
    name: zookeeper-addon
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