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

## Supported Memcached Features

| Features                               | Availability |
| ------------------------------------   | :----------: |
| Clustering                             |   &#10007;   |
| Persistent Volume                      |   &#10007;   |
| Instant Backup                         |   &#10007;   |
| Scheduled Backup                       |   &#10007;   |
| Initialize using Snapshot              |   &#10007;   |
| Initialize using Script                |   &#10007;   |
| Multiple Memcached Versions         |   &#10003;   |
| Custom Configuration                   |   &#10003;   |
| Externally manageable Auth Secret	     |   &#10007;   |
| Reconfigurable Health Checker		     |   &#10003;   |
| Using Custom docker image              |   &#10003;   |
| Builtin Prometheus Discovery           |   &#10003;   |
| Using Prometheus operator              |   &#10003;   |
| Automated Version Update               |   &#10003;   |
| Automated Vertical Scaling             |   &#10003;   |
| Automated Horizontal Scaling           |   &#10003;   |
| Automated db-configure Reconfiguration |   &#10003;   |
| TLS: Add, Remove, Update, Rotate ( Cert Manager )	|&#10007;|
| Automated Volume Expansion	           |   &#10007;   |
| Autoscaling (Vertically)               |   &#10003;   |
| Grafana Dashboard               |   &#10003;   |
| Alert Dashboard	               |   &#10007;   |



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
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
