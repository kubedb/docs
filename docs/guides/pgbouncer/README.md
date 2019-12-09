---
title: PgBouncer
menu:
  docs_{{ .version }}:
    identifier: pb-readme-pgbouncer
    name: PgBouncer
    parent: pb-pgbouncer-guides
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
url: /docs/{{ .version }}/guides/pgbouncer/
aliases:
  - /docs/{{ .version }}/guides/pgbouncer/README/
---

> New to KubeDB? Please start [here](/docs/concepts/README.md).
>
# Overview

[PgBouncer](https://pgbouncer.github.io/) is an open-source, lightweight, single-binary connection-pooling middleware for PostgreSQL. PgBouncer maintains a pool of connections for each locally stored user-database pair. It is typically configured to hand out one of these connections to a new incoming client connection, and return it back in to the pool when the client disconnects. PgBouncer can manage one or more PostgreSQL databases on possibly different servers and serve clients over TCP and Unix domain sockets. For a more hands-on experience, see this brief [tutorial on how to create a PgBouncer](https://pgdash.io/blog/pgbouncer-connection-pool.html) for PostgreSQL database.

KubeDB operator now comes bundled with PgBouncer crd to handle connection pooling. With connection pooling, clients connect to a proxy server which maintains a pool of direct connections to other real PostgreSQL servers. PgBouncer crd can handle multiple local or remote Postgres database connections across multiple users using PgBouncer's connection pooling mechanism.

## PgBouncer Features

| Features                           | Availability |
| ---------------------------------- | :----------: |
| Multiple PgBouncer Versions        |   &#10003;   |
| Customizable Pooling Configuration |   &#10003;   |
| Custom docker images               |   &#10003;   |
| Builtin Prometheus Discovery       |   &#10003;   |
| Using CoreOS Prometheus Operator   |   &#10003;   |

## Supported PgBouncer Versions

| KubeDB Version ↓|   1.7  |  1.7.1 |  1.7.2 |  1.8.1 |  1.9.0 |  1.10.0|  1.11.0|
| -------------- | :---:  | :---:  | :----: | :----: | :----: | :----: | :----: |
| v0.13.0-rc.1   |&#10003;|&#10003;|&#10003;|&#10003;|&#10003;|&#10003;|&#10003;|

## Supported PgBouncerVersion CRD
PgBouncerVersion crd is used to fetch a specific iteration of PgBouncer image for a given PgBouncer Version. For example, there can be two PgBouncerVersion crd (1.7, and 1.7-v2 ) specified for a single PgBouncer release version 1.7 and images specified in those crds will differ in terms of features, and improvements.

Here, &#10003; means supported and &#10007; means unsupported.

|   NAME  ↓| VERSION ↓| KubeDB: v0.13.0-rc.0 | KubeDB: v0.13.0-rc.1 |
| :------: | :------: | :------------------: | :------------------: |
|   1.7    |   1.7    |      &#10007;        |       &#10003;       |
|   1.7.1  |  1.7.1   |      &#10007;        |       &#10003;       |
|   1.7.2  |  1.7.2   |      &#10007;        |       &#10003;       |
|   1.8.1  |  1.8.1   |      &#10007;        |       &#10003;       |
|   1.9.0  |  1.9.0   |      &#10007;        |       &#10003;       |
|   1.10.0 |  1.10.0  |      &#10007;        |       &#10003;       |
|   1.11.0 |  1.11.0  |      &#10007;        |       &#10003;       |

## External tools dependency

| Tool                                    | Version |
| --------------------------------------- | :-----: |
| [pgbouncer_exporter](https://github.com/kubedb/pgbouncer_exporter) | 0.0.3  |

## User Guide

- [Quickstart PgBouncer](/docs/guides/pgbouncer/quickstart/quickstart.md) with KubeDB Operator.
- Monitor your PgBouncer with KubeDB using [`out-of-the-box` builtin-Prometheus](/docs/guides/pgbouncer/monitoring/using-builtin-prometheus.md).
- Monitor your PgBouncer with KubeDB using [`out-of-the-box` CoreOS Prometheus Operator](/docs/guides/pgbouncer/monitoring/using-coreos-prometheus-operator.md).
- Use [private Docker registry](/docs/guides/pgbouncer/private-registry/using-private-registry.md) to deploy PgBouncer with KubeDB.
- Detail concepts of [PgBouncer object](/docs/concepts/database-proxy/pgbouncer.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
