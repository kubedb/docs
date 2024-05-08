---
title: FerretDB
menu:
  docs_{{ .version }}:
    identifier: mg-readme-ferretdb
    name: FerretDB
    parent: fr-ferretdb-guides
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
url: /docs/{{ .version }}/guides/ferretdb/
aliases:
  - /docs/{{ .version }}/guides/ferretdb/README/
---

> New to KubeDB? Please start [here](/docs/README.md).

# Overview

FerretDB is an open-source proxy that translates MongoDB wire protocol queries to SQL, with PostgreSQL or SQLite as the database engine. FerretDB was founded to become the true open-source alternative to MongoDB. It uses the same commands, drivers, and tools as MongoDB.

## Supported FerretDB Features

| Features                              | Availability |
|---------------------------------------|:------------:|
| Internally  manageable Backend Engine |   &#10003;   |
| Externally manageable Backend Engine  |   &#10003;   |
| Authentication & Authorization        |   &#10003;   |
| TLS Support                           |   &#10003;   |
| Monitoring using Prometheus           |   &#10003;   |
| Builtin Prometheus Discovery          |   &#10003;   |
| Using Prometheus operator             |   &#10003;   |
| Reconfigurable Health Checker         |   &#10003;   |
| Persistent volume                     |   &#10003;   |

## Supported FerretDB Versions

KubeDB supports the following FerretDB Versions.
- `1.18.0`

## Life Cycle of a FerretDB Object

<!---
ref : https://app.diagrams.net/
--->

<p text-align="center">
    <img alt="lifecycle"  src="/docs/images/ferretdb/quick-start.png" >
</p>

## User Guide

- [Quickstart FerretDB](/docs/guides/ferretdb/quickstart/quickstart.md) with KubeDB Operator.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).