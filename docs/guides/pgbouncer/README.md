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

> New to KubeDB? Please start [here](/docs/README.md).

# Overview

[PgBouncer](https://pgbouncer.github.io/) is an open-source, lightweight, single-binary connection-pooling middleware for PostgreSQL. PgBouncer maintains a pool of connections for each locally stored user-database pair. It is typically configured to hand out one of these connections to a new incoming client connection, and return it back in to the pool when the client disconnects. PgBouncer can manage only one PostgreSQL database on possibly different servers and serve clients over TCP and Unix domain sockets. For a more hands-on experience, see this brief [tutorial on how to create a PgBouncer](https://pgdash.io/blog/pgbouncer-connection-pool.html) for PostgreSQL database.

KubeDB operator now comes bundled with PgBouncer crd to handle connection pooling. With connection pooling, clients connect to a proxy server which maintains a pool of direct connections to other real PostgreSQL servers. PgBouncer crd can handle multiple local or remote Postgres database connections across multiple users using PgBouncer's connection pooling mechanism.

## PgBouncer Features

| Features                                                    | Availability |
|-------------------------------------------------------------| :----------: |
| Multiple PgBouncer Versions                                 |   &#10003;   |
| Custom Configuration                                        |   &#10003;   |
| Externally manageable Auth Secret                           |   &#10003;   |
| Reconfigurable Health Checker                               |   &#10003;   |
| Integrate with externally managed PostgreSQL                |   &#10003;   |
| Sync Postgres Users to PgBouncer                            |   &#10003;   |
| Custom docker images                                        |   &#10003;   |
| TLS: Add ( [Cert Manager]((https://cert-manager.io/docs/))) |   &#10003;   |
| Reconfigure TLS                                             |   &#10003;   |
| Monitoring with Prometheus & Grafana                        |   &#10003;   |
| Builtin Prometheus Discovery                                |   &#10003;   |
| Using Prometheus operator                                   |   &#10003;   |
| Alert Dashboard                                             |   &#10003;   |
| Grafana Dashboard                                           |   &#10003;   |
| Update PgBouncer Version                                    |   &#10003;   |
| Horizontal Scaling                                          |   &#10003;   |
| Vertical Scaling                                            |   &#10003;   |
| Autoscaling (Compute Resources)                             |   &#10003;   |
| Restart                                                     |   &#10003;   |
| Rotate Authentication Credentials                           |   &#10003;   |
| Initialization from Git Repository                          |   &#10003;   |
| Virtual Secrets                                             |   &#10003;   |
| Private Docker Registry                                     |   &#10003;   |

## User Guide

- [Quickstart PgBouncer](/docs/guides/pgbouncer/quickstart/quickstart.md) with KubeDB Operator.
- [Update version](/docs/guides/pgbouncer/update-version/update_version.md) of PgBouncer.
- [Horizontal Scale](/docs/guides/pgbouncer/scaling/horizontal-scaling/horizontal-ops.md) PgBouncer.
- [Vertical Scale](/docs/guides/pgbouncer/scaling/vertical-scaling/vertical-ops.md) PgBouncer.
- [Autoscale](/docs/guides/pgbouncer/autoscaler/compute/compute-autoscale.md) compute resources of PgBouncer.
- [Reconfigure](/docs/guides/pgbouncer/reconfigure/reconfigure-pgbouncer.md) PgBouncer.
- [Configure TLS/SSL](/docs/guides/pgbouncer/tls/configure_ssl.md) for PgBouncer.
- [Reconfigure TLS/SSL](/docs/guides/pgbouncer/reconfigure-tls/reconfigure-tls.md) for PgBouncer.
- [Restart](/docs/guides/pgbouncer/restart/restart.md) PgBouncer.
- [Rotate Authentication Credentials](/docs/guides/pgbouncer/rotateauth/rotateauth.md) of PgBouncer.
- [Sync Users](/docs/guides/pgbouncer/sync-users/sync-users-pgbouncer.md) to PgBouncer at runtime.
- [Initialize PgBouncer from Git Repository](/docs/guides/pgbouncer/initialization/gitsync.md).
- [Use Virtual Secrets](/docs/guides/pgbouncer/virtual_secret/guide.md) for PgBouncer credentials.
- Monitor your PgBouncer with KubeDB using [`out-of-the-box` builtin-Prometheus](/docs/guides/pgbouncer/monitoring/using-builtin-prometheus.md).
- Monitor your PgBouncer with KubeDB using [`out-of-the-box` Prometheus operator](/docs/guides/pgbouncer/monitoring/using-prometheus-operator.md).
- Use [private Docker registry](/docs/guides/pgbouncer/private-registry/using-private-registry.md) to deploy PgBouncer with KubeDB.
- Setup [custom PgBouncer versions](/docs/guides/pgbouncer/custom-versions/setup.md) with KubeDB.
- Detail concepts of [PgBouncer object](/docs/guides/pgbouncer/concepts/pgbouncer.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
