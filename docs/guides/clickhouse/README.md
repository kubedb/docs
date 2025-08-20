---
title: ClickHouse
menu:
  docs_{{ .version }}:
    identifier: guides-clickhouse-readme
    name: ClickHouse
    parent: guides-clickhouse
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
url: /docs/{{ .version }}/guides/clickhouse/
aliases:
  - /docs/{{ .version }}/guides/clickhouse/README/
---

> New to KubeDB? Please start [here](/docs/README.md).

## Supported MySQL Features

| Features                                                      | Availability |
|---------------------------------------------------------------|:------------:|
| ClusterTopology                                               |   &#10003;   |
| Initialize using Script (\*.sql, \*sql.gz and/or \*.sh)       |   &#10003;   |
| Custom Configuration                                          |   &#10003;   |
| Builtin Prometheus Discovery                                  |   &#10003;   |
| Using Prometheus operator                                     |   &#10003;   |
| Authentication & Authorization (TLS)                          |   &#10003;   |
| Externally manageable Auth Secret                             |   &#10003;   |
| Reconfigurable TLS Certificates (Add, Remove, Rotate, Update) |   &#10003;   |

## Supported ClickHouse Versions

KubeDB supports the following ClickHouse Versions.
- `24.4.1`
- `25.7.1`

## Life Cycle of a ClickHouse Object

<p align="center">
  <img alt="lifecycle"  src="/docs/images/clickhouse/clickhouse-lifecycle.png" >
</p>

## User Guide

- [Quickstart ClickHouse](/docs/guides/clickhouse/quickstart/index.md) with KubeDB Operator.
- Monitor your ClickHouse database with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/mysql/monitoring/prometheus-operator/index.md).
- Monitor your ClickHouse database with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/mysql/monitoring/builtin-prometheus/index.md).
- Use [kubedb cli](/docs/guides/mysql/cli/index.md) to manage databases like kubectl for Kubernetes.
- Detail concepts of [ClickHouse object](/docs/guides/mysql/concepts/database/index.md).
- Detail concepts of [ClickHouseVersion object](/docs/guides/mysql/concepts/catalog/index.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
