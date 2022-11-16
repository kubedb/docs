---
title: ProxySQL
menu:
  docs_{{ .version }}:
    identifier: guides-proxysql
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
| Load balance PerconaXtraDB Cluster   |   &#10007;   |
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

#TODO : edit the links after adding all 
- Overview of KubeDB ProxySQL CRD 
- Configure KubeDB ProxySQL for MySQL Group Replication
- Deploy ProxySQL cluster with KubeDB 
- Initialize KubeDB ProxySQL with declarative configuration 
- Initialize KubeDB ProxySQL with configuration secret
- Reconfigure KubeDB ProxySQL with ops-request
- Deploy TLS/SSL secured KubeDB ProxySQL
- Reconfigure TLS/SSL for KubeDB ProxySQL
- Detail concepts of ProxySQLVersion CRD 
- Upgrade KubeDB ProxySQL version with ops-request
- Scale horizontally and vertically KubeDB ProxySQL with ops-request
- Learn auto-scaling for KubeDB ProxySQL
- Monitor your ProxySQL with KubeDB using prometheus
- Want to hack on KubeDB? Check our 
