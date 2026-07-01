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

- [Quickstart ClickHouse](/docs/guides/clickhouse/quickstart/guide/quickstart.md) with KubeDB Operator.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).

## Cross-DC Disaster Recovery (DC-DR)

Do you want to run your ClickHouse database across multiple data centers and recover from a full data center failure with a single, automatically re-routing write endpoint? KubeDB runs one logical `ReplicatedMergeTree` cluster across the data centers, spreads ClickHouse Keeper 3-site so no single data center holds a Keeper majority (the split-brain guarantee), and lets the `dr-controlplane` Lease route the write endpoint to a data center that still holds Keeper quorum. Follow [here](/docs/guides/clickhouse/dr/overview/index.md).
