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

## Supported Pgpool Features

| Features                                                    | Availability |
|-------------------------------------------------------------|:------------:|
| Clustering                                                  |   &#10003;   |
| Multiple Pgpool Versions                                    |   &#10003;   |
| Custom Configuration                                        |   &#10003;   |
| Externally manageable Auth Secret                           |   &#10003;   |
| Reconfigurable Health Checker                               |   &#10003;   |
| Integrate with externally managed PostgreSQL                |   &#10003;   |
| Sync Postgres Users to Pgpool                               |   &#10003;   |
| Custom docker images                                        |   &#10003;   |
| TLS: Add ( [Cert Manager]((https://cert-manager.io/docs/))) |   &#10003;   |
| Monitoring with Prometheus & Grafana                        |   &#10003;   |
| Builtin Prometheus Discovery                                |   &#10003;   |
| Using Prometheus operator                                   |   &#10003;   |
| Alert Dashboard                                             |   &#10003;   |
| Grafana Dashboard                                           |   &#10003;   |

## Supported Pgpool Versions

KubeDB supports the following Pgpool versions:
- `4.4.5`
- `4.5.0`

> The listed PgpoolVersions are tested and provided as a part of the installation process (ie. catalog chart), but you are open to create your own [PgpoolVersion](/docs/guides/pgpool/concepts/catalog.md) object with your custom pgpool image.

## Lifecycle of Pgpool Object

<p align="center">
  <img alt="lifecycle"  src="/docs/images/pgpool/quickstart/lifecycle.png">
</p>

## User Guide

- [Quickstart Pgpool](/docs/guides/pgpool/quickstart/quickstart.md) with KubeDB Operator.
- Detail concepts of [Pgpool object](/docs/guides/pgpool/concepts/pgpool.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
