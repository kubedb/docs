---
title: Postgres
menu:
  docs_{{ .version }}:
    identifier: pg-readme-postgres
    name: Postgres
    parent: pg-postgres-guides
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
url: /docs/{{ .version }}/guides/postgres/
aliases:
  - /docs/{{ .version }}/guides/postgres/README/
---

> New to KubeDB? Please start [here](/docs/README.md).

## Supported PostgreSQL Features

| Features                           | Availability |
|------------------------------------|:------------:|
| Clustering                         |   &#10003;   |
| Warm Standby                       |   &#10003;   |
| Hot Standby                        |   &#10003;   |
| Synchronous Replication            |   &#10003;   |
| Streaming Replication              |   &#10003;   |
| Automatic Failover                 |   &#10003;   |
| Continuous Archiving using `wal-g` |   &#10003;   |
| Initialization from WAL archive    |   &#10003;   |
| Persistent Volume                  |   &#10003;   |
| Instant Backup                     |   &#10003;   |
| Scheduled Backup                   |   &#10003;   |
| Initialization from Snapshot       |   &#10003;   |
| Initialization using Script        |   &#10003;   |
| Builtin Prometheus Discovery       |   &#10003;   |
| Using Prometheus operator          |   &#10003;   |
| Custom Configuration               |   &#10003;   |
| Using Custom docker image          |   &#10003;   |

## Life Cycle of a PostgreSQL Object

<p align="center">
  <img alt="lifecycle"  src="/docs/images/postgres/lifecycle.png">
</p>

## User Guide

- [Quickstart PostgreSQL](/docs/guides/postgres/quickstart/quickstart.md) with KubeDB Operator.
- How to [Backup & Restore](/docs/guides/postgres/backup/stash/overview/index.md) PostgreSQL database using Stash.
- Initialize [PostgreSQL with Script](/docs/guides/postgres/initialization/script_source.md).
- [PostgreSQL Clustering](/docs/guides/postgres/clustering/ha_cluster.md) supported by KubeDB Postgres.
- [Streaming Replication](/docs/guides/postgres/clustering/streaming_replication.md) for PostgreSQL clustering.
- Monitor your PostgreSQL database with KubeDB using [`out-of-the-box` builtin-Prometheus](/docs/guides/postgres/monitoring/using-builtin-prometheus.md).
- Monitor your PostgreSQL database with KubeDB using [`out-of-the-box` Prometheus operator](/docs/guides/postgres/monitoring/using-prometheus-operator.md).
- Use [private Docker registry](/docs/guides/postgres/private-registry/using-private-registry.md) to deploy PostgreSQL with KubeDB.
- Detail concepts of [Postgres object](/docs/guides/postgres/concepts/postgres.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).


## A Guide to Postgres Ops Requests

A `PostgresOpsRequest` lets you manage various database operational and day-2 features. For example, managing Database TLS, custom configuration, version upgrade, scaling, and so on.

### Managing Postgresql database TLS

If you want to use encrypted connection or certificate-based authentication for clients, you can use `PostgresOpsRequest`. Based on your requirements, you can add, remove or rotate tls certificates. For more information, please follow the documentation section [link1](/docs/guides/postgres/tls/overview), [link2](/docs/guides/postgres/reconfigure-tls/overview.md).

### Upgrade Postgresql Version

Upgrading a Postgresql version can be a nightmare for the DBA's. We make this process a lot easier. You can apply a `PostgresOpsRequest` and your database will be upgraded to your desired versions. For more information, check [this](/docs/guides/postgres/update-version/overview/index.md) section of documentation.

> **Note**: Before Upgrading, make sure your current version and the version you want to upgrade to, has the same base image. Also do not try to make a major jump where the major version difference is greater than one.

### Scaling Postgresql Database

Being able to scale the database both horizontally and vertically is a blessing for database to handle more incoming loads. But sadly, just increasing the database replica should not work for most of the databases. Because the databases need to join the cluster and perform a few other database-specific tasks before joining the cluster. Don't worry, we take care of those for you. You simply need to create a `PostgresOpsRequest`, and the scaling will be handled automatically.

#### Horizontal Scaling

For scaling Horizontally, follow [this](/docs/guides/postgres/scaling/horizontal-scaling/overview/index.md) section of the documentation.

#### Vertical Scaling

For vertical scaling, follow [this](/docs/guides/postgres/scaling/vertical-scaling/_index.md) section.

#### Auto Scaling

We also support autoscaling! You can configure auto-scaling your database and forget about the loads that your system might face during peak hours! To set up and configure, visit [here](/docs/guides/postgres/autoscaler/compute/overview.md) for compute autoscaling and [here](/docs/guides/postgres/autoscaler/storage/overview.md) for storage.

### VolumeExpansion of Postgresql Database

For volume expansion, follow [this](/docs/guides/postgres/volume-expansion/Overview/overview.md) section.

### Re-configure Postgresql configuration parameters

Do you need to update your PostgreSQL `shared_buffers`, `max_connections`, or other parameters? You can use our Reconfigure `PostgresOpsRequest`. Follow [here](/docs/guides/postgres/reconfigure/overview.md)

### Remote Replica Support

Do you want to have a backup data center where you want to run your postgresql database to recover from a data center failure as soon as possible?

Follow [here](/docs/guides/postgres/remote-replica/remotereplica.md).

