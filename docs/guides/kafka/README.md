---
title: Kafka
menu:
  docs_{{ .version }}:
    identifier: kf-readme-kafka
    name: Kafka
    parent: kf-kafka-guides
    weight: 10
menu_name: docs_{{ .version }}
section_menu_id: guides
url: /docs/{{ .version }}/guides/kafka/
aliases:
  - /docs/{{ .version }}/guides/kafka/README/
---

> New to KubeDB? Please start [here](/docs/README.md).

## Supported Kafka Features


| Features                                                       | Community | Enterprise |
|----------------------------------------------------------------|:---------:|:----------:|
| Clustering - Combined (shared controller and broker nodes)     | &#10003;  |  &#10003;  |
| Clustering - Topology (dedicated controllers and broker nodes) | &#10003;  |  &#10003;  |
| Custom Docker Image                                            | &#10003;  |  &#10003;  |
| Authentication & Authorization                                 | &#10003;  |  &#10003;  |
| Persistent Volume                                              | &#10003;  |  &#10003;  |
| Custom Volume                                                  | &#10003;  |  &#10003;  |
| TLS: using ( [Cert Manager](https://cert-manager.io/docs/) )   | &#10007;  |  &#10003;  |
| Reconfigurable Health Checker                                  | &#10007;  |  &#10003;  |
| Externally manageable Auth Secret                              | &#10007;  |  &#10003;  |

## Supported Kafka Versions

KubeDB supports Kafka versions `3.3.0` in Kraft mode






> The listed ElasticsearchVersions are tested and provided as a part of the installation process (ie. catalog chart), but you are open to create your own [ElasticsearchVersion](/docs/guides/elasticsearch/concepts/catalog/index.md) object with your custom Elasticsearch image.

## User Guide

- [Quickstart Elasticsearch](/docs/guides/elasticsearch/quickstart/overview/index.md) with KubeDB Operator.
- [Elasticsearch Clustering](/docs/guides/elasticsearch/clustering/combined-cluster/index.md) supported by KubeDB
- [Backup & Restore Elasticsearch](/docs/guides/elasticsearch/backup/overview/index.md) database using Stash.
- Monitor your Elasticsearch database with KubeDB using [`out-of-the-box` builtin-Prometheus](/docs/guides/elasticsearch/monitoring/using-builtin-prometheus.md).
- Monitor your Elasticsearch database with KubeDB using [`out-of-the-box` Prometheus operator](/docs/guides/elasticsearch/monitoring/using-prometheus-operator.md).
- Use [private Docker registry](/docs/guides/elasticsearch/private-registry/using-private-registry.md) to deploy Elasticsearch with KubeDB.
- Use [kubedb cli](/docs/guides/elasticsearch/cli/cli.md) to manage databases like kubectl for Kubernetes.
- Detail concepts of [Elasticsearch object](/docs/guides/elasticsearch/concepts/elasticsearch/index.md).
- Want to hack on KubeDB? Check our [contribution guidelines](/docs/CONTRIBUTING.md).
