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
| Features                                                                           | Community | Enterprise |
|------------------------------------------------------------------------------------|:---------:|:----------:|
| Clustering                                                                         | &#10003;  |  &#10003;  |
| Sentinel                                                                           | &#10003;  |  &#10003;  |
| Standalone                                                                         | &#10003;  |  &#10003;  |
| Authentication & Autorization                                                      | &#10003;  |  &#10003;  |
| Persistent Volume                                                                  | &#10003;  |  &#10003;  |
| Initializing from Snapshot ( [Stash](https://stash.run/) )                         | &#10003;  |  &#10003;  |
| Instant Backup (Sentinel and Standalone Mode)                                      | &#10003;  |  &#10003;  |
| Scheduled Backup (Sentinel and Standalone Mode)                                    | &#10003;  |  &#10003;  |
| Builtin Prometheus Discovery                                                       | &#10003;  |  &#10003;  |
| Using Prometheus operator                                                          | &#10003;  |  &#10003;  |
| Automated Version Update                                                           | &#10007;  |  &#10003;  |
| Automatic Vertical Scaling                                                         | &#10007;  |  &#10003;  |
| Automated Horizontal Scaling                                                       | &#10007;  |  &#10003;  |
| Automated db-configure Reconfiguration                                             | &#10007;  |  &#10003;  |
| TLS: Add, Remove, Update, Rotate ( [Cert Manager](https://cert-manager.io/docs/) ) | &#10007;  |  &#10003;  |
| Automated Volume Expansion                                                         | &#10007;  |  &#10003;  |
| Autoscaling (vertically)                                                           | &#10007;  |  &#10003;  |


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
