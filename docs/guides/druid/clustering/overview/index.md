---
title: Druid Topology Cluster Overview
menu:
  docs_{{ .version }}:
    identifier: guides-druid-clustering-overview
    name: Druid Clustering Overview
    parent: guides-druid-clustering
    weight: 15
menu_name: docs_{{ .version }}
section_menu_id: guides
---

> New to KubeDB? Please start [here](/docs/README.md).

# Druid Architecture

Druid has a distributed architecture that is designed to be cloud-friendly and easy to operate. You can configure and scale services independently for maximum flexibility over cluster operations. This design includes enhanced fault tolerance: an outage of one component does not immediately affect other components.

The following diagram shows the services that make up the Druid architecture, their typical arrangement across servers, and how queries and data flow through this architecture.

![Druid Architecture](/docs/guides/druid/clustering/overview/images/druid-architecture.svg)

Image ref: <https://druid.apache.org/assets/images/druid-architecture-7db1cd79d2d70b2e5ccc73b6bebfcaa4.svg>


## Druid services

Druid has several types of services:

- **Coordinator** manages data availability on the cluster.
- **Overlord** controls the assignment of data ingestion workloads.
- **Broker** handles queries from external clients.
- **Router** routes requests to Brokers, Coordinators, and Overlords.
- **Historical** stores queryable data.
- **MiddleManager** and Peon ingest data.
- **Indexer** serves an alternative to the MiddleManager + Peon task execution system.

## Druid servers

You can deploy Druid services according to your preferences. For ease of deployment, we recommend organizing them into three server types: **Master**, **Query**, and **Data**.

### Master server
A Master server manages data ingestion and availability. It is responsible for starting new ingestion jobs and coordinating availability of data on the Data server.

Master servers divide operations between **Coordinator** and **Overlord** services.

#### Coordinator service
[Coordinator](https://druid.apache.org/docs/latest/design/coordinator/) services watch over the Historical services on the Data servers. They are responsible for assigning segments to specific servers, and for ensuring segments are well-balanced across Historicals.

### Overlord service
[Overlord](https://druid.apache.org/docs/latest/design/overlord/) services watch over the MiddleManager services on the Data servers and are the controllers of data ingestion into Druid. They are responsible for assigning ingestion tasks to MiddleManagers and for coordinating segment publishing.

### Query server
A Query server provides the endpoints that users and client applications interact with, routing queries to Data servers or other Query servers (and optionally proxied Master server requests).

Query servers divide operations between Broker and Router services.

#### Broker service
[Broker](https://druid.apache.org/docs/latest/design/broker/) services receive queries from external clients and forward those queries to Data servers. When Brokers receive results from those subqueries, they merge those results and return them to the caller. Typically, you query Brokers rather than querying Historical or MiddleManager services on Data servers directly.

#### Router service
[Router](https://druid.apache.org/docs/latest/design/router/) services provide a unified API gateway in front of Brokers, Overlords, and Coordinators.

The Router service also runs the web console, a UI for loading data, managing datasources and tasks, and viewing server status and segment information.

### Data server
A Data server executes ingestion jobs and stores queryable data.

Data servers divide operations between Historical and MiddleManager services.

#### Historical service
[Historical](https://druid.apache.org/docs/latest/design/historical/) services handle storage and querying on historical data, including any streaming data that has been in the system long enough to be committed. Historical services download segments from deep storage and respond to queries about these segments. They don't accept writes.

#### MiddleManager service
[MiddleManager](https://druid.apache.org/docs/latest/design/middlemanager) services handle ingestion of new data into the cluster. They are responsible for reading from external data sources and publishing new Druid segments.

## External dependencies
In addition to its built-in service types, Druid also has three external dependencies. These are intended to be able to leverage existing infrastructure, where present.

### Deep storage
Druid uses deep storage to store any data that has been ingested into the system. Deep storage is shared file storage accessible by every Druid server. In a clustered deployment, this is typically a distributed object store like S3 or HDFS, or a network mounted filesystem. In a single-server deployment, this is typically local disk.

Druid uses deep storage for the following purposes:

- To store all the data you ingest. Segments that get loaded onto Historical services for low latency queries are also kept in deep storage for backup purposes. Additionally, segments that are only in deep storage can be used for queries from deep storage.
- As a way to transfer data in the background between Druid services. Druid stores data in files called segments.

Historical services cache data segments on local disk and serve queries from that cache as well as from an in-memory cache. Segments on disk for Historical services provide the low latency querying performance Druid is known for.

You can also query directly from deep storage. When you query segments that exist only in deep storage, you trade some performance for the ability to query more of your data without necessarily having to scale your Historical services.

When determining sizing for your storage, keep the following in mind:

- Deep storage needs to be able to hold all the data that you ingest into Druid.
- On disk storage for Historical services need to be able to accommodate the data you want to load onto them to run queries. The data on Historical services should be data you access frequently and need to run low latency queries for. 

- Deep storage is an important part of Druid's elastic, fault-tolerant design. Druid bootstraps from deep storage even if every single data server is lost and re-provisioned.

For more details, please see the [Deep storage](https://druid.apache.org/docs/latest/design/deep-storage/) page.

### Metadata storage
The metadata storage holds various shared system metadata such as segment usage information and task information. In a clustered deployment, this is typically a traditional RDBMS like PostgreSQL or MySQL. In a single-server deployment, it is typically a locally-stored Apache Derby database.

For more details, please see the [Metadata storage](https://druid.apache.org/docs/latest/design/metadata-storage/) page.

### ZooKeeper
Used for internal service discovery, coordination, and leader election.

For more details, please see the [ZooKeeper](https://druid.apache.org/docs/latest/design/zookeeper/) page.


## Next Steps

- [Deploy Druid Cluster](/docs/guides/druid/clustering/overview/index.md) using KubeDB.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md)
