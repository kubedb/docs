---
title: Use Stash to Backup MongoDB
menu:
  docs_{{ .version }}:
    identifier: mg-stash-backup
    name: Using Stash
    parent: mg-snapshot-mongodb
    weight: 5
menu_name: docs_{{ .version }}
section_menu_id: guides
---

# Use Stash to Backup MongoDB

[Stash](https://appscode.com/products/stash) by [AppsCode](https://appscode.com) is a Kubernetes operator for backup and recovery of Kubernetes stateful workloads. Stash v0.9.0+ supports backup and restoration of MongoDB databases. KubeDB v0.13.0+ comes with built-in support for Stash. We recommend to use Stash for backup and restore your MongoDB databases instead of KubeDB's native method.

This guide will give you an overview of why you should use Stash to backup and restore your MongoDB databases.

## Why use Stash

As a dedicated backup and recovery tool, Stash has the following key benefits:

- **Automatic Cleanup:** Stash lets you provide a retention policy for the snapshots. So, you don't have to worry about running out of space for your backed up data. Stash will automatically delete the old snapshots according to the retention policy.

- **Deduplication:** Stash does not upload the entire targeted data on each backup. Instead, it uploads only the changes since the last backup. This reduces network bandwidth usage and backup time.

- **Encryption:** Stash keeps all the data in backend encrypted. Hence, your data is safe even if your backend gets compromised.

- **Instant Backup:** Stash lets you trigger a backup instantly. This is particularly useful when you want to perform some experimental operations on your database and you want to make sure that you have backed up the current state of your database.

- **Auto Backup:** You can also configure a common backup template for your databases. Then, you can enable backup for a particular database by adding some annotations to the respective `AppBinding` crd. In Stash parlance, we call it **Auto-Backup**.

- **Rich Prometheus Metrics:** Stash provides rich Prometheus metrics for both backup and restore processes. So, you can always keep an eye on the backup process and configure an alert in case something goes wrong.

- **Independent of Database Life Cycle:** You can enable or disable backup for your databases without interrupting your services.

- **Customizability:** Stash gives you the ability to customize the backup process. You can pass various arguments to the backup and restore command. You can also create your own backup or restore flow through addon mechanism.

## How to use Stash

In order to backup MongoDB database using Stash, follow the following steps:

- **Install Stash:** At first, you have to install Stash. Please, follow the steps from [here](https://stash.run/docs/latest/setup/).

- **Install MongoDB Addon:** Then, you have to install MongoDB addon for Stash. Please, follow the steps from [here](https://appscode.com/products/stash/latest/addons/mongodb/setup/install/).

- **Understand the Backup and Restore Flow:** Now, you can read the following guide from [here](https://appscode.com/products/stash/latest/addons/mongodb/overview/) to understand how backup and restore of a MongoDB database works in Stash.

- **Get Started:** Now, follow the guidelines of your desired database version to go through the steps of backup and restore process from [here](https://appscode.com/products/stash/latest/addons/mongodb/).
