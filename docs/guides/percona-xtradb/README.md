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

> New to KubeDB? Please start [here](/docs/concepts/README.md).

## Supported PerconaXtraDB Features

|                        Features                         | Availability |
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
| Using CoreOS Prometheus Operator                        |   &#10003;   |

## Life Cycle of a PerconaXtraDB Object

<p align="center">
  <img alt="lifecycle" src="/docs/images/percona-xtradb/Lifecycle_of_a_PerconaXtraDB.svg" >
</p>

## Supported PerconaXtraDB Versions

| KubeDB Version | PerconaXtraDB:5.7 | PerconaXtraDB:5.7-cluster |
| :------------: | :---------------: | :-----------------------: |
|  v0.13.0-rc.1  |      &#10003;     |         &#10003;          |

## Supported PerconaXtraDBVersion CRD

Here, &#10003; means supported and &#10007; means deprecated.

|    NAME     | VERSION | KubeDB: v0.13.0-rc.0 | KubeDB: v0.13.0-rc.1 |
| :---------: | :-----: | :------------------: | :------------------: |
|     5.7     |   5.7   |       &#10007;       |       &#10003;       |
| 5.7-cluster |   5.7   |       &#10007;       |       &#10003;       |

## External tools dependency

|                                      Tool                                      | Version |
| :----------------------------------------------------------------------------: | :-----: |
| [peer-finder](https://github.com/kubernetes/contrib/tree/master/peer-finder)   | latest  |

## User Guide

- [Overview](/docs/guides/percona-xtradb/overview/overview.md) of PerconaXtraDB.
- [Quickstart PerconaXtraDB](/docs/guides/percona-xtradb/quickstart/quickstart.md) with KubeDB Operator.
- How to run [PerconaXtraDB Cluster](/docs/guides/percona-xtradb/clustering/percona-xtradb-cluster.md).
- Monitor your PerconaXtraDB database with KubeDB using [out-of-the-box CoreOS Prometheus Operator](/docs/guides/percona-xtradb/monitoring/using-coreos-prometheus-operator.md).
- Monitor your PerconaXtraDB database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/percona-xtradb/monitoring/using-builtin-prometheus.md).
- Use [private Docker registry](/docs/guides/percona-xtradb/private-registry/using-private-registry.md) to deploy PerconaXtraDB with KubeDB.
- Use Stash to [Backup PerconaXtraDB](/docs/guides/percona-xtradb/snapshot/stash.md).
- How to use [custom configuration](/docs/guides/percona-xtradb/configuration/using-custom-config.md).
- Detail concepts of [PerconaXtraDB object](/docs/concepts/databases/percona-xtradb.md).
- Detail concepts of [PerconaXtraDBVersion object](/docs/concepts/catalog/percona-xtradb.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
