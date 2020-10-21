---
title: Elasticsearch
menu:
  docs_{{ .version }}:
    identifier: es-readme-elasticsearch
    name: Elasticsearch
    parent: es-elasticsearch-guides
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
url: /docs/{{ .version }}/guides/elasticsearch/
aliases:
  - /docs/{{ .version }}/guides/elasticsearch/README/
---

> New to KubeDB? Please start [here](/docs/README.md).

## Supported Elasticsearch Features

| Features                                                                              | Availability |
| ------------------------------------------------------------------------------------- | :----------: |
| Clustering                                                                            |   &#10003;   |
| Authentication (using [Search Guard](https://github.com/floragunncom/search-guard))   |   &#10003;   |
| Authorization (using [Search Guard](https://github.com/floragunncom/search-guard))    |   &#10003;   |
| TLS certificates (using [Search Guard](https://github.com/floragunncom/search-guard)) |   &#10003;   |
| Persistent Volume                                                                     |   &#10003;   |
| Instant Backup                                                                        |   &#10003;   |
| Scheduled Backup                                                                      |   &#10003;   |
| Initialization from Script                                                            |   &#10007;   |
| Initialization from Snapshot                                                          |   &#10003;   |
| Builtin Prometheus Discovery                                                          |   &#10003;   |
| Using Prometheus operator                                                      |   &#10003;   |
| Custom Configuration                                                                  |   &#10003;   |
| Using Custom Docker Image                                                             |   &#10003;   |

## Life Cycle of an Elasticsearch Object

<p align="center">
  <img alt="lifecycle"  src="/docs/images/elasticsearch/lifecycle.png">
</p>

## Supported Elasticsearch Version

| KubeDB Version | Elasticsearch:2.3 | Elasticsearch:5.6 | Elasticsearch:6.2 | Elasticsearch:6.3 | Elasticsearch:6.4 | Elasticsearch:6.5 | Elasticsearch:6.8 | Elasticsearch:7.2 | Elasticsearch:7.3 |
| :------------: | :---------------: | :---------------: | :---------------: | :---------------: | :---------------: | :---------------: | :---------------: | :---------------: | :---------------: |
| 0.1.0 - 0.7.0  |     &#10003;      |     &#10007;      |     &#10007;      |     &#10007;      |     &#10007;      |     &#10007;      |     &#10007;      |     &#10007;      |     &#10007;      |
|     0.8.0      |     &#10007;      |     &#10003;      |     &#10003;      |     &#10007;      |     &#10007;      |     &#10007;      |     &#10007;      |     &#10007;      |     &#10007;      |
|     0.9.0      |     &#10007;      |     &#10003;      |     &#10003;      |     &#10003;      |     &#10003;      |     &#10007;      |     &#10007;      |     &#10007;      |     &#10007;      |
|     0.10.0     |     &#10007;      |     &#10003;      |     &#10003;      |     &#10003;      |     &#10003;      |     &#10003;      |     &#10007;      |     &#10007;      |     &#10007;      |
|     0.11.0     |     &#10007;      |     &#10003;      |     &#10003;      |     &#10003;      |     &#10003;      |     &#10003;      |     &#10007;      |     &#10007;      |     &#10007;      |
|     0.12.0     |     &#10007;      |     &#10003;      |     &#10003;      |     &#10003;      |     &#10003;      |     &#10003;      |     &#10007;      |     &#10007;      |     &#10007;      |
|    v0.13.0     |     &#10007;      |     &#10003;      |     &#10003;      |     &#10003;      |     &#10003;      |     &#10003;      |     &#10003;      |     &#10003;      |     &#10003;      |

## Supported ElasticsearchVersion CRD

Here, &#10003; means supported and &#10007; means deprecated.

|   NAME   | VERSION | KubeDB: 0.9.0 | KubeDB: 0.10.0 | KubeDB: 0.11.0 | KubeDB: 0.12.0 | KubeDB: v0.13.0 |
| :------: | :-----: | :-----------: | :------------: | :------------: | :------------: | :-------------: |
|   5.6    |   5.6   |   &#10007;    |    &#10007;    |    &#10007;    |    &#10007;    |    &#10007;     |
|  5.6-v1  |   5.6   |   &#10003;    |    &#10003;    |    &#10003;    |    &#10003;    |    &#10003;     |
|  5.6.4   |  5.6.4  |   &#10007;    |    &#10007;    |    &#10007;    |    &#10007;    |    &#10007;     |
| 5.6.4-v1 |  5.6.4  |   &#10003;    |    &#10003;    |    &#10003;    |    &#10003;    |    &#10003;     |
|   6.2    |   6.2   |   &#10007;    |    &#10007;    |    &#10007;    |    &#10007;    |    &#10007;     |
|  6.2-v1  |   6.2   |   &#10003;    |    &#10003;    |    &#10003;    |    &#10003;    |    &#10003;     |
|  6.2.4   |  6.2.4  |   &#10007;    |    &#10007;    |    &#10007;    |    &#10007;    |    &#10007;     |
| 6.2.4-v1 |  6.2.4  |   &#10003;    |    &#10003;    |    &#10003;    |    &#10003;    |    &#10003;     |
|   6.3    |   6.3   |   &#10007;    |    &#10007;    |    &#10007;    |    &#10007;    |    &#10007;     |
|  6.3-v1  |   6.3   |   &#10003;    |    &#10003;    |    &#10003;    |    &#10003;    |    &#10003;     |
|  6.3.0   |  6.3.0  |   &#10007;    |    &#10007;    |    &#10007;    |    &#10007;    |    &#10007;     |
| 6.3.0-v1 |  6.3.0  |   &#10003;    |    &#10003;    |    &#10003;    |    &#10003;    |    &#10003;     |
|   6.4    |   6.4   |   &#10003;    |    &#10003;    |    &#10003;    |    &#10003;    |    &#10003;     |
|  6.4.0   |  6.4.0  |   &#10003;    |    &#10003;    |    &#10003;    |    &#10003;    |    &#10003;     |
|  6.5.3   |  6.5.3  |   &#10007;    |    &#10003;    |    &#10003;    |    &#10003;    |    &#10003;     |
|   6.5    |  6.5.3  |   &#10007;    |    &#10003;    |    &#10003;    |    &#10003;    |    &#10003;     |
|  6.8.0   |  6.8.0  |   &#10007;    |    &#10007;    |    &#10007;    |    &#10007;    |    &#10003;     |
|   6.8    |  6.8.0  |   &#10007;    |    &#10007;    |    &#10007;    |    &#10007;    |    &#10003;     |
|  7.3.2   |   7.2   |   &#10007;    |    &#10007;    |    &#10007;    |    &#10007;    |    &#10003;     |
|   7.2    |  7.3.2  |   &#10007;    |    &#10007;    |    &#10007;    |    &#10007;    |    &#10003;     |
|  7.3.2   |  7.3.2  |   &#10007;    |    &#10007;    |    &#10007;    |    &#10007;    |    &#10003;     |
|   7.3    |   7.3   |   &#10007;    |    &#10007;    |    &#10007;    |    &#10007;    |    &#10003;     |

## External tools dependency

|                               Tool                               |           Version           |
| :--------------------------------------------------------------: | :-------------------------: |
|   [Search Guard](https://github.com/floragunncom/search-guard)   | `5.6.4-18` and `6.2.4-22.1` |
| [Elasticdump](https://github.com/taskrabbit/elasticsearch-dump/) |    `3.3.1` and `3.3.14`     |
|              [osm](https://github.com/appscode/osm)              |           `0.9.1`           |
|              [yq](https://github.com/mikefarah/yq)               |           `2.1.1`           |

## User Guide

- [Quickstart Elasticsearch](/docs/guides/elasticsearch/quickstart/quickstart.md) with KubeDB Operator.
- [Backup & Restore Elasticsearch](/docs/guides/elasticsearch/backup/stash.md) database using Stash.
- [Elasticsearch Topology](/docs/guides/elasticsearch/clustering/topology.md) supported by KubeDB
- Monitor your Elasticsearch database with KubeDB using [`out-of-the-box` builtin-Prometheus](/docs/guides/elasticsearch/monitoring/using-builtin-prometheus.md).
- Monitor your Elasticsearch database with KubeDB using [`out-of-the-box` Prometheus operator](/docs/guides/elasticsearch/monitoring/using-prometheus-operator.md).
- Use [private Docker registry](/docs/guides/elasticsearch/private-registry/using-private-registry.md) to deploy Elasticsearch with KubeDB.
- Use [kubedb cli](/docs/guides/elasticsearch/cli/cli.md) to manage databases like kubectl for Kubernetes.
- Detail concepts of [Elasticsearch object](/docs/guides/elasticsearch/concepts/elasticsearch.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
