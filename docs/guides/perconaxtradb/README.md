---
title: PerconaXtraDB
menu:
  docs_{{ .version }}:
    identifier: guides-perconaxtradb-overview
    name: PerconaXtraDB
    parent: guides-perconaxtradb
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
url: /docs/{{ .version }}/guides/perconaxtradb/
aliases:
  - /docs/{{ .version }}/guides/perconaxtradb/README/
---


> New to KubeDB? Please start [here](/docs/README.md).

## Supported PerconaXtraDB Features

| Features                                                | Availability |
| ------------------------------------------------------- | :----------: |
| Clustering                                              |   &#10003;   |
| Persistent Volume                                       |   &#10003;   |
| Instant Backup                                          |   &#10003;   |
| Scheduled Backup                                        |   &#10003;   |
| Initialize using Snapshot                               |   &#10003;   |
| Custom Configuration                                    |   &#10003;   |
| Using Custom docker image                               |   &#10003;   |
| Builtin Prometheus Discovery                            |   &#10003;   |
| Using Prometheus operator                               |   &#10003;   |

## Life Cycle of a PerconaXtraDB Object

<p align="center">
  <img alt="lifecycle"  src="/docs/guides/perconaxtradb/images/perconaxtradb-lifecycle.png" >
</p>

## User Guide

- [Quickstart PerconaXtraDB](/docs/guides/perconaxtradb/quickstart/overview) with KubeDB Operator.
- Detail concepts of [PerconaXtraDB object](/docs/guides/perconaxtradb/concepts/perconaxtradb).
- Detail concepts of [PerconaXtraDBVersion object](/docs/guides/perconaxtradb/concepts/perconaxtradb-version).
- Create [PerconaXtraDB Cluster](/docs/guides/perconaxtradb/clustering/galera-cluster).
- Create [PerconaXtraDB with Custom Configuration](/docs/guides/perconaxtradb/configuration/using-config-file).
- Use [Custom RBAC](/docs/guides/perconaxtradb/custom-rbac/using-custom-rbac).
- Use [private Docker registry](/docs/guides/perconaxtradb/private-registry/quickstart) to deploy MySQL with KubeDB.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
