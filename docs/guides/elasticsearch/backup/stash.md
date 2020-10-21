---
title: Backup & Restore Elasticsearch Using Stash
menu:
  docs_{{ .version }}:
    identifier: es-backup-stash
    name: Using Stash
    parent: es-backup
    weight: 5
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Backup & Restore Elasticsearch Using Stash

[Stash](https://stash.run/) by [AppsCode](https://appscode.com) is a Kubernetes operator for backup and recovery of Kubernetes stateful workloads. Stash v0.9.0+ supports backup and restoration of Elasticsearch databases. KubeDB v0.14.0+ comes with built-in support for Stash.

<figure align="center">
  <img alt="KubeDB + Stash" src="/docs/images/kubedb_plus_stash.svg">
<figcaption align="center">Fig: Backup KubeDB Databases Using Stash</figcaption>
</figure>

## How to use Stash

In order to backup Elasticsearch database using Stash, follow the following steps:

- **Install Stash Enterprise:** At first, you have to install Stash Enterprise Edition. Please, follow the steps from [here](https://stash.run/docs/latest/setup/install/enterprise/).

- **Install Elasticsearch Addon:** Then, you have to install Elasticsearch addon for Stash. Please, follow the steps from [here](https://stash.run/docs/latest/addons/elasticsearch/setup/install/).

- **Understand the Backup and Restore Flow:** Now, you can read about how Elasticsearch backup and restore works in Stash from [here](https://stash.run/docs/latest/addons/elasticsearch/overview/).

- **Get Started:** Finally, follow the guidelines of your desired database version to go through the steps of backup and restore process from [here](https://stash.run/docs/latest/addons/elasticsearch/).
