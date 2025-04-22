---
title: MariaDB Galera Cluster Overview
menu:
  docs_{{ .version }}:
    identifier: guides-mariadb-clustering-overview
    name: MariaDB Clustering Overview
    parent: guides-mariadb-clustering
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# MariaDB Clustering

KubeDB currently supports two cluster modes: Multi-Master MariaDB Galera Cluster and Master-Slave MariaDB Standard Replication.

Here we'll discuss some concepts about Cluster.

## So What is Replication

Replication involves copying data from one MariaDB server to one or more other MariaDB servers, rather than storing it on a single server. The replication mode determines read and write capabilities. In a multi-master setup, you can perform read and write operations on any server. In a master and slave architecture, read and write operations are supported on the master host, while slave hosts support only read operations.

The following figure shows a cluster of four MariaDB servers of multi-master replication:

![MariaDB Cluster](/docs/guides/mariadb/clustering/overview/images/galera_small.png)

Image ref: <https://mariadb.com/kb/en/what-is-mariadb-galera-cluster/+image/galera_small>

## Galera Replication

MariaDB Galera Cluster is a [virtually synchronous](https://mariadb.com/kb/en/about-galera-replication/#synchronous-vs-asynchronous-replication) multi-master cluster for MariaDB. The Server replicates a transaction at commit time by broadcasting the write set associated with the transaction to every node in the cluster. The client connects directly to the DBMS and experiences behavior that is similar to native MariaDB in most cases. The wsrep API (write set replication API) defines the interface between Galera replication and MariaDB.

Ref: [About Galera Replication](https://mariadb.com/kb/en/about-galera-replication/)

## MariaDB Galera Cluster Features

- Virtually synchronous replication
- Active-active multi-master topology
- Read and write to any cluster node
- Automatic membership control, failed nodes drop from the cluster
- Automatic node joining
- True parallel replication, on row level
- Direct client connections, native MariaDB look & feel

Ref: [What is MariaDB Galera Cluster?](https://mariadb.com/kb/en/what-is-mariadb-galera-cluster/#features)

### Limitations

There are some limitations in MariaDB Galera Cluster that are listed [here](https://mariadb.com/kb/en/mariadb-galera-cluster-known-limitations/).

## MariaDB Standard Replication

MariaDB Standard Replication is a widely used mechanism for copying data from one MariaDB server (the master) to one or more other MariaDB servers (the slaves). This replication mode ensures data redundancy, enhances availability, and supports read scalability by distributing read operations across multiple servers. It operates in a single-master and slave architecture, where the master server handles both read and write operations, while slave servers are limited to read operations.

Ref: [About MariadDB Standard Replication](https://mariadb.com/kb/en/replication-overview/#standard-replication)

## MariaDB Standard Replication Cluster Features

- Asynchronous Replication
- Active-passive single-master topology
- Read and write to master, and read only to slave node
- Automatic node joining
- Automatic failover using maxscale proxy
- Binary Log-Based replication
- Direct client connections, native MariaDB look & feel
- Load Balance using Maxscale Proxy Server
- Read Write Split using maxscale Proxy Server

Ref: [What is MariaDB Standard Replication Cluster?](https://mariadb.com/kb/en/what-is-mariadb-galera-cluster/#features)

### Limitations

There are some limitations in MariaDB Galera Cluster that are listed [here](https://mariadb.com/kb/en/mariadb-galera-cluster-known-limitations/).


## Next Steps

- [Deploy MariaDB Galera Cluster](/docs/guides/mariadb/clustering/galera-cluster) using KubeDB.
