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
url: /docs/{{ .version }}/guides/mariadb/
aliases:
  - /docs/{{ .version }}/guides/mariadb/README/
---


> New to KubeDB? Please start [here](/docs/README.md).

## Supported MariaDB Features

| Features                                                | Availability |
|---------------------------------------------------------| :----------: |
| Clustering                                              |   &#10003;   |
| Persistent Volume                                       |   &#10003;   |
| Instant Backup                                          |   &#10003;   |
| Scheduled Backup                                        |   &#10003;   |
| Continuous Archiving using `wal-g`                      |   &#10003;   |
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

- [Quickstart MariaDB](/docs/guides/mariadb/quickstart/overview) with KubeDB Operator.
- Detail concepts of [MariaDB object](/docs/guides/mariadb/concepts/mariadb).
- Detail concepts of [MariaDBVersion object](/docs/guides/mariadb/concepts/mariadb-version).
- Create [MariaDB Cluster](/docs/guides/mariadb/clustering/galera-cluster).
- Create [MariaDB with Custom Configuration](/docs/guides/mariadb/configuration/using-config-file).
- Use [Custom RBAC](/docs/guides/mariadb/custom-rbac/using-custom-rbac).
- Use [private Docker registry](/docs/guides/mariadb/private-registry/quickstart) to deploy MySQL with KubeDB.
- Initialize [MariaDB with Script](/docs/guides/mariadb/initialization/using-script).
- Backup and Restore [MariaDB](/docs/guides/mariadb/backup/stash/overview).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
