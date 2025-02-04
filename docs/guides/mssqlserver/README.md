---
title: Microsoft SQL Server
menu:
  docs_{{ .version }}:
    identifier: guides-mssqlserver-readme
    name: Microsoft SQL Server
    parent: guides-mssqlserver
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
url: /docs/{{ .version }}/guides/mssqlserver/
aliases:
  - /docs/{{ .version }}/guides/mssqlserver/README/
---
> New to KubeDB? Please start [here](/docs/README.md).

# Overview

Microsoft SQL Server is one of the most popular relational database management systems (RDBMS) in the world. KubeDB support provisioning for SQL Server Availability Group and Standalone SQL Server instances. Utilize SQL Server’s high availability features by deploying instances in availability group mode. KubeDB leverages the Raft Consensus Algorithm for cluster coordination, enabling automatic leader election and fail over decisions. Quorum support ensures the reliability and fault tolerance of your SQL Server deployments. You can also deploy SQL Server instances in standalone mode for simple, single-node configurations. KubeDB users can now seamlessly provision and manage SQL Server instances directly within their Kubernetes clusters.

## Supported Microsoft SQL Server Features

| Features                                                           | Availability |
|--------------------------------------------------------------------|:------------:|
| Standalone and Availability Group Cluster (HA configuration)       |   &#10003;   |
| Synchronous Replication                                            |   &#10003;   |
| Automatic Fail over                                                |   &#10003;   |
| Arbiter Node for quorum in even-sized clusters                     |   &#10003;   |
| Custom Configuration                                               |   &#10003;   |
| Authentication & Authorization                                     |   &#10003;   |
| Externally manageable Auth Secret                                  |   &#10003;   |
| Instant and Scheduled Backup ([KubeStash](https://kubestash.com/)) |   &#10003;   |
| Continuous Archiving using `wal-g`                                 |   &#10003;   |
| Initialization from WAL archive                                    |   &#10003;   |
| Initializing from Snapshot ([KubeStash](https://kubestash.com/))   |   &#10003;   |
| Reconfigurable Health Checker                                      |   &#10003;   |
| Persistent Volume                                                  |   &#10003;   |
| Builtin Prometheus Discovery                                       |   &#10003;   |
| Using Prometheus operator                                          |   &#10003;   |
| Automated Version Update                                           |   &#10003;   |
| Automated Vertical Scaling, Volume Expansion                       |   &#10003;   |
| Automated Horizontal Scaling                                       |   &#10003;   |
| Autoscaling Compute and Storage Resources                          |   &#10003;   |
| Reconfiguration                                                    |   &#10003;   |
| TLS configuration ([Cert Manager](https://cert-manager.io/docs/))  |   &#10003;   |
| Reconfiguration of TLS: Add, Remove, Update, Rotate                |   &#10003;   |
| Grafana Dashboards                                                 |   &#10003;   |


## Supported Microsoft SQL Server Versions

KubeDB supports the following Microsoft SQL Server Version.
- `2022-CU12-ubuntu-22.04`
- `2022-CU14-ubuntu-22.04`
- `2022-CU16-ubuntu-22.04`

## Life Cycle of a Microsoft SQL Server Object

<!---
ref : https://cacoo.com/diagrams/4PxSEzhFdNJRIbIb/0281B
--->

<p align="center">
  <img alt="lifecycle"  src="/docs/guides/mssqlserver/images/mssqlserver-lifecycle.png" >
</p>

## User Guide

- [Quickstart Microsoft SQL Server](/docs/guides/mssqlserver/quickstart/quickstart.md) with KubeDB Operator.
- Detail concepts of [MSSQLServer object](/docs/guides/mssqlserver/concepts/mssqlserver.md).
- [SQL Server Availability Group Clustering](/docs/guides/mssqlserver/clustering/ag_cluster.md) supported by KubeDB.
- How to [Backup & Restore](/docs/guides/mssqlserver/backup/overview/index.md) SQL Server using [KubeStash](https://kubestash.com/).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).