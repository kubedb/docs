---
title: ProxySQL
menu:
  docs_{{ .version }}:
    identifier: guides-proxysql-readme
    name: ProxySQL
    parent: guides
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
url: /docs/{{ .version }}/guides/proxysql/
aliases:
  - /docs/{{ .version }}/guides/proxysql/README/
---

> New to KubeDB? Please start [here](/docs/README.md).

{{< notice type="warning" message="This is an Enterprise-only feature. Please install [KubeDB Enterprise Edition](/docs/setup/install/enterprise.md) to try this feature." >}}

## Supported ProxySQL Features

| Features                             | Availability |
| ------------------------------------ | :----------: |
| Load balance MySQL Group Replication |   &#10003;   |
| Load balance PerconaXtraDB Cluster   |   &#10003;   |
| Custom Configuration                 |   &#10003;   |
| Declarative Configuration            |   &#10003;   |
| Version Update                       |   &#10003;   |
| Builtin Prometheus Discovery         |   &#10003;   |
| Using Prometheus operator            |   &#10003;   |
| ProxySQL server cluster              |   &#10003;   |
| ProxySQL server failure recovery     |   &#10003;   |
| TLS secured connection for backend   |   &#10003;   |
| TLS secured connection for frontend  |   &#10003;   |

## User Guide

- [Overview of KubeDB ProxySQL CRD](/docs/guides/proxysql/concepts/proxysql/index.md) 
- [Configure KubeDB ProxySQL for MySQL Group Replication](/docs/guides/proxysql/quickstart/mysqlgrp/index.md)
- [Deploy ProxySQL cluster with KubeDB](/docs/guides/proxysql/clustering/proxysql-cluster/index.md) 
- [Initialize KubeDB ProxySQL with declarative configuration](/docs/guides/proxysql/concepts/declarative-configuration/index.md) 
- [Reconfigure KubeDB ProxySQL with ops-request](/docs/guides/proxysql/concepts/opsrequest/index.md)
- [Deploy TLS/SSL secured KubeDB ProxySQL](/docs/guides/proxysql/tls/configure/index.md)
- [Reconfigure TLS/SSL for KubeDB ProxySQL](/docs/guides/proxysql/reconfigure-tls/cluster/index.md)
- [Detail concepts of ProxySQLVersion CRD](/docs/guides/proxysql/concepts/proxysql-version/index.md)
- [Upgrade KubeDB ProxySQL version with ops-request](/docs/guides/proxysql/upgrading/cluster/index.md)
- [Scale horizontally and vertically KubeDB ProxySQL with ops-request](/docs/guides/proxysql/scaling/horizontal-scaling/cluster/index.md)
- [Learn auto-scaling for KubeDB ProxySQL](/docs/guides/proxysql/autoscaler/compute/cluster/index.md)
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
