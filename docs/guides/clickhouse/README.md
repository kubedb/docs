---
title: ClickHouse
menu:
  docs_{{ .version }}:
    identifier: guides-clickhouse-readme
    name: ClickHouse
    parent: ch-clickhouse-guides
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
url: /docs/{{ .version }}/guides/clickhouse/
aliases:
  - /docs/{{ .version }}/guides/clickhouse/README/
---

> New to KubeDB? Please start [here](/docs/README.md).

## Supported ClickHouse Features

| Features                                                | Availability |
|---------------------------------------------------------|:------------:|
| ClusterTopology                                         |   &#10003;   |
| Initialize using Script (\*.sql, \*sql.gz and/or \*.sh) |   &#10003;   |
| Custom Configuration                                    |   &#10003;   |
| Monitoring (Prometheus)                                 |   &#10003;   |
| TLS/SSL Encryption                                      |   &#10003;   |
| Externally manageable Auth Secret                       |   &#10003;   |
| Reconfigure                                             |   &#10003;   |
| Horizontal & Vertical Scaling                           |   &#10003;   |
| Volume Expansion                                        |   &#10003;   |
| Update Version                                          |   &#10003;   |
| Restart                                                 |   &#10003;   |
| Rotate Authentication                                   |   &#10003;   |
| Autoscaling                                             |   &#10003;   |

## Supported ClickHouse Versions

KubeDB supports the following ClickHouse Versions.
- `24.4.1`
- `25.7.1`

## Life Cycle of a ClickHouse Object

<p align="center">
  <img alt="lifecycle"  src="/docs/images/clickhouse/clickhouse-lifecycle.png" >
</p>

## User Guide

- [Quickstart ClickHouse](/docs/guides/clickhouse/quickstart/guide/quickstart.md) with KubeDB Operator.
- [Custom Configuration](/docs/guides/clickhouse/configuration/using-config-file.md) of ClickHouse.
- [Using Builtin Prometheus](/docs/guides/clickhouse/monitoring/using-builtin-prometheus.md) for monitoring.
- [Using Prometheus Operator](/docs/guides/clickhouse/monitoring/using-prometheus-operator.md) for monitoring.
- [Configure TLS/SSL](/docs/guides/clickhouse/tls/cluster.md) for ClickHouse.
- [Reconfigure TLS](/docs/guides/clickhouse/reconfigure-tls/clickhouse.md) for ClickHouse.
- [Reconfigure](/docs/guides/clickhouse/reconfigure/reconfigure.md) ClickHouse.
- [Horizontal Scaling](/docs/guides/clickhouse/scaling/horizontal-scaling/cluster.md) of ClickHouse.
- [Vertical Scaling](/docs/guides/clickhouse/scaling/vertical-scaling/cluster.md) of ClickHouse.
- [Volume Expansion](/docs/guides/clickhouse/volume-expansion/cluster.md) of ClickHouse.
- [Update Version](/docs/guides/clickhouse/update-version/update-version.md) of ClickHouse.
- [Restart](/docs/guides/clickhouse/restart/restart.md) ClickHouse.
- [Rotate Authentication](/docs/guides/clickhouse/rotate-auth/rotateauth.md) for ClickHouse.
- [Compute Autoscaling](/docs/guides/clickhouse/autoscaler/compute/compute-autoscale.md) of ClickHouse.
- [Storage Autoscaling](/docs/guides/clickhouse/autoscaler/storage/storage-autoscale.md) of ClickHouse.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
