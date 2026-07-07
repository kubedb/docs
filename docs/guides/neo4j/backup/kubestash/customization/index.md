---
title: Neo4j Backup Customization | KubeStash
description: Customizing Neo4j Backup and Restore process with KubeStash
menu:
  docs_{{ .version }}:
    identifier: guides-neo4j-backup-customization-stashv2
    name: Customizing Backup & Restore Process
    parent: guides-neo4j-backup-stashv2
    weight: 50
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Customizing Backup and Restore Process

KubeStash provides rich customization supports for the backup and restore process to meet the requirements of various cluster configurations. This guide will show you some examples of these customizations.

KubeStash uses the `neo4j-admin database backup` command under the hood to take backups and the `neo4j-admin database restore` command to restore them. By default, it backs up every database of the instance (including the `system` database). The customizations shown below are passed through the `addon.tasks[*].params` and `addon.jobTemplate.spec` sections of the `BackupConfiguration` and `RestoreSession` CRs.

## Customizing Backup Process

In this section, we are going to show you how to customize the backup process. Here, we are going to show some examples of passing arguments to the backup process, backing up specific databases, running the backup process as a specific user, etc.

### Passing arguments to the backup process

KubeStash Neo4j addon uses the [neo4j-admin database backup](https://neo4j.com/docs/operations-manual/current/backup-restore/online-backup/) command for backups. You can pass any extra arguments supported by this command through the `neo4jAdminArgs` parameter under the `addon.tasks[*].params` section. To pass multiple arguments, provide them as a comma-separated list.

The below example shows how you can pass `--keep-failed=true` and `--parallel-recovery=true` to the backup command.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: sample-neo4j-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Neo4j
    namespace: demo
    name: sample-neo4j
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
        - name: s3-neo4j-repo
          backend: s3-backend
          directory: /neo4j
      addon:
        name: neo4j-addon
        tasks:
          - name: logical-backup
            params:
              neo4jAdminArgs: "--keep-failed=true,--parallel-recovery=true"
```

### Passing a target database to the backup process

By default, KubeStash Neo4j addon backs up every database of the instance (the `*` selector). If you want to back up only a specific set of databases, you can specify them using the `databases` parameter under the `addon.tasks[*].params` section. Provide the database names as a comma-separated list.

The below example shows how you can back up only the `neo4j` and `movies` databases.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: sample-neo4j-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Neo4j
    namespace: demo
    name: sample-neo4j
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
        - name: s3-neo4j-repo
          backend: s3-backend
          directory: /neo4j
      addon:
        name: neo4j-addon
        tasks:
          - name: logical-backup
            params:
              databases: "neo4j,movies"
```

> **WARNING**: Make sure that the databases you provide already exist in the target instance before taking the backup.

### Backing up from a specific server

By default, KubeStash takes the backup from the server resolved through the database `AppBinding` (over the backup port `6362`). If you want to take the backup from a specific server, for example to offload the backup load from the leader to a particular replica, you can set the `from` parameter under the `addon.tasks[*].params` section to the desired server address.

The below example shows how you can take the backup from the `sample-neo4j-2` pod.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: sample-neo4j-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Neo4j
    namespace: demo
    name: sample-neo4j
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
        - name: s3-neo4j-repo
          backend: s3-backend
          directory: /neo4j
      addon:
        name: neo4j-addon
        tasks:
          - name: logical-backup
            params:
              from: "sample-neo4j-2.demo.svc:6362"
```

> The `from` address must point to a server that exposes the backup port (`6362`).

### Running backup job as a specific user

If your cluster requires running the backup job as a specific user, you can provide `securityContext` under the `addon.jobTemplate.spec.securityContext` section. The below example shows how you can run the backup job as the `neo4j` user (`runAsUser: 7474`).

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: sample-neo4j-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Neo4j
    namespace: demo
    name: sample-neo4j
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
        - name: s3-neo4j-repo
          backend: s3-backend
          directory: /neo4j
      addon:
        name: neo4j-addon
        jobTemplate:
          spec:
            securityContext:
              runAsUser: 7474
              runAsGroup: 7474
        tasks:
          - name: logical-backup
```

### Specifying Memory/CPU limit/request for the backup job

If you want to specify the Memory/CPU limit/request for your backup job, you can specify the `resources` field under the `addon.jobTemplate.spec` section.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: BackupConfiguration
metadata:
  name: sample-neo4j-backup
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Neo4j
    namespace: demo
    name: sample-neo4j
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
        - name: s3-neo4j-repo
          backend: s3-backend
          directory: /neo4j
      addon:
        name: neo4j-addon
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

`KubeStash` uses the `neo4j-admin database restore` command during the restore process and then seeds the restored store into the cluster from a single bootstrap pod. In this section, we are going to show how you can restore a specific snapshot, restore specific databases, pass arguments to the restore process, run the restore job as a specific user, etc.

> **Note:** For a clustered `Neo4j` restore, KubeStash restores the store files into the seed pod's data volume and then bootstraps the other replicas from it. Therefore, in every `RestoreSession` you need to set the `seedServerName` parameter to the target seed pod (e.g. `restored-neo4j-0`) and mount that pod's data PVC into the restore `Job` as shown in the examples below.

### Restore specific snapshot

You can also restore a specific snapshot. At first, list the available snapshots as below,

```bash
kubectl get snapshots.storage.kubestash.com -n demo -l=kubestash.com/repo-name=s3-neo4j-repo
```
NAME                                                          REPOSITORY      SESSION           SNAPSHOT-TIME          DELETION-POLICY   PHASE       AGE
s3-neo4j-repo-sample-neo4j-backup-frequent-backup-1725257849   s3-neo4j-repo   frequent-backup   2024-09-02T06:18:01Z   Delete            Succeeded   15m
s3-neo4j-repo-sample-neo4j-backup-frequent-backup-1725258000   s3-neo4j-repo   frequent-backup   2024-09-02T06:20:00Z   Delete            Succeeded   13m
s3-neo4j-repo-sample-neo4j-backup-frequent-backup-1725258300   s3-neo4j-repo   frequent-backup   2024-09-02T06:25:00Z   Delete            Succeeded   8m34s
s3-neo4j-repo-sample-neo4j-backup-frequent-backup-1725258600   s3-neo4j-repo   frequent-backup   2024-09-02T06:30:00Z   Delete            Succeeded   3m34s

The below example shows how you can pass a specific snapshot name in the `.spec.dataSource` section.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: sample-neo4j-restore
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Neo4j
    namespace: demo
    name: restored-neo4j
  dataSource:
    repository: s3-neo4j-repo
    snapshot: s3-neo4j-repo-sample-neo4j-backup-frequent-backup-1725258000
  addon:
    name: neo4j-addon
    tasks:
      - name: logical-backup-restore
        params:
          seedServerName: "restored-neo4j-0" ## Neo4j Pod Name
    jobTemplate:
      spec:
        volumes:
          - name: data
            persistentVolumeClaim:
              claimName: data-restored-neo4j-0 # PVC Name
        volumeMounts:
          - mountPath: /data
            name: data
            subPath: data
        securityContext:
          runAsNonRoot: true
          runAsUser: 7474
```

### Restoring specific databases

A snapshot contains every database that was backed up. If you want to restore only a subset of them, you can specify the database names using the `databases` parameter under the `addon.tasks[*].params` section. Provide the database names as a comma-separated list.

> **Note:** The `system` database is never restored, regardless of the value of `databases`. If you omit the `databases` parameter, all user databases from the snapshot are restored.

The below example shows how you can restore only the `movies` database.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: sample-neo4j-restore
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Neo4j
    namespace: demo
    name: restored-neo4j
  dataSource:
    repository: s3-neo4j-repo
    snapshot: latest
  addon:
    name: neo4j-addon
    tasks:
      - name: logical-backup-restore
        params:
          seedServerName: "restored-neo4j-0" ## Neo4j Pod Name
          databases: "movies"
    jobTemplate:
      spec:
        volumes:
          - name: data
            persistentVolumeClaim:
              claimName: data-restored-neo4j-0 # PVC Name
        volumeMounts:
          - mountPath: /data
            name: data
            subPath: data
        securityContext:
          runAsNonRoot: true
          runAsUser: 7474
```

### Passing arguments to the restore process

You can pass any extra arguments supported by the `neo4j-admin database restore` command through the `neo4jAdminArgs` parameter under the `addon.tasks[*].params` section. To pass multiple arguments, provide them as a comma-separated list.

A common use case is restoring into a database that already exists. Passing `--overwrite-destination=true` tells KubeStash to stop and drop the existing database before restoring it.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: sample-neo4j-restore
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Neo4j
    namespace: demo
    name: restored-neo4j
  dataSource:
    repository: s3-neo4j-repo
    snapshot: latest
  addon:
    name: neo4j-addon
    tasks:
      - name: logical-backup-restore
        params:
          seedServerName: "restored-neo4j-0" ## Neo4j Pod Name
          neo4jAdminArgs: "--overwrite-destination=true"
    jobTemplate:
      spec:
        volumes:
          - name: data
            persistentVolumeClaim:
              claimName: data-restored-neo4j-0 # PVC Name
        volumeMounts:
          - mountPath: /data
            name: data
            subPath: data
        securityContext:
          runAsNonRoot: true
          runAsUser: 7474
```

### Specifying Memory/CPU limit/request for the restore job

Similar to the backup process, you can also provide the `resources` field under the `addon.jobTemplate.spec` section to limit the Memory/CPU for your restore job.

```yaml
apiVersion: core.kubestash.com/v1alpha1
kind: RestoreSession
metadata:
  name: sample-neo4j-restore
  namespace: demo
spec:
  target:
    apiGroup: kubedb.com
    kind: Neo4j
    namespace: demo
    name: restored-neo4j
  dataSource:
    repository: s3-neo4j-repo
    snapshot: latest
  addon:
    name: neo4j-addon
    tasks:
      - name: logical-backup-restore
        params:
          seedServerName: "restored-neo4j-0" ## Neo4j Pod Name
    jobTemplate:
      spec:
        resources:
          requests:
            cpu: "200m"
            memory: "1Gi"
          limits:
            cpu: "200m"
            memory: "1Gi"
        volumes:
          - name: data
            persistentVolumeClaim:
              claimName: data-restored-neo4j-0 # PVC Name
        volumeMounts:
          - mountPath: /data
            name: data
            subPath: data
        securityContext:
          runAsNonRoot: true
          runAsUser: 7474
```

> You can configure additional runtime settings for restore jobs within the `addon.jobTemplate.spec` sections. For further details, please refer to the [reference](https://kubestash.com/docs/latest/concepts/crds/restoresession/#podtemplate-spec).
