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

Microsoft SQL Server, one of the most popular relational database management systems (RDBMS) in the world. KubeDB support for Provisioning SQL Server Availability Group and Standalone SQL Server instance. Utilize SQL Server’s high availability features by deploying instances in availability group mode. KubeDB leverages the Raft Consensus Algorithm for cluster coordination, enabling automatic leader election and failover decisions. Quorum support ensures the reliability and fault tolerance of your SQL Server deployments. You can also deploy SQL Server instances in standalone mode for simple, single-node configurations. KubeDB users can now seamlessly provision and manage SQL Server instances directly within their Kubernetes clusters.

## Supported Microsoft SQL Server Features

| Features                                                                   | Availability |
|----------------------------------------------------------------------------|:------------:|
| Standalone and Availability Group                                          |   &#10003;   |
| Authentication & Authorization                                             |   &#10003;   |
| Backup/Recovery: Instant, Scheduled ( [KubeStash](https://kubestash.com/)) |   &#10003;   |
| Initializing from Snapshot ( [KubeStash](https://kubestash.com/))          |   &#10003;   |
| Externally manageable Auth Secret                                          |   &#10003;   |
| Reconfigurable Health Checker                                              |   &#10003;   |
| Persistent volume                                                          |   &#10003;   | 


## Supported Microsoft SQL Server Versions

KubeDB supports the following Microsoft SQL Server Version.
- `2022-CU12-ubuntu-22.04`

## Life Cycle of a Microsoft SQL Server Object

<!---
ref : https://cacoo.com/diagrams/4PxSEzhFdNJRIbIb/0281B
--->

<p align="center">
  <img alt="lifecycle"  src="/docs/guides/mssqlserver/images/mssqlserver-lifecycle.png" >
</p>

## User Guide

- [Quickstart MSSQLServer](/docs/guides/mssqlserver/quickstart/quickstart.md) with KubeDB Operator.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).