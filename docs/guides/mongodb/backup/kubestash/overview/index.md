---
title: MongoDB Backup & Restore Overview | KubeStash
menu:
  docs_{{ .version }}:
    identifier: guides-mongodb-backup-kubestash-overview
    name: Overview
    parent: guides-mongodb-backup-stashv2
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeStash Enterprise Edition] to try this feature. You can use KubeDB Enterprise license to install KubeStash Enterprise edition. Database backup with KubeStash is already included in the KubeDB Enterprise license. So, you don't need a separate license for KubeStash." >}}

# MongoDB Backup & Restore Overview

KubeDB can also uses [KubeStash](https://kubestash.com/) to backup and restore databases. KubeStash by AppsCode is a cloud native data backup and recovery solution for Kubernetes workloads. KubeStash utilizes [restic](https://github.com/restic/restic) to securely backup stateful applications to any cloud or on-prem storage backends (for example, S3, GCS, Azure Blob storage, Minio, NetApp, Dell EMC etc.).

## How Backup Works

The following diagram shows how KubeStash takes backup of a MongoDB database. Open the image in a new tab to see the enlarged version.

<figure align="center">
 <img alt="MongoDB Backup Overview" src="/docs/guides/mongodb/backup/kubestash/overview/images/backup_overview.svg">
  <figcaption align="center">Fig: MongoDB Backup Overview</figcaption>
</figure>

The backup process consists of the following steps:

1. At first, a user creates a secret with access credentials of the backend where the backed up data will be stored.

2. Then, she creates a `BackupStorage` crd that specifies the backend information along with the secret that holds the credentials to access the backend.

3. Then, she creates a `BackupConfiguration` crd targeting the crd of the desired `MongoDB` database. The `BackupConfiguration` object also specifies the `Sessions` to use to backup the database.

4. KubeStash operator watches for `BackupConfiguration` crd.

5. Once KubeStash operator finds a `BackupConfiguration` crd, it creates a CronJob with the schedule specified in `BackupConfiguration` object to trigger backup periodically.

6. On the next scheduled slot, the CronJob triggers a backup by creating a `BackupSession` crd.

7. KubeStash operator also watches for `BackupSession` crd.

8. When it finds a `BackupSession` object, It creates a `Snapshot` for holding backup information. 

9. KubeStash resolves the respective `Addon` and `Function` and prepares a Job definition to backup.

10. Then, it creates the Job to backup the targeted database.

11. The backup Job reads necessary information to connect with the database from the `AppBinding` crd of the targated `MongoDB` database. It also reads backend information and access credentials from `BackupStorage` crd and Storage Secret respectively through `Backend` section of `BackupConfiguration` crd

12. Then, the Job dumps the targeted database and uploads the output to the backend. KubeStash pipes the output of dump command to uploading process. Hence, backup Job does not require a large volume to hold the entire dump output.

13. Finally, when the backup is complete, the Job sends Prometheus metrics to the Pushgateway running inside KubeStash operator pod. It also updates the `BackupSession` and `Snapshot` status to reflect the backup procedure.

### Backup Different MongoDB Configurations

This section will show you how backup works for different MongoDB configurations.

#### Standalone MongoDB

For a standalone MongoDB database, the backup job directly dumps the database using `mongodump` and pipe the output to the backup process.

<figure align="center">
 <img alt="Standalone MongoDB Backup Overview" src="/docs/guides/mongodb/backup/kubestash/overview/images/standalone_backup.svg">
  <figcaption align="center">Fig: Standalone MongoDB Backup</figcaption>
</figure>

#### MongoDB ReplicaSet Cluster

For MongoDB ReplicaSet cluster, KubeStash takes backup from one of the secondary replicas. The backup process consists of the following steps:

1. Identify a secondary replica.
2. Lock the secondary replica.
3. Backup the secondary replica.
4. Unlock the secondary replica.

<figure align="center">
 <img alt="MongoDB ReplicaSet Cluster Backup Overview" src="/docs/guides/mongodb/backup/kubestash/overview/images/replicaset_backup.svg">
  <figcaption align="center">Fig: MongoDB ReplicaSet Cluster Backup</figcaption>
</figure>

#### MongoDB Sharded Cluster

For MongoDB sharded cluster, KubeStash takes backup of the individual shards as well as the config server. KubeStash takes backup from a secondary replica of the shards and the config server. If there is no secondary replica then KubeStash will take backup from the primary replica. The backup process consists of the following steps:

1. Disable balancer.
2. Lock config server.
3. Identify a secondary replica for each shard.
4. Lock the secondary replica.
5. Run backup on the secondary replica.
6. Unlock the secondary replica.
7. Unlock config server.
8. Enable balancer.

<figure align="center">
 <img alt="MongoDB Sharded Cluster Backup Overview" src="/docs/guides/mongodb/backup/kubestash/overview/images/sharded_backup.svg">
  <figcaption align="center">Fig: MongoDB Sharded Cluster Backup</figcaption>
</figure>

## How Restore Process Works

The following diagram shows how KubeStash restores backed up data into a MongoDB database. Open the image in a new tab to see the enlarged version.

<figure align="center">
 <img alt="Database Restore Overview" src="/docs/guides/mongodb/backup/kubestash/overview/images/restore_overview.svg">
  <figcaption align="center">Fig: MongoDB Restore Process Overview</figcaption>
</figure>

The restore process consists of the following steps:

1. At first, a user creates a `RestoreSession` crd targeting the `AppBinding` of the desired database where the backed up data will be restored. It also specifies the `Repository` crd which holds the backend information and the `Task` to use to restore the target.

2. KubeStash operator watches for `RestoreSession` object.

3. Once it finds a `RestoreSession` object, it resolves the respective `Task` and `Function` and prepares a Job definition to restore.

4. Then, it creates the Job to restore the target.

5. The Job reads necessary information to connect with the database from respective `AppBinding` crd. It also reads backend information and access credentials from `Repository` crd and Storage Secret respectively.

6. Then, the job downloads the backed up data from the backend and injects into the desired database. KubeStash pipes the downloaded data to the respective database tool to inject into the database. Hence, restore job does not require a large volume to download entire backup data inside it.

7. Finally, when the restore process is complete, the Job sends Prometheus metrics to the Pushgateway and update the `RestoreSession` status to reflect restore completion.

### Restoring Different MongoDB Configurations

This section will show you restore process works for different MongoDB configurations.

#### Standalone MongoDB

For a standalone MongoDB database, the restore job downloads the backed up data from the backend and pipe the downloaded data to `mongorestore` command which inserts the data into the desired MongoDB database.

<figure align="center">
 <img alt="Standalone MongoDB Restore Overview" src="/docs/guides/mongodb/backup/kubestash/overview/images/standalone_restore.svg">
  <figcaption align="center">Fig: Standalone MongoDB Restore</figcaption>
</figure>

#### MongoDB ReplicaSet Cluster

For MongoDB ReplicaSet cluster, KubeStash identifies the primary replica and restore into it.

<figure align="center">
 <img alt="MongoDB ReplicaSet Cluster Restore Overview" src="/docs/guides/mongodb/backup/kubestash/overview/images/replicaset_restore.svg">
  <figcaption align="center">Fig: MongoDB ReplicaSet Cluster Restore</figcaption>
</figure>

#### MongoDB Sharded Cluster

For MongoDB sharded cluster, KubeStash identifies the primary replica of each shard as well as the config server and restore respective backed up data into them.

<figure align="center">
 <img alt="MongoDB Sharded Cluster Restore" src="/docs/guides/mongodb/backup/kubestash/overview/images/sharded_restore.svg">
  <figcaption align="center">Fig: MongoDB Sharded Cluster Restore</figcaption>
</figure>

## Next Steps

- Backup a standalone MongoDB databases using KubeStash following the guide from [here](/docs/guides/mongodb/backup/kubestash/logical/standalone/index.md).
- Backup a MongoDB Replicaset cluster using KubeStash following the guide from [here](/docs/guides/mongodb/backup/kubestash/logical/replicaset/index.md).
- Backup a sharded MongoDB cluster using KubeStash following the guide from [here](/docs/guides/mongodb/backup/kubestash/logical/sharding/index.md).


