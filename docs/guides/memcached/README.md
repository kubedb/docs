---
title: Memcached
menu:
  docs_{{ .version }}:
    identifier: mc-readme-memcached
    name: Memcached
    parent: mc-memcached-guides
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
url: /docs/{{ .version }}/guides/memcached/
aliases:
  - /docs/{{ .version }}/guides/memcached/README/
---

> New to KubeDB? Please start [here](/docs/README.md).

## Overview
`Memcached` is an in-memory key-value store that allows for high-performance, low-latency data caching. It is often used to speed up dynamic web applications by offloading frequent, computationally expensive database queries and storing data that needs to be retrieved fast. Memcached is frequently used in contexts where the rapid retrieval of tiny data items, such as session data, query results, and user profiles, is critical for increasing application performance. It is especially well-suited for use cases requiring scalability, distributed caching, and high throughput, making it a popular choice for powering online and mobile apps, particularly in high-concurrency environments. Memcached is a simple, extremely efficient caching layer that minimizes stress on backend systems and improves performance for applications that require real-time data.

## Supported Memcached Features

| Features                                              | Availability |
| ----------------------------------------------------- | :----------: |
| Custom Configuration                                  |   &#10003;   |
| Externally manageable Auth Secret                     |   &#10003;   |
| Reconfigurable Health Checker                         |   &#10003;   |
| Using Custom docker image                             |   &#10003;   |
| Builtin Prometheus Discovery                          |   &#10003;   |
| Operator Managed Prometheus Discovery                 |   &#10003;   |
| Automated Version Update                              |   &#10003;   |
| Automated Vertical Scaling                            |   &#10003;   |
| Automated Horizontal Scaling                          |   &#10003;   |
| Automated db-configure Reconfiguration                |   &#10003;   |
| Authentication & Authorization                        |   &#10003;   |
| TLS: Add, Remove, Update, Rotate ( Cert Manager )     |   &#10003;   |
| Autoscaling (Vertically)                              |   &#10003;   |
| Multiple Memcached Versions                           |   &#10003;   |
| Monitoring using Prometheus and Grafana               |   &#10003;   |
| Restart                                               |   &#10003;   |
| Custom RBAC                                           |   &#10003;   |

## Life Cycle of a Memcached Object

<p align="center">
  <img alt="lifecycle"  src="/docs/images/memcached/memcached-lifecycle.png">
</p>

## User Guide

- [Quickstart Memcached](/docs/guides/memcached/quickstart/quickstart.md) with KubeDB Operator.
- Monitor your Memcached server with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/memcached/monitoring/using-prometheus-operator.md).
- Monitor your Memcached server with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/memcached/monitoring/using-builtin-prometheus.md).
- Use [private Docker registry](/docs/guides/memcached/private-registry/using-private-registry.md) to deploy Memcached with KubeDB.
- Use [kubedb cli](/docs/guides/memcached/cli/cli.md) to manage databases like kubectl for Kubernetes.
- Detail concepts of [Memcached object](/docs/guides/memcached/concepts/memcached.md).
- [Horizontal Scale](/docs/guides/memcached/scaling/horizontal-scaling/horizontal-scaling.md) your Memcached cluster.
- [Vertical Scale](/docs/guides/memcached/scaling/vertical-scaling/vertical-scaling.md) your Memcached cluster.
- [Autoscale](/docs/guides/memcached/autoscaler/compute/compute-autoscale.md) compute resources of your Memcached cluster.
- [Update Version](/docs/guides/memcached/update-version/update-version.md) of your Memcached cluster.
- [Reconfigure](/docs/guides/memcached/reconfigure/reconfigure.md) your Memcached cluster.
- Configure [TLS/SSL Encryption](/docs/guides/memcached/tls/tls.md) for your Memcached cluster.
- [Reconfigure TLS](/docs/guides/memcached/reconfigure-tls/reconfigure-tls.md) for your Memcached cluster.
- [Restart](/docs/guides/memcached/restart/restart.md) your Memcached cluster.
- [Rotate Authentication](/docs/guides/memcached/rotate-auth/rotateauth.md) credentials of your Memcached cluster.
- Use [Custom RBAC](/docs/guides/memcached/custom-rbac/using-custom-rbac.md) to manage Memcached with fine-grained access control.
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
