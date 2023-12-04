---
title: MongoDB Backup & Restore Overview | KubeStash
menu:
  docs_{{ .version }}:
    identifier: guides-mongodb-backup-kubestash-overview
    name: Overview
    parent: guides-mongodb-backup-kubestash
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