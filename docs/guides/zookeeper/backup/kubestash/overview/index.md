---
title: Backup & Restore ZooKeeper Using KubeStash
menu:
  docs_{{ .version }}:
    identifier: guides-zk-backup-overview-stashv2
    name: Overview
    parent: guides-zk-backup-stashv2
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="Please install [KubeStash](https://kubestash.com/docs/latest/setup/install/kubestash/) to try this feature. Database backup with KubeStash is already included in the KubeDB license. So, you don't need a separate license for KubeStash." >}}

# ZooKeeper Backup & Restore Overview

KubeDB also uses [KubeStash](https://kubestash.com) to backup and restore databases. KubeStash by AppsCode is a cloud native data backup and recovery solution for Kubernetes workloads and databases. KubeStash utilizes [restic](https://github.com/restic/restic) to securely backup stateful applications to any cloud or on-prem storage backends (for example, S3, GCS, Azure Blob storage, Minio, NetApp, Dell EMC etc.).

<figure align="center">
  <img alt="KubeDB + KubeStash" src="/docs/guides/zookeeper/backup/kubestash/overview/images/kubedb_plus_kubestash.svg">
<figcaption align="center">Fig: Backup KubeDB Databases Using KubeStash</figcaption>
</figure>

## How Backup Works

The following diagram shows how KubeStash takes backup of a `ZooKeeper` database. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="ZooKeeper Backup Overview" src="/docs/guides/zookeeper/backup/kubestash/overview/images/backup_overview.svg">
  <figcaption align="center">Fig: ZooKeeper Backup Overview</figcaption>
</figure>

The backup process consists of the following steps:

1. At first, a user creates a `Secret`. This secret holds the credentials to access the backend where the backed up data will be stored.

2. Then, he creates a `BackupStorage` custom resource that specifies the backend information, along with the `Secret` containing the credentials needed to access the backend.

3. KubeStash operator watches for `BackupStorage` custom resources. When it finds a `BackupStorage` object, it initializes the `BackupStorage` by uploading the `metadata.yaml` file to the specified backend.

4. Next, he creates a `BackupConfiguration` custom resource that specifies the target database, addon information (including backup tasks), backup schedules, storage backends for storing the backup data, and other additional settings.

5. KubeStash operator watches for `BackupConfiguration` objects.

6. Once the KubeStash operator finds a `BackupConfiguration` object, it creates `Repository` with the information specified in the `BackupConfiguration`.

7. KubeStash operator watches for `Repository` custom resources. When it finds the `Repository` object, it Initializes `Repository` by uploading `repository.yaml` file into the `spec.sessions[*].repositories[*].directory` path specified in `BackupConfiguration`.

8. Then, it creates a `CronJob` for each session with the schedule specified in `BackupConfiguration` to trigger backup periodically.

9. KubeStash operator triggers an instant backup as soon as the `BackupConfiguration` is ready. Backups are otherwise triggered by the `CronJob` based on the specified schedule.

10. KubeStash operator watches for `BackupSession` custom resources.

11. When it finds a `BackupSession` object, it creates a `Snapshot` custom resource for each `Repository` specified in the `BackupConfiguration`.

12. Then it resolves the respective `Addon` and `Function` and prepares backup `Job` definition.

13. Then, it creates the `Job` to backup the targeted `ZooKeeper` database.

14. The backup `Job` reads necessary information (e.g. auth secret, port)  to connect with the database from the `AppBinding` CR. It also reads backend information and access credentials from `BackupStorage` CR, Storage Secret and `Repository` path respectively.

15. Then, the `Job` dumps the targeted `ZooKeeper` database and uploads the output to the backend. KubeStash pipes the output of dump command to uploading process. Hence, backup `Job` does not require a large volume to hold the entire dump output.

16. After the backup process is completed, the backup `Job` updates the `status.components[dump]` field of the `Snapshot` resources with backup information of the target `ZooKeeper` database.

## How Restore Process Works

The following diagram shows how KubeStash restores backed up data into a `PostgreSQL` database. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Database Restore Overview" src="/docs/guides/zookeeper/backup/kubestash/overview/images/restore_overview.svg">
  <figcaption align="center">Fig: ZooKeeper Restore Process Overview</figcaption>
</figure>

The restore process consists of the following steps:

1. At first, a user creates a `ZooKeeper` database where the data will be restored or the user can use the same `ZooKeeper` database.

2. Then, he creates a `RestoreSession` custom resource that specifies the target database where the backed-up data will be restored, addon information (including restore tasks), the target snapshot to be restored, the Repository containing that snapshot, and other additional settings.

3. KubeStash operator watches for `RestoreSession` custom resources.

4. When it finds a `RestoreSession` custom resource, it resolves the respective `Addon` and `Function` and prepares a restore `Job` definition.

5. Then, it creates the `Job` to restore the target.

6. The `Job` reads necessary information to connect with the database from respective `AppBinding` CR. It also reads backend information and access credentials from `Repository` CR and storage `Secret` respectively.

7. Then, the `Job` downloads the backed up data from the backend and injects into the desired database. KubeStash pipes the downloaded data to the respective database tool to inject into the database. Hence, restore `Job` does not require a large volume to download entire backup data inside it.

8. Finally, when the restore process is completed, the `Job` updates the `status.components[*]` field of the `RestoreSession` with restore information of the target database.

## Next Steps

- Backup a `ZooKeeper` database using KubeStash by the following guides from [here](/docs/guides/zookeeper/backup/kubestash/logical/index.md).