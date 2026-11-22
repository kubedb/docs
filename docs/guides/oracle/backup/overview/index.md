---
title: Oracle Backup & Restore Overview
menu:
  docs_{{ .version }}:
    identifier: guides-oracle-backup-overview
    name: Overview
    parent: guides-oracle-backup
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="Please install [KubeStash](https://kubestash.com/docs/latest/setup/install/kubestash/) to try this feature. Database backup with KubeStash is already included in the KubeDB license. So, you don't need a separate license for KubeStash." >}}

# Oracle Backup & Restore Overview

KubeDB uses [KubeStash](https://kubestash.com) to backup and restore `Oracle` databases. KubeStash uses the `OracleBackup` driver, which runs Oracle's own backup tool, **RMAN** (Recovery Manager), inside the database pod. RMAN takes a physical backup of the database (datafiles, control file, SPFILE and archived redo logs) while the database stays online, and streams it straight to an S3 compatible bucket.

## Backup Types

KubeDB supports three kinds of Oracle backup. Pick the one that matches how much data you can afford to lose.

| Backup type | What is backed up | Restores to | Set up with | Guide |
|---|---|---|---|---|
| Base backup | A full physical backup of the whole database | The moment the backup finished | `BackupConfiguration` | [Base Backup](/docs/guides/oracle/backup/base-backup/index.md) |
| Incremental backup | A full (LEVEL 0) backup on a schedule, plus small LEVEL 1 backups and archived redo logs in between | Any point in time between backups | `OracleArchiver` | [Incremental Backup](/docs/guides/oracle/backup/incremental/index.md) |
| Point-in-time recovery | Uses the incremental backups above | Any second you choose, into a new database | `Oracle.spec.init.archiver` | [Point-in-time Recovery](/docs/guides/oracle/pitr/archiver.md) |

A few words used in these guides:

- **Base backup (full backup)**: a complete copy of the database. It can be restored on its own.
- **LEVEL 0**: an RMAN full backup that later incremental backups can build on. KubeDB takes every full backup of an `OracleArchiver` as a LEVEL 0.
- **LEVEL 1**: an RMAN incremental backup. It holds only the blocks changed since the LEVEL 0 (or since the previous LEVEL 1), so it is small and fast.
- **Archived redo logs**: Oracle writes every change into redo logs. Shipping these logs every minute is what lets KubeDB restore to any second, not just to the moment of a backup.
- **SCN**: the System Change Number, Oracle's internal clock. Every committed change has one.

## How Backup Works

The backup process consists of the following steps:

1. You create a `Secret` with the credentials of your S3 bucket, and a `BackupStorage` that points to the bucket.

2. You create a `BackupConfiguration` (base backup), or an `OracleArchiver` (incremental backup). For an `OracleArchiver`, the KubeDB operator creates the `BackupConfiguration` for you.

3. KubeStash creates a `Repository` and a `CronJob` for the schedule, and triggers the first backup right away.

4. For each backup, KubeStash creates a `BackupSession`, a `Snapshot`, and a backup `Job`.

5. Before the first backup of a database, the backup job prepares an **Oracle wallet**. It stores the bucket address in a `Secret` named `<database-name>-backup-wallet` and creates an `OracleOpsRequest`. KubeDB applies the wallet to the database, which restarts the database pod once. Later backups reuse the wallet.

6. The backup job runs RMAN inside the database pod. RMAN writes the backup pieces directly to the bucket through Oracle's S3 backup module. The data does not pass through the backup job.

7. When RMAN finishes, the job records the details of the backup (database ID, SCN, control file and SPFILE pieces) on the `Snapshot` and marks it `Succeeded`.

For an `OracleArchiver`, the KubeDB operator also runs a long-lived `Sidekick` pod named `<database-name>-sidekick`. Between full backups, it ships the archived redo logs every minute and takes a LEVEL 1 backup on a fixed interval. The [Incremental Backup](/docs/guides/oracle/backup/incremental/index.md) guide explains this loop.

## How Restore Works

The restore process consists of the following steps:

1. You create a `RestoreSession` that points to the target `Oracle` and to the `Repository` and `Snapshot` to restore from. For point-in-time recovery, you create a new `Oracle` with `spec.init.archiver`, and the KubeDB operator creates the `RestoreSession` for you.

2. KubeStash creates a restore `Job`.

3. The restore job prepares the Oracle wallet of the target database, the same way a backup does.

4. The job runs RMAN inside the target database pod. RMAN restores the datafiles, recovers them up to the chosen point, and opens the database with `RESETLOGS`.

5. The job marks the `RestoreSession` `Succeeded`.

## Things to Know

- **Storage:** Oracle backups are written by Oracle's S3 backup module, so the backend must be S3 or S3 compatible (for example MinIO). Put the scheme and port in the `BackupStorage` endpoint, for example `http://minio.demo.svc.cluster.local:80`.
- **Bucket prefix:** the `BackupStorage` prefix is used for KubeStash's own metadata only. RMAN writes its backup pieces at the root of the bucket. Use a bucket dedicated to Oracle backups.
- **One repository:** use exactly one repository per backup session.
- **ARCHIVELOG mode:** RMAN can only back up an online database in `ARCHIVELOG` mode. KubeDB creates Oracle databases in `ARCHIVELOG` mode.
- **Restore overwrites:** a restore replaces everything in the target database. Changes made after the restore point are lost.

## Next Steps

- Take a [base backup](/docs/guides/oracle/backup/base-backup/index.md) of an Oracle database and restore it.
- Set up [incremental backup](/docs/guides/oracle/backup/incremental/index.md) with an `OracleArchiver`.
- Restore an Oracle database to a [point in time](/docs/guides/oracle/pitr/archiver.md).
