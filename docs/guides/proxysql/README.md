---
title: ProxySQL
menu:
  docs_{{ .version }}:
    identifier: guides-proxysql-readme
    name: ProxySQL
    parent: guides-proxysql
    weight: 5
menu_name: docs_{{ .version }}
section_menu_id: guides
url: /docs/{{ .version }}/guides/proxysql/
aliases:
  - /docs/{{ .version }}/guides/proxysql/README/
---

> New to KubeDB? Please start [here](/docs/README.md).

## Supported ProxySQL Features

| Features                                                          | Availability |
|-------------------------------------------------------------------|:------------:|
| Load Balancing (MySQL Group Replication, PerconaXtraDB, MariaDB) |   &#10003;   |
| Custom & Declarative Configuration                                |   &#10003;   |
| Version Update                                                    |   &#10003;   |
| Monitoring (Prometheus)                                           |   &#10003;   |
| Clustering & Failure Recovery                                     |   &#10003;   |
| TLS/SSL Encryption (Frontend & Backend, Reconfigure)              |   &#10003;   |
| Horizontal & Vertical Scaling                                     |   &#10003;   |
| Autoscaling (Compute)                                             |   &#10003;   |
| Custom RBAC Resources                                             |   &#10003;   |
| Restart                                                           |   &#10003;   |

## User Guide

- [Overview of KubeDB ProxySQL CRD](/docs/guides/proxysql/concepts/proxysql/index.md) 
- [Configure KubeDB ProxySQL for MySQL Group Replication](/docs/guides/proxysql/quickstart/mysqlgrp/index.md)
- [Configure KubeDB ProxySQL for PerconaXtraDB Cluster (External)](/docs/guides/proxysql/quickstart/xtradbext/index.md)
- [Configure KubeDB ProxySQL for MariaDB Galera Cluster](/docs/guides/proxysql/backends/mariadb-galera/index.md)
- [Deploy ProxySQL cluster with KubeDB](/docs/guides/proxysql/clustering/proxysql-cluster/index.md) 
- [Initialize KubeDB ProxySQL with declarative configuration](/docs/guides/proxysql/concepts/declarative-configuration/index.md) 
- [Reconfigure KubeDB ProxySQL with ops-request](/docs/guides/proxysql/reconfigure/cluster/index.md)
- [Deploy TLS/SSL secured KubeDB ProxySQL](/docs/guides/proxysql/tls/configure/index.md)
- [Reconfigure TLS/SSL for KubeDB ProxySQL](/docs/guides/proxysql/reconfigure-tls/cluster/index.md)
- [Detail concepts of ProxySQLVersion CRD](/docs/guides/proxysql/concepts/proxysql-version/index.md)
- [Update KubeDB ProxySQL version with ops-request](/docs/guides/proxysql/update-version/cluster/index.md)
- [Horizontal Scale KubeDB ProxySQL with ops-request](/docs/guides/proxysql/scaling/horizontal-scaling/cluster/index.md)
- [Vertical Scale KubeDB ProxySQL with ops-request](/docs/guides/proxysql/scaling/vertical-scaling/cluster/index.md)
- [Learn auto-scaling for KubeDB ProxySQL](/docs/guides/proxysql/autoscaler/compute/cluster/index.md)
- [Run ProxySQL with Custom RBAC resources](/docs/guides/proxysql/custom-rbac/index.md)
- [Restart KubeDB ProxySQL with ops-request](/docs/guides/proxysql/restart/index.md)
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
