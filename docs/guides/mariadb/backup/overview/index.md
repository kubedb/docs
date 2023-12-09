---
title: Backup & Restore MariaDB Using Stash
menu:
  docs_{{ .version }}:
    identifier: guides-mariadb-backup-overview
    name: Overview
    parent: guides-mariadb-backup
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="Please install [Stash](https://stash.run/docs/latest/setup/install/enterprise/) to try this feature. Database backup with Stash is already included in the KubeDB license. So, you don't need a separate license for Stash." >}}

# MariaDB Backup & Restore Overview

KubeDB uses [Stash](https://stash.run) to backup and restore databases. Stash by AppsCode is a cloud native data backup and recovery solution for Kubernetes workloads. Stash utilizes [restic](https://github.com/restic/restic) to securely backup stateful applications to any cloud or on-prem storage backends (for example, S3, GCS, Azure Blob storage, Minio, NetApp, Dell EMC etc.).

<figure align="center">
  <img alt="KubeDB + Stash" src="/docs/images/kubedb_plus_stash.svg">
<figcaption align="center">Fig: Backup KubeDB Databases Using Stash</figcaption>
</figure>

# How Stash Backups & Restores MariaDB Database

Stash 0.9.0+ supports backup and restore operation of many databases. This guide will give you an overview of how MariaDB database backup and restore process works in Stash.

## Logical Backup

Stash supports taking [logical backup](https://mariadb.com/kb/en/backup-and-restore-overview/#logical-vs-physical-backups) of MariaDB databases using [mysqldump](https://mariadb.com/kb/en/mysqldump/). It is the most flexible way to perform a backup and restore, and a good choice when the data size is relatively small.

### How Logical Backup Works

The following diagram shows how Stash takes logical backup of a MariaDB database. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="MariaDB Backup Overview" src="/docs/guides/mariadb/backup/overview/images/mariadb-logical-backup.svg">
  <figcaption align="center">Fig: MariaDB Logical Backup Overview</figcaption>
</figure>

The backup process consists of the following steps:

1. At first, a user creates a secret with access credentials of the backend where the backed up data will be stored.

2. Then, she creates a `Repository` crd that specifies the backend information along with the secret that holds the credentials to access the backend.

3. Then, she creates a `BackupConfiguration` crd targeting the [AppBinding](/docs/guides/mariadb/concepts/appbinding/index.md) crd of the desired database. The `BackupConfiguration` object also specifies the `Task` to use to backup the database.

4. Stash operator watches for `BackupConfiguration` crd.

5. Once Stash operator finds a `BackupConfiguration` crd, it creates a CronJob with the schedule specified in `BackupConfiguration` object to trigger backup periodically.

6. On the next scheduled slot, the CronJob triggers a backup by creating a `BackupSession` crd.

7. Stash operator also watches for `BackupSession` crd.

8. When it finds a `BackupSession` object, it resolves the respective `Task` and `Function` and prepares a Job definition to backup.

9. Then, it creates the Job to backup the targeted database.

10. The backup Job reads necessary information to connect with the database from the `AppBinding` crd. It also reads backend information and access credentials from `Repository` crd and Storage Secret respectively.

11. Then, the Job dumps the targeted database and uploads the output to the backend. Stash pipes the output of dump command to uploading process. Hence, backup Job does not require a large volume to hold the entire dump output.

12. Finally, when the backup is complete, the Job sends Prometheus metrics to the Pushgateway running inside Stash operator pod. It also updates the `BackupSession` and `Repository` status to reflect the backup procedure.

### How Restore from Logical Backup Works

The following diagram shows how Stash restores a MariaDB database from a logical backup. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Database Restore Overview" src="/docs/guides/mariadb/backup/overview/images/mariadb-logical-restore.svg">
  <figcaption align="center">Fig: MariaDB Logical Restore Process Overview</figcaption>
</figure>

The restore process consists of the following steps:

1. At first, a user creates a `RestoreSession` crd targeting the `AppBinding` of the desired database where the backed up data will be restored. It also specifies the `Repository` crd which holds the backend information and the `Task` to use to restore the target.

2. Stash operator watches for `RestoreSession` object.

3. Once it finds a `RestoreSession` object, it resolves the respective `Task` and `Function` and prepares a Job definition to restore.

4. Then, it creates the Job to restore the target.

5. The Job reads necessary information to connect with the database from respective `AppBinding` crd. It also reads backend information and access credentials from `Repository` crd and Storage Secret respectively.

6. Then, the job downloads the backed up data from the backend and injects into the desired database. Stash pipes the downloaded data to the respective database tool to inject into the database. Hence, restore job does not require a large volume to download entire backup data inside it.

7. Finally, when the restore process is complete, the Job sends Prometheus metrics to the Pushgateway and update the `RestoreSession` status to reflect restore completion.

## Next Steps

- Backup a standalone MariaDB databases using Stash following the guide from [here](/docs/guides/mariadb/backup/logical/standalone/index.md).
- Backup a MariaDB cluster using Stash following the guide from [here](/docs/guides/mariadb/backup/logical/cluster/index.md).
- Configure a generic backup template for all the MariaDB databases of your cluster using Stash Auto-backup by following the guide from [here](/docs/guides/mariadb/backup/auto-backup/index.md).
- Customize the backup & restore process for your cluster by following the guides from [here](/docs/guides/mariadb/backup/customization/index.md).
