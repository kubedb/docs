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

KubeDB supports two primary clustering modes for MariaDB: Multi-Master MariaDB Galera Cluster and Master-Slave MariaDB Standard Replication. Below, we explore key concepts of these clustering approaches, highlighting their mechanisms, benefits, and use cases in a distributed database environment.

Below, we explore key concepts of MariaDB clustering, focusing on the two clustering modes supported by KubeDB: Multi-Master MariaDB Galera Cluster and Master-Slave MariaDB Standard Replication

## What is Replication

Replication in MariaDB involves duplicating data from one MariaDB server (the source) to one or more other MariaDB servers (replicas), rather than storing it on a single server. This process enhances data availability, scalability, and fault tolerance. The replication mode determines the read and write capabilities of each server in the cluster. MariaDB supports two primary replication architectures: multi-master and master-slave.

- Multi-Master Replication: In a multi-master setup, every server in the cluster can handle both read and write operations. This architecture, exemplified by the MariaDB Galera Cluster, ensures high availability and scalability by allowing applications to distribute workloads across all nodes.
- Master-Slave Replication: In a master-slave architecture, the master server supports both read and write operations, while slave servers are limited to read operations. This setup is ideal for read-heavy workloads, with slaves providing scalability for queries and serving as backups.

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

MariaDB Standard Replication is a widely used mechanism for copying data from one MariaDB server (master) to one or more other MariaDB servers (slave). This replication mode ensures data redundancy, enhances availability, and supports read scalability by distributing read operations across multiple servers. It operates in a single-master and slave architecture, where the master server handles both read and write operations, while slave servers are limited to read operations.

Ref: [About MariadDB Standard Replication](https://mariadb.com/kb/en/replication-overview/#standard-replication)

![MariaDB Standard Replication Cluster](/docs/guides/mariadb/clustering/overview/images/mariadb-standard-replication.png)


## MariaDB Standard Replication Cluster Features

- Asynchronous Replication
- Active-passive single-master topology
- Read and write to master, and read only to slave node
- Automatic node joining
- Automatic failover using Maxscale Proxy Server
- Binary Log-Based replication
- Direct client connections, native MariaDB look & feel
- Load Balance using Maxscale Proxy Server
- Read Write Split using Maxscale Proxy Server

### MariaDB Maxscale Proxy Server
MariaDB MaxScale is a tool that acts as a middleman between your applications and MariaDB Server databases. It helps improve the database's availability, ability to handle more users, and security. It also makes it easier for developers by separating the application from the database setup.

MaxScale has a flexible design that supports add-ons (plugins) to do more than just basic load balancing. For example, it can act as a database firewall to protect your data. It comes with built-in plugins for routing queries, filtering data, and supporting different protocols. You can set up MaxScale to direct database requests or modify responses based on your needs, such as hiding sensitive information or spreading read requests across multiple servers to handle more traffic.

One of its key features is auto failover, which ensures your database remains available even if a primary server fails. Auto failover automatically detects a failed primary node, promotes a replica to take its place, and redirects traffic to the new primary, minimizing downtime and keeping your applications running smoothly.


### How Auto Failover Works with MaxScale
MaxScale uses a monitor (like the MariaDB-Monitor plugin) to track the health of database nodes in a replication setup (e.g. MariaDB master-replica). If the primary node becomes unavailable—due to a crash, network issue, or maintenance—

MaxScale:
- Detects the Failure: The monitor(ReplicationMonitor) continuously checks node status (using MySQL pings or status variables).
- Selects a New Primary: It identifies the most suitable replica based on criteria like replication lag or server state.
- Promotes the Replica: MaxScale executes commands to promote the chosen replica to primary (e.g. STOP SLAVE; RESET SLAVE ALL;).
- Reconfigures Replicas: Other replicas are updated to replicate from the new primary.
- Redirects Traffic: MaxScale’s router (RW-Split-Router) seamlessly directs write queries to the new primary and read queries to replicas.

This process happens automatically, typically within seconds, ensuring minimal disruption.

# Next Steps 
- [Deploy MariaDB Galera Cluster](/docs/guides/mariadb/clustering/galera-cluster) using KubeDB.
