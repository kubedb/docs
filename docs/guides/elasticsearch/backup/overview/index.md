---
title: Backup & Restore Elasticsearch Using Stash
menu:
  docs_{{ .version }}:
    identifier: guides-es-backup-overview
    name: Overview
    parent: guides-es-backup
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="Please install [Stash](https://stash.run/docs/latest/setup/install/stash/) to try this feature. Database backup with Stash is already included in the KubeDB license. So, you don't need a separate license for Stash." >}}

# Backup & Restore Elasticsearch Using Stash

KubeDB uses [Stash](https://stash.run) to backup and restore databases. Stash by AppsCode is a cloud native data backup and recovery solution for Kubernetes workloads. Stash utilizes [restic](https://github.com/restic/restic) to securely backup stateful applications to any cloud or on-prem storage backends (for example, S3, GCS, Azure Blob storage, Minio, NetApp, Dell EMC etc.).

<figure align="center">
  <img alt="KubeDB + Stash" src="/docs/images/kubedb_plus_stash.svg">
<figcaption align="center">Fig: Backup KubeDB Databases Using Stash</figcaption>
</figure>

## How Backup Works

The following diagram shows how Stash takes a backup of an Elasticsearch database. Open the image in a new tab to see the enlarged version.

<figure align="center">
 <img alt="Elasticsearch Backup Overview" src="/docs/guides/elasticsearch/backup/overview/images/backup_overview.svg">
  <figcaption align="center">Fig: Elasticsearch Backup Overview</figcaption>
</figure>

The backup process consists of the following steps:

1. At first, a user creates a secret with access credentials of the backend where the backed up data will be stored.

2. Then, she creates a `Repository` crd that specifies the backend information along with the secret that holds the credentials to access the backend.

3. Then, she creates a `BackupConfiguration` crd targeting the [AppBinding](/docs/guides/elasticsearch/concepts/appbinding/index.md) crd of the desired database. The `BackupConfiguration` object also specifies the `Task` to use to backup the database.

4. Stash operator watches for `BackupConfiguration` crd.

5. Once Stash operator finds a `BackupConfiguration` crd, it creates a CronJob with the schedule specified in `BackupConfiguration` object to trigger backup periodically.

6. On the next scheduled slot, the CronJob triggers a backup by creating a `BackupSession` crd.

7. Stash operator also watches for `BackupSession` crd.

8. When it finds a `BackupSession` object, it resolves the respective `Task` and `Function` and prepares a Job definition to backup.

9. Then, it creates the Job to backup the targeted database.

10. The backup Job reads necessary information to connect with the database from the `AppBinding` crd. It also reads backend information and access credentials from `Repository` crd and Storage Secret respectively.

11. Then, the Job dumps the targeted database and uploads the output to the backend. Stash stores the dumped files temporarily before uploading into the backend. Hence, you should provide a PVC template using `spec.interimVolumeTemplate` field of `BackupConfiguration` crd to use to store those dumped files temporarily.

12. Finally, when the backup is completed, the Job sends Prometheus metrics to the Pushgateway running inside Stash operator pod. It also updates the `BackupSession` and `Repository` status to reflect the backup procedure.

## How Restore Process Works

The following diagram shows how Stash restores backed up data into an Elasticsearch database. Open the image in a new tab to see the enlarged version.

<figure align="center">
 <img alt="Database Restore Overview" src="/docs/guides/elasticsearch/backup/overview/images/restore_overview.svg">
  <figcaption align="center">Fig: Elasticsearch Restore Process</figcaption>
</figure>

The restore process consists of the following steps:

1. At first, a user creates a `RestoreSession` crd targeting the `AppBinding` of the desired database where the backed up data will be restored. It also specifies the `Repository` crd which holds the backend information and the `Task` to use to restore the target.

2. Stash operator watches for `RestoreSession` object.

3. Once it finds a `RestoreSession` object, it resolves the respective `Task` and `Function` and prepares a Job definition to restore.

4. Then, it creates the Job to restore the target.

5. The Job reads necessary information to connect with the database from respective `AppBinding` crd. It also reads backend information and access credentials from `Repository` crd and Storage Secret respectively.

6. Then, the job downloads the backed up data from the backend and insert into the desired database. Stash stores the downloaded files temporarily before inserting into the targeted database. Hence, you should provide a PVC template using `spec.interimVolumeTemplate` field of `RestoreSession` crd to use to store those restored files temporarily.

7. Finally, when the restore process is completed, the Job sends Prometheus metrics to the Pushgateway and update the `RestoreSession` status to reflect restore completion.

## Next Steps

- Backup your Elasticsearch databases using Stash following the guide from [here](/docs/guides/elasticsearch/backup/kubedb/index.md).
- Configure a generic backup template for all the Elasticsearch databases of your cluster using Stash Auto-backup by following the guide from [here](/docs/guides/elasticsearch/backup/auto-backup/index.md).
- Customize the backup & restore process for your cluster by following the guides from [here](/docs/guides/elasticsearch/backup/customization/index.md).
