---
title: MongoDB Backup & Restore Overview
menu:
  docs_{{ .version }}:
    identifier: guides-mongodb-backup-overview
    name: Overview
    parent: guides-mongodb-backup
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="Please install [Stash](https://stash.run/docs/latest/setup/install/enterprise/) to try this feature. Database backup with Stash is already included in the KubeDB license. So, you don't need a separate license for Stash." >}}


# MongoDB Backup & Restore Overview

KubeDB uses [Stash](https://stash.run) to backup and restore databases. Stash by AppsCode is a cloud native data backup and recovery solution for Kubernetes workloads. Stash utilizes [restic](https://github.com/restic/restic) to securely backup stateful applications to any cloud or on-prem storage backends (for example, S3, GCS, Azure Blob storage, Minio, NetApp, Dell EMC etc.).

<figure align="center">
  <img alt="KubeDB + Stash" src="/docs/images/kubedb_plus_stash.svg">
<figcaption align="center">Fig: Backup KubeDB Databases Using Stash</figcaption>
</figure>

## How Backup Works

The following diagram shows how Stash takes backup of a MongoDB database. Open the image in a new tab to see the enlarged version.

<figure align="center">
 <img alt="MongoDB Backup Overview" src="/docs/guides/mongodb/backup/overview/images/backup_overview.svg">
  <figcaption align="center">Fig: MongoDB Backup Overview</figcaption>
</figure>

The backup process consists of the following steps:

1. At first, a user creates a secret with access credentials of the backend where the backed up data will be stored.

2. Then, she creates a `Repository` crd that specifies the backend information along with the secret that holds the credentials to access the backend.

3. Then, she creates a `BackupConfiguration` crd targeting the [AppBinding](/docs/guides/mongodb/concepts/appbinding.md) crd of the desired database. The `BackupConfiguration` object also specifies the `Task` to use to backup the database.

4. Stash operator watches for `BackupConfiguration` crd.

5. Once Stash operator finds a `BackupConfiguration` crd, it creates a CronJob with the schedule specified in `BackupConfiguration` object to trigger backup periodically.

6. On the next scheduled slot, the CronJob triggers a backup by creating a `BackupSession` crd.

7. Stash operator also watches for `BackupSession` crd.

8. When it finds a `BackupSession` object, it resolves the respective `Task` and `Function` and prepares a Job definition to backup.

9. Then, it creates the Job to backup the targeted database.

10. The backup Job reads necessary information to connect with the database from the `AppBinding` crd. It also reads backend information and access credentials from `Repository` crd and Storage Secret respectively.

11. Then, the Job dumps the targeted database and uploads the output to the backend. Stash pipes the output of dump command to uploading process. Hence, backup Job does not require a large volume to hold the entire dump output.

12. Finally, when the backup is complete, the Job sends Prometheus metrics to the Pushgateway running inside Stash operator pod. It also updates the `BackupSession` and `Repository` status to reflect the backup procedure.

### Backup Different MongoDB Configurations

This section will show you how backup works for different MongoDB configurations.

#### Standalone MongoDB

For a standalone MongoDB database, the backup job directly dumps the database using `mongodump` and pipe the output to the backup process.

<figure align="center">
 <img alt="Standalone MongoDB Backup Overview" src="/docs/guides/mongodb/backup/overview/images/standalone_backup.svg">
  <figcaption align="center">Fig: Standalone MongoDB Backup</figcaption>
</figure>

#### MongoDB ReplicaSet Cluster

For MongoDB ReplicaSet cluster, Stash takes backup from one of the secondary replicas. The backup process consists of the following steps:

1. Identify a secondary replica.
2. Lock the secondary replica.
3. Backup the secondary replica.
4. Unlock the secondary replica.

<figure align="center">
 <img alt="MongoDB ReplicaSet Cluster Backup Overview" src="/docs/guides/mongodb/backup/overview/images/replicaset_backup.svg">
  <figcaption align="center">Fig: MongoDB ReplicaSet Cluster Backup</figcaption>
</figure>

#### MongoDB Sharded Cluster

For MongoDB sharded cluster, Stash takes backup of the individual shards as well as the config server. Stash takes backup from a secondary replica of the shards and the config server. If there is no secondary replica then Stash will take backup from the primary replica. The backup process consists of the following steps:

1. Disable balancer.
2. Lock config server.
3. Identify a secondary replica for each shard.
4. Lock the secondary replica.
5. Run backup on the secondary replica.
6. Unlock the secondary replica.
7. Unlock config server.
8. Enable balancer.

<figure align="center">
 <img alt="MongoDB Sharded Cluster Backup Overview" src="/docs/guides/mongodb/backup/overview/images/sharded_backup.svg">
  <figcaption align="center">Fig: MongoDB Sharded Cluster Backup</figcaption>
</figure>

## How Restore Process Works

The following diagram shows how Stash restores backed up data into a MongoDB database. Open the image in a new tab to see the enlarged version.

<figure align="center">
 <img alt="Database Restore Overview" src="/docs/guides/mongodb/backup/overview/images/restore_overview.svg">
  <figcaption align="center">Fig: MongoDB Restore Process Overview</figcaption>
</figure>

The restore process consists of the following steps:

1. At first, a user creates a `RestoreSession` crd targeting the `AppBinding` of the desired database where the backed up data will be restored. It also specifies the `Repository` crd which holds the backend information and the `Task` to use to restore the target.

2. Stash operator watches for `RestoreSession` object.

3. Once it finds a `RestoreSession` object, it resolves the respective `Task` and `Function` and prepares a Job definition to restore.

4. Then, it creates the Job to restore the target.

5. The Job reads necessary information to connect with the database from respective `AppBinding` crd. It also reads backend information and access credentials from `Repository` crd and Storage Secret respectively.

6. Then, the job downloads the backed up data from the backend and injects into the desired database. Stash pipes the downloaded data to the respective database tool to inject into the database. Hence, restore job does not require a large volume to download entire backup data inside it.

7. Finally, when the restore process is complete, the Job sends Prometheus metrics to the Pushgateway and update the `RestoreSession` status to reflect restore completion.

### Restoring Different MongoDB Configurations

This section will show you restore process works for different MongoDB configurations.

#### Standalone MongoDB

For a standalone MongoDB database, the restore job downloads the backed up data from the backend and pipe the downloaded data to `mongorestore` command which inserts the data into the desired MongoDB database.

<figure align="center">
 <img alt="Standalone MongoDB Restore Overview" src="/docs/guides/mongodb/backup/overview/images/standalone_restore.svg">
  <figcaption align="center">Fig: Standalone MongoDB Restore</figcaption>
</figure>

#### MongoDB ReplicaSet Cluster

For MongoDB ReplicaSet cluster, Stash identifies the primary replica and restore into it.

<figure align="center">
 <img alt="MongoDB ReplicaSet Cluster Restore Overview" src="/docs/guides/mongodb/backup/overview/images/replicaset_restore.svg">
  <figcaption align="center">Fig: MongoDB ReplicaSet Cluster Restore</figcaption>
</figure>

#### MongoDB Sharded Cluster

For MongoDB sharded cluster, Stash identifies the primary replica of each shard as well as the config server and restore respective backed up data into them.

<figure align="center">
 <img alt="MongoDB Sharded Cluster Restore" src="/docs/guides/mongodb/backup/overview/images/sharded_restore.svg">
  <figcaption align="center">Fig: MongoDB Sharded Cluster Restore</figcaption>
</figure>

## Next Steps

- Backup a standalone MongoDB databases using Stash following the guide from [here](/docs/guides/mongodb/backup/logical/standalone/index.md).
- Backup a MongoDB Replicaset cluster using Stash following the guide from [here](/docs/guides/mongodb/backup/logical/replicaset/index.md).
- Backup a sharded MongoDB cluster using Stash following the guide from [here](/docs/guides/mongodb/backup/logical/sharding/index.md).
