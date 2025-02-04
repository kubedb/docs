---
title: Redis
menu:
  docs_{{ .version }}:
    identifier: rd-readme-redis
    name: Redis
    parent: rd-redis-guides
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
url: /docs/{{ .version }}/guides/redis/
aliases:
  - /docs/{{ .version }}/guides/redis/README/
---

> New to KubeDB? Please start [here](/docs/README.md).

## Supported Redis Features
| Features                                                          | Availability |
|-------------------------------------------------------------------|:------------:|
| Clustering (Sharding, Replication)                                |   &#10003;   |
| Redis in Sentinel Mode (Use separate Sentinel Cluster for Redis)  |   &#10003;   |
| Standalone Mode                                                   |   &#10003;   |
| Custom Configuration                                              |   &#10003;   |
| Using Custom Docker Image                                         |   &#10003;   |
| Initialization From Script (shell or lua script)                  |   &#10003;   |
| Initializing from Snapshot ([KubeStash](https://kubestash.com/))  |   &#10003;   |
| Authentication & Authorization                                    |   &#10003;   |
| Externally manageable Authentication Secret                       |   &#10003;   |
| Persistent Volume                                                 |   &#10003;   |
| Reconfigurable Health Checker                                     |   &#10003;   |
| Backup (Instant, Scheduled)                                       |   &#10003;   |
| Builtin Prometheus Discovery                                      |   &#10003;   |
| Using Prometheus Operator                                         |   &#10003;   |
| Automated Version Update                                          |   &#10003;   |
| Automatic Vertical Scaling, Volume Expansion                      |   &#10003;   |
| Automated Horizontal Scaling                                      |   &#10003;   |
| Automated db-configure Reconfiguration                            |   &#10003;   |
| TLS configuration ([Cert Manager](https://cert-manager.io/docs/)) |   &#10003;   |
| Reconfiguration of TLS: Add, Remove, Update, Rotate               |   &#10003;   |
| Autoscaling Compute and Storage Resources (vertically)            |   &#10003;   |
| Grafana Dashboards                                                |   &#10003;   |



## Life Cycle of a Redis Object

<p align="center">
  <img alt="lifecycle"  src="/docs/images/redis/redis-lifecycle.png">
</p>

## User Guide

- [Quickstart Redis](/docs/guides/redis/quickstart/quickstart.md) with KubeDB Operator.
- [Deploy Redis Cluster](/docs/guides/redis/clustering/redis-cluster.md) using KubeDB.
- Monitor your Redis server with KubeDB using [out-of-the-box Prometheus operator](/docs/guides/redis/monitoring/using-prometheus-operator.md).
- Monitor your Redis server with KubeDB using [out-of-the-box builtin-Prometheus](/docs/guides/redis/monitoring/using-builtin-prometheus.md).
- Use [private Docker registry](/docs/guides/redis/private-registry/using-private-registry.md) to deploy Redis with KubeDB.
- Use [kubedb cli](/docs/guides/redis/cli/cli.md) to manage databases like kubectl for Kubernetes.
- Detail concepts of [Redis object](/docs/guides/redis/concepts/redis.md).
- Detail concepts of [RedisVersion object](/docs/guides/redis/concepts/catalog.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
