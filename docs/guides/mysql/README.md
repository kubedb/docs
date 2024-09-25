---
title: MySQL
menu:
  docs_{{ .version }}:
    identifier: guides-mysql-readme
    name: MySQL
    parent: guides-mysql
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
url: /docs/{{ .version }}/guides/mysql/
aliases:
  - /docs/{{ .version }}/guides/mysql/README/
---

> New to KubeDB? Please start [here](/docs/README.md).

## Supported MySQL Features

| Features                                                                                | Availability |
| --------------------------------------------------------------------------------------- | :----------: |
| Group Replication                                                                       |   &#10003;   |
| Innodb Cluster                                                                          |   &#10003;   |
| SemiSynchronous cluster                                                                 |   &#10003;   |
| Read Replicas                                                                           |   &#10003;   |
| TLS: Add, Remove, Update, Rotate ( [Cert Manager](https://cert-manager.io/docs/) )      |   &#10003;   |
| Automated Version update                                                               |   &#10003;   |
| Automatic Vertical Scaling                                                              |   &#10003;   |
| Automated Horizontal Scaling                                                            |   &#10003;   |
| Automated Volume Expansion                                                              |   &#10003;   |
| Backup/Recovery: Instant, Scheduled ( [Stash](https://stash.run/) )                     |   &#10003;   |
| Initialize using Snapshot                                                               |   &#10003;   |
| Initialize using Script (\*.sql, \*sql.gz and/or \*.sh)                                 |   &#10003;   |
| Custom Configuration                                                                    |   &#10003;   |
| Using Custom docker image                                                               |   &#10003;   |
| Builtin Prometheus Discovery                                                            |   &#10003;   |
| Using Prometheus operator                                                               |   &#10003;   |

## Life Cycle of a MySQL Object

<p align="center">
  <img alt="lifecycle"  src="/docs/images/mysql/mysql-lifecycle.png" >
</p>

## User Guide

- [Quickstart MySQL](/docs/guides/mysql/quickstart/index.md) with KubeDB Operator.
- [Backup & Restore](/docs/guides/mysql/backup/stash/overview/index.md) MySQL databases using Stash.
- Initialize [MySQL with Script](/docs/guides/mysql/initialization/index.md).
- Monitor your MySQL database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/mysql/monitoring/prometheus-operator/index.md).
- Monitor your MySQL database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/mysql/monitoring/builtin-prometheus/index.md).
- Use [private Docker registry](/docs/guides/mysql/private-registry/index.md) to deploy MySQL with KubeDB.
- Use [kubedb cli](/docs/guides/mysql/cli/index.md) to manage databases like kubectl for Kubernetes.
- Detail concepts of [MySQL object](/docs/guides/mysql/concepts/database/index.md).
- Detail concepts of [MySQLVersion object](/docs/guides/mysql/concepts/catalog/index.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
