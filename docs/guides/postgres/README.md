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
- Check Update Version of PostgreSQL database with KubeDB using [Update Version](/docs/guides/postgres/update-version/versionupgrading)
- Use [private Docker registry](/docs/guides/postgres/private-registry/using-private-registry.md) to deploy PostgreSQL with KubeDB.
- Detail concepts of [Postgres object](/docs/guides/postgres/concepts/postgres.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
