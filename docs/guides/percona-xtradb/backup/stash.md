---
title: Backup & Restore PerconaXtraDB Using Stash
menu:
  docs_{{ .version }}:
    identifier: px-backup-stash
    name: Using Stash
    parent: px-backup
    weight: 5
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

# Backup & Restore Percona XtraDB Using Stash

KubeDB uses [Stash](https://stash.run) to backup and restore databases. Stash by AppsCode is a cloud native data backup and recovery solution for Kubernetes workloads. Stash utilizes [restic](https://github.com/restic/restic) to securely backup stateful applications to any cloud or on-prem storage backends (for example, S3, GCS, Azure Blob storage, Minio, NetApp, Dell EMC etc.).

<figure align="center">
  <img alt="KubeDB + Stash" src="/docs/images/kubedb_plus_stash.svg">
<figcaption align="center">Fig: Backup KubeDB Databases Using Stash</figcaption>
</figure>

## How to use Stash

In order to backup PerconaXtraDB database using Stash, follow the following steps:

- **Install Stash Enterprise:** At first, you have to install Stash Enterprise Edition. Please, follow the steps from [here](https://stash.run/docs/latest/setup/install/enterprise/).

- **Install PerconaXtraDB Addon:** Then, you have to install PerconaXtraDB addon for Stash. Please, follow the steps from [here](https://stash.run/docs/latest/addons/percona-xtradb/setup/install/).

- **Understand the Backup and Restore Flow:** Now, you can read about how PerconaXtraDB backup and restore works in Stash from [here](https://stash.run/docs/latest/addons/percona-xtradb/overview/).

- **Get Started:** Finally, follow the step by step guideline to backup or restore your desired database version from [here](https://stash.run/docs/latest/addons/percona-xtradb/).
