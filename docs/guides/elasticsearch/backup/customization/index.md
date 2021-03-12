---
title: Elasticsearch Backup Customization | Stash
description: Customizing Elasticsearch Backup and Restore process with Stash
menu:
  docs_{{ .version }}:
    identifier: stash-elasticsearch-guides-{{ .subproject_version }}-customization
    name: Customizing Backup & Restore Process
    parent: stash-elasticsearch-guides-{{ .subproject_version }}
    weight: 50
product_name: stash
menu_name: docs_{{ .version }}
section_menu_id: stash-addons
---

# Customizing Backup and Restore Process

Stash provides rich customization supports for the backup and restore process to meet the requirements of various cluster configurations. This guide will show you some examples of these customizations.

## Customizing Backup Process

In this section, we are going to show you how to customize the backup process. Here, we are going to show some examples of providing arguments to the backup process, running the backup process as a specific user, ignoring some indexes during the backup process, etc.

### Passing arguments to the backup process

Stash Elasticsearch addon uses [multielasticdump](https://github.com/elasticsearch-dump/elasticsearch-dump#multielasticdump) for backup. You can pass arguments to the `multielasticdump` through `args` param under `task.params` section.

The below example shows how you can pass the `--ignoreType` argument to ignore `template` and `settings` during backup.

```yaml
apiVersion: stash.appscode.com/v1beta1
kind: BackupConfiguration
metadata:
  name: sample-elasticsearch-backup
  namespace: demo
spec:
  schedule: "*/2 * * * *"
  task:
#    name: elasticsearch-backup-{{< param "info.subproject_version" >}} # Uncomment if you are not using KubeDB to deploy your database
    params:
    - name: args
      value: --ignoreType=template,settings
  repository:
    name: gcs-repo
  target:
    ref:
      apiVersion: appcatalog.appscode.com/v1alpha1
      kind: AppBinding
      name: sample-elasticsearch
  interimVolumeTemplate:
    metadata:
      name: stash-tmp-backup-storage
    spec:
      accessModes: [ "ReadWriteOnce" ]
      storageClassName: "standard"
      resources:
        requests:
          storage: 1Gi
  retentionPolicy:
    name: keep-last-5
    keepLast: 5
    prune: true
```

### Ignoring Search Guard Indexes

If you are using the Search Guard variant for your Elasticsearch, you can pass a regex through the `--match` argument to ignore the Search Guard specific indexes during backup.

```yaml
apiVersion: stash.appscode.com/v1beta1
kind: BackupConfiguration
metadata:
  name: sample-elasticsearch-backup
  namespace: demo
spec:
  schedule: "*/2 * * * *"
  task:
#    name: elasticsearch-backup-{{< param "info.subproject_version" >}} # Uncomment if you are not using KubeDB to deploy your database
    params:
    - name: args
      value: --match=^(?![.])(?!searchguard).+ --ignoreType=template
  repository:
    name: gcs-repo
  target:
    ref:
      apiVersion: appcatalog.appscode.com/v1alpha1
      kind: AppBinding
      name: sample-elasticsearch
  interimVolumeTemplate:
    metadata:
      name: stash-tmp-backup-storage
    spec:
      accessModes: [ "ReadWriteOnce" ]
      storageClassName: "standard"
      resources:
        requests:
          storage: 1Gi
  retentionPolicy:
    name: keep-last-5
    keepLast: 5
    prune: true
```

### Running backup job as a specific user

If your cluster requires running the backup job as a specific user, you can provide `securityContext` under `runtimeSettings.pod` section. The below example shows how you can run the backup job as the root user.

```yaml
apiVersion: stash.appscode.com/v1beta1
kind: BackupConfiguration
metadata:
  name: sample-elasticsearch-backup
  namespace: demo
spec:
  schedule: "*/2 * * * *"
#  task: # Uncomment if you are not using KubeDB to deploy your database
#    name: elasticsearch-backup-{{< param "info.subproject_version" >}}
  repository:
    name: gcs-repo
  target:
    ref:
      apiVersion: appcatalog.appscode.com/v1alpha1
      kind: AppBinding
      name: sample-elasticsearch
  interimVolumeTemplate:
    metadata:
      name: stash-tmp-backup-storage
    spec:
      accessModes: [ "ReadWriteOnce" ]
      storageClassName: "standard"
      resources:
        requests:
          storage: 1Gi
  runtimeSettings:
    pod:
      securityContext:
        runAsUser: 0
        runAsGroup: 0
  retentionPolicy:
    name: keep-last-5
    keepLast: 5
    prune: true
```

### Specifying Memory/CPU limit/request for the backup job

If you want to specify the Memory/CPU limit/request for your backup job, you can specify `resources` field under `runtimeSettings.container` section.

```yaml
apiVersion: stash.appscode.com/v1beta1
kind: BackupConfiguration
metadata:
  name: sample-elasticsearch-backup
  namespace: demo
spec:
  schedule: "*/2 * * * *"
#  task: # Uncomment if you are not using KubeDB to deploy your database
#    name: elasticsearch-backup-{{< param "info.subproject_version" >}}
  repository:
    name: gcs-repo
  target:
    ref:
      apiVersion: appcatalog.appscode.com/v1alpha1
      kind: AppBinding
      name: sample-elasticsearch
  interimVolumeTemplate:
    metadata:
      name: stash-tmp-backup-storage
    spec:
      accessModes: ["ReadWriteOnce"]
      storageClassName: "standard"
      resources:
        requests:
          storage: 1Gi
  runtimeSettings:
    container:
      resources:
        requests:
          cpu: "200m"
          memory: "1Gi"
        limits:
          cpu: "200m"
          memory: "1Gi"
  retentionPolicy:
    name: keep-last-5
    keepLast: 5
    prune: true
```

### Using multiple retention policies

You can also specify multiple retention policies for your backed up data. For example, you may want to keep few daily snapshots, few weekly snapshots, and few monthly snapshots, etc. You just need to pass the desired number with the respective key under the `retentionPolicy` section.

```yaml
apiVersion: stash.appscode.com/v1beta1
kind: BackupConfiguration
metadata:
  name: sample-elasticsearch-backup
  namespace: demo
spec:
  schedule: "*/2 * * * *"
#  task: # Uncomment if you are not using KubeDB to deploy your database
#    name: elasticsearch-backup-{{< param "info.subproject_version" >}}
  repository:
    name: gcs-repo
  target:
    ref:
      apiVersion: appcatalog.appscode.com/v1alpha1
      kind: AppBinding
      name: sample-elasticsearch
  interimVolumeTemplate:
    metadata:
      name: stash-tmp-backup-storage
    spec:
      accessModes: [ "ReadWriteOnce" ]
      storageClassName: "standard"
      resources:
        requests:
          storage: 1Gi
  retentionPolicy:
    name: sample-es-retention
    keepLast: 5
    keepDaily: 10
    keepWeekly: 20
    keepMonthly: 50
    keepYearly: 100
    prune: true
```

To know more about the available options for retention policies, please visit [here](/docs/concepts/crds/backupconfiguration/#specretentionpolicy).

## Customizing Restore Process

Stash also uses `multielasticdump` during the restore process. In this section, we are going to show how you can pass arguments to the restore process, restore a specific snapshot, run restore job as a specific user, etc.

### Passing arguments to the restore process

Similar to the backup process, you can pass arguments to the restore process through the `args` params under `task.params` section.

```yaml
apiVersion: stash.appscode.com/v1beta1
kind: RestoreSession
metadata:
  name: sample-elasticsearch-restore
  namespace: demo
spec:
  task:
#    name: elasticsearch-restore-{{< param "info.subproject_version" >}} # Uncomment if you are not using KubeDB to deploy your database
    params:
    - name: args
      value: --ignoreType=template,settings
  repository:
    name: gcs-repo
  target:
    ref:
      apiVersion: appcatalog.appscode.com/v1alpha1
      kind: AppBinding
      name: sample-elasticsearch
  interimVolumeTemplate:
    metadata:
      name: stash-tmp-restore-storage
    spec:
      accessModes: [ "ReadWriteOnce" ]
      storageClassName: "standard"
      resources:
        requests:
          storage: 1Gi
  rules:
  - snapshots: [latest]
```

### Restore specific snapshot

You can also restore a specific snapshot. At first, list the available snapshot as bellow,

```bash
❯ kubectl get snapshots -n demo
NAME                REPOSITORY   HOSTNAME   CREATED AT
gcs-repo-4bc21d6f   gcs-repo     host-0     2021-02-12T14:54:27Z
gcs-repo-f0ac7cbd   gcs-repo     host-0     2021-02-12T14:56:26Z
gcs-repo-9210ebb6   gcs-repo     host-0     2021-02-12T14:58:27Z
gcs-repo-0aff8890   gcs-repo     host-0     2021-02-12T15:00:28Z
```

>You can also filter the snapshots as shown in the guide [here](https://stash.run/docs/v2021.01.21/concepts/crds/snapshot/#working-with-snapshot).

Stash adds the Repository name as a prefix of the Snapshot. You have to remove the repository prefix and use only the last 8 characters as the snapshot name during restore.

The below example shows how you can pass a specific snapshot name through the `snapshots` filed of `rules` section.

```yaml
apiVersion: stash.appscode.com/v1beta1
kind: RestoreSession
metadata:
  name: sample-elasticsearch-restore
  namespace: demo
spec:
#  task: # Uncomment if you are not using KubeDB to deploy your database
#    name: elasticsearch-restore-{{< param "info.subproject_version" >}}
  repository:
    name: gcs-repo
  target:
    ref:
      apiVersion: appcatalog.appscode.com/v1alpha1
      kind: AppBinding
      name: sample-elasticsearch
  interimVolumeTemplate:
    metadata:
      name: stash-tmp-restore-storage
    spec:
      accessModes: [ "ReadWriteOnce" ]
      storageClassName: "standard"
      resources:
        requests:
          storage: 1Gi
  rules:
  - snapshots: [4bc21d6f]
```

>Please, do not specify multiple snapshots here. Each snapshot represents a complete backup of your database. Multiple snapshots are only usable during file/directory restore.

### Running restore job as a specific user

You can provide `securityContext` under `runtimeSettings.pod` section to run the restore job as a specific user.

```yaml
apiVersion: stash.appscode.com/v1beta1
kind: RestoreSession
metadata:
  name: sample-elasticsearch-restore
  namespace: demo
spec:
#  task: # Uncomment if you are not using KubeDB to deploy your database
#    name: elasticsearch-restore-{{< param "info.subproject_version" >}}
  repository:
    name: gcs-repo
  target:
    ref:
      apiVersion: appcatalog.appscode.com/v1alpha1
      kind: AppBinding
      name: sample-elasticsearch
  interimVolumeTemplate:
    metadata:
      name: stash-tmp-restore-storage
    spec:
      accessModes: [ "ReadWriteOnce" ]
      storageClassName: "standard"
      resources:
        requests:
          storage: 1Gi
  runtimeSettings:
    pod:
      securityContext:
        runAsUser: 0
        runAsGroup: 0
  rules:
  - snapshots: [latest]
```

### Specifying Memory/CPU limit/request for the restore job

Similar to the backup process, you can also provide `resources` field under the `runtimeSettings.container` section to limit the Memory/CPU for your restore job.

```yaml
apiVersion: stash.appscode.com/v1beta1
kind: RestoreSession
metadata:
  name: sample-elasticsearch-restore
  namespace: demo
spec:
#  task: # Uncomment if you are not using KubeDB to deploy your database
#    name: elasticsearch-restore-{{< param "info.subproject_version" >}}
  repository:
    name: gcs-repo
  target:
    ref:
      apiVersion: appcatalog.appscode.com/v1alpha1
      kind: AppBinding
      name: sample-elasticsearch
  interimVolumeTemplate:
    metadata:
      name: stash-tmp-restore-storage
    spec:
      accessModes: [ "ReadWriteOnce" ]
      storageClassName: "standard"
      resources:
        requests:
          storage: 1Gi
  runtimeSettings:
    container:
      resources:
        requests:
          cpu: "200m"
          memory: "1Gi"
        limits:
          cpu: "200m"
          memory: "1Gi"
  rules:
  - snapshots: [latest]
```
