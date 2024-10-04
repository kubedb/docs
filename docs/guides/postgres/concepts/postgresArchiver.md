---
title: PostgresArchiver CRD
menu:
  docs_{{ .version }}:
    identifier: pg-PostgresArchiver-archiver-concepts
    name: PostgresArchiver
    parent: pg-concepts-PostgresArchiver
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).
> Also you need to have basic understanding of [Kubestash concepts](https://kubestash.com/docs/v2024.9.30/concepts/) before proceeding.
# PostgresArchiver

## What is PostgresArchiver

`PostgresArchiver` is a Kubernetes `Custom Resource Definitions` (CRD). It provides declarative configuration for taking `full-backup` and `wal-backup` for `point-in-time-recovery` in a Kubernetes native way. You only need to describe the desired `pitr` configuration in a PostgresArchiver Archiver object, and the KubeDB operator will create Kubernetes objects in the desired state for you.

## PostgresArchiver Spec

As with all other Kubernetes objects, a PostgresArchiver needs `apiVersion`, `kind`, and `metadata` fields. It also needs a `.spec` section.

Below is an example PostgresArchiver object.

```yaml
apiVersion: archiver.kubedb.com/v1alpha1
kind: PostgresArchiver
metadata:
  name: PostgresArchiver-sample
  namespace: demo
spec:
  pause: false
  databases:
    namespaces:
      from: Selector
      selector:
        matchLabels:
          kubernetes.io/metadata.name: demo
    selector:
      matchLabels:
        archiver: "true"
  retentionPolicy:
    name: PostgresArchiver-retention-policy
    namespace: demo
  encryptionSecret:
    name: "encrypt-secret"
    namespace: "demo"
  fullBackup:
    driver: "Restic"
    scheduler:
      successfulJobsHistoryLimit: 1
      failedJobsHistoryLimit: 1
      schedule: "*/10 * * * *"
    sessionHistoryLimit: 2
  manifestBackup:
    scheduler:
      successfulJobsHistoryLimit: 1
      failedJobsHistoryLimit: 1
      schedule: "30 * * * *"
    sessionHistoryLimit: 2
  backupStorage:
    ref:
      name: "gcs-storage"
      namespace: "restore"
```

### spec.databases

Databases define which Postgres databases are allowed to consume this archiver.

#### spec.databases.namespaces.from
This indicates namespaces from which Consumers may be attached to. Default value for this field is `same`. Which means only db's from same namespace can consume this postgres archiver.
Other supported values are `All`, `Selector`. If you use `Selector`, then your consumer namespace labels should be matched with the `.spec.databases.namespaces.selector`.

#### spec.databases.selector (selctor)

This specifies a selector for consumers that are allowed to bind to this database instance. For example, if you have following in your archiver spec,
```yaml
selector:
  matchLabels:
    archiver: "true"
```
then only those consumers will be allowed to consume this who has the below labels.
```yaml
metadata:
  labels:
    archiver: "true"
```
> Note: Only those consumers are allowed to consume this archiver which satisfies both `spec.databases.namespaces` & `spec.databases.selector`.

### spec.pause (bool)
Pause defines if the backup process should be paused or not. It is a boolean field.

### spec.retentionPolicy (objectReference)

RetentionPolicy field is the RetentionPolicy of the backupConfiguration's backend.
Check [here](https://kubestash.com/docs/v2024.9.30/concepts/crds/retentionpolicy/) for more details about retentionPolicy.

### spec.fullBackup

FullBackup defines the sessionConfig of the fullBackup. This options will eventually go to the full-backup job's yaml. 
For knowing more about sessionConfig and Backup, visit [this](https://kubestash.com/docs/v2024.9.30/concepts/crds/backupconfiguration/).

#### spec.fullBackup.driver (string)

Driver specifies the name of underlying tool that is being used to upload the backed up data.
Supported values are, `Restic`,`WalG`, `VolumeSnapshotter`.

#### spec.fullBackup.task.Params (RawExtension)
Task will let you provide backup params. To know about `params` section, visit [here](https://kubestash.com/docs/v2024.9.30/concepts/crds/backupconfiguration/#task-reference).

#### spec.fullBackup.scheduler

Checkout [here](https://kubestash.com/docs/v2024.9.30/concepts/crds/backupconfiguration/#scheduler-spec) to know about scheduler. Only `schedule`, `concurrencyPolicy`, `jobTemplate`, `successfulJobsHistoryLimit` and `successfulJobsHistoryLimit` are supported.

#### spec.fullBackup.containerRuntimeSettings
This provides settings required to run backup pods. Checkout [here](https://github.com/kmodules/offshoot-api/blob/6f79b0d0097965c01877e2e2fa8265b3909eb4de/api/v1/runtime_settings_types.go#L125) for more details.
#### spec.fullBackup.jobTemplate
Use to run the backup job. For more info, visit [here](https://github.com/kmodules/offshoot-api/blob/6f79b0d0097965c01877e2e2fa8265b3909eb4de/api/v1/types.go#L42).

#### spec.fullBackup.retryConfig.maxRetry
MaxRetry specifies the maximum number of times KubeStash should retry the backup/restore process. By default, KubeStash will retry only 1 time.
#### spec.fullBackup.retryConfig.delay
The amount of time to wait before next retry. If you don't specify this field, KubeStash will retry immediately. Format: 30s, 2m, 1h etc. 

### spec.walBackup

WalBackup defines the sessionConfig of the walBackup. This options will eventually go to the sidekick specification. We use to run a [sidekick](https://github.com/kubeops/sidekick) pod to take continuous wal backup using wal-g.

#### spec.walBackup.configSecret (genericSecretReference)
A secret reference which you can use to set custom env's to the wal-backup pod.

#### spec.walBackup.runTimeSettings

Settings to run wal-backup pod. More information can be found [here](https://github.com/kmodules/offshoot-api/blob/6f79b0d0097965c01877e2e2fa8265b3909eb4de/api/v1/runtime_settings_types.go#L26).

### spec.manifestBackup

ManifestBackup defines the sessionConfig of the manifestBackup. This options will eventually go to the manifest-backup job's yaml.  For knowing more about sessionConfig and Backup, visit [this](https://kubestash.com/docs/v2024.9.30/concepts/crds/backupconfiguration/).

Both `ManifestBackup` and `FullBackup` share the following fields:
- `Scheduler`
- `ContainerRuntimeSettings`
- `JobTemplate`
- `RetryConfig`
- `Timeout`
- `SessionHistoryLimit`

### spec.encryptionSecret (objectReference)

EncryptionSecret refers to the Secret containing the encryption key which will be used to encrypt the backed up data. You can refer to a Secret of a different namespace by providing name and namespace fields. This field is optional. No encryption secret is required for VolumeSnapshot backups. 
To know more about this, visit [here](https://kubestash.com/docs/v2024.9.30/concepts/crds/backupconfiguration/#backupconfiguration-spec). 

### spec.backupStorage

BackupStorage is the backend storageRef of the BackupConfiguration. To learn more about this, visit [here](https://kubestash.com/docs/v2024.9.30/concepts/crds/backupstorage/).

#### spec.backupStorage.ref (objectReference)

Reference to the backupStorage.
#### spec.backupStorage.subDir (string)

If this is set then this directory path will be used to store the backups.

### spec.deletionPolicy
DeletionPolicy defines the created repository's deletionPolicy.

## Next Steps

- Learn how to use KubeDB to run a PostgresArchiverQL database [here](/docs/guides/PostgresArchiver/README.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
