---
title: PerconaXtraDB
menu:
  docs_{{ .version }}:
    identifier: readme-percona-xtradb
    name: PerconaXtraDB
    parent: px-percona-xtradb-guides
    weight: 40
menu_name: docs_{{ .version }}
section_menu_id: guides
url: /docs/{{ .version }}/guides/percona-xtradb/
aliases:
  - /docs/{{ .version }}/guides/percona-xtradb/README/
---

> New to KubeDB? Please start [here](/docs/README.md).

## Supported PerconaXtraDB Features

| Features                                                | Availability |
| ------------------------------------------------------- | :----------: |
| Clustering                                              |   &#10003;   |
| Persistent Volume                                       |   &#10003;   |
| Instant Backup                                          |   &#10003;   |
| Scheduled Backup                                        |   &#10003;   |
| Initialize using Snapshot                               |   &#10003;   |
| Initialize using Script (\*.sql, \*sql.gz and/or \*.sh) |   &#10003;   |
| Custom Configuration                                    |   &#10003;   |
| Using Custom docker image                               |   &#10003;   |
| Builtin Prometheus Discovery                            |   &#10003;   |
| Using Prometheus operator                               |   &#10003;   |

## Life Cycle of a PerconaXtraDB Object

<p align="center">
  <img alt="lifecycle" src="/docs/images/percona-xtradb/Lifecycle_of_a_PerconaXtraDB.svg" >
</p>

## User Guide

- [Overview](/docs/guides/percona-xtradb/overview/overview.md) of PerconaXtraDB.
- [Quickstart PerconaXtraDB](/docs/guides/percona-xtradb/quickstart/quickstart.md) with KubeDB Operator.
- How to run [PerconaXtraDB Cluster](/docs/guides/percona-xtradb/clustering/percona-xtradb-cluster.md).
- Monitor your PerconaXtraDB database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/percona-xtradb/monitoring/using-prometheus-operator.md).
- Monitor your PerconaXtraDB database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/percona-xtradb/monitoring/using-builtin-prometheus.md).
- Use [private Docker registry](/docs/guides/percona-xtradb/private-registry/using-private-registry.md) to deploy PerconaXtraDB with KubeDB.
- Use Stash to [Backup PerconaXtraDB](/docs/guides/percona-xtradb/backup/stash.md).
- How to use [custom configuration](/docs/guides/percona-xtradb/configuration/using-config-file.md).
- Detail concepts of [PerconaXtraDB object](/docs/guides/percona-xtradb/concepts/percona-xtradb.md).
- Detail concepts of [PerconaXtraDBVersion object](/docs/guides/percona-xtradb/concepts/catalog.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
