---
title: PerconaXtraDB Galera Cluster Overview
menu:
  docs_{{ .version }}:
    identifier: guides-perconaxtradb-clustering-overview
    name: PerconaXtraDB Galera Cluster Overview
    parent: guides-perconaxtradb-clustering
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# PerconaXtraDB Galera Cluster

Here we'll discuss some concepts about PerconaXtraDB Galera Cluster.

## So What is Replication

Replication means data being copied from one PerconaXtraDB server to one or more other PerconaXtraDB servers, instead of only stored in one server. One can read or write in any server of the cluster. The following figure shows a cluster of four PerconaXtraDB servers:

![PerconaXtraDB Cluster](/docs/guides/perconaxtradb/clustering/overview/images/galera_small.png)

Image ref: <https://perconaxtradb.com/kb/en/what-is-perconaxtradb-galera-cluster/+image/galera_small>

## Galera Replication

PerconaXtraDB Galera Cluster is a [virtually synchronous](https://perconaxtradb.com/kb/en/about-galera-replication/#synchronous-vs-asynchronous-replication) multi-master cluster for PerconaXtraDB. The Server replicates a transaction at commit time by broadcasting the write set associated with the transaction to every node in the cluster. The client connects directly to the DBMS and experiences behavior that is similar to native PerconaXtraDB in most cases. The wsrep API (write set replication API) defines the interface between Galera replication and PerconaXtraDB.

Ref: [About Galera Replication](https://perconaxtradb.com/kb/en/about-galera-replication/)

## PerconaXtraDB Galera Cluster Features

- Virtually synchronous replication
- Active-active multi-master topology
- Read and write to any cluster node
- Automatic membership control, failed nodes drop from the cluster
- Automatic node joining
- True parallel replication, on row level
- Direct client connections, native PerconaXtraDB look & feel

Ref: [What is PerconaXtraDB Galera Cluster?](https://perconaxtradb.com/kb/en/what-is-perconaxtradb-galera-cluster/#features)

### Limitations

There are some limitations in PerconaXtraDB Galera Cluster that are listed [here](https://perconaxtradb.com/kb/en/perconaxtradb-galera-cluster-known-limitations/).

## Next Steps

- [Deploy PerconaXtraDB Galera Cluster](/docs/guides/perconaxtradb/clustering/galera-cluster) using KubeDB.
