---
title: MongoDBArchiver CRD
menu:
  docs_{{ .version }}:
    identifier: mg-archiver-concepts
    name: MongoDBArchiver
    parent: mg-concepts-mongodb
    weight: 30
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# MongoDBArchiver

## What is MongoDBArchiver

`MongoDBArchiver` is a Kubernetes `Custom Resource Definitions` (CRD). It provides a declarative configuration for backup and restore [MongoDB](https://www.mongodb.com/) database in a Kubernetes native way.

## MongoDBArchiver CRD Specifications

Like any official Kubernetes resource, a `MongoDBArchiver` has `TypeMeta`, `ObjectMeta`, `Spec` and `Status` sections.

Here, a sample `MongoDBArchiver` CRO for backing up a mongodb database is given below:

**Sample `MongoDBArchiver`:**

```yaml
apiVersion: archiver.kubedb.com/v1alpha1
kind: MongoDBArchiver
metadata:
  name: mongodbarchiver-sample
  namespace: demo
spec:
  pause: false
  databases:
    namespaces:
      from: "Same"
    selector:
      matchLabels:
        archiver: "true"
  retentionPolicy:
    name: mongodb-retention-policy
    namespace: demo
  fullBackup:
    driver: "CSISnapshotter"
    csiSnapshotter:
      volumeSnapshotClassName: "longhorn-snapshot-vsc"
    scheduler:
      schedule: "*/5 * * * *"
  manifestBackup:
    encryptionSecret:
      name: "encrypt-secret"
      namespace: "demo"
    scheduler:
      schedule: "*/5 * * * *"
  backupStorage:
    ref:
      apiGroup: "storage.kubestash.com"
      kind: "BackupStorage"
      name: "linode-storage"
      namespace: "demo"
```

Here, we are going to describe the various sections of a `MongoDBArchiver` crd.

A `MongoDBArchiver` object has the following fields in the `spec` section.

### spec.databases

`spec.databases` is a field that specifies the mongodb databases this mongodbarchiver will select. This field consists of the following sub-field:

- `spec.databases.namespaces` : specifies the allowed namespaces.
- `spec.databases.selector` : specifies the allowed database's selector.

### spec.retentionPolicy

`spec.retentionPolicy` is a field that specifies the retention policy that will be applied to the backup of the database. This field consists of the following sub-field:

- `spec.retentionPolicy.name` : specifies the name of the retention policy CR.
- `spec.retentionPolicy.namespace` : specifies the namespace of the retention policy CR.

### spec.pause
`spec.pause` is a boolean field that specifies if the archiver is currently paused or not. If the archiver is paused, the backup supporting resources such as backupConfiguration is paused and walg oplog backup is also paused by deleting the Sidekick.

### spec.backupStorage
`spec.backupStorage` is a field that specifies the backupStorage that will be used to store the backup. It is a reference to the BackupStorage object of kubeStash.
`spec.backupStorage` has `.ref` field which holds the full reference(group, kind, name, namespace) of the backup Storage object. It also has a `.prefix` field specifying the backup folder-name prefix.

### spec.fullBackup
`spec.fullBackup` is used for different options in fullBackup. It has the following sub-fields:

- `driver` specifies the driver we are using for full backup. Currently only one driver, `CSISnapshotter` driver is supported.
- `csiSnapshotter` specifies the csiSnapshotter driver options such as the volume snapshot class name.
- `scheduler` specifies the scheduler options for the full database backup such as the schedule, job template etc.
- `containerRuntimeSettings` specifies the container runtime settings for the full backup. For more information check [here](https://github.com/kmodules/offshoot-api/blob/master/api/v1/runtime_settings_types.go#L122-L173).
- `jobTemplate` specifies the job template that is used in the backup session created for the full backup. For more information check [here](https://github.com/kmodules/offshoot-api/blob/master/api/v1/types.go#L42-L57).
- `retryConfig` specifies the behavior of the retry.
- `timeout` specifies the timeout for the backup.
- `sessionHistoryLimit` specifies how many backup Jobs and associate resources Stash should keep for debugging purpose.


### spec.manifestBackup
`spec.manifestBackup` is used for different options in manifest (auth secret, config secret etc.) backup. It has the following sub-fields:

- `encryptionSecret` specifies the secret name and namespace which is used to encrypt the data that is being backed up.
- `scheduler` specifies the scheduler options for the full database backup such as the schedule, job template etc.
- `containerRuntimeSettings` specifies the container runtime settings for the manifest backup. For more information check [here](https://github.com/kmodules/offshoot-api/blob/master/api/v1/runtime_settings_types.go#L122-L173).
- `jobTemplate` specifies the job template that is used in the backup session created for the manifest backup. For more information check [here](https://github.com/kmodules/offshoot-api/blob/master/api/v1/types.go#L42-L57).
- `retryConfig` specifies the behavior of the retry.
- `timeout` specifies the timeout for the backup.
- `sessionHistoryLimit` specifies how many backup Jobs and associate resources Stash should keep for debugging purpose. 

### spec.walBackup
`spec.walBackup` is used for different options in walg backup. It has the following sub-fields:

- `runtimeSettings` specifies different runtime settings for pod and containers. For more information check [here](https://github.com/kmodules/offshoot-api/blob/master/api/v1/runtime_settings_types.go#L26-L29).
- `configSecret` specifies the name and namespace of the secret which contains different wal-g options. You can find the list of options that you can provide for walg [here](https://wal-g.readthedocs.io/MongoDB/#configuration).

### spec.deletionPolicy
`spec.deletionPolicy` is a field that specifies that what will happen to the data if the archiver is deleted. It has three options `WipeOut`, `Delete` & `DoNotDelete`.

### status
The status section of MongoDBArchiver only has one field `databaseRefs`. It holds the name & namespace of all the databases which are managed by this archiver.

## Next Steps

- Learn how to use MongoDBArchiver to backup and restore MongoDB database [here](/docs/guides/mongodb/backup/archiver).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
