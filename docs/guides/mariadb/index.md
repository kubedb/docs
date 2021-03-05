---
title: MariaDB
menu:
  docs_{{ .version }}:
    identifier: guides-mariadb-overview
    name: MariaDB
    parent: guides-mariadb
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
---


> New to KubeDB? Please start [here](/docs/README.md).

## Supported MariaDB Features

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

## Life Cycle of a MariaDB Object

<p align="center">
  <img alt="lifecycle"  src="/docs/guides/mariadb/images/mariadb-lifecycle.png" >
</p>

## User Guide

- [Quickstart MariaDB](/docs/guides/mariadb/quickstart/quickstart.md) with KubeDB Operator.
- [Backup & Restore](/docs/guides/mariadb/backup/stash.md) MariaDB databases using Stash.
- Initialize [MariaDB with Script](/docs/guides/mariadb/initialization/using-script.md).
- Monitor your MariaDB database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/mariadb/monitoring/using-prometheus-operator.md).
- Monitor your MariaDB database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/mariadb/monitoring/using-builtin-prometheus.md).
- Use [private Docker registry](/docs/guides/mariadb/private-registry/using-private-registry.md) to deploy MariaDB with KubeDB.
- Use [kubedb cli](/docs/guides/mariadb/cli/cli.md) to manage databases like kubectl for Kubernetes.
- Detail concepts of [MariaDB object](/docs/guides/mariadb/concepts/mariadb.md).
- Detail concepts of [MariaDBVersion object](/docs/guides/mariadb/concepts/catalog.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
