---
title: Pgpool
menu:
  docs_{{ .version }}:
    identifier: pp-readme-pgpool
    name: Pgpool
    parent: pp-pgpool-guides
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
url: /docs/{{ .version }}/guides/pgpool/
aliases:
  - /docs/{{ .version }}/guides/pgpool/README/
---

> New to KubeDB? Please start [here](/docs/README.md).

# Overview

[Pgpool](https://pgpool.net/) is a versatile proxy solution positioned between PostgreSQL servers and database clients. It offers essential functionalities such as Connection Pooling, Load Balancing, In-Memory Query Cache and many more. Pgpool enhances the performance, scalability, and reliability of PostgreSQL database systems.

KubeDB operator now comes bundled with Pgpool crd to manage all the essential features of Pgpool. 

## Pgpool Features

| Features                      | Availability |
|-------------------------------| :----------: |
| Clustering                    |   &#10003;   |
| Multiple Pgpool Versions      |   &#10003;   |
| Custom Configuration          |   &#10003;   |
| Sync Postgres Users to Pgpool |   &#10003;   |
| Custom docker images          |   &#10003;   |
| Enabling TLS                  |   &#10003;   |
| Builtin Prometheus Discovery  |   &#10003;   |
| Using Prometheus operator     |   &#10003;   |

## User Guide

- [Quickstart Pgpool](/docs/guides/pgpool/quickstart/quickstart.md) with KubeDB Operator.
- Detail concepts of [Pgpool object](/docs/guides/pgpool/concepts/pgpool.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
