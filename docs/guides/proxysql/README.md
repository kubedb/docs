---
title: ProxySQL
menu:
  docs_{{ .version }}:
    identifier: prx-readme-proxysql
    name: ProxySQL
    parent: prx-proxysql-guides
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
| Load balance PerconaXtraDB Cluster   |   &#10007;   |
| Custom Configuration                 |   &#10003;   |
| Using Custom docker image            |   &#10003;   |
| Builtin Prometheus Discovery         |   &#10003;   |
| Using Prometheus operator            |   &#10003;   |
| ProxySQL server cluster              |   &#10003;   |
| ProxySQL server failure recovery     |   &#10003;   |
| TLS secured connection for backend   |   &#10003;   |
| TLS secured connection for frontend  |   &#10003;   |

## User Guide

- Overview of ProxySQL [here](/docs/guides/proxysql/overview/overview.md).
- Configure ProxySQL for Group Replication [here](/docs/guides/proxysql/overview/configure-proxysql.md).
- Learn to use ProxySQL to Load Balance MySQL Group Replication with KubeDB Operator [here](/docs/guides/proxysql/quickstart/load-balance-mysql-group-replication.md).
- Monitor your ProxySQL with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/proxysql/monitoring/using-builtin-prometheus.md).
- Monitor your ProxySQL with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/proxysql/monitoring/using-prometheus-operator.md).
- Use private Docker registry to deploy ProxySQL with KubeDB [here](/docs/guides/proxysql/private-registry/using-private-registry.md).
- Use custom config file to configure ProxySQL [here](/docs/guides/proxysql/configuration/using-config-file.md).
- Detail concepts of ProxySQL CRD [here](/docs/guides/proxysql/concepts/proxysql.md).
- Detail concepts of ProxySQLVersion CRD [here](/docs/guides/proxysql/concepts/catalog.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
