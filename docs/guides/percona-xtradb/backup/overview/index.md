---
title: Backup & Restore PerconaXtraDB Using Stash
menu:
  docs_{{ .version }}:
    identifier: guides-px-backup-overview
    name: Overview
    parent: guides-px-backup
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [Stash Enterprise Edition](https://stash.run/docs/latest/setup/install/enterprise/) to try this feature. You can use KubeDB Enterprise license to install Stash Enterprise edition. Database backup with Stash is already included in the KubeDB Enterprise license. So, you don't need a separate license for Stash." >}}

# Percona XtraDB Backup & Restore Overview

KubeDB uses [Stash](https://stash.run) to backup and restore databases. Stash by AppsCode is a cloud native data backup and recovery solution for Kubernetes workloads. Stash utilizes [restic](https://github.com/restic/restic) to securely backup stateful applications to any cloud or on-prem storage backends (for example, S3, GCS, Azure Blob storage, Minio, NetApp, Dell EMC etc.).

<figure align="center">
  <img alt="KubeDB + Stash" src="/docs/images/kubedb_plus_stash.svg">
<figcaption align="center">Fig: Backup KubeDB Databases Using Stash</figcaption>
</figure>

## How Backup Works

The following diagram shows how Stash takes backup of a Percona XtraDB database. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Percona XtraDB Backup Overview" src="/docs/guides/percona-xtradb/backup/overview/images/backup_overview.svg">
  <figcaption align="center">Fig: Percona XtraDB Backup Overview</figcaption>
</figure>

The backup process consists of the following steps:

1. At first, a user creates a secret with access credentials of the backend where the backed up data will be stored.

2. Then, she creates a `Repository` crd that specifies the backend information along with the secret that holds the credentials to access the backend.

3. Then, she creates a `BackupConfiguration` crd targeting the [AppBinding](/docs/guides/percona-xtradb/concepts/appbinding.md) crd of the desired database. The `BackupConfiguration` object also specifies the `Task` to use to backup the database.

4. Stash operator watches for `BackupConfiguration` crd.

5. Once Stash operator finds a `BackupConfiguration` crd, it creates a CronJob with the schedule specified in `BackupConfiguration` object to trigger backup periodically.

6. On the next scheduled slot, the CronJob triggers a backup by creating a `BackupSession` crd.

7. Stash operator also watches for `BackupSession` crd.

8. When it finds a `BackupSession` object, it resolves the respective `Task` and `Function` and prepares a Job definition to backup.

9. Then, it creates the Job to backup the targeted database.

10. The backup Job reads necessary information to connect with the database from the `AppBinding` crd. It also reads backend information and access credentials from `Repository` crd and Storage Secret respectively.

11. Then, the Job dumps the targeted database(s) and uploads the output to the backend. Stash pipes the output of the dump command to the upload process. Hence, backup Job does not require a large volume to hold the entire dump output.

12. Finally, when the backup is complete, the Job sends Prometheus metrics to the Pushgateway running inside Stash operator pod. It also updates the `BackupSession` and `Repository` status to reflect the backup procedure.

### Backup Different Percona XtraDB Configurations

This section will show you how backup works for different Percona XtraDB Configurations.

#### Standalone Percona XtraDB

For a standalone Percona XtraDB database, the backup job directly dumps the database using `mysqldump` and pipe the output to the backup process.

<figure align="center">
 <img alt="Standalone Percona XtraDB Backup Overview" src="/docs/guides/percona-xtradb/backup/overview/images/standalone_backup.png">
  <figcaption align="center">Fig: Standalone Percona XtraDB Backup</figcaption>
</figure>

#### Percona XtraDB Cluster

For a standalone Percona XtraDB database, the backup Job runs the backup procedure to take the backup of the targeted databases and uploads the output to the backend. In backup procedure, the Job runs a process called `garbd` ([Galera Arbitrator](https://galeracluster.com/library/documentation/arbitrator.html)) which uses `xtrabackup-v2` script during State Snapshot Transfer (SST). Basically this Job takes a full copy of the data stored in  the data directory (`/var/lib/mysql`) and pipes the output of the backup procedure to the uploading process. Hence, backup Job does not require a large volume to hold the entire backed up data.

<figure align="center">
 <img alt="Percona XtraDB Cluster Backup Overview" src="/docs/guides/percona-xtradb/backup/overview/images/cluster_backup.png">
  <figcaption align="center">Fig: Percona XtraDB Cluster Backup</figcaption>
</figure>

## How Restore Process Works

The following diagram shows how Stash restores backed up data into a Percona XtraDB database. Open the image in a new tab to see the enlarged version.

<figure align="center">
  <img alt="Percona XtraDB Restore Overview" src="/docs/guides/percona-xtradb/backup/overview/images/restore_overview.svg">
  <figcaption align="center">Fig: Percona XtraDB Restore Process Overview</figcaption>
</figure>

The restore process consists of the following steps:

1. At first, a user creates a `RestoreSession` crd targeting the `AppBinding` of the desired database where the backed up data will be restored. It also specifies the `Repository` crd which holds the backend information and the `Task` to use to restore the target.

2. Stash operator watches for `RestoreSession` object.

3. Once it finds a `RestoreSession` object, it resolves the respective `Task` and `Function` and prepares a Job (in case of restoring cluster more than one Job and PVC) definition(s) to restore.

4. Then, it creates the Job(s) (as well as PVCs in case of cluster) to restore the target.

5. The Job(s) reads necessary information to connect with the database from respective `AppBinding` crd. It also reads backend information and access credentials from `Repository` crd and Storage Secret respectively.

6. Then, the Job(s) downloads the backed up data from the backend and injects into the desired database. Stash pipes the downloaded data to inject into the database. Hence, the restore Job(s) does not require a large volume to download entire backup data inside it.

7. Finally, when the restore process is complete, the Job(s) sends Prometheus metrics to the Pushgateway and update the `RestoreSession` status to reflect restore completion.

### Restore Different Percona XtraDB Configurations

This section will show you how restore works for different Percona XtraDB Configurations.

#### Standalone Percona XtraDB

For a standalone Percona XtraDB database, the restore Job downloads the backed up data from the backend and pipe the downloaded data to `mysql` command which inserts the data into the desired database.

<figure align="center">
 <img alt="Standalone Percona XtraDB Restore Overview" src="/docs/guides/percona-xtradb/backup/overview/images/standalone_restore.png">
  <figcaption align="center">Fig: Standalone Percona XtraDB Restore</figcaption>
</figure>

#### Percona XtraDB Cluster

For a Percona XtraDB Cluster, the Stash operator creates a number (equal to the value of `.spec.target.replicas` of `RestoreSession` object) of Jobs to restore. Each of these Jobs requires a PVC to store the previously backed up data of the data directory `/var/lib/mysql` from the backend. Then each Job downloads the backed up data from the backend and injects into the associated PVC.

<figure align="center">
 <img alt="Percona XtraDB Cluster Restore Overview" src="/docs/guides/percona-xtradb/backup/overview/images/cluster_restore.png">
  <figcaption align="center">Fig: Percona XtraDB Cluster Restore</figcaption>
</figure>

## Next Steps

- Backup a standalone Precona XtraDB server using Stash by following the guides from [here](/docs/guides/percona-xtradb/backup/standalone/index.md).
- Backup a Precona XtraDB cluster using Stash by following the guides from [here](/docs/guides/percona-xtradb/backup/cluster/index.md).
