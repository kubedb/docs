---
title: Weaviate Backup Overview
menu:
  docs_{{ .version }}:
    identifier: weaviate-backup-overview
    name: Overview
    parent: weaviate-backup
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Weaviate Backup Overview

This guide will give an overview of how KubeDB supports backup and restore for `Weaviate` databases using [KubeStash](https://kubestash.com).

## Before You Begin

- You should be familiar with the following `KubeDB` concepts:
  - [Weaviate](/docs/guides/weaviate/concepts/weaviate.md)
- You should be familiar with the following `KubeStash` concepts:
  - [BackupStorage](https://kubestash.com/docs/latest/concepts/crds/backupstorage/)
  - [BackupConfiguration](https://kubestash.com/docs/latest/concepts/crds/backupconfiguration/)
  - [BackupSession](https://kubestash.com/docs/latest/concepts/crds/backupsession/)
  - [RestoreSession](https://kubestash.com/docs/latest/concepts/crds/restoresession/)
  - [RetentionPolicy](https://kubestash.com/docs/latest/concepts/crds/retentionpolicy/)

## How Backup Works

KubeStash uses a sidecar-based approach to backup Weaviate databases. The backup process consists of the following steps:

1. At first, a user creates a `BackupStorage` CR that defines the backend storage location (e.g., S3, GCS, Azure Blob).

2. Then, the user creates a `RetentionPolicy` CR that defines how long backup snapshots will be retained.

3. Then, the user creates a `BackupConfiguration` CR that references the target `Weaviate` database, the `BackupStorage`, and the `RetentionPolicy`. A backup schedule (cron expression) can be defined.

4. When a `BackupConfiguration` CR is created, KubeStash creates a `CronJob` to trigger backup sessions at the scheduled time.

5. On each scheduled time, a `BackupSession` CR is created. KubeStash executes the backup in a temporary job that connects to the Weaviate database and writes a snapshot to the backend storage.

6. The backup snapshot is stored in the backend storage and a `Snapshot` CR is created to track the backup metadata.

## How Restore Works

The restore process consists of the following steps:

1. At first, the user creates a target `Weaviate` database (or uses an existing one).

2. Then, the user creates a `RestoreSession` CR referencing the `Snapshot` to restore and the target `Weaviate` database.

3. KubeStash executes the restore in a temporary job that reads the snapshot from the backend storage and restores the data to the target Weaviate database.

4. After the restore completes, the `RestoreSession` status transitions to `Succeeded`.

In the next docs, we are going to show step-by-step guides on backup and restore of Weaviate databases using KubeStash.
